package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"gopkg.in/yaml.v3"
)

// LimitUnlimitedValue is how an unlimited Limit is written in JSON, YAML and
// BSON.
const LimitUnlimitedValue = "unlimited"

// ErrLimitInvalid reports a limit that is neither a non-negative whole
// number nor "unlimited".
var ErrLimitInvalid = errors.New("invalid limit")

// Limit is one plan limit: a non-negative number, or unlimited. It is written
// as a number or as the string "unlimited" in JSON, YAML and BSON; there is
// no sentinel number. 0 means none. Check usage with Allows instead of
// comparing numbers.
type Limit struct {
	// Unlimited is set when there is no limit; Value is then ignored.
	Unlimited bool `json:"-" bson:"-" yaml:"-"`
	// Value is the limit when Unlimited is not set.
	Value int64 `json:"-" bson:"-" yaml:"-"`
}

// NewLimit returns a limit of n, for filling Limits.
func NewLimit(n int64) *Limit {
	return &Limit{Value: n}
}

// NewUnlimitedLimit returns an unlimited limit, for filling Limits.
func NewUnlimitedLimit() *Limit {
	return &Limit{Unlimited: true}
}

// Allows reports whether a usage of used stays within the limit: always when
// unlimited, otherwise when used is at most the value. To check whether n
// more fit, pass the current usage plus n.
func (l Limit) Allows(used int64) bool {
	return l.Unlimited || used <= l.Value
}

// String returns "unlimited" or the number.
func (l Limit) String() string {
	if l.Unlimited {
		return LimitUnlimitedValue
	}
	return strconv.FormatInt(l.Value, 10)
}

// MarshalJSON writes the limit as a number or "unlimited".
func (l Limit) MarshalJSON() ([]byte, error) {
	if l.Unlimited {
		return json.Marshal(LimitUnlimitedValue)
	}
	if l.Value < 0 {
		return nil, fmt.Errorf("%w: %d", ErrLimitInvalid, l.Value)
	}
	return []byte(strconv.FormatInt(l.Value, 10)), nil
}

// UnmarshalJSON reads a non-negative whole number or "unlimited".
func (l *Limit) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		return l.setText(text)
	}
	value, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrLimitInvalid, data)
	}
	return l.setValue(value)
}

// MarshalYAML writes the limit as a number or "unlimited".
func (l Limit) MarshalYAML() (interface{}, error) {
	if l.Unlimited {
		return LimitUnlimitedValue, nil
	}
	if l.Value < 0 {
		return nil, fmt.Errorf("%w: %d", ErrLimitInvalid, l.Value)
	}
	return l.Value, nil
}

// UnmarshalYAML reads a non-negative whole number or "unlimited".
func (l *Limit) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("%w at line %d: want a number or %q", ErrLimitInvalid, node.Line, LimitUnlimitedValue)
	}
	if node.ShortTag() == "!!null" {
		return nil
	}
	if node.Value == LimitUnlimitedValue {
		*l = Limit{Unlimited: true}
		return nil
	}
	if node.ShortTag() != "!!int" {
		return fmt.Errorf("%w at line %d: %q", ErrLimitInvalid, node.Line, node.Value)
	}
	var value int64
	if err := node.Decode(&value); err != nil {
		return fmt.Errorf("%w at line %d: %q", ErrLimitInvalid, node.Line, node.Value)
	}
	return l.setValue(value)
}

// MarshalBSONValue writes the limit as an int64 or "unlimited".
func (l Limit) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if l.Unlimited {
		return bson.MarshalValue(LimitUnlimitedValue)
	}
	if l.Value < 0 {
		return 0, nil, fmt.Errorf("%w: %d", ErrLimitInvalid, l.Value)
	}
	return bson.MarshalValue(l.Value)
}

// UnmarshalBSONValue reads an int32, int64, a whole double (as written by
// the mongo shell) or "unlimited".
func (l *Limit) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	raw := bson.RawValue{Type: t, Value: data}
	switch t {
	case bsontype.Null, bsontype.Undefined:
		return nil
	case bsontype.String:
		text, ok := raw.StringValueOK()
		if !ok {
			return fmt.Errorf("%w: malformed string", ErrLimitInvalid)
		}
		return l.setText(text)
	case bsontype.Int32:
		value, ok := raw.Int32OK()
		if !ok {
			return fmt.Errorf("%w: malformed int32", ErrLimitInvalid)
		}
		return l.setValue(int64(value))
	case bsontype.Int64:
		value, ok := raw.Int64OK()
		if !ok {
			return fmt.Errorf("%w: malformed int64", ErrLimitInvalid)
		}
		return l.setValue(value)
	case bsontype.Double:
		value, ok := raw.DoubleOK()
		if !ok || value != math.Trunc(value) || value > math.MaxInt64 || value < math.MinInt64 {
			return fmt.Errorf("%w: %v", ErrLimitInvalid, value)
		}
		return l.setValue(int64(value))
	}
	return fmt.Errorf("%w: BSON type %s", ErrLimitInvalid, t)
}

func (l *Limit) setText(text string) error {
	if text != LimitUnlimitedValue {
		return fmt.Errorf("%w: %q", ErrLimitInvalid, text)
	}
	*l = Limit{Unlimited: true}
	return nil
}

func (l *Limit) setValue(value int64) error {
	if value < 0 {
		return fmt.Errorf("%w: %d", ErrLimitInvalid, value)
	}
	*l = Limit{Value: value}
	return nil
}

// Limits are the numbers a plan allows, separate from its permissions. A nil
// field is not set: in a plan every field is required (PlanDefinition
// validation); in subscription overrides a nil field means "use the plan's
// value" and a set one replaces it, including with unlimited.
type Limits struct {
	// RetentionDays is how long recordings are kept. Replaces dayLimit and
	// custom_day_limit.
	RetentionDays *Limit `json:"retentionDays,omitempty" bson:"retentionDays,omitempty" yaml:"retentionDays,omitempty"`
	// UploadGBPerDay caps recording uploads per calendar day in the
	// organisation's timezone. Replaces usage.
	UploadGBPerDay *Limit `json:"uploadGBPerDay,omitempty" bson:"uploadGBPerDay,omitempty" yaml:"uploadGBPerDay,omitempty"`
	// UploadGBPerMonth caps recording uploads per calendar month in the
	// organisation's timezone. Replaces usage.
	UploadGBPerMonth *Limit `json:"uploadGBPerMonth,omitempty" bson:"uploadGBPerMonth,omitempty" yaml:"uploadGBPerMonth,omitempty"`
	// WorkflowCredits are the custom workflow runs per month (1 credit per
	// completed run).
	WorkflowCredits *Limit `json:"workflowCredits,omitempty" bson:"workflowCredits,omitempty" yaml:"workflowCredits,omitempty"`
	// Devices is the number of active devices.
	Devices *Limit `json:"devices,omitempty" bson:"devices,omitempty" yaml:"devices,omitempty"`
	// Projects is the number of projects.
	Projects *Limit `json:"projects,omitempty" bson:"projects,omitempty" yaml:"projects,omitempty"`
	// Members is the number of organisation members. Replaces
	// organisation.settings.maxMembers.
	Members *Limit `json:"members,omitempty" bson:"members,omitempty" yaml:"members,omitempty"`
}

// limitFields lists every limit by its JSON name, in declaration order.
func (l *Limits) limitFields() []struct {
	name  string
	field **Limit
} {
	return []struct {
		name  string
		field **Limit
	}{
		{"retentionDays", &l.RetentionDays},
		{"uploadGBPerDay", &l.UploadGBPerDay},
		{"uploadGBPerMonth", &l.UploadGBPerMonth},
		{"workflowCredits", &l.WorkflowCredits},
		{"devices", &l.Devices},
		{"projects", &l.Projects},
		{"members", &l.Members},
	}
}

// LimitNames returns the JSON names of every limit in declaration order.
func LimitNames() []string {
	var limits Limits
	var names []string
	for _, field := range limits.limitFields() {
		names = append(names, field.name)
	}
	return names
}

// Missing returns the JSON names of the limits that are not set.
func (l Limits) Missing() []string {
	var missing []string
	for _, field := range l.limitFields() {
		if *field.field == nil {
			missing = append(missing, field.name)
		}
	}
	return missing
}

// WithOverrides returns these limits with every limit set in overrides
// replacing its value; a limit not set in overrides keeps this one's. The
// result shares no pointers with either input.
func (l Limits) WithOverrides(overrides Limits) Limits {
	merged := l.Clone()
	mergedFields := merged.limitFields()
	for index, field := range overrides.limitFields() {
		if *field.field != nil {
			value := **field.field
			*mergedFields[index].field = &value
		}
	}
	return merged
}

// Clone returns a copy that shares no pointers with l.
func (l Limits) Clone() Limits {
	var clone Limits
	cloneFields := clone.limitFields()
	for index, field := range l.limitFields() {
		if *field.field != nil {
			value := **field.field
			*cloneFields[index].field = &value
		}
	}
	return clone
}
