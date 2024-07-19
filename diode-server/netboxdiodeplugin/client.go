package netboxdiodeplugin

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/netboxlabs/diode/diode-server/netbox"
)

const (
	// SDKName is the name of the SDK
	SDKName = "netbox-diode-plugin-sdk-go"

	// SDKVersion is the version of the SDK
	SDKVersion = "0.1.0"

	// BaseURLEnvVarName is the environment variable name for the NetBox Diode plugin HTTP base URL
	BaseURLEnvVarName = "NETBOX_DIODE_PLUGIN_API_BASE_URL"

	// TLSSkipVerifyEnvVarName is the environment variable name for Netbox Diode plugin TLS verification
	TLSSkipVerifyEnvVarName = "NETBOX_DIODE_PLUGIN_SKIP_TLS_VERIFY"

	// TimeoutSecondsEnvVarName is the environment variable name for the NetBox Diode plugin HTTP timeout
	TimeoutSecondsEnvVarName = "NETBOX_DIODE_PLUGIN_API_TIMEOUT_SECONDS"

	defaultBaseURL = "http://127.0.0.1:8080/api/plugins/diode"

	defaultHTTPTimeoutSeconds = 5
)

var (
	// ErrInvalidTimeout is an error for invalid timeout value
	ErrInvalidTimeout = errors.New("invalid timeout value")
)

type apiRoundTripper struct {
	transport http.RoundTripper
	apiKey    string
	userAgent string
}

func newAPIRoundTripper(apiKey string, next http.RoundTripper) (http.RoundTripper, error) {
	if len(apiKey) == 0 {
		return nil, fmt.Errorf("API key not provided")
	}

	return &apiRoundTripper{
		transport: next,
		apiKey:    apiKey,
		userAgent: userAgent(),
	}, nil
}

// RoundTrip implements the RoundTripper interface
func (rt *apiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to ensure thread safety
	req2 := req.Clone(req.Context())

	// Set authorization header
	req2.Header.Set("Authorization", fmt.Sprintf("Token %s", rt.apiKey))

	// Set user agent header
	req2.Header.Set("User-Agent", rt.userAgent)

	// Set content type header
	req2.Header.Set("Content-Type", "application/json")

	return rt.transport.RoundTrip(req2)
}

// NetBoxAPI is the interface for the NetBox Diode plugin API
type NetBoxAPI interface {
	// RetrieveObjectState retrieves the object state
	RetrieveObjectState(context.Context, RetrieveObjectStateQueryParams) (*ObjectState, error)

	// ApplyChangeSet applies a change set
	ApplyChangeSet(context.Context, ChangeSetRequest) (*ChangeSetResponse, error)
}

// Client is a NetBox Diode plugin client
type Client struct {
	logger     *slog.Logger
	httpClient *http.Client
	baseURL    *url.URL
}

// NewClient creates a new NetBox Diode plugin client
func NewClient(logger *slog.Logger, apiKey string) (*Client, error) {

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLS(),
		},
	}

	rt, err := newAPIRoundTripper(apiKey, transport)
	if err != nil {
		return nil, err
	}

	timeout, err := httpTimeout()
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Transport: rt,
		Timeout:   timeout,
	}

	u, err := url.Parse(baseURL())
	if err != nil {
		return nil, err
	}

	client := &Client{
		logger:     logger,
		httpClient: httpClient,
		baseURL:    u,
	}

	return client, nil
}

func userAgent() string {
	return fmt.Sprintf("%s/%s", SDKName, SDKVersion)
}

func baseURL() string {
	u, ok := os.LookupEnv(BaseURLEnvVarName)
	if !ok {
		u = defaultBaseURL
	}
	return u
}

func skipTLS() bool {
	skipTLS, ok := os.LookupEnv(TLSSkipVerifyEnvVarName)
	if !ok {
		return false
	}
	skip, err := strconv.ParseBool(skipTLS)
	if err != nil {
		return false
	}
	return skip
}

func httpTimeout() (time.Duration, error) {
	timeoutSecondsStr, ok := os.LookupEnv(TimeoutSecondsEnvVarName)
	if !ok || len(timeoutSecondsStr) == 0 {
		return defaultHTTPTimeoutSeconds * time.Second, nil
	}

	timeout, err := strconv.Atoi(timeoutSecondsStr)
	if err != nil || timeout <= 0 {
		return 0, ErrInvalidTimeout
	}
	return time.Duration(timeout) * time.Second, nil
}

type objectStateRaw struct {
	ObjectID       int    `json:"object_id"`
	ObjectType     string `json:"object_type"`
	ObjectChangeID int    `json:"object_change_id"`
	Object         any    `json:"object"`
}

// ObjectState represents the NetBox object state
type ObjectState struct {
	ObjectID       int                   `json:"object_id"`
	ObjectType     string                `json:"object_type"`
	ObjectChangeID int                   `json:"object_change_id"`
	Object         netbox.ComparableData `json:"object"`
}

// RetrieveObjectStateQueryParams represents the query parameters for retrieving the object state
type RetrieveObjectStateQueryParams struct {
	ObjectType string
	ObjectID   int
	Params     map[string]string
}

// RetrieveObjectState retrieves the object state
func (c *Client) RetrieveObjectState(ctx context.Context, params RetrieveObjectStateQueryParams) (*ObjectState, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/object-state/", c.baseURL.String()))
	if err != nil {
		return nil, err
	}
	queryParams := endpointURL.Query()

	queryParams.Set("object_type", params.ObjectType)
	if params.ObjectID > 0 {
		queryParams.Set("object_id", strconv.Itoa(params.ObjectID))
	}
	for k, v := range params.Params {
		queryParams.Set(k, v)
	}

	endpointURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn("failed to close response body", "error", closeErr)
		}
	}()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var objStateRaw objectStateRaw
	if err := json.Unmarshal(respBodyBytes, &objStateRaw); err != nil {
		return nil, err
	}

	objState, err := extractObjectState(&objStateRaw, params.ObjectType)
	if err != nil {
		return nil, err
	}

	return &ObjectState{
		ObjectID:       objStateRaw.ObjectID,
		ObjectType:     objStateRaw.ObjectType,
		ObjectChangeID: objStateRaw.ObjectChangeID,
		Object:         objState,
	}, nil
}

func extractObjectState(objState *objectStateRaw, objectType string) (netbox.ComparableData, error) {
	if objState == nil {
		return nil, fmt.Errorf("raw object state response is nil")
	}

	dw, err := netbox.NewDataWrapper(objectType)
	if err != nil {
		return nil, err
	}

	wrappedData, err := wrapObjectState(objectType, objState.Object)
	if err != nil {
		return nil, err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:    &dw,
		MatchName: netbox.IpamIPAddressAssignedObjectMatchName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			statusMapToStringHookFunc(),
			netbox.IpamIPAddressAssignedObjectHookFunc(),
		),
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(wrappedData); err != nil {
		return nil, fmt.Errorf("failed to decode ingest entity data %w", err)
	}

	return dw, nil
}

func statusMapToStringHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {

		if f != reflect.Map {
			return data, nil
		}

		raw := data.(map[string]any)

		if len(raw) == 0 {
			return data, nil
		}

		if t == reflect.String && f == reflect.Map {
			val, ok := raw["value"]
			if !ok {
				return data, nil
			}
			return val, nil
		}

		return data, nil
	}
}

// ChangeSetRequest represents a apply change set request
// type ChangeSetRequest changeset.ChangeSet
type ChangeSetRequest struct {
	ChangeSetID string   `json:"change_set_id"`
	ChangeSet   []Change `json:"change_set"`
}

// Change represents a change
type Change struct {
	ChangeID      string `json:"change_id"`
	ChangeType    string `json:"change_type"`
	ObjectType    string `json:"object_type"`
	ObjectID      *int   `json:"object_id,omitempty"`
	ObjectVersion *int   `json:"object_version,omitempty"`
	Data          any    `json:"data"`
}

// ChangeSetResponse represents an apply change set response
type ChangeSetResponse struct {
	ChangeSetID string `json:"change_set_id"`
	Result      string `json:"result"`
	Errors      any    `json:"errors"`
}

// ApplyChangeSet applies a change set
func (c *Client) ApplyChangeSet(ctx context.Context, payload ChangeSetRequest) (*ChangeSetResponse, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/apply-change-set/", c.baseURL.String()))
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	c.logger.Info("apply change set", "payload", string(reqBody))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn("failed to close response body", "error", closeErr)
		}
	}()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body %w", err)
	}

	c.logger.Info("apply change set", "response", string(respBytes))

	var changeSetResponse ChangeSetResponse
	if err = json.Unmarshal(respBytes, &changeSetResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Info(fmt.Sprintf("request POST %s failed", req.URL.String()), "statusCode", resp.StatusCode, "response", changeSetResponse)
		return &changeSetResponse, fmt.Errorf("request POST %s failed - %q", req.URL.String(), resp.Status)
	}
	return &changeSetResponse, nil
}

func wrapObjectState(dataType string, object any) (any, error) {
	switch dataType {
	case netbox.DcimDeviceObjectType:
		return struct {
			Device any
		}{
			Device: object,
		}, nil
	case netbox.DcimDeviceRoleObjectType:
		return struct {
			DeviceRole any
		}{
			DeviceRole: object,
		}, nil
	case netbox.DcimDeviceTypeObjectType:
		return struct {
			DeviceType any
		}{
			DeviceType: object,
		}, nil
	case netbox.DcimInterfaceObjectType:
		return struct {
			Interface any
		}{
			Interface: object,
		}, nil
	case netbox.DcimManufacturerObjectType:
		return struct {
			Manufacturer any
		}{
			Manufacturer: object,
		}, nil
	case netbox.DcimPlatformObjectType:
		return struct {
			Platform any
		}{
			Platform: object,
		}, nil
	case netbox.DcimSiteObjectType:
		return struct {
			Site any
		}{
			Site: object,
		}, nil
	case netbox.ExtrasTagObjectType:
		return struct {
			Tag any
		}{
			Tag: object,
		}, nil
	case netbox.IpamIPAddressObjectType:
		return struct {
			IPAddress any
		}{
			IPAddress: object,
		}, nil
	case netbox.IpamPrefixObjectType:
		return struct {
			Prefix any
		}{
			Prefix: object,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported data type %s", dataType)
	}
}
