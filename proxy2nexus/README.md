### proxy2nexus

A simple HTTP proxy server that forwards requests to a Nexus repository manager while optionally insert an Authorisation request header.

## Environment Variables
- `HTTP_PORT`: The port on which the proxy server will listen (default: 4711).
- `FORWARD_TO_HOST`: The hostname of the Nexus repository manager to which requests will be forwarded.
- `PROXY_USER`: The username for basic authentication when forwarding requests.
- `PROXY_PASSWORD`: The password for basic authentication when forwarding requests.

## Testing:

* Start the proxy using the above environment variables.
* cd to a test golang project directory, the one which we want to (test) build
* set GOMODCACHE to a temp directory: `export GOMODCACHE=/tmp/gomodcache`
* set GOPROXY to point to the proxy server: `export GOPROXY='http://localhost:4711/repository/golang/'` (include the correct path for your nexus golang repository)
* run `go build .` or `go mod download` to test if the proxy is working correctly.
* to repeat the test, make sure to clean the GOMODCACHE directory: `chmod -R 777 /tmp/gomodcache && rm -rf /tmp/gomodcache`
