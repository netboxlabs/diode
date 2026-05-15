package netboxdiodeplugin_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
)

func TestTransportSecurity(t *testing.T) {
	tests := []struct {
		name             string
		expectedInsecure bool
	}{
		{
			name:             "enable insecure mode",
			expectedInsecure: true,
		},
		{
			name:             "default secure TLS config",
			expectedInsecure: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			if tt.expectedInsecure {
				_ = os.Setenv(netboxdiodeplugin.TLSSkipVerifyEnvVarName, "true")
			}

			httpTransport := netboxdiodeplugin.NewHTTPTransport()
			assert.Equal(t, tt.expectedInsecure, httpTransport.TLSClientConfig.InsecureSkipVerify)
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL                   string
		diodeAuthTokenURL         string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		rateLimiterRPS            int
		rateLimiterBurst          int
		timeout                   string
		setTimeoutEnvVar          bool
		setTLSSkipEnvVar          bool
		shouldError               bool
	}{
		{
			name:                      "valid client",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               false,
		},
		{
			name:                      "invalid base URL",
			baseURL:                   "http://local\nhost",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               true,
		},
		{
			name:                      "default timeout",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "",
			setTimeoutEnvVar:          false,
			setTLSSkipEnvVar:          false,
			shouldError:               false,
		},
		{
			name:                      "invalid timeout",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "-1",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               true,
		},
		{
			name:                      "client ID not provided",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               true,
		},
		{
			name:                      "client secret not provided",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               true,
		},
		{
			name:                      "token URL not provided",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               true,
		},
		{
			name:                      "set TLS skip verify",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          true,
			shouldError:               false,
		},
		{
			name:                      "invalid rate limiter rps parameter",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            0,
			rateLimiterBurst:          1,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               true,
		},
		{
			name:                      "invalid rate limiter burst parameter",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			rateLimiterRPS:            1,
			rateLimiterBurst:          0,
			timeout:                   "5",
			setTimeoutEnvVar:          true,
			setTLSSkipEnvVar:          false,
			shouldError:               true,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			if tt.setTimeoutEnvVar {
				_ = os.Setenv(netboxdiodeplugin.TimeoutSecondsEnvVarName, tt.timeout)
			}
			if tt.setTLSSkipEnvVar {
				_ = os.Setenv(netboxdiodeplugin.TLSSkipVerifyEnvVarName, "true")
			}

			maxRetries := 3

			client, err := netboxdiodeplugin.NewClient(
				netboxdiodeplugin.ClientOptions{
					Logger:            logger,
					BaseURL:           tt.baseURL,
					ClientID:          tt.diodeToNetBoxClientID,
					ClientSecret:      tt.diodeToNetBoxClientSecret,
					TokenURL:          tt.diodeAuthTokenURL,
					RateLimitRPS:      tt.rateLimiterRPS,
					RateLimitBurstRPS: tt.rateLimiterBurst,
					MaxRetries:        maxRetries,
				})
			if tt.shouldError {
				require.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}


func TestGetDefaultBranch(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL                   string
		diodeAuthTokenURL         string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		mockServerResponse        string
		mockStatusCode            int
		rateLimiterRPS            int
		rateLimiterBurst          int
		maxRetries                int
		expectedBranch            *netboxdiodeplugin.Branch
		shouldError               bool
		expectedErrorString       string
	}{
		{
			name:                      "successful branch retrieval",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			mockServerResponse:        `{"branch": {"id": "abc123", "name": "dev-branch"}}`,
			mockStatusCode:            http.StatusOK,
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			maxRetries:                3,
			expectedBranch: &netboxdiodeplugin.Branch{
				ID:   "abc123",
				Name: "dev-branch",
			},
			shouldError: false,
		},
		{
			name:                      "no default branch (null)",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			mockServerResponse:        `{"branch": null}`,
			mockStatusCode:            http.StatusOK,
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			maxRetries:                3,
			expectedBranch:            nil,
			shouldError:               false,
		},
		{
			name:                      "HTTP 500 error",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			mockServerResponse:        `{"error": "Internal server error"}`,
			mockStatusCode:            http.StatusInternalServerError,
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			maxRetries:                3,
			expectedBranch:            nil,
			shouldError:               true,
			expectedErrorString:       "giving up after 4 attempt(s)",
		},
		{
			name:                      "HTTP 404 error returns sentinel error",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			mockServerResponse:        `{"error": "Not found"}`,
			mockStatusCode:            http.StatusNotFound,
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			maxRetries:                3,
			expectedBranch:            nil,
			shouldError:               true,
			expectedErrorString:       "default branch not found",
		},
		{
			name:                      "invalid JSON response",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			mockServerResponse:        `{invalid json}`,
			mockStatusCode:            http.StatusOK,
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			maxRetries:                3,
			expectedBranch:            nil,
			shouldError:               true,
			expectedErrorString:       "failed to unmarshal response body",
		},
		{
			name:                      "HTML error response",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			mockServerResponse:        `<html><body><h1>500 Internal Server Error</h1></body></html>`,
			mockStatusCode:            http.StatusOK,
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			maxRetries:                3,
			expectedBranch:            nil,
			shouldError:               true,
			expectedErrorString:       "failed to unmarshal response body",
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			expectedToken := "mocked-token"
			authTokenURL := "/diode/auth/token"
			mockOAuth2Server := newMockOAuth2Server(authTokenURL, requireCredentials(tt.diodeToNetBoxClientID, tt.diodeToNetBoxClientSecret), expectedToken)
			defer mockOAuth2Server.Close()

			mockOAuth2ServerURL := mockOAuth2Server.URL + authTokenURL

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/diode/default-branch/", r.URL.Path)
				assert.Equal(t, fmt.Sprintf("Bearer %s", expectedToken), r.Header.Get("Authorization"))
				assert.Equal(t, fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion), r.Header.Get("User-Agent"))

				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/default-branch/", handler)
			ts := httptest.NewServer(mux)
			defer ts.Close()

			baseURL := fmt.Sprintf("%s/api/diode", ts.URL)
			client, err := netboxdiodeplugin.NewClient(
				netboxdiodeplugin.ClientOptions{
					Logger:            logger,
					BaseURL:           baseURL,
					ClientID:          tt.diodeToNetBoxClientID,
					ClientSecret:      tt.diodeToNetBoxClientSecret,
					TokenURL:          mockOAuth2ServerURL,
					RateLimitRPS:      tt.rateLimiterRPS,
					RateLimitBurstRPS: tt.rateLimiterBurst,
					MaxRetries:        tt.maxRetries,
				})
			require.NoError(t, err)

			branch, err := client.GetDefaultBranch(context.Background())

			if tt.shouldError {
				require.Error(t, err)
				if tt.expectedErrorString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorString)
				}
				// For 404 errors, verify it's the sentinel error
				if tt.mockStatusCode == http.StatusNotFound {
					assert.ErrorIs(t, err, netboxdiodeplugin.ErrDefaultBranchNotFound)
				}
				assert.Nil(t, branch)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedBranch, branch)
		})
	}
}

func TestGetDefaultBranchRateLimiting(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL                   string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		expectedCalls             int
		mockServerResponse        string
		mockStatusCode            int
		rateLimiterRPS            int
		rateLimiterBurst          int
		maxRetries                int
		expectedBranch            *netboxdiodeplugin.Branch
		shouldError               bool
	}{
		{
			name:                      "rate limited requests",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			expectedCalls:             2,
			mockServerResponse:        `{"branch": {"id": "abc123", "name": "dev-branch"}}`,
			mockStatusCode:            http.StatusOK,
			rateLimiterRPS:            1,
			rateLimiterBurst:          1,
			maxRetries:                3,
			expectedBranch: &netboxdiodeplugin.Branch{
				ID:   "abc123",
				Name: "dev-branch",
			},
			shouldError: false,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()
			actualCalls := 0

			expectedToken := "mocked-token"
			authTokenURL := "/diode/auth/token"
			mockOAuth2Server := newMockOAuth2Server(authTokenURL, requireCredentials(tt.diodeToNetBoxClientID, tt.diodeToNetBoxClientSecret), expectedToken)
			defer mockOAuth2Server.Close()

			mockOAuth2ServerURL := mockOAuth2Server.URL + authTokenURL

			handler := func(w http.ResponseWriter, r *http.Request) {
				actualCalls++
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/diode/default-branch/", r.URL.Path)
				assert.Equal(t, fmt.Sprintf("Bearer %s", expectedToken), r.Header.Get("Authorization"))
				assert.Equal(t, fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion), r.Header.Get("User-Agent"))

				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/default-branch/", handler)
			ts := httptest.NewServer(mux)
			defer ts.Close()

			baseURL := fmt.Sprintf("%s/api/diode", ts.URL)

			client, err := netboxdiodeplugin.NewClient(
				netboxdiodeplugin.ClientOptions{
					Logger:            logger,
					BaseURL:           baseURL,
					ClientID:          tt.diodeToNetBoxClientID,
					ClientSecret:      tt.diodeToNetBoxClientSecret,
					TokenURL:          mockOAuth2ServerURL,
					RateLimitRPS:      tt.rateLimiterRPS,
					RateLimitBurstRPS: tt.rateLimiterBurst,
					MaxRetries:        tt.maxRetries,
				})
			require.NoError(t, err)

			// Make two calls to test rate limiting
			branch, err := client.GetDefaultBranch(context.Background())
			_, _ = client.GetDefaultBranch(context.Background())

			if tt.shouldError {
				require.Error(t, err)
				assert.Nil(t, branch)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedBranch, branch)
			assert.Equal(t, tt.mockStatusCode, http.StatusOK)
			assert.Equal(t, tt.expectedCalls, actualCalls)
		})
	}
}

func cleanUpEnvVars() {
	_ = os.Unsetenv(netboxdiodeplugin.TimeoutSecondsEnvVarName)
	_ = os.Unsetenv(netboxdiodeplugin.TLSSkipVerifyEnvVarName)
}

func requireCredentials(clientID, clientSecret string) map[string]string {
	return map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
	}
}

func newMockOAuth2Server(authTokenURL string, requiredParams map[string]string, mockedToken string) *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc(authTokenURL, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid form", http.StatusBadRequest)
			return
		}

		for k, v := range requiredParams {
			if r.PostForm.Get(k) != v {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				if err := json.NewEncoder(w).Encode(map[string]string{
					"error":             "unauthorized",
					"error_description": "Authentication required",
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				return
			}
		}

		// Simulate token response
		resp := map[string]any{
			"access_token": mockedToken,
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        "dummy-scope",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return httptest.NewServer(handler)
}
