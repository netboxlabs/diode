package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/kelseyhightower/envconfig"
)

// Server is a auth Server
type Server struct {
	config     Config
	logger     *slog.Logger
	httpServer *http.Server
	mux        *http.ServeMux
}

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
func NewServer(_ context.Context, logger *slog.Logger) (*Server, error) {
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

func (s *Server) GetMux() *http.ServeMux {
	return s.mux
}

// Stop stops the server
func (s *Server) Stop() error {
	s.logger.Info("stopping component", "name", s.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("error during server shutdown", "error", err)
		return fmt.Errorf("error during server shutdown: %w", err)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("error reading request body", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jwksURL := s.config.OAuth2.PublicServerURL + "/.well-known/jwks.json"
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		s.logger.Error("error getting JWKS", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := jwt.Parse(string(body), jwks.Keyfunc)

	if err != nil || !token.Valid {
		writeJSON(w, http.StatusOK, IntrospectResponse{Active: false})
		return
	}

	s.logger.Info("introspecting request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
		"payload", string(body),
		"token", token,
	)

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeJSON(w, http.StatusOK, IntrospectResponse{Active: false})
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

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
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
	req, err := http.NewRequest(r.Method, s.config.OAuth2.PublicServerURL+"/oauth2/token", r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Request failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
