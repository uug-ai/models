package models

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// conditionFieldKind describes how deep a condition path may reach into a known
// operation's result field.
type conditionFieldKind int

const (
	conditionFieldLeaf  conditionFieldKind = iota // scalar: matchable only at this exact segment
	conditionFieldArray                           // array: matchable directly, or fanned out with a "*" segment
	conditionFieldMap                             // nested map: deeper paths allowed
)

// conditionOperationFields declares the result fields a condition may target
// under inputs.<op> / results.<op> for operations whose shape the platform owns.
// Operations absent here carry worker-defined results and accept any field path.
var conditionOperationFields = map[string]map[string]conditionFieldKind{
	"classify": {
		"properties":  conditionFieldArray,
		"objectCount": conditionFieldLeaf,
		"details":     conditionFieldArray,
	},
}

// conditionDeviceFields and conditionUserFields are the curated envelope leaves
// a condition may target; credentials (user.storage) are deliberately absent.
var (
	conditionDeviceFields = map[string]bool{"deviceKey": true, "deviceName": true, "provider": true, "storageSolution": true, "siteIds": true}
	conditionUserFields   = map[string]bool{"organisationId": true}
)

// ValidateStageCondition checks that a condition's path is reachable on the
// credential-free run root (see StageCondition) and that a `matches` operand is a
// valid RE2 pattern. A nil condition (unconditional dependency) is valid. It is
// the shared validator for stage needs, trigger conditions, and graph edges.
func ValidateStageCondition(c *StageCondition) error {
	if err := validateConditionPath(c); err != nil {
		return err
	}
	return validateConditionValue(c)
}

func validateConditionPath(c *StageCondition) error {
	if c == nil {
		return nil
	}
	path := strings.TrimSpace(c.Path)
	if path == "" {
		return fmt.Errorf("condition has an empty path")
	}
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			return fmt.Errorf("condition path %q has an empty segment (check for a leading, trailing, or doubled '.')", path)
		}
	}

	root, rest, nested := strings.Cut(path, ".")
	switch root {
	case "storage":
		return fmt.Errorf("condition path %q targets storage credentials, which are not matchable", path)
	case "operation", "runId", "key", "traceId":
		if nested {
			return fmt.Errorf("condition path %q reaches into scalar %q, which cannot be traversed", path, root)
		}
		return nil
	case "device":
		return validateConditionLeafNamespace(path, root, rest, nested, conditionDeviceFields)
	case "user":
		if rest == "storage" || strings.HasPrefix(rest, "storage.") {
			return fmt.Errorf("condition path %q targets user.storage credentials, which are not matchable", path)
		}
		return validateConditionLeafNamespace(path, root, rest, nested, conditionUserFields)
	case "inputs", "results":
		return validateConditionOperationNamespace(path, root, rest, nested)
	default:
		return fmt.Errorf("condition path %q has unknown root %q (want one of: device, inputs, key, operation, results, runId, traceId, user)", path, root)
	}
}

func validateConditionValue(c *StageCondition) error {
	if c == nil || c.Op != ConditionOpMatches {
		return nil
	}
	pattern, ok := c.Value.(string)
	if !ok {
		return fmt.Errorf("condition %q op %q requires a string regular-expression value, got %T", c.Path, c.Op, c.Value)
	}
	if _, err := regexp.Compile(pattern); err != nil {
		return fmt.Errorf("condition %q op %q has an invalid regular expression %q: %w", c.Path, c.Op, pattern, err)
	}
	return nil
}

func validateConditionLeafNamespace(path, root, rest string, nested bool, fields map[string]bool) error {
	if !nested {
		return fmt.Errorf("condition path %q must name a field under %q (e.g. %s.<field>)", path, root, root)
	}
	field, _, deeper := strings.Cut(rest, ".")
	if !fields[field] {
		return fmt.Errorf("condition path %q targets unknown field %q under %q", path, field, root)
	}
	if deeper {
		return fmt.Errorf("condition path %q reaches past leaf %q.%q, which cannot be traversed", path, root, field)
	}
	return nil
}

func validateConditionOperationNamespace(path, root, rest string, nested bool) error {
	if !nested {
		return fmt.Errorf("condition path %q must name an operation under %q (e.g. %s.classify.<field>)", path, root, root)
	}
	op, fieldPath, hasField := strings.Cut(rest, ".")
	if op == "" {
		return fmt.Errorf("condition path %q has an empty operation under %q", path, root)
	}
	if !hasField {
		return nil
	}
	fields, known := conditionOperationFields[op]
	if !known {
		return nil
	}
	fieldRoot, remainder, deeper := strings.Cut(fieldPath, ".")
	kind, ok := fields[fieldRoot]
	if !ok {
		return fmt.Errorf("condition path %q targets unknown field %q on operation %q (known fields: %s)", path, fieldRoot, op, knownConditionFieldList(fields))
	}
	if !deeper {
		return nil
	}
	switch kind {
	case conditionFieldMap:
		return nil
	case conditionFieldArray:
		if next, _, _ := strings.Cut(remainder, "."); next != "*" {
			return fmt.Errorf("condition path %q reaches into array %q on operation %q without a %q fan-out segment; use %s.%s.%s.*.<field> to match across its elements", path, fieldRoot, op, "*", root, op, fieldRoot)
		}
		return nil
	default:
		return fmt.Errorf("condition path %q reaches into scalar %q on operation %q, which cannot be traversed", path, fieldRoot, op)
	}
}

func knownConditionFieldList(fields map[string]conditionFieldKind) string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
