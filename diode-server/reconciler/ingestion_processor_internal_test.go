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

	"github.com/andybalholm/brotli"
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
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

func strPtr(s string) *string { return &s }

// testCompressBrotli compresses data using brotli for test message construction.
func testCompressBrotli(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := brotli.NewWriterLevel(&buf, 1)
	_, err := w.Write(data)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf.Bytes()
}

func TestHandleStreamMessage(t *testing.T) {
	tests := []struct {
		name          string
		validMsg      bool
		entities      []*diodepb.Entity
		expectedError bool
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
			expectedError: false,
		},
		{
			name:     "unmarshal error",
			validMsg: false,
			entities: []*diodepb.Entity{
				{
					Entity: nil,
				},
			},
			expectedError: true,
		},
		{
			name:     "no entities",
			validMsg: true,
			entities: []*diodepb.Entity{
				{
					Entity: nil,
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockRedisClient := new(mr.RedisClient)
			mockRedisStreamClient := new(mr.RedisClient)
			mockNbClient := new(mnp.NetBoxAPI)
			mockRepository := new(mr.Repository)
			mockMetrics := mr.NewMetrics(t)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

			p := &IngestionProcessor{
				redisClient:       mockRedisClient,
				redisStreamClient: mockRedisStreamClient,
				logger:            logger,
				Config:            Config{},
				ops:               NewOps(mockRepository, mockNbClient, logger, nil),
				metrics:           mockMetrics,
			}

			request := redis.XMessage{}
			if tt.validMsg {
				reqBytes, err := proto.Marshal(&diodepb.IngestRequest{
					Id:       "req123",
					Entities: tt.entities,
				})
				if err == nil {
					compressed := testCompressBrotli(t, reqBytes)
					request = redis.XMessage{
						ID: "1",
						Values: map[string]interface{}{
							"request":      string(compressed),
							"encoding":     "br",
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
			mockNbClient.On("GetDefaultBranch", mock.Anything).Return((*netboxdiodeplugin.Branch)(nil), nil)
			if tt.entities[0].Entity != nil {
				mockRepository.On("FindPriorIngestionLogsByEntityHashes", mock.Anything, mock.Anything, mock.Anything).Return(map[string]*ops.PriorIngestionLog{}, nil)
				mockRepository.On("BulkCreateIngestionLogs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
					func(_ context.Context, logs []*reconcilerpb.IngestionLog, _ [][]byte, _ []string) map[string]int32 {
						result := make(map[string]int32, len(logs))
						for _, log := range logs {
							result[log.Id] = 1
						}
						return result
					}, nil)
			}

			mockRedisStreamClient.On("XAck", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))
			mockRedisStreamClient.On("XDel", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))
			mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()

			if tt.entities[0].Entity != nil {
				mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()
			}

			err := p.handleStreamMessage(ctx, request)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
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
			mockMetrics := mr.NewMetrics(t)
			if !tt.groupError {
				// Only expect metrics if we're actually processing messages (no group error)
				mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
			}

			p := &IngestionProcessor{
				redisStreamClient: mockRedisClient,
				logger:            logger,
				Config: Config{
					AutoApplyChangesets: true,
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

func TestHandleStreamMessageLegacyUncompressed(t *testing.T) {
	ctx := context.Background()
	mockRedisClient := new(mr.RedisClient)
	mockRedisStreamClient := new(mr.RedisClient)
	mockNbClient := new(mnp.NetBoxAPI)
	mockRepository := new(mr.Repository)
	mockMetrics := mr.NewMetrics(t)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	p := &IngestionProcessor{
		redisClient:       mockRedisClient,
		redisStreamClient: mockRedisStreamClient,
		logger:            logger,
		Config:            Config{},
		ops:               NewOps(mockRepository, mockNbClient, logger, nil),
		metrics:           mockMetrics,
	}

	reqBytes, err := proto.Marshal(&diodepb.IngestRequest{
		Id: "legacy-req",
		Entities: []*diodepb.Entity{
			{
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{Name: "legacy-site"},
				},
			},
		},
	})
	require.NoError(t, err)

	request := redis.XMessage{
		ID: "1",
		Values: map[string]interface{}{
			"request":      string(reqBytes),
			"ingestion_ts": "1720425600",
		},
	}

	mockNbClient.On("GetDefaultBranch", mock.Anything).Return((*netboxdiodeplugin.Branch)(nil), nil)
	mockRepository.On("FindPriorIngestionLogsByEntityHashes", mock.Anything, mock.Anything, mock.Anything).Return(map[string]*ops.PriorIngestionLog{}, nil)
	mockRepository.On("BulkCreateIngestionLogs", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		func(_ context.Context, logs []*reconcilerpb.IngestionLog, _ [][]byte, _ []string) map[string]int32 {
			result := make(map[string]int32, len(logs))
			for _, log := range logs {
				result[log.Id] = 1
			}
			return result
		}, nil)
	mockRedisStreamClient.On("XAck", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))
	mockRedisStreamClient.On("XDel", mock.Anything, mock.Anything, mock.Anything).Return(redis.NewIntCmd(ctx))
	mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
	mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()

	err = p.handleStreamMessage(ctx, request)
	require.NoError(t, err)

	mockRepository.AssertExpectations(t)
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
