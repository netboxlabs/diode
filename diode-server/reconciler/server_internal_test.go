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
		name     string
		in       reconcilerpb.RetrieveIngestionLogsRequest
		result   interface{}
		response *reconcilerpb.RetrieveIngestionLogsResponse
		failCmd  bool
		hasError bool
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
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-db0931ec-c119-4702-bd74-4f0bed4e110b",
						"values": []interface{}{},
					},
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.device","entity":{"device":{"name":"Conference_Room_AP_02","deviceType":{"model":"Cisco Aironet 3802","manufacturer":{"name":"Cisco"}},"role":{"name":"Wireless_AP"},"serial":"PQR456789012","site":{"name":"HQ"}}},"ingestionTs":1725552654541975975,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"bc1052e3-656a-42f0-b364-27b385e02a0c","sdkName":"diode-sdk-python","sdkVersion":"0.0.1","state":1}`,
							"ingestion_ts": "1725552654541976064",
						},
						"id":     "ingest-entity:dcim.device-1725552654541975975-a6123183-1a5b-4331-ad73-4713cbee85bb",
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
						Entity:             &diodepb.Entity{},
						Error:              nil,
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
						Entity:             &diodepb.Entity{},
						Error:              nil,
					},
				},
				NextPageToken: "F/Jk/zc08gA=",
			},
			failCmd:  false,
			hasError: false,
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
							"$":            `{"dataType":"ipam.ipaddress","entity":{"ip_address":{"address":"192.168.1.1","interface":null,"description":"Vendor: HUAWEI TECHNOLOGIES"}},"error":{"message":"failed to apply change set","code":400,"details":{"change_set_id":"6304c706-f955-4bcb-a1cc-514293d53d07","result":"failed","errors":[{"error":"address: Duplicate IP address found in global table: 192.168.1.1/32","change_id":"ff9e29b2-7a64-40ba-99a8-21f44768f60a"}]}},"ingestionTs":1725046967777525928,"producerAppName":"example-app","producerAppVersion":"0.1.0","request_id":"e03c4892-5b7e-4c39-b5e6-0225a264ab8b","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":2}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:ipam.ipaddress-1725046967777525928-5e272f66-61a9-4142-8013-a7aadb7264a8",
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
						Entity:             &diodepb.Entity{},
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
				NextPageToken: "F/JlO7d81QA=",
			},
			failCmd:  false,
			hasError: false,
		},
		{
			name: "filter by state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_RECONCILED.Enum()},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-db0931ec-c119-4702-bd74-4f0bed4e110b",
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
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity:             &diodepb.Entity{},
						Error:              nil,
					},
				},
				NextPageToken: "F/JlO7d81QA=",
			},
			failCmd:  false,
			hasError: false,
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
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-db0931ec-c119-4702-bd74-4f0bed4e110b",
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
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity:             &diodepb.Entity{},
						Error:              nil,
					},
				},
				NextPageToken: "F/JlO7d81QA=",
			},
			failCmd:  false,
			hasError: false,
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
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-db0931ec-c119-4702-bd74-4f0bed4e110b",
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
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity:             &diodepb.Entity{},
						Error:              nil,
					},
				},
				NextPageToken: "F/JlO7d81QA=",
			},
			failCmd:  false,
			hasError: false,
		},
		{
			name: "pagination check",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "F/Jk/zc08gA="},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": "1725552914392208640",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-db0931ec-c119-4702-bd74-4f0bed4e110b",
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
						State:              reconcilerpb.State_RECONCILED,
						RequestId:          "req-id",
						IngestionTs:        1725552914392208722,
						ProducerAppName:    "diode-agent",
						ProducerAppVersion: "0.0.1",
						SdkName:            "diode-sdk-go",
						SdkVersion:         "0.1.0",
						Entity:             &diodepb.Entity{},
						Error:              nil,
					},
				},
				NextPageToken: "F/JlO7d81QA=",
			},
			failCmd:  false,
			hasError: false,
		},
		{
			name: "error parsing extra attributes",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "F/Jk/zc08gA="},
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
			failCmd:  false,
			hasError: true,
		},
		{
			name: "error parsing timestamp",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "F/Jk/zc08gA="},
			result: interface{}(map[interface{}]interface{}{
				"attributes": []interface{}{},
				"format":     "STRING",
				"results": []interface{}{
					map[interface{}]interface{}{
						"extra_attributes": map[interface{}]interface{}{
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": "invalid_ts",
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-db0931ec-c119-4702-bd74-4f0bed4e110b",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			failCmd:  false,
			hasError: true,
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
							"$":            `{"dataType":"dcim.interface","entity":{"interface":{"device":{"name":"my_dev"},"name":"Gig 2"}},"ingestionTs":1725552914392208722,"producerAppName":"diode-agent","producerAppVersion":"0.0.1","request_id":"req-id","sdkName":"diode-sdk-go","sdkVersion":"0.1.0","state":1}`,
							"ingestion_ts": 123,
						},
						"id":     "ingest-entity:dcim.interface-1725552914392208722-db0931ec-c119-4702-bd74-4f0bed4e110b",
						"values": []interface{}{},
					},
				},
				"total_results": 1,
				"warning":       []interface{}{},
			}),
			failCmd:  false,
			hasError: true,
		},
		{
			name:     "redis error",
			in:       reconcilerpb.RetrieveIngestionLogsRequest{},
			failCmd:  true,
			hasError: true,
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
			mockRedisClient.On("Do", ctx, "FT.SEARCH", "ingest-entity", mock.Anything, "SORTBY", "ingestion_ts", "DESC", "LIMIT", 0, mock.Anything).
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
			}
		})
	}
}
