package reconciler

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

	"github.com/netboxlabs/diode/diode-sdk-go/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

const (
	redisStreamID = "diode.v1.ingest-stream"

	redisConsumerGroup = "diode-reconciler"

	// RedisConsumerGroupExistsErrMsg is the error message returned by the redis client when the consumer group already exists
	RedisConsumerGroupExistsErrMsg = "BUSYGROUP Consumer Group name already exists"
)

// IngestEntityState represents the state of an ingested entity
type IngestEntityState int

const (
	// IngestEntityStateNew is the state of an entity after it has been ingested
	IngestEntityStateNew IngestEntityState = iota

	// IngestEntityStateReconciled is the state of an entity after it has been reconciled
	IngestEntityStateReconciled

	// IngestEntityStateReconciliationFailed is the state of an entity after it has failed to be reconciled
	IngestEntityStateReconciliationFailed
)

// IngestionProcessor processes ingested data
type IngestionProcessor struct {
	config            Config
	logger            *slog.Logger
	hostname          string
	redisClient       *redis.Client
	redisStreamClient *redis.Client
	nbClient          *netboxdiodeplugin.Client
}

// NewIngestionProcessor creates a new ingestion processor
func NewIngestionProcessor(ctx context.Context, logger *slog.Logger) (*IngestionProcessor, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisClient.String(), err)
	}

	redisStreamClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisStreamDB,
	})

	if _, err := redisStreamClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed connection to %s: %v", redisStreamClient.String(), err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %v", err)
	}

	nbClient, err := netboxdiodeplugin.NewClient(logger, cfg.DiodeToNetBoxAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create netbox diode plugin client: %v", err)
	}

	component := &IngestionProcessor{
		config:            cfg,
		logger:            logger,
		hostname:          hostname,
		redisClient:       redisClient,
		redisStreamClient: redisStreamClient,
		nbClient:          nbClient,
	}

	return component, nil
}

// Name returns the name of the component
func (p *IngestionProcessor) Name() string {
	return "reconciler-ingestion-processor"
}

// Start starts the component
func (p *IngestionProcessor) Start(ctx context.Context) error {
	p.logger.Info("starting component", "name", p.Name())
	return p.consumeIngestionStream(ctx, redisStreamID, redisConsumerGroup, fmt.Sprintf("%s-%s", redisConsumerGroup, p.hostname))
}

// Stop stops the component
func (p *IngestionProcessor) Stop() error {
	p.logger.Info("stopping component", "name", p.Name())
	redisClientErr := p.redisClient.Close()
	redisStreamErr := p.redisStreamClient.Close()

	return errors.Join(redisStreamErr, redisClientErr)
}

func (p *IngestionProcessor) consumeIngestionStream(ctx context.Context, stream, group, consumer string) error {
	err := p.redisStreamClient.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil && err.Error() != RedisConsumerGroupExistsErrMsg {
		return err
	}

	for {
		streams, err := p.redisStreamClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    100,
		}).Result()
		if err != nil || len(streams) == 0 {
			continue
		}
		for _, msg := range streams[0].Messages {
			if err := p.handleStreamMessage(ctx, msg); err != nil {
				p.logger.Error("failed to handle stream message", "error", err, "message", msg)
				return err
			}
		}
	}
}

func (p *IngestionProcessor) handleStreamMessage(ctx context.Context, msg redis.XMessage) error {
	p.logger.Debug("received stream message", "message", msg.Values, "id", msg.ID)

	ingestReq := &diodepb.IngestRequest{}
	if err := proto.Unmarshal([]byte(msg.Values["request"].(string)), ingestReq); err != nil {
		return err
	}

	errs := make([]string, 0)

	ingestionTs := msg.Values["ingestion_ts"]

	p.logger.Debug("handling ingest request", "request", ingestReq)

	for i, v := range ingestReq.GetEntities() {
		if v.GetEntity() == nil {
			errs = append(errs, fmt.Sprintf("entity at index %d is nil", i))
			continue
		}

		objectType, err := extractObjectType(v)
		if err != nil {
			errs = append(errs, fmt.Sprintf("failed to extract data type for index %d: %v", i, err))
			continue
		}

		key := fmt.Sprintf("ingest-entity:%s-%s-%s", objectType, ingestionTs, uuid.NewString())
		p.logger.Debug("ingest entity key", "key", key)

		val := map[string]any{
			"request_id":           ingestReq.GetId(),
			"producer_app_name":    ingestReq.GetProducerAppName(),
			"producer_app_version": ingestReq.GetProducerAppVersion(),
			"sdk_name":             ingestReq.GetSdkName(),
			"sdk_version":          ingestReq.GetSdkVersion(),
			"data_type":            objectType,
			"entity":               v.GetEntity(),
			"ingestion_ts":         ingestionTs,
			"state":                IngestEntityStateNew,
		}

		encodedValue, err := p.writeJSON(ctx, key, val)
		if err != nil {
			errs = append(errs, fmt.Sprintf("failed to write JSON: %v", err))
			continue
		}

		changeSet, err := p.reconcileEntity(ctx, encodedValue)
		if err != nil {
			errs = append(errs, fmt.Sprintf("failed to reconcile entity: %v", err))
			val["state"] = IngestEntityStateReconciliationFailed
			_, err = p.writeJSON(ctx, key, val)
			if err != nil {
				errs = append(errs, fmt.Sprintf("failed to write JSON: %v", err))
			}
			continue
		}

		if changeSet != nil {
			val["state"] = IngestEntityStateReconciled
			_, err = p.writeJSON(ctx, key, val)
			if err != nil {
				errs = append(errs, fmt.Sprintf("failed to write JSON: %v", err))
				continue
			}
		}
	}

	if len(errs) > 0 {
		p.logger.Error("failed to handle ingest request", "errors", strings.Join(errs, ", "))
	}

	p.redisStreamClient.XAck(ctx, redisStreamID, redisConsumerGroup, msg.ID)

	if len(errs) > 0 {
		p.logger.Error("failed to handle ingest request", "errors", strings.Join(errs, ", "))
	} else {
		p.redisStreamClient.XDel(ctx, redisStreamID, msg.ID)
	}

	return nil
}

func (p *IngestionProcessor) reconcileEntity(ctx context.Context, encodedValue []byte) (*changeset.ChangeSet, error) {
	var ingestEntity changeset.IngestEntity
	_ = json.Unmarshal(encodedValue, &ingestEntity)

	ingestEntity.RequestID = uuid.NewString()

	cs, err := changeset.Prepare(ingestEntity, p.nbClient)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare change set: %v", err)
	}

	if len(cs.ChangeSet) == 0 {
		p.logger.Info("no changes to apply")
		return nil, nil
	}

	changes := make([]netboxdiodeplugin.Change, 0)
	for _, change := range cs.ChangeSet {
		changes = append(changes, netboxdiodeplugin.Change{
			ChangeID:      change.ChangeID,
			ChangeType:    change.ChangeType,
			ObjectType:    change.ObjectType,
			ObjectID:      change.ObjectID,
			ObjectVersion: change.ObjectVersion,
			Data:          change.Data,
		})
	}

	req := netboxdiodeplugin.ChangeSetRequest{
		ChangeSetID: cs.ChangeSetID,
		ChangeSet:   changes,
	}

	resp, err := p.nbClient.ApplyChangeSet(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to apply change set: %v", err)
	}

	p.logger.Debug("apply change set response", "response", resp)
	return cs, nil
}

func (p *IngestionProcessor) writeJSON(ctx context.Context, key string, value map[string]any) ([]byte, error) {
	encodedValue, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if _, err = p.redisClient.Do(ctx, "JSON.SET", key, "$", encodedValue).Result(); err != nil {
		return nil, fmt.Errorf("failed to set JSON redis key: %v", err)
	}

	return encodedValue, nil
}

func extractObjectType(in *diodepb.Entity) (string, error) {
	switch in.GetEntity().(type) {
	case *diodepb.Entity_Device:
		return netbox.DcimDeviceObjectType, nil
	case *diodepb.Entity_DeviceRole:
		return netbox.DcimDeviceRoleObjectType, nil
	case *diodepb.Entity_DeviceType:
		return netbox.DcimDeviceTypeObjectType, nil
	case *diodepb.Entity_Interface:
		return netbox.DcimInterfaceObjectType, nil
	case *diodepb.Entity_Manufacturer:
		return netbox.DcimManufacturerObjectType, nil
	case *diodepb.Entity_Platform:
		return netbox.DcimPlatformObjectType, nil
	case *diodepb.Entity_Site:
		return netbox.DcimSiteObjectType, nil
	default:
		return "", fmt.Errorf("unknown data type")
	}
}
