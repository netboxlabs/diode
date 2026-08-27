package cli

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/auth"
	"github.com/netboxlabs/diode/diode-server/auth/mocks"
)

const testCredentials = `[
  {
    "client_id": "diode-ingest",
    "client_secret": "ingest-secret-0123456789",
    "grant_types": ["client_credentials"],
    "scope": "diode:ingest"
  },
  {
    "client_id": "netbox-to-diode",
    "client_secret": "netbox-secret-0123456789",
    "grant_types": ["client_credentials"],
    "scope": "diode:read diode:write"
  }
]`

func writeCredentials(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "client-credentials.json")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

// newCommand builds a command with the readiness wait disabled, so tests only
// exercise the reconciliation itself.
func newCommand(clientManager auth.ClientManager) *BootstrapClientsCommand {
	return &BootstrapClientsCommand{clientManager: clientManager}
}

func notFound(clientID string) error {
	return auth.NewAuthError("client "+clientID+" not found", http.StatusNotFound)
}

func TestBootstrapCreatesMissingClients(t *testing.T) {
	clientManager := mocks.NewClientManager(t)

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, "diode-ingest").
		Return(auth.ClientInfo{}, notFound("diode-ingest")).Once()
	clientManager.EXPECT().CreateClient(mock.Anything, auth.ClientInfo{
		ClientID:     "diode-ingest",
		ClientSecret: "ingest-secret-0123456789",
		Scope:        "diode:ingest",
	}).Return(auth.ClientInfo{ClientID: "diode-ingest"}, nil).Once()

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, "netbox-to-diode").
		Return(auth.ClientInfo{}, notFound("netbox-to-diode")).Once()
	clientManager.EXPECT().CreateClient(mock.Anything, auth.ClientInfo{
		ClientID:     "netbox-to-diode",
		ClientSecret: "netbox-secret-0123456789",
		Scope:        "diode:read diode:write",
	}).Return(auth.ClientInfo{ClientID: "netbox-to-diode"}, nil).Once()

	err := newCommand(clientManager).bootstrap(context.Background(), writeCredentials(t, testCredentials))
	require.NoError(t, err)
}

func TestBootstrapSkipsExistingClients(t *testing.T) {
	clientManager := mocks.NewClientManager(t)

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, "diode-ingest").
		Return(auth.ClientInfo{ClientID: "diode-ingest"}, nil).Once()
	clientManager.EXPECT().RetrieveClientByID(mock.Anything, "netbox-to-diode").
		Return(auth.ClientInfo{ClientID: "netbox-to-diode"}, nil).Once()

	err := newCommand(clientManager).bootstrap(context.Background(), writeCredentials(t, testCredentials))
	require.NoError(t, err)
	// CreateClient is never asserted on the mock, so any call would fail the test.
}

func TestBootstrapTreatsCreateConflictAsExisting(t *testing.T) {
	clientManager := mocks.NewClientManager(t)

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, mock.Anything).
		Return(auth.ClientInfo{}, notFound("any")).Twice()
	clientManager.EXPECT().CreateClient(mock.Anything, mock.Anything).
		Return(auth.ClientInfo{}, auth.NewAuthError("already exists", http.StatusConflict)).Twice()

	err := newCommand(clientManager).bootstrap(context.Background(), writeCredentials(t, testCredentials))
	require.NoError(t, err)
}

func TestBootstrapFailsOnLookupError(t *testing.T) {
	clientManager := mocks.NewClientManager(t)

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, "diode-ingest").
		Return(auth.ClientInfo{}, auth.NewAuthError("boom", http.StatusInternalServerError)).Once()

	err := newCommand(clientManager).bootstrap(context.Background(), writeCredentials(t, testCredentials))
	require.ErrorContains(t, err, "failed to look up client diode-ingest")
}

func TestBootstrapFailsOnCreateError(t *testing.T) {
	clientManager := mocks.NewClientManager(t)

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, "diode-ingest").
		Return(auth.ClientInfo{}, notFound("diode-ingest")).Once()
	clientManager.EXPECT().CreateClient(mock.Anything, mock.Anything).
		Return(auth.ClientInfo{}, errors.New("connection refused")).Once()

	err := newCommand(clientManager).bootstrap(context.Background(), writeCredentials(t, testCredentials))
	require.ErrorContains(t, err, "failed to create client diode-ingest")
}

func TestBootstrapCredentialsFileErrors(t *testing.T) {
	tests := []struct {
		name         string
		contents     string
		missingFile  bool
		errorMessage string
	}{
		{
			name:         "missing file",
			missingFile:  true,
			errorMessage: "failed to read credentials file",
		},
		{
			name:         "invalid json",
			contents:     "not json",
			errorMessage: "failed to parse credentials file",
		},
		{
			name:         "missing client id",
			contents:     `[{"client_secret": "s", "scope": "diode:ingest"}]`,
			errorMessage: "entry 0: client_id is required",
		},
		{
			name:         "missing client secret",
			contents:     `[{"client_id": "diode-ingest", "scope": "diode:ingest"}]`,
			errorMessage: "entry 0 (diode-ingest): client_secret is required",
		},
		{
			name:         "missing scope",
			contents:     `[{"client_id": "diode-ingest", "client_secret": "s"}]`,
			errorMessage: "entry 0 (diode-ingest): scope is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientManager := mocks.NewClientManager(t)

			path := filepath.Join(t.TempDir(), "does-not-exist.json")
			if !tt.missingFile {
				path = writeCredentials(t, tt.contents)
			}

			err := newCommand(clientManager).bootstrap(context.Background(), path)
			require.ErrorContains(t, err, tt.errorMessage)
		})
	}
}

func TestWaitForOAuth2ServerRetriesUntilReady(t *testing.T) {
	clientManager := mocks.NewClientManager(t)

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, readinessProbeClientID).
		Return(auth.ClientInfo{}, errors.New("connection refused")).Once()
	clientManager.EXPECT().RetrieveClientByID(mock.Anything, readinessProbeClientID).
		Return(auth.ClientInfo{}, notFound(readinessProbeClientID)).Once()

	cmd := &BootstrapClientsCommand{
		clientManager: clientManager,
		readyTimeout:  time.Second,
		readyInterval: time.Millisecond,
	}

	require.NoError(t, cmd.waitForOAuth2Server(context.Background()))
}

func TestWaitForOAuth2ServerGivesUp(t *testing.T) {
	clientManager := mocks.NewClientManager(t)

	clientManager.EXPECT().RetrieveClientByID(mock.Anything, readinessProbeClientID).
		Return(auth.ClientInfo{}, errors.New("connection refused"))

	cmd := &BootstrapClientsCommand{
		clientManager: clientManager,
		readyTimeout:  10 * time.Millisecond,
		readyInterval: time.Millisecond,
	}

	err := cmd.waitForOAuth2Server(context.Background())
	require.ErrorContains(t, err, "oauth2 server not ready")
}

func TestBootstrapClientsRegisteredAsSubcommand(t *testing.T) {
	commands := MakeDefaultCommands(mocks.NewClientManager(t))
	cmd, ok := commands["bootstrap-clients"]
	require.True(t, ok, "bootstrap-clients subcommand should be registered")

	bootstrapCmd, ok := cmd.(*BootstrapClientsCommand)
	require.True(t, ok)
	require.Equal(t, defaultReadyTimeout, bootstrapCmd.readyTimeout)
	require.Equal(t, defaultReadyInterval, bootstrapCmd.readyInterval)
}
