package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config is the configuration for SSL/TLS settings
// Note the environment variables may be prefixed with REDIS_TLS_
// based on the application's configuration structure
type Config struct {
	Enabled        bool   `envconfig:"ENABLED" default:"false"`
	SkipVerify     bool   `envconfig:"SKIP_VERIFY" default:"false"`
	CaPath         string `envconfig:"CA_PATH" default:""`
	ClientKeyPath  string `envconfig:"CLIENT_KEY_PATH" default:""`
	ClientCertPath string `envconfig:"CLIENT_CERT_PATH" default:""`
}

// ToTLSConfig converts the Diode TLS config to a Go Crypto TLS config
func (cfg *Config) ToTLSConfig() (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	// Load CA certificate if provided
	if cfg.CaPath != "" {
		caCert, err := os.ReadFile(cfg.CaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append CA certificate to pool")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if provided
	if cfg.ClientCertPath != "" && cfg.ClientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate and key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
