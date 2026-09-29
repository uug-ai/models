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
// values (nil or "") are dropped, so an omitted value falls back to the
// parameter's default at runtime. Strings are trimmed (secrets are kept
// verbatim), numbers are float64, and select values must be one of Options.
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
		return number, true, nil
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
