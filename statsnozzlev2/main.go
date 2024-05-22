package main

import (
	"context"
	"crypto/tls"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/rabobank/go-utils/statsnozzlev2/conf"
	"github.com/rabobank/go-utils/statsnozzlev2/db"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"code.cloudfoundry.org/go-loggregator/v9"
	"code.cloudfoundry.org/go-loggregator/v9/rpc/loggregator_v2"
)

var (
	allSelectors = []*loggregator_v2.Selector{{Message: &loggregator_v2.Selector_Log{Log: &loggregator_v2.LogSelector{}}}}
	accessToken  string

	// taken from:  https://clavinjune.dev/en/blogs/create-log-parser-using-go/
	logFormat = `$route - \[$timestamp\] \"$method $path $protocol\" $response_code $body_size $response_time \"$referer\" \"$user_agent\" \"$remote_addr\" \"$upstream_addr\" x_forwarded_for:\"$x_forwarded_for\" x_forwarded_proto:\"$x_forwarded_proto\" vcap_request_id:\"$vcap_request_id\" response_time:$response_time gorouter_time:$gorouter_time app_id:\"$app_id\" app_index:\"$app_index\" instance_id:\"$instance_id\" x_cf_routererror:\"$x_cf_routererror\" x_rabo_client_ip:\"$x_rabo_client_ip\" x_session_id:\"$x_session_id\" traceparent:\"$traceparent\" tracestate:\"$tracestate\"`
	// transform all the defined variable into a regex-readable named format
	logRegexFormat = regexp.MustCompile(`\$([\w_]*)`).ReplaceAllString(logFormat, `(?P<$1>.*)`)
	// compile the result
	logRegex = regexp.MustCompile(logRegexFormat)
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetPrefix("")
	logErr := log.New(os.Stderr, "", 0)
	logErr.SetOutput(os.Stderr)
	logErr.SetPrefix("")

	if !conf.EnvironmentComplete() {
		os.Exit(8)
	}

	errorChan := make(chan error)

	go func() {
		for err := range errorChan {
			logErr.Printf("from errorChannel: %s\n", err.Error())
		}
	}()

	uaa, err := uaago.NewClient(strings.Replace(conf.ApiAddr, "api.sys", "uaa.sys", 1))
	if err != nil {
		logErr.Printf("error while getting uaaClient %s\n", err)
		os.Exit(1)
	}

	if conf.StoreDB {
		db.InitDB()
	}

	tokenAttacher := &TokenAttacher{}

	go func() {
		for {
			if accessToken, err = uaa.GetAuthToken(conf.Client, conf.Secret, true); err != nil {
				log.Fatalf("tokenRefresher failed : %s)", err)
			}
			tokenAttacher.refreshToken(accessToken)
			time.Sleep(15 * time.Minute)
		}
	}()

	c := loggregator.NewRLPGatewayClient(
		strings.Replace(conf.ApiAddr, "api.sys", "log-stream.sys", 1),
		//loggregator.WithRLPGatewayClientLogger(log.New(os.Stderr, "", log.LstdFlags)),
		loggregator.WithRLPGatewayHTTPClient(tokenAttacher),
		loggregator.WithRLPGatewayErrChan(errorChan),
	)

	time.Sleep(1 * time.Second) // wait for uaa token to be fetched
	envelopeStream := c.Stream(context.Background(), &loggregator_v2.EgressBatchRequest{ShardId: conf.ShardId, Selectors: allSelectors})

	var envelopeCount int
	var insertCount int
	payloadBatch := make([][]string, 0, 100)
	bufferedEnvelopes := make(chan *loggregator_v2.Envelope, conf.MaxMessages)
	uriRegexp := regexp.MustCompile(conf.URIFilter)
	go func() {
		for envelope := range bufferedEnvelopes {
			payloadSlice := parseLogLine(string(envelope.GetLog().Payload))
			if uriRegexp.MatchString(payloadSlice[0]) {
				envelopeCount++
				payloadSlice = append(payloadSlice, envelope.Tags["organization_name"], envelope.Tags["space_name"], envelope.Tags["app_name"])
				if conf.PrintLogs {
					log.Printf("%s - :    -  %v\n", time.Unix(0, envelope.Timestamp), payloadSlice)
				}
				if conf.StoreDB {
					payloadBatch = append(payloadBatch, payloadSlice)
					// batch insert every 100 rows:
					insertCount++
					if insertCount%100 == 0 {
						if err = db.InsertStats(payloadBatch); err != nil {
							logErr.Printf("failed to insert stats: %s", err)
						}
						payloadBatch = payloadBatch[:0]
					}
					if insertCount%(conf.MaxMessages/10) == 0 {
						log.Printf("processed %d envelopes, inserted %d rows, bufferedEnvelopes: %d\n", envelopeCount, insertCount, len(bufferedEnvelopes))
					}
				}
				if insertCount >= conf.MaxMessages {
					log.Printf("all rows (%d) inserted, exiting...\n", conf.MaxMessages)
					os.Exit(0)
				}
			}
		}
	}()

	for {
		if envelopeCount < conf.MaxMessages {
			for _, envelope := range envelopeStream() {
				if envelope.Tags["source_type"] == "RTR" {
					bufferedEnvelopes <- envelope
					if envelopeCount >= conf.MaxMessages {
						log.Printf("max envelopes reached: %d, inserted %d rows\n", conf.MaxMessages, insertCount)
						break
					}
				}
			}
		} else {
			if insertCount < conf.MaxMessages {
				time.Sleep(5 * time.Second)
			} else {
				break
			}
		}
	}
}

type TokenAttacher struct {
	token string
}

func (a *TokenAttacher) refreshToken(token string) {
	a.token = token
}

func (a *TokenAttacher) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", a.token)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return http.DefaultClient.Do(req)
}

// parseLogLine parses a log line into a slice of strings using the predefined regex
func parseLogLine(logLine string) []string {
	matches := logRegex.FindStringSubmatch(logLine)
	return matches[1:]
}
