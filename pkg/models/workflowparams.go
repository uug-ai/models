package models

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
)

// MaxStageParamStringLength bounds a single string or secret parameter value.
const MaxStageParamStringLength = 4096

// Param returns the declared parameter with the given name.
func (s WorkflowStage) Param(name string) (StageParam, bool) {
	for _, param := range s.Params {
		if param.Name == name {
			return param, true
		}
	}
	return StageParam{}, false
}

// IsSecretParam reports whether name is declared as a StageParamSecret.
func (s WorkflowStage) IsSecretParam(name string) bool {
	param, ok := s.Param(name)
	return ok && param.Type == StageParamSecret
}

// NormalizeParamValues validates a node's Data against the stage's declared
// Params and returns the canonical values to store. Undeclared keys, values of
// inactive params (see StageParam.When) and empty values (nil, "" or an empty
// list) are dropped; ApplyParamDefaults fills them from declared defaults.
// Inactive params are not validated. Strings are trimmed (secrets are kept
// verbatim), numbers are float64 within Minimum/Maximum, select values must be
// one of Options, and multiselect values are distinct strings, each one of
// Options. The result is nil when no value remains.
func (s WorkflowStage) NormalizeParamValues(data map[string]any) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	active := s.activeParams(data)
	normalized := make(map[string]any, len(data))
	for _, param := range s.Params {
		if !active[param.Name] {
			continue
		}
		raw, ok := data[param.Name]
		if !ok || raw == nil {
			continue
		}
		value, keep, err := normalizeParamValue(param, raw)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", param.Name, err)
		}
		if keep {
			normalized[param.Name] = value
		}
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}

// ApplyParamDefaults returns data with every active, declared, non-secret
// parameter that has no value set to its normalized Default. Existing values
// of active params and undeclared keys are kept; values of declared params
// that are inactive for the result (see StageParam.When) are removed, so a
// value restored or kept from an earlier save cannot outlive its condition.
func (s WorkflowStage) ApplyParamDefaults(data map[string]any) (map[string]any, error) {
	result := make(map[string]any, len(data)+len(s.Params))
	for key, value := range data {
		result[key] = value
	}
	active := s.activeParams(data)
	for _, param := range s.Params {
		if !active[param.Name] {
			delete(result, param.Name)
			continue
		}
		if param.Default == nil || param.Type == StageParamSecret {
			continue
		}
		if value, ok := result[param.Name]; ok && value != nil {
			continue
		}
		value, keep, err := normalizeParamValue(param, param.Default)
		if err != nil {
			return nil, fmt.Errorf("parameter %q default: %w", param.Name, err)
		}
		if keep {
			result[param.Name] = value
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

// MissingRequiredParams returns, in sorted order, the required active
// parameters that have neither a value in data nor a declared default.
func (s WorkflowStage) MissingRequiredParams(data map[string]any) []string {
	active := s.activeParams(data)
	var missing []string
	for _, param := range s.Params {
		if !param.Required || param.Default != nil || !active[param.Name] {
			continue
		}
		value, ok := data[param.Name]
		if !ok || value == nil {
			missing = append(missing, param.Name)
			continue
		}
		if text, isText := value.(string); isText && strings.TrimSpace(text) == "" {
			missing = append(missing, param.Name)
		}
	}
	sort.Strings(missing)
	return missing
}

// MaxStageParamConditionDepth bounds a chain of conditional params (a param
// whose condition refers to a param that is itself conditional, and so on).
const MaxStageParamConditionDepth = 3

// ActiveParams returns, in declaration order, the params that apply to a node
// with the given Data: unconditional params, plus conditional params whose
// controlling param is active and whose value, or default when unset, is one
// of the condition's values.
func (s WorkflowStage) ActiveParams(data map[string]any) []StageParam {
	active := s.activeParams(data)
	result := make([]StageParam, 0, len(s.Params))
	for _, param := range s.Params {
		if active[param.Name] {
			result = append(result, param)
		}
	}
	return result
}

// activeParams evaluates every condition in declaration order, so a condition
// can only see params declared before it. A condition that refers to an
// unknown, later or inactive param never holds. A controlling value that does
// not normalize counts as unset and falls back to its default.
func (s WorkflowStage) activeParams(data map[string]any) map[string]bool {
	active := make(map[string]bool, len(s.Params))
	effective := make(map[string]any, len(s.Params))
	for _, param := range s.Params {
		on := param.When == nil
		if !on {
			controller := param.When.Param
			on = active[controller] && conditionHolds(param.When.In, effective[controller])
		}
		active[param.Name] = on
		if !on || (param.Type != StageParamSelect && param.Type != StageParamBoolean) {
			continue
		}
		if raw, ok := data[param.Name]; ok && raw != nil {
			if value, keep, err := normalizeParamValue(param, raw); err == nil && keep {
				effective[param.Name] = value
				continue
			}
		}
		if param.Default != nil {
			if value, keep, err := normalizeParamValue(param, param.Default); err == nil && keep {
				effective[param.Name] = value
			}
		}
	}
	return active
}

func conditionHolds(in []any, value any) bool {
	if value == nil {
		return false
	}
	for _, candidate := range in {
		switch candidate.(type) {
		case string, bool:
			if candidate == value {
				return true
			}
		}
	}
	return false
}

// ValidateParamConditions checks every StageParam.When: the controlling param
// is declared earlier and is a select or boolean, the condition lists at least
// one value, every value is one of the select's options (any string when it
// declares none) or a boolean, and no chain is deeper than
// MaxStageParamConditionDepth.
func (s WorkflowStage) ValidateParamConditions() error {
	declared := make(map[string]StageParam, len(s.Params))
	depth := make(map[string]int, len(s.Params))
	for _, param := range s.Params {
		if when := param.When; when != nil {
			controller, ok := declared[when.Param]
			if !ok {
				return fmt.Errorf("parameter %q: condition refers to %q, which is not declared before it", param.Name, when.Param)
			}
			if controller.Type != StageParamSelect && controller.Type != StageParamBoolean {
				return fmt.Errorf("parameter %q: condition refers to %q, which is not a select or boolean", param.Name, when.Param)
			}
			if len(when.In) == 0 {
				return fmt.Errorf("parameter %q: condition lists no values", param.Name)
			}
			for _, value := range when.In {
				switch typed := value.(type) {
				case string:
					if controller.Type != StageParamSelect || (len(controller.Options) > 0 && !slices.Contains(controller.Options, typed)) {
						return fmt.Errorf("parameter %q: condition value %q is not an option of %q", param.Name, typed, when.Param)
					}
				case bool:
					if controller.Type != StageParamBoolean {
						return fmt.Errorf("parameter %q: condition value %v does not match select %q", param.Name, typed, when.Param)
					}
				default:
					return fmt.Errorf("parameter %q: condition values must be strings or booleans", param.Name)
				}
			}
			depth[param.Name] = depth[when.Param] + 1
			if depth[param.Name] > MaxStageParamConditionDepth {
				return fmt.Errorf("parameter %q: conditions are nested deeper than %d", param.Name, MaxStageParamConditionDepth)
			}
		}
		declared[param.Name] = param
	}
	return nil
}

func normalizeParamValue(param StageParam, raw any) (any, bool, error) {
	switch param.Type {
	case StageParamString, StageParamSecret, StageParamSelect:
		text, ok := raw.(string)
		if !ok {
			return nil, false, fmt.Errorf("must be a string")
		}
		if param.Type != StageParamSecret {
			text = strings.TrimSpace(text)
		}
		if text == "" {
			return nil, false, nil
		}
		if len(text) > MaxStageParamStringLength {
			return nil, false, fmt.Errorf("must be at most %d characters", MaxStageParamStringLength)
		}
		if param.Type == StageParamSelect && len(param.Options) > 0 && !slices.Contains(param.Options, text) {
			return nil, false, fmt.Errorf("must be one of %s", strings.Join(param.Options, ", "))
		}
		return text, true, nil
	case StageParamNumber:
		number, ok := paramNumber(raw)
		if !ok {
			return nil, false, fmt.Errorf("must be a finite number")
		}
		if param.Minimum != nil && number < *param.Minimum {
			return nil, false, fmt.Errorf("must be at least %v", *param.Minimum)
		}
		if param.Maximum != nil && number > *param.Maximum {
			return nil, false, fmt.Errorf("must be at most %v", *param.Maximum)
		}
		return number, true, nil
	case StageParamMultiSelect:
		items, ok := paramStrings(raw)
		if !ok {
			return nil, false, fmt.Errorf("must be a list of strings")
		}
		values := make([]string, 0, len(items))
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item == "" || slices.Contains(values, item) {
				continue
			}
			if len(param.Options) > 0 && !slices.Contains(param.Options, item) {
				return nil, false, fmt.Errorf("values must be among %s", strings.Join(param.Options, ", "))
			}
			values = append(values, item)
		}
		if len(values) == 0 {
			return nil, false, nil
		}
		return values, true, nil
	case StageParamBoolean:
		flag, ok := raw.(bool)
		if !ok {
			return nil, false, fmt.Errorf("must be a boolean")
		}
		return flag, true, nil
	default:
		return nil, false, fmt.Errorf("has unsupported type %q", param.Type)
	}
}

func paramStrings(raw any) ([]string, bool) {
	switch value := raw.(type) {
	case []string:
		return value, true
	case []any:
		items := make([]string, 0, len(value))
		for _, item := range value {
			text, ok := item.(string)
			if !ok {
				return nil, false
			}
			items = append(items, text)
		}
		return items, true
	}
	// BSON arrays decode as primitive.A, a named []interface{}.
	if list, ok := asList(raw); ok {
		return paramStrings(list)
	}
	return nil, false
}

func paramNumber(raw any) (float64, bool) {
	var number float64
	switch value := raw.(type) {
	case float64:
		number = value
	case float32:
		number = float64(value)
	case int:
		number = float64(value)
	case int32:
		number = float64(value)
	case int64:
		number = float64(value)
	default:
		return 0, false
	}
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, false
	}
	return number, true
}
