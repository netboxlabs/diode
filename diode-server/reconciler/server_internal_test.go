package reconciler

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
)

func TestIsAuthenticated(t *testing.T) {
	tests := []struct {
		name            string
		rpcMethod       string
		authorization   []string
		apiKeys         map[string]string
		isAuthenticated bool
	}{
		{
			name:          "missing authorization header",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{},
			apiKeys: map[string]string{
				"INGESTER_TO_RECONCILER": "test",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion data sources with valid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"INGESTER_TO_RECONCILER": "test",
			},
			isAuthenticated: true,
		},
		{
			name:          "retrieve ingestion data sources with invalid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{"test0"},
			apiKeys: map[string]string{
				"INGESTER_TO_RECONCILER": "test",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion data sources for server without api key configured",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionDataSources_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"DIODE": "foorbar",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion logs with valid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionLogs_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"NETBOX_TO_DIODE": "test",
			},
			isAuthenticated: true,
		},
		{
			name:          "retrieve ingestion logs with invalid authorization",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionLogs_FullMethodName,
			authorization: []string{"test0"},
			apiKeys: map[string]string{
				"NETBOX_TO_DIODE": "test",
			},
			isAuthenticated: false,
		},
		{
			name:          "retrieve ingestion logs for server without api key configured",
			rpcMethod:     reconcilerpb.ReconcilerService_RetrieveIngestionLogs_FullMethodName,
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"DIODE": "foorbar",
			},
			isAuthenticated: false,
		},
		{
			name:          "authorization for unknown rpc method",
			rpcMethod:     "/diode.v1.ReconcilerService/UnknownMethod",
			authorization: []string{"test"},
			apiKeys: map[string]string{
				"DIODE": "foorbar",
			},
			isAuthenticated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			assert.Equal(t, tt.isAuthenticated, isAuthenticated(logger, tt.rpcMethod, tt.apiKeys, tt.authorization))
		})
	}
}
