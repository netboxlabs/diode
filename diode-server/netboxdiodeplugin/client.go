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
	"sync"
	"time"

	diodeErrors "github.com/netboxlabs/diode/diode-server/errors"
	"github.com/netboxlabs/diode/diode-server/reconciler/changeset"

	"golang.org/x/time/rate"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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
	userAgent string
}

func newAPIRoundTripper(next http.RoundTripper) (http.RoundTripper, error) {
	return &apiRoundTripper{
		transport: next,
		userAgent: userAgent(),
	}, nil
}

// RoundTrip implements the RoundTripper interface
func (rt *apiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to ensure thread safety
	req2 := req.Clone(req.Context())

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
	logger       *slog.Logger
	httpClient   *http.Client
	baseURL      *url.URL
	limiter      *rate.Limiter
	clientID     string
	clientSecret string
	tokenUrl     string
	token        string
	tokenMutex   sync.Mutex
	maxRetries   int
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
func NewClient(logger *slog.Logger, clientID, clientSecret, tokenUrl string, rateLimitRps, rateLimitBurstRps int, maxRetries int) (*Client, error) {
	transport := NewHTTPTransport()

	rt, err := newAPIRoundTripper(transport)
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
		logger:       logger,
		httpClient:   httpClient,
		baseURL:      u,
		limiter:      rate.NewLimiter(rate.Limit(rateLimitRps), rateLimitBurstRps),
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenUrl:     tokenUrl,
		tokenMutex:   sync.Mutex{},
		maxRetries:   maxRetries,
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
	jsonBytes, err := protojson.Marshal(proto)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(jsonBytes), nil
}

func (c *Client) Authenticate(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	body := url.Values{}
	body.Add("client_id", c.clientID)
	body.Add("client_secret", c.clientSecret)
	body.Add("grant_type", "client_credentials")
	req.Body = io.NopCloser(strings.NewReader(body.Encode()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, respBytes)
	}

	var tokenResponse struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(respBytes, &tokenResponse); err != nil {
		return err
	}

	if tokenResponse.Token == "" {
		return fmt.Errorf("token is required but not provided in response")
	}

	c.tokenMutex.Lock()
	c.token = tokenResponse.Token
	c.tokenMutex.Unlock()
	return nil
}

func (c *Client) doRequestWithRetries(ctx context.Context, method, url string, body io.Reader, headers map[string]string) ([]byte, error) {
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			return nil, err
		}

		for key, value := range headers {
			req.Header.Set(key, value)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		c.logger.Debug("Request to netbox", "statusCode", resp.StatusCode, "response", string(respBytes))

		if resp.StatusCode == http.StatusUnauthorized && attempt < c.maxRetries-1 {
			c.logger.Info("received 401, attempting reauthentication and retry", "attempt", attempt+1)
			if err := c.Authenticate(ctx); err != nil {
				return nil, fmt.Errorf("reauthentication failed: %w", err)
			}
			continue // retry
		}

		// return errors with 4xx status code
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, changeset.NewError("generate diff failed", diodeErrors.ErrCodeOpsGenerateDiff, respBytes)
		}

		return respBytes, nil // success
	}

	return nil, fmt.Errorf("max retries reached after receiving 401 responses")
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

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", c.token),
	}

	branchID := strings.TrimSpace(payload.BranchID)
	if branchID != "" {
		headers[NetBoxBranchHeader] = branchID
	}

	respBytes, err := c.doRequestWithRetries(ctx, http.MethodPost, endpointURL.String(), bytes.NewBuffer(reqBody), headers)
	if err != nil {
		return nil, err
	}

	var changeSetResult ChangeSetResult
	if err = json.Unmarshal(respBytes, &changeSetResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
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

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", c.token),
	}

	branchID := strings.TrimSpace(payload.BranchID)
	if branchID != "" {
		headers[NetBoxBranchHeader] = branchID
	}

	respBytes, err := c.doRequestWithRetries(ctx, http.MethodPost, endpointURL.String(), bytes.NewBuffer(reqBody), headers)
	if err != nil {
		return nil, err
	}

	var changeSetResult ChangeSetResult
	if err = json.Unmarshal(respBytes, &changeSetResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return &changeSetResult, nil
}

// GetToken returns the current token
func (c *Client) GetToken() string {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	return c.token
}
