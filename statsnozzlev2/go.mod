module github.com/rabobank/go-utils/statsnozzlev2

go 1.26.0

replace (
	github.com/onsi/ginkgo => github.com/onsi/ginkgo v1.16.5
	golang.org/x/crypto => golang.org/x/crypto v0.57.0
	golang.org/x/net => golang.org/x/net v0.59.0
	golang.org/x/text => golang.org/x/text v0.42.0
	google.golang.org/protobuf => google.golang.org/protobuf v1.36.12
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.4.0
)

require (
	code.cloudfoundry.org/go-loggregator/v9 v9.2.1
	github.com/cloudfoundry-incubator/uaago v0.0.0-20190307164349-8136b7bbe76e
	github.com/mattn/go-sqlite3 v1.14.52
)

require (
	code.cloudfoundry.org/go-diodes v0.0.0-20260928063035-f81ac938b818 // indirect
	code.cloudfoundry.org/tlsconfig v0.68.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260928230214-8a89bd6388cc // indirect
	google.golang.org/grpc v1.86.0-dev // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
