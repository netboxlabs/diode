package authutil

import (
	"context"
	"log/slog"
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

// TokenStringFromMetadata returns the bearer token from incoming metadata
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

// Authorizer is an interface for checking if a token has the required scopes
type Authorizer interface {
	RequireScopes(token string, scopes []string) error
	RequireScopesContext(ctx context.Context, scopes []string) error
}

// UnverifiedJWTAuthorizer is an authorizer that does not verify the JWT token's
// signature. It is only used to check if the token has the required scopes
// and is otherwise well formed.
type UnverifiedJWTAuthorizer struct {
	logger *slog.Logger
}

// NewUnverifiedJWTAuthorizer creates a new UnverifiedJWTAuthorizer
func NewUnverifiedJWTAuthorizer(logger *slog.Logger) *UnverifiedJWTAuthorizer {
	return &UnverifiedJWTAuthorizer{logger: logger}
}

// RequireScopesContext checks if the token in context has the required scopes
func (j *UnverifiedJWTAuthorizer) RequireScopesContext(ctx context.Context, scopes []string) error {
	tokenString, err := TokenStringFromMetadata(ctx)
	if err != nil {
		return err
	}
	return j.RequireScopes(tokenString, scopes)
}

// RequireScopes checks if the token has the required scopes
func (j *UnverifiedJWTAuthorizer) RequireScopes(tokenString string, requiredScopes []string) error {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		j.logger.Debug("failed to parse token", "error", err)
		return status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		j.logger.Debug("invalid claims type", "claims", claims)
		return status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}

	scopeClaim, ok := claims["scope"]
	if !ok {
		j.logger.Debug("missing scope claim")
		return status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}
	scopeClaimStr, ok := scopeClaim.(string)
	if !ok {
		j.logger.Debug("scope claim is not a string", "scope", scopeClaim)
		return status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}
	scopeList := strings.Split(scopeClaimStr, " ")
	scopeSet := make(map[string]bool)
	for _, scope := range scopeList {
		scopeSet[scope] = true
	}

	for _, scope := range requiredScopes {
		if !scopeSet[scope] {
			j.logger.Debug("missing scope", "scope", scope)
			return status.Errorf(codes.Unauthenticated, ErrMissingScopeMsg)
		}
	}
	return nil
}
