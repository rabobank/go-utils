package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type SpaceConfig struct {
	Org      string `yaml:"org"`
	Space    string `yaml:"space"`
	Metadata struct {
		Manager string `yaml:"manager"`
		Contact string `yaml:"contact"`
	} `yaml:"metadata"`
}

type AlertMail struct {
	To      string `json:"to"`
	Cc      string `json:"cc"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

var (
	thresholdDaysStr   = os.Getenv("THRESHOLD_DAYS")
	thresholdDays      int
	orgsSpacesRepoPath = os.Getenv("ORGS_SPACES_REPO_PATH")
	awsRegion          = os.Getenv("AWS_REGION")
	mailCC             = os.Getenv("MAIL_CC")
	rdsClient          *rds.Client
	ctx                = context.Background()
	filterNameEngine   = "engine"
	totalDBInstances   = 0
	totalSnapshots     = 0
	orgsSpaces         map[string]SpaceConfig
	totalSpaces        = 0
	alertMails         []AlertMail
)

func main() {
	if !environmentComplete() {
		os.Exit(1)
	}
	if awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion)); err != nil {
		log.Fatalf("error loading aws config: %s", err)
	} else {
		rdsClient = rds.New(rds.Options{Credentials: awsConfig.Credentials, Region: awsConfig.Region})
		if instancesOutput, err := rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{Filters: []types.Filter{{Name: &filterNameEngine, Values: []string{"postgres"}}}}); err != nil {
			log.Fatalf("error describing rds instances: %s", err)
		} else {
			for _, dbInstance := range instancesOutput.DBInstances {
				totalDBInstances++
				//fmt.Printf("DBInstance: %s\n", *dbInstance.DBInstanceIdentifier)
				if snapshotsOutput, err := rdsClient.DescribeDBSnapshots(ctx, &rds.DescribeDBSnapshotsInput{DBInstanceIdentifier: dbInstance.DBInstanceIdentifier}); err != nil {
					log.Printf("error describing snapshots for instance %s: %s", *dbInstance.DBInstanceIdentifier, err)
				} else {
					var mostRecentSnapshot *types.DBSnapshot
					for _, snapshot := range snapshotsOutput.DBSnapshots {
						totalSnapshots++
						if mostRecentSnapshot == nil || snapshot.SnapshotCreateTime.After(*mostRecentSnapshot.SnapshotCreateTime) {
							mostRecentSnapshot = &snapshot
						}
					}
					if mostRecentSnapshot.SnapshotCreateTime.Before(time.Now().Add(-time.Duration(thresholdDays)*24*time.Hour)) || *mostRecentSnapshot.Status != "available" {
						//fmt.Printf("  Most recent SnapshotIdentifier: %s, SnapshotCreateTime: %s is too old\n", mostRecentSnapshotID, mostRecentSnapshotTime.Format(time.RFC3339))
						createAlert(dbInstance)
					}
				}
			}
			ba, err := json.MarshalIndent(alertMails, "", "  ")
			if err != nil {
				log.Printf("failed to format alert mail: %s", err)
			} else {
				if ba != nil && string(ba) != "null" {
					fmt.Println(string(ba))
				}
			}
			log.Printf("Total DB Instances checked: %d, total snapshots checked: %d\n", totalDBInstances, totalSnapshots)
		}
	}
}

func createAlert(dbInstance types.DBInstance) {
	if orgsSpaces == nil {
		orgsSpaces = make(map[string]SpaceConfig)
		loadOrgsSpaces(orgsSpacesRepoPath)
	}
	var org, space, serviceInstance string
	tagNameOrg := "OrganizationName"
	tagNameSpace := "SpaceName"
	tagNameInstanceName := "ServiceInstanceName"
	for _, tag := range dbInstance.TagList {
		if *tag.Key == tagNameOrg {
			org = *tag.Value
		} else if *tag.Key == tagNameSpace {
			space = *tag.Value
		} else if *tag.Key == tagNameInstanceName {
			serviceInstance = *tag.Value
		}
	}
	if spaceConfig, found := orgsSpaces[fmt.Sprintf("%s/%s", org, space)]; found {
		log.Printf("  Alerting org: %s, space: %s contact: %s\n", org, space, spaceConfig.Metadata.Contact)
		subject := fmt.Sprintf("RDS Snapshot alert for org: %s, space: %s, DB: %s, Service Instance: %s", org, space, *dbInstance.DBInstanceIdentifier, serviceInstance)
		body := fmt.Sprintf("The most recent snapshot for RDS instance %s (org: %s, space: %s, service instance: %s) is older than %d days.\n", *dbInstance.DBInstanceIdentifier, org, space, serviceInstance, thresholdDays)
		body += fmt.Sprintf("Contact team Panzer for more details.\n")
		alertMail := AlertMail{To: spaceConfig.Metadata.Contact, Cc: mailCC, Subject: subject, Message: body}
		alertMails = append(alertMails, alertMail)
	} else {
		fmt.Printf("  No spaceConfig.yml found for org: %s, space: %s\n", org, space)
	}
}

func loadOrgsSpaces(path string) {
	if err := filepath.Walk(path, handlePath); err != nil {
		fmt.Println(err)
	}
	log.Printf("Total spaces: %d\n", totalSpaces)
}

func handlePath(fullPath string, info os.FileInfo, err error) error {
	if err == nil && info.Name() == "spaceConfig.yml" {
		var file *os.File
		if file, err = os.Open(fullPath); err != nil {
			return errors.New(fmt.Sprintf("%s could not be opened: %s\n", fullPath, err))
		} else {
			defer func() { _ = file.Close() }()
			decoder := yaml.NewDecoder(bufio.NewReader(file))
			//decoder.KnownFields(true)
			spaceConfig := SpaceConfig{}
			if err = decoder.Decode(&spaceConfig); err != nil {
				return errors.New(fmt.Sprintf("%s could not be parsed: %s\n", fullPath, err))
			} else {
				totalSpaces++
				orgsSpaces[fmt.Sprintf("%s/%s", spaceConfig.Org, spaceConfig.Space)] = spaceConfig
			}
		}
	} else {
		if err != nil {
			return errors.New(fmt.Sprintf("error accessing path %s: %s\n", fullPath, err))
		}
	}
	return nil
}

func environmentComplete() bool {
	var err error
	envComplete := true
	if thresholdDaysStr == "" {
		thresholdDays = 1
	} else {
		if thresholdDays, err = strconv.Atoi(thresholdDaysStr); err != nil {
			fmt.Printf("invalid value (%s) for THRESHOLD_DAYS: %s", thresholdDaysStr, err)
			envComplete = false
		}
	}
	if orgsSpacesRepoPath == "" {
		fmt.Println("missing envvar : ORGS_SPACES_REPO_PATH")
		envComplete = false
	}
	if awsRegion == "" {
		awsRegion = "eu-west-1"
	}
	if envComplete {
		log.Printf("Running with the following options:\n")
		log.Printf(" THRESHOLD_DAYS: %d\n", thresholdDays)
		log.Printf(" ORGS_SPACES_REPO_PATH:%s\n", orgsSpacesRepoPath)
		log.Printf(" AWS_REGION: %s\n", awsRegion)
	}
	return envComplete
}
