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
	apiAddress                 = os.Getenv("CF_API_ADDR")
	cfUsername                 = os.Getenv("CF_USERNAME")
	cfPassword                 = os.Getenv("CF_PASSWORD")
	skipSSLValidationStr       = os.Getenv("SKIP_SSL_VALIDATION")
	skipSSLValidation          bool
	memScrapePercentageStr     = os.Getenv("MEM_SCRAPE_PERCENTAGE")
	memScrapePercentage        int
	diskScrapePercentageStr    = os.Getenv("DISK_SCRAPE_PERCENTAGE")
	diskScrapePercentage       int
	javaMinimumMemoryStr       = os.Getenv("JAVA_MINIMUM_MEM_MB")
	javaMinimumMemory          int
	minimumDiskStr             = os.Getenv("MINIMUM_DISK_MB")
	minimumDisk                int
	memUsageThresholdStr       = os.Getenv("MEM_USAGE_THRESHOLD_PERCENTAGE")
	memUsageThreshold          int
	diskUsageThresholdStr      = os.Getenv("DISK_USAGE_THRESHOLD_PERCENTAGE")
	diskUsageThreshold         int
	lastUpdatedAgeThresholdStr = os.Getenv("LAST_UPDATED_AGE_THRESHOLD")
	lastUpdatedAgeThreshold    float64
	dryRun                     = os.Getenv("DRY_RUN")
	excludedOrgsStr            = os.Getenv("EXCLUDED_ORGS")
	excludedOrgs               []string
	excludedSpacesStr          = os.Getenv("EXCLUDED_SPACES")
	excludedSpaces             []string
	cfConfig                   *config.Config
	ctx                        = context.Background()
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
	if memScrapePercentageStr == "" {
		memScrapePercentage = 20
	} else {
		if memScrapePercentage, err = strconv.Atoi(memScrapePercentageStr); err != nil {
			fmt.Printf("invalid value (%s) for MEM_SCRAPE_PERCENTAGE: %s", memScrapePercentageStr, err)
			envComplete = false
		}
		if memScrapePercentage < 1 || memScrapePercentage > 100 {
			fmt.Printf("invalid value (%s) for MEM_SCRAPE_PERCENTAGE: must be between 1 and 100", memScrapePercentageStr)
			envComplete = false
		}
	}
	if diskScrapePercentageStr == "" {
		diskScrapePercentage = 20
	} else {
		if diskScrapePercentage, err = strconv.Atoi(diskScrapePercentageStr); err != nil {
			fmt.Printf("invalid value (%s) for DISK_SCRAPE_PERCENTAGE: %s", diskScrapePercentageStr, err)
			envComplete = false
		}
		if diskScrapePercentage < 1 || diskScrapePercentage > 100 {
			fmt.Printf("invalid value (%s) for DISK_SCRAPE_PERCENTAGE: must be between 1 and 100", diskScrapePercentageStr)
			envComplete = false
		}
	}
	if memUsageThresholdStr == "" {
		memUsageThreshold = 20
	} else {
		if memUsageThreshold, err = strconv.Atoi(memUsageThresholdStr); err != nil {
			fmt.Printf("invalid value (%s) for MEM_USAGE_THRESHOLD_PERCENTAGE: %s", memUsageThresholdStr, err)
			envComplete = false
		}
		if memUsageThreshold < 1 || memUsageThreshold > 100 {
			fmt.Printf("invalid value (%s) for MEM_USAGE_THRESHOLD_PERCENTAGE: must be between 1 and 100", memUsageThresholdStr)
			envComplete = false
		}
	}
	if diskUsageThresholdStr == "" {
		diskUsageThreshold = 20
	} else {
		if diskUsageThreshold, err = strconv.Atoi(diskUsageThresholdStr); err != nil {
			fmt.Printf("invalid value (%s) for DISKUSAGE_THRESHOLD_PERCENTAGE: %s", diskUsageThresholdStr, err)
			envComplete = false
		}
		if diskUsageThreshold < 1 || diskUsageThreshold > 100 {
			fmt.Printf("invalid value (%s) for DISK_USAGE_THRESHOLD_PERCENTAGE: must be between 1 and 100", diskUsageThresholdStr)
			envComplete = false
		}
	}
	if lastUpdatedAgeThresholdStr == "" {
		lastUpdatedAgeThreshold = 5
	} else {
		if lastUpdatedAgeThreshold, err = strconv.ParseFloat(lastUpdatedAgeThresholdStr, 64); err != nil {
			fmt.Printf("invalid value (%s) for LAST_UPDATED_AGE_THRESHOLD: %s", lastUpdatedAgeThresholdStr, err)
			envComplete = false
		}
		if lastUpdatedAgeThreshold < 0 {
			fmt.Printf("invalid value (%s) for LAST_UPDATED_AGE_THRESHOLD: must be greater than 0", lastUpdatedAgeThresholdStr)
			envComplete = false
		}
	}
	if javaMinimumMemoryStr == "" {
		javaMinimumMemory = 768
	} else {
		if javaMinimumMemory, err = strconv.Atoi(javaMinimumMemoryStr); err != nil {
			fmt.Printf("invalid value (%s) for JAVA_MINIMUM_MEM_MB: %s", javaMinimumMemoryStr, err)
			envComplete = false
		}
	}
	if minimumDiskStr == "" {
		minimumDisk = 512
	} else {
		if minimumDisk, err = strconv.Atoi(minimumDiskStr); err != nil {
			fmt.Printf("invalid value (%s) for MINIMUM_DISK_MB: %s", minimumDiskStr, err)
			envComplete = false
		}
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
		fmt.Printf(" MEM_SCRAPE_PERCENTAGE: %d\n", memScrapePercentage)
		fmt.Printf(" DISK_SCRAPE_PERCENTAGE: %d\n", diskScrapePercentage)
		fmt.Printf(" MEM_USAGE_THRESHOLD_PERCENTAGE: %d\n", memUsageThreshold)
		fmt.Printf(" DISK_USAGE_THRESHOLD_PERCENTAGE: %d\n", diskUsageThreshold)
		fmt.Printf(" LAST_UPDATED_AGE_THRESHOLD: %2.f\n", lastUpdatedAgeThreshold)
		fmt.Printf(" JAVA_MINIMUM_MEM_MB: %d\n", javaMinimumMemory)
		fmt.Printf(" MINIMUM_DISK_MB: %d\n", minimumDisk)
		fmt.Printf(" EXCLUDED_ORGS: %s\n", excludedOrgs)
		fmt.Printf(" EXCLUDED_SPACES: %s\n", excludedSpaces)
		fmt.Printf(" DRY_RUN: %s\n\n", dryRun)
	}
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
	cfClient := getCFClient()
	if orgs, err := cfClient.Organizations.ListAll(ctx, nil); err != nil {
		fmt.Printf("failed to list orgs: %s", err)
		os.Exit(1)
	} else {
		var totalMemVictims, totalDiskVictims, totalMemAllocated, totalMemUsed, totalDiskAllocated, totalDiskUsed int
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
									optedOut := app.Metadata.Labels["NO_MEMNDISK_TRIM"]
									if app.State == "STARTED" && (optedOut == nil || *optedOut != "true") {
										usedBuildpack := getBuildPackForApp(app)
										createdAge := time.Now().Sub(app.CreatedAt).Hours() / 24
										updatedAge := time.Now().Sub(app.UpdatedAt).Hours() / 24
										if processes, err := cfClient.Processes.ListForAppAll(ctx, app.GUID, &client.ProcessListOptions{}); err != nil {
											log.Printf("failed to list all processes for app %s: %s", app.Name, err)
										} else {
											for _, process := range processes {
												if process.Type == "web" {
													if stats, err := cfClient.Processes.GetStats(ctx, process.GUID); err != nil {
														log.Printf("failed to get stats for process %s: %s", process.GUID, err)
													} else {
														highestStatMemory := 0 // we can have multiple instances, we pick the highest value
														highestStatDisk := 0   // we can have multiple instances, we pick the highest value
														crashedProcessFound := false
														for _, stat := range stats.Stats {
															if stat.State != "RUNNING" {
																crashedProcessFound = true
															} else {
																if stat.Usage.Memory > highestStatMemory { // we pick the highest value
																	highestStatMemory = stat.Usage.Memory
																}
																if stat.Usage.Disk > highestStatDisk { // we pick the highest value
																	highestStatDisk = stat.Usage.Disk
																}
															}
														}
														isVictim := false
														if !crashedProcessFound {
															usageMemMB := highestStatMemory / 1024 / 1024
															usageMemPercentMB := (usageMemMB * 100) / process.MemoryInMB
															usageDiskMB := highestStatDisk / 1024 / 1024
															usageDiskPercentMB := (usageDiskMB * 100) / process.DiskInMB
															newProcess := &resource.Process{}
															if usageMemPercentMB < memUsageThreshold && updatedAge > lastUpdatedAgeThreshold && (!strings.Contains(usedBuildpack, "java_") || (strings.Contains(usedBuildpack, "java_") && process.MemoryInMB > javaMinimumMemory)) {
																isVictim = true
																totalMemVictims++
																totalMemUsed = totalMemUsed + usageMemMB
																totalMemAllocated = totalMemAllocated + process.MemoryInMB
																osa := fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name)
																fmt.Printf("\n%25v  mem usage (used/alloc):%4d/%4d (%2d%%) created/lastupdate age: %3.0f / %3.0f - %s", usedBuildpack, usageMemMB, process.MemoryInMB, usageMemPercentMB, createdAge, updatedAge, osa)
																newMemory := process.MemoryInMB * (100 - memScrapePercentage) / 100
																if newMemory < javaMinimumMemory && strings.Contains(usedBuildpack, "java_") {
																	newMemory = javaMinimumMemory
																}
																if dryRun != "true" {
																	if newProcess, err = cfClient.Processes.Scale(ctx, process.GUID, &resource.ProcessScale{MemoryInMB: &newMemory}); err != nil { // scale down to 80% of the original value
																		fmt.Printf("\nfailed to mem scale down app %s: %s\n", app.Name, err)
																	}
																}
															}
															if usageDiskPercentMB < diskUsageThreshold && usageDiskMB > minimumDisk && updatedAge > lastUpdatedAgeThreshold {
																isVictim = true
																totalDiskVictims++
																totalDiskAllocated = totalDiskAllocated + process.DiskInMB
																totalDiskUsed = totalDiskUsed + usageDiskMB
																osa := fmt.Sprintf("%s/%s/%s", org.Name, space.Name, app.Name)
																fmt.Printf("\n%25v disk usage (used/alloc):%4d/%4d (%2d%%) created/lastupdate age: %3.0f / %3.0f - %s", usedBuildpack, usageDiskMB, process.DiskInMB, usageDiskPercentMB, createdAge, updatedAge, osa)
																newDisk := process.DiskInMB * (100 - memScrapePercentage) / 100
																if dryRun != "true" {
																	if newProcess, err = cfClient.Processes.Scale(ctx, process.GUID, &resource.ProcessScale{DiskInMB: &newDisk}); err != nil { // scale down to 80% of the original value
																		fmt.Printf("\nfailed to disk scale down app %s: %s\n", app.Name, err)
																	}
																}
															}
															if isVictim {
																if dryRun != "true" {
																	if _, err = cfClient.Applications.Restart(ctx, app.GUID); err != nil {
																		fmt.Printf("\nfailed to restart app %s: %s\n", app.Name, err)
																	} else {
																		time.Sleep(3 * time.Second) // the docs say that the restart is synchronous, but to me it looks like it is not
																		fmt.Printf("  ==>  %d MB Mem, ==> %d MB Disk", newProcess.MemoryInMB, newProcess.DiskInMB)
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
					}
				}
			}
		}
		fmt.Printf("\nexecutionTime: %.0f secs, Memory: victims:%d, used/allocated/freed: %d/%d/%d  /  Disk: victims:%d used/allocated/freed: %d/%d/%d\n", time.Now().Sub(startTime).Seconds(), totalMemVictims, totalMemUsed, totalMemAllocated, memScrapePercentage*totalMemAllocated/100, totalDiskVictims, totalDiskUsed, totalDiskAllocated, diskScrapePercentage*totalDiskAllocated/100)
	}
}

func getBuildPackForApp(app *resource.App) string {
	if len(app.Lifecycle.BuildpackData.Buildpacks) > 0 {
		return app.Lifecycle.BuildpackData.Buildpacks[0]
	}
	if droplets, _, err := getCFClient().Droplets.List(ctx, &client.DropletListOptions{AppGUIDs: client.Filter{Values: []string{app.GUID}}}); err == nil {
		if len(droplets) > 0 && droplets[0].Buildpacks != nil && len(droplets[0].Buildpacks) > 0 {
			return droplets[0].Buildpacks[0].Name
		}
	}
	return ""
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
