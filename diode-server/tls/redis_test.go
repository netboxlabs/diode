package tls

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisOptions(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     string
		username string
		password string
		db       int
		tlsCfg   *Config
		wantAddr string
		wantUser string
		wantPass string
		wantDB   int
		wantTLS  bool
		wantErr  bool
	}{
		{
			name:     "basic options without TLS",
			host:     "localhost",
			port:     "6379",
			username: "",
			password: "secret",
			db:       0,
			tlsCfg:   &Config{Enabled: false},
			wantAddr: "localhost:6379",
			wantUser: "",
			wantPass: "secret",
			wantDB:   0,
			wantTLS:  false,
		},
		{
			name:     "with ACL username",
			host:     "redis.example.com",
			port:     "6380",
			username: "redis-user",
			password: "secret",
			db:       1,
			tlsCfg:   &Config{Enabled: false},
			wantAddr: "redis.example.com:6380",
			wantUser: "redis-user",
			wantPass: "secret",
			wantDB:   1,
			wantTLS:  false,
		},
		{
			name:     "with TLS enabled",
			host:     "redis.example.com",
			port:     "6380",
			username: "redis-user",
			password: "secret",
			db:       0,
			tlsCfg:   &Config{Enabled: true, SkipVerify: true},
			wantAddr: "redis.example.com:6380",
			wantUser: "redis-user",
			wantPass: "secret",
			wantDB:   0,
			wantTLS:  true,
		},
		{
			name:     "empty username is not set",
			host:     "localhost",
			port:     "6379",
			username: "",
			password: "secret",
			db:       0,
			tlsCfg:   &Config{Enabled: false},
			wantAddr: "localhost:6379",
			wantUser: "",
			wantPass: "secret",
			wantDB:   0,
			wantTLS:  false,
		},
		{
			name:     "nil TLS config returns error",
			host:     "localhost",
			port:     "6379",
			username: "",
			password: "secret",
			db:       0,
			tlsCfg:   nil,
			wantErr:  true,
		},
		{
			name:     "invalid CA path returns error",
			host:     "localhost",
			port:     "6379",
			username: "",
			password: "secret",
			db:       0,
			tlsCfg:   &Config{Enabled: true, CaPath: "/nonexistent/ca.crt"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := NewRedisOptions(tt.host, tt.port, tt.username, tt.password, tt.db, tt.tlsCfg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantAddr, opts.Addr)
			assert.Equal(t, tt.wantUser, opts.Username)
			assert.Equal(t, tt.wantPass, opts.Password)
			assert.Equal(t, tt.wantDB, opts.DB)

			if tt.wantTLS {
				assert.NotNil(t, opts.TLSConfig, "TLSConfig should be set when TLS is enabled")
			} else {
				assert.Nil(t, opts.TLSConfig, "TLSConfig should be nil when TLS is disabled")
			}
		})
	}
}

func TestNewRedisOptionsUsernameNotSetWhenEmpty(t *testing.T) {
	// This test specifically verifies the core fix: when username is empty,
	// the Username field on redis.Options must remain empty ("") rather than
	// being set to "". go-redis treats an empty Username as "use the default
	// Redis user", which is the correct behavior when no ACL username is
	// configured.
	opts, err := NewRedisOptions("localhost", "6379", "", "pass", 0, &Config{})
	require.NoError(t, err)
	assert.Empty(t, opts.Username)
}

func TestNewRedisOptionsUsernameSetWhenProvided(t *testing.T) {
	// This test verifies the core fix: when a username IS provided, it MUST
	// be propagated to redis.Options.Username. Missing this causes go-redis
	// to authenticate as the "default" Redis user instead of the intended
	// ACL user, resulting in WRONGPASS errors when the default user is
	// disabled or has a different password.
	opts, err := NewRedisOptions("localhost", "6379", "my-acl-user", "pass", 0, &Config{})
	require.NoError(t, err)
	assert.Equal(t, "my-acl-user", opts.Username)
}
