package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kelseyhightower/envconfig"
)

// Server is a auth Server
type Server struct {
	config      Config
	logger      *slog.Logger
	httpServer  *http.Server
	mux         *http.ServeMux
	tokenParser TokenParser
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

// NewServer creates a new auth server
func NewServer(_ context.Context, logger *slog.Logger, tokenParser TokenParser) (*Server, error) {
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
		tokenParser: tokenParser,
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
}

// introspect handles the introspect request
func (s *Server) introspect(w http.ResponseWriter, r *http.Request) {
	jwtToken := r.Header.Get("Authorization")
	if jwtToken == "" {
		s.logger.Warn("missing Authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	const bearerPrefix = "Bearer "
	if len(jwtToken) > len(bearerPrefix) && jwtToken[:len(bearerPrefix)] == bearerPrefix {
		jwtToken = jwtToken[len(bearerPrefix):]
	}

	jwksURL := s.config.OAuth2.PublicServerURL + "/.well-known/jwks.json"
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		s.logger.Error("failed to get JWKS", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := s.tokenParser.Parse(jwtToken, jwks.Keyfunc)
	if err != nil {
		// Invalid token format or signature
		s.logger.Warn("failed to validate token", "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !token.Valid {
		// Token is invalid (e.g., expired)
		s.logger.Warn("token is invalid")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
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

	err = writeJSON(w, http.StatusOK, resp)
	if err != nil {
		s.logger.Error("failed to write response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
