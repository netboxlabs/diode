package authutil_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/netboxlabs/diode/diode-server/authutil"
)

func TestRequireScopes(t *testing.T) {
	authorizer := authutil.NewUnverifiedJWTAuthorizer(slog.Default())

	type test struct {
		name          string
		token         string
		requireScopes []string
		wantErr       error
	}

	validToken := validWithScopes(t, []string{authutil.ScopeDiodeRead, authutil.ScopeDiodeWrite})

	tests := []test{
		{
			name:          "valid with valid scopes",
			token:         validToken,
			requireScopes: []string{authutil.ScopeDiodeRead, authutil.ScopeDiodeWrite},
			wantErr:       nil,
		},
		{
			name:          "valid with valid read scope",
			token:         validToken,
			requireScopes: []string{authutil.ScopeDiodeRead},
			wantErr:       nil,
		},
		{
			name:          "valid with valid write scopes",
			token:         validToken,
			requireScopes: []string{authutil.ScopeDiodeWrite},
			wantErr:       nil,
		},
		{
			name:          "valid with missing scope",
			token:         validToken,
			requireScopes: []string{"diode:sleep"},
			wantErr:       status.Errorf(codes.Unauthenticated, authutil.ErrMissingScopeMsg),
		},
		{
			name:          "valid with some missing scopes",
			token:         validToken,
			requireScopes: []string{authutil.ScopeDiodeRead, authutil.ScopeDiodeWrite, "diode:sleep"},
			wantErr:       status.Errorf(codes.Unauthenticated, authutil.ErrMissingScopeMsg),
		},
		{
			name:          "invalid token",
			token:         "some-invalid-token",
			requireScopes: []string{authutil.ScopeDiodeRead},
			wantErr:       status.Errorf(codes.Unauthenticated, authutil.ErrUnauthenticatedMsg),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authorizer.RequireScopes(tt.token, tt.requireScopes)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRequireScopesContext(t *testing.T) {
	authorizer := authutil.NewUnverifiedJWTAuthorizer(slog.Default())
	validToken := validWithScopes(t, []string{authutil.ScopeDiodeRead, authutil.ScopeDiodeWrite})

	type test struct {
		name          string
		ctx           context.Context
		requireScopes []string
		wantErr       error
	}

	tests := []test{
		{
			name:          "valid with valid scopes",
			ctx:           contextWithToken(validToken),
			requireScopes: []string{authutil.ScopeDiodeRead, authutil.ScopeDiodeWrite},
			wantErr:       nil,
		},
		{
			name:          "missing scope",
			ctx:           contextWithToken(validToken),
			requireScopes: []string{"diode:sleep"},
			wantErr:       status.Errorf(codes.Unauthenticated, authutil.ErrMissingScopeMsg),
		},
		{
			name:          "no metadata",
			ctx:           context.Background(),
			requireScopes: []string{authutil.ScopeDiodeRead},
			wantErr:       status.Errorf(codes.InvalidArgument, authutil.ErrMetadataNotFoundMsg),
		},
		{
			name:          "invalid token",
			ctx:           contextWithToken("some-invalid-token"),
			requireScopes: []string{authutil.ScopeDiodeRead},
			wantErr:       status.Errorf(codes.Unauthenticated, authutil.ErrUnauthenticatedMsg),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authorizer.RequireScopesContext(tt.ctx, tt.requireScopes)
			if tt.wantErr != nil {
				require.Error(t, err)
				require.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func validWithScopes(t *testing.T, scopes []string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"scope": strings.Join(scopes, " "),
	})
	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)
	return tokenString
}

func contextWithToken(tokenString string) context.Context {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		"authorization": "Bearer " + tokenString,
	}))
	return ctx
}
