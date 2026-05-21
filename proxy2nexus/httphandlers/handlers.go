package httphandlers

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rabobank/proxy2nexus/conf"
)

func HandleRequest(req *http.Request) {
	req.URL.Scheme = "https"
	req.URL.Host = conf.ForwardTo
	req.Host = conf.ForwardTo
	if len(conf.ProxyUser) > 0 && len(conf.ProxyPassword) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", conf.ProxyUser, conf.ProxyPassword)))))
	}
}

type MyRoundTripper struct {
	transport http.RoundTripper
}

func NewRoundTripper() *MyRoundTripper {
	transport := &http.Transport{MaxIdleConnsPerHost: 10, IdleConnTimeout: 3 * time.Second, TLSHandshakeTimeout: 5 * time.Second, MaxConnsPerHost: 100, MaxIdleConns: 5} //TLSClientConfig:     &tls.Config{InsecureSkipVerify: false, MinVersion: tls.VersionTLS12}
	return &MyRoundTripper{transport: transport}
}

func (lrt *MyRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	var err error
	var response *http.Response
	response, err = lrt.transport.RoundTrip(request)
	if err != nil {
		log.Printf("failed to proxy request %s: %s", request.URL.String(), err)
		return nil, err
	}
	conf.RequestCounter++
	log.Printf("%5d - %d : %s %s (%s)", conf.RequestCounter, response.StatusCode, request.Method, strings.Split(request.URL.String(), "?")[0], request.Host)
	return response, err
}
