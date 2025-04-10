package auth_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/auth"
)

func TestNewServer(t *testing.T) {
	ctx := context.Background()

	setupEnv()
	defer teardownEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := auth.NewServer(ctx, logger)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Start and stop the server in a separate goroutine
	go func() {
		err = server.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for the server to start and stop
	time.Sleep(50 * time.Millisecond)
}

func TestIntrospect(t *testing.T) {
	ctx := context.Background()

	setupEnv()
	defer teardownEnv()

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			// Return a mock JWKS response
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"keys": [
					{
						"kty": "RSA",
						"kid": "test-key",
						"use": "sig",
						"alg": "RS256",
						"n": "n_value",
						"e": "AQAB"
					}
				]
			}`))
		}
	}))
	defer mockJWKSServer.Close()

	os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := auth.NewServer(ctx, logger)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create a test server using the server's mux
	testServer := httptest.NewServer(server.GetMux())
	defer testServer.Close()

	// Test case 1: Invalid token
	t.Run("Invalid Token", func(t *testing.T) {
		resp, err := http.Post(
			testServer.URL+"/introspect",
			"application/json",
			strings.NewReader("invalid.token.string"),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var introspectResp auth.IntrospectResponse
		err = json.NewDecoder(resp.Body).Decode(&introspectResp)
		require.NoError(t, err)
		require.False(t, introspectResp.Active)
	})

	// Test case 2: Missing token
	t.Run("Missing Token", func(t *testing.T) {
		resp, err := http.Post(
			testServer.URL+"/introspect",
			"application/json",
			strings.NewReader(""),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var introspectResp auth.IntrospectResponse
		err = json.NewDecoder(resp.Body).Decode(&introspectResp)
		require.NoError(t, err)
		require.False(t, introspectResp.Active)
	})
	// Test case 3: Valid token
	t.Run("Valid Token", func(t *testing.T) {
		// Test with a token that our mock will consider valid
		resp, err := http.Post(
			testServer.URL+"/introspect",
			"application/json",
			strings.NewReader("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwic2NvcGUiOiJyZWFkIHdyaXRlIiwiZXhwIjoxNjk1MjM0MDAwLCJpYXQiOjE2OTUyMzA0MDAsImlzcyI6Imh0dHBzOi8vYXV0aC5leGFtcGxlLmNvbSIsImNsaWVudF9pZCI6ImNsaWVudDEyMyIsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.signature"),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var introspectResp auth.IntrospectResponse
		err = json.NewDecoder(resp.Body).Decode(&introspectResp)
		require.NoError(t, err)
		require.True(t, introspectResp.Active)
		require.Equal(t, "user123", introspectResp.Subject)
		require.Equal(t, "read write", introspectResp.Scope)
		require.Equal(t, "https://auth.example.com", introspectResp.Issuer)
		require.Equal(t, "client123", introspectResp.ClientID)
		require.Equal(t, "testuser", introspectResp.Username)
	})
}
func getFreePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return strconv.Itoa(0), err
	}

	addr := listener.Addr().(*net.TCPAddr)

	if err = listener.Close(); err != nil {
		return strconv.Itoa(0), err
	}
	return strconv.Itoa(addr.Port), nil
}

func setupEnv() {
	httpPort, _ := getFreePort()
	_ = os.Setenv("HTTP_PORT", httpPort)
}

func teardownEnv() {
	_ = os.Unsetenv("HTTP_PORT")
}
