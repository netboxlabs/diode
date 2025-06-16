package reconciler_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler"
	"github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func int32Ptr(i int32) *int32 { return &i }

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

	expectedToken := "mocked-token"
	authTokenURL := "/diode/auth/token"
	mockOAuth2Server := newMockOAuth2Server(authTokenURL, cfg.DiodeToNetBoxClientID, cfg.DiodeToNetBoxClientSecret, expectedToken)
	defer mockOAuth2Server.Close()

	mockOAuth2ServerURL := mockOAuth2Server.URL + authTokenURL

	nbClient, err := netboxdiodeplugin.NewClient(
		netboxdiodeplugin.ClientOptions{
			Logger:            logger,
			BaseURL:           cfg.NetBoxDiodePluginAPIBaseURL,
			ClientID:          cfg.DiodeToNetBoxClientID,
			ClientSecret:      cfg.DiodeToNetBoxClientSecret,
			TokenURL:          mockOAuth2ServerURL,
			RateLimitRPS:      cfg.DiodeToNetBoxRateLimiterRPS,
			RateLimitBurstRPS: cfg.DiodeToNetBoxRateLimiterBurst,
			MaxRetries:        0,
		})
	require.NoError(t, err)
	metrics := mocks.NewIngestionProcessorMetrics(t)

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

	processor, err := reconciler.NewIngestionProcessor(ctx, logger, cfg, redisClient, redisStreamClient, reconciler.DefaultRedisStreamID, reconciler.DefaultRedisConsumerGroup, reconciler.NewOps(mockRepository, nbClient, logger), metrics)
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

	expectedToken := "mocked-token"
	authTokenURL := "/diode/auth/token"
	mockOAuth2Server := newMockOAuth2Server(authTokenURL, cfg.DiodeToNetBoxClientID, cfg.DiodeToNetBoxClientSecret, expectedToken)
	defer mockOAuth2Server.Close()

	mockOAuth2ServerURL := mockOAuth2Server.URL + authTokenURL

	maxRetries := 3
	nbClient, err := netboxdiodeplugin.NewClient(
		netboxdiodeplugin.ClientOptions{
			Logger:            logger,
			BaseURL:           cfg.NetBoxDiodePluginAPIBaseURL,
			ClientID:          cfg.DiodeToNetBoxClientID,
			ClientSecret:      cfg.DiodeToNetBoxClientSecret,
			TokenURL:          mockOAuth2ServerURL,
			RateLimitRPS:      cfg.DiodeToNetBoxRateLimiterRPS,
			RateLimitBurstRPS: cfg.DiodeToNetBoxRateLimiterBurst,
			MaxRetries:        maxRetries,
		})
	require.NoError(t, err)
	mockMetrics := new(mocks.IngestionProcessorMetrics)
	mockMetrics.On("RecordHandleMessage", mock.Anything, mock.Anything).Return()
	mockMetrics.On("RecordIngestionLogCreate", mock.Anything, mock.Anything).Return()
	mockMetrics.On("RecordChangeSetCreate", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("RecordChangeSetApply", mock.Anything, mock.Anything, mock.Anything).Return()

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

	processor, err := reconciler.NewIngestionProcessor(ctx, logger, cfg, redisClient, redisStreamClient, reconciler.DefaultRedisStreamID, reconciler.DefaultRedisConsumerGroup, reconciler.NewOps(mockRepository, nbClient, logger), mockMetrics)
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

	mockRepository.On("CreateIngestionLog", mock.Anything, mock.Anything, mock.Anything).Return(int32Ptr(1), nil)
	mockRepository.On("UpdateIngestionLogStateWithError", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRepository.On("CreateChangeSet", mock.Anything, mock.Anything, mock.Anything).Return(int32Ptr(1), nil)

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
	for {
		streamLen, err := redisStreamClient.XLen(context.Background(), streamID).Result()
		assert.NoError(t, err)
		if streamLen == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Stop the processor
	err = processor.Stop()
	assert.NoError(t, err)
	mockRepository.AssertExpectations(t)
}

func newMockOAuth2Server(authTokenURL, wantClientID, wantClientSecret, mockedToken string) *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc(authTokenURL, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		// Optional: Validate client credentials
		if r.PostForm.Get("client_id") != wantClientID || r.PostForm.Get("client_secret") != wantClientSecret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			if err := json.NewEncoder(w).Encode(map[string]string{
				"error":             "unauthorized",
				"error_description": "Authentication required",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		// Simulate token response
		resp := map[string]any{
			"access_token": mockedToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return httptest.NewServer(handler)
}
