package reconciler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	mnp "github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	mr "github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }

func TestHandleStreamMessage(t *testing.T) {
	tests := []struct {
		name              string
		validMsg          bool
		entities          []*diodepb.Entity
		mockChangeSet     *changeset.ChangeSet
		changeSetResponse *netboxdiodeplugin.ApplyChangeSetResponse
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
			changeSetResponse: &netboxdiodeplugin.ApplyChangeSetResponse{},
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
			changeSetResponse: &netboxdiodeplugin.ApplyChangeSetResponse{},
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
			changeSetResponse: &netboxdiodeplugin.ApplyChangeSetResponse{},
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
				ID:      "cs123",
				Changes: []changeset.Change{},
			},
			changeSetResponse: &netboxdiodeplugin.ApplyChangeSetResponse{
				ID:     "cs123",
				Result: "changed",
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
				ID:      "cs123",
				Changes: []changeset.Change{},
			},
			changeSetResponse: &netboxdiodeplugin.ApplyChangeSetResponse{
				ID:     "cs123",
				Result: "changed",
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
			mockRepository := new(mr.Repository)
			mockMetrics := new(mr.IngestionProcessorMetrics)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			p := &IngestionProcessor{
				redisClient:       mockRedisClient,
				redisStreamClient: mockRedisStreamClient,
				logger:            logger,
				Config: Config{
					AutoApplyChangesets:        true,
					ReconcilerRateLimiterRPS:   20,
					ReconcilerRateLimiterBurst: 1,
				},
				ops:     NewOps(mockRepository, mockNbClient, logger),
				metrics: mockMetrics,
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
				mockNbClient.On("GenerateDiff", mock.Anything, mock.Anything).Return(nil, errors.New("prepare error"))
			} else {
				mockNbClient.On("GenerateDiff", mock.Anything, mock.Anything).Return(&netboxdiodeplugin.GenerateDiffResponse{
					Changes: []netboxdiodeplugin.Change{
						{
							ID:         "00000000-0000-0000-0000-000000000000",
							ChangeType: "create",
							ObjectType: "dcim.site",
							Data:       json.RawMessage(`{"name": "Site A"}`),
						},
					},
				}, nil)
			}
			mockNbClient.On("ApplyChangeSet", mock.Anything, mock.Anything).Return(tt.changeSetResponse, tt.changeSetError)
			if tt.entities[0].Entity != nil {
				mockRepository.On("CreateIngestionLog", mock.Anything, mock.Anything, mock.Anything).Return(int32Ptr(1), nil)
				mockRepository.On("CreateChangeSet", mock.Anything, mock.Anything, mock.Anything).Return(int32Ptr(1), nil)
				mockRepository.On("UpdateIngestionLogStateWithError", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			}
			mockRedisStreamClient.On("XAck", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))
			mockRedisStreamClient.On("XDel", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))
			mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordChangeSetCreate", mock.Anything, mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordChangeSetApply", mock.Anything, mock.Anything, mock.Anything).Return()

			allDone, err := p.handleStreamMessage(ctx, request)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			select {
			case <-allDone:
			case <-time.After(1 * time.Second):
				require.Fail(t, "allDone channel did not close")
			}

			if tt.validMsg {
				mockRepository.AssertExpectations(t)
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
			mockMetrics := new(mr.IngestionProcessorMetrics)
			mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordChangeSetCreate", mock.Anything, mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordChangeSetApply", mock.Anything, mock.Anything, mock.Anything).Return()

			p := &IngestionProcessor{
				redisStreamClient: mockRedisClient,
				logger:            logger,
				Config: Config{
					AutoApplyChangesets:        true,
					ReconcilerRateLimiterRPS:   20,
					ReconcilerRateLimiterBurst: 1,
				},
				metrics: mockMetrics,
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
		ID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
		Changes: []changeset.Change{
			{
				ID:            "5663a77e-9bad-4981-afe9-77d8a9f2b8b6",
				ChangeType:    changeset.ChangeTypeCreate,
				ObjectType:    "extras.tag",
				ObjectID:      nil,
				ObjectVersion: nil,
				After:         json.RawMessage(`{"name": "tag 2", "slug": "tag-2"}`),
			},
			{
				ID:            "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeType:    changeset.ChangeTypeUpdate,
				ObjectType:    "dcim.site",
				ObjectVersion: nil,
				After:         json.RawMessage(`{"name": "Site A", "slug": "site-a", "status": "active", "tags": [{"id": 1, "name": "tag 1", "slug": "tag-1"}, {"id": 3, "name": "tag 3", "slug": "tag-3"}, {"id": 2, "name": "tag 2", "slug": "tag-2"}]}`),
			},
		},
	}
	compressed, err := changeset.CompressChangeSet(&cs)
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
		name                       string
		ingestionLog               *reconcilerpb.IngestionLog
		mockGenerateDiffResponse   *netboxdiodeplugin.GenerateDiffResponse
		mockApplyChangeSetResponse *netboxdiodeplugin.ApplyChangeSetResponse
		autoApplyChangesets        bool
		expectedStatus             reconcilerpb.State
		expectedError              bool
	}{
		{
			name: "generate and apply change set",
			ingestionLog: &reconcilerpb.IngestionLog{
				Id:                 uuid.NewString(),
				RequestId:          "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "0.1.0",
				SdkName:            "diode-sdk-go",
				SdkVersion:         "0.2.0",
				ObjectType:         "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
				IngestionTs: time.Now().UnixNano(),
				SourceTs:    time.Now().UnixNano(),
				State:       reconcilerpb.State_QUEUED,
			},
			mockGenerateDiffResponse: &netboxdiodeplugin.GenerateDiffResponse{
				Changes: []netboxdiodeplugin.Change{
					{
						ID:         "00000000-0000-0000-0000-000000000000",
						ChangeType: "create",
						ObjectType: "dcim.site",
						Data:       json.RawMessage(`{"name": "Site A"}`),
					},
				},
			},
			mockApplyChangeSetResponse: &netboxdiodeplugin.ApplyChangeSetResponse{
				ID:     "00000000-0000-0000-0000-000000000000",
				Result: "success",
			},
			autoApplyChangesets: true,
			expectedStatus:      reconcilerpb.State_APPLIED,
			expectedError:       false,
		},
		{
			name: "generate change set only",
			ingestionLog: &reconcilerpb.IngestionLog{
				Id:                 uuid.NewString(),
				RequestId:          "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "0.1.0",
				SdkName:            "diode-sdk-go",
				SdkVersion:         "0.2.0",
				ObjectType:         "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
				IngestionTs: time.Now().UnixNano(),
				SourceTs:    time.Now().UnixNano(),
				State:       reconcilerpb.State_OPEN,
			},
			mockGenerateDiffResponse: &netboxdiodeplugin.GenerateDiffResponse{
				Changes: []netboxdiodeplugin.Change{
					{
						ID:         "00000000-0000-0000-0000-000000000000",
						ChangeType: "create",
						ObjectType: "dcim.site",
						Data:       json.RawMessage(`{"name": "Site A"}`),
					},
				},
			},
			autoApplyChangesets: false,
			expectedStatus:      reconcilerpb.State_OPEN,
			expectedError:       false,
		},
		{
			name: "generate change set without changes",
			ingestionLog: &reconcilerpb.IngestionLog{
				Id:                 uuid.NewString(),
				RequestId:          "cfa0f129-125c-440d-9e41-e87583cd7d89",
				ProducerAppName:    "test-app",
				ProducerAppVersion: "0.1.0",
				SdkName:            "diode-sdk-go",
				SdkVersion:         "0.2.0",
				ObjectType:         "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
				IngestionTs: time.Now().UnixNano(),
				SourceTs:    time.Now().UnixNano(),
				State:       reconcilerpb.State_QUEUED,
			},
			mockGenerateDiffResponse: &netboxdiodeplugin.GenerateDiffResponse{
				Changes: []netboxdiodeplugin.Change{},
			},
			autoApplyChangesets: false,
			expectedStatus:      reconcilerpb.State_NO_CHANGES,
			expectedError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRedisClient := new(mr.RedisClient)
			mockNbClient := new(mnp.NetBoxAPI)
			mockRepository := new(mr.Repository)
			mockMetrics := new(mr.IngestionProcessorMetrics)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			p := &IngestionProcessor{
				redisClient: mockRedisClient,
				logger:      logger,
				Config: Config{
					AutoApplyChangesets:        tt.autoApplyChangesets,
					ReconcilerRateLimiterRPS:   20,
					ReconcilerRateLimiterBurst: 1,
				},
				ops:     NewOps(mockRepository, mockNbClient, logger),
				metrics: mockMetrics,
			}

			ingestionLogID := int32(1)

			mockNbClient.On("GenerateDiff", ctx, mock.Anything).Return(tt.mockGenerateDiffResponse, nil)
			if tt.autoApplyChangesets {
				mockRepository.On("UpdateIngestionLogStateWithError", ctx, ingestionLogID, reconcilerpb.State_OPEN, mock.Anything).Return(nil)
				mockNbClient.On("ApplyChangeSet", ctx, mock.Anything).Return(tt.mockApplyChangeSetResponse, nil)
			}
			mockRepository.On("UpdateIngestionLogStateWithError", ctx, ingestionLogID, tt.expectedStatus, mock.Anything).Return(nil)
			mockRepository.On("CreateChangeSet", ctx, mock.Anything, ingestionLogID).Return(int32Ptr(1), nil)
			mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordChangeSetCreate", mock.Anything, mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordChangeSetApply", mock.Anything, mock.Anything, mock.Anything).Return()

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
				ingestionLogID: ingestionLogID,
				ingestionLog:   tt.ingestionLog,
			}
			close(generateChangeSetChannel)

			<-generateChangeSetDone
			if tt.autoApplyChangesets {
				<-applyChangeSetDone
			}

			mockRepository.AssertExpectations(t)
			require.Equal(t, tt.expectedStatus, tt.ingestionLog.State)
		})
	}
}
