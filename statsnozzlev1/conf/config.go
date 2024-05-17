package conf

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	CreateTablesFile = "resources/sql/create-tables.sql"
)

var (
	ApiAddress       = os.Getenv("CF_API_ADDR")
	CfUsername       = os.Getenv("CF_USERNAME")
	CfPassword       = os.Getenv("CF_PASSWORD")
	statsSizeStr     = os.Getenv("STATS_SIZE")
	StatsSize        int64
	statsIntervalStr = os.Getenv("STATS_INTERVAL")
	StatsInterval    int64
	printStatsStr    = os.Getenv("PRINT_STATS")
	PrintStats       = false
	printLogsStr     = os.Getenv("PRINT_LOGS")
	PrintLogs        = false
	storeDBStr       = os.Getenv("STORE_IN_DB")
	StoreDB          = false
	DBFile           = os.Getenv("DB_FILE")
	URIFilter        = os.Getenv("URI_FILTER")
)

func EnvironmentComplete() bool {
	envComplete := true
	if ApiAddress == "" {
		fmt.Println("missing envvar : API_ADDR")
		envComplete = false
	}
	if CfUsername == "" {
		fmt.Println("missing envvar : CF_USERNAME")
		envComplete = false
	}
	if CfPassword == "" {
		fmt.Println("missing envvar : CF_PASSWORD")
		envComplete = false
	}
	if len(printStatsStr) == 0 || printStatsStr == "true" {
		PrintStats = true
	}
	if len(printLogsStr) == 0 || printLogsStr == "true" {
		PrintLogs = true
	}
	if len(storeDBStr) == 0 || storeDBStr == "true" {
		StoreDB = true
	}
	if DBFile == "" {
		fmt.Println("missing envvar : DB_FILE")
		envComplete = false
	}
	if statsSizeStr == "" {
		StatsSize = 25
	} else {
		StatsSize, _ = strconv.ParseInt(statsSizeStr, 10, 64)
	}
	if statsIntervalStr == "" {
		StatsInterval = 10
	} else {
		StatsInterval, _ = strconv.ParseInt(statsIntervalStr, 10, 64)
	}
	if URIFilter == "" {
		URIFilter = ".*"
	}
	log.SetOutput(os.Stdout)
	return envComplete
}
