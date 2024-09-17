package reconciler

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	mr "github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

// MockPipeliner is a mock implementation of the redis Pipeliner interface.
type MockPipeliner struct {
	mock.Mock
	redis.Pipeliner
}

// Do is a mock of Pipeliner's Do method.
func (m *MockPipeliner) Do(ctx context.Context, args ...interface{}) *redis.Cmd {
	calledArgs := m.Called(ctx, args)
	return calledArgs.Get(0).(*redis.Cmd)
}

// Exec is a mock of Pipeliner's Exec method.
func (m *MockPipeliner) Exec(ctx context.Context) ([]redis.Cmder, error) {
	args := m.Called(ctx)
	cmds := make([]redis.Cmder, 0)
	return cmds, args.Error(0)
}

func TestIsAuthenticated(t *testing.T) {
	tests := []struct {
		name            string
		rpcMethod       string
		authorization   []string
		apiKeys         map[string]string
		isAuthenticated bool
	}{
		{
			name:          "missing authorization header",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{},
			apiKeys: map[string]string{
				"INGESTER_TO_RECONCILER": "test",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion data sources with valid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"INGESTER_TO_RECONCILER": "test",
			},
			isAuthenticated: true,
		},
		{
			name:          "retrieve ingestion data sources with invalid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{"test0"},
			apiKeys: map[string]string{
				"INGESTER_TO_RECONCILER": "test",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion data sources for server without api key configured",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"DIODE": "foorbar",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion logs with valid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionLogs_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"NETBOX_TO_DIODE": "test",
			},
			isAuthenticated: true,
		},
		{
			name:          "retrieve ingestion logs with invalid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionLogs_FullMethodName,
			authorization: []string{"test0"},
			apiKeys: map[string]string{
				"NETBOX_TO_DIODE": "test",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion logs for server without api key configured",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionLogs_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"DIODE": "foorbar",
			},
			isAuthenticated: false,
		},
		{
			name:          "authorization for unknown rpc method",
			rpcMethod:     "/diode.v1.ReconcilerService/UnknownMethod",
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"DIODE": "foorbar",
			},
			isAuthenticated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			assert.Equal(t, tt.isAuthenticated, isAuthenticated(logger, tt.rpcMethod, tt.apiKeys, tt.authorization))
		})
	}
}

func TestRetrieveLogs(t *testing.T) {
	tests := []struct {
		name             string
		in               reconcilerpb.RetrieveIngestionLogsRequest
		result           interface{}
		response         *reconcilerpb.RetrieveIngestionLogsResponse
		queryFilter      string
		queryLimitOffset int32
		failCmd          bool
		hasError         bool
	}{
		{
			name: "valid request",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":2}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.device","entity":{"device":{"name":"Conference_Room_AP_02","deviceType":{"model":"Cisco Aironet 3802","manufacturer":{"name":"Cisco"}},"role":{"name":"Wireless_AP"},"serial":"PQR456789012","site":{"name":"HQ"}}},"id":"2mC8GVBGFg6NyLsQxuS4IYMB6FI","ingestionTs":1725552654541975975,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"bc1052e3-656a-42f0-b364-27b385e02a0c","sdkName":"diode-sdk-python","sdkVersion":"0.0.1","state":2}`,
							"ingestion_ts": "1725552654541976064",
						},
						"id":     "ingest-entity:dcim.device-1725552654541975975-2mC8GVBGFg6NyLsQxuS4IYMB6FI",
						"values": []interface{}{},
					},
				},
				"total_results": 2,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
					{
						DataType:           "dcim.device",
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "bc1052e3-656a-42f0-b364-27b385e02a0c",
						IngestionTs:        1725552654541975975,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-python",
						SdkVersion:         "0.0.1",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Device{
								Device: &diodepb.Device{
									Name: "Conference_Room_AP_02",
									DeviceType: &diodepb.DeviceType{
										Model: "Cisco Aironet 3802",
										Manufacturer: &diodepb.Manufacturer{
											Name: "Cisco",
										},
									},
									Role:   &diodepb.Role{Name: "Wireless_AP"},
									Serial: strPtr("PQR456789012"),
									Site:   &diodepb.Site{Name: "HQ"},
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					Total: 2,
				},
				NextPageToken: "F/Jk/zc08gA=",
			},
			queryFilter:      "*",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "request with reconciliation error",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"ipam.ipaddress","entity":{"ip_address":{"address":"192.168.1.1","interface":null,"description":"Vendor: HUAWEI TECHNOLOGIES"}},"error":{"message":"failed to apply change set","code":400,"details":{"change_set_id":"6304c706-f955-4bcb-a1cc-514293d53d07","result":"failed","errors":[{"error":"address: Duplicate IP address found in global table: 192.168.1.1/32","change_id":"ff9e29b2-7a64-40ba-99a8-21f44768f60a"}]}},"id":"2mC8KCvHNasrYlfxSASk9hatfYC","ingestionTs":1725046967777525928,"producerAppName":"example-app","producerAppVersion":"0.1.0","request_id":"e03c4892-5b7e-4c39-b5e6-0225a264ab8b","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":3}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:ipam.ipaddress-1725046967777525928-2mC8KCvHNasrYlfxSASk9hatfYC",
						"values": []interface{}{},
					},
				},
				"total_results": 2,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						DataType:           "ipam.ipaddress",
						State:              reconcilerpb.State_FAILED,
						RequestId:          "e03c4892-5b7e-4c39-b5e6-0225a264ab8b",
						IngestionTs:        1725046967777525928,
						ProducerAppName:    "example-app",
						ProducerAppVersion: "0.1.0",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_IpAddress{
								IpAddress: &diodepb.IPAddress{
									Address:     "192.168.1.1",
									Description: strPtr("Vendor: HUAWEI TECHNOLOGIES"),
								},
							},
						},
						Error: &reconcilerpb.IngestionError{
							Message: "failed to apply change set",
							Code:    400,
							Details: &reconcilerpb.IngestionError_Details{
								ChangeSetId: "6304c706-f955-4bcb-a1cc-514293d53d07",
								Result:      "failed",
								Errors: []*reconcilerpb.IngestionError_Details_Error{
									{
										ChangeId: "ff9e29b2-7a64-40ba-99a8-21f44768f60a",
										Error:    "address: Duplicate IP address found in global table: 192.168.1.1/32",
									},
								},
							},
						},
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					Total: 2,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "*",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "filter by new state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_NEW.Enum()},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mC8NYwfIKM5rFDibDBuytASSOi","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mC8NYwfIKM5rFDibDBuytASSOi",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_NEW,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					New: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "@state:[1 1]",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "filter by reconciled state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_RECONCILED.Enum()},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":2}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					Reconciled: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "@state:[2 2]",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "filter by failed state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_FAILED.Enum()},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":3}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_FAILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					Failed: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "@state:[3 3]",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "filter by no changes state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_NO_CHANGES.Enum()},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":4}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_NO_CHANGES,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					NoChanges: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "@state:[4 4]",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "filter by data type",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{DataType: "dcim.interface"},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":2}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					Total: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "@data_type:dcim.interface",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "filter by timestamp",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{IngestionTsStart: 1725552914392208639},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":2}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					Total: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "@ingestion_ts:[1725552914392208639 inf]",
			queryLimitOffset: 0,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "pagination check",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "AAAFlg=="},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":2}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						DataType:           "dcim.interface",
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity: &diodepb.Entity{
							Entity: &diodepb.Entity_Interface{
								Interface: &diodepb.Interface{
									Device: &diodepb.Device{
										Name: "my_dev",
									},
									Name: "Gig 2",
								},
							},
						},
						Error: nil,
					},
				},
				Metrics: &reconcilerpb.IngestionMetrics{
					Total: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			queryFilter:      "*",
			queryLimitOffset: 1430,
			failCmd:          false,
			hasError:         false,
		},
		{
			name: "error parsing extra attributes",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "AAAFlg=="},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `"extra":is":"invalid"`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			queryFilter:      "*",
			queryLimitOffset: 1430,
			failCmd:          false,
			hasError:         true,
		},
		{
			name:     "error decoding page token",
			in:       reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "invalid"},
			failCmd:  false,
			hasError: true,
		},
		{
			name: "error parsing response json",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"id":"2mAT7vZ38H4ttI0i5dBebwJbSnZ","ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":2}`,
							"ingestion_ts": 123,
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			queryFilter: "*",
			failCmd:     false,
			hasError:    true,
		},
		{
			name:        "redis error",
			in:          reconcilerpb.RetrieveIngestionLogsRequest{},
			queryFilter: "*",
			failCmd:     true,
			hasError:    true,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			mockRedisClient := new(mr.RedisClient)

			cmd := redis.NewCmd(ctx)
			cmd.SetVal(tt.result)
			if tt.failCmd {
				cmd.SetErr(errors.New("error"))
			}
			mockRedisClient.On("Do", ctx, "FT.SEARCH", "ingest-entity", tt.queryFilter, "SORTBY", "id", "DESC", "LIMIT", tt.queryLimitOffset, int32(100)).
				Return(cmd)

			server := &Server{
				redisClient: mockRedisClient,
				logger:      logger,
			}

			response, err := server.RetrieveIngestionLogs(ctx, &tt.in)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.response.Logs), len(response.Logs))
				for i := range response.Logs {
					assert.Equal(t, tt.response.Logs[i].DataType, response.Logs[i].DataType)
					assert.Equal(t, tt.response.Logs[i].State, response.Logs[i].State)
					assert.Equal(t, tt.response.Logs[i].RequestId, response.Logs[i].RequestId)
					assert.Equal(t, tt.response.Logs[i].IngestionTs, response.Logs[i].IngestionTs)
					assert.Equal(t, tt.response.Logs[i].ProducerAppName, response.Logs[i].ProducerAppName)
					assert.Equal(t, tt.response.Logs[i].ProducerAppVersion, response.Logs[i].ProducerAppVersion)
					assert.Equal(t, tt.response.Logs[i].SdkName, response.Logs[i].SdkName)
					assert.Equal(t, tt.response.Logs[i].SdkVersion, response.Logs[i].SdkVersion)
					assert.Equal(t, tt.response.Logs[i].Entity.String(), response.Logs[i].Entity.String())
				}
				require.Equal(t, tt.response.Metrics, response.Metrics)
			}
		})
	}
}

func TestRetrieveIngestionLogsMetricsOnly(t *testing.T) {
	tests := []struct {
		name          string
		expectedTotal interface{}
		cmdError      bool
		execError     error
		hasError      bool
		errorMsg      string
	}{
		{
			name:          "valid request",
			expectedTotal: int64(10),
			cmdError:      false,
			hasError:      false,
		},
		{
			name:     "query error",
			cmdError: true,
			hasError: true,
			errorMsg: "failed to retrieve ingestion logs: cmd error",
		},
		{
			name:      "exec error",
			cmdError:  false,
			execError: errors.New("exec error"),
			hasError:  true,
			errorMsg:  "failed to retrieve ingestion logs: exec error",
		},
		{
			name:          "error getting total results",
			expectedTotal: nil,
			cmdError:      false,
			hasError:      true,
			errorMsg:      "failed to retrieve ingestion logs: failed to parse total_results",
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			expected := &reconcilerpb.IngestionMetrics{
				New:        3,
				Reconciled: 3,
				Failed:     2,
				NoChanges:  2,
				Total:      10,
			}

			mockRedisClient := new(mr.RedisClient)

			mockPipeliner := new(MockPipeliner)

			cmdTotal := redis.NewCmd(ctx)
			if tt.cmdError {
				cmdTotal.SetErr(errors.New("cmd error"))
			}
			cmdTotal.SetVal(interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{},
				},
				"total_results": tt.expectedTotal,
				"warning":       []interface{}{},
			}))
			mockPipeliner.On("Do", ctx, []interface{}{"FT.SEARCH", "ingest-entity", "*", "LIMIT", 0, 0}).Return(cmdTotal)

			cmdNew := redis.NewCmd(ctx)
			cmdNew.SetVal(interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{},
				},
				"total_results": int64(expected.New),
				"warning":       []interface{}{},
			}))
			mockPipeliner.On("Do", ctx, []interface{}{"FT.SEARCH", "ingest-entity", "@state:NEW", "LIMIT", 0, 0}).Return(cmdNew)

			cmdReconciled := redis.NewCmd(ctx)
			cmdReconciled.SetVal(interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{},
				},
				"total_results": int64(expected.Reconciled),
				"warning":       []interface{}{},
			}))
			mockPipeliner.On("Do", ctx, []interface{}{"FT.SEARCH", "ingest-entity", "@state:RECONCILED", "LIMIT", 0, 0}).Return(cmdReconciled)

			cmdFailed := redis.NewCmd(ctx)
			cmdFailed.SetVal(interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{},
				},
				"total_results": int64(expected.Failed),
				"warning":       []interface{}{},
			}))
			mockPipeliner.On("Do", ctx, []interface{}{"FT.SEARCH", "ingest-entity", "@state:FAILED", "LIMIT", 0, 0}).Return(cmdFailed)

			cmdNoChanges := redis.NewCmd(ctx)
			cmdNoChanges.SetVal(interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{},
				},
				"total_results": int64(expected.NoChanges),
				"warning":       []interface{}{},
			}))
			mockPipeliner.On("Do", ctx, []interface{}{"FT.SEARCH", "ingest-entity", "@state:NO_CHANGES", "LIMIT", 0, 0}).Return(cmdNoChanges)

			mockPipeliner.On("Exec", ctx).Return(tt.execError)
			mockRedisClient.On("Pipeline").Return(mockPipeliner)

			in := reconcilerpb.RetrieveIngestionLogsRequest{OnlyMetrics: true}

			server := &Server{
				redisClient: mockRedisClient,
				logger:      logger,
			}

			response, err := server.RetrieveIngestionLogs(ctx, &in)
			if tt.hasError {
				require.Error(t, err)
				require.Equal(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, expected, response.Metrics)
			}
		})
	}
}
