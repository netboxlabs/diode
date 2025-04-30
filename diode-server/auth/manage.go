package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

// ClientInfo is a struct that contains information about a client.
type ClientInfo struct {
	ClientID     string `json:"client_id"`
	Scope        string `json:"scope"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// RetrieveClientsRequest is a struct that contains information about a request to retrieve clients.
type RetrieveClientsRequest struct {
	PageToken string
}

// RetrieveClientsResponse reponse struct for listing clients
type RetrieveClientsResponse struct {
	Clients       []ClientInfo
	NextPageToken string
}

// ClientManager is an interface for managing oauth2 clients.
type ClientManager interface {
	// CreateClient creates a new oauth2 client
	CreateClient(ctx context.Context, clientInfo ClientInfo) (ClientInfo, error)
	// RetrieveClientByID retrieves information about an oauth2 client by id
	RetrieveClientByID(ctx context.Context, clientID string) (ClientInfo, error)
	// RetrieveClients retrieves a list of oauth2 clients
	RetrieveClients(ctx context.Context, q RetrieveClientsRequest) (RetrieveClientsResponse, error)
	// DeleteClientByID deletes an oauth2 client by id
	DeleteClientByID(ctx context.Context, clientID string) error
}

// GenerateClientSecret generates a random 32 byte client secret.
func GenerateClientSecret() (string, error) {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(secret), nil
}
