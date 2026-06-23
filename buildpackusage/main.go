package main

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	apiAddress          = os.Getenv("CF_API_ADDR")
	cfUsername          = os.Getenv("CF_USERNAME")
	cfPassword          = os.Getenv("CF_PASSWORD")
	cfConfig            *config.Config
	ctx                 = context.Background()
	cfClient            *client.Client
	cfClientMu          sync.Mutex
	offeringNamesByGuid = make(map[string]string)
	orgCacheByGuid      = make(map[string]*resource.Organization)
	spaceCacheByGuid    = make(map[string]*resource.Space)
)

func environmentComplete() bool {
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

	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" CF_API_ADDR: %s\n", apiAddress)
		fmt.Printf(" CF_USERNAME: %s\n", cfUsername)
		getCFClient()
	}

	return envComplete
}

func getCFClient() {
	var err error
	if cfConfig, err = config.New(apiAddress, config.ClientCredentials(cfUsername, cfPassword), config.SkipTLSValidation()); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		if cfClient, err = client.New(cfConfig); err != nil {
			log.Fatalf("failed to create new client: %s", err)
		} else {
			// refresh the client every 90 minutes to get a new refresh token
			go func() {
				ticker := time.NewTicker(90 * time.Minute)
				defer ticker.Stop()
				for range ticker.C {
					newClient, err := client.New(cfConfig)
					if err != nil {
						log.Printf("failed to refresh cfclient, error is %s", err)
					} else {
						cfClientMu.Lock()
						cfClient = newClient
						cfClientMu.Unlock()
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

	buildpackVersions := make(map[string]int) // key is buildpack name and version, value is count of droplets using that buildpack

	dropletListOptions := client.DropletListOptions{ListOptions: &client.ListOptions{}}
	if droplets, err := cfClient.Droplets.ListAll(ctx, &dropletListOptions); err != nil {
		log.Fatalf("failed to list droplets: %s", err)
	} else {
		dropletCount := 0
		for _, droplet := range droplets {
			if droplet.Lifecycle.Type == "buildpack" {
				dropletCount++
				for _, buildpack := range droplet.Buildpacks {
					if strings.Contains(buildpack.Version, "-offline-") {
						buildpack.Version = strings.Split(buildpack.Version, "-offline-")[0]
					}
					if currentVersion := buildpackVersions[buildpack.Name+"/"+buildpack.Version]; currentVersion == 0 {
						buildpackVersions[buildpack.Name+"/"+buildpack.Version] = 1
					} else {
						buildpackVersions[buildpack.Name+"/"+buildpack.Version] = currentVersion + 1
					}
				}
			}

		}
		fmt.Printf("Total buildpack droplets: %d\n", dropletCount)
		// range over the buildpackVersions map and print out the buildpacks and their counts
		for buildpack, count := range buildpackVersions {
			fmt.Printf("%s: %d\n", buildpack, count)
		}
	}

}
