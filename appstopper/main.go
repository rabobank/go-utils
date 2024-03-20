package main

import (
	"context"
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient/v3/client"
	"github.com/cloudfoundry-community/go-cfclient/v3/config"
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
	skipSSLValidationStr = os.Getenv("SKIP_SSL_VALIDATION")
	skipSSLValidation    bool
	dryRun               = os.Getenv("DRY_RUN")
	runType              = os.Getenv("RUN_TYPE")
	missingLabelAction   = os.Getenv("MISSING_LABEL_ACTION")
	excludedOrgsStr      = os.Getenv("EXCLUDED_ORGS")
	excludedOrgs         []string
	excludedSpacesStr    = os.Getenv("EXCLUDED_SPACES")
	excludedSpaces       []string
	cfConfig             *config.Config
	ctx                  = context.Background()
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
	if runType == "" {
		runType = "weekly"
	} else if runType != "weekly" && runType != "daily" && runType != "daily,weekly" {
		fmt.Printf("invalid value (%s) for RUN_TYPE, must be either 'weekly', 'daily' or 'daily,weekly'\n", runType)
		envComplete = false
	}

	if missingLabelAction == "" {
		missingLabelAction = "daily"
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
	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" CF_API_ADDR: %s\n", apiAddress)
		fmt.Printf(" CF_USERNAME: %s\n", cfUsername)
		fmt.Printf(" SKIP_SSL_VALIDATION: %t\n", skipSSLValidation)
		fmt.Printf(" EXCLUDED_ORGS: %s\n", excludedOrgs)
		fmt.Printf(" EXCLUDED_SPACES: %s\n", excludedSpaces)
		fmt.Printf(" DRY_RUN: %s\n", dryRun)
		fmt.Printf(" RUN_TYPE: %s\n\n", runType)
	}
	return envComplete
}

func getCFClient() (cfClient *client.Client) {
	var err error
	if cfConfig, err = config.NewClientSecret(apiAddress, cfUsername, cfPassword); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		cfConfig.WithSkipTLSValidation(skipSSLValidation)
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
	cfClient := getCFClient()
	if orgs, err := cfClient.Organizations.ListAll(ctx, nil); err != nil {
		fmt.Printf("failed to list orgs: %s", err)
		os.Exit(1)
	} else {
		var totalVictims int
		startTime := time.Now()
		for _, org := range orgs {
			if !orgNameExcluded(org.Name) {
				if spaces, _, err := cfClient.Spaces.List(ctx, &client.SpaceListOptions{OrganizationGUIDs: client.Filter{Values: []string{org.GUID}}}); err != nil {
					log.Fatalf("failed to list spaces: %s", err)
				} else {
					for _, space := range spaces {
						if !spaceNameExcluded(space.Name) {
							if apps, _, err := cfClient.Applications.List(ctx, &client.AppListOptions{SpaceGUIDs: client.Filter{Values: []string{space.GUID}}}); err != nil {
								log.Fatalf("failed to list all apps: %s", err)
							} else {
								for _, app := range apps {
									autostopLabel := app.Metadata.Labels["AUTOSTOP"]
									if app.State == "STARTED" && ((autostopLabel == nil && strings.Contains(runType, missingLabelAction)) || (autostopLabel != nil && strings.Contains(runType, *autostopLabel))) {
										totalVictims++
										if dryRun != "true" {
											if _, err := cfClient.Applications.Stop(ctx, app.GUID); err != nil {
												fmt.Printf("failed to stop app %s: %s\n", app.Name, err)
											} else {
												fmt.Printf("stopped  %s\n", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
											}
										} else {
											fmt.Printf("(because of DRYRUN=true) not stopped  %s\n", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
										}
									}
								}
							}
						}
					}
				}
			}
		}
		fmt.Printf("\nexecutionTime: %.0f secs, total victims: %d\n", time.Now().Sub(startTime).Seconds(), totalVictims)
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
