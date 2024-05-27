package server

// Config is the configuration for the server
type Config struct {
	Environment   string `envconfig:"ENVIRONMENT" default:"development"`
	LoggingFormat string `envconfig:"LOGGING_FORMAT" default:"json"`
	LoggingLevel  string `envconfig:"LOGGING_LEVEL" default:"info"`
}
