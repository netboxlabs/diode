package auth

import (
	"bytes"
	"context"
	"encoding/json"
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

	"github.com/netboxlabs/diode/diode-server/authutil"
)

const (
	// DefaultTokenOwnerID is the default owner of user created clients
	DefaultTokenOwnerID = "diode/user"
)

// Server is a auth Server
type Server struct {
	config         Config
	logger         *slog.Logger
	httpServer     *http.Server
	mux            *http.ServeMux
	tokenParser    TokenParser
	clientManager  ClientManager
	tokenOwnership TokenOwnershipProvider
}

// IntrospectResponse is the response for the introspect request
type IntrospectResponse struct {
	Active    bool   `json:"active"`
	Subject   string `json:"sub,omitempty"`
	Scope     string `json:"scope,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
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
	Data          []ClientResponse `json:"data"`
	NextPageToken string           `json:"next_page_token,omitempty"`
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

// TokenOwnershipProvider determines the owner of a token
type TokenOwnershipProvider interface {
	TokenOwnerID(ctx context.Context, token string) (string, error)
}

// DefaultTokenOwner is a default implementation of TokenOwnershipProvider
type DefaultTokenOwner struct{}

// TokenOwnerID returns the owner of a token
func (p *DefaultTokenOwner) TokenOwnerID(_ context.Context, _ string) (string, error) {
	return DefaultTokenOwnerID, nil
}

// NewServer creates a new auth server
func NewServer(_ context.Context, logger *slog.Logger, tokenParser TokenParser, clientManager ClientManager, tokenOwnership TokenOwnershipProvider) (*Server, error) {
	var cfg Config
	envconfig.MustProcess("", &cfg)

	mux := http.NewServeMux()

	server := &Server{
		config: cfg,
		logger: logger,
		mux:    mux,
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
			Handler: mux,
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
	s.mux.HandleFunc("DELETE /clients/{clientID}", s.deleteClient)
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
	jwksURL := s.config.OAuth2.PublicServerURL + "/.well-known/jwks.json"
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, NewAuthError("failed to get JWKS", http.StatusInternalServerError)
	}

	token, err := s.tokenParser.Parse(jwtToken, jwks.Keyfunc)
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

	// Create a new request with the same method and buffered body
	req, err := http.NewRequestWithContext(r.Context(), r.Method, s.config.OAuth2.PublicServerURL+"/oauth2/token", bytes.NewReader(bodyBuf.Bytes()))
	if err != nil {
		s.logger.Error("failed to create request to token endpoint", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Use a custom HTTP client with a timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("failed to send token request", "error", err)
		http.Error(w, "failed to obtain the token", http.StatusBadGateway)
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// Copy headers from the response (avoid duplication if needed)
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value) // Use Set() instead if you want to overwrite duplicates
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		// Response headers already sent — cannot modify response at this point
		s.logger.Error("failed to stream response body to client", "error", err)
	}
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
		Data:          make([]ClientResponse, 0, len(clients.Clients)),
		NextPageToken: clients.NextPageToken,
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
