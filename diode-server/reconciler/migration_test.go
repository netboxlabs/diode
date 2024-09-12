package reconciler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mr "github.com/netboxlabs/diode/diode-server/reconciler/mocks"
)

func TestMigrate(t *testing.T) {
	tests := []struct {
		name              string
		appliedMigrations []MigrationLog
		err               error
	}{
		{
			name:              "no applied migrations found",
			appliedMigrations: nil,
			err:               nil,
		},
		{
			name:              "applied migrations found",
			appliedMigrations: []MigrationLog{{Name: "0001_initial", ApplyTs: time.Now().Unix()}},
			err:               nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRedisClient := new(mr.RedisClient)

			processor := &IngestionProcessor{
				redisClient: mockRedisClient,
				logger:      slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false})),
			}

			ctx := context.Background()

			if tt.appliedMigrations == nil {
				cmd := redis.NewCmd(ctx)
				if tt.err != nil {
					cmd.SetErr(errors.New("error"))
				} else {
					cmd.SetVal(nil)
					cmd.SetErr(nil)
				}
				mockRedisClient.On("Do", context.Background(), "JSON.GET", RedisDiodeMigrationsKey).Return(cmd)
				mockRedisClient.On("Do", context.Background(), "FT.DROPINDEX", RedisIngestEntityIndexName).Return(cmd)
				scanResults := []string{"ingest-entity:1", "ingest-entity:2", "ingest-entity:3"}
				mockRedisClient.On("Scan", context.Background(), uint64(0), "ingest-entity:*", int64(10)).Return(redis.NewScanCmdResult(scanResults, 0, nil))
				for _, key := range scanResults {
					mockRedisClient.On("Del", context.Background(), key).Return(redis.NewIntResult(0, nil))
				}
				mockRedisClient.On("Do", context.Background(),
					"FT.CREATE",
					RedisIngestEntityIndexName,
					"ON",
					"JSON",
					"PREFIX",
					"1",
					"ingest-entity:",
					"SCHEMA",
					"$.dataType",
					"AS",
					"data_type",
					"TEXT",
					"$.state",
					"AS",
					"state",
					"TEXT",
					"$.requestId",
					"AS",
					"request_id",
					"TEXT",
					"$.ingestionTs",
					"AS",
					"ingestion_ts",
					"NUMERIC",
					"SORTABLE",
				).Return(cmd)
				mockRedisClient.On("Do", context.Background(), "JSON.SET", RedisDiodeMigrationsKey, "$", mock.Anything).Return(cmd)
			} else {
				getAppliedMigrationsRespCmd := redis.NewCmd(ctx)
				appliedMigrationsJSON, _ := json.Marshal(tt.appliedMigrations)
				getAppliedMigrationsRespCmd.SetVal(string(appliedMigrationsJSON))
				getAppliedMigrationsRespCmd.SetErr(nil)
				mockRedisClient.On("Do", context.Background(), "JSON.GET", RedisDiodeMigrationsKey).Return(getAppliedMigrationsRespCmd)
				mockRedisClient.On("Do", context.Background(), "JSON.SET", RedisDiodeMigrationsKey, "$", appliedMigrationsJSON).Return(redis.NewCmd(ctx))
			}

			err := migrate(ctx, processor.logger, mockRedisClient)
			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
