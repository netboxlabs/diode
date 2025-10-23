package netboxdiodeplugin_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/errors"
	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
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

func TestGenerateDiff(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL                   string
		diodeAuthTokenURL         string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		generateDiffRequest       netboxdiodeplugin.GenerateDiffRequest
		mockStatusCode            int
		expectedBody              string
		mockServerResponse        string
		rateLimiterRPS            int
		rateLimiterBurst          int
		maxRetries                int
		response                  *netboxdiodeplugin.ChangeSetResult
		shouldError               bool
		expectedError             error
		expectedErrorString       string
	}{
		{
			name:                      "valid generate diff response",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			generateDiffRequest: netboxdiodeplugin.GenerateDiffRequest{
				ObjectType: "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: strPtr("test"),
							Site: &diodepb.Site{
								Name: "test-site",
							},
						},
					},
				},
			},
			mockStatusCode:     http.StatusOK,
			expectedBody:       `{"object_type":"dcim.device","entity":{"device":{"name":"test","site":{"name":"test-site"}}}}`,
			mockServerResponse: `{"id": "00000000-0000-0000-0000-000000000001", "change_set": {"id": "00000000-0000-0000-0000-000000000001", "changes": [{"id": "00000000-0000-0000-0000-000000000002", "change_type": "create", "object_type": "dcim.device", "data": {"name": "test"}}]}}`,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			maxRetries:         3,
			response: &netboxdiodeplugin.ChangeSetResult{
				ID: "00000000-0000-0000-0000-000000000001",
				ChangeSet: &netboxdiodeplugin.ChangeSet{
					ID: "00000000-0000-0000-0000-000000000001",
					Changes: []netboxdiodeplugin.Change{
						{
							ID:         "00000000-0000-0000-0000-000000000002",
							ChangeType: "create",
							ObjectType: "dcim.device",
							Data:       json.RawMessage(`{"name": "test"}`),
						},
					},
				},
			},
			shouldError: false,
		},
		{
			name:                      "valid error diff response",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			generateDiffRequest: netboxdiodeplugin.GenerateDiffRequest{
				ObjectType: "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: strPtr("test"),
							Site: &diodepb.Site{
								Name: "test-site",
							},
						},
					},
				},
			},
			mockStatusCode:     http.StatusBadRequest,
			expectedBody:       `{"object_type":"dcim.device","entity":{"device":{"name":"test","site":{"name":"test-site"}}}}`,
			mockServerResponse: `{"id": "00000000-0000-0000-0000-000000000001", "errors": {"dcim.device": {"name": ["illegal name"]}}}`,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			maxRetries:         3,
			shouldError:        true,
			expectedError: &changeset.Error{
				Message: "generate diff failed",
				Code:    errors.ErrCodeOpsGenerateDiff,
				Details: json.RawMessage(`{"id": "00000000-0000-0000-0000-000000000001", "errors": {"dcim.device": {"name": ["illegal name"]}}}`),
			},
		},
		{
			name:                      "invalid error diff response",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			generateDiffRequest: netboxdiodeplugin.GenerateDiffRequest{
				ObjectType: "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: strPtr("test"),
							Site: &diodepb.Site{
								Name: "test-site",
							},
						},
					},
				},
			},
			mockStatusCode:      http.StatusInternalServerError,
			expectedBody:        `{"object_type":"dcim.device","entity":{"device":{"name":"test","site":{"name":"test-site"}}}}`,
			mockServerResponse:  `<html><body><h1>500 Internal Server Error</h1></body></html>`,
			rateLimiterRPS:      1,
			rateLimiterBurst:    1,
			maxRetries:          3,
			shouldError:         true,
			expectedErrorString: "failed to unmarshal response body invalid character '<' looking for beginning of value",
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
				assert.Equal(t, r.Method, http.MethodPost)
				assert.Equal(t, r.URL.Path, "/api/diode/generate-diff/")
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", expectedToken))
				assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion))
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
				assert.Equal(t, r.Header.Get("Accept"), "application/json")

				if tt.generateDiffRequest.BranchID != "" {
					assert.Equal(t, r.Header.Get(netboxdiodeplugin.NetBoxBranchHeader), tt.generateDiffRequest.BranchID)
				} else {
					assert.Len(t, r.Header.Values(netboxdiodeplugin.NetBoxBranchHeader), 0)
				}
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(body))
				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/generate-diff/", handler)
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
			resp, err := client.GenerateDiff(context.Background(), tt.generateDiffRequest)
			if tt.shouldError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				// changeset.Error details must be valid json
				if cse, ok := err.(*changeset.Error); ok {
					assert.True(t, json.Valid(cse.Details))
				}
				assert.Equal(t, tt.response, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response, resp)
			assert.Equal(t, tt.mockStatusCode, http.StatusOK)
		})
	}
}

func TestGenerateDiffRateLimiting(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL                   string
		diodeAuthTokenURL         string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		expectedCalls             int
		generateDiffRequest       netboxdiodeplugin.GenerateDiffRequest
		mockStatusCode            int
		expectedBody              string
		mockServerResponse        string
		rateLimiterRPS            int
		rateLimiterBurst          int
		maxRetries                int
		response                  *netboxdiodeplugin.ChangeSetResult
		shouldError               bool
	}{
		{
			name:                      "rate limited requests",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			expectedCalls:             2,
			generateDiffRequest: netboxdiodeplugin.GenerateDiffRequest{
				ObjectType: "dcim.device",
				Entity: &diodepb.Entity{
					Entity: &diodepb.Entity_Device{
						Device: &diodepb.Device{
							Name: strPtr("test"),
							Site: &diodepb.Site{
								Name: "test-site",
							},
						},
					},
				},
			},
			mockStatusCode:     http.StatusOK,
			expectedBody:       `{"object_type":"dcim.device","entity":{"device":{"name":"test","site":{"name":"test-site"}}}}`,
			mockServerResponse: `{"id": "00000000-0000-0000-0000-000000000001", "change_set": {"id": "00000000-0000-0000-0000-000000000001", "changes": [{"id": "00000000-0000-0000-0000-000000000002", "change_type": "create", "object_type": "dcim.device", "data": {"name": "test"}}]}}`,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			maxRetries:         3,
			response: &netboxdiodeplugin.ChangeSetResult{
				ID: "00000000-0000-0000-0000-000000000001",
				ChangeSet: &netboxdiodeplugin.ChangeSet{
					ID: "00000000-0000-0000-0000-000000000001",
					Changes: []netboxdiodeplugin.Change{
						{
							ID:         "00000000-0000-0000-0000-000000000002",
							ChangeType: "create",
							ObjectType: "dcim.device",
							Data:       json.RawMessage(`{"name": "test"}`),
						},
					},
				},
			},
			shouldError: false,
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

			actualCalls := 0
			handler := func(w http.ResponseWriter, r *http.Request) {
				actualCalls++
				assert.Equal(t, r.Method, http.MethodPost)
				assert.Equal(t, r.URL.Path, "/api/diode/generate-diff/")
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", expectedToken))
				assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion))
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")

				if tt.generateDiffRequest.BranchID != "" {
					assert.Equal(t, r.Header.Get(netboxdiodeplugin.NetBoxBranchHeader), tt.generateDiffRequest.BranchID)
				} else {
					assert.Len(t, r.Header.Values(netboxdiodeplugin.NetBoxBranchHeader), 0)
				}
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, string(body))
				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/generate-diff/", handler)
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
			resp, err := client.GenerateDiff(context.Background(), tt.generateDiffRequest)
			_, _ = client.GenerateDiff(context.Background(), tt.generateDiffRequest)
			if tt.shouldError {
				require.Error(t, err)
				assert.Equal(t, tt.response, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response, resp)
			assert.Equal(t, tt.mockStatusCode, http.StatusOK)
			assert.Equal(t, tt.expectedCalls, actualCalls)
		})
	}
}

func TestAuthenticationParams(t *testing.T) {
	testRequest := netboxdiodeplugin.GenerateDiffRequest{
		ObjectType: "dcim.device",
		Entity: &diodepb.Entity{
			Entity: &diodepb.Entity_Device{
				Device: &diodepb.Device{
					Name: strPtr("test"),
					Site: &diodepb.Site{
						Name: "test-site",
					},
				},
			},
		},
	}
	mockServerResponse := `{"id": "00000000-0000-0000-0000-000000000001", "change_set": {"id": "00000000-0000-0000-0000-000000000001", "changes": [{"id": "00000000-0000-0000-0000-000000000002", "change_type": "create", "object_type": "dcim.device", "data": {"name": "test"}}]}}`

	tests := []struct {
		name                      string
		baseURL                   string
		diodeAuthTokenURL         string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		extraParams               map[string]string
		requiredParams            map[string]string
		shouldError               bool
		expectedError             error
	}{
		{
			name:                      "valid authentication params, no extra params",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test_client",
			diodeToNetBoxClientSecret: "test_secret",
			requiredParams:            requireCredentials("test_client", "test_secret"),
			shouldError:               false,
		},
		{
			name:                      "valid authentication params, with extra params",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test_client",
			diodeToNetBoxClientSecret: "test_secret",
			extraParams:               map[string]string{"test_param": "test_value"},
			requiredParams: map[string]string{
				"client_id":     "test_client",
				"client_secret": "test_secret",
				"test_param":    "test_value",
			},
			shouldError: false,
		},
		{
			name:                      "valid authentication params, missing required params",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test_client",
			diodeToNetBoxClientSecret: "test_secret",
			requiredParams: map[string]string{
				"client_id":     "test_client",
				"client_secret": "test_secret",
				"test_param":    "test_value",
			},
			shouldError: true,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			expectedToken := "mocked-token"
			authTokenURL := "/diode/auth/token"

			mockOAuth2Server := newMockOAuth2Server(authTokenURL, tt.requiredParams, expectedToken)
			defer mockOAuth2Server.Close()

			mockOAuth2ServerURL := mockOAuth2Server.URL + authTokenURL

			handler := func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/generate-diff/", handler)
			ts := httptest.NewServer(mux)
			defer ts.Close()

			baseURL := fmt.Sprintf("%s/api/diode", ts.URL)

			extraParams := url.Values{}
			for k, v := range tt.extraParams {
				extraParams.Set(k, v)
			}
			client, err := netboxdiodeplugin.NewClient(
				netboxdiodeplugin.ClientOptions{
					Logger:              logger,
					BaseURL:             baseURL,
					ClientID:            tt.diodeToNetBoxClientID,
					ClientSecret:        tt.diodeToNetBoxClientSecret,
					TokenURL:            mockOAuth2ServerURL,
					TokenEndpointParams: extraParams,
					RateLimitRPS:        1,
					RateLimitBurstRPS:   1,
					MaxRetries:          0,
				})
			require.NoError(t, err)
			_, err = client.GenerateDiff(context.Background(), testRequest)
			if tt.shouldError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestApplyChangeSet(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL                   string
		diodeAuthTokenURL         string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		changeSetRequest          netboxdiodeplugin.ApplyChangeSetRequest
		mockServerResponse        string
		mockStatusCode            int
		rateLimiterRPS            int
		rateLimiterBurst          int
		maxRetries                int
		response                  *netboxdiodeplugin.ChangeSetResult
		shouldError               bool
		expectedError             error
		expectedErrorString       string
	}{
		{
			name:                      "valid apply change set response",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			changeSetRequest: netboxdiodeplugin.ApplyChangeSetRequest{
				ID: "00000000-0000-0000-0000-000000000000",
				Changes: []netboxdiodeplugin.Change{
					{
						ID:            "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data:          json.RawMessage(`{"name": "test"}`),
					},
					{
						ID:            "00000000-0000-0000-0000-000000000002",
						ChangeType:    "update",
						ObjectType:    "dcim.device",
						ObjectID:      ptrInt(1),
						ObjectVersion: ptrInt(2),
						Data:          json.RawMessage(`{"name": "test"}`),
					},
				},
			},
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000"}`,
			mockStatusCode:     http.StatusOK,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			maxRetries:         3,
			response: &netboxdiodeplugin.ChangeSetResult{
				ID: "00000000-0000-0000-0000-000000000000",
			},
			shouldError: false,
		},
		{
			name:                      "valid apply change set response with branch",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			changeSetRequest: netboxdiodeplugin.ApplyChangeSetRequest{
				ID:       "00000000-0000-0000-0000-000000000000",
				BranchID: "test-branch",
				Changes: []netboxdiodeplugin.Change{
					{
						ID:            "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data:          json.RawMessage(`{"name": "test"}`),
					},
				},
			},
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000"}`,
			mockStatusCode:     http.StatusOK,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			maxRetries:         3,
			response: &netboxdiodeplugin.ChangeSetResult{
				ID: "00000000-0000-0000-0000-000000000000",
			},
			shouldError: false,
		},
		{
			name:                      "invalid request",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			changeSetRequest: netboxdiodeplugin.ApplyChangeSetRequest{
				ID: "00000000-0000-0000-0000-000000000000",
				Changes: []netboxdiodeplugin.Change{
					{
						ID:            "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data:          nil,
					},
				},
			},
			mockStatusCode:   http.StatusBadRequest,
			rateLimiterRPS:   1,
			rateLimiterBurst: 1,
			maxRetries:       3,
			response:         nil,
			shouldError:      true,
		},
		{
			name:                      "invalid post message",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			changeSetRequest: netboxdiodeplugin.ApplyChangeSetRequest{
				ID: "00000000-0000-0000-0000-000000000000",
				Changes: []netboxdiodeplugin.Change{
					{
						ID:            "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data:          json.RawMessage(`{"name": "test"}`),
					},
				},
			},
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000","errors": {"dcim.device": {"name": ["illegal name"]}}}`,
			mockStatusCode:     http.StatusBadRequest,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			maxRetries:         3,
			response:           nil,
			shouldError:        true,
			expectedError: &changeset.Error{
				Message: "apply change set failed",
				Code:    errors.ErrCodeOpsApplyChangeSet,
				Details: json.RawMessage(`{"id":"00000000-0000-0000-0000-000000000000","errors": {"dcim.device": {"name": ["illegal name"]}}}`),
			},
		},
		{
			name:                      "unmarshal error",
			baseURL:                   "http://",
			diodeAuthTokenURL:         "http://diode-auth:8000/diode/auth/token",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			changeSetRequest: netboxdiodeplugin.ApplyChangeSetRequest{
				ID: "00000000-0000-0000-0000-000000000000",
				Changes: []netboxdiodeplugin.Change{
					{
						ID:            "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data:          json.RawMessage(`{"name": "test"}`),
					},
				},
			},
			mockServerResponse:  `{"id"  - "00000000-0000-0000\-0000-000000000000","result":"error"}`,
			mockStatusCode:      http.StatusBadRequest,
			rateLimiterRPS:      1,
			rateLimiterBurst:    1,
			maxRetries:          3,
			response:            nil,
			shouldError:         true,
			expectedErrorString: "failed to unmarshal response body invalid character '-' after object key",
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
				assert.Equal(t, r.Method, http.MethodPost)
				assert.Equal(t, r.URL.Path, "/api/diode/apply-change-set/")
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", expectedToken))
				assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion))
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
				if tt.changeSetRequest.BranchID != "" {
					assert.Equal(t, r.Header.Get(netboxdiodeplugin.NetBoxBranchHeader), tt.changeSetRequest.BranchID)
				} else {
					assert.Len(t, r.Header.Values(netboxdiodeplugin.NetBoxBranchHeader), 0)
				}
				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/apply-change-set/", handler)
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
			resp, err := client.ApplyChangeSet(context.Background(), tt.changeSetRequest)
			if tt.shouldError {
				require.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				if tt.expectedErrorString != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorString)
				}
				// changeset.Error details must be valid json
				if cse, ok := err.(*changeset.Error); ok {
					assert.True(t, json.Valid(cse.Details))
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response, resp)
			assert.Equal(t, tt.mockStatusCode, http.StatusOK)
		})
	}
}

func TestApplyChangeSetRateLimiting(t *testing.T) {
	tests := []struct {
		name                      string
		baseURL                   string
		diodeToNetBoxClientID     string
		diodeToNetBoxClientSecret string
		changeSetRequest          netboxdiodeplugin.ApplyChangeSetRequest
		expectedCalls             int
		mockServerResponse        string
		mockStatusCode            int
		rateLimiterRPS            int
		rateLimiterBurst          int
		maxRetries                int
		response                  *netboxdiodeplugin.ChangeSetResult
		shouldError               bool
	}{
		{
			name:                      "rate limit error",
			diodeToNetBoxClientID:     "test",
			diodeToNetBoxClientSecret: "test",
			changeSetRequest: netboxdiodeplugin.ApplyChangeSetRequest{
				ID:       "00000000-0000-0000-0000-000000000000",
				BranchID: "test-branch",
				Changes: []netboxdiodeplugin.Change{
					{
						ID:            "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data:          json.RawMessage(`{"name": "test"}`),
					},
				},
			},
			expectedCalls:      2,
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000"}`,
			mockStatusCode:     http.StatusOK,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			maxRetries:         3,
			response: &netboxdiodeplugin.ChangeSetResult{
				ID: "00000000-0000-0000-0000-000000000000",
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
				assert.Equal(t, r.Method, http.MethodPost)
				assert.Equal(t, r.URL.Path, "/api/diode/apply-change-set/")
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", expectedToken))
				assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion))
				assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
				if tt.changeSetRequest.BranchID != "" {
					assert.Equal(t, r.Header.Get(netboxdiodeplugin.NetBoxBranchHeader), tt.changeSetRequest.BranchID)
				} else {
					assert.Len(t, r.Header.Values(netboxdiodeplugin.NetBoxBranchHeader), 0)
				}
				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/apply-change-set/", handler)
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
			resp, err := client.ApplyChangeSet(context.Background(), tt.changeSetRequest)
			_, _ = client.ApplyChangeSet(context.Background(), tt.changeSetRequest)
			if tt.shouldError {
				require.Error(t, err)
				assert.Equal(t, tt.response, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response, resp)
			assert.Equal(t, tt.mockStatusCode, http.StatusOK)
			assert.Equal(t, tt.expectedCalls, actualCalls)
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
			expectedErrorString:       "get default branch failed with status 500",
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

func ptrInt(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
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
