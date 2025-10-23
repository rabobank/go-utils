package main

import (
	"fmt"
	"github.com/cloudfoundry-community/go-uaa"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	uaaApiURL            = os.Getenv("UAA_API_ADDR")
	uaaClientId          = os.Getenv("UAA_CLIENTID")
	uaaClientSecret      = os.Getenv("UAA_CLIENTSECRET")
	skipSSLValidationStr = os.Getenv("SKIP_SSL_VALIDATION")
	deleteUsersStr       = os.Getenv("DELETE_USERS")
	createdDaysAgoStr    = os.Getenv("CREATED_DAYS_AGO")
	createdDaysAgo       int
	lastLogonDaysAgoStr  = os.Getenv("LAST_LOGON_DAYS_AGO")
	lastLogonDaysAgo     int
	skipSSLValidation    bool
	deleteUsers          bool
	api                  *uaa.API
	magicCreatedTime     = "2006-01-02T15:04:05.000Z"
	//magicLastLogonTime   = "2006-01-02 15:04:05 +0100 CET"
)

func environmentComplete() bool {
	var err error
	envComplete := true
	if uaaApiURL == "" {
		fmt.Println("missing envvar : UAA_API_ADDR")
		envComplete = false
	}
	if uaaClientId == "" {
		fmt.Println("missing envvar : UAA_CLIENTID")
		envComplete = false
	}
	if uaaClientSecret == "" {
		fmt.Println("missing envvar : UAA_CLIENTSECRET")
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
	if createdDaysAgoStr == "" {
		createdDaysAgo = 400
	} else {
		if createdDaysAgo, err = strconv.Atoi(createdDaysAgoStr); err != nil {
			fmt.Printf("invalid value (%s) for CREATED_DAYS_AGO: %s", createdDaysAgoStr, err)
			envComplete = false
		}
	}
	if lastLogonDaysAgoStr == "" {
		lastLogonDaysAgo = 400
	} else {
		if lastLogonDaysAgo, err = strconv.Atoi(lastLogonDaysAgoStr); err != nil {
			fmt.Printf("invalid value (%s) for LAST_LOGON_DAYS_AGO: %s", lastLogonDaysAgoStr, err)
			envComplete = false
		}
	}
	if deleteUsersStr == "" {
		deleteUsers = false
	} else {
		if deleteUsers, err = strconv.ParseBool(deleteUsersStr); err != nil {
			fmt.Printf("invalid value (%s) for DELETE_USERS: %s", deleteUsersStr, err)
			envComplete = false
		}
	}
	if envComplete {
		fmt.Printf("Running with the following options:\n")
		fmt.Printf(" UAA_API_ADDR: %s\n", uaaApiURL)
		fmt.Printf(" UAA_CLIENTID: %s\n", uaaClientId)
		fmt.Printf(" SKIP_SSL_VALIDATION: %t\n", skipSSLValidation)
		fmt.Printf(" CREATED_DAYS_AGO: %d\n", createdDaysAgo)
		fmt.Printf(" LAST_LOGON_DAYS_AGO: %s\n", lastLogonDaysAgoStr)
		fmt.Printf(" DELETE_USERS: %t\n", deleteUsers)
	}
	return envComplete
}

func main() {
	if !environmentComplete() {
		os.Exit(8)
	}
	log.SetOutput(os.Stdout)
	if err := initializeUaa(); err != nil {
		fmt.Printf("failed to initialize UAA: %s\n", err)
		os.Exit(8)
	}
	log.Printf("UAA initialized, getting users...")
	pageSize := 250
	startIndex := 0
	eligibleUsers := 0
	totalUsers := 0
	deletedUsers := 0
	for {
		if users, page, err := api.ListUsers("", "", "", "", startIndex-deletedUsers, pageSize); err != nil {
			log.Printf("failed to list users: %s\n", err)
			//os.Exit(8)
		} else {
			//log.Printf("Total %d (startIndex %d)\n", page.TotalResults, page.StartIndex)
			if startIndex == 0 {
				totalUsers = page.TotalResults
			}
			startIndex += page.ItemsPerPage
			if startIndex > page.TotalResults {
				break
			}
			var createdTime time.Time
			for _, user := range users {
				if createdTime, err = time.Parse(magicCreatedTime, user.Meta.Created); err != nil {
					log.Printf("failed to parse created time: %s\n", err)
				} else {
					lastLogonTime := time.Unix(int64(user.LastLogonTime/1000), 0)
					if time.Since(createdTime).Hours() > float64(createdDaysAgo*24) && time.Since(lastLogonTime).Hours() > float64(lastLogonDaysAgo*24) {
						log.Printf("creationTime: %s, lastLogonTime: %s, ID: %s origin: %s, User: %s\n", createdTime.Format(time.RFC3339), lastLogonTime.Format(time.RFC3339), user.ID, user.Origin, user.Username)
						eligibleUsers++
						if deleteUsers {
							if deletedUser, err := api.DeleteUser(user.ID); err != nil {
								log.Printf("failed to delete user (%s) %s: %s\n", user.ID, user.Username, err)
							} else {
								log.Printf("deleted user (%s) %s\n", deletedUser.ID, deletedUser.Username)
								deletedUsers++
							}
						}
					}
				}
			}
		}
	}
	log.Printf("eligible users: %d (out of %d), deleted users: %d\n", eligibleUsers, totalUsers, deletedUsers)
}

func initializeUaa() error {
	if a, e := uaa.New(uaaApiURL, uaa.WithClientCredentials(uaaClientId, uaaClientSecret, uaa.JSONWebToken), uaa.WithSkipSSLValidation(true), uaa.WithVerbosity(false)); e != nil {
		return e
	} else {
		api = a
	}
	return nil
}
