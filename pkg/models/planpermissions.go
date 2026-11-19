package models

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Plan gating: which permissions an organisation's plan decides.
//
// Gating is declared by domain (the part of a permission before the dot), so
// a new permission in a gated domain is gated too, plus single permissions
// where a domain is mixed. It does not follow permissionGroups: live.* and
// ptz.* are registered with the device permissions, and every page shares one
// group. Page permissions are named <domain>.page, so live.page,
// videowalls.page, analytics.page, cases.page and workflows.page are gated
// with their domain; the legacy pages.* names map onto them through
// CanonicalPagePermission.

// planGatedDomains are the domains whose every permission is plan gated.
var planGatedDomains = []string{
	"live",
	"ptz",
	"videowalls",
	"analytics",
	"detections",
	"workflows",
	"workflow-runs",
	"cases",
	"storage",
}

// planGatedPermissions are single plan gated permissions in domains that are
// otherwise ungated.
var planGatedPermissions = []Permission{
	PermissionMediaPlayHD,
}

// neverPlanGatedDomains and neverPlanGatedPermissions can never be gated: an
// organisation can always pay, manage itself and see what it is missing.
var (
	neverPlanGatedDomains     = []string{"billing", "organisation", "members", "roles", "teams"}
	neverPlanGatedPermissions = []Permission{PermissionSubscriptionPage, PermissionPagesSubscription}
)

var (
	// ErrPlanGatingInvalid reports a gating list that names an unknown
	// domain or permission, or gates one that must stay ungated.
	ErrPlanGatingInvalid = errors.New("invalid plan gating")
	// ErrPlanPermissionPatternInvalid reports a plan permission pattern that
	// matches no gated permission.
	ErrPlanPermissionPatternInvalid = errors.New("invalid plan permission pattern")
)

// PermissionDomain returns the domain of a permission: the part before the
// first dot, or the whole identifier when it has none.
func PermissionDomain(permission Permission) string {
	domain, _, _ := strings.Cut(string(permission), ".")
	return domain
}

func isPlanGated(permission Permission, domains []string, singles []Permission) bool {
	permission = CanonicalPagePermission(permission)
	return slices.Contains(domains, PermissionDomain(permission)) || slices.Contains(singles, permission)
}

// IsPlanGated reports whether a catalog permission is plan gated. Legacy
// page names are mapped to their canonical name first; unknown permissions
// are not gated (they grant nothing anyway).
func IsPlanGated(permission Permission) bool {
	definition, ok := LookupPermission(permission)
	return ok && definition.PlanGated
}

// PlanGatedPermissions returns every plan gated permission in catalog order.
func PlanGatedPermissions() []Permission {
	var gated []Permission
	for _, definition := range permissionDefinitions {
		if definition.PlanGated {
			gated = append(gated, definition.Permission)
		}
	}
	return gated
}

// UngatedPermissions returns every permission that is allowed on every plan,
// in catalog order.
func UngatedPermissions() []Permission {
	var ungated []Permission
	for _, definition := range permissionDefinitions {
		if !definition.PlanGated {
			ungated = append(ungated, definition.Permission)
		}
	}
	return ungated
}

// validatePlanGating rejects gating on the domains and permissions that must
// stay ungated.
func validatePlanGating(definitions []PermissionDefinition) error {
	for _, definition := range definitions {
		if !definition.PlanGated {
			continue
		}
		if slices.Contains(neverPlanGatedDomains, PermissionDomain(definition.Permission)) ||
			slices.Contains(neverPlanGatedPermissions, definition.Permission) {
			return fmt.Errorf("%w: %q must never be plan gated", ErrPlanGatingInvalid, definition.Permission)
		}
	}
	return nil
}

// validatePlanGatingLists checks that every gated domain has at least one
// permission and every single gated permission is in the catalog, so a
// misspelt entry fails instead of gating nothing, and that neither list names
// something that must stay ungated.
func validatePlanGatingLists(definitions []PermissionDefinition, domains []string, singles []Permission) error {
	for _, domain := range domains {
		if slices.Contains(neverPlanGatedDomains, domain) {
			return fmt.Errorf("%w: domain %q must never be plan gated", ErrPlanGatingInvalid, domain)
		}
		if !slices.ContainsFunc(definitions, func(definition PermissionDefinition) bool {
			return PermissionDomain(definition.Permission) == domain
		}) {
			return fmt.Errorf("%w: gated domain %q has no permissions", ErrPlanGatingInvalid, domain)
		}
	}
	for _, permission := range singles {
		if slices.Contains(neverPlanGatedPermissions, permission) || slices.Contains(neverPlanGatedDomains, PermissionDomain(permission)) {
			return fmt.Errorf("%w: %q must never be plan gated", ErrPlanGatingInvalid, permission)
		}
		if !slices.ContainsFunc(definitions, func(definition PermissionDefinition) bool {
			return definition.Permission == permission
		}) {
			return fmt.Errorf("%w: gated permission %q is not in the catalog", ErrPlanGatingInvalid, permission)
		}
	}
	return nil
}

// ValidatePermissionCatalog checks the release's permission catalog,
// including plan gating. Services may call it at startup; the package tests
// run it on every change.
func ValidatePermissionCatalog() error {
	if err := validatePermissionGroups(permissionGroups); err != nil {
		return err
	}
	return validatePlanGatingLists(permissionDefinitions, planGatedDomains, planGatedPermissions)
}

// MatchPlanPermissionPattern reports whether a plan permission pattern covers
// a permission. A pattern is "domain.*", covering every permission of the
// domain including ones added later, or a single "domain.action"; legacy
// page names are mapped to their canonical name.
func MatchPlanPermissionPattern(pattern string, permission Permission) bool {
	permission = CanonicalPagePermission(permission)
	if domain, ok := strings.CutSuffix(pattern, ".*"); ok {
		return domain != "" && PermissionDomain(permission) == domain
	}
	return CanonicalPagePermission(Permission(pattern)) == permission
}

// planPatternPermissions returns the gated permissions a pattern covers, in
// catalog order.
func planPatternPermissions(definitions []PermissionDefinition, pattern string) []Permission {
	var matched []Permission
	for _, definition := range definitions {
		if definition.PlanGated && MatchPlanPermissionPattern(pattern, definition.Permission) {
			matched = append(matched, definition.Permission)
		}
	}
	return matched
}

// ValidatePlanPermissionPatterns checks that every pattern covers at least one
// gated permission, as plan files and subscription overrides require.
func ValidatePlanPermissionPatterns(patterns []string) error {
	for _, pattern := range patterns {
		if len(planPatternPermissions(permissionDefinitions, pattern)) == 0 {
			return fmt.Errorf("%w: %q matches no plan gated permission", ErrPlanPermissionPatternInvalid, pattern)
		}
	}
	return nil
}

// ResolvePlanPermissions returns the gated permissions a plan allows, in
// catalog order: those covered by patterns and additions together with
// everything they imply, without those covered by removals. A removal also
// removes every permission that implies a removed one, so the result stays
// closed under Implies (removing ptz.panTilt also removes ptz.control).
// Patterns that match no gated permission add or remove nothing.
func ResolvePlanPermissions(patterns, add, remove []string) []Permission {
	return resolvePlanPermissions(permissionDefinitions, permissionIndex, patterns, add, remove)
}

func resolvePlanPermissions(definitions []PermissionDefinition, index map[Permission]int, patterns, add, remove []string) []Permission {
	var allowed []Permission
	for _, pattern := range slices.Concat(patterns, add) {
		allowed = append(allowed, planPatternPermissions(definitions, pattern)...)
	}
	allowed = expandPermissions(definitions, index, allowed)

	removed := make(map[Permission]bool)
	for _, pattern := range remove {
		for _, permission := range planPatternPermissions(definitions, pattern) {
			removed[permission] = true
		}
	}
	// Remove every permission that implies a removed one, until stable.
	for changed := len(removed) > 0; changed; {
		changed = false
		for _, definition := range definitions {
			if removed[definition.Permission] {
				continue
			}
			if slices.ContainsFunc(definition.Implies, func(implied Permission) bool { return removed[implied] }) {
				removed[definition.Permission] = true
				changed = true
			}
		}
	}

	var resolved []Permission
	for _, permission := range allowed {
		if !removed[permission] && definitions[index[permission]].PlanGated {
			resolved = append(resolved, permission)
		}
	}
	return resolved
}
