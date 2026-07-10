package models

import (
	"regexp"
	"strings"
)

// EvaluateCondition tests a StageCondition against a root — the credential-free
// projection of a workflow envelope as a nested map[string]any (inputs.<op>,
// results.<op>, device, user, and the top-level identity scalars). A nil
// condition always matches. Path is an absolute dot-separated lookup into root.
//
// It is the single, dependency-free source of condition-matching semantics for
// the whole workflows system: the hub-workflows stage engine delegates to it for
// stage conditions and automatic workflow triggers reuse it for device (and any
// envelope) scoping, so triggers and stages match with the very same operators
// and semantics. There is no second implementation to keep in sync — any change
// to operator behaviour lives here and applies everywhere.
//
// A "*" path segment fans out across the elements of the array at that position
// and continues resolving the remaining path from each element, so the path
// resolves to a SET of candidate values: the positive operators (exists, eq,
// contains, in, matches, gt, gte, lt, lte) match when ANY candidate satisfies
// them, while the negative operator (ne) matches when NO candidate equals the
// operand — the empty set passing vacuously. A path with no wildcard resolves to
// at most one candidate, so these semantics reduce exactly to a single-value
// lookup and stay backward compatible.
func EvaluateCondition(c *StageCondition, root map[string]any) bool {
	if c == nil {
		return true
	}

	candidates, found := ResolveCandidates(root, c.Path)

	switch c.Op {
	case ConditionOpExists:
		return found
	case ConditionOpNe:
		// Universal: every candidate must differ. An empty set passes vacuously.
		for _, actual := range candidates {
			if equalValues(actual, c.Value) {
				return false
			}
		}
		return true
	case ConditionOpEq:
		return anyCandidate(candidates, func(a any) bool { return equalValues(a, c.Value) })
	case ConditionOpContains:
		return anyCandidate(candidates, func(a any) bool { return containsValue(a, c.Value) })
	case ConditionOpIn:
		return anyCandidate(candidates, func(a any) bool { return inValue(a, c.Value) })
	case ConditionOpMatches:
		re, ok := compileMatchPattern(c.Value)
		if !ok {
			return false
		}
		return anyCandidate(candidates, func(a any) bool { return matchesRegex(a, re) })
	case ConditionOpGt:
		return anyCandidate(candidates, func(a any) bool { x, y, ok := numericPair(a, c.Value); return ok && x > y })
	case ConditionOpGte:
		return anyCandidate(candidates, func(a any) bool { x, y, ok := numericPair(a, c.Value); return ok && x >= y })
	case ConditionOpLt:
		return anyCandidate(candidates, func(a any) bool { x, y, ok := numericPair(a, c.Value); return ok && x < y })
	case ConditionOpLte:
		return anyCandidate(candidates, func(a any) bool { x, y, ok := numericPair(a, c.Value); return ok && x <= y })
	default:
		return false
	}
}

// anyCandidate reports whether at least one candidate satisfies pred.
func anyCandidate(candidates []any, pred func(any) bool) bool {
	for _, actual := range candidates {
		if pred(actual) {
			return true
		}
	}
	return false
}

// ResolveCandidates resolves a dot-separated path into the set of values it
// reaches. A "*" segment fans out across every element of the array at that
// position and continues resolving the remaining path from each element. Without
// a wildcard the path yields at most one candidate. The bool reports whether at
// least one candidate was found. It is exported so the hub-workflows engine can
// render the actual values seen at a condition path in its debug explainer while
// still evaluating conditions through this single shared implementation.
func ResolveCandidates(root map[string]any, path string) ([]any, bool) {
	if path == "" {
		return nil, false
	}
	out := resolveParts(root, strings.Split(path, "."))
	return out, len(out) > 0
}

// resolveParts walks the remaining path segments from current, fanning out on a
// "*" segment across array elements. When the segments are exhausted the value
// reached is emitted as a candidate. A non-wildcard segment applied to anything
// that is not a string-keyed map yields no candidate.
func resolveParts(current any, parts []string) []any {
	if len(parts) == 0 {
		return []any{current}
	}
	part, rest := parts[0], parts[1:]

	if part == "*" {
		items, ok := current.([]any)
		if !ok {
			return nil
		}
		var out []any
		for _, item := range items {
			out = append(out, resolveParts(item, rest)...)
		}
		return out
	}

	m, ok := current.(map[string]any)
	if !ok {
		return nil
	}
	next, ok := m[part]
	if !ok {
		return nil
	}
	return resolveParts(next, rest)
}

func equalValues(a, b any) bool {
	if af, aok := toFloat(a); aok {
		if bf, bok := toFloat(b); bok {
			return af == bf
		}
		return false
	}
	return a == b
}

// containsValue is true when actual is a slice containing the wanted value, or a
// string containing the wanted substring.
func containsValue(actual, wanted any) bool {
	switch v := actual.(type) {
	case []any:
		for _, item := range v {
			if equalValues(item, wanted) {
				return true
			}
		}
		return false
	case string:
		if s, ok := wanted.(string); ok {
			return strings.Contains(v, s)
		}
		return false
	default:
		return false
	}
}

// inValue is true when actual equals one of the values in wanted, where wanted
// is the operand list. It is the inverse of containsValue: here the operand is
// the set and actual is the candidate member.
func inValue(actual, wanted any) bool {
	list, ok := wanted.([]any)
	if !ok {
		return false
	}
	for _, item := range list {
		if equalValues(actual, item) {
			return true
		}
	}
	return false
}

// compileMatchPattern compiles the operand of a "matches" condition into a
// regular expression. The operand must be a string holding an RE2 pattern; any
// other type, or a pattern that fails to compile, yields ok=false so the
// operator fails closed (never matches).
func compileMatchPattern(wanted any) (*regexp.Regexp, bool) {
	pattern, ok := wanted.(string)
	if !ok {
		return nil, false
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, false
	}
	return re, true
}

// matchesRegex is true when the pattern matches a string candidate anywhere (a
// partial, unanchored match — anchor with ^…$ for a full match). It mirrors
// containsValue's string/array duality: a bare string candidate matches
// directly, and an array candidate matches when ANY of its string elements
// matches. Non-string, non-array values never match.
func matchesRegex(actual any, re *regexp.Regexp) bool {
	switch v := actual.(type) {
	case string:
		return re.MatchString(v)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && re.MatchString(s) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func numericPair(a, b any) (float64, float64, bool) {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	return af, bf, aok && bok
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}
