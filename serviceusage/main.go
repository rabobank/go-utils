package main

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
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
	cfConfig             *config.Config
	ctx                  = context.Background()
	cfClient             *client.Client
	cfContext            = context.TODO()
	offeringNamesByGuid  = make(map[string]string)
	orgCacheByGuid       = make(map[string]*resource.Organization)
	spaceCacheByGuid     = make(map[string]*resource.Space)
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
	if skipSSLValidationStr == "" {
		skipSSLValidation = false
	} else {
		if skipSSLValidation, err = strconv.ParseBool(skipSSLValidationStr); err != nil {
			fmt.Printf("invalid value (%s) for SKIP_SSL_VALIDATION: %s", skipSSLValidationStr, err)
			envComplete = false
		}
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
		fmt.Printf(" SERVICE_OFFERINGS: %s\n\n", serviceOfferings)
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
	servicePlanListOptions := client.ServicePlanListOptions{ListOptions: &client.ListOptions{}, ServiceOfferingNames: client.Filter{Values: serviceOfferings}}
	if servicePlans, err := cfClient.ServicePlans.ListAll(cfContext, &servicePlanListOptions); err != nil {
		log.Fatalf("failed to list service plans: %s", err)
	} else {
		fmt.Printf("%-15s %-60s %-20s %-20s %-30s %-40s %-40s\n", "Offering name", "Service Instance", "Updated At", "Last operation", "Org", "Space", "Bound App")

		for _, servicePlan := range servicePlans {
			serviceInstanceListOptions := client.ServiceInstanceListOptions{
				ListOptions:      &client.ListOptions{PerPage: 4999},
				ServicePlanGUIDs: client.Filter{Values: []string{servicePlan.GUID}},
			}
			if serviceInstances, err := cfClient.ServiceInstances.ListAll(ctx, &serviceInstanceListOptions); err != nil {
				log.Printf("failed to list service instances: %s", err)
			} else {
				for _, serviceInstance := range serviceInstances {
					if serviceBindings, err := cfClient.ServiceCredentialBindings.ListAll(cfContext, &client.ServiceCredentialBindingListOptions{ListOptions: &client.ListOptions{}, ServiceInstanceGUIDs: client.Filter{Values: []string{serviceInstance.GUID}}}); err != nil {
						log.Printf("failed to list service bindings for service instance guid %s: %s", serviceInstance.GUID, err)
					} else {
						for _, serviceBinding := range serviceBindings {
							if app, err := cfClient.Applications.Get(ctx, serviceBinding.Relationships.App.Data.GUID); err != nil {
								log.Printf("failed to get app for guid %s: %s", serviceBinding.Relationships.App.Data.GUID, err)
							} else {
								if space, err := getSpaceByGuidCached(serviceInstance.Relationships.Space.Data.GUID); err != nil {
									log.Printf("failed to get space for guid %s: %s", serviceInstance.Relationships.Space.Data.GUID, err)
								} else {
									if org, err := getOrgByGuidCached(space.Relationships.Organization.Data.GUID); err != nil {
										log.Printf("failed to get org for guid %s: %s", space.Relationships.Organization.Data.GUID, err)
									} else {
										fmt.Printf("%-15s %-60s %-20s %-20s %-30s %-40s %-40s\n", offeringNameByGuid(servicePlan.Relationships.ServiceOffering.Data.GUID), serviceInstance.Name, serviceInstance.UpdatedAt.Format(time.RFC3339), serviceInstance.LastOperation.UpdatedAt.Format(time.RFC3339), org.Name, space.Name, app.Name)
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

func getOrgByGuidCached(orgGuid string) (org *resource.Organization, err error) {
	inCache := false
	if org, inCache = orgCacheByGuid[orgGuid]; inCache {
		return org, nil
	}
	if org, err = cfClient.Organizations.Get(ctx, orgGuid); err != nil {
		return nil, err
	} else {
		orgCacheByGuid[orgGuid] = org
		return org, nil
	}
}

func getSpaceByGuidCached(spaceGuid string) (space *resource.Space, err error) {
	inCache := false
	if space, inCache = spaceCacheByGuid[spaceGuid]; inCache {
		return space, nil
	}
	if space, err = cfClient.Spaces.Get(ctx, spaceGuid); err != nil {
		return nil, err
	} else {
		spaceCacheByGuid[spaceGuid] = space
		return space, nil
	}
}
