package netboxdiodeplugin_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/netbox"
	"github.com/netboxlabs/diode/diode-server/netboxdiodeplugin"
)

func TestTransfortSecurity(t *testing.T) {
	tests := []struct {
		name             string
		expectedInsecure bool
	}{
		{
			name:             "valid client",
			expectedInsecure: true,
		},
		{
			name:             "default base URL",
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

			client, err := netboxdiodeplugin.NewClient(logger, tt.apiKey)
			if tt.shouldError {
				require.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}

func TestRetrieveObjectState(t *testing.T) {
	tests := []struct {
		name               string
		params             netboxdiodeplugin.RetrieveObjectStateQueryParams
		apiKey             string
		mockServerResponse string
		response           any
		tlsSkipVerify      bool
		shouldError        bool
	}{
		{
			name:               "valid response for DCIM device",
			params:             netboxdiodeplugin.RetrieveObjectStateQueryParams{ObjectType: netbox.DcimDeviceObjectType, ObjectID: 1},
			mockServerResponse: `{"object_type":"dcim.device","object_change_id":1,"object":{"id":1,"name":"test"}}`,
			apiKey:             "foobar",
			response: &netboxdiodeplugin.ObjectState{
				ObjectType:     netbox.DcimDeviceObjectType,
				ObjectChangeID: 1,
				Object: &netbox.DcimDeviceDataWrapper{
					Device: &netbox.DcimDevice{
						ID:   1,
						Name: "test",
					},
				},
			},
			tlsSkipVerify: true,
			shouldError:   false,
		},
		{
			name:               "valid response for DCIM site with query",
			params:             netboxdiodeplugin.RetrieveObjectStateQueryParams{ObjectType: netbox.DcimSiteObjectType, Params: map[string]string{"q": "site 01"}},
			mockServerResponse: `{"object_type":"dcim.site","object_change_id":1,"object":{"id":1,"name":"site 01", "slug": "site-01"}}`,
			apiKey:             "foobar",
			response: &netboxdiodeplugin.ObjectState{
				ObjectType:     netbox.DcimSiteObjectType,
				ObjectChangeID: 1,
				Object: &netbox.DcimSiteDataWrapper{
					Site: &netbox.DcimSite{
						ID:   1,
						Name: "site 01",
						Slug: "site-01",
					},
				},
			},
			tlsSkipVerify: true,
			shouldError:   false,
		},
		{
			name: "valid response for DCIM device with query and additional attributes",
			params: netboxdiodeplugin.RetrieveObjectStateQueryParams{
				ObjectType: netbox.DcimDeviceObjectType,
				ObjectID:   1,
				Params:     map[string]string{"q": "dev1", "attr_name": "site.id", "attr_value": "2"}},
			mockServerResponse: `{"object_type":"dcim.device","object_change_id":1,"object":{"id":1,"name":"dev1", "site": {"id": 2}}}`,
			apiKey:             "foobar",
			response: &netboxdiodeplugin.ObjectState{
				ObjectType:     netbox.DcimDeviceObjectType,
				ObjectChangeID: 1,
				Object: &netbox.DcimDeviceDataWrapper{
					Device: &netbox.DcimDevice{
						ID:   1,
						Name: "dev1",
						Site: &netbox.DcimSite{
							ID: 2,
						},
					},
				},
			},
			tlsSkipVerify: true,
			shouldError:   false,
		},
		{
			name:               "response for invalid object - empty object",
			params:             netboxdiodeplugin.RetrieveObjectStateQueryParams{ObjectType: netbox.DcimDeviceObjectType, ObjectID: 1},
			mockServerResponse: `{"object_type":"dcim.device","object_change_id":1,"object":{"InvalidObjectType": {"id":1,"name":"test"}}}`,
			apiKey:             "foobar",
			response: &netboxdiodeplugin.ObjectState{
				ObjectType:     netbox.DcimDeviceObjectType,
				ObjectChangeID: 1,
				Object: &netbox.DcimDeviceDataWrapper{
					Device: &netbox.DcimDevice{},
				},
			},
			tlsSkipVerify: true,
			shouldError:   false,
		},
		{
			name:               "invalid server response",
			params:             netboxdiodeplugin.RetrieveObjectStateQueryParams{ObjectType: netbox.DcimDeviceObjectType, ObjectID: 1},
			apiKey:             "barfoo",
			mockServerResponse: ``,
			tlsSkipVerify:      true,
			shouldError:        true,
		},
		{
			name:               "tls bad certificate",
			params:             netboxdiodeplugin.RetrieveObjectStateQueryParams{ObjectType: netbox.DcimDeviceObjectType, ObjectID: 1},
			apiKey:             "barfoo",
			mockServerResponse: ``,
			tlsSkipVerify:      false,
			shouldError:        true,
		},
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanUpEnvVars()

			handler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, r.Method, http.MethodGet)
				assert.Equal(t, r.URL.Path, "/api/diode/object-state/")
				assert.Equal(t, r.URL.Query().Get("object_type"), tt.params.ObjectType)
				var objectID string
				if tt.params.ObjectID > 0 {
					objectID = strconv.Itoa(tt.params.ObjectID)
				}
				for k, v := range tt.params.Params {
					assert.Equal(t, r.URL.Query().Get(k), v)
				}
				assert.Equal(t, r.URL.Query().Get("object_id"), objectID)
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Token %s", tt.apiKey))
				assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion))
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}

			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/object-state/", handler)
			ts := httptest.NewTLSServer(mux)
			defer ts.Close()

			_ = os.Setenv(netboxdiodeplugin.BaseURLEnvVarName, fmt.Sprintf("%s/api/diode", ts.URL))
			if tt.tlsSkipVerify {
				_ = os.Setenv(netboxdiodeplugin.TLSSkipVerifyEnvVarName, "true")
			}

			client, err := netboxdiodeplugin.NewClient(logger, tt.apiKey)
			require.NoError(t, err)
			resp, err := client.RetrieveObjectState(context.Background(), tt.params)
			if tt.shouldError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response, resp)
		})
	}
}

func TestApplyChangeSet(t *testing.T) {
	tests := []struct {
		name               string
		apiKey             string
		changeSetRequest   netboxdiodeplugin.ChangeSetRequest
		mockServerResponse string
		mockStatusCode     int
		response           any
		shouldError        bool
	}{
		{
			name:   "valid apply change set response",
			apiKey: "foobar",
			changeSetRequest: netboxdiodeplugin.ChangeSetRequest{
				ChangeSetID: "00000000-0000-0000-0000-000000000000",
				ChangeSet: []netboxdiodeplugin.Change{
					{
						ChangeID:      "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "dcim.device",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data: &netbox.DcimDevice{
							Name: "test",
						},
					},
					{
						ChangeID:      "00000000-0000-0000-0000-000000000002",
						ChangeType:    "update",
						ObjectType:    "dcim.device",
						ObjectID:      ptrInt(1),
						ObjectVersion: ptrInt(2),
						Data: &netbox.DcimDevice{
							Name: "test",
						},
					},
				},
			},
			mockServerResponse: `{"change_set_id":"00000000-0000-0000-0000-000000000000","result":"success"}`,
			mockStatusCode:     http.StatusOK,
			response: &netboxdiodeplugin.ChangeSetResponse{
				ChangeSetID: "00000000-0000-0000-0000-000000000000",
				Result:      "success",
			},
			shouldError: false,
		},
		{
			name:   "invalid request",
			apiKey: "foobar",
			changeSetRequest: netboxdiodeplugin.ChangeSetRequest{
				ChangeSetID: "00000000-0000-0000-0000-000000000000",
				ChangeSet: []netboxdiodeplugin.Change{
					{
						ChangeID:      "00000000-0000-0000-0000-000000000001",
						ChangeType:    "create",
						ObjectType:    "",
						ObjectID:      nil,
						ObjectVersion: nil,
						Data:          nil,
					},
				},
			},
			mockServerResponse: `{"change_set_id":"00000000-0000-0000-0000-000000000000","result":"failure","errors":["invalid object type"]}`,
			mockStatusCode:     http.StatusBadRequest,
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
				w.WriteHeader(tt.mockStatusCode)
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/apply-change-set/", handler)
			ts := httptest.NewServer(mux)
			defer ts.Close()

			_ = os.Setenv(netboxdiodeplugin.BaseURLEnvVarName, fmt.Sprintf("%s/api/diode", ts.URL))

			client, err := netboxdiodeplugin.NewClient(logger, tt.apiKey)
			require.NoError(t, err)
			resp, err := client.ApplyChangeSet(context.Background(), tt.changeSetRequest)
			if tt.shouldError {
				require.Error(t, err)
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
