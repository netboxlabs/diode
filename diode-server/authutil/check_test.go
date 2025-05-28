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

func TestRequireScopesToken(t *testing.T) {
	authorizer := authutil.NewContextAuthorizer(slog.Default())
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
			interceptor := authutil.NewUnverifiedJWTInterceptor(slog.Default())
			_, err := interceptor(tt.ctx, nil, nil, func(ctx context.Context, _ any) (any, error) {
				return nil, authorizer.RequireScopes(ctx, tt.requireScopes)
			})

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
