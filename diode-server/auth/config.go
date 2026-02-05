package auth

import "github.com/netboxlabs/diode/diode-server/telemetry"

// Config is the configuration for the auth service
type Config struct {
	HTTPPort  int              `envconfig:"HTTP_PORT" default:"8080"`
	PProfAddr string           `envconfig:"AUTH_PPROF_ADDR"`
	OAuth2    OAuth2Config     `envconfig:"OAUTH2"`
	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}

// OAuth2Config is the configuration for the OAuth2 server
type OAuth2Config struct {
	PublicServerURL string `envconfig:"OAUTH2_PUBLIC_SERVER_URL" default:"http://localhost:4444"`
	AdminServerURL  string `envconfig:"OAUTH2_ADMIN_SERVER_URL" default:"http://localhost:4445"`
}
