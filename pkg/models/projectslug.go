package models

import (
	"errors"
)

// ErrProjectSlugInvalid is returned when a project slug is not in canonical form.
var ErrProjectSlugInvalid = errors.New("project slug is not well formed")

// MaxProjectSlugLength is the longest a project slug may be. See maxSlugLength.
const MaxProjectSlugLength = maxSlugLength

// reservedProjectSlugs are the project slugs no caller may claim.
//
// Reserving is close to one-way. Removing a word later is free, but adding one
// later can invalidate slugs that are already stored and already referenced, so
// the list errs towards generous now while the project collection is still
// effectively empty.
//
// This set is deliberately kept separate from reservedOrganisationSlugs rather
// than derived from it. The two slugs occupy different path positions, so a word
// added to one for a routing reason is usually irrelevant to the other; a shared
// base would make every such addition a change to both.
//
// The groups below are the reasons a word is here; they are not a taxonomy
// anyone needs to preserve when editing the list.
var reservedProjectSlugs = map[string]struct{}{
	// The default project identity, plus the words that read as "no project" or
	// "every project". A slug that looks like a sentinel invites code that
	// special-cases it as one.
	"default": {}, "none": {}, "null": {}, "nil": {}, "undefined": {},
	"all": {}, "any": {}, "unassigned": {},

	// Path segments a future /<orgSlug>/<projectSlug> scheme would collide with,
	// or that name the platform's own surfaces.
	"api": {}, "admin": {}, "app": {}, "assets": {}, "static": {}, "public": {},
	"internal": {}, "system": {}, "health": {}, "healthz": {}, "status": {},
	"metrics": {}, "well-known": {},

	// Authentication routes.
	"login": {}, "logout": {}, "signin": {}, "signout": {}, "signup": {},
	"auth": {}, "oauth": {}, "callback": {}, "token": {}, "session": {},

	// Resource collection names. These are the ones most likely to appear as a
	// sibling segment of a project slug.
	"organisation": {}, "organisations": {}, "org": {}, "orgs": {},
	"project": {}, "projects": {}, "user": {}, "users": {},
	"account": {}, "accounts": {},

	// Generic action segments.
	"new": {}, "create": {}, "edit": {}, "update": {}, "delete": {},
	"search": {}, "settings": {}, "config": {},
}

// IsReservedProjectSlug reports whether slug is reserved by the platform.
//
// This is deliberately separate from ValidateProjectSlugFormat rather than
// folded into it. DefaultProjectSlug is both reserved and legitimately held —
// by exactly one document per organisation, the server-minted default — so a
// single function returning "invalid" for it would reject the one write that is
// supposed to succeed. Callers validate the format for every slug, and consult
// this only for caller-supplied ones. The organisation slug has no such
// server-minted holder, which is why ValidateOrganisationSlug can combine the
// two checks and this one cannot.
//
// The argument is case-sensitive and assumes canonical form. Pass it through
// ValidateProjectSlugFormat first: without that, "Default" misses this check,
// misses an exact comparison against DefaultProjectSlug, and misses a
// byte-comparison unique index.
func IsReservedProjectSlug(slug string) bool {
	_, reserved := reservedProjectSlugs[slug]
	return reserved
}

// ReservedProjectSlugs returns the reserved project slugs in sorted order. It
// returns a fresh slice on every call so a caller cannot mutate the set.
// Intended for surfacing the list in API documentation and error responses.
func ReservedProjectSlugs() []string {
	return sortedSlugs(reservedProjectSlugs)
}

// ValidateProjectSlugFormat reports whether slug is in canonical form. It does
// not consider whether the slug is reserved; see IsReservedProjectSlug.
//
// The rules live here, next to DefaultProjectId and ResolveProjectId, for the
// same reason those do: a slug rule that two services disagree about is worse
// than no rule. One service accepting what another rejects means a slug that is
// writable but unroutable, or a reserved word that is unforgeable in one
// service and forgeable in the next.
func ValidateProjectSlugFormat(slug string) error {
	return validateSlugFormat(slug, ErrProjectSlugInvalid)
}
