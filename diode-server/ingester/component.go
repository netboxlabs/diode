package ingester

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"

	pb "github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/netbox"
)

const (
	streamID = "diode.v1.ingest-stream"

	consumerGroup = "diode-ingester"

	// RedisConsumerGroupExistsErrMsg is the error message returned by the redis client when the consumer group already exists
	RedisConsumerGroupExistsErrMsg = "BUSYGROUP Consumer Group name already exists"
)

// IngestEntityState represents the state of an ingested entity
type IngestEntityState int

const (
	// IngestEntityStateNew is the state of an entity after it has been ingested
	IngestEntityStateNew IngestEntityState = iota
)

// Component asynchronously ingests data from the distributor
type Component struct {
	ctx               context.Context
	config            Config
	logger            *slog.Logger
	hostname          string
	redisStreamClient *redis.Client
	redisClient       *redis.Client
}

// New creates a new ingester component
func New(ctx context.Context, logger *slog.Logger) (*Component, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	redisStreamClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisStreamDB,
	})

	if _, err := redisStreamClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisStreamClient.String(), err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisClient.String(), err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	return &Component{
		ctx:               ctx,
		config:            cfg,
		logger:            logger,
		hostname:          hostname,
		redisStreamClient: redisStreamClient,
		redisClient:       redisClient,
	}, nil
}

// Name returns the name of the component
func (c *Component) Name() string {
	return "ingester"
}

// Start starts the component
func (c *Component) Start(ctx context.Context) error {
	c.logger.Info("starting component", "name", c.Name())

	return c.consumeStream(ctx, streamID, consumerGroup, fmt.Sprintf("%s-%s", consumerGroup, c.hostname))
}

func (c *Component) consumeStream(ctx context.Context, stream, group, consumer string) error {
	err := c.redisStreamClient.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil && err.Error() != RedisConsumerGroupExistsErrMsg {
		return err
	}

	for {
		streams, err := c.redisStreamClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    100,
		}).Result()
		if err != nil || len(streams) == 0 {
			continue
		}
		for _, msg := range streams[0].Messages {
			if err := c.handleStreamMessage(ctx, msg); err != nil {
				c.logger.Error("failed to handle stream message", "error", err, "message", msg)
			}
		}
	}
}

func (c *Component) handleStreamMessage(ctx context.Context, msg redis.XMessage) error {
	c.logger.Info("received stream message", "message", msg.Values, "id", msg.ID)

	pushReq := &pb.PushRequest{}
	if err := proto.Unmarshal([]byte(msg.Values["request"].(string)), pushReq); err != nil {
		return err
	}

	errs := make([]string, 0)

	ingestionTs := msg.Values["ingestion_ts"]

	c.logger.Info("handling push request", "request", pushReq)

	for i, v := range pushReq.GetData() {
		if v.GetData() == nil {
			errs = append(errs, fmt.Sprintf("data for index %d is nil", i))
			continue
		}

		objectType, err := extractObjectType(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("failed to extract data type for index %d: %v", i, err))
			continue
		}

		key := fmt.Sprintf("ingest-entity:%s-%s-%s", objectType, ingestionTs, uuid.NewString())
		c.logger.Info("ingest entity key", "key", key)

		val := map[string]interface{}{
			"request_id":           pushReq.GetId(),
			"producer_app_name":    pushReq.GetProducerAppName(),
			"producer_app_version": pushReq.GetProducerAppVersion(),
			"sdk_name":             pushReq.GetSdkName(),
			"sdk_version":          pushReq.GetSdkVersion(),
			"data_type":            objectType,
			"data":                 v.GetData(),
			"ingestion_ts":         ingestionTs,
			"state":                IngestEntityStateNew,
		}
		encodedValue, err := json.Marshal(val)
		if err != nil {
			c.logger.Error("failed to marshal JSON", "value", val, "error", err)
			continue
		}
		if _, err = c.redisClient.Do(ctx, "JSON.SET", key, "$", encodedValue).Result(); err != nil {
			c.logger.Error("failed to set JSON redis key", "key", key, "value", encodedValue, "error", err)
			continue
		}
	}

	if len(errs) > 0 {
		c.logger.Error("failed to handle push request", "errors", strings.Join(errs, ", "))
	}

	c.redisStreamClient.XAck(ctx, streamID, consumerGroup, msg.ID)
	c.redisStreamClient.XDel(ctx, streamID, msg.ID)
	return nil
}

// Stop stops the component
func (c *Component) Stop() error {
	c.logger.Info("stopping component", "name", c.Name())

	redisStreamErr := c.redisStreamClient.Close()
	redisClientErr := c.redisClient.Close()

	return errors.Join(redisStreamErr, redisClientErr)
}

func extractObjectType(in *pb.IngestEntity) (string, error) {
	switch in.GetData().(type) {
	case *pb.IngestEntity_Device:
		return netbox.DcimDeviceTypeObjectType, nil
	case *pb.IngestEntity_DeviceRole:
		return netbox.DcimDeviceRoleObjectType, nil
	case *pb.IngestEntity_DeviceType:
		return netbox.DcimDeviceTypeObjectType, nil
	case *pb.IngestEntity_Interface:
		return netbox.DcimInterfaceObjectType, nil
	case *pb.IngestEntity_Manufacturer:
		return netbox.DcimManufacturerObjectType, nil
	case *pb.IngestEntity_Platform:
		return netbox.DcimPlatformObjectType, nil
	case *pb.IngestEntity_Site:
		return netbox.DcimSiteObjectType, nil
	default:
		return "", fmt.Errorf("unknown data type")
	}
}
