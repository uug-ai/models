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
// conditionDeviceFields and conditionUserFields are the curated envelope leaves
// a condition may target; credentials (user.storage) are deliberately absent.
// All three derive from WorkflowConditionRootSchema, the single description.
var conditionDeviceFields, conditionUserFields, conditionOperationFields = conditionSchemaFields()

// conditionSchemaFields reduces the schema to the first field segment under
// each namespace: a field followed by "*" is an array, one with deeper named
// segments is a map, and anything else is a leaf.
func conditionSchemaFields() (map[string]conditionFieldKind, map[string]conditionFieldKind, map[string]map[string]conditionFieldKind) {
	device, user := map[string]conditionFieldKind{}, map[string]conditionFieldKind{}
	operations := map[string]map[string]conditionFieldKind{}
	record := func(fields map[string]conditionFieldKind, rest string) {
		field, remainder, deeper := strings.Cut(rest, ".")
		kind := conditionFieldLeaf
		if deeper {
			kind = conditionFieldMap
			if next, _, _ := strings.Cut(remainder, "."); next == "*" {
				kind = conditionFieldArray
			}
		}
		if previous, seen := fields[field]; !seen || previous == conditionFieldLeaf {
			fields[field] = kind
		}
	}
	for _, entry := range WorkflowConditionRootSchema() {
		root, rest, nested := strings.Cut(entry.Path, ".")
		if !nested || entry.Open {
			continue
		}
		switch root {
		case "device":
			record(device, rest)
		case "user":
			record(user, rest)
		case "inputs", "results":
			op, fieldPath, _ := strings.Cut(rest, ".")
			if operations[op] == nil {
				operations[op] = map[string]conditionFieldKind{}
			}
			record(operations[op], fieldPath)
		}
	}
	return device, user, operations
}

// ValidateStageCondition is the legacy entrypoint for shared validation.
func ValidateStageCondition(c *StageCondition) error {
	return ValidateWorkflowCondition(c)
}

// ValidateWorkflowCondition checks the credential-free path, operator, operand,
// and bounded logical/anyMatch structure. A nil condition remains unconditional.
func ValidateWorkflowCondition(c *WorkflowCondition) error {
	remaining := MaxWorkflowConditionCount
	return validateWorkflowConditionTree(c, 0, &remaining, true)
}

// Logical groups are bounded independently of anyMatch's nonrecursive,
// same-element predicate set. The count includes groups and relative leaves.
const (
	MaxWorkflowConditionDepth = 8
	MaxWorkflowConditionCount = 256
)

func validateWorkflowConditionTree(c *WorkflowCondition, depth int, remaining *int, full bool) error {
	if c == nil {
		return nil
	}
	*remaining -= 1
	if *remaining < 0 {
		return fmt.Errorf("condition set exceeds %d predicates", MaxWorkflowConditionCount)
	}
	if c.Op == ConditionOpAll || c.Op == ConditionOpAny {
		if depth >= MaxWorkflowConditionDepth {
			return fmt.Errorf("logical conditions exceed depth %d", MaxWorkflowConditionDepth)
		}
		if len(c.Conditions) == 0 {
			return fmt.Errorf("condition op %q requires nonempty conditions", c.Op)
		}
		if c.Path != "" || c.Value != nil || c.Match != nil {
			return fmt.Errorf("condition op %q uses only conditions, not path, value or match", c.Op)
		}
		for i := range c.Conditions {
			if err := validateWorkflowConditionTree(&c.Conditions[i], depth+1, remaining, full); err != nil {
				return fmt.Errorf("condition op %q child %d: %w", c.Op, i, err)
			}
		}
		return nil
	}
	if c.Conditions != nil {
		return fmt.Errorf("condition %q op %q cannot have logical conditions", c.Path, c.Op)
	}
	if full {
		if err := validateConditionPath(c); err != nil {
			return err
		}
	}
	if c.Op == ConditionOpAnyMatch && c.Match != nil {
		*remaining -= len(c.Match.Conditions)
		if *remaining < 0 {
			return fmt.Errorf("condition set exceeds %d predicates", MaxWorkflowConditionCount)
		}
	}
	if err := validateWorkflowLeafStructure(c); err != nil {
		return err
	}
	if c.Op == ConditionOpAnyMatch {
		if !full {
			return nil
		}
		for i, p := range c.Match.Conditions {
			if err := validateConditionValue(&WorkflowCondition{Path: p.Path, Op: p.Op, Value: p.Value}); err != nil {
				return fmt.Errorf("condition %q match predicate %d: %w", c.Path, i, err)
			}
		}
		return nil
	}
	if !full {
		return nil
	}
	return validateConditionValue(c)
}

func ValidateWorkflowConditionSet(set WorkflowConditionSet) error {
	return validateWorkflowConditionSet(set, true)
}

func validateWorkflowConditionSet(set WorkflowConditionSet, full bool) error {
	if err := validateConditionMode(set.ConditionMode); err != nil {
		return err
	}
	remaining := MaxWorkflowConditionCount
	for i := range set.Conditions {
		if err := validateWorkflowConditionTree(&set.Conditions[i], 0, &remaining, full); err != nil {
			return fmt.Errorf("condition %d: %w", i, err)
		}
	}
	return nil
}

// Structural checks also run before evaluation can short-circuit. Operand
// validation (including regex compilation) remains an authoring/load concern.
func validateWorkflowConditionStructure(c *WorkflowCondition) error {
	remaining := MaxWorkflowConditionCount
	return validateWorkflowConditionTree(c, 0, &remaining, false)
}

func validateWorkflowLeafStructure(c *WorkflowCondition) error {
	if c.Op != ConditionOpAnyMatch {
		if c.Match != nil {
			return fmt.Errorf("condition %q op %q cannot have match predicates", c.Path, c.Op)
		}
		if !isScalarConditionOp(c.Op) {
			return fmt.Errorf("condition %q has unknown condition operator %q", c.Path, c.Op)
		}
		return nil
	}
	if c.Value != nil {
		return fmt.Errorf("condition %q op anyMatch uses match, not value", c.Path)
	}
	if c.Match == nil || len(c.Match.Conditions) == 0 {
		return fmt.Errorf("condition %q op anyMatch requires nonempty match predicates", c.Path)
	}
	if err := validateConditionMode(c.Match.ConditionMode); err != nil {
		return fmt.Errorf("condition %q match: %w", c.Path, err)
	}
	if err := validateConditionPath(c); err != nil {
		return err
	}
	if err := validateAnyMatchPath(c.Path); err != nil {
		return err
	}
	for i, p := range c.Match.Conditions {
		if !isScalarConditionOp(p.Op) {
			return fmt.Errorf("condition %q match predicate %d has unsupported operator %q; nested anyMatch is not allowed", c.Path, i, p.Op)
		}
		if err := validateRelativeConditionPath(p.Path); err != nil {
			return fmt.Errorf("condition %q match predicate %d: %w", c.Path, i, err)
		}
	}
	return nil
}

func isScalarConditionOp(op ConditionOp) bool {
	switch op {
	case ConditionOpEq, ConditionOpNe, ConditionOpContains, ConditionOpIn, ConditionOpExists,
		ConditionOpMatches, ConditionOpGt, ConditionOpGte, ConditionOpLt, ConditionOpLte:
		return true
	default:
		return false
	}
}

func validateConditionPath(c *StageCondition) error {
	if c == nil {
		return nil
	}
	path := c.Path
	if err := validateConditionPathSegments(path); err != nil {
		return err
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
	if c == nil {
		return nil
	}
	switch c.Op {
	case ConditionOpExists:
		return nil // Legacy operands are intentionally ignored.
	case ConditionOpEq, ConditionOpNe, ConditionOpContains:
		if !isConditionScalar(c.Value) {
			return fmt.Errorf("condition %q op %q requires a scalar value, got %T", c.Path, c.Op, c.Value)
		}
	case ConditionOpIn:
		items, ok := asList(c.Value)
		if !ok {
			return fmt.Errorf("condition %q op in requires a list value, got %T", c.Path, c.Value)
		}
		for _, item := range items {
			if !isConditionScalar(item) {
				return fmt.Errorf("condition %q op in requires scalar list elements, got %T", c.Path, item)
			}
		}
	case ConditionOpGt, ConditionOpGte, ConditionOpLt, ConditionOpLte:
		if _, ok := toFloat(c.Value); !ok {
			return fmt.Errorf("condition %q op %q requires a numeric value, got %T", c.Path, c.Op, c.Value)
		}
	case ConditionOpMatches:
		pattern, ok := c.Value.(string)
		if !ok {
			return fmt.Errorf("condition %q op %q requires a string regular-expression value, got %T", c.Path, c.Op, c.Value)
		}
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("condition %q op %q has an invalid regular expression %q: %w", c.Path, c.Op, pattern, err)
		}
	default:
		return fmt.Errorf("condition %q has unknown condition operator %q", c.Path, c.Op)
	}
	return nil
}

func isConditionScalar(v any) bool {
	switch v.(type) {
	case nil, string, bool:
		return true
	default:
		_, ok := toFloat(v)
		return ok
	}
}

func validateConditionPathSegments(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("condition has an empty path")
	}
	for _, seg := range strings.Split(path, ".") {
		if seg == "" {
			return fmt.Errorf("condition path %q has an empty segment (check for a leading, trailing, or doubled '.')", path)
		}
		if seg != strings.TrimSpace(seg) || strings.ContainsAny(seg, "$[]") || (strings.Contains(seg, "*") && seg != "*") {
			return fmt.Errorf("condition path %q has an invalid segment %q", path, seg)
		}
	}
	return nil
}

func validateRelativeConditionPath(path string) error {
	if err := validateConditionPathSegments(path); err != nil {
		return err
	}
	root, _, _ := strings.Cut(path, ".")
	switch root {
	case "*", "inputs", "results", "device", "user", "storage":
		return fmt.Errorf("match path %q must be relative to an array object, not an envelope namespace or credentials", path)
	}
	for _, part := range strings.Split(path, ".") {
		if strings.Trim(part, "0123456789") == "" {
			return fmt.Errorf("match path %q uses a numeric array index; use a wildcard instead", path)
		}
	}
	return nil
}

func validateAnyMatchPath(path string) error {
	parts := strings.Split(path, ".")
	switch parts[0] {
	case "operation", "runId", "key", "traceId", "user":
		return fmt.Errorf("condition path %q is scalar, not an anyMatch array root", path)
	case "device":
		if len(parts) != 2 || conditionDeviceFields[parts[1]] != conditionFieldArray {
			return fmt.Errorf("condition path %q is scalar, not an anyMatch array root", path)
		}
	case "inputs", "results":
		if len(parts) == 2 {
			return fmt.Errorf("condition path %q is an operation object, not an anyMatch array root", path)
		}
		if len(parts) == 3 {
			if fields, known := conditionOperationFields[parts[1]]; known && fields[parts[2]] != conditionFieldArray {
				return fmt.Errorf("condition path %q is not an anyMatch array root", path)
			}
		}
	}
	return nil
}

func validateConditionLeafNamespace(path, root, rest string, nested bool, fields map[string]conditionFieldKind) error {
	if !nested {
		return fmt.Errorf("condition path %q must name a field under %q (e.g. %s.<field>)", path, root, root)
	}
	field, remainder, deeper := strings.Cut(rest, ".")
	kind, known := fields[field]
	if !known {
		return fmt.Errorf("condition path %q targets unknown field %q under %q", path, field, root)
	}
	if deeper {
		if kind == conditionFieldArray && remainder == "*" {
			return nil
		}
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
