package main

import (
	"fmt"
	"github.com/rabobank/proxy2nexus/conf"
	"github.com/rabobank/proxy2nexus/httphandlers"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

func main() {

	conf.InitConfig()

	proxy := &httputil.ReverseProxy{Director: httphandlers.HandleRequest, Transport: httphandlers.NewRoundTripper()}
	httpServer := &http.Server{Addr: fmt.Sprintf("localhost:%d", conf.HttpPort), Handler: proxy, IdleTimeout: time.Second * 2}
	log.Printf("starting http server on port %d, forwarding to %s", conf.HttpPort, conf.ForwardTo)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("failed to start http server on port %d: %s", conf.HttpPort, err)
	}
}
