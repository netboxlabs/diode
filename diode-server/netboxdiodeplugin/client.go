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

	"github.com/hashicorp/go-retryablehttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/netboxlabs/diode/diode-server/authutil"
	diodeErrors "github.com/netboxlabs/diode/diode-server/errors"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"
)

const (
	// SDKName is the name of the SDK
	SDKName = "netbox-diode-plugin-sdk-go"

	// SDKVersion is the version of the SDK
	SDKVersion = "0.1.0" // TODO: consider making this same as diode-reconciler version

	// TLSSkipVerifyEnvVarName is the environment variable name for NetBox Diode plugin TLS verification
	TLSSkipVerifyEnvVarName = "NETBOX_DIODE_PLUGIN_SKIP_TLS_VERIFY"

	// TimeoutSecondsEnvVarName is the environment variable name for the NetBox Diode plugin HTTP timeout
	TimeoutSecondsEnvVarName = "NETBOX_DIODE_PLUGIN_API_TIMEOUT_SECONDS"

	// defaultHTTPTimeoutSeconds is the default HTTP timeout
	defaultHTTPTimeoutSeconds = 5

	// NetBoxBranchHeader is an HTTP header that indicates the NetBox branch to target
	NetBoxBranchHeader = "X-NetBox-Branch"

	// NetBoxBranchParam is a query parameter that indicates the NetBox branch to target
	NetBoxBranchParam = "_branch"
)

// ErrInvalidTimeout is an error for invalid timeout value
var ErrInvalidTimeout = errors.New("invalid timeout value")

// ErrDefaultBranchNotFound is returned when the default branch is not found (e.g. no endpoint available in older plugin versions)
var ErrDefaultBranchNotFound = errors.New("default branch not found")

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

// Client is a NetBox Diode plugin client
type Client struct {
	logger    *slog.Logger
	http      *http.Client
	baseURL   *url.URL
	userAgent string
	limiter   *rate.Limiter
}

// headerRoundTripper adds common headers to all requests
type headerRoundTripper struct {
	transport http.RoundTripper
}

// RoundTrip adds common headers to all requests
func (rt *headerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())

	reqClone.Header.Set("User-Agent", SDKName+"/"+SDKVersion)
	reqClone.Header.Set("Content-Type", "application/json")
	reqClone.Header.Set("Accept", "application/json")

	return rt.transport.RoundTrip(reqClone)
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

// ClientOptions represents the options for creating a new NetBox Diode plugin client
type ClientOptions struct {
	// BaseURL is the base URL of the NetBox Diode plugin API
	BaseURL string

	// ClientID is the client ID for the NetBox Diode plugin API
	ClientID string
	// ClientSecret is the client secret for the NetBox Diode plugin API
	ClientSecret string
	// TokenURL is the URL of the token endpoint
	TokenURL string
	// TokenEndpointParams are extra parameters provided to the token endpoint
	TokenEndpointParams url.Values

	// Logger is the logger for the NetBox Diode plugin client
	Logger *slog.Logger

	// RateLimitRPS is the rate limit for the NetBox Diode plugin client
	RateLimitRPS int
	// RateLimitBurstRPS is the rate limit burst for the NetBox Diode plugin client
	RateLimitBurstRPS int
	// MaxRetries is the maximum number of retries for the NetBox Diode plugin client
	MaxRetries int
}

// NewClient creates a new NetBox Diode plugin client
func NewClient(options ClientOptions) (*Client, error) {
	u, err := url.Parse(options.BaseURL)
	if err != nil {
		return nil, err
	}

	if len(options.ClientID) == 0 || len(options.ClientSecret) == 0 {
		return nil, fmt.Errorf("client ID or secret not provided")
	}

	if len(options.TokenURL) == 0 {
		return nil, fmt.Errorf("token URL not provided")
	}

	timeout, err := httpTimeout()
	if err != nil {
		return nil, err
	}

	if options.RateLimitRPS <= 0 || options.RateLimitBurstRPS <= 0 {
		return nil, fmt.Errorf("invalid rate limit values: %d %d", options.RateLimitRPS, options.RateLimitBurstRPS)
	}

	t, err := url.Parse(options.TokenURL)
	if err != nil {
		return nil, err
	}

	baseTransport := NewHTTPTransport()
	// Use per-request timeout on the transport instead of http.Client.Timeout.
	// Client.Timeout cancels the request context, which prevents retryablehttp
	// from retrying. ResponseHeaderTimeout only limits the server response wait
	// per attempt, allowing retries on timeout errors.
	baseTransport.ResponseHeaderTimeout = timeout

	headerRT := &headerRoundTripper{
		transport: baseTransport,
	}

	oauthConfig := &clientcredentials.Config{
		ClientID:       options.ClientID,
		ClientSecret:   options.ClientSecret,
		TokenURL:       t.String(),
		Scopes:         []string{authutil.ScopeNetBoxRead, authutil.ScopeNetBoxWrite},
		EndpointParams: options.TokenEndpointParams,
	}

	oauthTransport := &oauth2.Transport{
		Source: oauthConfig.TokenSource(context.Background()),
		Base:   otelhttp.NewTransport(headerRT),
	}

	rhttp := retryablehttp.NewClient()
	rhttp.HTTPClient = &http.Client{Transport: oauthTransport}
	rhttp.RetryMax = options.MaxRetries
	rhttp.RetryWaitMin = 150 * time.Millisecond
	rhttp.RetryWaitMax = 2 * time.Second
	rhttp.Logger = options.Logger

	httpClient := rhttp.StandardClient()

	client := &Client{
		logger:    options.Logger,
		http:      httpClient,
		baseURL:   u,
		userAgent: fmt.Sprintf("%s/%s", SDKName, SDKVersion),
		limiter:   rate.NewLimiter(rate.Limit(options.RateLimitRPS), options.RateLimitBurstRPS),
	}

	return client, nil
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

// NetBoxAPI is the interface for the NetBox Diode plugin API
type NetBoxAPI interface {
	// GenerateDiff generates diff between ingested entity and NetBox object state
	GenerateDiff(context.Context, GenerateDiffRequest) (*ChangeSetResult, error)

	// ApplyChangeSet applies a change set
	ApplyChangeSet(context.Context, ApplyChangeSetRequest) (*ChangeSetResult, error)

	// GetDefaultBranch gets the default branch from NetBox plugin settings
	GetDefaultBranch(context.Context) (*Branch, error)
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

	branchID := strings.TrimSpace(payload.BranchID)
	if branchID != "" {
		req.Header.Set(NetBoxBranchHeader, branchID)
	}

	resp, err := c.http.Do(req)
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

	branchID := strings.TrimSpace(payload.BranchID)
	if branchID != "" {
		req.Header.Set(NetBoxBranchHeader, branchID)
	}

	resp, err := c.http.Do(req)
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

// GetDefaultBranchResponse represents the response from the default-branch endpoint
type GetDefaultBranchResponse struct {
	Branch *Branch `json:"branch"`
}

// GetDefaultBranch gets the default branch from NetBox plugin settings
func (c *Client) GetDefaultBranch(ctx context.Context) (*Branch, error) {
	endpointURL, err := url.Parse(fmt.Sprintf("%s/default-branch/", c.baseURL.String()))
	if err != nil {
		return nil, err
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
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

	// Log response body only for successful responses to avoid noise from 404s (endpoint not available on older plugin versions)
	if resp.StatusCode < http.StatusBadRequest {
		c.logger.Debug("get default branch", "statusCode", resp.StatusCode, "response", string(respBytes))
	} else {
		c.logger.Debug("get default branch", "statusCode", resp.StatusCode)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		// Return sentinel error for 404 to allow callers to cache and handle gracefully
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrDefaultBranchNotFound
		}
		return nil, fmt.Errorf("get default branch failed with status %d: %s", resp.StatusCode, string(respBytes))
	}

	var defaultBranchResponse GetDefaultBranchResponse
	if err = json.Unmarshal(respBytes, &defaultBranchResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body %w", err)
	}

	return defaultBranchResponse.Branch, nil
}
