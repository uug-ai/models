package models

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

// PlanDefinition is one version of a plan: which gated permissions it allows
// and its limits. Plans are fixed in the Hub chart's plan file (LoadPlans), not
// stored in the database; a subscription is an instance of one
// {key, version}. A change to permissions or limits adds a new version, and
// every version stays in the file while a subscription references it.
//
// It is named PlanDefinition rather than Plan because Plan is the legacy plan
// (levels and upload counts) that legacy organisations still use.
type PlanDefinition struct {
	// Key is the stable identifier subscriptions reference, e.g. "trial",
	// "community", "enterprise", or an archived legacy key.
	Key string `json:"key" yaml:"key"`
	// Version increments on any change to permissions or limits.
	Version int `json:"version" yaml:"version"`
	// Name is the display name.
	Name string `json:"name" yaml:"name"`
	// Permissions are the gated permission patterns the plan allows: a
	// "domain.*" pattern or a single "domain.action" (see
	// MatchPlanPermissionPattern). Ungated permissions are allowed on every
	// plan and are not listed.
	Permissions []string `json:"permissions" yaml:"permissions"`
	// Limits are the plan's limits; every limit is required.
	Limits Limits `json:"limits" yaml:"limits"`
	// Level is kept only for clients that still compare levels during the
	// transition. It is optional and not authoritative.
	Level int `json:"level,omitempty" yaml:"level,omitempty"`
	// StripeLookupKeys are the Stripe price lookup keys that map to this plan
	// (e.g. "community_monthly"). The same keys exist in Stripe test and live
	// mode, so the plan file is identical on staging and production. Empty
	// for plans that are never billed through Stripe.
	StripeLookupKeys []string `json:"stripeLookupKeys,omitempty" yaml:"stripeLookupKeys,omitempty"`
	// Public plans are offered in self-service checkout.
	Public bool `json:"public" yaml:"public"`
	// Archived plans take no new subscriptions; existing ones keep working.
	Archived bool `json:"archived" yaml:"archived"`
}

func (p PlanDefinition) clone() PlanDefinition {
	p.Permissions = slices.Clone(p.Permissions)
	p.StripeLookupKeys = slices.Clone(p.StripeLookupKeys)
	p.Limits = p.Limits.Clone()
	return p
}

// Validate checks one plan on its own: a key, a positive version, every
// limit set, and every permission pattern covering at least one gated
// permission.
func (p PlanDefinition) Validate() error {
	if p.Key == "" {
		return fmt.Errorf("%w: plan without key", ErrPlanInvalid)
	}
	if p.Version < 1 {
		return fmt.Errorf("%w: plan %q has version %d, want 1 or more", ErrPlanInvalid, p.Key, p.Version)
	}
	if missing := p.Limits.Missing(); len(missing) > 0 {
		return fmt.Errorf("%w: plan %s misses limits %v", ErrPlanInvalid, p.ref(), missing)
	}
	if err := ValidatePlanPermissionPatterns(p.Permissions); err != nil {
		return fmt.Errorf("%w: plan %s: %w", ErrPlanInvalid, p.ref(), err)
	}
	return nil
}

func (p PlanDefinition) ref() string {
	return fmt.Sprintf("%s v%d", p.Key, p.Version)
}

// PlanFile is the plan file shipped with the Hub chart (files/plans.yaml):
//
//	plans:
//	  - key: community
//	    version: 1
//	    name: Community
//	    permissions: [live.*, ptz.*, media.playHD]
//	    limits: {devices: 30, uploadGBPerDay: 10, uploadGBPerMonth: 300,
//	             retentionDays: 30, workflowCredits: 100,
//	             projects: unlimited, members: unlimited}
//	    stripeLookupKeys: [community_monthly]
//	    public: true
//	    archived: false
type PlanFile struct {
	Plans []PlanDefinition `json:"plans" yaml:"plans"`
}

// ErrPlanInvalid reports an invalid plan or plan file.
var ErrPlanInvalid = errors.New("invalid plan")

type planRef struct {
	key     string
	version int
}

// PlanCatalog holds every plan version of the plan file, validated. The zero
// value is an empty catalog. It is read-only and safe for concurrent use.
type PlanCatalog struct {
	plans []PlanDefinition
	index map[planRef]int
}

// NewPlanCatalog validates the plans and returns them as a catalog. It fails
// when a plan is invalid (PlanDefinition.Validate), a {key, version} appears
// twice, or a Stripe lookup key maps to more than one plan key.
func NewPlanCatalog(plans []PlanDefinition) (PlanCatalog, error) {
	catalog := PlanCatalog{index: make(map[planRef]int, len(plans))}
	lookupKeys := make(map[string]string)
	for _, plan := range plans {
		if err := plan.Validate(); err != nil {
			return PlanCatalog{}, err
		}
		ref := planRef{plan.Key, plan.Version}
		if _, exists := catalog.index[ref]; exists {
			return PlanCatalog{}, fmt.Errorf("%w: duplicate plan %s", ErrPlanInvalid, plan.ref())
		}
		for _, lookupKey := range plan.StripeLookupKeys {
			if lookupKey == "" {
				return PlanCatalog{}, fmt.Errorf("%w: plan %s has an empty Stripe lookup key", ErrPlanInvalid, plan.ref())
			}
			if other, exists := lookupKeys[lookupKey]; exists && other != plan.Key {
				return PlanCatalog{}, fmt.Errorf("%w: Stripe lookup key %q maps to plans %q and %q", ErrPlanInvalid, lookupKey, other, plan.Key)
			}
			lookupKeys[lookupKey] = plan.Key
		}
		catalog.index[ref] = len(catalog.plans)
		catalog.plans = append(catalog.plans, plan.clone())
	}
	return catalog, nil
}

// ParsePlans reads a plan file (YAML, or JSON, which is valid YAML) and
// returns its validated catalog. Unknown fields are rejected so a misspelt
// limit or flag fails instead of being ignored.
func ParsePlans(data []byte) (PlanCatalog, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var file PlanFile
	if err := decoder.Decode(&file); err != nil {
		if errors.Is(err, io.EOF) {
			return PlanCatalog{}, fmt.Errorf("%w: empty plan file", ErrPlanInvalid)
		}
		return PlanCatalog{}, fmt.Errorf("%w: %w", ErrPlanInvalid, err)
	}
	if len(file.Plans) == 0 {
		return PlanCatalog{}, fmt.Errorf("%w: plan file has no plans", ErrPlanInvalid)
	}
	return NewPlanCatalog(file.Plans)
}

// LoadPlans reads and validates the plan file at path. Services call it at
// startup and fail when it returns an error.
func LoadPlans(path string) (PlanCatalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return PlanCatalog{}, fmt.Errorf("read plan file: %w", err)
	}
	catalog, err := ParsePlans(data)
	if err != nil {
		return PlanCatalog{}, fmt.Errorf("plan file %s: %w", path, err)
	}
	return catalog, nil
}

// Get returns one plan version.
func (c PlanCatalog) Get(key string, version int) (PlanDefinition, bool) {
	position, ok := c.index[planRef{key, version}]
	if !ok {
		return PlanDefinition{}, false
	}
	return c.plans[position].clone(), true
}

// Latest returns the highest version of a plan.
func (c PlanCatalog) Latest(key string) (PlanDefinition, bool) {
	var latest *PlanDefinition
	for index := range c.plans {
		if c.plans[index].Key == key && (latest == nil || c.plans[index].Version > latest.Version) {
			latest = &c.plans[index]
		}
	}
	if latest == nil {
		return PlanDefinition{}, false
	}
	return latest.clone(), true
}

// Plans returns every plan version in file order.
func (c PlanCatalog) Plans() []PlanDefinition {
	plans := make([]PlanDefinition, len(c.plans))
	for index, plan := range c.plans {
		plans[index] = plan.clone()
	}
	return plans
}
