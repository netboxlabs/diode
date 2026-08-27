package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/netboxlabs/diode/diode-server/auth"
)

const (
	// defaultCredentialsFile is where the oauth2 client credentials secret is mounted.
	defaultCredentialsFile = "/etc/config/oauth2/client/client-credentials.json"

	// The oauth2 server is usually still starting when the bootstrap runs, so the
	// first admin call is retried until it answers rather than blindly sleeping.
	defaultReadyTimeout  = 60 * time.Second
	defaultReadyInterval = 2 * time.Second

	// readinessProbeClientID is only used to check whether the admin API answers;
	// it is never created, and a "not found" response counts as ready.
	readinessProbeClientID = "diode-bootstrap-readiness-probe"
)

// BootstrapClient is a single entry of the oauth2 client credentials file.
type BootstrapClient struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Scope        string `json:"scope"`
}

// BootstrapClientsCommand is a subcommand that creates every oauth2 client
// listed in the client credentials file, skipping the ones that already exist.
type BootstrapClientsCommand struct {
	clientManager auth.ClientManager
	// readyTimeout bounds the wait for the oauth2 admin API. Zero skips the wait.
	readyTimeout  time.Duration
	readyInterval time.Duration
}

// Run implements the SubCommand interface.
func (c *BootstrapClientsCommand) Run() error {
	cmd := flag.NewFlagSet("bootstrap-clients", flag.ExitOnError)
	credentialsFile := cmd.String("credentials-file", defaultCredentialsFile, "path to the oauth2 client credentials file")

	if err := cmd.Parse(os.Args[2:]); err != nil {
		return err
	}

	return c.bootstrap(context.Background(), *credentialsFile)
}

// bootstrap reads the credentials file and reconciles every client in it.
func (c *BootstrapClientsCommand) bootstrap(ctx context.Context, credentialsFile string) error {
	clients, err := readBootstrapClients(credentialsFile)
	if err != nil {
		return err
	}

	if err := c.waitForOAuth2Server(ctx); err != nil {
		return err
	}

	for _, client := range clients {
		if err := c.ensureClient(ctx, client); err != nil {
			return err
		}
	}

	return nil
}

// ensureClient creates the client unless the oauth2 server already has it.
func (c *BootstrapClientsCommand) ensureClient(ctx context.Context, client BootstrapClient) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	_, err := c.clientManager.RetrieveClientByID(ctx, client.ClientID)
	if err == nil {
		CliLog(fmt.Sprintf("INFO: client %s already exists", client.ClientID))
		return nil
	}
	if !hasStatusCode(err, http.StatusNotFound) {
		return fmt.Errorf("failed to look up client %s: %w", client.ClientID, err)
	}

	if _, err := c.clientManager.CreateClient(ctx, auth.ClientInfo{
		ClientID:     client.ClientID,
		ClientSecret: client.ClientSecret,
		Scope:        client.Scope,
	}); err != nil {
		// Another bootstrap run may have won the race between the lookup and here.
		if hasStatusCode(err, http.StatusConflict) {
			CliLog(fmt.Sprintf("INFO: client %s already exists", client.ClientID))
			return nil
		}
		return fmt.Errorf("failed to create client %s: %w", client.ClientID, err)
	}

	CliLog(fmt.Sprintf("INFO: client %s created", client.ClientID))
	return nil
}

// waitForOAuth2Server blocks until the admin API answers or readyTimeout elapses.
func (c *BootstrapClientsCommand) waitForOAuth2Server(ctx context.Context) error {
	if c.readyTimeout <= 0 {
		return nil
	}

	deadline := time.Now().Add(c.readyTimeout)
	var lastErr error

	for {
		probeCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
		_, err := c.clientManager.RetrieveClientByID(probeCtx, readinessProbeClientID)
		cancel()

		// Any well-formed answer, including "not found", means the server is up.
		if err == nil || hasStatusCode(err, http.StatusNotFound) {
			return nil
		}
		lastErr = err

		if time.Now().Add(c.readyInterval).After(deadline) {
			return fmt.Errorf("oauth2 server not ready after %s: %w", c.readyTimeout, lastErr)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.readyInterval):
		}
	}
}

// readBootstrapClients loads and validates the client credentials file.
func readBootstrapClients(credentialsFile string) ([]BootstrapClient, error) {
	contents, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file %s: %w", credentialsFile, err)
	}

	var clients []BootstrapClient
	if err := json.Unmarshal(contents, &clients); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file %s: %w", credentialsFile, err)
	}

	for i, client := range clients {
		switch {
		case client.ClientID == "":
			return nil, fmt.Errorf("credentials entry %d: client_id is required", i)
		case client.ClientSecret == "":
			return nil, fmt.Errorf("credentials entry %d (%s): client_secret is required", i, client.ClientID)
		case client.Scope == "":
			return nil, fmt.Errorf("credentials entry %d (%s): scope is required", i, client.ClientID)
		}
	}

	return clients, nil
}

// hasStatusCode reports whether err is an auth error carrying the given status.
func hasStatusCode(err error, statusCode int) bool {
	var authErr *auth.Error
	return errors.As(err, &authErr) && authErr.StatusCode == statusCode
}
