package conf

import (
	"fmt"
	"os"
	"strconv"
)

const (
	CreateTablesFile = "resources/sql/create-tables.sql"
)

var (
	ApiAddr        = os.Getenv("API_ADDR")
	ShardId        = os.Getenv("SHARD_ID")
	Client         = os.Getenv("CLIENT_ID")
	Secret         = os.Getenv("CLIENT_SECRET")
	printLogsStr   = os.Getenv("PRINT_LOGS")
	PrintLogs      = false
	storeDBStr     = os.Getenv("STORE_IN_DB")
	StoreDB        = false
	DBFile         = os.Getenv("DB_FILE")
	URIFilter      = os.Getenv("URI_FILTER")
	maxMessagesStr = os.Getenv("MAX_MESSAGES")
	MaxMessages    int
)

func EnvironmentComplete() bool {
	envComplete := true
	if ApiAddr == "" {
		fmt.Println("missing envvar: API_ADDR")
		envComplete = false
	}
	if ShardId == "" {
		fmt.Println("missing envvar: SHARD_ID")
		envComplete = false
	}
	if Client == "" {
		fmt.Println("missing envvar: CLIENT_ID")
		envComplete = false
	}
	if Secret == "" {
		fmt.Println("missing envvar: CLIENT_SECRET")
		envComplete = false
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
	if URIFilter == "" {
		URIFilter = ".*"
	}
	if maxMessagesStr == "" {
		MaxMessages = 10000
	} else {
		var err error
		if MaxMessages, err = strconv.Atoi(maxMessagesStr); err != nil {
			fmt.Println("error parsing MAX_MESSAGES: ", err)
			envComplete = false
		}
	}
	return envComplete
}
