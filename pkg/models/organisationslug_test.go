package models

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateOrganisationSlugAcceptsCanonicalForms pins the shapes a caller is
// allowed to keep.
func TestValidateOrganisationSlugAcceptsCanonicalForms(t *testing.T) {
	for _, slug := range []string{
		"a",
		"7",
		"acme-corp",
		"acme-2",
		"a-b-c-d",
		strings.Repeat("a", MaxOrganisationSlugLength),
	} {
		t.Run(slug, func(t *testing.T) {
			if err := ValidateOrganisationSlug(slug); err != nil {
				t.Fatalf("validate %q = %v, want nil", slug, err)
			}
		})
	}
}

// TestValidateOrganisationSlugRejectsNonCanonicalForms documents every rule. The
// uppercase cases matter most: the global unique index compares bytes, so
// without this check "Admin" would clear the reserved list and still occupy a
// name no other tenant could take in canonical form.
func TestValidateOrganisationSlugRejectsNonCanonicalForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		slug string
	}{
		{name: "empty", slug: ""},
		{name: "whitespace only", slug: "   "},
		{name: "inner space", slug: "acme corp"},
		{name: "leading space", slug: " acme-corp"},
		{name: "trailing space", slug: "acme-corp "},
		{name: "uppercase", slug: "Acme-Corp"},
		{name: "reserved slug in another case", slug: "Admin"},
		{name: "reserved slug shouting", slug: "API"},
		{name: "underscore", slug: "acme_corp"},
		{name: "path traversal", slug: "../etc"},
		{name: "slash", slug: "acme/corp"},
		{name: "percent encoded", slug: "acme%20corp"},
		{name: "leading hyphen", slug: "-acme"},
		{name: "trailing hyphen", slug: "acme-"},
		{name: "consecutive hyphens", slug: "acme--corp"},
		{name: "too long", slug: strings.Repeat("a", MaxOrganisationSlugLength+1)},
		{name: "object id shaped", slug: "507f1f77bcf86cd799439011"},
		{name: "non ascii", slug: "äcme"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOrganisationSlug(tc.slug)
			if !errors.Is(err, ErrOrganisationSlugInvalid) {
				t.Fatalf("validate %q = %v, want ErrOrganisationSlugInvalid", tc.slug, err)
			}
			if tc.slug != "" && strings.Contains(err.Error(), tc.slug) {
				t.Fatalf("error message echoes the supplied value: %v", err)
			}
		})
	}
}

// TestValidateOrganisationSlugRejectsEveryReservedSlug walks the whole set
// rather than sampling it, so a word added to the list without being reachable
// through the validator is caught here rather than in production.
func TestValidateOrganisationSlugRejectsEveryReservedSlug(t *testing.T) {
	reserved := ReservedOrganisationSlugs()
	if len(reserved) == 0 {
		t.Fatal("reserved slugs = 0, want a non-empty set")
	}

	for _, slug := range reserved {
		if err := ValidateOrganisationSlug(slug); !errors.Is(err, ErrOrganisationSlugInvalid) {
			t.Errorf("validate reserved %q = %v, want ErrOrganisationSlugInvalid", slug, err)
		}
	}
}

// TestReservedOrganisationSlugsAreThemselvesCanonical keeps the list reachable.
// A reserved word that fails format validation is dead weight: no caller could
// ever submit it, so reserving it protects nothing and misleads whoever reads
// the list.
func TestReservedOrganisationSlugsAreThemselvesCanonical(t *testing.T) {
	for _, slug := range ReservedOrganisationSlugs() {
		if err := ValidateOrganisationSlugFormat(slug); err != nil {
			t.Errorf("reserved slug %q is not itself a valid slug: %v", slug, err)
		}
	}
}

// TestValidateOrganisationSlugFormatIgnoresTheReservedSet separates the two
// checks. The format function is exported on its own, so it has to stay a pure
// shape check even though the combined validator is what callers normally use.
func TestValidateOrganisationSlugFormatIgnoresTheReservedSet(t *testing.T) {
	for _, slug := range ReservedOrganisationSlugs() {
		if err := ValidateOrganisationSlugFormat(slug); err != nil {
			t.Fatalf("format check rejected reserved slug %q: %v", slug, err)
		}
	}
}

// TestIsReservedOrganisationSlugIsCaseSensitive documents the contract callers
// rely on: this check assumes canonical input and is not a substitute for
// format validation.
func TestIsReservedOrganisationSlugIsCaseSensitive(t *testing.T) {
	for _, slug := range []string{"Admin", "ADMIN", "Api", "acme-corp", ""} {
		if IsReservedOrganisationSlug(slug) {
			t.Errorf("IsReservedOrganisationSlug(%q) = true, want false", slug)
		}
	}
}

// TestReservedOrganisationSlugsReturnsAnIsolatedCopy stops a caller mutating the
// set through the accessor.
func TestReservedOrganisationSlugsReturnsAnIsolatedCopy(t *testing.T) {
	first := ReservedOrganisationSlugs()
	if len(first) == 0 {
		t.Fatal("reserved slugs = 0, want a non-empty set")
	}
	first[0] = "mutated"

	for _, slug := range ReservedOrganisationSlugs() {
		if slug == "mutated" {
			t.Fatal("mutating the returned slice changed the reserved set")
		}
	}
}

// TestOrganisationAndProjectSlugSetsAreIndependent pins the decision to keep two
// literal lists instead of deriving one from the other. The organisation slug is
// globally unique and sits in the outer path segment, so it reserves words the
// project slug has no reason to; this fails if someone collapses them into a
// shared base and quietly widens the project set.
func TestOrganisationAndProjectSlugSetsAreIndependent(t *testing.T) {
	project := map[string]struct{}{}
	for _, slug := range ReservedProjectSlugs() {
		project[slug] = struct{}{}
	}

	var onlyOrganisation int
	for _, slug := range ReservedOrganisationSlugs() {
		if _, shared := project[slug]; !shared {
			onlyOrganisation++
		}
	}
	if onlyOrganisation == 0 {
		t.Fatal("no organisation-only reserved slugs, want the sets to differ")
	}

	// The default project slug is reserved for projects because a real document
	// holds it. Nothing holds it for organisations, but an organisation called
	// "default" reads as "no organisation" everywhere the project sentinel reads
	// as "no project", so it is reserved on both sides for different reasons.
	if !IsReservedOrganisationSlug(DefaultProjectSlug) {
		t.Errorf("IsReservedOrganisationSlug(%q) = false, want true", DefaultProjectSlug)
	}
}

// TestOrganisationAndProjectSlugFormatsAgree pins the shared shape engine: the
// reserved sets are allowed to diverge, the shape rules are not. A caller should
// never have to remember which slug kind tolerates an underscore.
func TestOrganisationAndProjectSlugFormatsAgree(t *testing.T) {
	for _, slug := range []string{
		"a", "acme-corp", "site-2", "Acme", "acme_corp", "acme--corp",
		"-acme", "acme-", "", "   ", "../etc", "507f1f77bcf86cd799439011",
		strings.Repeat("a", maxSlugLength),
		strings.Repeat("a", maxSlugLength+1),
	} {
		t.Run(slug, func(t *testing.T) {
			organisation := ValidateOrganisationSlugFormat(slug) != nil
			project := ValidateProjectSlugFormat(slug) != nil
			if organisation != project {
				t.Fatalf("format checks disagree on %q: organisation rejected = %v, project rejected = %v",
					slug, organisation, project)
			}
		})
	}
}
