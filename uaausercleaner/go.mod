module github.com/rabobank/go-utils/uaausercleaner

go 1.26.0

replace (
	golang.org/x/net => golang.org/x/net v0.59.0
	golang.org/x/text => golang.org/x/text v0.42.0
)

require github.com/cloudfoundry-community/go-uaa v0.5.0

require (
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
)
