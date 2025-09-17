module github.com/rabobank/go-utils/statsnozzlev2

go 1.25

replace (
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.16.5
	golang.org/x/crypto => golang.org/x/crypto v0.42.0
	golang.org/x/net => golang.org/x/net v0.44.0
	golang.org/x/text => golang.org/x/text v0.29.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.36.9
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
)

require (
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e
	github.com/mattn/go-sqlite3 v1.14.32
)

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20250909124000-1dfc755f0d96 // indirect
	code.cloudfoundry.org/tlsconfig v0.35.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)
