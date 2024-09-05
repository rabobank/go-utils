package main

import (
	"context"
	"fmt"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/logger"
	"github.com/hazelcast/hazelcast-go-client/types"
	"os"
	"time"
)

var (
	clusterAddress = os.Getenv("HZ_CLUSTERADDR")
	clusterName    = os.Getenv("HZ_CLUSTERNAME")
	hzUser         = os.Getenv("HZ_USERNAME")
	hzPassword     = os.Getenv("HZ_PASSWORD")
	dryRunStr      = os.Getenv("DRY_RUN")
	dryRun         = true
	ctx            = context.TODO()
)

func environmentComplete() bool {
	envComplete := true
	if clusterAddress == "" {
		fmt.Println("missing envvar : HZ_CLUSTERADDR")
		envComplete = false
	}
	if clusterName == "" {
		fmt.Println("missing envvar : HZ_CLUSTERNAME")
		envComplete = false
	}
	if hzUser == "" {
		fmt.Println("missing envvar : HZ_USERNAME")
		envComplete = false
	}
	if hzPassword == "" {
		fmt.Println("missing envvar : HZ_PASSWORD")
		envComplete = false
	}
	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" HZ_CLUSTERADDR: %s\n", clusterAddress)
		fmt.Printf(" HZ_CLUSTERNAME: %s\n", clusterName)
		fmt.Printf(" HZ_USERNAME: %s\n", hzUser)
		fmt.Printf(" DRY_RUN: %s\n", dryRunStr)
	}
	return envComplete
}

func main() {
	if !environmentComplete() {
		os.Exit(8)
	} else {
		config := hazelcast.Config{}
		config.ClientName = "panzer-hzutil"
		config.Cluster.Network.SetAddresses(clusterAddress)
		config.Cluster.Name = clusterName
		config.Cluster.Security.Credentials.Username = hzUser
		config.Cluster.Security.Credentials.Password = hzPassword
		if dryRunStr == "false" {
			dryRun = false
		}
		config.Logger.Level = logger.WarnLevel
		//log.SetOutput(os.Stdout)

		var err error
		var client *hazelcast.Client
		if client, err = hazelcast.StartNewClientWithConfig(ctx, config); err != nil {
			fmt.Printf("failed to start new client: %s\n", err)
		} else {
			var distObjects []types.DistributedObjectInfo
			var findCount, destroyCount int
			if distObjects, err = client.GetDistributedObjectsInfo(ctx); err != nil {
				fmt.Printf("failed to get distributed objects: %s\n", err)
			} else {
				for _, distObject := range distObjects {
					var mappie *hazelcast.Map
					mappie, err = client.GetMap(ctx, distObject.Name)
					if err != nil {
						fmt.Printf("failed to get map: %s\n", err)
					} else {
						if mapSize, err := mappie.Size(ctx); err != nil {
							fmt.Printf("failed to get map size: %s\n", err)
						} else if mapSize == 0 {
							findCount++
							_, _ = fmt.Fprintf(os.Stderr, "%d - %s\n", findCount, mappie.Name())
							if !dryRun {
								if err = mappie.Destroy(ctx); err != nil {
									fmt.Printf("      failed to destroy : %s\n", err)
								} else {
									destroyCount++
								}
							}
						}
					}
					time.Sleep(time.Millisecond * 50)
				}
				fmt.Printf("found %d distributed objects @ %s (%d destroyed)\n", len(distObjects), config.Cluster.Name, destroyCount)
			}
		}
		if err = client.Shutdown(ctx); err != nil {
			fmt.Printf("failed to shutdown client: %s\n", err)
		}
	}
}
