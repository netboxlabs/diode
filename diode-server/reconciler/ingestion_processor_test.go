package reconciler_test

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	pluginmocks "github.com/netboxlabs/diode/diode-server/netboxdiodeplugin/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
	"github.com/netboxlabs/diode/diode-server/reconciler/ops"
)

func int32Ptr(i int32) *int32 { return &i }
func intPtr(i int) *int       { return &i }

func TestNewIngestionProcessor(t *testing.T) {
	ctx := context.Background()
	s := miniredis.RunT(t)
	defer s.Close()

	setupEnv(s.Addr())
	defer teardownEnv()
	var cfg reconciler.Config
	envconfig.MustProcess("", &cfg)

	mockRepository := mocks.NewRepository(t)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	mockMetrics := mocks.NewMetrics(t)

	redisClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   0,
	})
	defer func() {
		_ = redisClient.Close()
	}()
	redisStreamClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   1,
	})
	defer func() {
		_ = redisStreamClient.Close()
	}()

	processor, err := reconciler.NewIngestionProcessor(ctx, logger, cfg, redisClient, redisStreamClient, reconciler.DefaultRedisStreamID, reconciler.DefaultRedisConsumerGroup, reconciler.NewOps(mockRepository, mockNetBoxClient, logger), mockMetrics)
	require.NoError(t, err)
	require.NotNil(t, processor)

	err = processor.Stop()
	require.NoError(t, err)
}

func TestIngestionProcessorStart(t *testing.T) {
	s := miniredis.RunT(t)
	s.DB(1)
	defer s.Close()

	setupEnv(s.Addr())
	defer teardownEnv()
	var cfg reconciler.Config
	envconfig.MustProcess("", &cfg)

	mockRepository := mocks.NewRepository(t)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	mockNetBoxClient := pluginmocks.NewNetBoxAPI(t)
	mockMetrics := mocks.NewMetrics(t)

	redisClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   0,
	})
	defer func() {
		_ = redisClient.Close()
	}()
	redisStreamClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   1,
	})
	defer func() {
		_ = redisStreamClient.Close()
	}()

	processor, err := reconciler.NewIngestionProcessor(ctx, logger, cfg, redisClient, redisStreamClient, reconciler.DefaultRedisStreamID, reconciler.DefaultRedisConsumerGroup, reconciler.NewOps(mockRepository, mockNetBoxClient, logger), mockMetrics)
	require.NoError(t, err)
	require.NotNil(t, processor)

	ingestReq := &diodepb.IngestRequest{
		Id:                 "test-request-id",
		ProducerAppName:    "test-app",
		ProducerAppVersion: "1.0",
		SdkName:            "test-sdk",
		SdkVersion:         "1.0",
		Entities: []*diodepb.Entity{
			{
				Entity: &diodepb.Entity_Manufacturer{
					Manufacturer: &diodepb.Manufacturer{
						Name: "test-manufacturer",
					},
				},
			},
			{
				Entity: &diodepb.Entity_Platform{
					Platform: &diodepb.Platform{
						Name: "test-platform",
						Manufacturer: &diodepb.Manufacturer{
							Name: "test-manufacturer",
						},
					},
				},
			},
			{
				Entity: &diodepb.Entity_DeviceType{
					DeviceType: &diodepb.DeviceType{
						Model: "test-model",
						Manufacturer: &diodepb.Manufacturer{
							Name: "test-manufacturer",
						},
					},
				},
			},
			{
				Entity: &diodepb.Entity_DeviceRole{
					DeviceRole: &diodepb.DeviceRole{
						Name: "test-device-role",
					},
				},
			},
			{
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{
						Name: "test-site-name",
					},
				},
			},
			{
				Entity: &diodepb.Entity_Device{
					Device: &diodepb.Device{
						Name: strPtr("test-device-name"),
						Site: &diodepb.Site{
							Name: "test-site-name",
						},
						DeviceType: &diodepb.DeviceType{
							Model: "test-model",
							Manufacturer: &diodepb.Manufacturer{
								Name: "test-manufacturer",
							},
						},
						Platform: &diodepb.Platform{
							Name: "test-platform",
							Manufacturer: &diodepb.Manufacturer{
								Name: "test-manufacturer",
							},
						},
					},
				},
			},
			{
				Entity: &diodepb.Entity_Interface{
					Interface: &diodepb.Interface{
						Name: "test-interface",
						Device: &diodepb.Device{
							Name: strPtr("test-device-name"),
							Site: &diodepb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},
			{
				Entity: &diodepb.Entity_IpAddress{
					IpAddress: &diodepb.IPAddress{
						Address: "192.168.0.1/32",
						AssignedObject: &diodepb.IPAddress_AssignedObjectInterface{
							AssignedObjectInterface: &diodepb.Interface{
								Name: "test-interface",
								Device: &diodepb.Device{
									Name: strPtr("test-device-name"),
									Site: &diodepb.Site{
										Name: "test-site-name",
									},
								},
							},
						},
					},
				},
			},
			{
				Entity: &diodepb.Entity_Prefix{
					Prefix: &diodepb.Prefix{
						Prefix: "192.168.0.0/32",
						Scope: &diodepb.Prefix_ScopeSite{
							ScopeSite: &diodepb.Site{
								Name: "test-site-name",
							},
						},
					},
				},
			},
			{
				Entity: &diodepb.Entity_ClusterGroup{
					ClusterGroup: &diodepb.ClusterGroup{
						Name: "test-cluster-group",
					},
				},
			},
			{
				Entity: &diodepb.Entity_ClusterType{
					ClusterType: &diodepb.ClusterType{
						Name: "test-cluster-type",
					},
				},
			},
			{
				Entity: &diodepb.Entity_Cluster{
					Cluster: &diodepb.Cluster{
						Name: "test-cluster",
					},
				},
			},
			{
				Entity: &diodepb.Entity_VirtualMachine{
					VirtualMachine: &diodepb.VirtualMachine{
						Name: "test-vm",
					},
				},
			},
			{
				Entity: &diodepb.Entity_VmInterface{
					VmInterface: &diodepb.VMInterface{
						Name: "test-vm-interface",
					},
				},
			},
			{
				Entity: &diodepb.Entity_VirtualDisk{
					VirtualDisk: &diodepb.VirtualDisk{
						Name: "test-virtual-disk",
					},
				},
			},
		},
	}
	reqBytes, err := proto.Marshal(ingestReq)
	assert.NoError(t, err)

	// Start processor in a separate goroutine
	go func() {
		err := processor.Start(ctx)
		assert.NoError(t, err)
	}()
	// Wait server
	time.Sleep(50 * time.Millisecond)

	mockRepository.On("CreateIngestionLog", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int32Ptr(1), nil)
	mockRepository.On("FindPriorIngestionLogByEntityHash", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, sql.ErrNoRows)
	mockRepository.On("UpdateIngestionLogStateWithError", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRepository.On("CreateChangeSet", mock.Anything, mock.Anything, mock.Anything).Return(int32Ptr(1), nil)

	mockNetBoxClient.On("GenerateDiff", mock.Anything, mock.Anything).Return(&netboxdiodeplugin.ChangeSetResult{
		ChangeSet: &netboxdiodeplugin.ChangeSet{
			ID: "test-changeset-id",
			Changes: []netboxdiodeplugin.Change{
				{
					ID:         "change-1",
					ChangeType: "create",
					ObjectType: "manufacturer",
					ObjectID:   intPtr(1),
				},
			},
		},
	}, nil)
	mockNetBoxClient.On("ApplyChangeSet", mock.Anything, mock.Anything).Return(&netboxdiodeplugin.ChangeSetResult{
		ChangeSet: &netboxdiodeplugin.ChangeSet{
			ID: "test-changeset-id",
		},
	}, nil)

	mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
	mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()
	mockMetrics.On("RecordChangeSetCreate", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("RecordChangeSetApply", mock.Anything, mock.Anything, mock.Anything).Return()

	// Add a message to the Redis stream
	metadata := []string{
		"request", string(reqBytes),
		"ingestion_ts", "1720425600",
	}
	streamID := reconciler.DefaultRedisStreamID
	err = redisStreamClient.XAdd(context.Background(), &redis.XAddArgs{
		Stream: streamID,
		Values: metadata,
	}).Err()
	assert.NoError(t, err)

	// Wait for the stream to be empty (message processed)
	timeout := time.After(2 * time.Second)
	for {
		streamLen, err := redisStreamClient.XLen(context.Background(), streamID).Result()
		assert.NoError(t, err)
		if streamLen == 0 {
			break
		}
		select {
		case <-timeout:
			t.Fatal("timeout waiting for stream to be empty")
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Stop the processor
	err = processor.Stop()
	assert.NoError(t, err)
	mockRepository.AssertExpectations(t)
	mockNetBoxClient.AssertExpectations(t)
}

func TestIngestionProcessor_DuplicateHandling(t *testing.T) {
	tests := []struct {
		name                  string
		existingLogState      reconcilerpb.State
		stateAfterChangeset   reconcilerpb.State
		makePrimary           bool
		expectSkipProcessing  bool
		expectReuseExistingID bool
		changeSetHasChanges   bool
	}{
		{
			name:                 "duplicate of IGNORED - skip processing",
			existingLogState:     reconcilerpb.State_IGNORED,
			expectSkipProcessing: true,
		},
		{
			name:                  "duplicate of QUEUED - reprocess existing",
			existingLogState:      reconcilerpb.State_QUEUED,
			stateAfterChangeset:   reconcilerpb.State_OPEN,
			expectSkipProcessing:  false,
			expectReuseExistingID: true,
			changeSetHasChanges:   true,
		},
		{
			name:                  "duplicate of OPEN - reprocess existing",
			existingLogState:      reconcilerpb.State_OPEN,
			stateAfterChangeset:   reconcilerpb.State_OPEN,
			expectSkipProcessing:  false,
			expectReuseExistingID: true,
			changeSetHasChanges:   true,
		},
		{
			name:                  "duplicate of FAILED - reprocess existing",
			existingLogState:      reconcilerpb.State_FAILED,
			stateAfterChangeset:   reconcilerpb.State_OPEN,
			expectSkipProcessing:  false,
			expectReuseExistingID: true,
			changeSetHasChanges:   true,
		},
		{
			name:                  "duplicate of APPLIED with changes - promote to primary",
			existingLogState:      reconcilerpb.State_APPLIED,
			stateAfterChangeset:   reconcilerpb.State_OPEN,
			expectSkipProcessing:  false,
			expectReuseExistingID: false,
			makePrimary:           true,
			changeSetHasChanges:   true,
		},
		{
			name:                  "duplicate of APPLIED without changes - no promotion",
			existingLogState:      reconcilerpb.State_APPLIED,
			stateAfterChangeset:   reconcilerpb.State_NO_CHANGES,
			expectSkipProcessing:  false,
			expectReuseExistingID: false,
			makePrimary:           false,
			changeSetHasChanges:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := miniredis.RunT(t)
			s.DB(1)
			defer s.Close()

			setupEnv(s.Addr())
			defer teardownEnv()
			var cfg reconciler.Config
			envconfig.MustProcess("", &cfg)

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

			mockOps := mocks.NewIngestionProcessorOps(t)
			mockMetrics := mocks.NewMetrics(t)

			redisClient := redis.NewClient(&redis.Options{Addr: s.Addr(), DB: 0})
			defer func() {
				_ = redisClient.Close()
			}()
			redisStreamClient := redis.NewClient(&redis.Options{Addr: s.Addr(), DB: 1})
			defer func() {
				_ = redisStreamClient.Close()
			}()

			testEntity := &diodepb.Entity{
				Entity: &diodepb.Entity_Site{
					Site: &diodepb.Site{
						Name: "test-site-name",
					},
				},
			}

			newLogID := int32(101)
			existingLogID := int32(100)

			existingLog := &reconcilerpb.IngestionLog{
				Id:         "existing-log-id",
				ObjectType: "dcim.site",
				State:      tt.existingLogState,
				Entity:     testEntity,
			}

			duplicateResult := &ops.CreateIngestionLogResult{
				Created: ops.IngestionLogRef{
					ID: newLogID,
					IngestionLog: &reconcilerpb.IngestionLog{
						Id:         "new-log-id",
						ObjectType: "dcim.site",
						State:      reconcilerpb.State_DUPLICATE,
						Entity:     testEntity,
					},
				},
				DuplicateOf: &ops.IngestionLogRef{
					ID:           existingLogID,
					IngestionLog: existingLog,
				},
			}

			mockOps.On("CreateIngestionLog", mock.Anything, mock.Anything, mock.Anything).Return(duplicateResult, nil)

			if !tt.expectSkipProcessing {
				changeSetLogID := newLogID
				ingestionLogForChangeset := duplicateResult.Created.IngestionLog
				if tt.expectReuseExistingID {
					changeSetLogID = existingLogID
					ingestionLogForChangeset = existingLog
				}

				changes := []changeset.Change{}
				if tt.changeSetHasChanges {
					changes = append(changes, changeset.Change{
						ID:         "test-change",
						ChangeType: "create",
						ObjectType: "dcim.site",
					})
				}

				mockChangeSet := &changeset.ChangeSet{
					ID:      "changeset-id",
					Changes: changes,
				}

				mockOps.On("GenerateChangeSet", mock.Anything, changeSetLogID, ingestionLogForChangeset, "").Run(func(_ mock.Arguments) {
					ingestionLogForChangeset.State = tt.stateAfterChangeset
					if tt.makePrimary {
						ingestionLogForChangeset.State = reconcilerpb.State_OPEN
					}
				}).Return(int32Ptr(1), mockChangeSet, nil)

				mockMetrics.On("RecordChangeSetCreate", mock.Anything, mock.Anything, mock.Anything).Return()

				if tt.stateAfterChangeset == reconcilerpb.State_OPEN {
					mockOps.On("ApplyChangeSet", mock.Anything, changeSetLogID, ingestionLogForChangeset, int32(1), mockChangeSet).Return(nil)
					mockMetrics.On("RecordChangeSetApply", mock.Anything, mock.Anything, mock.Anything).Return()
				}

				if tt.makePrimary {
					mockOps.On("ApplyChangeSet", mock.Anything, newLogID, ingestionLogForChangeset, int32(1), mockChangeSet).Return(nil)
					mockMetrics.On("RecordChangeSetApply", mock.Anything, mock.Anything, mock.Anything).Return()
				}
			}

			mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
			mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()

			processor, err := reconciler.NewIngestionProcessor(
				ctx, logger, cfg, redisClient, redisStreamClient,
				reconciler.DefaultRedisStreamID, reconciler.DefaultRedisConsumerGroup,
				mockOps, mockMetrics)
			require.NoError(t, err)

			go func() {
				err := processor.Start(ctx)
				assert.NoError(t, err)
			}()
			time.Sleep(50 * time.Millisecond)

			ingestReq := &diodepb.IngestRequest{
				Id:       "test-request-id",
				Entities: []*diodepb.Entity{testEntity},
			}
			reqBytes, err := proto.Marshal(ingestReq)
			require.NoError(t, err)

			streamID := reconciler.DefaultRedisStreamID
			err = redisStreamClient.XAdd(ctx, &redis.XAddArgs{
				Stream: streamID,
				Values: []string{
					"request", string(reqBytes),
					"ingestion_ts", "1720425600",
				},
			}).Err()
			require.NoError(t, err)

			timeout := time.After(2 * time.Second)
			for {
				streamLen, err := redisStreamClient.XLen(ctx, streamID).Result()
				require.NoError(t, err)
				if streamLen == 0 {
					break
				}
				select {
				case <-timeout:
					t.Fatal("timeout waiting for stream to be empty")
				default:
				}
				time.Sleep(100 * time.Millisecond)
			}

			err = processor.Stop()
			require.NoError(t, err)

			mockOps.AssertExpectations(t)
			mockMetrics.AssertExpectations(t)
		})
	}
}
