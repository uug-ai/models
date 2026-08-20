package models

import (
	"fmt"
	"sort"
)

// maxSlugLength follows the DNS label limit. A slug is intended to be usable as
// a URL path segment, so bounding it here keeps that option open.
const maxSlugLength = 63

// objectIDHexLength is the width of a hex-encoded ObjectID.
const objectIDHexLength = 24

// validateSlugFormat reports whether slug is in canonical form, wrapping the
// caller's sentinel so each slug kind keeps its own errors.Is target.
//
// The organisation slug and the project slug sit at different points of a URL
// and reserve different words, but they are the same kind of identifier and a
// caller should never have to remember which one tolerates an underscore. The
// shape rules are therefore shared verbatim; only the reserved sets differ.
//
// The rules are deliberately narrow, because a slug is the one identifier meant
// to be stable and human-visible: once it is written and referenced, tightening
// the rules means rewriting stored values and breaking whatever pointed at them.
//
//   - lowercase letters, digits and "-" only, so a slug is unambiguous in a URL
//     and comparisons need no case folding or collation.
//   - no leading, trailing or repeated "-", so one logical name has one spelling.
//   - not shaped like an ObjectID, so a route carrying both a slug and an id can
//     tell them apart without guessing.
//
// Error messages describe the rule that failed and never echo the supplied
// value, which keeps caller input out of logs and responses.
func validateSlugFormat(slug string, sentinel error) error {
	if slug == "" {
		return fmt.Errorf("%w: slug is empty", sentinel)
	}
	if len(slug) > maxSlugLength {
		return fmt.Errorf("%w: slug is longer than %d characters", sentinel, maxSlugLength)
	}

	for i := 0; i < len(slug); i++ {
		c := slug[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '-':
			if i == 0 || i == len(slug)-1 {
				return fmt.Errorf("%w: slug may not start or end with a hyphen", sentinel)
			}
			if slug[i-1] == '-' {
				return fmt.Errorf("%w: slug may not contain consecutive hyphens", sentinel)
			}
		default:
			return fmt.Errorf("%w: slug may only contain lowercase letters, digits and hyphens", sentinel)
		}
	}

	if isObjectIDShapedSlug(slug) {
		return fmt.Errorf("%w: slug may not be shaped like an object id", sentinel)
	}

	return nil
}

// isObjectIDShapedSlug reports whether slug could be read as a hex-encoded
// ObjectID. Only the lowercase form is checked because validateSlugFormat has
// already rejected uppercase by the time this runs.
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

// sortedSlugs returns the members of a reserved set in sorted order, as a fresh
// slice so a caller cannot mutate the set through the accessor.
func sortedSlugs(set map[string]struct{}) []string {
	slugs := make([]string, 0, len(set))
	for slug := range set {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs
}
