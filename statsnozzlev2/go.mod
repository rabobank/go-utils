module github.com/rabobank/go-utils/statsnozzlev2

go 1.23

replace (
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.16.5
	golang.org/x/crypto => golang.org/x/crypto v0.31.0
	golang.org/x/net => golang.org/x/net v0.33.0
	golang.org/x/text => golang.org/x/text v0.21.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.36.0
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
)

require (
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e
	github.com/mattn/go-sqlite3 v1.14.24
)

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20241223074059-7f8c1f03edeb // indirect
	code.cloudfoundry.org/tlsconfig v0.13.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241219192143-6b3ec007d9bb // indirect
	google.golang.org/grpc v1.69.2 // indirect
	google.golang.org/protobuf v1.36.0 // indirect
)
