module github.com/rabobank/go-utils/uaausercleaner

go 1.26

replace (
	golang.org/x/net => golang.org/x/net v0.56.0
	golang.org/x/text => golang.org/x/text v0.38.0
)

require github.com/cloudfoundry-community/go-uaa v0.4.0

require (
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
