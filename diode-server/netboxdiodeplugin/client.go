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

	"golang.org/x/time/rate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	diodeErrors "github.com/netboxlabs/diode/diode-server/errors"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

const (
	// SDKName is the name of the SDK
	SDKName = "netbox-diode-plugin-sdk-go"

	// SDKVersion is the version of the SDK
	SDKVersion = "0.1.0" // TODO: consider making this same as diode-reconciler version

	// BaseURLEnvVarName is the environment variable name for the NetBox Diode plugin HTTP base URL
	BaseURLEnvVarName = "NETBOX_DIODE_PLUGIN_API_BASE_URL"

	// TLSSkipVerifyEnvVarName is the environment variable name for NetBox Diode plugin TLS verification
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

// ErrInvalidTimeout is an error for invalid timeout value
var ErrInvalidTimeout = errors.New("invalid timeout value")

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

// ChangeSetResult represents a change set result
type ChangeSetResult struct {
	ID        string          `json:"id"`
	ChangeSet *ChangeSet      `json:"change_set"`
	Errors    json.RawMessage `json:"errors"`
}

// ChangeSet represents a change set
type ChangeSet struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
	Branch  *Branch  `json:"branch"`
}

// Change represents a change
type Change struct {
	ID                 string          `json:"id"`
	ChangeType         string          `json:"change_type"`
	ObjectType         string          `json:"object_type"`
	ObjectID           *int            `json:"object_id,omitempty"`
	RefID              *string         `json:"ref_id,omitempty"`
	ObjectVersion      *int            `json:"object_version,omitempty"`
	Data               json.RawMessage `json:"data"`
	Before             json.RawMessage `json:"before,omitempty"`
	ObjectPrimaryValue string          `json:"object_primary_value,omitempty"`
	NewRefs            []string        `json:"new_refs,omitempty"`
}

// Branch represents a NetBox branch details
type Branch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NetBoxAPI is the interface for the NetBox Diode plugin API
type NetBoxAPI interface {
	// GenerateDiff generates diff between ingested entity and NetBox object state
	GenerateDiff(context.Context, GenerateDiffRequest) (*ChangeSetResult, error)

	// ApplyChangeSet applies a change set
	ApplyChangeSet(context.Context, ApplyChangeSetRequest) (*ChangeSetResult, error)
}

// Client is a NetBox Diode plugin client
type Client struct {
	logger     *slog.Logger
	httpClient *http.Client
	baseURL    *url.URL
	limiter    *rate.Limiter
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
func NewClient(logger *slog.Logger, apiKey string, rateLimitRps, rateLimitBurstRps int) (*Client, error) {
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

	if rateLimitRps <= 0 || rateLimitBurstRps <= 0 {
		return nil, fmt.Errorf("invalid rate limit values: %d %d", rateLimitRps, rateLimitBurstRps)
	}

	client := &Client{
		logger:     logger,
		httpClient: httpClient,
		baseURL:    u,
		limiter:    rate.NewLimiter(rate.Limit(rateLimitRps), rateLimitBurstRps),
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

func protoToJSON(proto proto.Message) (json.RawMessage, error) {
	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
	}
	jsonBytes, err := marshaler.Marshal(proto)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(jsonBytes), nil
}

// GenerateDiff generates a diff between an ingested entity and NetBox object state
func (c *Client) GenerateDiff(ctx context.Context, payload GenerateDiffRequest) (*ChangeSetResult, error) {
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

	if err := c.limiter.Wait(ctx); err != nil {
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

	c.logger.Debug("generate diff", "statusCode", resp.StatusCode, "response", string(respBytes))

	var changeSetResult ChangeSetResult
	if err = json.Unmarshal(respBytes, &changeSetResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %w", err)
	}

	// return errors with 4xx status code
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, changeset.NewError("generate diff failed", diodeErrors.ErrCodeOpsGenerateDiff, respBytes)
	}

	return &changeSetResult, nil
}

// ApplyChangeSetRequest represents a apply change set request
// type ChangeSetRequest changeset.ChangeSet
type ApplyChangeSetRequest struct {
	ID       string   `json:"id"`
	Changes  []Change `json:"changes"`
	BranchID string   `json:"-"` // Supplied as header
}

// ApplyChangeSet applies a change set
func (c *Client) ApplyChangeSet(ctx context.Context, payload ApplyChangeSetRequest) (*ChangeSetResult, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/apply-change-set/", c.baseURL.String()))
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if err := c.limiter.Wait(ctx); err != nil {
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

	var changeSetResult ChangeSetResult
	if err = json.Unmarshal(respBytes, &changeSetResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %w", err)
	}

	// return errors with 4xx status code
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, changeset.NewError("apply change set failed", diodeErrors.ErrCodeOpsApplyChangeSet, respBytes)
	}

	return &changeSetResult, nil
}
