package reconciler_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode-sdk-go/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/reconciler"
)

func TestNewIngestionProcessor(t *testing.T) {
	ctx := context.Background()
	s := miniredis.RunT(t)
	defer s.Close()

	setupEnv(s.Addr())
	defer teardownEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	processor, err := reconciler.NewIngestionProcessor(ctx, logger)
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := context.Background()

	processor, err := reconciler.NewIngestionProcessor(ctx, logger)
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
					DeviceRole: &diodepb.Role{
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
						Name: "test-device-name",
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
							Name: "test-device-name",
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
						AssignedObject: &diodepb.IPAddress_Interface{
							Interface: &diodepb.Interface{
								Name: "test-interface",
								Device: &diodepb.Device{
									Name: "test-device-name",
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
						Site: &diodepb.Site{
							Name: "test-site-name",
						},
					},
				},
			},
		},
	}
	reqBytes, err := proto.Marshal(ingestReq)
	assert.NoError(t, err)

	// Start processor in a separate goroutine
	go func() {
		err = processor.Start(ctx)
		assert.NoError(t, err)
	}()
	// Wait server
	time.Sleep(50 * time.Millisecond)

	redisClient := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
		DB:   1,
	})

	// Add a message to the Redis stream
	metadata := []string{
		"request", string(reqBytes),
		"ingestion_ts", "1720425600",
	}
	err = redisClient.XAdd(context.Background(), &redis.XAddArgs{
		Stream: "diode.v1.ingest-stream",
		Values: metadata,
	}).Err()
	assert.NoError(t, err)

	// Wait for the message to be processed
	time.Sleep(100 * time.Millisecond)

	// Stop the processor
	err = processor.Stop()
	assert.NoError(t, err)
}
