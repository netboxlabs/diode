package netboxdiodeplugin_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/diodepb"
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
		name             string
		apiKey           string
		baseURL          string
		timeout          string
		setBaseURLEnvVar bool
		setTimeoutEnvVar bool
		setTLSSkipEnvVar bool
		shouldError      bool
	}{
		{
			name:             "valid client",
			apiKey:           "test",
			baseURL:          "http://",
			timeout:          "5",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: true,
			setTLSSkipEnvVar: false,
			shouldError:      false,
		},
		{
			name:             "default base URL",
			apiKey:           "test",
			baseURL:          "",
			timeout:          "5",
			setBaseURLEnvVar: false,
			setTimeoutEnvVar: true,
			shouldError:      false,
		},
		{
			name:             "invalid base URL",
			apiKey:           "test",
			baseURL:          "http://local\nhost",
			timeout:          "5",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: true,
			setTLSSkipEnvVar: false,
			shouldError:      true,
		},
		{
			name:             "default timeout",
			apiKey:           "test",
			baseURL:          "http://",
			timeout:          "",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: false,
			setTLSSkipEnvVar: false,
			shouldError:      false,
		},
		{
			name:             "invalid timeout",
			apiKey:           "test",
			baseURL:          "http://",
			timeout:          "-1",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: true,
			setTLSSkipEnvVar: false,
			shouldError:      true,
		},
		{
			name:             "API key not provided",
			apiKey:           "",
			baseURL:          "http://",
			timeout:          "5",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: true,
			setTLSSkipEnvVar: false,
			shouldError:      true,
		},
		{
			name:             "set TLS skip verify",
			apiKey:           "test",
			baseURL:          "",
			timeout:          "5",
			setBaseURLEnvVar: false,
			setTimeoutEnvVar: true,
			setTLSSkipEnvVar: true,
			shouldError:      false,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			if tt.setBaseURLEnvVar {
				_ = os.Setenv(netboxdiodeplugin.BaseURLEnvVarName, tt.baseURL)
			}
			if tt.setTimeoutEnvVar {
				_ = os.Setenv(netboxdiodeplugin.TimeoutSecondsEnvVarName, tt.timeout)
			}
			if tt.setTLSSkipEnvVar {
				_ = os.Setenv(netboxdiodeplugin.TLSSkipVerifyEnvVarName, "true")
			}

			client, err := netboxdiodeplugin.NewClient(logger, tt.apiKey, 1, 1)
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
		name                string
		apiKey              string
		generateDiffRequest netboxdiodeplugin.GenerateDiffRequest
		mockStatusCode      int
		expectedBody        string
		mockServerResponse  string
		rateLimiterRPS      int
		rateLimiterBurst    int
		response            *netboxdiodeplugin.GenerateDiffResponse
		shouldError         bool
	}{
		{
			name:   "valid generate diff response",
			apiKey: "foobar",
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
			mockServerResponse: `{"changes": [{"id": "00000000-0000-0000-0000-000000000001", "change_type": "create", "object_type": "dcim.device", "data": {"name": "test"}}]}`,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			response: &netboxdiodeplugin.GenerateDiffResponse{
				Changes: []netboxdiodeplugin.Change{
					{
						ID:         "00000000-0000-0000-0000-000000000001",
						ChangeType: "create",
						ObjectType: "dcim.device",
						Data:       json.RawMessage(`{"name": "test"}`),
					},
				},
			},
			shouldError: false,
		},
		{
			name:   "valid generate diff response",
			apiKey: "foobar",
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
			mockServerResponse: `{"changes": [{"id": "00000000-0000-0000-0000-000000000001", "change_type": "create", "object_type": "dcim.device", "data": {"name": "test"}}]}`,
			rateLimiterRPS:     0, // Any calls made to netbox will be rate limited
			rateLimiterBurst:   0,
			response:           nil,
			shouldError:        true,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, http.MethodPost)
				assert.Equal(t, r.URL.Path, "/api/diode/generate-diff/")
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Token %s", tt.apiKey))
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

			_ = os.Setenv(netboxdiodeplugin.BaseURLEnvVarName, fmt.Sprintf("%s/api/diode", ts.URL))

			client, err := netboxdiodeplugin.NewClient(logger, tt.apiKey, tt.rateLimiterRPS, tt.rateLimiterBurst)
			require.NoError(t, err)
			resp, err := client.GenerateDiff(context.Background(), tt.generateDiffRequest)
			if tt.shouldError {
				require.Error(t, err)
				assert.Equal(t, tt.response, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response, resp)
			assert.Equal(t, tt.mockStatusCode, http.StatusOK)
		})
	}
}

func TestApplyChangeSet(t *testing.T) {
	tests := []struct {
		name               string
		apiKey             string
		changeSetRequest   netboxdiodeplugin.ApplyChangeSetRequest
		mockServerResponse string
		mockStatusCode     int
		rateLimiterRPS     int
		rateLimiterBurst   int
		response           *netboxdiodeplugin.ApplyChangeSetResponse
		shouldError        bool
	}{
		{
			name:   "valid apply change set response",
			apiKey: "foobar",
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
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000","result":"success"}`,
			mockStatusCode:     http.StatusOK,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			response: &netboxdiodeplugin.ApplyChangeSetResponse{
				ID:     "00000000-0000-0000-0000-000000000000",
				Result: "success",
			},
			shouldError: false,
		},
		{
			name:   "valid apply change set response with branch",
			apiKey: "foobar",
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
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000","result":"success"}`,
			mockStatusCode:     http.StatusOK,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			response: &netboxdiodeplugin.ApplyChangeSetResponse{
				ID:     "00000000-0000-0000-0000-000000000000",
				Result: "success",
			},
			shouldError: false,
		},
		{
			name:   "rate limit error",
			apiKey: "foobar",
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
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000","result":"success"}`,
			mockStatusCode:     http.StatusOK,
			rateLimiterRPS:     0, // Any calls made to netbox will be rate limited
			rateLimiterBurst:   0,
			response:           nil,
			shouldError:        true,
		},
		{
			name:   "invalid request",
			apiKey: "foobar",
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
			rateLimiterRPS:   1,
			rateLimiterBurst: 1,
			response:         nil,
			shouldError:      true,
		},
		{
			name:   "invalid post message",
			apiKey: "foobar",
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
			mockServerResponse: `{"id":"00000000-0000-0000-0000-000000000000","result":"error"}`,
			mockStatusCode:     http.StatusBadRequest,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			response:           nil,
			shouldError:        true,
		},
		{
			name:   "unmarshal error",
			apiKey: "foobar",
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
			mockServerResponse: `{"id"  - "00000000-0000-0000\-0000-000000000000","result":"error"}`,
			mockStatusCode:     http.StatusBadRequest,
			rateLimiterRPS:     1,
			rateLimiterBurst:   1,
			response:           nil,
			shouldError:        true,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, http.MethodPost)
				assert.Equal(t, r.URL.Path, "/api/diode/apply-change-set/")
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Token %s", tt.apiKey))
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

			_ = os.Setenv(netboxdiodeplugin.BaseURLEnvVarName, fmt.Sprintf("%s/api/diode", ts.URL))

			client, err := netboxdiodeplugin.NewClient(logger, tt.apiKey, tt.rateLimiterRPS, tt.rateLimiterBurst)
			require.NoError(t, err)
			resp, err := client.ApplyChangeSet(context.Background(), tt.changeSetRequest)
			if tt.shouldError {
				require.Error(t, err)
				assert.Equal(t, tt.response, resp)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response, resp)
			assert.Equal(t, tt.mockStatusCode, http.StatusOK)
		})
	}
}

func cleanUpEnvVars() {
	_ = os.Unsetenv(netboxdiodeplugin.BaseURLEnvVarName)
	_ = os.Unsetenv(netboxdiodeplugin.TimeoutSecondsEnvVarName)
	_ = os.Unsetenv(netboxdiodeplugin.TLSSkipVerifyEnvVarName)
}

func ptrInt(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}
