package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	apiAddress           = os.Getenv("CF_API_ADDR")
	cfUsername           = os.Getenv("CF_USERNAME")
	cfPassword           = os.Getenv("CF_PASSWORD")
	serviceOfferingsStr  = os.Getenv("SERVICE_OFFERINGS")
	serviceOfferings     []string
	skipSSLValidationStr = os.Getenv("SKIP_SSL_VALIDATION")
	skipSSLValidation    bool
	dryRunStr            = os.Getenv("DRY_RUN")
	dryRun               = true
	excludedOrgsStr      = os.Getenv("EXCLUDED_ORGS")
	excludedOrgs         []string
	excludedSpacesStr    = os.Getenv("EXCLUDED_SPACES")
	gracePeriodStr       = os.Getenv("GRACE_PERIOD")
	graceDate            time.Time
	excludedSpaces       []string
	cfConfig             *config.Config
	ctx                  = context.Background()
	cfClient             *client.Client
	cfContext            = context.TODO()
	offeringNamesByGuid  = make(map[string]string)
	totalVictims         int
)

func environmentComplete() bool {
	var err error
	envComplete := true
	if apiAddress == "" {
		fmt.Println("missing envvar : CF_API_ADDR")
		envComplete = false
	}
	if cfUsername == "" {
		fmt.Println("missing envvar : CF_USERNAME")
		envComplete = false
	}
	if cfPassword == "" {
		fmt.Println("missing envvar : CF_PASSWORD")
		envComplete = false
	}
	if cfPassword == "" {
		fmt.Println("missing envvar : CF_PASSWORD")
		envComplete = false
	}
	if skipSSLValidationStr == "" {
		skipSSLValidation = false
	} else {
		if skipSSLValidation, err = strconv.ParseBool(skipSSLValidationStr); err != nil {
			fmt.Printf("invalid value (%s) for SKIP_SSL_VALIDATION: %s", skipSSLValidationStr, err)
			envComplete = false
		}
	}

	if dryRunStr == "false" {
		dryRun = false
	}

	if excludedOrgsStr == "" {
		excludedOrgs = []string{"system"}
	} else {
		excludedOrgs = strings.Split(excludedOrgsStr, ",")
	}
	if excludedSpacesStr == "" {
		excludedSpaces = []string{""}
	} else {
		excludedSpaces = strings.Split(excludedSpacesStr, ",")
	}

	if gracePeriod, err := strconv.Atoi(gracePeriodStr); err != nil {
		log.Fatalf("failed to parse grace period: %s", err)
	} else {
		graceDate = time.Now().Add(-time.Hour * 24 * time.Duration(gracePeriod))
	}

	if serviceOfferingsStr == "" {
		fmt.Println("missing envvar : SERVICE_OFFERINGS")
		envComplete = false
	} else {
		serviceOfferings = strings.Split(serviceOfferingsStr, ",")
		if len(serviceOfferings) == 0 {
			fmt.Println("missing envvar : SERVICE_OFFERINGS")
			envComplete = false
		} else {
			for i, serviceOffering := range serviceOfferings {
				serviceOfferings[i] = strings.TrimSpace(serviceOffering)
			}
		}
	}

	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" CF_API_ADDR: %s\n", apiAddress)
		fmt.Printf(" CF_USERNAME: %s\n", cfUsername)
		fmt.Printf(" SKIP_SSL_VALIDATION: %t\n", skipSSLValidation)
		fmt.Printf(" EXCLUDED_ORGS: %s\n", excludedOrgs)
		fmt.Printf(" EXCLUDED_SPACES: %s\n", excludedSpaces)
		fmt.Printf(" DRY_RUN: %t\n", dryRun)
		fmt.Printf(" SERVICE_OFFERINGS: %s\n", serviceOfferings)
		fmt.Printf(" GRACE_PERIOD: %s\n\n", graceDate.Format(time.RFC3339))
	}

	cfClient = getCFClient()

	return envComplete
}

func getCFClient() (cfClient *client.Client) {
	var err error
	if cfConfig, err = config.New(apiAddress, config.ClientCredentials(cfUsername, cfPassword), config.SkipTLSValidation()); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		if cfClient, err = client.New(cfConfig); err != nil {
			log.Fatalf("failed to create new client: %s", err)
		} else {
			// refresh the client every hour to get a new refresh token
			go func() {
				channel := time.Tick(time.Duration(90) * time.Minute)
				for range channel {
					cfClient, err = client.New(cfConfig)
					if err != nil {
						log.Printf("failed to refresh cfclient, error is %s", err)
					}
				}
			}()
		}
	}
	return
}

func main() {
	if !environmentComplete() {
		os.Exit(8)
	}
	startTime := time.Now()
	servicePlanListOptions := client.ServicePlanListOptions{ListOptions: &client.ListOptions{}, ServiceOfferingNames: client.Filter{Values: serviceOfferings}}
	if servicePlans, err := cfClient.ServicePlans.ListAll(cfContext, &servicePlanListOptions); err != nil {
		log.Fatalf("failed to list service plans: %s", err)
	} else {
		dateFilter := client.TimestampFilterList{}
		dateFilter.Before(graceDate)
		if orgs, err := cfClient.Organizations.ListAll(ctx, nil); err != nil {
			log.Printf("failed to list orgs: %s", err)
			os.Exit(1)
		} else {
			fmt.Printf("%-15s %-60s %-20s %-20s %-30s %-40s\n", "Offering name", "Service Instance", "Updated At", "Last operation", "Org", "Space")
			for _, org := range orgs {
				if !orgNameExcluded(org.Name) {
					if spaces, _, err := cfClient.Spaces.List(ctx, &client.SpaceListOptions{OrganizationGUIDs: client.Filter{Values: []string{org.GUID}}}); err != nil {
						log.Fatalf("failed to list spaces: %s", err)
					} else {
						for _, space := range spaces {
							if !spaceNameExcluded(space.Name) {
								for _, servicePlan := range servicePlans {
									serviceInstanceListOptions := client.ServiceInstanceListOptions{
										ListOptions:      &client.ListOptions{PerPage: 4999, UpdatedAts: dateFilter},
										SpaceGUIDs:       client.Filter{Values: []string{space.GUID}},
										ServicePlanGUIDs: client.Filter{Values: []string{servicePlan.GUID}},
									}
									if serviceInstances, err := cfClient.ServiceInstances.ListAll(ctx, &serviceInstanceListOptions); err != nil {
										log.Printf("failed to list service instances: %s", err)
									} else {
										for _, serviceInstance := range serviceInstances {
											if serviceInstance.LastOperation.UpdatedAt.Before(graceDate) {
												if serviceBindings, err := cfClient.ServiceCredentialBindings.ListAll(cfContext, &client.ServiceCredentialBindingListOptions{ListOptions: &client.ListOptions{}, ServiceInstanceGUIDs: client.Filter{Values: []string{serviceInstance.GUID}}}); err != nil {
													log.Printf("failed to list service bindings for service instance guid %s: %s", serviceInstance.GUID, err)
												} else {
													if len(serviceBindings) == 0 {
														if !hasRecentBindingActions(serviceInstance.GUID, space.GUID) {
															if !dryRun {
																if _, err = cfClient.ServiceInstances.Delete(cfContext, serviceInstance.GUID); err != nil {
																	log.Printf("failed to delete service instance %s: %s", serviceInstance.Name, err)
																}
															}
															fmt.Printf("%-15s %-60s %-20s %-20s %-30s %-40s\n", offeringNameByGuid(servicePlan.Relationships.ServiceOffering.Data.GUID), serviceInstance.Name, serviceInstance.UpdatedAt.Format(time.RFC3339), serviceInstance.LastOperation.UpdatedAt.Format(time.RFC3339), org.Name, space.Name)
															totalVictims++
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
			fmt.Println("")
			log.Printf("executionTime: %.0f secs, total victims: %d\n", time.Now().Sub(startTime).Seconds(), totalVictims)
		}
	}
}

func orgNameExcluded(orgName string) bool {
	for _, excludedOrg := range excludedOrgs {
		if orgName == excludedOrg {
			return true
		}
	}
	return false
}

func spaceNameExcluded(spaceName string) bool {
	for _, excludedSpace := range excludedSpaces {
		if spaceName == excludedSpace {
			return true
		}
	}
	return false
}

func offeringNameByGuid(offeringGuid string) string {
	if offeringNamesByGuid[offeringGuid] != "" {
		return offeringNamesByGuid[offeringGuid]
	}
	if offering, err := cfClient.ServiceOfferings.Get(cfContext, offeringGuid); err != nil {
		log.Printf("failed to get service offering for guid %s: %s", offeringGuid, err)
		return ""
	} else {
		offeringNamesByGuid[offeringGuid] = offering.Name
	}
	return offeringNamesByGuid[offeringGuid]
}

func hasRecentBindingActions(serviceInstanceGuid string, spaceGuid string) bool {
	auditEventListOptions := client.AuditEventListOptions{
		ListOptions: &client.ListOptions{PerPage: 4999},
		Types:       client.Filter{Values: []string{"audit.service_binding.create", "audit.service_binding.delete"}},
		SpaceGUIDs:  client.Filter{Values: []string{spaceGuid}},
	}
	if events, err := cfClient.AuditEvents.ListAll(cfContext, &auditEventListOptions); err != nil {
		log.Printf("failed to list audit events for service instance guid %s: %s", serviceInstanceGuid, err)
		return true // just so we don't delete stuff in case of error
	} else {
		for _, event := range events {
			if eventDataBytes, err := event.Data.MarshalJSON(); err != nil {
				log.Printf("failed to marshal event data: %s", err)
				return true // just so we don't delete stuff in case of error
			} else {
				eventData := EventData{}
				if err = json.Unmarshal(eventDataBytes, &eventData); err != nil {
					log.Printf("failed to unmarshal event data: %s", err)
					return true // just so we don't delete stuff in case of error
				} else {
					if eventData.Request.Relationships.ServiceInstance.Data.GUID == serviceInstanceGuid {
						return true
					}
				}
			}
		}
	}
	return false
}

type EventData struct {
	Request struct {
		Type       string `json:"type"`
		Name       any    `json:"name"`
		Parameters struct {
		} `json:"parameters"`
		Relationships struct {
			ServiceInstance struct {
				Data struct {
					GUID string `json:"guid"`
				} `json:"data"`
			} `json:"service_instance"`
			App struct {
				Data struct {
					GUID string `json:"guid"`
				} `json:"data"`
			} `json:"app"`
		} `json:"relationships"`
	} `json:"request"`
	ManifestTriggered bool `json:"manifest_triggered"`
}
