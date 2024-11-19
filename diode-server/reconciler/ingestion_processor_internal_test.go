package reconciler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	mnp "github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	mr "github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func strPtr(s string) *string { return &s }

func TestWriteIngestionLog(t *testing.T) {
	tests := []struct {
		name         string
		ingestionLog *reconcilerpb.IngestionLog
		hasError     bool
		hasMock      bool
	}{
		{
			name: "write successful",
			ingestionLog: &reconcilerpb.IngestionLog{
				RequestId: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
			},
			hasError: false,
			hasMock:  true,
		},
		{
			name: "redis error",
			ingestionLog: &reconcilerpb.IngestionLog{
				RequestId: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
				IngestionTs: time.Now().UnixNano(),
			},
			hasError: true,
			hasMock:  true,
		},
	}
	for i := range tests {
		tt := tests[i]

		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			key := "test-key"

			// Create a mock Redis client
			mockRedisClient := new(mr.RedisClient)
			p := &IngestionProcessor{
				redisClient: mockRedisClient,
				logger:      slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false})),
				Config: Config{
					AutoApplyChangesets:        true,
					ReconcilerRateLimiterRPS:   20,
					ReconcilerRateLimiterBurst: 1,
				},
			}

			// Set up the mock expectation
			cmd := redis.NewCmd(ctx)
			if tt.hasError {
				cmd.SetErr(errors.New("error"))
			}
			mockRedisClient.On("Do", ctx, "JSON.SET", "test-key", "$", mock.Anything).
				Return(cmd)

			// Call the method
			_, err := p.writeIngestionLog(ctx, key, tt.ingestionLog)

			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			// Assert the expectations
			if tt.hasMock {
				mockRedisClient.AssertExpectations(t)
			}
		})
	}
}

func TestHandleStreamMessage(t *testing.T) {
	tests := []struct {
		name              string
		validMsg          bool
		entities          []*diodepb.Entity
		mockChangeSet     *changeset.ChangeSet
		changeSetResponse *netboxdiodeplugin.ChangeSetResponse
		changeSetError    error
		reconcilerError   bool
		expectedError     bool
	}{
		{
			name:     "successful processing",
			validMsg: true,
			entities: []*diodepb.Entity{
				{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "test-site-name",
						},
					},
				},
			},
			changeSetResponse: &netboxdiodeplugin.ChangeSetResponse{},
			reconcilerError:   false,
			expectedError:     false,
		},
		{
			name:     "unmarshal error",
			validMsg: false,
			entities: []*diodepb.Entity{
				{
					Entity: nil,
				},
			},
			reconcilerError: false,
			expectedError:   true,
		},
		{
			name:     "reconcile error",
			validMsg: true,
			entities: []*diodepb.Entity{
				{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "test-site-name",
						},
					},
				},
			},
			changeSetResponse: &netboxdiodeplugin.ChangeSetResponse{},
			reconcilerError:   true,
			expectedError:     false,
		},
		{
			name:     "no entities",
			validMsg: true,
			entities: []*diodepb.Entity{
				{
					Entity: nil,
				},
			},
			changeSetResponse: &netboxdiodeplugin.ChangeSetResponse{},
			reconcilerError:   false,
			expectedError:     false,
		},
		{
			name:     "change set available",
			validMsg: true,
			entities: []*diodepb.Entity{
				{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "test-site-name",
						},
					},
				},
			},
			mockChangeSet: &changeset.ChangeSet{
				ChangeSetID: "cs123",
				ChangeSet:   []changeset.Change{},
			},
			changeSetResponse: &netboxdiodeplugin.ChangeSetResponse{
				ChangeSetID: "cs123",
				Result:      "changed",
			},
			reconcilerError: false,
			expectedError:   false,
		},
		{
			name:     "change set apply error",
			validMsg: true,
			entities: []*diodepb.Entity{
				{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "test-site-name",
						},
					},
				},
			},
			mockChangeSet: &changeset.ChangeSet{
				ChangeSetID: "cs123",
				ChangeSet:   []changeset.Change{},
			},
			changeSetResponse: &netboxdiodeplugin.ChangeSetResponse{
				ChangeSetID: "cs123",
				Result:      "changed",
			},
			changeSetError:  errors.New("apply error"),
			reconcilerError: false,
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRedisClient := new(mr.RedisClient)
			mockRedisStreamClient := new(mr.RedisClient)
			mockNbClient := new(mnp.NetBoxAPI)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			p := &IngestionProcessor{
				nbClient:          mockNbClient,
				redisClient:       mockRedisClient,
				redisStreamClient: mockRedisStreamClient,
				logger:            logger,
				Config: Config{
					AutoApplyChangesets:        true,
					ReconcilerRateLimiterRPS:   20,
					ReconcilerRateLimiterBurst: 1,
				},
			}

			request := redis.XMessage{}
			if tt.validMsg {
				reqBytes, err := proto.Marshal(&diodepb.IngestRequest{
					Id:       "req123",
					Entities: tt.entities,
				})
				if err == nil {
					request = redis.XMessage{
						ID: "1",
						Values: map[string]interface{}{
							"request":      string(reqBytes),
							"ingestion_ts": "1720425600",
						},
					}
				}
			} else {
				request = redis.XMessage{
					ID: "2",
					Values: map[string]interface{}{
						"request":      "invalid-request",
						"ingestion_ts": "1720425600",
					},
				}
			}
			if tt.reconcilerError {
				mockNbClient.On("RetrieveObjectState", ctx, mock.Anything).Return(&netboxdiodeplugin.ObjectState{}, errors.New("prepare error"))
			} else {
				mockNbClient.On("RetrieveObjectState", ctx, mock.Anything).Return(&netboxdiodeplugin.ObjectState{ObjectType: "dcim.site",
					ObjectID:       0,
					ObjectChangeID: 0,
					Object: &netbox.DcimSiteDataWrapper{
						Site: nil,
					}}, nil)
			}
			mockNbClient.On("ApplyChangeSet", ctx, mock.Anything).Return(tt.changeSetResponse, tt.changeSetError)
			if tt.entities[0].Entity != nil {
				mockRedisClient.On("Do", ctx, "JSON.SET", mock.Anything, "$", mock.Anything).Return(redis.NewCmd(ctx))
			}
			mockRedisStreamClient.On("XAck", ctx, mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))
			mockRedisStreamClient.On("XDel", ctx, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))

			err := p.handleStreamMessage(ctx, request)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.validMsg {
				mockRedisClient.AssertExpectations(t)
			}
		})
	}
}

func TestConsumeIngestionStream(t *testing.T) {
	tests := []struct {
		name          string
		groupError    bool
		expectedError bool
	}{
		{
			name:          "group create error",
			groupError:    true,
			expectedError: true,
		},
		{
			name:          "handle stream message error",
			groupError:    false,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRedisClient := new(mr.RedisClient)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			cmdSlice := redis.NewXStreamSliceCmd(ctx)
			streams := []redis.XStream{
				{
					Stream: "test-stream",
					Messages: []redis.XMessage{
						{
							ID: "1",
							Values: map[string]interface{}{
								"request":      "invalid-request",
								"ingestion_ts": "timestamp",
							},
						},
					},
				},
			}
			cmdSlice.SetVal(streams)

			status := redis.NewStatusCmd(ctx)
			if tt.groupError {
				status.SetErr(errors.New("group create error"))
			} else {
				mockRedisClient.On("XReadGroup", mock.Anything, mock.Anything).Return(cmdSlice)
			}
			mockRedisClient.On("XGroupCreateMkStream", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(status)

			p := &IngestionProcessor{
				redisStreamClient: mockRedisClient,
				logger:            logger,
				Config: Config{
					AutoApplyChangesets:        true,
					ReconcilerRateLimiterRPS:   20,
					ReconcilerRateLimiterBurst: 1,
				},
			}

			err := p.consumeIngestionStream(ctx, "test-stream", "test-group", "test-consumer")

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockRedisClient.AssertExpectations(t)
		})
	}
}

func TestCompressChangeSet(t *testing.T) {
	cs := changeset.ChangeSet{
		ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
		ChangeSet: []changeset.Change{
			{
				ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b6",
				ChangeType:    changeset.ChangeTypeCreate,
				ObjectType:    "extras.tag",
				ObjectID:      nil,
				ObjectVersion: nil,
				Data: &netbox.Tag{
					Name: "tag 2",
					Slug: "tag-2",
				},
			},
			{
				ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeType:    changeset.ChangeTypeUpdate,
				ObjectType:    "dcim.site",
				ObjectVersion: nil,
				Data: &netbox.DcimSite{
					ID:     1,
					Name:   "Site A",
					Slug:   "site-a",
					Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
					Tags: []*netbox.Tag{
						{
							ID:   1,
							Name: "tag 1",
							Slug: "tag-1",
						},
						{
							ID:   3,
							Name: "tag 3",
							Slug: "tag-3",
						},
						{
							Name: "tag 2",
							Slug: "tag-2",
						},
					},
				},
			},
		},
	}
	compressed, err := compressChangeSet(&cs)
	require.NoError(t, err)

	// Decompress the compressed data
	r := brotli.NewReader(bytes.NewReader(compressed))
	var decodedOutput bytes.Buffer
	n, err := io.Copy(&decodedOutput, r)
	require.NoError(t, err)

	csJSON, err := json.Marshal(cs)
	require.NoError(t, err)

	// Assert the decompressed data is the same as the original data
	require.Equal(t, int64(len(csJSON)), n)
	require.Equal(t, csJSON, decodedOutput.Bytes())
	require.Contains(t, decodedOutput.String(), "5663a77e-9bad-4981-afe9-77d8a9f2b8b5")
}

func TestIngestionProcessor_GenerateAndApplyChangeSet(t *testing.T) {
	tests := []struct {
		name                            string
		ingestionLog                    *reconcilerpb.IngestionLog
		mockRetrieveObjectStateResponse *netboxdiodeplugin.ObjectState
		mockApplyChangeSetResponse      *netboxdiodeplugin.ChangeSetResponse
		autoApplyChangesets             bool
		expectedStatus                  reconcilerpb.State
		expectedError                   bool
	}{
		{
			name: "generate and apply changeset",
			ingestionLog: &reconcilerpb.IngestionLog{
				Id:                 ksuid.New().String(),
				RequestId:          "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "0.1.0",
				SdkName:            "diode-sdk-go",
				SdkVersion:         "0.2.0",
				DataType:           "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
				IngestionTs: time.Now().UnixNano(),
				State:       reconcilerpb.State_QUEUED,
			},
			mockRetrieveObjectStateResponse: &netboxdiodeplugin.ObjectState{
				ObjectType: "dcim.site",
				ObjectID:   0,
				Object: &netbox.DcimSiteDataWrapper{
					Site: nil,
				},
			},
			mockApplyChangeSetResponse: &netboxdiodeplugin.ChangeSetResponse{
				ChangeSetID: "00000000-0000-0000-0000-000000000000",
				Result:      "success",
			},
			autoApplyChangesets: true,
			expectedStatus:      reconcilerpb.State_RECONCILED,
			expectedError:       false,
		},
		{
			name: "generate changeset only",
			ingestionLog: &reconcilerpb.IngestionLog{
				Id:                 ksuid.New().String(),
				RequestId:          "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "0.1.0",
				SdkName:            "diode-sdk-go",
				SdkVersion:         "0.2.0",
				DataType:           "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
				IngestionTs: time.Now().UnixNano(),
				State:       reconcilerpb.State_QUEUED,
			},
			mockRetrieveObjectStateResponse: &netboxdiodeplugin.ObjectState{
				ObjectType: "dcim.site",
				ObjectID:   0,
				Object: &netbox.DcimSiteDataWrapper{
					Site: nil,
				},
			},
			autoApplyChangesets: false,
			expectedStatus:      reconcilerpb.State_QUEUED,
			expectedError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRedisClient := new(mr.RedisClient)
			mockNbClient := new(mnp.NetBoxAPI)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			p := &IngestionProcessor{
				redisClient: mockRedisClient,
				nbClient:    mockNbClient,
				logger:      logger,
				Config: Config{
					AutoApplyChangesets:        tt.autoApplyChangesets,
					ReconcilerRateLimiterRPS:   20,
					ReconcilerRateLimiterBurst: 1,
				},
			}

			// Set up the mock expectation
			cmd := redis.NewCmd(ctx)
			if tt.expectedError {
				cmd.SetErr(errors.New("error"))
			}
			redisKey := fmt.Sprintf("ingest-entity:%s-%d-%s", tt.ingestionLog.DataType, tt.ingestionLog.IngestionTs, tt.ingestionLog.Id)
			mockRedisClient.On("Do", ctx, "JSON.SET", redisKey, "$", mock.Anything).
				Return(cmd)

			mockNbClient.On("RetrieveObjectState", ctx, mock.Anything).Return(tt.mockRetrieveObjectStateResponse, nil)
			if tt.autoApplyChangesets {
				mockNbClient.On("ApplyChangeSet", ctx, mock.Anything).Return(tt.mockApplyChangeSetResponse, nil)
			}

			bufCapacity := 1

			generateChangeSetChannel := make(chan IngestionLogToProcess, bufCapacity)
			var applyChangeSetChannel chan IngestionLogToProcess
			if tt.autoApplyChangesets {
				applyChangeSetChannel = make(chan IngestionLogToProcess, bufCapacity)
			}
			generateChangeSetDone := make(chan struct{})
			applyChangeSetDone := make(chan struct{})

			p.GenerateChangeSet(ctx, generateChangeSetChannel, applyChangeSetChannel, generateChangeSetDone)
			if tt.autoApplyChangesets {
				p.ApplyChangeSet(ctx, applyChangeSetChannel, applyChangeSetDone)
			}

			generateChangeSetChannel <- IngestionLogToProcess{
				key:          redisKey,
				ingestionLog: tt.ingestionLog,
			}
			close(generateChangeSetChannel)

			<-generateChangeSetDone
			if tt.autoApplyChangesets {
				<-applyChangeSetDone
			}

			mockRedisClient.AssertExpectations(t)
			require.NotNil(t, tt.ingestionLog.ChangeSet)
			require.Equal(t, tt.expectedStatus, tt.ingestionLog.State)
		})
	}
}
