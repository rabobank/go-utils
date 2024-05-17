package main

import (
	"crypto/tls"
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient/v3/config"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/rabobank/go-utils/statsnozzlev1/conf"
	"github.com/rabobank/go-utils/statsnozzlev1/db"
	"github.com/rabobank/go-utils/statsnozzlev1/model"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	lock         = sync.RWMutex{}
	eventTypes   = make(map[string]int)
	origins      = make(map[string]int)
	jobs         = make(map[string]int)
	ips          = make(map[string]int)
	RTRRemote    = make(map[string]int)
	RTRIP        = make(map[string]int)
	RTRForwarded = make(map[string]int)
	RTRUserAgent = make(map[string]int)
	RTRUri       = make(map[string]int)
	cfConfig     *config.Config
)

type tokenRefresher struct {
	uaaClient *uaago.Client
}

func (t *tokenRefresher) RefreshAuthToken() (string, error) {
	token, err := t.uaaClient.GetAuthToken(conf.CfUsername, conf.CfPassword, true)
	if err != nil {
		log.Fatalf("tokenRefresher failed : %s)", err)
	}
	return token, nil
}

func getCFConfig() {
	var err error
	if cfConfig, err = config.NewClientSecret(conf.ApiAddress, conf.CfUsername, conf.CfPassword); err != nil {
		log.Fatalf("failed to create new config: %s", err)
	} else {
		cfConfig.WithSkipTLSValidation(true)
	}
	return
}

func printStatistics() {
	lock.Lock()
	defer lock.Unlock()
	fmt.Print("\n=================================================================================================\nEventTypes\n")

	var arrays2sort = []map[string]int{eventTypes, origins, jobs, ips, RTRIP, RTRUri, RTRRemote, RTRForwarded, RTRUserAgent}
	var names = []string{"Event Types", "Event Origins", "Event Jobs", "Event IPs", "RTR IPs", "RTR URIs", "RTR Remotes", "RTR Forwarded for", "RTR UserAgent"}

	for ix, array2sort := range arrays2sort {
		vs := NewValSorter(array2sort, true)
		vs.Sort()
		fmt.Print("\n", names[ix], ":\n")
		var resultCount int64
		for ix2, key := range vs.Keys {
			resultCount++
			if resultCount > conf.StatsSize {
				break
			}
			fmt.Printf("  %s : %d\n", key, vs.Vals[ix2])
		}
	}
}

func main() {
	if !conf.EnvironmentComplete() {
		os.Exit(8)
	}

	log.SetPrefix("")
	logErr := log.New(os.Stderr, "", 0)
	logErr.SetOutput(os.Stderr)
	logErr.SetPrefix("")
	getCFConfig()
	//
	// start sucking the firehose and handle events
	//
	dopplerAddress := strings.Replace(conf.ApiAddress, "https://api.", "wss://doppler.", 1)
	cons := consumer.New(dopplerAddress, &tls.Config{InsecureSkipVerify: true}, nil)
	//cons.SetDebugPrinter(ConsoleDebugPrinter{})
	uaa, err := uaago.NewClient(strings.ReplaceAll(conf.ApiAddress, "api.sys", "uaa.sys"))
	if err != nil {
		log.Printf("error from uaaClient %s\n", err)
		os.Exit(1)
	}
	refresher := tokenRefresher{uaaClient: uaa}
	cons.RefreshTokenFrom(&refresher)
	firehoseChan, errorChan := cons.Firehose("StatsNozzle", cfConfig.AccessToken)

	go func() {
		for err := range errorChan {
			log.Printf("%v\n", err.Error())
		}
	}()

	if conf.PrintStats {
		go func() {
			for i := 0; i < 99999; i++ {
				printStatistics()
				time.Sleep(time.Duration(conf.StatsInterval) * time.Second)
			}
		}()
	}

	if conf.StoreDB {
		db.InitDB()
	}

	var msgCount int
	statsBatch := make([]model.Stats, 0, 100)
	uriRegexp := regexp.MustCompile(conf.URIFilter)
	for msg := range firehoseChan {
		msgCount++
		if msg.GetEventType() == events.Envelope_HttpStartStop {
			lock.Lock()
			if uriRegexp.MatchString(*msg.HttpStartStop.Uri) {
				RTRRemote[strings.Split(*msg.HttpStartStop.RemoteAddress, ":")[0]]++
				RTRIP[msg.GetIp()]++
				RTRForwarded[fmt.Sprintf("%s", msg.HttpStartStop.Forwarded)]++
				RTRUserAgent[*msg.HttpStartStop.UserAgent]++
				RTRUri[*msg.HttpStartStop.Uri]++
				stats := model.Stats{
					Time:          time.Unix(0, msg.GetTimestamp()),
					IP:            msg.GetIp(),
					PeerType:      msg.HttpStartStop.GetPeerType().String(),
					Method:        msg.HttpStartStop.GetMethod().String(),
					StatusCode:    int(msg.HttpStartStop.GetStatusCode()),
					ContentLength: int(msg.HttpStartStop.GetContentLength()),
					URI:           *msg.HttpStartStop.Uri,
					Remote:        strings.Split(*msg.HttpStartStop.RemoteAddress, ":")[0],
					RemotePort:    strings.Split(*msg.HttpStartStop.RemoteAddress, ":")[1],
					ForwardedFor:  fmt.Sprintf("%s", msg.HttpStartStop.Forwarded),
					UserAgent:     *msg.HttpStartStop.UserAgent,
				}
				if conf.PrintLogs {
					logErr.Println(stats)
				}
				if conf.StoreDB {
					statsBatch = append(statsBatch, stats)
					// batch insert every 100 rows:
					if msgCount%100 == 0 {
						if err := db.InsertStats(statsBatch); err != nil {
							log.Printf("failed to insert stats: %s", err)
						}
						statsBatch = statsBatch[:0]
					}
				}
			}
			lock.Unlock()
		}
		eventTypes[msg.GetEventType().String()]++
		origins[msg.GetOrigin()]++
		jobs[msg.GetJob()]++
		ips[msg.GetIp()]++
	}
}
