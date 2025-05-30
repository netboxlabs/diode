package authutil

import (
	"context"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ContextAuthorizer is an Authorizer that checks
// values that have been placed in the context by an
// upstream interceptor.  It does not parse or validate
// the JWT token.
type ContextAuthorizer struct {
	logger *slog.Logger
}

// NewContextAuthorizer creates a new ContextAuthorizer
func NewContextAuthorizer(logger *slog.Logger) *ContextAuthorizer {
	return &ContextAuthorizer{logger: logger}
}

// RequireScopes checks if the scope claim in the context
// is present and has the required scopes.
func (c *ContextAuthorizer) RequireScopes(ctx context.Context, scopes []string) error {
	scope, ok := ctx.Value(ContextKeyScope).(string)
	if !ok {
		c.logger.Debug("no scope found in context")
		return status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}
	if err := RequireScopes(scope, scopes); err != nil {
		c.logger.Debug("missing scope", "error", err)
		return status.Errorf(codes.Unauthenticated, ErrMissingScopeMsg)
	}
	return nil
}

// NewUnverifiedJWTInterceptor returns a grpc.UnaryServerInterceptor
// that creates an authenticated context for the request.  It does
// not verify the JWT token, it just parses it and puts the token
// scope in the context.  It is expected that the token has
// already been verified by an upstream service.
func NewUnverifiedJWTInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		tokenString, err := TokenStringFromMetadata(ctx)
		if err != nil {
			return nil, err
		}
		token, err := ParseTokenUnverified(tokenString, logger)
		if err != nil {
			return nil, err
		}
		ctx, err = ContextWithTokenScope(ctx, token, logger)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// ParseTokenUnverified is a helper function that parses a JWT token
// without verifying it.  It is expected that the token has already
// been verified by an upstream service.
func ParseTokenUnverified(tokenString string, logger *slog.Logger) (*jwt.Token, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		logger.Debug("failed to parse token", "error", err)
		return nil, status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}

	return token, nil
}

// ContextWithTokenScope is a ContextClaimsFunc that places the scope claim from a
// parsed JWT token in the context.  It does not verify or validate the token,
// but it does return an error if the scope claim is not present or invalid.
func ContextWithTokenScope(ctx context.Context, token *jwt.Token, logger *slog.Logger) (context.Context, error) {
	if token == nil {
		return nil, status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}
	scope, err := GetScopeClaim(token)
	if err != nil {
		logger.Debug("failed to get scope claim", "error", err)
		return nil, status.Errorf(codes.Unauthenticated, ErrUnauthenticatedMsg)
	}

	ctx = context.WithValue(ctx, ContextKeyScope, scope)

	return ctx, nil
}
