package netboxdiodeplugin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	// SDKName is the name of the SDK
	SDKName = "netbox-diode-plugin-sdk-go"

	// SDKVersion is the version of the SDK
	SDKVersion = "0.1.0"

	// BaseURLEnvVarName is the environment variable name for the NetBox Diode plugin HTTP base URL
	BaseURLEnvVarName = "NETBOX_DIODE_PLUGIN_API_BASE_URL"

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

// Client is a NetBox Diode plugin client
type Client struct {
	logger     *slog.Logger
	httpClient *http.Client
	baseURL    *url.URL
}

// NewClient creates a new NetBox Diode plugin client
func NewClient(apiKey string, logger *slog.Logger) (*Client, error) {
	rt, err := newAPIRoundTripper(apiKey, http.DefaultTransport)
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

type rawObjectStateResponse struct {
	ObjectType     string          `json:"object_type"`
	ObjectChangeID int             `json:"object_change_id"`
	Object         json.RawMessage `json:"object"`
}

func (c *Client) retrieveRawObjectState(ctx context.Context, objectType string, objectID int, query string) (*rawObjectStateResponse, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/object-state", c.baseURL.String()))
	if err != nil {
		return nil, err
	}
	queryParams := endpointURL.Query()

	queryParams.Set("object_type", objectType)
	if objectID > 0 {
		queryParams.Set("object_id", strconv.Itoa(objectID))
	}
	if len(query) > 0 {
		queryParams.Set("q", query)
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

	var rawResp rawObjectStateResponse
	if err := json.Unmarshal(respBodyBytes, &rawResp); err != nil {
		return nil, err
	}

	return &rawResp, nil
}

// RetrieveDcimDeviceState retrieves the state of a DCIM device
func (c *Client) RetrieveDcimDeviceState(ctx context.Context, objectID int, query string) (*DcimDeviceState, error) {
	resp, err := c.retrieveRawObjectState(ctx, "dcim.device", objectID, query)
	if err != nil {
		return nil, err
	}

	dcimDevice := &DcimDevice{}
	if err := json.Unmarshal(resp.Object, dcimDevice); err != nil {
		return nil, err
	}

	return &DcimDeviceState{
		ObjectChangeID: resp.ObjectChangeID,
		Object:         *dcimDevice,
	}, nil
}

// ApplyChangeSet applies a change set
func (c *Client) ApplyChangeSet(ctx context.Context, changeSet ChangeSetRequest) (*ChangeSetResponse, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/apply-change-set", c.baseURL.String()))
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(changeSet)
	if err != nil {
		return nil, err
	}

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

	var changeSetResponse ChangeSetResponse
	if err = json.Unmarshal(respBytes, &changeSetResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Debug(fmt.Sprintf("request POST %s failed", req.URL.String()), "statusCode", resp.StatusCode, "response", changeSetResponse)
		return &changeSetResponse, fmt.Errorf("request POST %s failed - %q", req.URL.String(), resp.Status)
	}
	return &changeSetResponse, nil
}

// ChangeSetRequest represents a apply change set request
type ChangeSetRequest struct {
	ChangeSetID string   `json:"change_set_id"`
	ChangeSet   []Change `json:"change_set"`
}

// Change represents a change to apply
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
	ChangeSetID string   `json:"change_set_id"`
	Result      string   `json:"result"`
	Errors      []string `json:"errors"`
}

// DcimDevice represents a DCIM device
type DcimDevice struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// DcimDeviceState represents the state of a DCIM device
type DcimDeviceState struct {
	ObjectChangeID int        `json:"object_change_id"`
	Object         DcimDevice `json:"object"`
}
