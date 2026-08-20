package models

import (
	"errors"
	"fmt"
)

// ErrOrganisationSlugInvalid is returned when an organisation slug is not in
// canonical form, or when it is reserved by the platform.
var ErrOrganisationSlugInvalid = errors.New("organisation slug is not well formed")

// MaxOrganisationSlugLength is the longest an organisation slug may be. See
// maxSlugLength.
const MaxOrganisationSlugLength = maxSlugLength

// reservedOrganisationSlugs are the organisation slugs no caller may claim.
//
// This list is longer than reservedProjectSlugs, and deliberately so. An
// organisation slug is unique *globally* — the index is a single-key unique
// partial index on {slug: 1}, not the compound {organisationId, slug} the
// project collection uses — and it is the outermost path segment, so it is the
// one that would collide with the platform's own top-level routes. A word
// claimed here is claimed for every tenant at once, which makes both the
// routing collisions and the squatting worth guarding against up front.
//
// Reserving is close to one-way: removing a word later is free, adding one
// later can invalidate slugs already stored and referenced. No organisation in
// production holds a slug yet, so the cost of being generous is zero today and
// only rises from here.
var reservedOrganisationSlugs = map[string]struct{}{
	// Words that read as "no organisation" or "every organisation". An
	// organisation whose slug is a sentinel invites code that treats it as one.
	"default": {}, "none": {}, "null": {}, "nil": {}, "undefined": {},
	"all": {}, "any": {}, "unassigned": {}, "unknown": {},

	// Platform surfaces and infrastructure hostnames. These are the segments a
	// /<orgSlug>/... scheme collides with directly.
	"api": {}, "admin": {}, "app": {}, "assets": {}, "static": {}, "public": {},
	"internal": {}, "system": {}, "health": {}, "healthz": {}, "readyz": {},
	"status": {}, "metrics": {}, "well-known": {}, "www": {}, "mail": {},
	"smtp": {}, "ftp": {}, "cdn": {}, "ns": {}, "mx": {}, "localhost": {},
	"graphql": {}, "grpc": {}, "ws": {}, "websocket": {}, "webhook": {},
	"webhooks": {}, "sitemap": {}, "robots": {}, "favicon": {},

	// Authentication and account lifecycle routes.
	"login": {}, "logout": {}, "signin": {}, "signout": {}, "signup": {},
	"register": {}, "auth": {}, "oauth": {}, "sso": {}, "saml": {}, "callback": {},
	"token": {}, "session": {}, "password": {}, "forgot": {}, "reset": {},
	"verify": {}, "confirm": {}, "activate": {}, "invite": {}, "invites": {},
	"invitation": {}, "invitations": {}, "mfa": {}, "2fa": {},

	// Resource collection names, which are what a reader most easily mistakes an
	// organisation slug for.
	"organisation": {}, "organisations": {}, "organization": {},
	"organizations": {}, "org": {}, "orgs": {}, "project": {}, "projects": {},
	"user": {}, "users": {}, "account": {}, "accounts": {}, "team": {},
	"teams": {}, "member": {}, "members": {}, "role": {}, "roles": {},
	"permission": {}, "permissions": {}, "group": {}, "groups": {},
	"device": {}, "devices": {}, "media": {}, "site": {}, "sites": {},

	// Generic action and navigation segments.
	"new": {}, "create": {}, "edit": {}, "update": {}, "delete": {},
	"search": {}, "settings": {}, "config": {}, "dashboard": {}, "home": {},
	"index": {}, "me": {}, "my": {}, "self": {}, "profile": {},

	// Commercial and legal surfaces, which tend to become top-level routes on
	// the marketing side of the same hostname.
	"billing": {}, "invoice": {}, "invoices": {}, "subscription": {},
	"subscriptions": {}, "plan": {}, "plans": {}, "pricing": {}, "checkout": {},
	"legal": {}, "privacy": {}, "terms": {}, "security": {}, "abuse": {},
	"support": {}, "help": {}, "docs": {}, "documentation": {}, "blog": {},
	"about": {}, "contact": {}, "careers": {},

	// Environment names, which are the words most likely to be typed into a
	// tenant-creation form during a test and then never cleaned up.
	"test": {}, "demo": {}, "example": {}, "sandbox": {}, "staging": {},
	"production": {}, "dev": {}, "development": {}, "preview": {},

	// Product and company names, held back from first-come-first-served because
	// the namespace is global.
	"kerberos": {}, "kerberosio": {}, "uug": {}, "uugai": {}, "hub": {},
}

// IsReservedOrganisationSlug reports whether slug is reserved by the platform.
//
// The argument is case-sensitive and assumes canonical form; ValidateOrganisationSlug
// runs the format check first for exactly that reason. Prefer that function —
// this one is exported for callers that need to explain *why* a slug was
// refused, or to surface the distinction in documentation.
func IsReservedOrganisationSlug(slug string) bool {
	_, reserved := reservedOrganisationSlugs[slug]
	return reserved
}

// ReservedOrganisationSlugs returns the reserved organisation slugs in sorted
// order. It returns a fresh slice on every call so a caller cannot mutate the
// set.
func ReservedOrganisationSlugs() []string {
	return sortedSlugs(reservedOrganisationSlugs)
}

// ValidateOrganisationSlugFormat reports whether slug is in canonical form. It
// does not consider whether the slug is reserved; see IsReservedOrganisationSlug.
//
// Most callers want ValidateOrganisationSlug instead. This exists for the same
// reason its project counterpart does: to keep the shape check available on its
// own, so a caller that has already established a slug's provenance is not
// forced to re-apply the reserved list to it.
func ValidateOrganisationSlugFormat(slug string) error {
	return validateSlugFormat(slug, ErrOrganisationSlugInvalid)
}

// ValidateOrganisationSlug reports whether slug may be written to an
// organisation. It combines the format and reserved checks, which the project
// equivalent deliberately keeps apart.
//
// The asymmetry is not an inconsistency. DefaultProjectSlug is reserved *and*
// held by a real document that the server mints itself, so folding the checks
// together there would reject the one write meant to succeed. Nothing mints an
// organisation slug — the bootstrap tool is forbidden from synthesizing one, and
// the field is optional precisely so it can stay absent — so for organisations
// every write is caller-supplied and both checks always apply.
//
// The slug is optional: an organisation with no slug is valid, and its absence
// is what the partial unique index is built around. Callers must therefore test
// for an absent slug before calling this, rather than passing "" and expecting
// it to pass. Writing "" is not the same as writing nothing: the index treats
// the empty string as a value, so the second organisation to store one would
// collide with the first, globally.
func ValidateOrganisationSlug(slug string) error {
	if err := ValidateOrganisationSlugFormat(slug); err != nil {
		return err
	}
	if IsReservedOrganisationSlug(slug) {
		return fmt.Errorf("%w: slug is reserved by the platform", ErrOrganisationSlugInvalid)
	}
	return nil
}
