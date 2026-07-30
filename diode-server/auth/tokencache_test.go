package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testClientID     = "diode-ingest"
	testClientSecret = "correct-secret"
)

func testCacheConfig() TokenCacheConfig {
	return TokenCacheConfig{
		Enabled:     true,
		MaxEntries:  128,
		MaxTTL:      15 * time.Minute,
		NegativeTTL: 5 * time.Second,
	}
}

// upstream is a stand-in for the authorization server's token endpoint. It counts calls
// so tests can assert what the cache did or did not prevent.
type upstream struct {
	server *httptest.Server
	calls  atomic.Int64
}

// newUpstream starts a token endpoint that issues a token when the presented secret is
// testClientSecret, and rejects the request otherwise. The issued access token embeds
// the requested scope so tests can tell responses apart.
func newUpstream(t *testing.T) *upstream {
	t.Helper()

	u := &upstream{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		u.calls.Add(1)

		require.NoError(t, r.ParseForm())

		if r.Form.Get("client_secret") != testClientSecret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "token-for-" + r.Form.Get("scope"),
			"token_type":   "bearer",
			"expires_in":   3600,
			"scope":        r.Form.Get("scope"),
		})
	})

	u.server = httptest.NewServer(mux)
	t.Cleanup(u.server.Close)
	return u
}

func newTestServer(t *testing.T, upstreamURL string, cfg TokenCacheConfig) *Server {
	t.Helper()

	var cache *tokenCache
	if cfg.Enabled {
		var err error
		cache, err = newTokenCache(cfg)
		require.NoError(t, err)
	}

	return &Server{
		config:     Config{OAuth2: OAuth2Config{PublicServerURL: upstreamURL, TokenCache: cfg}},
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		httpClient: &http.Client{Timeout: upstreamTokenTimeout},
		tokenCache: cache,
	}
}

// tokenRequest issues a token request against the handler under test.
func tokenRequest(t *testing.T, s *Server, form url.Values) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	s.token(rec, req)
	return rec
}

func clientCredentialsForm(secret, scope string) url.Values {
	return url.Values{
		"grant_type":    {grantTypeClientCredentials},
		"client_id":     {testClientID},
		"client_secret": {secret},
		"scope":         {scope},
	}
}

func decodeToken(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

// gate is a TokenIssuanceGate whose verdict tests control.
type gate struct {
	allowed bool
	err     error
	calls   atomic.Int64
}

func (g *gate) Allow(_ context.Context, _ string) (bool, error) {
	g.calls.Add(1)
	return g.allowed, g.err
}

// The security properties below are the reason this cache is allowed to exist. If any of
// them regress the cache is an authentication bypass, not an optimisation.

func TestTokenCacheNeverServesTokenToWrongSecret(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, testCacheConfig())

	rec := tokenRequest(t, s, clientCredentialsForm(testClientSecret, "diode:ingest"))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(1), up.calls.Load())

	// Same client ID, wrong secret. This must not be served from the entry above.
	rec = tokenRequest(t, s, clientCredentialsForm("wrong-secret", "diode:ingest"))

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.NotContains(t, rec.Body.String(), "access_token")
	require.Equal(t, int64(2), up.calls.Load(), "wrong secret must reach the upstream")
}

func TestTokenCacheNeverServesBroaderScope(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, testCacheConfig())

	rec := tokenRequest(t, s, clientCredentialsForm(testClientSecret, "diode:read diode:write"))
	require.Equal(t, http.StatusOK, rec.Code)

	rec = tokenRequest(t, s, clientCredentialsForm(testClientSecret, "diode:read"))
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, "token-for-diode:read", decodeToken(t, rec)["access_token"])
	require.Equal(t, int64(2), up.calls.Load(), "a narrower scope must not be served from a broader entry")
}

func TestTokenCacheNeverCachesOtherGrants(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, testCacheConfig())

	form := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {"some-refresh-token"},
		"client_id":     {testClientID},
		"client_secret": {testClientSecret},
	}

	tokenRequest(t, s, form)
	tokenRequest(t, s, form)

	require.Equal(t, int64(2), up.calls.Load(), "only client_credentials responses may be cached")
}

func TestTokenCacheGateDenialFallsThroughToUpstream(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, testCacheConfig())

	form := clientCredentialsForm(testClientSecret, "diode:ingest")

	require.Equal(t, http.StatusOK, tokenRequest(t, s, form).Code)
	require.Equal(t, int64(1), up.calls.Load())

	// A hit while the gate denies must go upstream for the authoritative answer.
	denying := &gate{allowed: false}
	s.SetTokenIssuanceGate(denying)

	require.Equal(t, http.StatusOK, tokenRequest(t, s, form).Code)
	require.Equal(t, int64(1), denying.calls.Load())
	require.Equal(t, int64(2), up.calls.Load())
}

func TestTokenCacheGateErrorIsTreatedAsDenial(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, testCacheConfig())

	form := clientCredentialsForm(testClientSecret, "diode:ingest")
	require.Equal(t, http.StatusOK, tokenRequest(t, s, form).Code)

	s.SetTokenIssuanceGate(&gate{allowed: true, err: errors.New("tenant lookup unavailable")})

	require.Equal(t, http.StatusOK, tokenRequest(t, s, form).Code)
	require.Equal(t, int64(2), up.calls.Load(), "a gate error must not serve the cached entry")
}

func TestTokenCacheServesRepeatedRequests(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, testCacheConfig())

	form := clientCredentialsForm(testClientSecret, "diode:ingest")

	first := tokenRequest(t, s, form)
	second := tokenRequest(t, s, form)

	require.Equal(t, http.StatusOK, second.Code)
	require.Equal(t, decodeToken(t, first)["access_token"], decodeToken(t, second)["access_token"])
	require.Equal(t, int64(1), up.calls.Load())

	require.Equal(t, "application/json", second.Header().Get("Content-Type"))
	require.Equal(t, "no-store", second.Header().Get("Cache-Control"))
	require.Equal(t, "no-cache", second.Header().Get("Pragma"))
}

func TestTokenCacheDisabledProxiesEveryRequest(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, TokenCacheConfig{})

	form := clientCredentialsForm(testClientSecret, "diode:ingest")
	tokenRequest(t, s, form)
	tokenRequest(t, s, form)

	require.Equal(t, int64(2), up.calls.Load())
}

func TestTokenCacheCollapsesConcurrentMisses(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{}, 1)

	var calls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("POST /oauth2/token", func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		select {
		case started <- struct{}{}:
		default:
		}
		<-release

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "token",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	s := newTestServer(t, server.URL, testCacheConfig())
	form := clientCredentialsForm(testClientSecret, "diode:ingest")

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			require.Equal(t, http.StatusOK, tokenRequest(t, s, form).Code)
		})
	}

	<-started
	close(release)
	wg.Wait()

	require.Equal(t, int64(1), calls.Load(), "concurrent misses on one key must collapse")
}

func TestTokenCacheNegativeEntryIsScopedToPresentedSecret(t *testing.T) {
	up := newUpstream(t)
	s := newTestServer(t, up.server.URL, testCacheConfig())

	// Two rejections of the same wrong secret cost the upstream one verification.
	require.Equal(t, http.StatusUnauthorized, tokenRequest(t, s, clientCredentialsForm("wrong", "diode:ingest")).Code)
	require.Equal(t, http.StatusUnauthorized, tokenRequest(t, s, clientCredentialsForm("wrong", "diode:ingest")).Code)
	require.Equal(t, int64(1), up.calls.Load())

	// The correct secret hashes to a different key and cannot be locked out by it.
	rec := tokenRequest(t, s, clientCredentialsForm(testClientSecret, "diode:ingest"))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(2), up.calls.Load())
}

func TestCacheableTokenRequest(t *testing.T) {
	basic := func(id, secret string) http.Header {
		h := http.Header{}
		h.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(id)+":"+url.QueryEscape(secret))))
		return h
	}

	tests := []struct {
		name     string
		body     string
		header   http.Header
		wantOK   bool
		wantCred tokenCredentials
	}{
		{
			name:   "client secret post",
			body:   "grant_type=client_credentials&client_id=a&client_secret=b&scope=y+x",
			header: http.Header{},
			wantOK: true,
			wantCred: tokenCredentials{
				grantType: grantTypeClientCredentials, clientID: "a", clientSecret: "b",
				scopes: []string{"x", "y"},
			},
		},
		{
			// Every client this service registers uses client_secret_post and the
			// upstream rejects anything else, so basic credentials cannot yield a
			// cacheable response. Bypassing is also what stops us keying on the form
			// while the upstream authenticates the header.
			name:   "credentials in an authorization header bypass the cache",
			body:   "grant_type=client_credentials",
			header: basic("a", "b"),
			wantOK: false,
		},
		{
			name:   "form credentials alongside an authorization header are ambiguous",
			body:   "grant_type=client_credentials&client_id=a&client_secret=b",
			header: basic("a", "b"),
			wantOK: false,
		},
		{
			name:   "any authorization scheme bypasses the cache",
			body:   "grant_type=client_credentials&client_id=a&client_secret=b",
			header: http.Header{"Authorization": {"Bearer some-token"}},
			wantOK: false,
		},
		{
			name:   "other grant type",
			body:   "grant_type=refresh_token&client_id=a&client_secret=b",
			header: http.Header{},
			wantOK: false,
		},
		{
			name:   "missing secret",
			body:   "grant_type=client_credentials&client_id=a",
			header: http.Header{},
			wantOK: false,
		},
		{
			name:   "no credentials at all",
			body:   "grant_type=client_credentials",
			header: http.Header{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred, ok := cacheableTokenRequest([]byte(tt.body), tt.header)
			require.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				require.Equal(t, tt.wantCred, cred)
			}
		})
	}
}

func TestTokenCacheKeyDerivation(t *testing.T) {
	c, err := newTokenCache(testCacheConfig())
	require.NoError(t, err)

	base := tokenCredentials{
		grantType: grantTypeClientCredentials, clientID: "a", clientSecret: "b",
		scopes: []string{"read", "write"},
	}

	t.Run("scope order does not matter", func(t *testing.T) {
		first, ok := cacheableTokenRequest([]byte("grant_type=client_credentials&client_id=a&client_secret=b&scope=read+write"), http.Header{})
		require.True(t, ok)
		second, ok := cacheableTokenRequest([]byte("grant_type=client_credentials&client_id=a&client_secret=b&scope=write+read"), http.Header{})
		require.True(t, ok)

		require.Equal(t, c.key(first), c.key(second))
	})

	t.Run("secret is part of the key", func(t *testing.T) {
		other := base
		other.clientSecret = "b2"
		require.NotEqual(t, c.key(base), c.key(other))
	})

	t.Run("scope is part of the key", func(t *testing.T) {
		other := base
		other.scopes = []string{"read"}
		require.NotEqual(t, c.key(base), c.key(other))
	})

	t.Run("audience is part of the key", func(t *testing.T) {
		other := base
		other.audiences = []string{"https://example.test"}
		require.NotEqual(t, c.key(base), c.key(other))
	})

	t.Run("field boundaries are unambiguous", func(t *testing.T) {
		// Without length prefixing these two tuples would concatenate identically, and a
		// caller could construct a client ID that borrowed another caller's secret.
		left := tokenCredentials{grantType: grantTypeClientCredentials, clientID: "ab", clientSecret: "c"}
		right := tokenCredentials{grantType: grantTypeClientCredentials, clientID: "a", clientSecret: "bc"}
		require.NotEqual(t, c.key(left), c.key(right))
	})
}

func TestTokenCacheWindow(t *testing.T) {
	now := time.Now()
	cred := tokenCredentials{grantType: grantTypeClientCredentials, clientID: "a", clientSecret: "b"}

	t.Run("bounded by the configured maximum", func(t *testing.T) {
		cfg := testCacheConfig()
		cfg.MaxTTL = time.Minute
		c, err := newTokenCache(cfg)
		require.NoError(t, err)

		require.True(t, c.putToken("k", cred, map[string]any{"expires_in": float64(3600)}, now))

		_, ok := c.get("k", now.Add(59*time.Second))
		require.True(t, ok)
		_, ok = c.get("k", now.Add(61*time.Second))
		require.False(t, ok, "entry must not outlive the configured maximum")
	})

	t.Run("bounded by a fraction of the token lifetime", func(t *testing.T) {
		c, err := newTokenCache(testCacheConfig())
		require.NoError(t, err)

		// 100s token: reusable for 80s, well inside the 15m configured maximum.
		require.True(t, c.putToken("k", cred, map[string]any{"expires_in": float64(100)}, now))

		_, ok := c.get("k", now.Add(79*time.Second))
		require.True(t, ok)
		_, ok = c.get("k", now.Add(81*time.Second))
		require.False(t, ok)
	})

	t.Run("responses without expires_in are not cached", func(t *testing.T) {
		c, err := newTokenCache(testCacheConfig())
		require.NoError(t, err)

		require.False(t, c.putToken("k", cred, map[string]any{"access_token": "t"}, now))
		_, ok := c.get("k", now)
		require.False(t, ok)
	})
}

func TestTokenCacheRecomputesExpiresIn(t *testing.T) {
	now := time.Now()
	c, err := newTokenCache(testCacheConfig())
	require.NoError(t, err)

	cred := tokenCredentials{grantType: grantTypeClientCredentials, clientID: "a", clientSecret: "b"}
	require.True(t, c.putToken("k", cred, map[string]any{
		"access_token": "t",
		"expires_in":   float64(3600),
	}, now))

	entry, ok := c.get("k", now.Add(10*time.Minute))
	require.True(t, ok)

	body, ok := entry.responseBody(now.Add(10 * time.Minute))
	require.True(t, ok)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(body, &decoded))

	require.EqualValues(t, 3000, decoded["expires_in"], "expires_in must reflect remaining life")
	require.Equal(t, "t", decoded["access_token"], "unknown upstream fields must survive")
}

func TestTokenCacheEvictsOldestBeyondMaxEntries(t *testing.T) {
	cfg := testCacheConfig()
	cfg.MaxEntries = 2
	c, err := newTokenCache(cfg)
	require.NoError(t, err)

	now := time.Now()
	cred := tokenCredentials{grantType: grantTypeClientCredentials, clientID: "a", clientSecret: "b"}
	body := map[string]any{"expires_in": float64(3600)}

	for _, key := range []string{"a", "b", "c"} {
		require.True(t, c.putToken(key, cred, body, now))
	}

	_, ok := c.get("a", now)
	require.False(t, ok, "cycling keys must not grow the cache without bound")
	_, ok = c.get("c", now)
	require.True(t, ok)
}

func TestNegativelyCacheable(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{"invalid client", http.StatusUnauthorized, `{"error":"invalid_client"}`, true},
		{"invalid client on 400", http.StatusBadRequest, `{"error":"invalid_client"}`, true},
		{"other oauth2 error", http.StatusBadRequest, `{"error":"invalid_scope"}`, false},
		{"server error", http.StatusInternalServerError, `{"error":"invalid_client"}`, false},
		{"throttled", http.StatusTooManyRequests, `{"error":"invalid_client"}`, false},
		{"unparseable", http.StatusUnauthorized, `not json`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, negativelyCacheable(tt.status, []byte(tt.body)))
		})
	}
}
