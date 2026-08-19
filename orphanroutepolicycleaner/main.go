package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
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
	debugStr             = os.Getenv("DEBUG")
	debug                bool
	cfConfig             *config.Config
	ctx                  = context.Background()
)

// debugPrintf prints the given message only when DEBUG=true
func debugPrintf(format string, args ...any) {
	if debug {
		fmt.Printf(format, args...)
	}
}

// sourceExistence caches the existence check per source string (cf:<type>:<guid>)
var sourceExistence = make(map[string]bool)

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
			fmt.Printf("invalid value (%s) for SKIP_SSL_VALIDATION: %s\n", skipSSLValidationStr, err)
			envComplete = false
		}
	}
	if debugStr == "" {
		debug = false
	} else {
		if debug, err = strconv.ParseBool(debugStr); err != nil {
			fmt.Printf("invalid value (%s) for DEBUG: %s\n", debugStr, err)
			envComplete = false
		}
	}
	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" CF_API_ADDR: %s\n", apiAddress)
		fmt.Printf(" CF_USERNAME: %s\n", cfUsername)
		fmt.Printf(" SKIP_SSL_VALIDATION: %t\n", skipSSLValidation)
		fmt.Printf(" DRY_RUN: %s\n", dryRun)
		fmt.Printf(" DEBUG: %t\n\n", debug)
	}
	return envComplete
}

func getCFClient() (cfClient *client.Client) {
	var err error
	options := []config.Option{config.ClientCredentials(cfUsername, cfPassword)}
	if skipSSLValidation {
		options = append(options, config.SkipTLSValidation())
	}
	if cfConfig, err = config.New(apiAddress, options...); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	}
	if cfClient, err = client.New(cfConfig); err != nil {
		log.Fatalf("failed to create new client: %s", err)
	}
	return
}

func main() {
	startTime := time.Now()
	if !environmentComplete() {
		os.Exit(8)
	}
	cfClient := getCFClient()

	fmt.Println("listing all route policies...")
	routePolicies, err := cfClient.RoutePolicies.ListAll(ctx, nil)
	if err != nil {
		log.Fatalf("failed to list route policies: %s", err)
	}
	fmt.Printf("found %d route policies\n\n", len(routePolicies))

	// step 1 : gather and deduplicate all sources
	uniqueSources := make(map[string]int) // source -> number of route policies using it
	for _, routePolicy := range routePolicies {
		uniqueSources[routePolicy.Source]++
	}
	sources := make([]string, 0, len(uniqueSources))
	for source := range uniqueSources {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	fmt.Printf("found %d unique sources in %d route policies:\n", len(sources), len(routePolicies))
	for _, source := range sources {
		debugPrintf("  %s (used by %d route policies)\n", source, uniqueSources[source])
	}

	if len(sources) == 0 {
		fmt.Println("no route policies found, exiting")
		return
	}

	// step 2 : check for each unique source if the org/space/app still exists
	fmt.Printf("\nchecking existence of %d unique sources...\n", len(sources))
	orphanSources := 0
	for _, source := range sources {
		exists := sourceStillExists(cfClient, source)
		sourceExistence[source] = exists
		if exists {
			debugPrintf("  %-60s : still exists\n", source)
		} else {
			orphanSources++
			fmt.Printf("  %-60s : ORPHAN (no longer exists)\n", source)
		}
	}
	fmt.Printf("\n%d of the %d unique sources are orphans\n", orphanSources, len(sources))

	deleted := 0
	failed := 0
	skipped := 0

	if orphanSources == 0 {
		fmt.Println("no orphan sources found, exiting")
	} else {

		// step 3 : delete the route policies that have an orphan source
		fmt.Printf("\ndeleting route policies with an orphan source...\n")
		for _, routePolicy := range routePolicies {
			if sourceExistence[routePolicy.Source] {
				continue // source still exists, keep the route policy
			}
			routeGUID := ""
			if routePolicy.Relationships.Route != nil && routePolicy.Relationships.Route.Data != nil {
				routeGUID = routePolicy.Relationships.Route.Data.GUID
			}
			if dryRun == "true" {
				skipped++
				fmt.Printf("  DRY_RUN: would delete route policy %s (source: %s, route: %s)\n", routePolicy.GUID, routePolicy.Source, routeGUID)
				continue
			}
			fmt.Printf("  deleting route policy %s (source: %s, route: %s)... ", routePolicy.GUID, routePolicy.Source, routeGUID)
			_, err := cfClient.RoutePolicies.Delete(ctx, routePolicy.GUID)
			if err != nil {
				failed++
				fmt.Printf("deleting route policy FAILED: %s\n", err)
				continue
			}
			deleted++
			fmt.Printf("route policy deleted\n")
		}
	}

	fmt.Printf("\nsummary: route policies: %d, unique sources: %d, orphan sources: %d, deleted: %d, failed: %d, dry-run skipped: %d, execution time: %.0f secs\n",
		len(routePolicies), len(sources), orphanSources, deleted, failed, skipped, time.Since(startTime).Seconds())
}

// sourceStillExists checks if the org/space/app referenced by the given source (cf:<type>:<guid>) still exists.
// When we cannot determine it (unexpected error or unknown source format) we return true, so that we never delete
// a route policy we are not sure about.
func sourceStillExists(cfClient *client.Client, source string) bool {
	parts := strings.Split(source, ":")
	if len(parts) != 3 || parts[0] != "cf" {
		fmt.Printf("  %-60s : unrecognized source format, keeping it\n", source)
		return true
	}
	sourceType := parts[1]
	guid := parts[2]
	var err error
	switch sourceType {
	case "org":
		_, err = cfClient.Organizations.Get(ctx, guid)
	case "space":
		_, err = cfClient.Spaces.Get(ctx, guid)
	case "app":
		_, err = cfClient.Applications.Get(ctx, guid)
	default:
		fmt.Printf("  %-60s : unknown source type %s, keeping it\n", source, sourceType)
		return true
	}
	if err == nil {
		return true
	}
	if resource.IsResourceNotFoundError(err) || resource.IsNotFoundError(err) {
		return false
	}
	fmt.Printf("  %-60s : failed to check existence (%s), keeping it\n", source, err)
	return true
}
