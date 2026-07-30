package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"
)

const (
	// grantTypeClientCredentials is the only grant whose responses may be cached. Other
	// grants are proxied untouched.
	grantTypeClientCredentials = "client_credentials"

	// errInvalidClient is the only upstream rejection that is negatively cached.
	errInvalidClient = "invalid_client"

	// tokenReuseNumerator and tokenReuseDenominator bound reuse to a fraction of a
	// token's own lifetime, so a cached token is never served close to its expiry.
	tokenReuseNumerator   = 8
	tokenReuseDenominator = 10
)

// tokenCacheOutcome describes how a token request interacted with the cache. Outcomes
// are mutually exclusive: exactly one is recorded per request.
type tokenCacheOutcome string

const (
	// tokenCacheHit is a token served from cache.
	tokenCacheHit tokenCacheOutcome = "hit"
	// tokenCacheMiss is a cacheable request that had to go upstream.
	tokenCacheMiss tokenCacheOutcome = "miss"
	// tokenCacheBypass is a request the cache is not allowed to serve, such as a grant
	// other than client credentials or a request whose credentials could not be located.
	tokenCacheBypass tokenCacheOutcome = "bypass"
	// tokenCacheNegativeHit is a cached upstream rejection served without going upstream.
	tokenCacheNegativeHit tokenCacheOutcome = "negative_hit"
	// tokenCacheGateDenied is a cache hit refused because the issuance gate denied it.
	tokenCacheGateDenied tokenCacheOutcome = "gate_denied"
)

// TokenIssuanceGate decides whether a client may still be issued a token.
//
// On a cache miss the upstream authorization server performs its own checks, including
// any token hook. A cache hit skips the upstream entirely and therefore skips those
// checks, so the gate is consulted on every hit to keep them in force. Implementations
// must be cheap enough to sit on the hot path and are expected to fail closed.
type TokenIssuanceGate interface {
	Allow(ctx context.Context, clientID string) (bool, error)
}

// tokenCredentials are the parts of a token request that identify the caller and the
// token being asked for. Every field participates in the cache key.
type tokenCredentials struct {
	grantType    string
	clientID     string
	clientSecret string
	scopes       []string
	audiences    []string
}

// tokenCacheEntry is a cached upstream token response. Exactly one of the positive or
// negative field sets is populated, discriminated by negative.
type tokenCacheEntry struct {
	negative bool

	// Positive entries.
	clientID       string
	body           map[string]any
	tokenExpiresAt time.Time

	// Negative entries.
	statusCode  int
	contentType string
	rawBody     []byte

	servableUntil time.Time
}

// tokenCache caches client credentials token responses keyed by the presented
// credentials.
//
// The presented secret is part of every key. A caller holding a valid client ID and the
// wrong secret therefore derives a different key and can never be served a token minted
// for the correct secret. Nothing in this package may key an entry without it.
type tokenCache struct {
	entries     *lru.Cache[string, tokenCacheEntry]
	group       singleflight.Group
	keyMaterial []byte
	maxTTL      time.Duration
	negativeTTL time.Duration
}

// newTokenCache creates a token cache sized and bounded by cfg.
func newTokenCache(cfg TokenCacheConfig) (*tokenCache, error) {
	entries, err := lru.New[string, tokenCacheEntry](cfg.MaxEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to create token cache: %w", err)
	}

	// Per-process key material, so a cache key is never a credential and never
	// meaningful outside the process that derived it.
	keyMaterial := make([]byte, sha256.Size)
	if _, err := rand.Read(keyMaterial); err != nil {
		return nil, fmt.Errorf("failed to generate token cache key material: %w", err)
	}

	return &tokenCache{
		entries:     entries,
		keyMaterial: keyMaterial,
		maxTTL:      cfg.MaxTTL,
		negativeTTL: cfg.NegativeTTL,
	}, nil
}

// key derives the cache key for cred.
//
// Fields are length-prefixed so that no two distinct credential tuples can produce the
// same input, and the digest is an HMAC so that the key is not itself a usable
// credential even if it leaks through a log or a metric label.
func (c *tokenCache) key(cred tokenCredentials) string {
	mac := hmac.New(sha256.New, c.keyMaterial)
	for _, field := range []string{
		cred.grantType,
		cred.clientID,
		cred.clientSecret,
		strings.Join(cred.scopes, " "),
		strings.Join(cred.audiences, " "),
	} {
		_, _ = fmt.Fprintf(mac, "%d:%s", len(field), field)
	}
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// get returns a live entry for key. Entries past their serving window are removed and
// reported as a miss.
func (c *tokenCache) get(key string, now time.Time) (tokenCacheEntry, bool) {
	entry, ok := c.entries.Get(key)
	if !ok {
		return tokenCacheEntry{}, false
	}
	if !now.Before(entry.servableUntil) {
		c.entries.Remove(key)
		return tokenCacheEntry{}, false
	}
	return entry, true
}

// putToken caches a successful token response and reports whether it was stored.
//
// Responses without a usable expires_in are not cached: remaining lifetime could not be
// recomputed on a later hit, and serving a stale expires_in would just move the failure
// downstream into the caller.
func (c *tokenCache) putToken(key string, cred tokenCredentials, body map[string]any, now time.Time) bool {
	lifetime, ok := tokenLifetime(body)
	if !ok || lifetime <= 0 {
		return false
	}

	reusable := min(lifetime*tokenReuseNumerator/tokenReuseDenominator, c.maxTTL)
	if reusable <= 0 {
		return false
	}

	c.entries.Add(key, tokenCacheEntry{
		clientID:       cred.clientID,
		body:           body,
		tokenExpiresAt: now.Add(lifetime),
		servableUntil:  now.Add(reusable),
	})
	return true
}

// putNegative caches an upstream client rejection.
//
// This is safe only because the presented secret is part of the key: a caller that later
// presents the correct secret derives a different key and cannot be locked out by a
// cached rejection.
func (c *tokenCache) putNegative(key string, statusCode int, contentType string, body []byte, now time.Time) {
	if c.negativeTTL <= 0 {
		return
	}
	c.entries.Add(key, tokenCacheEntry{
		negative:      true,
		statusCode:    statusCode,
		contentType:   contentType,
		rawBody:       body,
		servableUntil: now.Add(c.negativeTTL),
	})
}

// remove drops the entry for key, used when the issuance gate refuses a hit.
func (c *tokenCache) remove(key string) {
	c.entries.Remove(key)
}

// responseBody re-marshals a cached token with expires_in recomputed from the remaining
// lifetime, so a caller never receives the originally issued value on a hit. It reports
// false once nothing useful is left of the token.
func (e tokenCacheEntry) responseBody(now time.Time) ([]byte, bool) {
	remaining := int64(e.tokenExpiresAt.Sub(now).Seconds())
	if remaining <= 0 {
		return nil, false
	}

	// Copy before mutating: the stored map is shared with concurrent readers.
	body := make(map[string]any, len(e.body))
	maps.Copy(body, e.body)
	body["expires_in"] = remaining

	out, err := json.Marshal(body)
	if err != nil {
		return nil, false
	}
	return out, true
}

// cacheableTokenRequest reports whether a token request may be served from cache, and
// returns the credentials to key it by.
//
// It is deliberately conservative: anything it cannot positively identify is proxied
// rather than cached.
func cacheableTokenRequest(body []byte, header http.Header) (tokenCredentials, bool) {
	form, err := url.ParseQuery(string(body))
	if err != nil {
		return tokenCredentials{}, false
	}

	// The endpoint proxies whatever the caller sends. Caching a refresh_token or
	// authorization_code response would be incorrect as well as unsafe.
	if form.Get("grant_type") != grantTypeClientCredentials {
		return tokenCredentials{}, false
	}

	basicID, basicSecret, hasBasic := basicAuthCredentials(header)
	formID, formSecret := form.Get("client_id"), form.Get("client_secret")
	hasForm := formID != "" || formSecret != ""

	var clientID, clientSecret string
	switch {
	case hasBasic && hasForm:
		// RFC 6749 section 2.3 forbids more than one authentication method. We cannot
		// know which one the upstream will validate, so we must not key on either.
		return tokenCredentials{}, false
	case hasBasic:
		clientID, clientSecret = basicID, basicSecret
	case hasForm:
		clientID, clientSecret = formID, formSecret
	default:
		return tokenCredentials{}, false
	}

	if clientID == "" || clientSecret == "" {
		return tokenCredentials{}, false
	}

	return tokenCredentials{
		grantType:    grantTypeClientCredentials,
		clientID:     clientID,
		clientSecret: clientSecret,
		scopes:       sortedFields(form["scope"]),
		audiences:    sortedFields(form["audience"]),
	}, true
}

// basicAuthCredentials extracts client_secret_basic credentials.
//
// RFC 6749 section 2.3.1 requires both halves to be form-urlencoded before they are
// base64 encoded, which http.Request.BasicAuth does not undo.
func basicAuthCredentials(header http.Header) (string, string, bool) {
	const prefix = "Basic "

	value := header.Get("Authorization")
	if len(value) <= len(prefix) || !strings.EqualFold(value[:len(prefix)], prefix) {
		return "", "", false
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value[len(prefix):]))
	if err != nil {
		return "", "", false
	}

	clientID, clientSecret, ok := strings.Cut(string(raw), ":")
	if !ok {
		return "", "", false
	}

	clientID, err = url.QueryUnescape(clientID)
	if err != nil {
		return "", "", false
	}
	clientSecret, err = url.QueryUnescape(clientSecret)
	if err != nil {
		return "", "", false
	}

	return clientID, clientSecret, true
}

// sortedFields splits each value on whitespace and returns the result in a stable order,
// so that semantically identical scope or audience sets derive the same cache key.
func sortedFields(values []string) []string {
	var fields []string
	for _, value := range values {
		fields = append(fields, strings.Fields(value)...)
	}
	sort.Strings(fields)
	return fields
}

// tokenLifetime reads expires_in from a decoded token response.
func tokenLifetime(body map[string]any) (time.Duration, bool) {
	switch v := body["expires_in"].(type) {
	case float64:
		return time.Duration(v) * time.Second, true
	case json.Number:
		seconds, err := v.Int64()
		if err != nil {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	default:
		return 0, false
	}
}

// negativelyCacheable reports whether an upstream rejection may be remembered.
//
// Only a definitive rejection of the presented credentials qualifies. Transport errors,
// server errors and throttling must be retried against the upstream.
func negativelyCacheable(statusCode int, body []byte) bool {
	if statusCode != http.StatusBadRequest && statusCode != http.StatusUnauthorized {
		return false
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	return payload.Error == errInvalidClient
}
