package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	hydra "github.com/ory/hydra-client-go/v25"
)

const (
	hydraAuthMethodClientSecretPost = "client_secret_post"
	hydraGrantTypeClientCredentials = "client_credentials"
	hydraResponseTypeToken          = "token"
	hydraClientFormatJSON           = "json"
)

// HydraClientManager is a ClientManager for a hydra server
type HydraClientManager struct {
	hydraAdmin *hydra.APIClient
	logger     *slog.Logger
}

// NewHydraClientManager creates a new HydraClientManager
func NewHydraClientManager(adminURL string, logger *slog.Logger) *HydraClientManager {
	hydraConfig := hydra.NewConfiguration()
	hydraConfig.Servers = []hydra.ServerConfiguration{
		{
			URL: adminURL,
		},
	}

	return &HydraClientManager{
		hydraAdmin: hydra.NewAPIClient(hydraConfig),
		logger:     logger,
	}
}

// CreateClient creates a new client
func (h *HydraClientManager) CreateClient(ctx context.Context, clientInfo ClientInfo) (ClientInfo, error) {
	authMethod := hydraAuthMethodClientSecretPost
	newClient := hydra.OAuth2Client{
		ClientId:                &clientInfo.ClientID,
		Scope:                   &clientInfo.Scope,
		GrantTypes:              []string{hydraGrantTypeClientCredentials},
		TokenEndpointAuthMethod: &authMethod,
		ResponseTypes:           []string{hydraResponseTypeToken},
	}
	if clientInfo.ClientSecret != "" {
		newClient.ClientSecret = &clientInfo.ClientSecret
	}
	if clientInfo.Owner != "" {
		newClient.Owner = &clientInfo.Owner
	}
	if clientInfo.ClientName != "" {
		newClient.ClientName = &clientInfo.ClientName
	}
	if clientInfo.Audience != nil {
		newClient.Audience = clientInfo.Audience
	}

	createdClient, response, err := h.hydraAdmin.OAuth2API.CreateOAuth2Client(ctx).OAuth2Client(newClient).Execute()
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				h.logger.Error("failed to close response body", "error", err)
			}
		}()
		if response.StatusCode == 409 {
			return ClientInfo{}, NewAuthError(fmt.Sprintf("failed to create client: client with id %s already exists", *newClient.ClientId), http.StatusConflict)
		}
		if response.StatusCode == 400 {
			return ClientInfo{}, NewAuthError("failed to create client: invalid request", http.StatusBadRequest)
		}
		if response.StatusCode != 201 {
			return ClientInfo{}, NewAuthError("failed to create client", response.StatusCode)
		}
	}
	// these can be confusing and related to internal client failures, so handled after http status codes
	if err != nil {
		return ClientInfo{}, fmt.Errorf("failed to create client: %w", err)
	}

	return clientInfoFromHydraClient(createdClient), nil
}

// DeleteClientByID deletes a client by id
func (h *HydraClientManager) DeleteClientByID(ctx context.Context, clientID string) error {
	response, err := h.hydraAdmin.OAuth2API.DeleteOAuth2Client(ctx, clientID).Execute()
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				h.logger.Error("failed to close response body", "error", err)
			}
		}()
		if response.StatusCode == 404 {
			return NewAuthError(fmt.Sprintf("client %s not found", clientID), http.StatusNotFound)
		}
		if response.StatusCode != 204 {
			return NewAuthError("failed to delete client from hydra", response.StatusCode)
		}
	}
	// these can be confusing and related to internal client failures, so handled after http status codes
	if err != nil {
		return fmt.Errorf("failed to delete client from hydra: %w", err)
	}

	return nil
}

// RetrieveClientByID retrieves a client by id
func (h *HydraClientManager) RetrieveClientByID(ctx context.Context, clientID string) (ClientInfo, error) {
	client, response, err := h.hydraAdmin.OAuth2API.GetOAuth2Client(ctx, clientID).Execute()
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				h.logger.Error("failed to close response body", "error", err)
			}
		}()
		if response.StatusCode == 404 {
			return ClientInfo{}, NewAuthError(fmt.Sprintf("client %s not found", clientID), http.StatusNotFound)
		}
		if response.StatusCode != 200 {
			return ClientInfo{}, NewAuthError("failed to retrieve client", response.StatusCode)
		}
	}
	// these tend to be confusing and related to internal client failures, so handled after http status codes
	if err != nil {
		return ClientInfo{}, fmt.Errorf("failed to retrieve client: %w", err)
	}

	return clientInfoFromHydraClient(client), nil
}

// RetrieveClients retrieves a list of clients
func (h *HydraClientManager) RetrieveClients(ctx context.Context, q RetrieveClientsRequest) (RetrieveClientsResponse, error) {
	var out RetrieveClientsResponse

	req := h.hydraAdmin.OAuth2API.ListOAuth2Clients(ctx)
	if q.Owner != "" {
		req = req.Owner(q.Owner)
	}
	if q.PageToken != "" {
		req = req.PageToken(q.PageToken)
	}
	if q.PageSize > 0 {
		req = req.PageSize(int64(q.PageSize))
	}

	clients, response, err := req.Execute()
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				h.logger.Error("failed to close response body", "error", err)
			}
		}()
		if response.StatusCode != 200 {
			return out, NewAuthError("failed to retrieve clients", response.StatusCode)
		}
	}
	if err != nil {
		return out, fmt.Errorf("failed to retrieve clients: %w", err)
	}

	out.Clients = make([]ClientInfo, 0, len(clients))
	for _, client := range clients {
		out.Clients = append(out.Clients, clientInfoFromHydraClient(&client))
	}

	out.FirstPageToken, out.NextPageToken = getHydraPagingTokens(response, h.logger)
	return out, nil
}

func clientInfoFromHydraClient(client *hydra.OAuth2Client) ClientInfo {
	var clientInfo ClientInfo

	if client == nil {
		return clientInfo
	}
	if client.ClientId != nil {
		clientInfo.ClientID = *client.ClientId
	}
	if client.Scope != nil {
		clientInfo.Scope = *client.Scope
	}
	if client.Owner != nil {
		clientInfo.Owner = *client.Owner
	}
	if client.ClientName != nil {
		clientInfo.ClientName = *client.ClientName
	}
	if client.CreatedAt != nil {
		clientInfo.CreatedAt = client.CreatedAt.Format(time.RFC3339)
	}
	if client.ClientSecret != nil {
		clientInfo.ClientSecret = *client.ClientSecret
	}
	if client.Audience != nil {
		clientInfo.Audience = client.Audience
	}
	return clientInfo
}

// getHydraPagingLinks returns the first and next page tokens from the response header
func getHydraPagingTokens(response *http.Response, logger *slog.Logger) (string, string) {
	first := ""
	next := ""
	linkHeaders := response.Header.Values("Link")
	for _, linkHeader := range linkHeaders {
		links := strings.Split(linkHeader, ",")
		for _, link := range links {
			params := strings.Split(link, ";")
			if len(params) < 2 {
				continue
			}
			link := params[0]
			params = params[1:]
			for _, param := range params {
				vs := strings.Split(param, "=")
				if len(vs) != 2 {
					continue
				}
				k, v := strings.TrimSpace(vs[0]), strings.TrimSpace(vs[1])
				if k == "rel" {
					switch v {
					case "first", "\"first\"":
						first = getHydraPageToken(link, logger)
					case "next", "\"next\"":
						next = getHydraPageToken(link, logger)
					}
				}
			}
		}
	}
	return first, next
}

func getHydraPageToken(link string, logger *slog.Logger) string {
	link = strings.TrimPrefix(link, "<")
	link = strings.TrimSuffix(link, ">")
	parsedURL, err := url.Parse(link)
	if err != nil {
		logger.Warn("failed to parse url in hydra paging link", "error", err, "link", link)
		return ""
	}
	queryParams := parsedURL.Query()
	for key, values := range queryParams {
		if key == "page_token" {
			return values[0]
		}
	}
	logger.Warn("failed to find page_token in hydra paging link", "link", link)
	return ""
}
