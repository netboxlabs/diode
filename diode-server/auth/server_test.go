package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gosimple/slug"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/netboxlabs/diode/diode-server/auth"
	"github.com/netboxlabs/diode/diode-server/auth/mocks"
)

type InvalidParser struct{}

func (p InvalidParser) Parse(_ string, _ jwt.Keyfunc) (*jwt.Token, error) {
	return nil, fmt.Errorf("invalid token")
}

type MockTokenParser struct {
	tokenMap map[string]jwt.Token
}

func (p MockTokenParser) Parse(token string, _ jwt.Keyfunc) (*jwt.Token, error) {
	if tok, ok := p.tokenMap[token]; ok {
		return &tok, nil
	}
	return nil, fmt.Errorf("token not found")
}

type ownerInvalid struct{}

func (o ownerInvalid) TokenOwnerID(_ context.Context, _ string) (string, error) {
	return auth.DefaultTokenOwnerID, nil
}

func (o ownerInvalid) ValidateTokenOwnership(_ auth.TokenOwnershipValidationData, _ jwt.MapClaims) error {
	return errors.New("invalid token owner")
}

func TestNewServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupEnv()
	defer teardownEnv()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	defaultOwnership := &auth.DefaultTokenOwner{}
	server, err := auth.NewServer(ctx, logger, InvalidParser{}, nil, defaultOwnership)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Start and stop the server in a separate goroutine
	go func() {
		err = server.Start(ctx)
		require.NoError(t, err)
	}()

	// Wait for the server to start and stop
	time.Sleep(50 * time.Millisecond)
}

func TestIntrospectForInvalidTokens(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupEnv()
	defer teardownEnv()

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			// Return a mock JWKS response
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"keys": [
					{
						"kty": "RSA",
						"kid": "test-key",
						"use": "sig",
						"alg": "RS256",
						"n": "n_value",
						"e": "AQAB"
					}
				]
			}`))
		}
	}))
	defer mockJWKSServer.Close()

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
	defaultOwnership := &auth.DefaultTokenOwner{}
	server, err := auth.NewServer(ctx, logger, InvalidParser{}, nil, defaultOwnership)
	require.NoError(t, err)
	require.NotNil(t, server)

	// Create a test server using the server's mux
	testServer := httptest.NewServer(server.GetMux())
	defer testServer.Close()

	t.Run("Invalid Token", func(t *testing.T) {
		resp, err := makeIntrospectRequest(testServer.URL, "invalid.token.string")
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
	})

	t.Run("Missing Token", func(t *testing.T) {
		resp, err := makeIntrospectRequest(testServer.URL, "")
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
	})

	t.Run("Missing auth header", func(t *testing.T) {
		resp, err := makeIntrospectRequestWithoutAuth(testServer.URL)
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		defer func() {
			_ = resp.Body.Close()
		}()
	})
}

func TestIntrospectForValidTokens(t *testing.T) {
	testToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3Qta2V5IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoiYXBpIiwiZXhwIjoxNjUwMDAwMDAwLCJpYXQiOjE1MDAwMDAwMDAsImNsaWVudF9pZCI6ImNsaWVudDEyMyIsInNjb3BlIjoicmVhZCB3cml0ZSIsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.WcPGXClpKD7Bc1C0CCDA1060E2GGlTfamrd8-W0ghBE"
	tests := []struct {
		name             string
		token            string
		tokenParser      auth.TokenParser
		invalidOwner     bool
		expectedStatus   int
		expectedAudience []string
		expectedSubject  string
		expectedScope    string
		expectedIssuer   string
		expectedClientID string
		expectedUsername string
	}{
		{
			name:  "Valid Token",
			token: testToken,
			tokenParser: &MockTokenParser{
				tokenMap: map[string]jwt.Token{
					testToken: {
						Claims: jwt.MapClaims{
							"iss":       "https://auth.example.com",
							"sub":       "user123",
							"aud":       "api",
							"exp":       time.Now().Add(time.Hour).Unix(),
							"iat":       time.Now().Unix(),
							"client_id": "client123",
							"scope":     "read write",
							"username":  "testuser",
						},
						Valid: true,
					},
				},
			},
			expectedStatus:   http.StatusOK,
			expectedAudience: []string{"api"},
			expectedSubject:  "user123",
			expectedScope:    "read write",
			expectedIssuer:   "https://auth.example.com",
			expectedClientID: "client123",
			expectedUsername: "testuser",
		},
		{
			name: "Valid Token with empty audience",
			tokenParser: &MockTokenParser{
				tokenMap: map[string]jwt.Token{
					testToken: {
						Claims: jwt.MapClaims{
							"iss":       "https://auth.example.com",
							"sub":       "user123",
							"exp":       time.Now().Add(time.Hour).Unix(),
							"iat":       time.Now().Unix(),
							"client_id": "client123",
							"scope":     "read write",
							"username":  "testuser",
						},
						Valid: true,
					},
				},
			},
			token:            testToken,
			expectedStatus:   http.StatusOK,
			expectedAudience: nil,
			expectedSubject:  "user123",
			expectedScope:    "read write",
			expectedIssuer:   "https://auth.example.com",
			expectedClientID: "client123",
			expectedUsername: "testuser",
		},
		{
			name:  "Valid Token with invalid owner",
			token: testToken,
			tokenParser: &MockTokenParser{
				tokenMap: map[string]jwt.Token{
					testToken: {
						Claims: jwt.MapClaims{
							"iss":       "https://auth.example.com",
							"sub":       "user123",
							"aud":       "api",
							"exp":       time.Now().Add(time.Hour).Unix(),
							"iat":       time.Now().Unix(),
							"client_id": "client123",
							"scope":     "read write",
							"username":  "testuser",
						},
						Valid: true,
					},
				},
			},
			invalidOwner:     true,
			expectedStatus:   http.StatusForbidden,
			expectedAudience: []string{"api"},
			expectedSubject:  "user123",
			expectedScope:    "read write",
			expectedIssuer:   "https://auth.example.com",
			expectedClientID: "client123",
			expectedUsername: "testuser",
		},
	}

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			// Return a mock JWKS response
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"keys": [
					{
						"kty": "RSA",
						"kid": "test-key",
						"use": "sig",
						"alg": "RS256",
						"n": "n_value",
						"e": "AQAB"
					}
				]
			}`))
		}
	}))
	defer mockJWKSServer.Close()

	setupEnv()
	defer teardownEnv()

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var ownerProvider auth.TokenOwnershipProvider = &auth.DefaultTokenOwner{}
			if test.invalidOwner {
				ownerProvider = ownerInvalid{}
			}
			server, err := auth.NewServer(ctx, logger, test.tokenParser, nil, ownerProvider)
			require.NoError(t, err)
			require.NotNil(t, server)

			// Create a test server using the server's mux
			testServer := httptest.NewServer(server.GetMux())
			defer testServer.Close()

			data := url.Values{}
			data.Set("token", test.token)

			resp, err := makeIntrospectRequest(testServer.URL, test.token)

			require.NoError(t, err)
			require.Equal(t, test.expectedStatus, resp.StatusCode)
			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				return
			}

			defer func() {
				_ = resp.Body.Close()
			}()

			var introspectResp auth.IntrospectResponse
			err = json.NewDecoder(resp.Body).Decode(&introspectResp)
			require.NoError(t, err)
			require.True(t, introspectResp.Active)
			require.Equal(t, test.expectedSubject, introspectResp.Subject)
			require.Equal(t, test.expectedScope, introspectResp.Scope)
			require.Equal(t, test.expectedIssuer, introspectResp.Issuer)
			require.Equal(t, test.expectedClientID, introspectResp.ClientID)
			require.Equal(t, test.expectedAudience, introspectResp.Audience)
			require.Equal(t, test.expectedUsername, introspectResp.Username)
		})
	}
}

func TestCreateClient(t *testing.T) {
	writeToken := jwt.Token{
		Claims: jwt.MapClaims{
			"exp":       time.Now().Add(time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"client_id": "client123",
			"scope":     "diode:write diode:read",
		},
		Valid: true,
	}
	readOnlyToken := jwt.Token{
		Claims: jwt.MapClaims{
			"exp":       time.Now().Add(time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"client_id": "client123",
			"scope":     "diode:read",
		},
		Valid: true,
	}
	invalidToken := jwt.Token{
		Valid: false,
	}

	validAccessToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3Qta2V5IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoiYXBpIiwiZXhwIjoxNjUwMDAwMDAwLCJpYXQiOjE1MDAwMDAwMDAsImNsaWVudF9pZCI6ImNsaWVudDEyMyIsInNjb3BlIjoicmVhZCB3cml0ZSIsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.WcPGXClpKD7Bc1C0CCDA1060E2GGlTfamrd8-W0ghBE"
	invalidAccessToken := "invalid.token.string"

	type createClientRequest struct {
		ClientName string `json:"client_name"`
		Scope      string `json:"scope"`
	}

	tests := []struct {
		name         string
		accessToken  string
		request      createClientRequest
		parsedToken  jwt.Token
		expectClient auth.ClientInfo
		expectStatus int
	}{
		{
			name:        "can create ingest client",
			parsedToken: writeToken,
			request: createClientRequest{
				ClientName: "Test Client 1",
				Scope:      "diode:ingest",
			},
			expectStatus: http.StatusCreated,
		},
		{
			name:        "cannot create ingest client with read only access",
			parsedToken: readOnlyToken,
			request: createClientRequest{
				ClientName: "Test Client 2",
				Scope:      "diode:ingest",
			},
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "cannot create ingest client with invalid access token",
			accessToken:  invalidAccessToken,
			parsedToken:  invalidToken,
			request:      createClientRequest{},
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:        "cannot create with non ingest scope",
			parsedToken: writeToken,
			request: createClientRequest{
				ClientName: "Test Client 3",
				Scope:      "diode:read",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:        "cannot create with additional scopes",
			parsedToken: writeToken,
			request: createClientRequest{
				ClientName: "Test Client 4",
				Scope:      "diode:ingest diode:write",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:        "cannot create with no scopes",
			parsedToken: writeToken,
			request: createClientRequest{
				ClientName: "Test Client 5",
				Scope:      "",
			},
			expectStatus: http.StatusBadRequest,
		},
		{
			name:        "cannot create with no name",
			parsedToken: writeToken,
			request: createClientRequest{
				ClientName: " ",
				Scope:      "diode:ingest",
			},
			expectStatus: http.StatusBadRequest,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupEnv()
	defer teardownEnv()

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := mockJWKSServer()
	defer mockJWKSServer.Close()

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defaultOwnership := &auth.DefaultTokenOwner{}
			accessToken := test.accessToken
			if accessToken == "" {
				accessToken = validAccessToken
			}
			mockTokenParser := &MockTokenParser{
				tokenMap: map[string]jwt.Token{
					accessToken: test.parsedToken,
				},
			}
			mockClientManager := &mocks.ClientManager{}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
			server, err := auth.NewServer(ctx, logger, mockTokenParser, mockClientManager, defaultOwnership)
			require.NoError(t, err)
			require.NotNil(t, server)

			// Create a test server using the server's mux
			testServer := httptest.NewServer(server.GetMux())
			defer testServer.Close()
			createdAt := time.Now().Format(time.RFC3339)
			if test.expectStatus == http.StatusCreated {
				mockClientManager.On("CreateClient", mock.Anything, mock.Anything).Return(func(_ context.Context, clientInfo auth.ClientInfo) (auth.ClientInfo, error) {
					info := auth.ClientInfo{
						ClientID:     clientInfo.ClientID,
						ClientName:   clientInfo.ClientName,
						Scope:        clientInfo.Scope,
						Owner:        clientInfo.Owner,
						CreatedAt:    createdAt,
						ClientSecret: clientInfo.ClientSecret,
					}
					return info, nil
				})
			}

			body, _ := json.Marshal(test.request)
			req, _ := http.NewRequest(
				"POST",
				testServer.URL+"/clients",
				bytes.NewBuffer(body),
			)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			require.Equal(t, test.expectStatus, resp.StatusCode)

			if test.expectStatus == http.StatusCreated {
				var clientInfo auth.ClientResponse
				err = json.NewDecoder(resp.Body).Decode(&clientInfo)
				require.NoError(t, err)
				require.Equal(t, test.request.ClientName, clientInfo.ClientName)
				require.Equal(t, test.request.Scope, clientInfo.Scope)
				require.Equal(t, createdAt, clientInfo.CreatedAt)
				require.NotEmpty(t, clientInfo.ClientSecret)
				require.Greater(t, len(clientInfo.ClientSecret), 16)
				require.NotEmpty(t, clientInfo.ClientID)
				require.Greater(t, len(clientInfo.ClientID), 17)
				slug := slug.Make(clientInfo.ClientName)
				if len(slug) > 15 {
					slug = slug[:15]
				}
				require.True(t, strings.HasPrefix(clientInfo.ClientID, slug))
			}
		})
	}
}

func TestListClients(t *testing.T) {
	readOnlyToken := jwt.Token{
		Claims: jwt.MapClaims{
			"exp":       time.Now().Add(time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"client_id": "client123",
			"scope":     "diode:read",
		},
		Valid: true,
	}
	ingestOnlyToken := jwt.Token{
		Claims: jwt.MapClaims{
			"exp":       time.Now().Add(time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"client_id": "client124",
			"scope":     "diode:ingest",
		},
		Valid: true,
	}

	invalidToken := jwt.Token{
		Valid: false,
	}

	validAccessToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3Qta2V5IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoiYXBpIiwiZXhwIjoxNjUwMDAwMDAwLCJpYXQiOjE1MDAwMDAwMDAsImNsaWVudF9pZCI6ImNsaWVudDEyMyIsInNjb3BlIjoicmVhZCB3cml0ZSIsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.WcPGXClpKD7Bc1C0CCDA1060E2GGlTfamrd8-W0ghBE"
	invalidAccessToken := "invalid.token.string"

	type listClientsRequest struct {
		PageToken string
		PageSize  string
	}

	tests := []struct {
		name         string
		accessToken  string
		request      listClientsRequest
		parsedToken  jwt.Token
		retrieved    auth.RetrieveClientsResponse
		expect       auth.ListClientsResponse
		expectError  string
		expectStatus int
	}{
		{
			name:        "can list clients",
			parsedToken: readOnlyToken,
			request: listClientsRequest{
				PageSize: "2",
			},
			retrieved: auth.RetrieveClientsResponse{
				Clients: []auth.ClientInfo{
					{
						ClientID:   "test-client-1-abcdef0123567890",
						ClientName: "Test Client 1",
						Scope:      "diode:read",
						Owner:      "diode/user",
						CreatedAt:  "2021-01-01T00:00:00Z",
					},
					{
						ClientID:   "test-client-2-abcdef0123567890",
						ClientName: "Test Client 2",
						Scope:      "diode:read",
						Owner:      "diode/user",
						CreatedAt:  "2021-01-02T00:00:00Z",
					},
				},
				NextPageToken: "",
			},
			expect: auth.ListClientsResponse{
				Data: []auth.ClientResponse{
					{
						ClientID:   "test-client-1-abcdef0123567890",
						ClientName: "Test Client 1",
						Scope:      "diode:read",
						CreatedAt:  "2021-01-01T00:00:00Z",
					},
					{
						ClientID:   "test-client-2-abcdef0123567890",
						ClientName: "Test Client 2",
						Scope:      "diode:read",
						CreatedAt:  "2021-01-02T00:00:00Z",
					},
				},
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "cannot list clients with invalid access token",
			accessToken:  invalidAccessToken,
			parsedToken:  invalidToken,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "cannot list clients without read access",
			parsedToken:  ingestOnlyToken,
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "cannot list clients with invalid page size",
			parsedToken:  readOnlyToken,
			request:      listClientsRequest{PageSize: "not-a-number"},
			expectStatus: http.StatusBadRequest,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupEnv()
	defer teardownEnv()

	mockJWKSServer := mockJWKSServer()
	defer mockJWKSServer.Close()

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defaultOwnership := &auth.DefaultTokenOwner{}
			accessToken := test.accessToken
			if accessToken == "" {
				accessToken = validAccessToken
			}
			mockTokenParser := &MockTokenParser{
				tokenMap: map[string]jwt.Token{
					accessToken: test.parsedToken,
				},
			}
			mockClientManager := &mocks.ClientManager{}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
			server, err := auth.NewServer(ctx, logger, mockTokenParser, mockClientManager, defaultOwnership)
			require.NoError(t, err)
			require.NotNil(t, server)

			testServer := httptest.NewServer(server.GetMux())
			defer testServer.Close()
			if test.expectStatus == http.StatusOK {
				mockClientManager.EXPECT().RetrieveClients(mock.Anything, mock.Anything).Return(test.retrieved, nil)
			}

			u, err := url.Parse(testServer.URL + "/clients")
			require.NoError(t, err)
			q := url.Values{}
			if test.request.PageToken != "" {
				q.Set("page_token", test.request.PageToken)
			}
			if test.request.PageSize != "" {
				q.Set("page_size", test.request.PageSize)
			}
			u.RawQuery = q.Encode()
			req, _ := http.NewRequest("GET", u.String(), nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			require.Equal(t, test.expectStatus, resp.StatusCode)

			if test.expectStatus == http.StatusOK {
				out := auth.ListClientsResponse{}
				err = json.NewDecoder(resp.Body).Decode(&out)
				require.NoError(t, err)
				require.Equal(t, test.expect.Data, out.Data)
				require.Equal(t, test.expect.NextPageToken, out.NextPageToken)
			}
		})
	}
}

func TestDeleteClient(t *testing.T) {
	writeToken := jwt.Token{
		Claims: jwt.MapClaims{
			"exp":       time.Now().Add(time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"client_id": "client123",
			"scope":     "diode:write diode:read",
		},
		Valid: true,
	}
	readOnlyToken := jwt.Token{
		Claims: jwt.MapClaims{
			"exp":       time.Now().Add(time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"client_id": "client123",
			"scope":     "diode:read",
		},
		Valid: true,
	}
	invalidToken := jwt.Token{
		Valid: false,
	}

	validAccessToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3Qta2V5IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoiYXBpIiwiZXhwIjoxNjUwMDAwMDAwLCJpYXQiOjE1MDAwMDAwMDAsImNsaWVudF9pZCI6ImNsaWVudDEyMyIsInNjb3BlIjoicmVhZCB3cml0ZSIsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.WcPGXClpKD7Bc1C0CCDA1060E2GGlTfamrd8-W0ghBE"
	invalidAccessToken := "invalid.token.string"

	tests := []struct {
		name         string
		accessToken  string
		clientID     string
		parsedToken  jwt.Token
		lookupResult *auth.ClientInfo
		lookupErr    error
		expectStatus int
	}{
		{
			name:        "can delete client",
			accessToken: validAccessToken,
			clientID:    "test-client-1-abcdef0123567890",
			parsedToken: writeToken,
			lookupResult: &auth.ClientInfo{
				ClientID:   "test-client-1-abcdef0123567890",
				ClientName: "Test Client 1",
				Scope:      "diode:ingest",
				Owner:      "diode/user",
				CreatedAt:  "2021-01-01T00:00:00Z",
			},
			expectStatus: http.StatusNoContent,
		},
		{
			name:         "cannot delete client with invalid access token",
			accessToken:  invalidAccessToken,
			clientID:     "test-client-1-abcdef0123567890",
			parsedToken:  invalidToken,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "cannot delete client with read only access",
			accessToken:  validAccessToken,
			clientID:     "test-client-1-abcdef0123567890",
			parsedToken:  readOnlyToken,
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "cannot delete client that does not exist",
			accessToken:  validAccessToken,
			clientID:     "test-client-1-abcdef0123567890",
			parsedToken:  writeToken,
			lookupErr:    auth.NewAuthError("client not found", http.StatusNotFound),
			expectStatus: http.StatusNotFound,
		},
		{
			name:        "cannot delete a client with the wrong owner",
			accessToken: validAccessToken,
			clientID:    "test-client-1-abcdef0123567890",
			parsedToken: writeToken,
			lookupResult: &auth.ClientInfo{
				ClientID:   "test-client-1-abcdef0123567890",
				ClientName: "Test Client 1",
				Owner:      "diode/system",
				Scope:      "diode:read diode:write",
				CreatedAt:  "2021-01-01T00:00:00Z",
			},
			expectStatus: http.StatusNotFound,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupEnv()
	defer teardownEnv()

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := mockJWKSServer()
	defer mockJWKSServer.Close()

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defaultOwnership := &auth.DefaultTokenOwner{}
			accessToken := test.accessToken
			if accessToken == "" {
				accessToken = validAccessToken
			}
			mockTokenParser := &MockTokenParser{
				tokenMap: map[string]jwt.Token{
					accessToken: test.parsedToken,
				},
			}
			mockClientManager := &mocks.ClientManager{}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
			server, err := auth.NewServer(ctx, logger, mockTokenParser, mockClientManager, defaultOwnership)
			require.NoError(t, err)
			require.NotNil(t, server)

			testServer := httptest.NewServer(server.GetMux())
			defer testServer.Close()

			if test.lookupResult != nil {
				mockClientManager.EXPECT().RetrieveClientByID(mock.Anything, test.clientID).Return(*test.lookupResult, test.lookupErr)
			} else if test.lookupErr != nil {
				mockClientManager.EXPECT().RetrieveClientByID(mock.Anything, test.clientID).Return(auth.ClientInfo{}, test.lookupErr)
			}

			if test.expectStatus == http.StatusNoContent {
				mockClientManager.EXPECT().DeleteClientByID(mock.Anything, test.clientID).Return(nil)
			}

			req, _ := http.NewRequest("DELETE", testServer.URL+"/clients/"+test.clientID, nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			require.Equal(t, test.expectStatus, resp.StatusCode)
		})
	}
}

func TestGetClient(t *testing.T) {
	readOnlyToken := jwt.Token{
		Claims: jwt.MapClaims{
			"exp":       time.Now().Add(time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"client_id": "client123",
			"scope":     "diode:read",
		},
		Valid: true,
	}
	invalidToken := jwt.Token{
		Valid: false,
	}

	validAccessToken := "eyJhbGciOiJSUzI1NiIsImtpZCI6InRlc3Qta2V5IiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJzdWIiOiJ1c2VyMTIzIiwiYXVkIjoiYXBpIiwiZXhwIjoxNjUwMDAwMDAwLCJpYXQiOjE1MDAwMDAwMDAsImNsaWVudF9pZCI6ImNsaWVudDEyMyIsInNjb3BlIjoicmVhZCB3cml0ZSIsInVzZXJuYW1lIjoidGVzdHVzZXIifQ.WcPGXClpKD7Bc1C0CCDA1060E2GGlTfamrd8-W0ghBE"
	invalidAccessToken := "invalid.token.string"

	tests := []struct {
		name         string
		accessToken  string
		clientID     string
		parsedToken  jwt.Token
		lookupResult *auth.ClientInfo
		lookupErr    error
		expectStatus int
		expect       auth.ClientResponse
	}{
		{
			name:        "can get client",
			accessToken: validAccessToken,
			clientID:    "test-client-1-abcdef0123567890",
			parsedToken: readOnlyToken,
			lookupResult: &auth.ClientInfo{
				ClientID:   "test-client-1-abcdef0123567890",
				ClientName: "Test Client 1",
				Scope:      "diode:ingest",
				Owner:      "diode/user",
				CreatedAt:  "2021-01-01T00:00:00Z",
			},
			expect: auth.ClientResponse{
				ClientID:   "test-client-1-abcdef0123567890",
				ClientName: "Test Client 1",
				Scope:      "diode:ingest",
				CreatedAt:  "2021-01-01T00:00:00Z",
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "cannot get client with invalid access token",
			accessToken:  invalidAccessToken,
			clientID:     "test-client-1-abcdef0123567890",
			parsedToken:  invalidToken,
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "cannot get client that does not exist",
			accessToken:  validAccessToken,
			clientID:     "test-client-1-abcdef0123567890",
			parsedToken:  readOnlyToken,
			lookupErr:    auth.NewAuthError("client not found", http.StatusNotFound),
			expectStatus: http.StatusNotFound,
		},
		{
			name:        "cannot get a client with the wrong owner",
			accessToken: validAccessToken,
			clientID:    "test-client-1-abcdef0123567890",
			parsedToken: readOnlyToken,
			lookupResult: &auth.ClientInfo{
				ClientID:   "test-client-1-abcdef0123567890",
				ClientName: "Test Client 1",
				Owner:      "diode/system",
				Scope:      "diode:read diode:write",
				CreatedAt:  "2021-01-01T00:00:00Z",
			},
			expectStatus: http.StatusNotFound,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupEnv()
	defer teardownEnv()

	// Setup a test server to mock the OAuth2 server
	mockJWKSServer := mockJWKSServer()
	defer mockJWKSServer.Close()

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", mockJWKSServer.URL)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
	}()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defaultOwnership := &auth.DefaultTokenOwner{}
			accessToken := test.accessToken
			if accessToken == "" {
				accessToken = validAccessToken
			}
			mockTokenParser := &MockTokenParser{
				tokenMap: map[string]jwt.Token{
					accessToken: test.parsedToken,
				},
			}
			mockClientManager := &mocks.ClientManager{}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: false}))
			server, err := auth.NewServer(ctx, logger, mockTokenParser, mockClientManager, defaultOwnership)
			require.NoError(t, err)
			require.NotNil(t, server)

			testServer := httptest.NewServer(server.GetMux())
			defer testServer.Close()

			if test.lookupResult != nil {
				mockClientManager.EXPECT().RetrieveClientByID(mock.Anything, test.clientID).Return(*test.lookupResult, test.lookupErr)
			} else if test.lookupErr != nil {
				mockClientManager.EXPECT().RetrieveClientByID(mock.Anything, test.clientID).Return(auth.ClientInfo{}, test.lookupErr)
			}

			req, _ := http.NewRequest("GET", testServer.URL+"/clients/"+test.clientID, nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()
			require.Equal(t, test.expectStatus, resp.StatusCode)
		})
	}
}

func makeIntrospectRequest(serverURL, token string) (*http.Response, error) {
	req, _ := http.NewRequest(
		"POST",
		serverURL+"/introspect",
		strings.NewReader(""),
	)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	return client.Do(req)
}

func makeIntrospectRequestWithoutAuth(serverURL string) (*http.Response, error) {
	req, _ := http.NewRequest(
		"POST",
		serverURL+"/introspect",
		strings.NewReader(""),
	)

	client := &http.Client{}
	return client.Do(req)
}

func getFreePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return strconv.Itoa(0), err
	}

	addr := listener.Addr().(*net.TCPAddr)

	if err = listener.Close(); err != nil {
		return strconv.Itoa(0), err
	}
	return strconv.Itoa(addr.Port), nil
}

func mockJWKSServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			// Return a mock JWKS response
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"keys": [
					{
						"kty": "RSA",
						"kid": "test-key",
						"use": "sig",
						"alg": "RS256",
						"n": "n_value",
						"e": "AQAB"
					}
				]
			}`))
		}
	}))
}

func setupEnv() {
	httpPort, _ := getFreePort()
	_ = os.Setenv("HTTP_PORT", httpPort)
}

func teardownEnv() {
	_ = os.Unsetenv("HTTP_PORT")
}
