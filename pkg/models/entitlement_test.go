package models

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var entitlementNow = time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)

func chartPlans(t *testing.T) PlanCatalog {
	t.Helper()
	catalog, err := LoadPlans(filepath.Join("testdata", "plans.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func subscriptionOn(key string, version int) *Subscription {
	return &Subscription{Id: primitive.NewObjectID(), PlanKey: key, PlanVersion: version, Status: "active"}
}

func gatedIn(permissions []Permission) []Permission {
	var gated []Permission
	for _, permission := range permissions {
		if IsPlanGated(permission) {
			gated = append(gated, permission)
		}
	}
	return gated
}

func TestEntitlementActive(t *testing.T) {
	organisation := primitive.NewObjectID()
	active := subscriptionOn("community", 1)
	entitlement := ComputeEntitlement(organisation, active, nil, chartPlans(t), entitlementNow)

	if entitlement.State != EntitlementStateActive || entitlement.SubscriptionId != active.Id ||
		entitlement.PlanKey != "community" || entitlement.PlanVersion != 1 || entitlement.OrganisationId != organisation {
		t.Fatalf("entitlement = %+v", entitlement)
	}
	for _, permission := range []Permission{PermissionLiveStream, PermissionPTZPresets, PermissionWorkflowsCreate, PermissionCasesPage, PermissionMediaPlayHD, PermissionMediaRead, PermissionBillingManage} {
		if !entitlement.Allows(permission) {
			t.Errorf("community lacks %s", permission)
		}
	}
	for _, permission := range []Permission{PermissionStorageRead, PermissionStorageUpdate} {
		if entitlement.Allows(permission) {
			t.Errorf("community allows %s", permission)
		}
	}
	for _, permission := range UngatedPermissions() {
		if !slices.Contains(entitlement.Permissions, permission) {
			t.Errorf("ungated %s missing", permission)
		}
	}
	if !slices.Equal(entitlement.Permissions, catalogOrder(entitlement.Permissions)) {
		t.Error("permissions not in catalog order")
	}
	if entitlement.Limits.Devices.Value != 30 || !entitlement.Limits.Members.Unlimited || len(entitlement.Limits.Missing()) != 0 {
		t.Errorf("limits = %+v", entitlement.Limits)
	}
}

func TestEntitlementTrialExcludesWorkflows(t *testing.T) {
	entitlement := ComputeEntitlement(primitive.NewObjectID(), subscriptionOn("trial", 1), nil, chartPlans(t), entitlementNow)
	for _, permission := range []Permission{PermissionWorkflowsRead, PermissionWorkflowRunsCreate, PermissionCasesRead, PermissionWorkflowsPage, PermissionPagesCases, PermissionStorageRead} {
		if entitlement.Allows(permission) {
			t.Errorf("trial allows %s", permission)
		}
	}
	if !entitlement.Allows(PermissionLiveStream) || !entitlement.Allows(PermissionPagesLiveView) {
		t.Error("trial lacks live view")
	}
}

func TestEntitlementOverrides(t *testing.T) {
	active := subscriptionOn("community", 1)
	active.Overrides = &SubscriptionOverrides{
		Limits:      Limits{Devices: NewUnlimitedLimit(), RetentionDays: NewLimit(90)},
		Permissions: PermissionOverrides{Add: []string{"storage.*"}, Remove: []string{"cases.*", "live.stream"}},
	}
	entitlement := ComputeEntitlement(primitive.NewObjectID(), active, nil, chartPlans(t), entitlementNow)
	if !entitlement.Allows(PermissionStorageUpdate) {
		t.Error("added storage missing")
	}
	for _, permission := range []Permission{PermissionCasesRead, PermissionCasesPage, PermissionLiveStream} {
		if entitlement.Allows(permission) {
			t.Errorf("removed %s allowed", permission)
		}
	}
	if !entitlement.Allows(PermissionLivePreview) {
		t.Error("live.preview is not removed with live.stream")
	}
	if !entitlement.Limits.Devices.Unlimited || entitlement.Limits.RetentionDays.Value != 90 || entitlement.Limits.UploadGBPerDay.Value != 10 {
		t.Errorf("limits = %+v", entitlement.Limits)
	}
	if plan, _ := chartPlans(t).Get("community", 1); plan.Limits.Devices.Unlimited {
		t.Error("overrides changed the plan")
	}
}

func TestEntitlementWithoutSubscription(t *testing.T) {
	entitlement := ComputeEntitlement(primitive.NewObjectID(), nil, nil, chartPlans(t), entitlementNow)
	if entitlement.State != EntitlementStateInactive || !entitlement.SubscriptionId.IsZero() || entitlement.PlanKey != "" {
		t.Fatalf("entitlement = %+v", entitlement)
	}
	if gated := gatedIn(entitlement.Permissions); len(gated) != 0 {
		t.Errorf("gated permissions %v", gated)
	}
	if !slices.Equal(entitlement.Permissions, UngatedPermissions()) {
		t.Error("inactive holds the ungated permissions")
	}
	limits := entitlement.Limits
	if limits.Devices.Allows(1) || limits.UploadGBPerDay.Allows(1) || limits.WorkflowCredits.Allows(1) ||
		!limits.Projects.Unlimited || !limits.Members.Unlimited || len(limits.Missing()) != 0 {
		t.Errorf("limits = %+v", limits)
	}
}

func TestEntitlementLapsedKeepsLastLimits(t *testing.T) {
	plans := chartPlans(t)
	last := subscriptionOn("community", 1)
	last.EndsAt = entitlementNow.Add(-24 * time.Hour)
	last.Overrides = &SubscriptionOverrides{Limits: Limits{RetentionDays: NewLimit(60)}}

	for name, entitlement := range map[string]Entitlement{
		"last ended":        ComputeEntitlement(primitive.NewObjectID(), nil, last, plans, entitlementNow),
		"active row lapsed": ComputeEntitlement(primitive.NewObjectID(), last, nil, plans, entitlementNow),
	} {
		t.Run(name, func(t *testing.T) {
			if entitlement.State != EntitlementStateInactive || entitlement.SubscriptionId != last.Id ||
				entitlement.PlanKey != "community" || !entitlement.SubscriptionEndedAt.Equal(last.EndsAt) {
				t.Fatalf("entitlement = %+v", entitlement)
			}
			if gated := gatedIn(entitlement.Permissions); len(gated) != 0 {
				t.Errorf("gated permissions %v", gated)
			}
			if entitlement.Limits.RetentionDays.Value != 60 || entitlement.Limits.Devices.Value != 30 {
				t.Errorf("limits = %+v", entitlement.Limits)
			}
		})
	}
}

func TestEntitlementUnknown(t *testing.T) {
	plans := chartPlans(t)
	missingVersion := subscriptionOn("community", 7)
	lastMissing := subscriptionOn("gold", 1)
	lastMissing.EndsAt = entitlementNow.Add(-time.Hour)
	for name, entitlement := range map[string]Entitlement{
		"zero organisation":   ComputeEntitlement(primitive.NilObjectID, subscriptionOn("community", 1), nil, plans, entitlementNow),
		"plan not in catalog": ComputeEntitlement(primitive.NewObjectID(), missingVersion, nil, plans, entitlementNow),
		"no version":          ComputeEntitlement(primitive.NewObjectID(), &Subscription{StripePlan: "enterprise"}, nil, plans, entitlementNow),
		"last not in catalog": ComputeEntitlement(primitive.NewObjectID(), nil, lastMissing, plans, entitlementNow),
	} {
		t.Run(name, func(t *testing.T) {
			if entitlement.State != EntitlementStateUnknown {
				t.Fatalf("state = %s", entitlement.State)
			}
			if gated := gatedIn(entitlement.Permissions); len(gated) != 0 {
				t.Errorf("gated permissions %v", gated)
			}
			if len(entitlement.Limits.Missing()) != len(LimitNames()) {
				t.Errorf("limits = %+v", entitlement.Limits)
			}
		})
	}
}

func TestEntitlementLegacyPlanKeyAlias(t *testing.T) {
	active := &Subscription{Id: primitive.NewObjectID(), StripePlan: "enterprise", PlanVersion: 1}
	entitlement := ComputeEntitlement(primitive.NewObjectID(), active, nil, chartPlans(t), entitlementNow)
	if entitlement.State != EntitlementStateActive || entitlement.PlanKey != "enterprise" || !entitlement.Allows(PermissionStorageUpdate) {
		t.Fatalf("entitlement = %+v", entitlement)
	}
}

// The plan caps the Owner like any member: the owner bypass yields every
// permission, the entitlement removes the gated ones the plan does not allow.
func TestEntitlementCapsOwner(t *testing.T) {
	ids := legacyIds(2)
	_, owner := legacyAccessFor(t, LegacyUserAccess{OrganisationId: ids[0], UserId: ids[1], Role: "owner"})
	project := primitive.NewObjectID()
	plans := chartPlans(t)

	inactive := ComputeEntitlement(ids[0], nil, nil, plans, entitlementNow)
	if capped := inactive.CapPermissions(owner.Permissions(project)); !slices.Equal(capped, UngatedPermissions()) {
		t.Errorf("owner without subscription holds %v", gatedIn(capped))
	}
	trial := ComputeEntitlement(ids[0], subscriptionOn("trial", 1), nil, plans, entitlementNow)
	capped := trial.CapPermissions(owner.Permissions(project))
	if slices.Contains(capped, PermissionWorkflowsCreate) || slices.Contains(capped, PermissionStorageUpdate) {
		t.Error("trial owner holds workflows or storage")
	}
	if !slices.Contains(capped, PermissionLiveStream) || !slices.Contains(capped, PermissionBillingManage) {
		t.Error("trial owner lacks live view or billing")
	}
	if !owner.Organisation(PermissionStorageUpdate, primitive.NilObjectID) {
		t.Error("Access itself is unchanged until the ceiling is applied")
	}
}

type fakeEntitlementStore struct {
	active, last           *Subscription
	activeErr, lastErr     error
	activeCalls, lastCalls int
}

func (s *fakeEntitlementStore) GetActiveSubscriptionForOrganisation(context.Context, primitive.ObjectID, time.Time) (*Subscription, error) {
	s.activeCalls++
	return s.active, s.activeErr
}

func (s *fakeEntitlementStore) GetLastEndedSubscriptionForOrganisation(context.Context, primitive.ObjectID, time.Time) (*Subscription, error) {
	s.lastCalls++
	return s.last, s.lastErr
}

func TestOrganisationEntitlement(t *testing.T) {
	plans := chartPlans(t)
	organisation := primitive.NewObjectID()
	ctx := context.Background()

	store := &fakeEntitlementStore{active: subscriptionOn("enterprise", 1)}
	entitlement, err := OrganisationEntitlement(ctx, store, plans, organisation, entitlementNow)
	if err != nil || entitlement.State != EntitlementStateActive || store.lastCalls != 0 {
		t.Fatalf("active: %+v, %v, last read %d times", entitlement, err, store.lastCalls)
	}

	last := subscriptionOn("trial", 1)
	last.EndsAt = entitlementNow.Add(-time.Hour)
	store = &fakeEntitlementStore{last: last}
	entitlement, err = OrganisationEntitlement(ctx, store, plans, organisation, entitlementNow)
	if err != nil || entitlement.State != EntitlementStateInactive || entitlement.Limits.RetentionDays.Value != 7 {
		t.Fatalf("lapsed: %+v, %v", entitlement, err)
	}

	store = &fakeEntitlementStore{}
	if entitlement, err = OrganisationEntitlement(ctx, store, plans, organisation, entitlementNow); err != nil || entitlement.State != EntitlementStateInactive {
		t.Fatalf("none: %+v, %v", entitlement, err)
	}

	failure := errors.New("connection reset")
	for name, test := range map[string]struct {
		store        EntitlementStore
		organisation primitive.ObjectID
	}{
		"active read fails": {&fakeEntitlementStore{activeErr: failure}, organisation},
		"last read fails":   {&fakeEntitlementStore{lastErr: failure}, organisation},
		"plan missing":      {&fakeEntitlementStore{active: subscriptionOn("community", 9)}, organisation},
		"zero organisation": {&fakeEntitlementStore{active: subscriptionOn("community", 1)}, primitive.NilObjectID},
		"no store":          {nil, organisation},
	} {
		t.Run(name, func(t *testing.T) {
			entitlement, err := OrganisationEntitlement(ctx, test.store, plans, test.organisation, entitlementNow)
			if !errors.Is(err, ErrEntitlementUnknown) || entitlement.State != EntitlementStateUnknown {
				t.Fatalf("%+v, %v", entitlement, err)
			}
			if entitlement.OrganisationId != test.organisation {
				t.Error("organisation id not kept")
			}
			if gated := gatedIn(entitlement.Permissions); len(gated) != 0 {
				t.Errorf("gated permissions %v", gated)
			}
		})
	}
}

func TestSubscriptionPlanFieldsBSON(t *testing.T) {
	subscription := Subscription{
		Id: primitive.NewObjectID(), PlanKey: "community", PlanVersion: 1,
		Type: SubscriptionTypeCommunity, Billing: SubscriptionBillingStripe,
		Overrides: &SubscriptionOverrides{
			Limits:      Limits{UploadGBPerDay: NewUnlimitedLimit()},
			Permissions: PermissionOverrides{Add: []string{"storage.*"}},
		},
		ReachedLimit: true, ReachedLimitAt: entitlementNow.Truncate(time.Millisecond), ReachedLimitPeriod: "2026-11-19",
	}
	data, err := bson.Marshal(subscription)
	if err != nil {
		t.Fatal(err)
	}
	raw := bson.Raw(data)
	for _, key := range []string{"plan_key", "plan_version", "type", "billing", "reached_limit", "reached_limit_at", "reached_limit_period"} {
		if _, err := raw.LookupErr(key); err != nil {
			t.Errorf("%s not stored", key)
		}
	}
	if value := raw.Lookup("overrides", "limits", "uploadGBPerDay"); value.StringValue() != "unlimited" {
		t.Errorf("overrides.limits.uploadGBPerDay = %v", value)
	}
	var decoded Subscription
	if err := bson.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.PlanKey != "community" || decoded.Type != SubscriptionTypeCommunity || !decoded.Overrides.Limits.UploadGBPerDay.Unlimited ||
		!slices.Equal(decoded.Overrides.Permissions.Add, []string{"storage.*"}) || decoded.ReachedLimitPeriod != "2026-11-19" {
		t.Errorf("decoded = %+v", decoded)
	}
	if !SubscriptionTypeTrial.IsValid() || SubscriptionType("gold").IsValid() || !SubscriptionBillingInvoice.IsValid() || SubscriptionBilling("card").IsValid() {
		t.Error("type or billing validation")
	}
}

func TestSubscriptionIsActiveAt(t *testing.T) {
	if !(Subscription{}).IsActiveAt(entitlementNow) {
		t.Error("no ends_at is active")
	}
	if !(Subscription{EndsAt: entitlementNow.Add(time.Second)}).IsActiveAt(entitlementNow) {
		t.Error("future ends_at is active")
	}
	if (Subscription{EndsAt: entitlementNow}).IsActiveAt(entitlementNow) {
		t.Error("ends_at at now has lapsed")
	}
}
