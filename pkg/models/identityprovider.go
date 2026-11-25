package models

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Identity provider connections (RBAC spec 4.16): an organisation connects
// its own identity provider. Its people sign in through it, become members
// on their first sign-in, and their provider groups decide which of the
// organisation's synced teams they are in. The identity provider decides who
// is in a team; Hub decides what the team may do.
//
// A connection is reached through its organisation, never through an email
// domain: the organisation's Slug is its sign-in name (the sign-in link, and
// on multi-tenant deployments the subdomain and the Domain field). Because a
// connection can then name any address, a person proves their email once
// before Hub links an account to the connection or creates one.

// PermissionOrganisationManageIdentityProvider manages the organisation's
// identity provider connections: who may join it and how. Owner-only and
// tenant-wide; never an access token scope.
const PermissionOrganisationManageIdentityProvider Permission = "organisation.manageIdentityProvider"

var identityProviderPermissions = []Permission{
	PermissionOrganisationManageIdentityProvider,
}

// IdentityProviderPermissions returns the canonical identity provider
// permission catalog.
func IdentityProviderPermissions() []Permission {
	return append([]Permission(nil), identityProviderPermissions...)
}

// ErrIdentityProviderInvalid wraps every validation error of a connection or
// an identity link.
var ErrIdentityProviderInvalid = errors.New("invalid identity provider connection")

// IdentityProviderProtocol is how Hub talks to the provider.
type IdentityProviderProtocol string

const (
	// IdentityProviderProtocolOIDC is OpenID Connect. SAML may follow on the
	// same connection.
	IdentityProviderProtocolOIDC IdentityProviderProtocol = "oidc"
)

// Default claim names of an OpenID Connect ID token.
const (
	DefaultSubjectClaim = "sub"
	DefaultEmailClaim   = "email"
	DefaultGroupsClaim  = "groups"
)

// IdentityProviderClaims names the claims Hub reads from the provider's
// token. An empty name means the default.
type IdentityProviderClaims struct {
	// Subject identifies the person at the provider; it never changes.
	Subject string `json:"subject,omitempty" bson:"subject,omitempty"`
	Email   string `json:"email,omitempty" bson:"email,omitempty"`
	// Groups lists the ids of the person's groups, never display names:
	// each is matched to a synced team's identityProvider.externalGroupId.
	Groups string `json:"groups,omitempty" bson:"groups,omitempty"`
}

// SubjectClaim, EmailClaim and GroupsClaim return the claim to read.
func (c IdentityProviderClaims) SubjectClaim() string {
	return orDefault(c.Subject, DefaultSubjectClaim)
}

func (c IdentityProviderClaims) EmailClaim() string {
	return orDefault(c.Email, DefaultEmailClaim)
}

func (c IdentityProviderClaims) GroupsClaim() string {
	return orDefault(c.Groups, DefaultGroupsClaim)
}

func orDefault(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

// IdentityProviderConnection is an organisation's identity provider (RBAC
// spec 4.16), stored in identity_provider_connections. Only an owner manages
// it (PermissionOrganisationManageIdentityProvider).
type IdentityProviderConnection struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId"`
	// Name identifies the connection in its organisation, and is what a
	// synced team names in identityProvider.provider. It never changes once
	// created, so the teams keep pointing at it.
	Name     string                   `json:"name" bson:"name"`
	Protocol IdentityProviderProtocol `json:"protocol" bson:"protocol"`

	// OpenID Connect.
	Issuer   string `json:"issuer" bson:"issuer"`
	ClientId string `json:"clientId" bson:"clientId"`
	// ClientSecret is write-only: stored, never sent (json "-"). Responses
	// say whether one is set.
	ClientSecret string   `json:"-" bson:"clientSecret,omitempty"`
	Scopes       []string `json:"scopes,omitempty" bson:"scopes,omitempty"`

	Claims IdentityProviderClaims `json:"claims" bson:"claims"`
	// JoinOnSignIn lets a first sign-in create the account and the
	// membership. Off (the default), sign-in links existing accounts and
	// adds memberships only; new people need an invitation.
	JoinOnSignIn bool  `json:"joinOnSignIn" bson:"joinOnSignIn"`
	Enabled      bool  `json:"enabled" bson:"enabled"`
	Audit        Audit `json:"audit" bson:"audit"`
}

// Validate checks the connection's own fields: a name, OpenID Connect with
// an https issuer (http only on a loopback host) and a client id, and valid
// claim names.
func (c IdentityProviderConnection) Validate() error {
	name := strings.TrimSpace(c.Name)
	if name == "" || len(name) > 64 || name != c.Name {
		return fmt.Errorf("%w: a name of 1 to 64 characters without surrounding spaces", ErrIdentityProviderInvalid)
	}
	if c.Protocol != IdentityProviderProtocolOIDC {
		return fmt.Errorf("%w: protocol %q", ErrIdentityProviderInvalid, c.Protocol)
	}
	if err := validateIssuer(c.Issuer); err != nil {
		return err
	}
	if strings.TrimSpace(c.ClientId) == "" {
		return fmt.Errorf("%w: client id", ErrIdentityProviderInvalid)
	}
	for _, claim := range []string{c.Claims.Subject, c.Claims.Email, c.Claims.Groups} {
		if claim != strings.TrimSpace(claim) || strings.ContainsAny(claim, " \t\n") {
			return fmt.Errorf("%w: claim name %q", ErrIdentityProviderInvalid, claim)
		}
	}
	return nil
}

func validateIssuer(issuer string) error {
	parsed, err := url.Parse(issuer)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%w: issuer %q", ErrIdentityProviderInvalid, issuer)
	}
	host := parsed.Hostname()
	loopback := host == "localhost" || host == "127.0.0.1" || host == "::1"
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && loopback) {
		return fmt.Errorf("%w: issuer %q must use https", ErrIdentityProviderInvalid, issuer)
	}
	return nil
}

// IdentityLink ties a Hub account to a person at an identity provider
// connection (RBAC spec 4.16), stored in identity_links: one per
// (connection, subject), at most one per (connection, user). Sign-in finds
// the account through it.
type IdentityLink struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId         primitive.ObjectID `json:"userId" bson:"userId"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId"`
	ConnectionId   primitive.ObjectID `json:"connectionId" bson:"connectionId"`
	// Subject is the provider's stable id of the person (the subject claim).
	Subject string `json:"subject" bson:"subject"`
	// Email is the address the provider sent when the link was made, which
	// the person proved before the link was made.
	Email    string    `json:"email" bson:"email"`
	LinkedAt time.Time `json:"linkedAt" bson:"linkedAt"`
	Audit    Audit     `json:"audit" bson:"audit"`
}

// Validate checks that the link names an account, a connection of an
// organisation and a subject.
func (l IdentityLink) Validate() error {
	if l.UserId.IsZero() || l.OrganisationId.IsZero() || l.ConnectionId.IsZero() || strings.TrimSpace(l.Subject) == "" {
		return fmt.Errorf("%w: a link needs a user, an organisation, a connection and a subject", ErrIdentityProviderInvalid)
	}
	return nil
}
