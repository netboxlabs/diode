package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kelseyhightower/envconfig"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"

	"github.com/netboxlabs/diode/diode-server/authutil"
	"github.com/netboxlabs/diode/diode-server/telemetry"
)

const (
	// DefaultTokenOwnerID is the default owner of user created clients
	DefaultTokenOwnerID = "diode/user"
)

// upstreamTokenTimeout bounds a single upstream token request. The upstream call is
// detached from the caller's context, so this timeout is the only thing that unwinds a
// stalled upstream transaction and must not be removed.
const upstreamTokenTimeout = 10 * time.Second

// Server is a auth Server
type Server struct {
	config         Config
	keyfunc        keyfunc.Keyfunc
	logger         *slog.Logger
	httpServer     *http.Server
	httpClient     *http.Client
	mux            *http.ServeMux
	tokenParser    TokenParser
	clientManager  ClientManager
	tokenOwnership TokenOwnershipProvider
	decorators     []ClientInfoDecorator
	tokenCache     *tokenCache
	issuanceGate   TokenIssuanceGate
	metrics        Metrics
}

// upstreamTokenResponse is a buffered upstream token response. Token responses are
// small, and buffering lets one upstream call be replayed to every single-flight waiter
// as well as stored in the cache.
type upstreamTokenResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

// IntrospectResponse is the response for the introspect request
type IntrospectResponse struct {
	Active    bool     `json:"active"`
	Subject   string   `json:"sub,omitempty"`
	Scope     string   `json:"scope,omitempty"`
	ExpiresAt int64    `json:"exp,omitempty"`
	IssuedAt  int64    `json:"iat,omitempty"`
	Issuer    string   `json:"iss,omitempty"`
	ClientID  string   `json:"client_id,omitempty"`
	Username  string   `json:"username,omitempty"`
	Audience  []string `json:"aud,omitempty"`
}

// CreateClientRequest request to create client
// contains only fields of ClientInfo that are allowed to be set by the HTTP API
type CreateClientRequest struct {
	ClientName string `json:"client_name"`
	Scope      string `json:"scope"`
}

// ClientResponse response to create/list clients
// contains only fields of ClientInfo that are returned by the HTTP API
type ClientResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	ClientName   string `json:"client_name"`
	Scope        string `json:"scope"`
	CreatedAt    string `json:"created_at"`
}

// ListClientsResponse response to list clients
type ListClientsResponse struct {
	Data           []ClientResponse `json:"data"`
	FirstPageToken string           `json:"first_page_token,omitempty"`
	NextPageToken  string           `json:"next_page_token,omitempty"`
}

// ClientErrorResponse error response to client requests
type ClientErrorResponse struct {
	Error string `json:"error"`
}

func statusFromError(err error) int {
	if authErr, ok := err.(*Error); ok {
		return authErr.StatusCode
	}
	return http.StatusInternalServerError
}

// TokenOwnershipValidationData contains data for validating token ownership
type TokenOwnershipValidationData struct {
	Headers http.Header
}

// TokenOwnershipProvider determines the owner of a token
type TokenOwnershipProvider interface {
	TokenOwnerID(ctx context.Context, token string) (string, error)
	ValidateTokenOwnership(data TokenOwnershipValidationData, claims jwt.MapClaims) error
}

// DefaultTokenOwner is a default implementation of TokenOwnershipProvider
type DefaultTokenOwner struct{}

// TokenOwnerID returns the owner of a token
func (p *DefaultTokenOwner) TokenOwnerID(_ context.Context, _ string) (string, error) {
	return DefaultTokenOwnerID, nil
}

// ValidateTokenOwnership validates the ownership of a token
func (p *DefaultTokenOwner) ValidateTokenOwnership(_ TokenOwnershipValidationData, _ jwt.MapClaims) error {
	return nil
}

// ClientInfoDecorator attaches additional information to a client info
type ClientInfoDecorator interface {
	VisitClientInfo(ctx context.Context, clientInfo *ClientInfo) error
}

// NewServer creates a new auth server
func NewServer(ctx context.Context, logger *slog.Logger, tokenParser TokenParser, clientManager ClientManager, tokenOwnership TokenOwnershipProvider) (*Server, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	mux := http.NewServeMux()

	jwkSetURL := cfg.OAuth2.PublicServerURL + "/.well-known/jwks.json"
	k, err := keyfunc.NewDefaultCtx(ctx, []string{jwkSetURL})
	if err != nil {
		return nil, fmt.Errorf("failed to create keyfunc: %w", err)
	}

	var cache *tokenCache
	if cfg.OAuth2.TokenCache.Enabled {
		cache, err = newTokenCache(cfg.OAuth2.TokenCache)
		if err != nil {
			return nil, err
		}
	}

	server := &Server{
		config:  cfg,
		keyfunc: k,
		logger:  logger,
		mux:     mux,
		httpClient: &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   upstreamTokenTimeout,
		},
		tokenCache: cache,
		httpServer: &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: otelhttp.NewHandler(mux, "auth-http-server", otelhttp.WithMetricAttributesFn(
				func(r *http.Request) []attribute.KeyValue {
					return []attribute.KeyValue{
						attribute.String("http.route", telemetry.ExtractPathFromPattern(r.Pattern)),
					}
				},
			)),
		},
		tokenParser:    tokenParser,
		clientManager:  clientManager,
		tokenOwnership: tokenOwnership,
	}

	server.RegisterHandlers()

	return server, nil
}

// Name returns the name of the server
func (s *Server) Name() string {
	return "auth-http-server"
}

// Start starts the server
func (s *Server) Start(_ context.Context) error {
	s.logger.Info("starting component", "name", s.Name(), "port", s.config.HTTPPort)
	return s.httpServer.ListenAndServe()
}

// GetMux returns the http.ServeMux of the auth server
func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

// Stop stops the server
func (s *Server) Stop() error {
	s.logger.Info("stopping component", "name", s.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shutdown server", "error", err)
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

// RegisterHandlers registers the handlers for the server
func (s *Server) RegisterHandlers() {
	// Handle both with and without trailing slash
	s.mux.HandleFunc("POST /introspect", s.introspect)
	s.mux.HandleFunc("POST /token", s.token)
	s.mux.HandleFunc("POST /clients", s.createClient)
	s.mux.HandleFunc("GET /clients", s.listClients)
	s.mux.HandleFunc("GET /clients/{clientID}", s.getClient)
	s.mux.HandleFunc("DELETE /clients/{clientID}", s.deleteClient)
}

// AddClientInfoDecorator adds a ClientInfoDecorator to the server
// these are called prior to a user generated client being created.
func (s *Server) AddClientInfoDecorator(decorator ClientInfoDecorator) {
	s.decorators = append(s.decorators, decorator)
}

// SetTokenIssuanceGate installs a gate consulted on every token cache hit, so that
// checks the upstream would have performed stay in force for cached responses.
//
// Deployments that enable the token cache and have such checks must install a gate.
// Without one, a cache hit is served on the strength of the credentials alone.
func (s *Server) SetTokenIssuanceGate(gate TokenIssuanceGate) {
	s.issuanceGate = gate
}

// SetMetrics installs a metrics recorder. Metrics are optional; without one the server
// records nothing.
func (s *Server) SetMetrics(metrics Metrics) {
	s.metrics = metrics
}

// introspect handles the introspect request
func (s *Server) introspect(w http.ResponseWriter, r *http.Request) {
	jwtToken, err := s.getAuthToken(r)
	if err != nil {
		s.logger.Error("failed to get auth token", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	claims, err := s.validateToken(jwtToken)
	if err != nil {
		s.logger.Error("failed to validate token", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	err = s.tokenOwnership.ValidateTokenOwnership(TokenOwnershipValidationData{Headers: r.Header}, claims)
	if err != nil {
		s.logger.Error("failed to validate token ownership", "error", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	resp := IntrospectResponse{
		Active:    true,
		Subject:   getStringClaim(claims, "sub"),
		Scope:     getStringClaim(claims, "scope"),
		ExpiresAt: getInt64Claim(claims, "exp"),
		IssuedAt:  getInt64Claim(claims, "iat"),
		Issuer:    getStringClaim(claims, "iss"),
		ClientID:  getStringClaim(claims, "client_id"),
		Username:  getStringClaim(claims, "username"),
	}

	aud := getStringOrStringArrayClaim(claims, "aud")
	if len(aud) > 0 {
		resp.Audience = aud
	}

	err = writeJSON(w, http.StatusOK, resp)
	if err != nil {
		s.logger.Error("failed to write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) getAuthToken(r *http.Request) (string, error) {
	jwtToken := r.Header.Get("Authorization")
	if jwtToken == "" {
		return "", NewAuthError("missing Authorization header", http.StatusUnauthorized)
	}

	// Remove "Bearer " prefix
	const bearerPrefix = "Bearer "
	if len(jwtToken) <= len(bearerPrefix) || jwtToken[:len(bearerPrefix)] != bearerPrefix {
		return "", NewAuthError("invalid Authorization header", http.StatusUnauthorized)
	}

	jwtToken = jwtToken[len(bearerPrefix):]
	return jwtToken, nil
}

func (s *Server) validateToken(jwtToken string) (jwt.MapClaims, error) {
	token, err := s.tokenParser.Parse(jwtToken, s.keyfunc.Keyfunc)
	if err != nil {
		// Invalid token format or signature
		return nil, NewAuthError("failed to validate token", http.StatusUnauthorized)
	}

	if !token.Valid {
		// Token is invalid (e.g., expired)
		return nil, NewAuthError("token is invalid", http.StatusUnauthorized)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, NewAuthError("invalid token claims", http.StatusForbidden)
	}

	return claims, nil
}

func (s *Server) hasScopes(claims jwt.MapClaims, requiredScopes []string) bool {
	scopeClaim := getStringClaim(claims, "scope")
	scopeList := strings.Split(scopeClaim, " ")
	scopeSet := make(map[string]bool)
	for _, scope := range scopeList {
		scopeSet[scope] = true
	}
	for _, scope := range requiredScopes {
		if !scopeSet[scope] {
			return false
		}
	}
	return true
}

func (s *Server) authorizeCall(w http.ResponseWriter, r *http.Request, requiredScopes []string) (string, jwt.MapClaims, bool) {
	jwtToken, err := s.getAuthToken(r)
	if err != nil {
		s.logger.Error("failed to get auth token", "error", err)
		w.WriteHeader(statusFromError(err))
		return "", nil, false
	}

	claims, err := s.validateToken(jwtToken)
	if err != nil {
		s.logger.Error("failed to validate token", "error", err)
		w.WriteHeader(statusFromError(err))
		return "", nil, false
	}

	if !s.hasScopes(claims, requiredScopes) {
		s.logger.Error("missing required scope", "scope_claim", getStringClaim(claims, "scope"), "required_scope", requiredScopes)
		w.WriteHeader(http.StatusForbidden)
		return "", nil, false
	}

	return jwtToken, claims, true
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func getStringOrStringArrayClaim(claims jwt.MapClaims, key string) []string {
	// specification allows some standard claims to be either
	// single strings or list of strings
	if val, ok := claims[key].(string); ok {
		if val == "" {
			return []string{}
		}
		return []string{val}
	}
	if val, ok := claims[key].([]interface{}); ok {
		out := make([]string, 0, len(val))
		for _, v := range val {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return []string{}
}

func getInt64Claim(claims jwt.MapClaims, key string) int64 {
	switch val := claims[key].(type) {
	case float64:
		return int64(val)
	case json.Number:
		i, _ := val.Int64()
		return i
	default:
		return 0
	}
}

// token handles the token request
func (s *Server) token(w http.ResponseWriter, r *http.Request) {
	// Copy and buffer the request body in case it needs to be read again
	var bodyBuf bytes.Buffer
	if _, err := io.Copy(&bodyBuf, r.Body); err != nil {
		s.logger.Error("failed to read request body", "error", err)
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}

	defer func() {
		_ = r.Body.Close()
	}()

	if s.tokenCache == nil {
		s.proxyToken(w, r, bodyBuf.Bytes())
		return
	}

	cred, ok := cacheableTokenRequest(bodyBuf.Bytes(), r.Header)
	if !ok {
		s.recordTokenCacheOutcome(r.Context(), tokenCacheBypass)
		s.proxyToken(w, r, bodyBuf.Bytes())
		return
	}

	key := s.tokenCache.key(cred)

	outcome, served := s.serveCachedToken(w, r, key)
	s.recordTokenCacheOutcome(r.Context(), outcome)
	if served {
		return
	}

	s.fetchToken(w, r, key, cred, bodyBuf.Bytes())
}

// serveCachedToken serves a live cache entry, if one exists and the issuance gate still
// allows it. It returns the outcome and whether a response was written.
func (s *Server) serveCachedToken(w http.ResponseWriter, r *http.Request, key string) (tokenCacheOutcome, bool) {
	now := time.Now()

	entry, ok := s.tokenCache.get(key, now)
	if !ok {
		return tokenCacheMiss, false
	}

	if entry.negative {
		s.writeCachedResponse(w, entry.statusCode, entry.contentType, entry.rawBody)
		return tokenCacheNegativeHit, true
	}

	if !s.allowIssuance(r.Context(), entry.clientID) {
		// Fall through to the upstream rather than inventing a rejection here, so the
		// caller sees the authoritative error. This is never worse than not caching.
		s.tokenCache.remove(key)
		return tokenCacheGateDenied, false
	}

	body, ok := entry.responseBody(now)
	if !ok {
		s.tokenCache.remove(key)
		return tokenCacheMiss, false
	}

	s.writeCachedResponse(w, http.StatusOK, "application/json", body)
	return tokenCacheHit, true
}

// allowIssuance consults the issuance gate.
//
// A gate error is treated as a denial. Denial means the cached response is not served
// and the request goes upstream, so failing closed here costs a cache hit rather than
// availability.
func (s *Server) allowIssuance(ctx context.Context, clientID string) bool {
	if s.issuanceGate == nil {
		return true
	}

	allowed, err := s.issuanceGate.Allow(ctx, clientID)
	if err != nil {
		s.logger.Error("failed to evaluate token issuance gate", "error", err, "client_id", clientID)
		return false
	}
	return allowed
}

// fetchToken obtains a token from the upstream and caches the result.
//
// Concurrent misses on the same key are collapsed into a single upstream request.
// Without this, a synchronised token expiry across callers stampedes the upstream. The
// collapsed request carries the leading caller's headers; callers that share a key
// present the same credentials, so only incidental headers can differ.
func (s *Server) fetchToken(w http.ResponseWriter, r *http.Request, key string, cred tokenCredentials, body []byte) {
	result, err, _ := s.tokenCache.group.Do(key, func() (any, error) {
		resp, err := s.requestToken(r.Context(), r.Header, body)
		if err != nil {
			return nil, err
		}
		s.storeTokenResponse(key, cred, resp)
		return resp, nil
	})
	if err != nil {
		s.logger.Error("failed to send token request", "error", err)
		http.Error(w, "failed to obtain the token", http.StatusBadGateway)
		return
	}

	resp, ok := result.(*upstreamTokenResponse)
	if !ok {
		s.logger.Error("unexpected token request result type")
		http.Error(w, "failed to obtain the token", http.StatusInternalServerError)
		return
	}

	s.writeUpstreamResponse(w, resp)
}

// proxyToken forwards a token request that the cache may not serve.
func (s *Server) proxyToken(w http.ResponseWriter, r *http.Request, body []byte) {
	resp, err := s.requestToken(r.Context(), r.Header, body)
	if err != nil {
		s.logger.Error("failed to send token request", "error", err)
		http.Error(w, "failed to obtain the token", http.StatusBadGateway)
		return
	}

	s.writeUpstreamResponse(w, resp)
}

// requestToken performs the upstream token request.
//
// The call is deliberately detached from the caller's context. A caller that gives up
// must not abort an in-flight upstream transaction, and under single-flight it must not
// abort the request other waiters are relying on. upstreamTokenTimeout is what bounds
// the call instead.
func (s *Server) requestToken(ctx context.Context, header http.Header, body []byte) (*upstreamTokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), upstreamTokenTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.OAuth2.PublicServerURL+"/oauth2/token", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request to token endpoint: %w", err)
	}
	req.Header = header.Clone()

	started := time.Now()
	resp, err := s.httpClient.Do(req)
	s.recordUpstreamTokenRequest(ctx, time.Since(started), err)
	if err != nil {
		return nil, fmt.Errorf("failed to send token request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	return &upstreamTokenResponse{
		statusCode: resp.StatusCode,
		header:     resp.Header.Clone(),
		body:       respBody,
	}, nil
}

// storeTokenResponse caches a successful token response, or a definitive upstream
// rejection of the presented credentials.
func (s *Server) storeTokenResponse(key string, cred tokenCredentials, resp *upstreamTokenResponse) {
	now := time.Now()

	if resp.statusCode != http.StatusOK {
		if negativelyCacheable(resp.statusCode, resp.body) {
			s.tokenCache.putNegative(key, resp.statusCode, resp.header.Get("Content-Type"), resp.body, now)
		}
		return
	}

	// UseNumber keeps upstream numbers exact, so re-marshalling a hit reproduces the
	// original response rather than a float64 approximation of it.
	decoder := json.NewDecoder(bytes.NewReader(resp.body))
	decoder.UseNumber()

	var body map[string]any
	if err := decoder.Decode(&body); err != nil {
		s.logger.Warn("token response is not cacheable", "error", err)
		return
	}

	s.tokenCache.putToken(key, cred, body, now)
}

// writeUpstreamResponse relays an upstream response verbatim.
func (s *Server) writeUpstreamResponse(w http.ResponseWriter, resp *upstreamTokenResponse) {
	for name, values := range resp.header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.statusCode)
	if _, err := w.Write(resp.body); err != nil {
		// Response headers already sent, cannot modify the response at this point
		s.logger.Error("failed to write response body to client", "error", err)
	}
}

// writeCachedResponse writes a response served from cache. There is no upstream response
// to copy headers from, so the headers the upstream would have set are set here.
func (s *Server) writeCachedResponse(w http.ResponseWriter, statusCode int, contentType string, body []byte) {
	if contentType == "" {
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	w.WriteHeader(statusCode)
	if _, err := w.Write(body); err != nil {
		s.logger.Error("failed to write cached response body to client", "error", err)
	}
}

// recordTokenCacheOutcome records how a token request interacted with the cache.
func (s *Server) recordTokenCacheOutcome(ctx context.Context, outcome tokenCacheOutcome) {
	if s.metrics == nil {
		return
	}
	s.metrics.RecordTokenCacheOutcome(ctx, string(outcome))
}

// recordUpstreamTokenRequest records the duration of an upstream token request, split by
// outcome. Separating a timeout from other transport failures is what distinguishes a
// stalled upstream from an unreachable one.
func (s *Server) recordUpstreamTokenRequest(ctx context.Context, duration time.Duration, err error) {
	if s.metrics == nil {
		return
	}

	outcome := "ok"
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		outcome = "deadline_exceeded"
	case err != nil:
		outcome = "error"
	}

	s.metrics.RecordUpstreamTokenRequest(ctx, duration, outcome)
}

func (s *Server) createClient(w http.ResponseWriter, r *http.Request) {
	jwtToken, _, ok := s.authorizeCall(w, r, []string{authutil.ScopeDiodeWrite})
	if !ok {
		return
	}

	ownerID, err := s.tokenOwnership.TokenOwnerID(r.Context(), jwtToken)
	if err != nil {
		s.logger.Error("failed to get token owner ID", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	// Parse the request body
	var createRequest CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
		s.logger.Error("failed to parse request body", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	// Only diode:ingest is allowed/supported currently
	scopes := strings.Split(createRequest.Scope, " ")
	if len(scopes) != 1 || scopes[0] != authutil.ScopeDiodeIngest {
		s.logger.Error("invalid/unsupported scope", "scope", createRequest.Scope)
		err = writeJSON(w, http.StatusBadRequest, ClientErrorResponse{Error: "invalid scope"})
		if err != nil {
			s.logger.Error("failed to write response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Client name must be specified
	if strings.TrimSpace(createRequest.ClientName) == "" {
		s.logger.Error("client name is required")
		err = writeJSON(w, http.StatusBadRequest, ClientErrorResponse{Error: "client name is required"})
		if err != nil {
			s.logger.Error("failed to write response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	clientInfo := ClientInfo{
		ClientName: createRequest.ClientName,
		Scope:      createRequest.Scope,
		Owner:      ownerID,
	}

	// Generate a client ID for the client
	clientInfo.ClientID, err = GenerateClientID(clientInfo)
	if err != nil {
		s.logger.Error("failed to generate client ID", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	// Generate a client secret for the client
	clientInfo.ClientSecret, err = GenerateClientSecret()
	if err != nil {
		s.logger.Error("failed to generate client secret", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, decorator := range s.decorators {
		err = decorator.VisitClientInfo(r.Context(), &clientInfo)
		if err != nil {
			s.logger.Error("failed to decorate client info", "error", err)
			w.WriteHeader(statusFromError(err))
			return
		}
	}

	// Create the client
	created, err := s.clientManager.CreateClient(r.Context(), clientInfo)
	if err != nil {
		s.logger.Error("failed to create client", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	out := ClientResponse{
		ClientID:     created.ClientID,
		ClientSecret: created.ClientSecret,
		ClientName:   created.ClientName,
		Scope:        created.Scope,
		CreatedAt:    created.CreatedAt,
	}

	err = writeJSON(w, http.StatusCreated, out)
	if err != nil {
		s.logger.Error("failed to write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) listClients(w http.ResponseWriter, r *http.Request) {
	jwtToken, _, ok := s.authorizeCall(w, r, []string{authutil.ScopeDiodeRead})
	if !ok {
		return
	}

	ownerID, err := s.tokenOwnership.TokenOwnerID(r.Context(), jwtToken)
	if err != nil {
		s.logger.Error("failed to get token owner ID", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	req := RetrieveClientsRequest{
		Owner:     ownerID,
		PageToken: r.URL.Query().Get("page_token"),
	}

	pageSizeStr := r.URL.Query().Get("page_size")
	if pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			err = writeJSON(w, http.StatusBadRequest, ClientErrorResponse{Error: "invalid page size"})
			if err != nil {
				s.logger.Error("failed to write response", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
		req.PageSize = pageSize
	}

	clients, err := s.clientManager.RetrieveClients(r.Context(), req)
	if err != nil {
		s.logger.Error("failed to list clients", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	out := ListClientsResponse{
		Data:           make([]ClientResponse, 0, len(clients.Clients)),
		FirstPageToken: clients.FirstPageToken,
		NextPageToken:  clients.NextPageToken,
	}
	for _, client := range clients.Clients {
		out.Data = append(out.Data, ClientResponse{
			ClientID:   client.ClientID,
			ClientName: client.ClientName,
			Scope:      client.Scope,
			CreatedAt:  client.CreatedAt,
		})
	}

	err = writeJSON(w, http.StatusOK, out)
	if err != nil {
		s.logger.Error("failed to write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) deleteClient(w http.ResponseWriter, r *http.Request) {
	jwtToken, _, ok := s.authorizeCall(w, r, []string{authutil.ScopeDiodeWrite})
	if !ok {
		return
	}

	ownerID, err := s.tokenOwnership.TokenOwnerID(r.Context(), jwtToken)
	if err != nil {
		s.logger.Error("failed to get token owner ID", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	clientID := r.PathValue("clientID")
	if clientID == "" {
		err = writeJSON(w, http.StatusBadRequest, ClientErrorResponse{Error: "client ID is required"})
		if err != nil {
			s.logger.Error("failed to write response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// get the client and verify ownership
	client, err := s.clientManager.RetrieveClientByID(r.Context(), clientID)
	if err != nil {
		s.logger.Error("failed to get client", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	if client.Owner != ownerID {
		s.logger.Error("client does not belong to requestor", "client_id", clientID, "owner_id", ownerID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// now we can delete the client
	err = s.clientManager.DeleteClientByID(r.Context(), clientID)
	if err != nil {
		s.logger.Error("failed to delete client", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getClient(w http.ResponseWriter, r *http.Request) {
	jwtToken, _, ok := s.authorizeCall(w, r, []string{authutil.ScopeDiodeRead})
	if !ok {
		return
	}

	ownerID, err := s.tokenOwnership.TokenOwnerID(r.Context(), jwtToken)
	if err != nil {
		s.logger.Error("failed to get token owner ID", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	clientID := r.PathValue("clientID")
	if clientID == "" {
		err = writeJSON(w, http.StatusBadRequest, ClientErrorResponse{Error: "client ID is required"})
		if err != nil {
			s.logger.Error("failed to write response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// get the client and verify ownership
	client, err := s.clientManager.RetrieveClientByID(r.Context(), clientID)
	if err != nil {
		s.logger.Error("failed to get client", "error", err)
		w.WriteHeader(statusFromError(err))
		return
	}

	if client.Owner != ownerID {
		s.logger.Error("client does not belong to requestor", "client_id", clientID, "owner_id", ownerID)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = writeJSON(w, http.StatusOK, client)
	if err != nil {
		s.logger.Error("failed to write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
