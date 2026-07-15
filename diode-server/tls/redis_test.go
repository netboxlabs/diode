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
			// Core fix: an empty username must leave Username empty rather than
			// being set to "". go-redis treats an empty Username as "use the
			// default Redis user", which is correct when no ACL user is set.
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
			// Core fix: a provided username MUST propagate to Username, else
			// go-redis authenticates as the "default" user and hits WRONGPASS
			// when that user is disabled or has a different password.
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
			opts, err := NewRedisOptions(RedisParams{
				Host:     tt.host,
				Port:     tt.port,
				Username: tt.username,
				Password: tt.password,
				DB:       tt.db,
				TLS:      tt.tlsCfg,
			})
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
