package models

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateProjectSlugFormatAcceptsCanonicalForms pins the shapes a caller is
// allowed to keep, including the reserved default slug: the server mints that
// document itself, so the constant has to survive format validation even though
// IsReservedProjectSlug refuses it to callers.
func TestValidateProjectSlugFormatAcceptsCanonicalForms(t *testing.T) {
	for _, slug := range []string{
		"a",
		"7",
		"warehouse-north",
		"site-2",
		"a-b-c-d",
		DefaultProjectSlug,
		strings.Repeat("a", MaxProjectSlugLength),
	} {
		t.Run(slug, func(t *testing.T) {
			if err := ValidateProjectSlugFormat(slug); err != nil {
				t.Fatalf("validate %q = %v, want nil", slug, err)
			}
		})
	}
}

// TestValidateProjectSlugFormatRejectsNonCanonicalForms documents every rule.
// The uppercase cases matter most: they are what makes a reserved slug
// unforgeable, since IsReservedProjectSlug and the unique {organisationId, slug}
// index both compare bytes.
func TestValidateProjectSlugFormatRejectsNonCanonicalForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		slug string
	}{
		{name: "empty", slug: ""},
		{name: "whitespace only", slug: "   "},
		{name: "inner space", slug: "warehouse north"},
		{name: "leading space", slug: " warehouse-north"},
		{name: "trailing space", slug: "warehouse-north "},
		{name: "uppercase", slug: "Warehouse-North"},
		{name: "default slug in another case", slug: "Project-1"},
		{name: "default slug shouting", slug: "PROJECT-1"},
		{name: "reserved slug in another case", slug: "Default"},
		{name: "reserved slug shouting", slug: "DEFAULT"},
		{name: "underscore", slug: "warehouse_north"},
		{name: "path traversal", slug: "../etc"},
		{name: "slash", slug: "warehouse/north"},
		{name: "percent encoded", slug: "warehouse%20north"},
		{name: "leading hyphen", slug: "-warehouse"},
		{name: "trailing hyphen", slug: "warehouse-"},
		{name: "consecutive hyphens", slug: "warehouse--north"},
		{name: "too long", slug: strings.Repeat("a", MaxProjectSlugLength+1)},
		{name: "object id shaped", slug: "507f1f77bcf86cd799439011"},
		{name: "non ascii", slug: "wärehouse"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateProjectSlugFormat(tc.slug)
			if !errors.Is(err, ErrProjectSlugInvalid) {
				t.Fatalf("validate %q = %v, want ErrProjectSlugInvalid", tc.slug, err)
			}
			if tc.slug != "" && strings.Contains(err.Error(), tc.slug) {
				t.Fatalf("error message echoes the supplied value: %v", err)
			}
		})
	}
}

// TestValidateProjectSlugFormatRejectsIdShapedSlugsOnly guards the narrowness of
// the object-id rule: it must reject a 24-character hex string without rejecting
// every string that happens to be 24 characters long.
func TestValidateProjectSlugFormatRejectsIdShapedSlugsOnly(t *testing.T) {
	idShaped := "507f1f77bcf86cd799439011"
	if len(idShaped) != objectIDHexLength {
		t.Fatalf("fixture length = %d, want %d", len(idShaped), objectIDHexLength)
	}
	if err := ValidateProjectSlugFormat(idShaped); !errors.Is(err, ErrProjectSlugInvalid) {
		t.Fatalf("validate %q = %v, want ErrProjectSlugInvalid", idShaped, err)
	}

	// Same length, but the trailing "z" puts it outside the hex alphabet.
	sameLength := "507f1f77bcf86cd79943901z"
	if len(sameLength) != objectIDHexLength {
		t.Fatalf("fixture length = %d, want %d", len(sameLength), objectIDHexLength)
	}
	if err := ValidateProjectSlugFormat(sameLength); err != nil {
		t.Fatalf("validate %q = %v, want nil", sameLength, err)
	}
}

// TestReservedProjectSlugsAreThemselvesCanonical keeps the list reachable. A
// reserved word that fails format validation is dead weight: no caller could
// ever submit it, so reserving it protects nothing and misleads whoever reads
// the list.
func TestReservedProjectSlugsAreThemselvesCanonical(t *testing.T) {
	for _, slug := range ReservedProjectSlugs() {
		if err := ValidateProjectSlugFormat(slug); err != nil {
			t.Errorf("reserved slug %q is not itself a valid slug: %v", slug, err)
		}
	}
}

// TestDefaultProjectSlugIsReserved ties the constant to the list. The default
// project identity is only unforgeable while no caller can claim its slug.
func TestDefaultProjectSlugIsReserved(t *testing.T) {
	if !IsReservedProjectSlug(DefaultProjectSlug) {
		t.Fatalf("IsReservedProjectSlug(%q) = false, want true", DefaultProjectSlug)
	}
}

// TestIsReservedProjectSlugIsCaseSensitive documents the contract callers rely
// on: this check assumes canonical input and is not a substitute for format
// validation.
func TestIsReservedProjectSlugIsCaseSensitive(t *testing.T) {
	for _, slug := range []string{"Project-1", "PROJECT-1", "Default", "DEFAULT", "Admin", "warehouse-north", ""} {
		if IsReservedProjectSlug(slug) {
			t.Errorf("IsReservedProjectSlug(%q) = true, want false", slug)
		}
	}
}

// TestReservedProjectSlugsReturnsAnIsolatedCopy stops a caller mutating the set
// through the accessor.
func TestReservedProjectSlugsReturnsAnIsolatedCopy(t *testing.T) {
	first := ReservedProjectSlugs()
	if len(first) == 0 {
		t.Fatal("reserved slugs = 0, want a non-empty set")
	}
	first[0] = "mutated"

	for _, slug := range ReservedProjectSlugs() {
		if slug == "mutated" {
			t.Fatal("mutating the returned slice changed the reserved set")
		}
	}
}
