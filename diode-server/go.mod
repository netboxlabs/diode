module github.com/netboxlabs/diode-internal/diode-server

go 1.21

require (
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/netboxlabs/diode-internal/diode-sdk-go v0.0.0
	github.com/oklog/run v1.1.0
	github.com/redis/go-redis/v9 v9.4.0
	google.golang.org/grpc v1.61.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/envoyproxy/protoc-gen-validate v1.0.4 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

replace github.com/netboxlabs/diode-internal/diode-sdk-go v0.0.0 => ../diode-sdk-go
