package authutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// ErrMetadataNotFoundMsg is an error for missing metadata
	ErrMetadataNotFoundMsg = "no request metadata found"

	// ErrUnauthenticatedMsg is an error for unauthenticated requests
	ErrUnauthenticatedMsg = "missing or invalid authorization header"

	// ErrMissingScopeMsg is an error for missing scopes
	ErrMissingScopeMsg = "missing required scope"
)

// Authorizer is an interface for making authorization checks
// on the contextual caller.
type Authorizer interface {
	RequireScopes(ctx context.Context, scopes []string) error
}

// TokenStringFromMetadata returns the bearer token from incoming grpc metadata
func TokenStringFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.InvalidArgument, ErrMetadataNotFoundMsg)
	}
	auth := md["authorization"]
	if len(auth) == 0 {
		return "", status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}

	return strings.TrimPrefix(auth[0], "Bearer "), nil
}

// GetScopeClaim returns the scope claim from a JWT token
func GetScopeClaim(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims type")
	}

	scopeClaim, ok := claims["scope"]
	if !ok {
		return "", fmt.Errorf("missing scope claim")
	}
	scopeClaimStr, ok := scopeClaim.(string)
	if !ok {
		return "", fmt.Errorf("scope claim is not a string")
	}
	return scopeClaimStr, nil
}

// RequireScopes checks if the scope claim contains all required scopes
func RequireScopes(scopeClaim string, requiredScopes []string) error {
	scopeList := strings.Split(scopeClaim, " ")
	scopeSet := make(map[string]bool)
	for _, scope := range scopeList {
		scopeSet[scope] = true
	}

	for _, scope := range requiredScopes {
		if !scopeSet[scope] {
			return fmt.Errorf("missing scope %s", scope)
		}
	}
	return nil
}
