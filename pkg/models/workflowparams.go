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
// Params and returns the canonical values to store. Undeclared keys and empty
// values (nil, "" or an empty list) are dropped; ApplyParamDefaults fills them
// from declared defaults. Strings are trimmed (secrets are kept verbatim),
// numbers are float64 within Minimum/Maximum, select values must be one of
// Options, and multiselect values are distinct strings, each one of Options.
// The result is nil when no value remains.
func (s WorkflowStage) NormalizeParamValues(data map[string]any) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	normalized := make(map[string]any, len(data))
	for _, param := range s.Params {
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

// ApplyParamDefaults returns data with every declared, non-secret parameter
// that has no value set to its normalized Default. Existing values are kept.
func (s WorkflowStage) ApplyParamDefaults(data map[string]any) (map[string]any, error) {
	result := make(map[string]any, len(data)+len(s.Params))
	for key, value := range data {
		result[key] = value
	}
	for _, param := range s.Params {
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

// MissingRequiredParams returns, in sorted order, the required parameters that
// have neither a value in data nor a declared default.
func (s WorkflowStage) MissingRequiredParams(data map[string]any) []string {
	var missing []string
	for _, param := range s.Params {
		if !param.Required || param.Default != nil {
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
