module github.com/rabobank/go-utils/statsnozzlev2

go 1.23

replace (
	github.com/yuin/goldmark => github.com/yuin/goldmark v1.4.0
	golang.org/x/crypto => golang.org/x/crypto v0.28.0
	golang.org/x/net => golang.org/x/net v0.30.0
	golang.org/x/text => golang.org/x/text v0.19.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.35.1
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.16.5
)

require (
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e
	github.com/mattn/go-sqlite3 v1.14.24
)

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20241007161556-ec30366c7912 // indirect
	code.cloudfoundry.org/tlsconfig v0.7.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241021214115-324edc3d5d38 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)
