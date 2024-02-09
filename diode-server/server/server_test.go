package server_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/netboxlabs/diode/diode-server/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		desc          string
		serverName    string
		loggingLevel  string
		loggingFormat string
	}{
		{
			desc:          "diode-test-server with debug level and json format",
			serverName:    "diode-test-server",
			loggingLevel:  "debug",
			loggingFormat: "json",
		},
		{
			desc:          "diode-test-server2 with debug level and text format",
			serverName:    "diode-test-server2",
			loggingLevel:  "debug",
			loggingFormat: "text",
		},
		{
			desc:          "diode-test-server with info level and json format",
			serverName:    "diode-test-server",
			loggingLevel:  "info",
			loggingFormat: "json",
		},
		{
			desc:          "diode-test-server with info level and text format",
			serverName:    "diode-test-server",
			loggingLevel:  "warn",
			loggingFormat: "json",
		},
		{
			desc:          "diode-test-server with error level and text format",
			serverName:    "diode-test-server",
			loggingLevel:  "error",
			loggingFormat: "text",
		},
		{
			desc:          "diode-test-server with error level and empty format",
			serverName:    "diode-test-server",
			loggingLevel:  "error",
			loggingFormat: "",
		},
		{
			desc:          "diode-test-server with empty level and text format",
			serverName:    "diode-test-server",
			loggingLevel:  "",
			loggingFormat: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx := context.Background()
			err := os.Setenv("LOGGING_LEVEL", tt.loggingLevel)
			require.NoError(t, err)
			err = os.Setenv("LOGGING_FORMAT", tt.loggingFormat)
			require.NoError(t, err)

			s := server.New(ctx, tt.serverName)

			assert.Equal(t, tt.serverName, s.Name())
			require.NotNil(t, s.Logger())
			//assert.True(t, s.Logger().Enabled(ctx, slog.LevelDebug))

			handlerOK := false
			if tt.loggingFormat == "text" {
				_, handlerOK = s.Logger().Handler().(*slog.TextHandler)
			} else {
				_, handlerOK = s.Logger().Handler().(*slog.JSONHandler)
			}
			assert.True(t, handlerOK)
		})
	}
}

func TestRegisterComponent(t *testing.T) {
	tests := []struct {
		desc             string
		registrationsNum int
		err              error
	}{
		{
			desc:             "registering a component",
			registrationsNum: 1,
			err:              nil,
		},
		{
			desc:             "registering a component twice",
			registrationsNum: 2,
			err:              fmt.Errorf("Server.RegisterComponent found duplicate component registration for noop"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx := context.Background()
			s := server.New(ctx, "diode-test-server")

			var err error
			for i := 0; i < tt.registrationsNum; i++ {
				err = s.RegisterComponent(&NoopComponent{})
			}
			if tt.err != nil {
				require.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		desc      string
		component server.Component
		err       error
	}{
		{
			desc:      "running a server with the NoopComponent",
			component: &NoopComponent{},
			err:       nil,
		},
		{
			desc:      "running a server with the FailingComponent",
			component: &FailingComponent{},
			err:       fmt.Errorf("start failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx := context.Background()
			s := server.New(ctx, "diode-test-server")

			require.NoError(t, s.RegisterComponent(tt.component))
			err := s.Run()

			if tt.err != nil {
				require.EqualError(t, err, tt.err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type NoopComponent struct{}

func (c *NoopComponent) Name() string {
	return "noop"
}

func (c *NoopComponent) Start(_ context.Context) error {
	return nil
}

func (c *NoopComponent) Stop() error {
	return nil
}

type FailingComponent struct{}

func (c *FailingComponent) Name() string {
	return "failing"
}

func (c *FailingComponent) Start(_ context.Context) error {
	return errors.New("start failed")
}

func (c *FailingComponent) Stop() error {
	return errors.New("stop failed")
}
