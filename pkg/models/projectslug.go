package models

import (
	"errors"
	"fmt"
	"sort"
)

// ErrProjectSlugInvalid is returned when a slug is not in canonical form.
var ErrProjectSlugInvalid = errors.New("project slug is not well formed")

const (
	// MaxProjectSlugLength follows the DNS label limit. A slug is intended to be
	// usable as a URL path segment, so bounding it here keeps that option open.
	MaxProjectSlugLength = 63

	// objectIDHexLength is the width of a hex-encoded ObjectID.
	objectIDHexLength = 24
)

// reservedProjectSlugs are the slugs no caller may claim.
//
// Reserving is close to one-way. Removing a word later is free, but adding one
// later can invalidate slugs that are already stored and already referenced, so
// the list errs towards generous now while the project collection is still
// effectively empty.
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
// this only for caller-supplied ones.
//
// The argument is case-sensitive and assumes canonical form. Pass it through
// ValidateProjectSlugFormat first: without that, "Default" misses this check,
// misses an exact comparison against DefaultProjectSlug, and misses a
// byte-comparison unique index.
func IsReservedProjectSlug(slug string) bool {
	_, reserved := reservedProjectSlugs[slug]
	return reserved
}

// ReservedProjectSlugs returns the reserved slugs in sorted order. It returns a
// fresh slice on every call so a caller cannot mutate the set. Intended for
// surfacing the list in API documentation and error responses.
func ReservedProjectSlugs() []string {
	slugs := make([]string, 0, len(reservedProjectSlugs))
	for slug := range reservedProjectSlugs {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs
}

// ValidateProjectSlugFormat reports whether slug is in canonical form. It does
// not consider whether the slug is reserved; see IsReservedProjectSlug.
//
// The rules live here, next to DefaultProjectId and ResolveProjectId, for the
// same reason those do: a slug rule that two services disagree about is worse
// than no rule. One service accepting what another rejects means a slug that is
// writable but unroutable, or a reserved word that is unforgeable in one
// service and forgeable in the next.
//
// The rules are deliberately narrow, because a slug is the one project
// identifier meant to be stable and human-visible: once it is written and
// referenced, tightening the rules means rewriting stored values and breaking
// whatever pointed at them.
//
//   - lowercase letters, digits and "-" only, so a slug is unambiguous in a URL
//     and comparisons need no case folding or collation.
//   - no leading, trailing or repeated "-", so one logical name has one spelling.
//   - not shaped like an ObjectID, so a route carrying both a slug and an id can
//     tell them apart without guessing.
//
// Error messages describe the rule that failed and never echo the supplied
// value, which keeps caller input out of logs and responses.
func ValidateProjectSlugFormat(slug string) error {
	if slug == "" {
		return fmt.Errorf("%w: slug is empty", ErrProjectSlugInvalid)
	}
	if len(slug) > MaxProjectSlugLength {
		return fmt.Errorf("%w: slug is longer than %d characters", ErrProjectSlugInvalid, MaxProjectSlugLength)
	}

	for i := 0; i < len(slug); i++ {
		c := slug[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '-':
			if i == 0 || i == len(slug)-1 {
				return fmt.Errorf("%w: slug may not start or end with a hyphen", ErrProjectSlugInvalid)
			}
			if slug[i-1] == '-' {
				return fmt.Errorf("%w: slug may not contain consecutive hyphens", ErrProjectSlugInvalid)
			}
		default:
			return fmt.Errorf("%w: slug may only contain lowercase letters, digits and hyphens", ErrProjectSlugInvalid)
		}
	}

	if isObjectIDShapedSlug(slug) {
		return fmt.Errorf("%w: slug may not be shaped like an object id", ErrProjectSlugInvalid)
	}

	return nil
}

// isObjectIDShapedSlug reports whether slug could be read as a hex-encoded
// ObjectID. Only the lowercase form is checked because ValidateProjectSlugFormat
// has already rejected uppercase by the time this runs.
func isObjectIDShapedSlug(slug string) bool {
	if len(slug) != objectIDHexLength {
		return false
	}
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') {
			continue
		}
		return false
	}
	return true
}
