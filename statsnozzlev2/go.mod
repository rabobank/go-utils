module github.com/rabobank/go-utils/statsnozzlev2

go 1.24.0

replace (
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.16.5
	golang.org/x/crypto => golang.org/x/crypto v0.39.0
	golang.org/x/net => golang.org/x/net v0.40.0
	golang.org/x/text => golang.org/x/text v0.25.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
)

require (
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e
	github.com/mattn/go-sqlite3 v1.14.28
)

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20250505082646-e4c2d772c2ec // indirect
	code.cloudfoundry.org/tlsconfig v0.30.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
