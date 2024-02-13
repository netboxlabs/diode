package ingester

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
)

const (
	streamID = "diode.v1.ingest-stream"

	consumerGroup = "diode-ingester"

	// RedisConsumerGroupExistsErrMsg is the error message returned by the redis client when the consumer group already exists
	RedisConsumerGroupExistsErrMsg = "BUSYGROUP Consumer Group name already exists"
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
				return err
			}
		}
	}
}

func (c *Component) handleStreamMessage(ctx context.Context, msg redis.XMessage) error {
	c.logger.Info("received message in stream", "message", msg.Values, "id", msg.ID)

	c.redisStreamClient.XAck(ctx, streamID, consumerGroup, msg.ID)
	return nil
}

// Stop stops the component
func (c *Component) Stop() error {
	c.logger.Info("stopping component", "name", c.Name())
	return c.redisClient.Close()
}
