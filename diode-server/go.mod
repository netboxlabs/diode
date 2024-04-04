module github.com/netboxlabs/diode/diode-server

go 1.21

require (
	github.com/envoyproxy/protoc-gen-validate v1.0.4
	github.com/evanphx/json-patch/v5 v5.9.0
	github.com/google/uuid v1.6.0
	github.com/gosimple/slug v1.14.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/netboxlabs/diode/diode-sdk-go v0.0.0
	github.com/oklog/run v1.1.0
	github.com/redis/go-redis/v9 v9.5.1
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.63.0
	google.golang.org/protobuf v1.33.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240401170217-c3f982113cda // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/netboxlabs/diode/diode-sdk-go v0.0.0 => ../diode-sdk-go
