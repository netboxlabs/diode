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
	"github.com/redis/go-redis/v9"
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

func TestReconcileEntity(t *testing.T) {
	tests := []struct {
		name                   string
		retrieveObjectStateErr error
		applyErr               error
		expectedError          bool
		expectedCS             *changeset.ChangeSet
	}{
		{
			name:          "successful reconciliation",
			expectedError: false,
			expectedCS: &changeset.ChangeSet{
				ChangeSetID: "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
				ChangeSet: []changeset.Change{
					{
						ChangeID:      "5663a77e-9bad-4981-afe9-77d8a9f2b8b5",
						ChangeType:    changeset.ChangeTypeCreate,
						ObjectType:    "dcim.site",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimSite{
							Name:   "Site A",
							Slug:   "site-a",
							Status: (*netbox.DcimSiteStatus)(strPtr(string(netbox.DcimSiteStatusActive))),
						},
					},
				},
			},
		},
		{
			name:                   "prepare error",
			retrieveObjectStateErr: errors.New("prepare error"),
			expectedError:          true,
		},
		{
			name: "apply error",
			expectedCS: &changeset.ChangeSet{
				ChangeSetID: "cs123",
				ChangeSet:   []changeset.Change{},
			},
			applyErr:      errors.New("apply error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Mock nbClient
			mockNbClient := new(mnp.NetBoxAPI)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
			// Create IngestionProcessor
			p := &IngestionProcessor{
				nbClient: mockNbClient,
				logger:   logger,
			}

			// Setup mock for RetrieveObjectState
			if tt.retrieveObjectStateErr != nil {
				mockNbClient.On("RetrieveObjectState", ctx, mock.Anything).Return(&netboxdiodeplugin.ObjectState{}, tt.retrieveObjectStateErr)
			} else {

				mockNbClient.On("RetrieveObjectState", ctx, mock.Anything).Return(&netboxdiodeplugin.ObjectState{ObjectType: "dcim.site",
					ObjectID:       0,
					ObjectChangeID: 0,
					Object: &netbox.DcimSiteDataWrapper{
						Site: nil,
					}}, nil)
			}
			// Setup mock for ApplyChangeSet
			if tt.expectedCS != nil {
				mockNbClient.On("ApplyChangeSet", ctx, mock.Anything).Return(&netboxdiodeplugin.ChangeSetResponse{}, tt.applyErr)
			}

			// Call reconcileEntity
			ingestEntity := changeset.IngestEntity{
				RequestID: "cfa0f129-125c-440d-9e41-e87583cd7d89",
				DataType:  "dcim.site",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Site{
						Site: &diodepb.Site{
							Name: "Site A",
						},
					},
				},
			}

			cs, err := p.reconcileEntity(ctx, ingestEntity)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedCS.ChangeSet[0].ChangeType, cs.ChangeSet[0].ChangeType)
				require.Equal(t, tt.expectedCS.ChangeSet[0].ObjectType, cs.ChangeSet[0].ObjectType)
				require.Equal(t, tt.expectedCS.ChangeSet[0].Data, cs.ChangeSet[0].Data)
			}

			// Assert expectations
			mockNbClient.AssertExpectations(t)
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
