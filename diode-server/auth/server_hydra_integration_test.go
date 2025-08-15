//go:build integration_test

package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/netboxlabs/diode/diode-server/auth"
)

func TestServerHydraIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	creq := testcontainers.ContainerRequest{
		Image:        "oryd/hydra:v2.3.0",
		ExposedPorts: []string{"4445/tcp", "4444/tcp"},
		WaitingFor:   wait.ForLog("Setting up http server on :4445"),
		Cmd:          []string{"serve", "all", "--dev"},
		Env: map[string]string{
			"DSN":                                      "memory",
			"STRATEGIES_ACCESS_TOKEN":                  "jwt",
			"STRATEGIES_REFRESH_TOKEN":                 "jwt",
			"STRATEGIES_JWT_SCOPE_CLAIM":               "both",
			"TTL_ACCESS_TOKEN":                         "1h",
			"OIDC_SUBJECT_IDENTIFIERS_SUPPORTED_TYPES": "public",
			"URLS_SELF_ISSUER":                         "http://127.0.0.1:4444",
			"SECRETS_SYSTEM":                           "some_very_secret_bytes",
			"LOG_LEAK_SENSITIVE_VALUES":                "true",
		},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{&logDumpConsumer{}},
		},
	}
	hydraC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: creq,
		Started:          true,
	})
	defer testcontainers.TerminateContainer(hydraC)
	require.NoError(t, err)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var manager auth.ClientManager
	hydraHost, err := hydraC.Host(ctx)
	require.NoError(t, err)
	adminPort, err := hydraC.MappedPort(ctx, "4445/tcp")
	require.NoError(t, err)
	publicPort, err := hydraC.MappedPort(ctx, "4444/tcp")
	require.NoError(t, err)
	adminEndpoint := fmt.Sprintf("http://%s:%s", hydraHost, adminPort.Port())
	publicEndpoint := fmt.Sprintf("http://%s:%s", hydraHost, publicPort.Port())
	manager = auth.NewHydraClientManager(adminEndpoint, logger)

	// no clients initially
	currentClients, err := manager.RetrieveClients(ctx, auth.RetrieveClientsRequest{})
	require.NoError(t, err)
	require.Equal(t, 0, len(currentClients.Clients))

	testClientID := "integration-test-client"
	testClientSecret := "integration-test-secret"
	testClientScope := "diode:read diode:write"

	// create a client to test with
	_, err = manager.CreateClient(ctx, auth.ClientInfo{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		Scope:        testClientScope,
	})
	require.NoError(t, err)

	setupEnv()
	defer teardownEnv()

	_ = os.Setenv("OAUTH2_PUBLIC_SERVER_URL", publicEndpoint)
	_ = os.Setenv("OAUTH2_ADMIN_SERVER_URL", adminEndpoint)
	defer func() {
		_ = os.Unsetenv("OAUTH2_PUBLIC_SERVER_URL")
		_ = os.Unsetenv("OAUTH2_ADMIN_SERVER_URL")
	}()

	defaultOwnership := &auth.DefaultTokenOwner{}
	tokenParser := &auth.JWTParser{}
	server, err := auth.NewServer(ctx, logger, tokenParser, manager, defaultOwnership)
	require.NoError(t, err)
	require.NotNil(t, server)

	testAudience := []string{"aud:1", "aud:2"}
	server.AddClientInfoDecorator(&audienceDecorator{aud: testAudience})

	testServer := httptest.NewServer(server.GetMux())
	defer testServer.Close()

	client := &authTestClient{
		endpoint: testServer.URL,
	}
	client.authenticate(t, testClientID, testClientSecret, testClientScope, []string{})

	// list clients (should be empty, only includes user created clients)
	result := client.listClients(t, "", 0)
	require.Equal(t, 0, len(result.Data))

	ingestClientName := "Ingest Client"
	ingestClientScope := "diode:ingest"

	ingestClientInfo := client.createClient(t, ingestClientName, ingestClientScope)

	require.NotNil(t, ingestClientInfo)
	require.Equal(t, ingestClientName, ingestClientInfo.ClientName)
	require.Equal(t, ingestClientScope, ingestClientInfo.Scope)
	require.NotEmpty(t, ingestClientInfo.ClientID)
	require.NotEmpty(t, ingestClientInfo.ClientSecret)

	// list clients, should include the ingest client, not the test client
	result = client.listClients(t, "", 0)
	require.Equal(t, 1, len(result.Data))
	require.Equal(t, ingestClientInfo.ClientID, result.Data[0].ClientID)
	require.Equal(t, "", result.Data[0].ClientSecret)
	require.Equal(t, ingestClientInfo.ClientName, result.Data[0].ClientName)
	require.Equal(t, ingestClientInfo.Scope, result.Data[0].Scope)

	// delete the ingest client
	err = manager.DeleteClientByID(ctx, ingestClientInfo.ClientID)
	require.NoError(t, err)

	// list clients, should be empty
	result = client.listClients(t, "", 0)
	require.Equal(t, 0, len(result.Data))

	// create 10 clients
	var clients []auth.ClientResponse
	for i := range 10 {
		clients = append(clients, client.createClient(t, fmt.Sprintf("test-client-%d", i), "diode:ingest"))
	}

	// list clients, should include the 10 test clients
	result = client.listClients(t, "", 100)
	require.Equal(t, 10, len(result.Data))
	require.Equal(t, result.NextPageToken, "")

	// get the first client
	firstClient := client.getClient(t, result.Data[0].ClientID)
	require.Equal(t, result.Data[0].ClientID, firstClient.ClientID)
	require.Equal(t, result.Data[0].ClientName, firstClient.ClientName)
	require.Equal(t, result.Data[0].Scope, firstClient.Scope)

	// delete the first client
	client.deleteClient(t, result.Data[0].ClientID)

	// list clients, should include the 9 remaining test clients
	result = client.listClients(t, "", 100)
	require.Equal(t, 9, len(result.Data))

	// page through the 9 remaining clients in pages of size 2
	var priorResult auth.ListClientsResponse
	pageSize := 2
	nextToken := ""
	seen := make(map[string]bool)
	pages := 0
	for range 6 { // should be 5 pages, stop after 6
		result = client.listClients(t, nextToken, pageSize)
		// previous page should be the same as the prior result
		if pages > 0 {
			prevPage := client.listClients(t, result.PrevPageToken, pageSize)
			require.Equal(t, priorResult.Data, prevPage.Data)
		}
		priorResult = result

		pages++
		for _, c := range result.Data {
			seen[c.ClientID] = true
		}
		nextToken = result.NextPageToken
		if nextToken == "" {
			break
		} else {
			require.Equal(t, pageSize, len(result.Data))
		}
	}

	// verify that we saw all 9 clients
	require.Equal(t, 9, len(seen))
	require.Equal(t, 5, pages)

	tokenClientInfo := client.createClient(t, "test-client-token-auth", ingestClientScope)

	// call the token endpoint with the credentials and verify that a token comes back ...
	resp := client.getToken(t, tokenClientInfo.ClientID, tokenClientInfo.ClientSecret, ingestClientScope, []string{"aud:2"})
	defer func() {
		_ = resp.Body.Close()
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResult struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResult)
	require.NoError(t, err)
	require.NotEmpty(t, tokenResult.AccessToken)
	accessToken := tokenResult.AccessToken
	require.NotEmpty(t, accessToken)

	token, _, err := jwt.NewParser().ParseUnverified(accessToken, jwt.MapClaims{})
	require.NoError(t, err)
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	scopeClaim, ok := claims["scope"]
	require.True(t, ok)
	require.Equal(t, ingestClientScope, scopeClaim)
	audienceClaim, ok := claims["aud"]
	require.True(t, ok)
	// either of these is technically valid according to the spec
	if audience, ok := audienceClaim.(string); ok {
		require.Equal(t, "aud:2", audience)
	}
	if audiences, ok := audienceClaim.([]interface{}); ok {
		require.Equal(t, 1, len(audiences))
		audience, ok := audiences[0].(string)
		require.True(t, ok)
		require.Equal(t, "aud:2", audience)
	}

	// try to use the credentials to create a token with a different scope ...
	resp = client.getToken(t, tokenClientInfo.ClientID, tokenClientInfo.ClientSecret, "netbox:read", []string{"aud:2"})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

type authTestClient struct {
	endpoint string
	token    string
}

func (c *authTestClient) getToken(t *testing.T, clientID string, clientSecret string, scope string, audience []string) *http.Response {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", scope)
	if len(audience) > 0 {
		data.Set("audience", strings.Join(audience, " "))
	}
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/token", strings.NewReader(data.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func (c *authTestClient) listClients(t *testing.T, pageToken string, pageSize int) auth.ListClientsResponse {
	u, err := url.Parse(c.endpoint + "/clients")
	require.NoError(t, err)
	q := url.Values{}
	if pageToken != "" {
		q.Set("page_token", pageToken)
	}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+c.token)
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var result auth.ListClientsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func (c *authTestClient) createClient(t *testing.T, clientName string, scope string) auth.ClientResponse {
	createReq := auth.CreateClientRequest{
		ClientName: clientName,
		Scope:      scope,
	}
	createReqBytes, err := json.Marshal(createReq)
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/clients", bytes.NewBuffer(createReqBytes))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdClient auth.ClientResponse
	err = json.NewDecoder(resp.Body).Decode(&createdClient)
	require.NoError(t, err)
	return createdClient
}

func (c *authTestClient) getClient(t *testing.T, clientID string) auth.ClientResponse {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+"/clients/"+clientID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+c.token)
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result auth.ClientResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func (c *authTestClient) deleteClient(t *testing.T, clientID string) {
	req, err := http.NewRequest(http.MethodDelete, c.endpoint+"/clients/"+clientID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+c.token)
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func (c *authTestClient) authenticate(t *testing.T, clientID string, clientSecret string, scope string, audience []string) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", scope)
	if len(audience) > 0 {
		data.Set("audience", strings.Join(audience, " "))
	}
	req, err := http.NewRequest(http.MethodPost, c.endpoint+"/token", strings.NewReader(data.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		AccessToken string `json:"access_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)

	c.token = result.AccessToken
}

type audienceDecorator struct {
	aud []string
}

func (d *audienceDecorator) VisitClientInfo(_ context.Context, clientInfo *auth.ClientInfo) error {
	clientInfo.Audience = d.aud
	return nil
}
