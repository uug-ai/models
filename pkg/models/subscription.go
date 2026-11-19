package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionBillingRead   Permission = "billing.read"
	PermissionBillingManage Permission = "billing.manage"
)

var billingPermissions = []Permission{
	PermissionBillingRead,
	PermissionBillingManage,
}

// BillingPermissions returns the canonical billing permission catalog.
func BillingPermissions() []Permission {
	return append([]Permission(nil), billingPermissions...)
}

// SubscriptionType decides a subscription's lifecycle, who may create it and
// what the UI shows. It does not decide permissions or limits; the plan
// version does.
type SubscriptionType string

const (
	// SubscriptionTypeTrial is the free entry tier on the trial plan; it
	// does not end unless revoked.
	SubscriptionTypeTrial SubscriptionType = "trial"
	// SubscriptionTypeCommunity is paid monthly through Stripe on the
	// community plan.
	SubscriptionTypeCommunity SubscriptionType = "community"
	// SubscriptionTypeEnterprise is paid, on any plan, with optional
	// per-customer overrides; billed through Stripe or by invoice.
	SubscriptionTypeEnterprise SubscriptionType = "enterprise"
)

// IsValid reports whether the type is a defined one.
func (t SubscriptionType) IsValid() bool {
	switch t {
	case SubscriptionTypeTrial, SubscriptionTypeCommunity, SubscriptionTypeEnterprise:
		return true
	}
	return false
}

// SubscriptionBilling is how a subscription is paid, independent of its type.
type SubscriptionBilling string

const (
	// SubscriptionBillingStripe requires a verified card on the
	// organisation's Stripe customer.
	SubscriptionBillingStripe SubscriptionBilling = "stripe"
	// SubscriptionBillingInvoice is activated by an application admin and
	// is exempt from the card requirement.
	SubscriptionBillingInvoice SubscriptionBilling = "invoice"
)

// IsValid reports whether the billing is a defined one.
func (b SubscriptionBilling) IsValid() bool {
	switch b {
	case SubscriptionBillingStripe, SubscriptionBillingInvoice:
		return true
	}
	return false
}

// SubscriptionOverrides are per-customer changes to the subscription's plan
// version.
type SubscriptionOverrides struct {
	// Limits replace the plan's limits; a nil limit means the plan's value.
	Limits Limits `json:"limits" bson:"limits"`
	// Permissions adds or removes gated permission patterns, with the same
	// patterns as plans.
	Permissions PermissionOverrides `json:"permissions" bson:"permissions"`
}

// PermissionOverrides are per-customer add-ons and exclusions of gated
// permission patterns.
type PermissionOverrides struct {
	Add    []string `json:"add,omitempty" bson:"add,omitempty"`
	Remove []string `json:"remove,omitempty" bson:"remove,omitempty"`
}

type Subscription struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisation_id,omitempty" bson:"organisation_id,omitempty"` // Organisation this subscription belongs to
	UserId         string             `json:"user_id,omitempty" bson:"user_id,omitempty"`                 // Legacy: user who created/owns the subscription
	StripeId       string             `json:"stripe_id,omitempty" bson:"stripe_id,omitempty"`
	// Deprecated: use PlanKey; read both through EffectivePlanKey while
	// StripePlan is kept as an alias during the transition.
	StripePlan string `json:"stripe_plan,omitempty" bson:"stripe_plan,omitempty"`
	// Deprecated: unused; the device limit comes from the plan
	// (Limits.Devices) and its overrides.
	Quantity    int       `json:"quantity,omitempty" bson:"quantity,omitempty"`
	Status      string    `json:"status,omitempty" bson:"status,omitempty"` // active, cancelled, past_due, trialing, etc.
	TrialEndsAt time.Time `json:"trial_ends_at,omitempty" bson:"trial_ends_at,omitempty"`
	EndsAt      time.Time `json:"ends_at,omitempty" bson:"ends_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`

	// PlanKey and PlanVersion pin the plan version this subscription is an
	// instance of (PlanCatalog.Get). PlanKey replaces StripePlan.
	PlanKey     string `json:"plan_key,omitempty" bson:"plan_key,omitempty"`
	PlanVersion int    `json:"plan_version,omitempty" bson:"plan_version,omitempty"`
	// Type is trial, community or enterprise.
	Type SubscriptionType `json:"type,omitempty" bson:"type,omitempty"`
	// Billing is how the subscription is paid: stripe or invoice.
	Billing SubscriptionBilling `json:"billing,omitempty" bson:"billing,omitempty"`
	// Overrides are per-customer limits and permission changes on top of
	// the plan version.
	Overrides *SubscriptionOverrides `json:"overrides,omitempty" bson:"overrides,omitempty"`

	// ReachedLimit is set while the organisation is over an upload limit
	// (uploadGBPerDay or uploadGBPerMonth); only recording uploads are
	// refused. It replaces users.reachedLimit.
	ReachedLimit bool `json:"reached_limit,omitempty" bson:"reached_limit,omitempty"`
	// ReachedLimitAt is when ReachedLimit was set.
	ReachedLimitAt time.Time `json:"reached_limit_at,omitempty" bson:"reached_limit_at,omitempty"`
	// ReachedLimitPeriod is the period ReachedLimit was set for: a day
	// ("2026-11-16") or a month ("2026-11") in the organisation's timezone
	// (UsageDayPeriod, UsageMonthPeriod). The first upload in a later period
	// clears the flag.
	ReachedLimitPeriod string `json:"reached_limit_period,omitempty" bson:"reached_limit_period,omitempty"`
}

// EffectivePlanKey returns PlanKey, or the legacy StripePlan alias when
// PlanKey is not set.
func (s Subscription) EffectivePlanKey() string {
	if s.PlanKey != "" {
		return s.PlanKey
	}
	return s.StripePlan
}

// IsActiveAt reports whether the subscription has not lapsed at now: EndsAt
// is unset or after now. It is the in-memory form of ActiveSubscriptionFilter
// and has the same meaning.
func (s Subscription) IsActiveAt(now time.Time) bool {
	return s.EndsAt.IsZero() || s.EndsAt.After(now)
}
