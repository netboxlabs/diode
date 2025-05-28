package authutil

const (
	// ScopeDiodeRead is the OAuth2 scope for read access to diode services
	ScopeDiodeRead = "diode:read"
	// ScopeDiodeWrite is the OAuth2 scope for write access to diode services
	ScopeDiodeWrite = "diode:write"
	// ScopeDiodeIngest is the OAuth2 scope for ingest access to diode services
	ScopeDiodeIngest = "diode:ingest"
	// ScopeNetBoxRead is the OAuth2 scope for read access to NetBox diode data plugin
	ScopeNetBoxRead = "netbox:read"
	// ScopeNetBoxWrite is the OAuth2 scope for write access to NetBox diode data plugin
	ScopeNetBoxWrite = "netbox:write"
)

// authutilContextKey is a type for context keys to avoid collisions
type authutilContextKey string

const (
	// ContextKeyScope is the context key for the auth token scopes
	ContextKeyScope authutilContextKey = "tokenScope"
)
