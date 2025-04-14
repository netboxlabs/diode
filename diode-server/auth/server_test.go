package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/auth"
)

type InvalidParser struct{}

func (p InvalidParser) Parse(tokenString string, keyfunc jwt.Keyfunc) (*jwt.Token, error) {
	return nil, fmt.Errorf("invalid token")
}

type ValidTokenParser struct{}

func (p ValidTokenParser) Parse(tokenString string, keyfunc jwt.Keyfunc) (*jwt.Token, error) {
	claims := jwt.MapClaims{
		"iss":       "https://auth.example.com",
		"sub":       "user123",
		"aud":       "api",
		"exp":       time.Now().Add(time.Hour).Unix(),
		"iat":       time.Now().Unix(),
		"client_id": "client123",
		"scope":     "read write",
		"username":  "testuser",
	}

	token := &jwt.Token{
		Claims: claims,
		Valid:  true,
	}
	return token, nil
}

func TestNewServer(t *testing.T) {
	ctx := context.Background()

	setupEnv()
	defer teardownEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := auth.NewServer(ctx, logger, InvalidParser{})
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

func TestIntrospectForInvalidTokens(t *testing.T) {
	ctx := context.Background()

	setupEnv()
	defer teardownEnv()

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			// Return a mock JWKS response
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
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

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := auth.NewServer(ctx, logger, InvalidParser{})
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create a test server using the server's mux
	testServer := httptest.NewServer(server.GetMux())
	defer testServer.Close()

	// Test case 1: Invalid token
	t.Run("Invalid Token", func(t *testing.T) {
		resp, err := http.Post(
			testServer.URL+"/introspect",
			"application/x-www-form-urlencoded",
			strings.NewReader("invalid.token.string"),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
	})

	// Test case 2: Missing token
	t.Run("Missing Token", func(t *testing.T) {
		resp, err := http.Post(
			testServer.URL+"/introspect",
			"application/x-www-form-urlencoded",
			strings.NewReader(""),
		)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
	})
}

func TestIntrospectForValidTokens(t *testing.T) {
	ctx := context.Background()

	setupEnv()
	defer teardownEnv()

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			// Return a mock JWKS response
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
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

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	server, err := auth.NewServer(ctx, logger, ValidTokenParser{})
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create a test server using the server's mux
	testServer := httptest.NewServer(server.GetMux())
	defer testServer.Close()

	t.Run("Valid Token", func(t *testing.T) {
		// This is just a dummy token for testing purposes
		testToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3Qta2V5IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoiYXBpIiwiZXhwIjoxNjUwMDAwMDAwLCJpYXQiOjE1MDAwMDAwMDAsImNsaWVudF9pZCI6ImNsaWVudDEyMyIsInNjb3BlIjoicmVhZCB3cml0ZSIsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.WcPGXClpKD7Bc1C0CCDA1060E2GGlTfamrd8-W0ghBE"

		data := url.Values{}
		data.Set("token", testToken)

		resp, err := http.Post(
			testServer.URL+"/introspect",
			"application/x-www-form-urlencoded",
			strings.NewReader(data.Encode()),
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
