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
	"strconv"
	"strings"
	"time"

	"github.com/netboxlabs/diode/diode-server/gen/diode/v1/reconcilerpb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

	// NetBoxBranchHeader is an HTTP header that indicates the NetBox branch to target
	NetBoxBranchHeader = "X-NetBox-Branch"
	// NetBoxBranchParam is a query parameter that indicates the NetBox branch to target
	NetBoxBranchParam = "_branch"
)

var (
	// ErrInvalidTimeout is an error for invalid timeout value
	ErrInvalidTimeout = errors.New("invalid timeout value")

	// ErrApplyChangeSetFailed is an error for failed to apply change set
	ErrApplyChangeSetFailed = errors.New("failed to apply change set")
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

// ApplyChangeSetError represents an error when applying a change set
type ApplyChangeSetError struct {
	Message string
	Code    int
	Details ApplyChangeSetResponse
}

// Error returns the NetBoxDiodePluginError message
func (e *ApplyChangeSetError) Error() string {
	detailsErrorsJSON, _ := json.Marshal(e.Details.Errors)
	return fmt.Sprintf("msg: %s, code: %d, change set id: %s, result: %s, errors: %s", e.Message, e.Code, e.Details.ChangeSetID, e.Details.Result, detailsErrorsJSON)
}

// NewApplyChangeSetError creates a new ApplyChangeSetError
func NewApplyChangeSetError(msg string, code int, response ApplyChangeSetResponse) error {
	return &ApplyChangeSetError{
		Message: msg,
		Code:    code,
		Details: response,
	}
}

// ToIngestionError converts ApplyChangeSetError to *reconcilerpb.IngestionError
func (e *ApplyChangeSetError) ToIngestionError() *reconcilerpb.IngestionError {
	changeSetErrors := make([]*reconcilerpb.IngestionError_Details_Error, 0)

	ingestionErr := &reconcilerpb.IngestionError{
		Message: e.Message,
		Code:    int32(e.Code),
		Details: &reconcilerpb.IngestionError_Details{
			ChangeSetId: e.Details.ChangeSetID,
			Result:      e.Details.Result,
		},
	}

	if len(e.Details.Errors) > 0 {
		for _, detailsErr := range e.Details.Errors {
			changeID := detailsErr["change_id"]
			for errKey, errValue := range detailsErr {
				if errKey == "change_id" {
					continue
				}
				changeSetErrors = append(changeSetErrors, &reconcilerpb.IngestionError_Details_Error{
					ChangeId: changeID,
					Error:    errValue,
				})
			}
		}
		ingestionErr.Details.Errors = changeSetErrors
	}

	return ingestionErr
}

// NetBoxAPI is the interface for the NetBox Diode plugin API
type NetBoxAPI interface {
	// GenerateDiff generates diff between ingested entity and NetBox object state
	GenerateDiff(context.Context, GenerateDiffRequest) (*GenerateDiffResponse, error)

	// ApplyChangeSet applies a change set
	ApplyChangeSet(context.Context, ApplyChangeSetRequest) (*ApplyChangeSetResponse, error)
}

// Client is a NetBox Diode plugin client
type Client struct {
	logger     *slog.Logger
	httpClient *http.Client
	baseURL    *url.URL
}

// NewHTTPTransport creates a http Transport Layer
func NewHTTPTransport() *http.Transport {
	return &http.Transport{
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
}

// NewClient creates a new NetBox Diode plugin client
func NewClient(logger *slog.Logger, apiKey string) (*Client, error) {
	transport := NewHTTPTransport()

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

// GenerateDiffRequest represents a generate diff request
type GenerateDiffRequest struct {
	ObjectType string          `json:"object_type"`
	BranchID   string          `json:"-"` // Supplied as header
	Entity     proto.Message   `json:"-"`
	EntityJSON json.RawMessage `json:"entity"` // Variable structure based on object type
}

// GenerateDiffResponse represents a diff generated by
// NetBox against an ingested entity
type GenerateDiffResponse struct {
	ChangeSet []Change `json:"change_set"`
	BranchID  string   `json:"branch_id"`
}

func protoToJSON(proto proto.Message) (json.RawMessage, error) {
	jsonBytes, err := protojson.Marshal(proto)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(jsonBytes), nil
}

// GenerateDiff generates a diff between an ingested entity and NetBox object state
func (c *Client) GenerateDiff(ctx context.Context, payload GenerateDiffRequest) (*GenerateDiffResponse, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/generate-diff/", c.baseURL.String()))
	if err != nil {
		return nil, err
	}

	if payload.Entity != nil {
		payload.EntityJSON, err = protoToJSON(payload.Entity)
		if err != nil {
			return nil, err
		}
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	branchID := strings.TrimSpace(payload.BranchID)
	if branchID != "" {
		req.Header.Set(NetBoxBranchHeader, branchID)
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

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body %w", err)
	}

	var generateDiffResponse GenerateDiffResponse
	if err = json.Unmarshal(respBytes, &generateDiffResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %w", err)
	}

	return &generateDiffResponse, nil
}

// ApplyChangeSetRequest represents a apply change set request
// type ChangeSetRequest changeset.ChangeSet
type ApplyChangeSetRequest struct {
	ChangeSetID string   `json:"change_set_id"`
	ChangeSet   []Change `json:"change_set"`
	BranchID    string   `json:"-"` // Supplied as header
}

// Change represents a change
type Change struct {
	ChangeID           string          `json:"change_id"`
	ChangeType         string          `json:"change_type"`
	ObjectType         string          `json:"object_type"`
	ObjectID           *int            `json:"object_id,omitempty"`
	ObjectVersion      *int            `json:"object_version,omitempty"`
	Data               json.RawMessage `json:"data"`
	Before             json.RawMessage `json:"before,omitempty"`
	ObjectPrimaryValue string          `json:"object_primary_value,omitempty"`
}

// ApplyChangeSetResponse represents an apply change set response
type ApplyChangeSetResponse struct {
	ChangeSetID string              `json:"change_set_id"`
	Result      string              `json:"result"`
	Errors      []map[string]string `json:"errors"`
}

// ApplyChangeSet applies a change set
func (c *Client) ApplyChangeSet(ctx context.Context, payload ApplyChangeSetRequest) (*ApplyChangeSetResponse, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/apply-change-set/", c.baseURL.String()))
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("apply change set", "payload", string(reqBody))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL.String(), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	branchID := strings.TrimSpace(payload.BranchID)
	if branchID != "" {
		req.Header.Set(NetBoxBranchHeader, branchID)
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

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body %w", err)
	}

	c.logger.Debug("apply change set", "response", string(respBytes))

	var changeSetResponse ApplyChangeSetResponse
	if err = json.Unmarshal(respBytes, &changeSetResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %w", err)
	}

	// return errors with 4xx status code
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, NewApplyChangeSetError(ErrApplyChangeSetFailed.Error(), resp.StatusCode, changeSetResponse)
	}

	return &changeSetResponse, nil
}
