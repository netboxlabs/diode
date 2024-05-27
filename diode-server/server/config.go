package server

// Config is the configuration for the server
type Config struct {
	Environment            string  `envconfig:"ENVIRONMENT" default:"development"`
	LoggingFormat          string  `envconfig:"LOGGING_FORMAT" default:"json"`
	LoggingLevel           string  `envconfig:"LOGGING_LEVEL" default:"info"`
	SentryDSN              string  `envconfig:"SENTRY_DSN"`
	SentryDebug            bool    `envconfig:"SENTRY_DEBUG" default:"false"`
	SentrySampleRate       float64 `envconfig:"SENTRY_SAMPLE_RATE" default:"1.0"`
	SentryEnableTracing    bool    `envconfig:"SENTRY_ENABLE_TRACING" default:"true"`
	SentryTracesSampleRate float64 `envconfig:"SENTRY_TRACES_SAMPLE_RATE" default:"1.0"`
	SentryAttachStacktrace bool    `envconfig:"SENTRY_ATTACH_STACKTRACE" default:"true"`
}
