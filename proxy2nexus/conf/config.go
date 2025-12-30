package conf

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	DefaultHttpPort = 4711
)

var (
	HttpPort      int
	err           error
	ForwardTo     string
	ProxyUser     string
	ProxyPassword string
)

func InitConfig() {
	HttpPort = DefaultHttpPort
	var portStr string
	portStr = os.Getenv("HTTP_PORT")
	if len(portStr) != 0 {
		HttpPort, err = strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("failed to parse HTTP_PORT environment variable %s: %s", portStr, err)
		}
	}
	ForwardTo = os.Getenv("FORWARD_TO_HOST")
	if len(ForwardTo) == 0 {
		log.Fatalf("FORWARD_TO_HOST environment variable is not set")
	} else {
		if strings.HasPrefix(ForwardTo, "http") {
			log.Fatalf("FORWARD_TO_HOST environment variable %s should not contain http/https prefix", ForwardTo)
		}
	}
	ProxyUser = os.Getenv("PROXY_USER")
	ProxyPassword = os.Getenv("PROXY_PASSWORD")
}
