package reconciler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// RedisDiodeMigrationsKey is the key for the redis diode migrations
	RedisDiodeMigrationsKey = "diode.migrations"
)

// AppliedMigrations is a list of applied migrations
type AppliedMigrations []MigrationLog

// MigrationLog is a log of a migration
type MigrationLog struct {
	Name    string `json:"name"`
	ApplyTs int64  `json:"apply_ts"`
}

type migration struct {
	name string
	run  func(context.Context, *slog.Logger, RedisClient) error
}

func migrate(ctx context.Context, logger *slog.Logger, redisClient RedisClient) error {
	res, err := redisClient.Do(ctx, "JSON.GET", RedisDiodeMigrationsKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("failed to get JSON redis key %s: %v", RedisDiodeMigrationsKey, err)
	}

	var appliedMigrations AppliedMigrations
	if res != nil {
		_ = json.Unmarshal([]byte(res.(string)), &appliedMigrations)
	}

	logger.Debug("migrations", "appliedMigrations", appliedMigrations)

	if len(appliedMigrations) == 0 {
		logger.Debug("no applied migrations found")
	}

	migrations := []migration{
		{
			name: "0001_initial",
			run:  initialMigration(),
		},
	}

	for _, m := range migrations {
		var found bool
		for _, am := range appliedMigrations {
			if am.Name == m.name {
				found = true
				break
			}
		}

		if !found {
			logger.Debug("applying migration", "name", m.name)

			if err := m.run(ctx, logger, redisClient); err != nil {
				return fmt.Errorf("failed to run migration %s: %v", m.name, err)
			}

			logger.Debug("migration applied", "name", m.name)

			appliedMigrations = append(appliedMigrations, MigrationLog{
				Name:    m.name,
				ApplyTs: time.Now().UnixNano(),
			})
		}
	}

	appliedMigrationsJSON, err := json.Marshal(appliedMigrations)
	if err != nil {
		return fmt.Errorf("failed to marshal applied migrations %#v: %v", appliedMigrations, err)
	}

	if _, err = redisClient.Do(ctx, "JSON.SET", RedisDiodeMigrationsKey, "$", appliedMigrationsJSON).Result(); err != nil {
		return fmt.Errorf("failed to set JSON redis key %s with value %s: %v", RedisDiodeMigrationsKey, appliedMigrationsJSON, err)
	}

	return nil
}

func initialMigration() func(context.Context, *slog.Logger, RedisClient) error {
	return func(ctx context.Context, logger *slog.Logger, redisClient RedisClient) error {
		// Drop FT index ingest-entity due to schema change
		logger.Debug("dropping index", "name", RedisIngestEntityIndexName)
		_, err := redisClient.Do(ctx, "FT.DROPINDEX", RedisIngestEntityIndexName).Result()
		if err != nil && !errors.Is(err, redis.Nil) && err.Error() != "Unknown Index name" {
			return fmt.Errorf("failed to drop FT index %s: %v", RedisIngestEntityIndexName, err)
		}

		// Delete all keys with prefix ingest-entity
		logger.Debug("deleting keys with prefix", "prefix", "ingest-entity:*")
		iter := redisClient.Scan(ctx, 0, "ingest-entity:*", 10).Iterator()
		for iter.Next(ctx) {
			if err := redisClient.Del(ctx, iter.Val()).Err(); err != nil {
				return fmt.Errorf("failed to delete key %s: %v", iter.Val(), err)
			}
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("failed to iterate over keys with prefix %s: %v", RedisIngestEntityIndexName, err)
		}

		// Create new FT index ingest-entity
		logger.Debug("creating index", "name", RedisIngestEntityIndexName)
		queryArgs := []interface{}{
			"FT.CREATE",
			RedisIngestEntityIndexName,
			"ON",
			"JSON",
			"PREFIX",
			"1",
			"ingest-entity:",
			"SCHEMA",
			"$.id",
			"AS",
			"id",
			"TEXT",
			"SORTABLE",
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
		}

		if _, err = redisClient.Do(ctx, queryArgs...).Result(); err != nil {
			return fmt.Errorf("failed to create FT index %s: %v", RedisIngestEntityIndexName, err)
		}

		return nil
	}
}
