package models

// TokenType says which kind of credential a signed Hub JWT is, and so which
// authorization path the API applies to it. It is stamped into the token as
// the TokenTypeClaim claim when the token is minted and is never inferred
// from other claims (for example from missing scopes).
type TokenType string

// TokenTypeClaim is the JWT claim that carries the TokenType.
const TokenTypeClaim = "tokenType"

const (
	// TokenTypeSession is an interactive login (password, passkey, SSO).
	// Nothing about the user's permissions is frozen into it: every request
	// resolves the user's current role assignments and resource access.
	TokenTypeSession TokenType = "session"

	// TokenTypeAccess is a personal access token a user creates for a tool,
	// script or MCP client. A request may use only what both the token's
	// fixed scopes and the creator's current access allow.
	TokenTypeAccess TokenType = "access"

	// TokenTypeService is a token for a service identity (the application
	// user) acting for the platform rather than for a person. A request may
	// use only what both the token's fixed scopes and the service identity's
	// current authorization allow.
	TokenTypeService TokenType = "service"
)

// TokenTypes lists every known token type.
func TokenTypes() []TokenType {
	return []TokenType{TokenTypeSession, TokenTypeAccess, TokenTypeService}
}

// Valid reports whether t is a known token type.
func (t TokenType) Valid() bool {
	switch t {
	case TokenTypeSession, TokenTypeAccess, TokenTypeService:
		return true
	}
	return false
}

// HasFixedScopes reports whether the token's scopes claim is a fixed upper
// bound on what it may do (access and service tokens). Session scopes are
// not authoritative; sessions resolve current authorization instead.
func (t TokenType) HasFixedScopes() bool {
	return t == TokenTypeAccess || t == TokenTypeService
}

// IsBound reports whether the token is bound to a stored access-token record
// (accessTokenId, organisationId, projectId, userId claims) that must still
// exist on every request, and so can be revoked and cannot be refreshed.
func (t TokenType) IsBound() bool {
	return t.HasFixedScopes()
}

// ParseTokenType reads a TokenTypeClaim value. ok is false when the claim is
// absent, not a string, or not a known type; callers must then fall back to
// the legacy (pre-token-type) behaviour, never to TokenTypeSession.
func ParseTokenType(value any) (TokenType, bool) {
	var raw string
	switch v := value.(type) {
	case string:
		raw = v
	case TokenType:
		raw = string(v)
	default:
		return "", false
	}
	t := TokenType(raw)
	return t, t.Valid()
}

// TokenClaims documents the claims a signed Hub JWT carries.
type TokenClaims struct {
	// Subject is the authenticated user (or service identity) id.
	Subject string `json:"sub"`

	// TokenType is absent on tokens minted before token types existed.
	TokenType TokenType `json:"tokenType,omitempty"`

	// Scopes are authoritative only when TokenType.HasFixedScopes().
	Scopes []Permission `json:"scopes,omitempty"`

	// Bound access/service token claims.
	AccessTokenId  string `json:"accessTokenId,omitempty"`
	OrganisationId string `json:"organisationId,omitempty"`
	ProjectId      string `json:"projectId,omitempty"`
	UserId         string `json:"userId,omitempty"`
}
