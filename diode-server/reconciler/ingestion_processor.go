package reconciler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"

	"github.com/andybalholm/brotli"
	"github.com/kelseyhightower/envconfig"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
	"github.com/netboxlabs/diode/diode-server/sentry"
)

const (
	redisStreamID = "diode.v1.ingest-stream"

	redisConsumerGroup = "diode-reconciler"

	// RedisIngestEntityIndexName is the name of the redis index for ingest entities
	RedisIngestEntityIndexName = "ingest-entity"

	// RedisConsumerGroupExistsErrMsg is the error message returned by the redis client when the consumer group already exists
	RedisConsumerGroupExistsErrMsg = "BUSYGROUP Consumer Group name already exists"
)

// RedisClient is an interface that represents the methods used from redis.Client
type RedisClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) *redis.StatusCmd
	XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd
	XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd
	XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Pipeline() redis.Pipeliner
}

// IngestionProcessor processes ingested data
type IngestionProcessor struct {
	config            Config
	logger            *slog.Logger
	hostname          string
	redisClient       RedisClient
	redisStreamClient RedisClient
	nbClient          netboxdiodeplugin.NetBoxAPI
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

	if p.config.MigrationEnabled {
		if err := migrate(ctx, p.logger, p.redisClient); err != nil {
			return fmt.Errorf("failed to migrate: %v", err)
		}
	}

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

				contextMap := map[string]any{
					"redis_stream_msg_id": msg.ID,
					"consumer":            consumer,
					"hostname":            p.hostname,
				}
				sentry.CaptureError(fmt.Errorf("failed to handle stream message: %v", err), nil, "Ingestion stream", contextMap)

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

	errs := make([]error, 0)

	ingestionTs, err := strconv.Atoi(msg.Values["ingestion_ts"].(string))
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to convert ingestion timestamp: %v", err))
	}

	p.logger.Debug("handling ingest request", "request", ingestReq)

	for i, v := range ingestReq.GetEntities() {
		if v.GetEntity() == nil {
			errs = append(errs, fmt.Errorf("entity at index %d is nil", i))
			continue
		}

		objectType, err := extractObjectType(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to extract data type for index %d: %v", i, err))
			continue
		}

		ingestionLogID := ksuid.New().String()

		key := fmt.Sprintf("ingest-entity:%s-%d-%s", objectType, ingestionTs, ingestionLogID)
		p.logger.Debug("ingest entity key", "key", key)

		ingestionLog := &reconcilerpb.IngestionLog{
			Id:                 ingestionLogID,
			RequestId:          ingestReq.GetId(),
			ProducerAppName:    ingestReq.GetProducerAppName(),
			ProducerAppVersion: ingestReq.GetProducerAppVersion(),
			SdkName:            ingestReq.GetSdkName(),
			SdkVersion:         ingestReq.GetSdkVersion(),
			DataType:           objectType,
			Entity:             v,
			IngestionTs:        int64(ingestionTs),
			State:              reconcilerpb.State_NEW,
		}

		if _, err = p.writeIngestionLog(ctx, key, ingestionLog); err != nil {
			errs = append(errs, fmt.Errorf("failed to write JSON: %v", err))
			continue
		}

		ingestEntity := changeset.IngestEntity{
			RequestID: ingestReq.GetId(),
			DataType:  objectType,
			Entity:    v.GetEntity(),
			State:     int(reconcilerpb.State_NEW),
		}

		changeSet, err := p.reconcileEntity(ctx, ingestEntity)
		if err != nil {
			errs = append(errs, err)

			ingestionLog.State = reconcilerpb.State_FAILED
			ingestionLog.Error = extractIngestionError(err)

			if changeSet != nil {
				ingestionLog.ChangeSet = &reconcilerpb.ChangeSet{Id: changeSet.ChangeSetID}
				csCompressed, err := compressChangeSet(changeSet)
				if err != nil {
					errs = append(errs, err)
				} else {
					ingestionLog.ChangeSet.Data = csCompressed
				}
			}

			if _, err = p.writeIngestionLog(ctx, key, ingestionLog); err != nil {
				errs = append(errs, err)
			}
			continue
		}

		if changeSet != nil {
			ingestionLog.State = reconcilerpb.State_RECONCILED
			ingestionLog.ChangeSet = &reconcilerpb.ChangeSet{Id: changeSet.ChangeSetID}
			csCompressed, err := compressChangeSet(changeSet)
			if err != nil {
				errs = append(errs, err)
			} else {
				ingestionLog.ChangeSet.Data = csCompressed
			}
		} else {
			ingestionLog.State = reconcilerpb.State_NO_CHANGES
		}

		if _, err = p.writeIngestionLog(ctx, key, ingestionLog); err != nil {
			errs = append(errs, fmt.Errorf("failed to write JSON: %v", err))
			continue
		}
	}

	p.redisStreamClient.XAck(ctx, redisStreamID, redisConsumerGroup, msg.ID)

	if len(errs) > 0 {
		errsStr := make([]string, 0)
		for _, err := range errs {
			errsStr = append(errsStr, err.Error())
		}
		p.logger.Warn("failed to handle ingest request", slog.String("request_id", ingestReq.Id), slog.Any("errors", errsStr))

		contextMap := map[string]any{
			"redis_stream_msg_id": msg.ID,
			"consumer":            fmt.Sprintf("%s-%s", redisConsumerGroup, p.hostname),
			"hostname":            p.hostname,
		}
		sentry.CaptureError(fmt.Errorf("failed to handle ingest request: %v", errs), nil, "Ingestion request", contextMap)
	} else {
		p.redisStreamClient.XDel(ctx, redisStreamID, msg.ID)
	}

	return nil
}

func extractIngestionError(err error) *reconcilerpb.IngestionError {
	var ingestionErr *reconcilerpb.IngestionError
	var applyChangeSetErr *netboxdiodeplugin.ApplyChangeSetError

	switch {
	case errors.As(err, &applyChangeSetErr):
		ingestionErr = applyChangeSetErr.ToIngestionError()
	default:
		ingestionErr = &reconcilerpb.IngestionError{
			Message: err.Error(),
			Code:    0,
		}
	}

	return ingestionErr
}

func (p *IngestionProcessor) reconcileEntity(ctx context.Context, ingestEntity changeset.IngestEntity) (*changeset.ChangeSet, error) {
	cs, err := changeset.Prepare(ingestEntity, p.nbClient)
	if err != nil {
		tags := map[string]string{
			"request_id": ingestEntity.RequestID,
		}
		contextMap := map[string]any{
			"request_id": ingestEntity.RequestID,
			"data_type":  ingestEntity.DataType,
		}
		sentry.CaptureError(err, tags, "Ingest Entity", contextMap)
		return nil, fmt.Errorf("failed to prepare change set: %v", err)
	}

	if len(cs.ChangeSet) == 0 {
		p.logger.Debug("no changes to apply", "request_id", ingestEntity.RequestID)
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
		return cs, err
	}

	p.logger.Debug("apply change set response", "response", resp)
	return cs, nil
}

func (p *IngestionProcessor) writeIngestionLog(ctx context.Context, key string, ingestionLog *reconcilerpb.IngestionLog) ([]byte, error) {
	ingestionLogJSON, err := protojson.Marshal(ingestionLog)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	ingestionLogJSON = normalizeIngestionLog(ingestionLogJSON)

	if _, err := p.redisClient.Do(ctx, "JSON.SET", key, "$", ingestionLogJSON).Result(); err != nil {
		return nil, fmt.Errorf("failed to set JSON redis key: %v", err)
	}

	return ingestionLogJSON, nil
}

func normalizeIngestionLog(l []byte) []byte {
	//replace ingestionTs string value as integer, see: https://github.com/golang/protobuf/issues/1414
	re := regexp.MustCompile(`"ingestionTs":"(\d+)"`)
	return re.ReplaceAll(l, []byte(`"ingestionTs":$1`))
}

func compressChangeSet(cs *changeset.ChangeSet) (string, error) {
	csJSON, err := json.Marshal(cs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	var brotliBuf bytes.Buffer
	brotliWriter := brotli.NewWriter(&brotliBuf)
	brotliWriter.Write(csJSON)
	brotliWriter.Close()

	return base64.StdEncoding.EncodeToString(brotliBuf.Bytes()), nil
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
	case *diodepb.Entity_IpAddress:
		return netbox.IpamIPAddressObjectType, nil
	case *diodepb.Entity_Prefix:
		return netbox.IpamPrefixObjectType, nil
	case *diodepb.Entity_ClusterGroup:
		return netbox.VirtualizationClusterGroupObjectType, nil
	case *diodepb.Entity_ClusterType:
		return netbox.VirtualizationClusterTypeObjectType, nil
	case *diodepb.Entity_Cluster:
		return netbox.VirtualizationClusterObjectType, nil
	case *diodepb.Entity_VirtualMachine:
		return netbox.VirtualizationVirtualMachineObjectType, nil
	case *diodepb.Entity_Vminterface:
		return netbox.VirtualizationVMInterfaceObjectType, nil
	case *diodepb.Entity_VirtualDisk:
		return netbox.VirtualizationVirtualDiskObjectType, nil
	default:
		return "", fmt.Errorf("unknown data type")
	}
}
