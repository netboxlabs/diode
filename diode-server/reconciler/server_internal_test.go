package reconciler

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

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
		name                  string
		in                    reconcilerpb.RetrieveIngestionLogsRequest
		ingestionLogsPerState map[reconcilerpb.State]int32
		ingestionLogs         []*reconcilerpb.IngestionLog
		response              *reconcilerpb.RetrieveIngestionLogsResponse
		hasError              bool
	}{
		{
			name: "valid request",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_APPLIED: 2,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
					ObjectType:         "dcim.interface",
					State:              reconcilerpb.State_APPLIED,
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
					Id:                 "2mC8GVBGFg6NyLsQxuS4IYMB6FI",
					ObjectType:         "dcim.device",
					State:              reconcilerpb.State_APPLIED,
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
				},
			},
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						ObjectType:         "dcim.interface",
						State:              reconcilerpb.State_APPLIED,
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
						ObjectType:         "dcim.device",
						State:              reconcilerpb.State_APPLIED,
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
					Total:      2,
					Reconciled: 2,
				},
				NextPageToken: "F/Jk/zc08gA=",
			},
			hasError: false,
		},
		{
			name: "request with reconciliation error",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_FAILED: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					ObjectType:         "ipam.ipaddress",
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						ObjectType:         "ipam.ipaddress",
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
					Total:  1,
					Failed: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name: "filter by queued state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_QUEUED.Enum()},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_QUEUED: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					ObjectType:         "dcim.interface",
					State:              reconcilerpb.State_QUEUED,
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						ObjectType:         "dcim.interface",
						State:              reconcilerpb.State_QUEUED,
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
					Total:  1,
					Queued: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name: "filter by applied state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_APPLIED.Enum()},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_APPLIED: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
					ObjectType:         "dcim.interface",
					State:              reconcilerpb.State_APPLIED,
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						ObjectType:         "dcim.interface",
						State:              reconcilerpb.State_APPLIED,
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
					Total:      1,
					Reconciled: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name: "filter by failed state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_FAILED.Enum()},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_FAILED: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
					ObjectType:         "dcim.interface",
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						ObjectType:         "dcim.interface",
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
					Total:  1,
					Failed: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name: "filter by no changes state",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{State: reconcilerpb.State_NO_CHANGES.Enum()},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_NO_CHANGES: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
					ObjectType:         "dcim.interface",
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						ObjectType:         "dcim.interface",
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
					Total:     1,
					NoChanges: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name: "filter by object type",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{ObjectType: "dcim.interface"},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_APPLIED: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
					ObjectType:         "dcim.interface",
					State:              reconcilerpb.State_APPLIED,
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						ObjectType:         "dcim.interface",
						State:              reconcilerpb.State_APPLIED,
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
					Total:      1,
					Reconciled: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name: "filter by timestamp",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{IngestionTsStart: 1725552914392208639},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_APPLIED: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
					ObjectType:         "dcim.interface",
					State:              reconcilerpb.State_APPLIED,
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						ObjectType:         "dcim.interface",
						State:              reconcilerpb.State_APPLIED,
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
					Total:      1,
					Reconciled: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name: "pagination check",
			in:   reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "AAAFlg=="},
			ingestionLogsPerState: map[reconcilerpb.State]int32{
				reconcilerpb.State_APPLIED: 1,
			},
			ingestionLogs: []*reconcilerpb.IngestionLog{
				{
					Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
					ObjectType:         "dcim.interface",
					State:              reconcilerpb.State_APPLIED,
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
			response: &reconcilerpb.RetrieveIngestionLogsResponse{
				Logs: []*reconcilerpb.IngestionLog{
					{
						Id:                 "2mAT7vZ38H4ttI0i5dBebwJbSnZ",
						ObjectType:         "dcim.interface",
						State:              reconcilerpb.State_APPLIED,
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
					Total:      1,
					Reconciled: 1,
				},
				NextPageToken: "AAAFlw==",
			},
			hasError: false,
		},
		{
			name:     "error decoding page token",
			in:       reconcilerpb.RetrieveIngestionLogsRequest{PageToken: "invalid"},
			hasError: true,
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			mockRedisClient := mr.NewRedisClient(t)
			mockRepository := mr.NewRepository(t)
			server := &Server{
				redisClient: mockRedisClient,
				logger:      logger,
				repository:  mockRepository,
			}

			var retrieveErr error
			if tt.hasError {
				retrieveErr = errors.New("failed to retrieve ingestion logs")
			}

			mockRepository.On("CountIngestionLogsPerState", ctx).Return(tt.ingestionLogsPerState, nil)

			if !tt.hasError {
				mockRepository.On("RetrieveIngestionLogs", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.ingestionLogs, retrieveErr)
			}

			response, err := server.RetrieveIngestionLogs(ctx, &tt.in)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.response.Logs), len(response.Logs))
				for i := range response.Logs {
					assert.Equal(t, tt.response.Logs[i].ObjectType, response.Logs[i].ObjectType)
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
			mockRepository.AssertExpectations(t)
		})
	}
}

func TestRetrieveIngestionLogsMetricsOnly(t *testing.T) {
	tests := []struct {
		name          string
		expectedTotal interface{}
		hasError      bool
		errorMsg      string
	}{
		{
			name:          "valid request",
			expectedTotal: int64(10),
			hasError:      false,
		},
		{
			name:     "query error",
			hasError: true,
			errorMsg: "failed to retrieve ingestion logs: cmd error",
		},
		{
			name:     "exec error",
			hasError: true,
			errorMsg: "failed to retrieve ingestion logs: exec error",
		},
		{
			name:          "error getting total results",
			expectedTotal: nil,
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
				Queued:     3,
				Reconciled: 3,
				Failed:     2,
				NoChanges:  2,
				Total:      10,
			}

			mockRedisClient := mr.NewRedisClient(t)
			mockRepository := mr.NewRepository(t)
			server := &Server{
				redisClient: mockRedisClient,
				logger:      logger,
				repository:  mockRepository,
			}

			ingestionLogStateMetricsMap := map[reconcilerpb.State]int32{
				reconcilerpb.State_QUEUED:     expected.Queued,
				reconcilerpb.State_APPLIED:    expected.Reconciled,
				reconcilerpb.State_FAILED:     expected.Failed,
				reconcilerpb.State_NO_CHANGES: expected.NoChanges,
			}

			var countErr error
			if tt.hasError {
				countErr = errors.New(tt.errorMsg)
			}

			mockRepository.On("CountIngestionLogsPerState", ctx).Return(ingestionLogStateMetricsMap, countErr)

			in := reconcilerpb.RetrieveIngestionLogsRequest{OnlyMetrics: true}

			response, err := server.RetrieveIngestionLogs(ctx, &in)
			if tt.hasError {
				require.Error(t, err)
				require.Equal(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, expected, response.Metrics)
			}
			mockRepository.AssertExpectations(t)
		})
	}
}
