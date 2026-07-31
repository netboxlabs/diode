package auth

import (
	"time"

	"github.com/netboxlabs/diode/diode-server/telemetry"
)

// Config is the configuration for the auth service
type Config struct {
	HTTPPort  int              `envconfig:"HTTP_PORT" default:"8080"`
	PProfAddr string           `envconfig:"PPROF_ADDR"`
	OAuth2    OAuth2Config     `envconfig:"OAUTH2"`
	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}

// OAuth2Config is the configuration for the OAuth2 server
type OAuth2Config struct {
	PublicServerURL string           `envconfig:"OAUTH2_PUBLIC_SERVER_URL" default:"http://localhost:4444"`
	AdminServerURL  string           `envconfig:"OAUTH2_ADMIN_SERVER_URL" default:"http://localhost:4445"`
	TokenCache      TokenCacheConfig `envconfig:"OAUTH2_TOKEN_CACHE"`
}

// TokenCacheConfig is the configuration for the client credentials token cache
// fronting the OAuth2 token endpoint.
type TokenCacheConfig struct {
	Enabled bool `envconfig:"OAUTH2_TOKEN_CACHE_ENABLED" default:"false"`
	// MaxEntries bounds the cache so that cycling client IDs cannot exhaust memory.
	MaxEntries int `envconfig:"OAUTH2_TOKEN_CACHE_MAX_ENTRIES" default:"4096"`
	// MaxTTL caps how long a token may be reused, independently of its own lifetime.
	// It bounds how long a rotated or deleted client secret keeps working from cache.
	MaxTTL time.Duration `envconfig:"OAUTH2_TOKEN_CACHE_MAX_TTL" default:"15m"`
	// NegativeTTL caps how long an upstream client rejection is remembered. Rejections
	// are keyed by the presented secret, so a caller with correct credentials hashes to
	// a different key and cannot be locked out.
	NegativeTTL time.Duration `envconfig:"OAUTH2_TOKEN_CACHE_NEGATIVE_TTL" default:"5s"`
}
