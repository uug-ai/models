package models

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EntitlementState is whether an organisation's subscription is in force.
type EntitlementState string

const (
	// EntitlementStateActive: the organisation has an active subscription
	// whose plan version is in the plan catalog.
	EntitlementStateActive EntitlementState = "active"
	// EntitlementStateInactive: the ownership query succeeded and there is
	// no active subscription (never subscribed, revoked or lapsed).
	EntitlementStateInactive EntitlementState = "inactive"
	// EntitlementStateUnknown: the subscription could not be read or
	// resolved (query or decode error, a plan version missing from the
	// catalog). New access fails closed; destructive cleanup skips and
	// retries.
	EntitlementStateUnknown EntitlementState = "unknown"
)

// Entitlement is what an organisation's subscription allows now: the
// permission ceiling and the limits. It is computed, never stored
// (ComputeEntitlement, OrganisationEntitlement), and carries the
// organisation and the subscription it came from.
//
// A plan caps, it never grants: Permissions is a ceiling on what role
// assignments grant, and it caps the Owner too (CapPermissions).
type Entitlement struct {
	State          EntitlementState   `json:"state"`
	OrganisationId primitive.ObjectID `json:"organisationId"`
	// SubscriptionId, PlanKey and PlanVersion identify the subscription and
	// plan version the entitlement came from: the active subscription, or
	// the last one after a lapse. Unset when there is none.
	SubscriptionId primitive.ObjectID `json:"subscriptionId,omitempty"`
	PlanKey        string             `json:"planKey,omitempty"`
	PlanVersion    int                `json:"planVersion,omitempty"`
	// SubscriptionEndedAt is when the last subscription ended, set when the
	// organisation is inactive after a lapse. Cleanup deletes the
	// organisation's data 30 days after it.
	SubscriptionEndedAt time.Time `json:"subscriptionEndedAt,omitempty"`
	// Permissions are the ungated permissions plus the gated ones the plan
	// allows after overrides, in catalog order. Inactive and unknown
	// entitlements hold the ungated permissions only.
	Permissions []Permission `json:"permissions"`
	// Limits are the plan's limits with the subscription's overrides
	// applied. After a lapse they are the last subscription's (retention
	// keeps running with them until the data is deleted); check State
	// before allowing uploads or devices. Without any subscription: no
	// active devices, no uploads, no credits, no retention; projects and
	// members unlimited.
	Limits Limits `json:"limits"`
}

// Allows reports whether the entitlement allows a permission: every ungated
// permission, and the gated ones in Permissions. Unknown permissions are not
// allowed.
func (e Entitlement) Allows(permission Permission) bool {
	definition, ok := LookupPermission(permission)
	if !ok {
		return false
	}
	return !definition.PlanGated || slices.Contains(e.Permissions, definition.Permission)
}

// CapPermissions returns the permissions the entitlement allows, in the
// given order: the plan ceiling applied to resolved access, for the Owner as
// for any member (e.g. CapPermissions(access.Permissions(projectId))).
func (e Entitlement) CapPermissions(permissions []Permission) []Permission {
	var capped []Permission
	for _, permission := range permissions {
		if e.Allows(permission) {
			capped = append(capped, permission)
		}
	}
	return capped
}

// noSubscriptionLimits are the limits of an organisation that never had a
// subscription.
func noSubscriptionLimits() Limits {
	return Limits{
		RetentionDays:    NewLimit(0),
		UploadGBPerDay:   NewLimit(0),
		UploadGBPerMonth: NewLimit(0),
		WorkflowCredits:  NewLimit(0),
		Devices:          NewLimit(0),
		Projects:         NewUnlimitedLimit(),
		Members:          NewUnlimitedLimit(),
	}
}

// ComputeEntitlement computes an organisation's entitlement at now, without
// I/O, from:
//   - active: the organisation's active subscription
//     (GetActiveSubscriptionForOrganisation), nil when it has none; a
//     subscription that is not active at now (Subscription.IsActiveAt) is
//     treated as ended;
//   - last: the organisation's most recent ended subscription, nil when it
//     has none; used for the limits after a lapse;
//   - plans: the plan catalog.
//
// The result is:
//   - active when active resolves to a plan version in the catalog, with
//     permissions = ungated ∪ resolve(plan + overrides.add − overrides.remove)
//     and limits = plan limits with overrides;
//   - inactive without an active subscription: ungated permissions only,
//     with the last subscription's limits, or no-subscription limits when
//     the organisation never had one;
//   - unknown when the organisation id is zero or the plan version of the
//     subscription it depends on is not in the catalog: ungated permissions
//     only and no limits.
func ComputeEntitlement(organisationId primitive.ObjectID, active, last *Subscription, plans PlanCatalog, now time.Time) Entitlement {
	entitlement := Entitlement{
		State:          EntitlementStateUnknown,
		OrganisationId: organisationId,
		Permissions:    UngatedPermissions(),
	}
	if organisationId.IsZero() {
		return entitlement
	}
	if active != nil && !active.IsActiveAt(now) {
		if last == nil || active.EndsAt.After(last.EndsAt) {
			last = active
		}
		active = nil
	}

	if active != nil {
		plan, ok := subscriptionPlan(*active, plans)
		if !ok {
			return entitlement.from(*active)
		}
		var overrides SubscriptionOverrides
		if active.Overrides != nil {
			overrides = *active.Overrides
		}
		entitlement = entitlement.from(*active)
		entitlement.State = EntitlementStateActive
		entitlement.Permissions = mergeCatalogOrder(UngatedPermissions(),
			ResolvePlanPermissions(plan.Permissions, overrides.Permissions.Add, overrides.Permissions.Remove))
		entitlement.Limits = plan.Limits.WithOverrides(overrides.Limits)
		return entitlement
	}

	if last == nil {
		entitlement.State = EntitlementStateInactive
		entitlement.Limits = noSubscriptionLimits()
		return entitlement
	}
	plan, ok := subscriptionPlan(*last, plans)
	if !ok {
		return entitlement.from(*last)
	}
	entitlement = entitlement.from(*last)
	entitlement.State = EntitlementStateInactive
	entitlement.SubscriptionEndedAt = last.EndsAt
	entitlement.Limits = plan.Limits
	if last.Overrides != nil {
		entitlement.Limits = plan.Limits.WithOverrides(last.Overrides.Limits)
	}
	return entitlement
}

func (e Entitlement) from(subscription Subscription) Entitlement {
	e.SubscriptionId = subscription.Id
	e.PlanKey = subscription.EffectivePlanKey()
	e.PlanVersion = subscription.PlanVersion
	return e
}

func subscriptionPlan(subscription Subscription, plans PlanCatalog) (PlanDefinition, bool) {
	return plans.Get(subscription.EffectivePlanKey(), subscription.PlanVersion)
}

// mergeCatalogOrder returns the union of two permission lists in catalog
// order.
func mergeCatalogOrder(a, b []Permission) []Permission {
	merged := slices.Concat(a, b)
	slices.SortFunc(merged, func(x, y Permission) int { return permissionIndex[x] - permissionIndex[y] })
	return slices.Compact(merged)
}

// EntitlementStore reads what OrganisationEntitlement needs. Both reads are
// scoped to the organisation (SubscriptionOwnershipFilter) and ordered with
// LatestSubscriptionSort; a nil subscription with a nil error means there is
// none.
type EntitlementStore interface {
	// GetActiveSubscriptionForOrganisation returns the organisation's
	// active subscription at now (ActiveSubscriptionFilter).
	GetActiveSubscriptionForOrganisation(ctx context.Context, organisationId primitive.ObjectID, now time.Time) (*Subscription, error)
	// GetLastEndedSubscriptionForOrganisation returns the organisation's
	// most recently ended subscription at now (ends_at at or before now).
	GetLastEndedSubscriptionForOrganisation(ctx context.Context, organisationId primitive.ObjectID, now time.Time) (*Subscription, error)
}

// ErrEntitlementUnknown is returned with an unknown entitlement.
var ErrEntitlementUnknown = errors.New("entitlement unknown")

// OrganisationEntitlement reads the organisation's subscriptions and computes
// its entitlement (ComputeEntitlement). It is the one implementation shared by
// Hub API, the pipelines, hub-cleanup and the CLI. On a read error, a zero
// organisation id or a plan version missing from the catalog it returns an
// unknown entitlement (ungated permissions only) together with an error
// wrapping ErrEntitlementUnknown.
func OrganisationEntitlement(ctx context.Context, store EntitlementStore, plans PlanCatalog, organisationId primitive.ObjectID, now time.Time) (Entitlement, error) {
	unknown := ComputeEntitlement(primitive.NilObjectID, nil, nil, plans, now)
	unknown.OrganisationId = organisationId
	if organisationId.IsZero() {
		return unknown, fmt.Errorf("%w: no organisation", ErrEntitlementUnknown)
	}
	if store == nil {
		return unknown, fmt.Errorf("%w: no store", ErrEntitlementUnknown)
	}
	active, err := store.GetActiveSubscriptionForOrganisation(ctx, organisationId, now)
	if err != nil {
		return unknown, fmt.Errorf("%w: active subscription of %s: %w", ErrEntitlementUnknown, organisationId.Hex(), err)
	}
	var last *Subscription
	if active == nil || !active.IsActiveAt(now) {
		last, err = store.GetLastEndedSubscriptionForOrganisation(ctx, organisationId, now)
		if err != nil {
			return unknown, fmt.Errorf("%w: last subscription of %s: %w", ErrEntitlementUnknown, organisationId.Hex(), err)
		}
	}
	entitlement := ComputeEntitlement(organisationId, active, last, plans, now)
	if entitlement.State == EntitlementStateUnknown {
		return entitlement, fmt.Errorf("%w: plan %q version %d of organisation %s is not in the plan catalog",
			ErrEntitlementUnknown, entitlement.PlanKey, entitlement.PlanVersion, organisationId.Hex())
	}
	return entitlement, nil
}
