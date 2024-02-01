package server

type Config struct {
	LoggingFormat string `envconfig:"LOGGING_FORMAT" default:"json"`
	LoggingLevel  string `envconfig:"LOGGING_LEVEL" default:"info"`
}
