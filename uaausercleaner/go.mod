module github.com/rabobank/go-utils/uaausercleaner

go 1.25

replace (
	golang.org/x/net => golang.org/x/net v0.46.0
	golang.org/x/text => golang.org/x/text v0.30.0
)

require github.com/cloudfoundry-community/go-uaa v0.3.5

require (
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
