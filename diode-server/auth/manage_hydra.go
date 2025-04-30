package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	hydra "github.com/ory/hydra-client-go/v2"
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
	newClient := hydra.OAuth2Client{
		ClientId:   &clientInfo.ClientID,
		Scope:      &clientInfo.Scope,
		GrantTypes: []string{"client_credentials"},
	}
	if clientInfo.ClientSecret != "" {
		newClient.ClientSecret = &clientInfo.ClientSecret
	}

	createdClient, response, err := h.hydraAdmin.OAuth2API.CreateOAuth2Client(ctx).OAuth2Client(newClient).Execute()
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				h.logger.Error("failed to close response body", "error", err)
			}
		}()
	}
	if response.StatusCode == 409 {
		return ClientInfo{}, fmt.Errorf("failed to create client: client with id %s already exists", *newClient.ClientId)
	}
	if response.StatusCode == 400 {
		return ClientInfo{}, fmt.Errorf("failed to create client: invalid request")
	}
	if response.StatusCode != 201 {
		return ClientInfo{}, fmt.Errorf("failed to create client: status=%s", response.Status)
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
	}

	if response.StatusCode == 404 {
		return fmt.Errorf("client %s not found", clientID)
	}
	if response.StatusCode != 204 {
		return fmt.Errorf("failed to delete client from hydra: status=%s", response.Status)
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
	}

	if response.StatusCode == 404 {
		return ClientInfo{}, fmt.Errorf("client %s not found", clientID)
	}
	if response.StatusCode != 200 {
		return ClientInfo{}, fmt.Errorf("failed to retrieve client: status=%s", response.Status)
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

	clients, response, err := h.hydraAdmin.OAuth2API.ListOAuth2Clients(ctx).PageToken(q.PageToken).Execute()
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				h.logger.Error("failed to close response body", "error", err)
			}
		}()
	}
	if response.StatusCode != 200 {
		return out, fmt.Errorf("failed to retrieve clients: status=%s", response.Status)
	}
	// these can be confusing and related to internal client failures, so handled after http status codes
	if err != nil {
		return out, fmt.Errorf("failed to retrieve clients: %w", err)
	}

	out.Clients = make([]ClientInfo, 0, len(clients))
	for _, client := range clients {
		out.Clients = append(out.Clients, clientInfoFromHydraClient(&client))
	}

	out.NextPageToken = getHydraNextPageToken(response, h.logger)
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

	return clientInfo
}

func getHydraNextPageToken(response *http.Response, logger *slog.Logger) string {
	for _, linkHeader := range response.Header.Values("Link") {
		params := strings.Split(linkHeader, ";")
		link := params[0]
		params = params[1:]
		// search for rel="next"
		for _, param := range params {
			vs := strings.Split(param, "=")
			if len(vs) != 2 {
				continue
			}
			k, v := strings.TrimSpace(vs[0]), strings.TrimSpace(vs[1])
			if k == "rel" && (v == "next" || v == "\"next\"") {
				parsedURL, err := url.Parse(link)
				if err != nil {
					logger.Warn("failed to parse url in rel=next link", "error", err, "link", linkHeader)
					return ""
				}
				queryParams := parsedURL.Query()
				for key, values := range queryParams {
					if key == "page_token" {
						return values[0]
					}
				}
				logger.Warn("failed to find next page token in rel=next url", "link", linkHeader)
				return ""
			}
		}
	}
	return ""
}
