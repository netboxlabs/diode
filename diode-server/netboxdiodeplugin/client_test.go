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

func TestNewClient(t *testing.T) {
	tests := []struct {
		name             string
		apiKey           string
		baseURL          string
		timeout          string
		setBaseURLEnvVar bool
		setTimeoutEnvVar bool
		shouldError      bool
	}{
		{
			name:             "valid client",
			apiKey:           "test",
			baseURL:          "http://",
			timeout:          "5",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: true,
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
			shouldError:      true,
		},
		{
			name:             "default timeout",
			apiKey:           "test",
			baseURL:          "http://",
			timeout:          "",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: false,
			shouldError:      false,
		},
		{
			name:             "invalid timeout",
			apiKey:           "test",
			baseURL:          "http://",
			timeout:          "-1",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: true,
			shouldError:      true,
		},
		{
			name:             "API key not provided",
			apiKey:           "",
			baseURL:          "http://",
			timeout:          "5",
			setBaseURLEnvVar: true,
			setTimeoutEnvVar: true,
			shouldError:      true,
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
		objectType         string
		objectID           int
		query              string
		apiKey             string
		mockServerResponse string
		response           any
		shouldError        bool
	}{
		{
			name:               "valid response for DCIM device",
			objectType:         netbox.DcimDeviceObjectType,
			objectID:           1,
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
			shouldError: false,
		},
		{
			name:               "valid response for DCIM site with query",
			objectType:         netbox.DcimSiteObjectType,
			objectID:           0,
			query:              "site 01",
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
			shouldError: false,
		},
		{
			name:               "response for invalid object - empty object",
			objectType:         netbox.DcimDeviceObjectType,
			objectID:           1,
			mockServerResponse: `{"object_type":"dcim.device","object_change_id":1,"object":{"InvalidObjectType": {"id":1,"name":"test"}}}`,
			apiKey:             "foobar",
			response: &netboxdiodeplugin.ObjectState{
				ObjectType:     netbox.DcimDeviceObjectType,
				ObjectChangeID: 1,
				Object: &netbox.DcimDeviceDataWrapper{
					Device: &netbox.DcimDevice{},
				},
			},
			shouldError: false,
		},
		{
			name:               "invalid server response",
			objectType:         netbox.DcimDeviceObjectType,
			objectID:           100,
			apiKey:             "barfoo",
			mockServerResponse: ``,
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
				assert.Equal(t, r.URL.Query().Get("object_type"), tt.objectType)
				var objectID string
				if tt.objectID > 0 {
					objectID = strconv.Itoa(tt.objectID)
				}
				assert.Equal(t, r.URL.Query().Get("object_id"), objectID)
				assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Token %s", tt.apiKey))
				assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("%s/%s", netboxdiodeplugin.SDKName, netboxdiodeplugin.SDKVersion))
				_, _ = w.Write([]byte(tt.mockServerResponse))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("/api/diode/object-state/", handler)
			ts := httptest.NewServer(mux)
			defer ts.Close()

			_ = os.Setenv(netboxdiodeplugin.BaseURLEnvVarName, fmt.Sprintf("%s/api/diode", ts.URL))

			client, err := netboxdiodeplugin.NewClient(logger, tt.apiKey)
			require.NoError(t, err)
			resp, err := client.RetrieveObjectState(context.Background(), tt.objectType, tt.objectID, tt.query)
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
}

func ptrInt(i int) *int {
	return &i
}
