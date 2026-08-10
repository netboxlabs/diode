package tls

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisParams holds the connection parameters for building redis.Options.
type RedisParams struct {
	Host     string
	Port     string
	Username string
	Password string
	DB       int
	TLS      *Config
}

// NewRedisOptions builds redis.Options with consistent TLS and ACL username
// handling. All Redis client creation should use this helper to ensure the
// username is never accidentally omitted (which would cause go-redis to
// authenticate as the "default" Redis user instead of the intended ACL user).
func NewRedisOptions(p RedisParams) (redis.Options, error) {
	if p.TLS == nil {
		return redis.Options{}, fmt.Errorf("TLS config must not be nil (use &tls.Config{} for no TLS)")
	}

	goTLS, err := p.TLS.ToTLSConfig()
	if err != nil {
		return redis.Options{}, fmt.Errorf("failed to create TLS config for Redis: %w", err)
	}

	return redis.Options{
		Addr:      fmt.Sprintf("%s:%s", p.Host, p.Port),
		Username:  p.Username,
		Password:  p.Password,
		DB:        p.DB,
		TLSConfig: goTLS,
	}, nil
}
