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
	"github.com/rabobank/go-utils/buildpackusage/db"
)

const defaultDBFile = "buildpackusage.db"

var (
	apiAddress = os.Getenv("CF_API_ADDR")
	cfUsername = os.Getenv("CF_USERNAME")
	cfPassword = os.Getenv("CF_PASSWORD")
	dbFile     = os.Getenv("DB_FILE")
	storeDBStr = os.Getenv("STORE_IN_DB")
	storeDB    = true
	ctx        = context.Background()
	orgCache   = map[string]*resource.Organization{}
	spaceCache = map[string]*resource.Space{}
)

type buildpackRow struct {
	OrgName          string
	SpaceName        string
	AppName          string
	DateCreated      time.Time
	DateLastUpdated  time.Time
	BuildpackName    string
	BuildpackVersion string
	Stack            string
}

type buildpackNameStack struct {
	Name  string
	Stack string
}

type buildpackNameVersionStack struct {
	Name    string
	Version string
	Stack   string
}

func environmentComplete() bool {
	envComplete := true
	if strings.TrimSpace(apiAddress) == "" {
		fmt.Println("missing envvar : CF_API_ADDR")
		envComplete = false
	}
	if strings.TrimSpace(cfUsername) == "" {
		fmt.Println("missing envvar : CF_USERNAME")
		envComplete = false
	}
	if strings.TrimSpace(cfPassword) == "" {
		fmt.Println("missing envvar : CF_PASSWORD")
		envComplete = false
	}
	if strings.TrimSpace(dbFile) == "" {
		dbFile = defaultDBFile
	}
	if strings.TrimSpace(storeDBStr) != "" {
		parsed, err := strconv.ParseBool(storeDBStr)
		if err != nil {
			fmt.Printf("invalid value (%s) for STORE_IN_DB: %s\n", storeDBStr, err)
			envComplete = false
		} else {
			storeDB = parsed
		}
	}

	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" CF_API_ADDR: %s\n", apiAddress)
		fmt.Printf(" CF_USERNAME: %s\n", cfUsername)
		fmt.Printf(" STORE_IN_DB: %t\n", storeDB)
		if storeDB {
			fmt.Printf(" DB_FILE: %s\n", dbFile)
		}
	}

	return envComplete
}

func getCFClient() *client.Client {
	cfConfig, err := config.New(apiAddress, config.ClientCredentials(cfUsername, cfPassword), config.SkipTLSValidation())
	if err != nil {
		log.Fatalf("failed to create new config: %s", err)
	}
	cfClient, err := client.New(cfConfig)
	if err != nil {
		log.Fatalf("failed to create new client: %s", err)
	}
	return cfClient
}

func main() {
	if !environmentComplete() {
		os.Exit(8)
	}

	cfClient := getCFClient()
	rows, byName, byNameStack, byNameVersionStack := collectBuildpackRows(cfClient)

	fmt.Printf("Buildpack droplets (total / unique versions): %d / %d\n\n", len(rows), len(byNameVersionStack))
	printNameCounts(byName)
	printNameStackCounts(byNameStack)
	printNameVersionStackCounts(byNameVersionStack)

	if storeDB {
		store, err := db.Open(dbFile)
		if err != nil {
			log.Fatalf("failed to open sqlite database %s: %s", dbFile, err)
		}
		defer func() {
			if err := store.Close(); err != nil {
				log.Printf("failed to close sqlite database: %s", err)
			}
		}()

		if err := store.InsertMany(ctx, toDBRows(rows)); err != nil {
			log.Fatalf("failed to insert buildpack rows into sqlite database: %s", err)
		}
	}
}

func collectBuildpackRows(cfClient *client.Client) ([]buildpackRow, map[string]int64, map[buildpackNameStack]int64, map[buildpackNameVersionStack]int64) {
	apps, err := cfClient.Applications.ListAll(ctx, &client.AppListOptions{ListOptions: &client.ListOptions{PerPage: 5000}})
	if err != nil {
		log.Fatalf("failed to list apps: %s", err)
	}
	droplets, err := cfClient.Droplets.ListAll(ctx, &client.DropletListOptions{ListOptions: &client.ListOptions{PerPage: 5000}})
	if err != nil {
		log.Fatalf("failed to list droplets: %s", err)
	}
	dropletByGUID := make(map[string]*resource.Droplet, len(droplets))
	for _, droplet := range droplets {
		dropletByGUID[droplet.GUID] = droplet
	}

	rows := make([]buildpackRow, 0, len(apps))
	byName := map[string]int64{}
	byNameStack := map[buildpackNameStack]int64{}
	byNameVersionStack := map[buildpackNameVersionStack]int64{}

	for _, app := range apps {
		currentDropletGUID := ""
		if app.Relationships.CurrentDroplet.Data != nil {
			currentDropletGUID = app.Relationships.CurrentDroplet.Data.GUID
		}
		if currentDropletGUID == "" {
			continue
		}

		droplet, ok := dropletByGUID[currentDropletGUID]
		if !ok {
			log.Printf("failed to find droplet %s for app %s", currentDropletGUID, app.Name)
			continue
		}
		if len(droplet.Buildpacks) == 0 {
			continue
		}

		space, org, err := getSpaceAndOrg(cfClient, app)
		if err != nil {
			log.Printf("failed to resolve org/space for app %s: %s", app.Name, err)
			continue
		}

		for _, detectedBuildpack := range droplet.Buildpacks {
			buildpackName := detectedBuildpack.BuildpackName
			if strings.TrimSpace(buildpackName) == "" {
				buildpackName = detectedBuildpack.Name
			}
			row := buildpackRow{
				OrgName:          org.Name,
				SpaceName:        space.Name,
				AppName:          app.Name,
				DateCreated:      droplet.CreatedAt,
				DateLastUpdated:  droplet.UpdatedAt,
				BuildpackName:    buildpackName,
				BuildpackVersion: detectedBuildpack.Version,
				Stack:            droplet.Stack,
			}
			rows = append(rows, row)
			byName[buildpackName]++
			byNameStack[buildpackNameStack{Name: buildpackName, Stack: droplet.Stack}]++
			byNameVersionStack[buildpackNameVersionStack{Name: buildpackName, Version: detectedBuildpack.Version, Stack: droplet.Stack}]++
		}
	}

	return rows, byName, byNameStack, byNameVersionStack
}

func getSpaceAndOrg(cfClient *client.Client, app *resource.App) (*resource.Space, *resource.Organization, error) {
	if app.Relationships.Space.Data == nil {
		return nil, nil, fmt.Errorf("app %s has no space relationship", app.Name)
	}
	space, err := getSpaceByGuidCached(cfClient, app.Relationships.Space.Data.GUID)
	if err != nil {
		return nil, nil, err
	}
	if space.Relationships == nil || space.Relationships.Organization == nil || space.Relationships.Organization.Data == nil {
		return nil, nil, fmt.Errorf("space %s has no organization relationship", space.Name)
	}
	org, err := getOrgByGuidCached(cfClient, space.Relationships.Organization.Data.GUID)
	if err != nil {
		return nil, nil, err
	}
	return space, org, nil
}

func getOrgByGuidCached(cfClient *client.Client, orgGUID string) (*resource.Organization, error) {
	if org, ok := orgCache[orgGUID]; ok {
		return org, nil
	}
	org, err := cfClient.Organizations.Get(ctx, orgGUID)
	if err != nil {
		return nil, err
	}
	orgCache[orgGUID] = org
	return org, nil
}

func getSpaceByGuidCached(cfClient *client.Client, spaceGUID string) (*resource.Space, error) {
	if space, ok := spaceCache[spaceGUID]; ok {
		return space, nil
	}
	space, err := cfClient.Spaces.Get(ctx, spaceGUID)
	if err != nil {
		return nil, err
	}
	spaceCache[spaceGUID] = space
	return space, nil
}

func toDBRows(rows []buildpackRow) []db.DropletBuildpack {
	dbRows := make([]db.DropletBuildpack, 0, len(rows))
	for _, row := range rows {
		dbRows = append(dbRows, db.DropletBuildpack{
			OrgName:          row.OrgName,
			SpaceName:        row.SpaceName,
			AppName:          row.AppName,
			DateCreated:      row.DateCreated,
			DateLastUpdated:  row.DateLastUpdated,
			BuildpackName:    row.BuildpackName,
			BuildpackVersion: row.BuildpackVersion,
			Stack:            row.Stack,
		})
	}
	return dbRows
}

func printNameCounts(counts map[string]int64) {
	fmt.Println("=== Counts by Buildpack Name ===")
	for _, name := range sortedStringKeys(counts) {
		fmt.Printf("%s: %d\n", name, counts[name])
	}
	fmt.Println()
}

func printNameStackCounts(counts map[buildpackNameStack]int64) {
	fmt.Println("=== Counts by Buildpack Name and stack ===")
	for _, key := range sortedNameStackKeys(counts) {
		fmt.Printf("%s/%s: %d\n", key.Name, key.Stack, counts[key])
	}
	fmt.Println()
}

func printNameVersionStackCounts(counts map[buildpackNameVersionStack]int64) {
	fmt.Println("=== Counts by Buildpack Name/Version ===")
	for _, key := range sortedNameVersionStackKeys(counts) {
		fmt.Printf("%s/%s/%s: %d\n", key.Name, key.Version, key.Stack, counts[key])
	}
	fmt.Println()
}

func sortedStringKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedNameStackKeys(m map[buildpackNameStack]int64) []buildpackNameStack {
	keys := make([]buildpackNameStack, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Name != keys[j].Name {
			return keys[i].Name < keys[j].Name
		}
		return keys[i].Stack < keys[j].Stack
	})
	return keys
}

func sortedNameVersionStackKeys(m map[buildpackNameVersionStack]int64) []buildpackNameVersionStack {
	keys := make([]buildpackNameVersionStack, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Name != keys[j].Name {
			return keys[i].Name < keys[j].Name
		}
		if keys[i].Version != keys[j].Version {
			return keys[i].Version < keys[j].Version
		}
		return keys[i].Stack < keys[j].Stack
	})
	return keys
}
