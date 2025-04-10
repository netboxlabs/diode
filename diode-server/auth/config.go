package auth

import "github.com/netboxlabs/diode/diode-server/telemetry"

// Config is the configuration for the auth service
type Config struct {
	HTTPPort int `envconfig:"HTTP_PORT" default:"8080"`

	Telemetry telemetry.Config `envconfig:"TELEMETRY"`
}
