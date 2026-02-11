package tls

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NewRedisOptions builds redis.Options with consistent TLS and ACL username
// handling. All Redis client creation should use this helper to ensure the
// username is never accidentally omitted (which would cause go-redis to
// authenticate as the "default" Redis user instead of the intended ACL user).
func NewRedisOptions(host, port, username, password string, db int, tlsCfg *Config) (redis.Options, error) {
	if tlsCfg == nil {
		return redis.Options{}, fmt.Errorf("TLS config must not be nil (use &tls.Config{} for no TLS)")
	}

	goTLS, err := tlsCfg.ToTLSConfig()
	if err != nil {
		return redis.Options{}, fmt.Errorf("failed to create TLS config for Redis: %w", err)
	}

	opts := redis.Options{
		Addr:      fmt.Sprintf("%s:%s", host, port),
		Password:  password,
		DB:        db,
		TLSConfig: goTLS,
	}
	if username != "" {
		opts.Username = username
	}

	return opts, nil
}
