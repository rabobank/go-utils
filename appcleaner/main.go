package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
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
	gracePeriodStr       = os.Getenv("GRACE_PERIOD")
	graceDate            time.Time
	excludedSpaces       []string
	cfConfig             *config.Config
	ctx                  = context.Background()
	totalVictims         int
	cfClient             *client.Client
	cfContext            = context.TODO()
)

const (
	RunTypeStopDaily                        = "stopDaily"
	RunTypeStopWeekly                       = "stopWeekly"
	RunTypeDailyAndWeekly                   = "stopDaily,stopWeekly"
	RunTypeStopCrashing                     = "stopCrashing"
	RunTypeStopOld                          = "stopOld"
	RunTypeDeleteStopped                    = "deleteStopped"
	RuntypeRestartWeeklyOutsideOfficeHours  = "RESTART_WEEKLY_OUTSIDE_OFFICE_HOURS"
	RuntypeRestartWeeklyOfficeHours         = "RESTART_WEEKLY_OFFICE_HOURS"
	RuntypeRestartMonthlyOutsideOfficeHours = "RESTART_MONTHLY_OUTSIDE_OFFICE_HOURS"
	RuntypeRestartMonthlyOfficeHours        = "RESTART_MONTHLY_OFFICE_HOURS"
	ProcessStateDown                        = "DOWN"
	ProcessStateCrashed                     = "CRASHED"
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
		runType = RunTypeStopWeekly
	} else if runType != RunTypeStopWeekly && runType != RunTypeStopDaily && runType != RunTypeDailyAndWeekly && runType != RunTypeStopCrashing && runType != RunTypeStopOld && runType != RunTypeDeleteStopped && runType != RuntypeRestartWeeklyOfficeHours && runType != RuntypeRestartWeeklyOutsideOfficeHours && runType != RuntypeRestartMonthlyOfficeHours && runType != RuntypeRestartMonthlyOutsideOfficeHours {
		log.Printf("invalid value (%s) for RUN_TYPE, must be one of %s, %s, %s, %s, %s, %s, %s, %s, %s, %s ", runType, RunTypeStopDaily, RunTypeStopWeekly, RunTypeDailyAndWeekly, RunTypeStopCrashing, RunTypeStopOld, RunTypeDeleteStopped, RuntypeRestartWeeklyOfficeHours, RuntypeRestartWeeklyOutsideOfficeHours, RuntypeRestartMonthlyOfficeHours, RuntypeRestartMonthlyOutsideOfficeHours)
		envComplete = false
	}

	if missingLabelAction == "" {
		missingLabelAction = RunTypeStopWeekly
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

	if runType == RunTypeStopOld || runType == RunTypeDeleteStopped || runType == RunTypeStopCrashing {
		if gracePeriod, err := strconv.Atoi(gracePeriodStr); err != nil {
			log.Fatalf("failed to parse grace period: %s", err)
		} else {
			graceDate = time.Now().Add(-time.Hour * 24 * time.Duration(gracePeriod))
		}
	}

	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" CF_API_ADDR: %s\n", apiAddress)
		fmt.Printf(" CF_USERNAME: %s\n", cfUsername)
		fmt.Printf(" SKIP_SSL_VALIDATION: %t\n", skipSSLValidation)
		fmt.Printf(" EXCLUDED_ORGS: %s\n", excludedOrgs)
		fmt.Printf(" EXCLUDED_SPACES: %s\n", excludedSpaces)
		fmt.Printf(" DRY_RUN: %s\n", dryRun)
		fmt.Printf(" RUN_TYPE: %s\n", runType)
		fmt.Printf(" MISSING_LABEL_ACTION: %s\n", missingLabelAction)
		if runType == RunTypeStopOld || runType == RunTypeDeleteStopped || runType == RunTypeStopCrashing {
			fmt.Printf(" GRACE_PERIOD: %s\n\n", graceDate.Format(time.RFC3339))
		}
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
	if strings.Contains(strings.ToLower(cfConfig.ApiURL("")), ".cfp") && runType == RunTypeStopOld {
		log.Println("skip stopping old apps because this is a production environment")
	} else {
		startTime := time.Now()
		if runType == RuntypeRestartWeeklyOfficeHours || runType == RuntypeRestartWeeklyOutsideOfficeHours || runType == RuntypeRestartMonthlyOfficeHours || runType == RuntypeRestartMonthlyOutsideOfficeHours {
			restartApps()
		} else {
			if orgs, err := cfClient.Organizations.ListAll(ctx, nil); err != nil {
				log.Printf("failed to list orgs: %s", err)
				os.Exit(1)
			} else {
				for _, org := range orgs {
					if !orgNameExcluded(org.Name) {
						if spaces, err := cfClient.Spaces.ListAll(ctx, &client.SpaceListOptions{ListOptions: &client.ListOptions{}, OrganizationGUIDs: client.Filter{Values: []string{org.GUID}}}); err != nil {
							log.Fatalf("failed to list spaces: %s", err)
						} else {
							for _, space := range spaces {
								if !spaceNameExcluded(space.Name) {
									if apps, err := cfClient.Applications.ListAll(ctx, &client.AppListOptions{ListOptions: &client.ListOptions{}, SpaceGUIDs: client.Filter{Values: []string{space.GUID}}}); err != nil {
										log.Fatalf("failed to list all apps: %s", err)
									} else {
										for _, app := range apps {
											if runType == RunTypeStopDaily || runType == RunTypeStopWeekly || runType == RunTypeDailyAndWeekly {
												dailyOrWeeklyStop(org, space, *app)
											}
											if runType == RunTypeStopCrashing {
												stopCrashing(org, space, *app)
											}
											if runType == RunTypeStopOld {
												stopOld(org, space, *app)
											}
											if runType == RunTypeDeleteStopped {
												deleteStopped(org, space, *app)
											}
										}
									}
								}
							}
						}
					}
				}
			}
			log.Printf("\nexecutionTime: %.0f secs, total victims: %d", time.Now().Sub(startTime).Seconds(), totalVictims)
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

func dailyOrWeeklyStop(org *resource.Organization, space *resource.Space, app resource.App) {
	autostopLabel := app.Metadata.Labels["AUTOSTOP"]
	//runType := strings.ToLower(runType)
	if app.State == "STARTED" && ((autostopLabel == nil && strings.Contains(runType, missingLabelAction)) || (autostopLabel != nil && strings.HasSuffix(runType, strings.ToLower(*autostopLabel)))) {
		if dryRun != "true" {
			if _, err := cfClient.Applications.Stop(ctx, app.GUID); err != nil {
				log.Printf("failed to stop app %s: %s", app.Name, err)
			} else {
				log.Printf("stopped  %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
				totalVictims++
			}
		} else {
			log.Printf("(because of DRYRUN=true) not stopped app %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
		}
	} else {
		//if app.State != "STARTED" {
		//	log.Printf("not stopping app %s because it is not started", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
		//} else if autostopLabel == nil {
		//	log.Printf("not stopping app %s because it has no AUTOSTOP label and missing label action is %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name), missingLabelAction)
		//} else {
		//	log.Printf("not stopping app %s because its AUTOSTOP label is %s and run type is %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name), *autostopLabel, runType)
		//}
	}
}

func stopCrashing(org *resource.Organization, space *resource.Space, app resource.App) {
	if app.State == "STARTED" {
		if processStats, err := cfClient.Processes.GetStatsForApp(cfContext, app.GUID, "web"); err != nil {
			log.Printf("failed to get process stats for app %s: %s", app.Name, err)
		} else {
			for _, processStat := range processStats.Stats {
				if processStat.State == ProcessStateDown || processStat.State == ProcessStateCrashed {
					if dryRun != "true" {
						if stoppedApp, err := cfClient.Applications.Stop(cfContext, app.GUID); err != nil {
							if stoppedApp != nil {
								log.Printf("failed to stop app %s: %s", stoppedApp.Name, err)
							}
						} else {
							log.Printf("stopped crashing app %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
							totalVictims++
						}
					} else {
						log.Printf("(because of DRYRUN=true) not stopped crashing app %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
					}
				}
			}
		}
	}
}

func stopOld(org *resource.Organization, space *resource.Space, app resource.App) {
	if app.State == "STARTED" {
		if app.UpdatedAt.Before(graceDate) {
			if dryRun != "true" {
				_, err := cfClient.Applications.Stop(cfContext, app.GUID)
				if err != nil {
					log.Printf("failed to stop app %s: %s", app.Name, err)
				} else {
					log.Printf("stopped old app %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
					totalVictims++
				}

			} else {
				log.Printf("(because of DRYRUN=true) not stopped old app %s", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name))
			}
		}
	}
}

func deleteStopped(org *resource.Organization, space *resource.Space, app resource.App) {
	if app.State == "STOPPED" {
		if app.UpdatedAt.Before(graceDate) {
			if dryRun != "true" {
				if _, err := cfClient.Applications.Delete(cfContext, app.GUID); err != nil {
					log.Printf("failed to delete app %s: %s", app.Name, err)
				} else {
					log.Printf("deleted stopped app %s (last update: %s)", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name), app.UpdatedAt.Format(time.RFC3339))
					totalVictims++
				}
			} else {
				log.Printf("(because of DRYRUN=true) not deleted stopped app %s (last update: %s)", fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name), app.UpdatedAt.Format(time.RFC3339))
			}
		}
	}
}

func restartApps() {
	labelSelector := make(client.LabelSelector)
	if runType == RuntypeRestartWeeklyOfficeHours {
		labelSelector.Existence(RuntypeRestartWeeklyOfficeHours)
	}
	if runType == RuntypeRestartWeeklyOutsideOfficeHours {
		labelSelector.Existence(RuntypeRestartWeeklyOutsideOfficeHours)
	}
	if runType == RuntypeRestartMonthlyOfficeHours {
		labelSelector.Existence(RuntypeRestartMonthlyOfficeHours)
	}
	if runType == RuntypeRestartMonthlyOutsideOfficeHours {
		labelSelector.Existence(RuntypeRestartMonthlyOutsideOfficeHours)
	}

	if apps, err := cfClient.Applications.ListAll(ctx, &client.AppListOptions{ListOptions: &client.ListOptions{LabelSel: labelSelector}}); err != nil {
		log.Fatalf("failed to list all apps for selector %v: %s", labelSelector, err)
	} else {
		for _, app := range apps {
			if app.State == "STARTED" {
				if dryRun != "true" {
					_, err = cfClient.Applications.Restart(cfContext, app.GUID)
					if err != nil {
						log.Printf("failed to restart app %s: %s", app.Name, err)
					} else {
						log.Printf("restarted app %s", fmt.Sprintf("%s", app.Name))
						totalVictims++
					}
				} else {
					log.Printf("(because of DRYRUN=true) not restarted app %s", fmt.Sprintf("%s", app.Name))
				}
			}
		}
	}
}
