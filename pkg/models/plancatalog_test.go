package models

import (
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// testdata/plans.yaml is the Hub chart's files/plans.yaml.
func TestLoadPlansChartFile(t *testing.T) {
	catalog, err := LoadPlans(filepath.Join("testdata", "plans.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if plans := catalog.Plans(); len(plans) != 3 {
		t.Fatalf("plans = %d, want 3", len(plans))
	}
	trial, ok := catalog.Get("trial", 1)
	if !ok {
		t.Fatal("trial v1 not found")
	}
	if trial.Level != 0 || len(trial.StripeLookupKeys) != 0 || !trial.Public || trial.Archived {
		t.Errorf("trial = %+v", trial)
	}
	if !trial.Limits.Projects.Unlimited || trial.Limits.Devices.Value != 3 || trial.Limits.RetentionDays.Value != 7 {
		t.Errorf("trial limits = %v %v %v", trial.Limits.Projects, trial.Limits.Devices, trial.Limits.RetentionDays)
	}
	enterprise, _ := catalog.Get("enterprise", 1)
	if !enterprise.Limits.Devices.Unlimited || !slices.Equal(enterprise.StripeLookupKeys, []string{"enterprise_monthly"}) {
		t.Errorf("enterprise = %+v", enterprise)
	}
	if _, ok := catalog.Get("trial", 2); ok {
		t.Error("trial v2 found")
	}
	if _, ok := catalog.Get("basic", 1); ok {
		t.Error("unknown plan found")
	}
}

func TestPlanCatalogReturnsCopies(t *testing.T) {
	catalog, err := LoadPlans(filepath.Join("testdata", "plans.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	plan, _ := catalog.Get("community", 1)
	plan.Permissions[0] = "changed"
	plan.Limits.Devices.Value = 1000
	again, _ := catalog.Get("community", 1)
	if again.Permissions[0] == "changed" || again.Limits.Devices.Value != 30 {
		t.Error("Get returned shared storage")
	}
}

func TestPlanCatalogLatest(t *testing.T) {
	v1, v2 := testPlan("community", 1), testPlan("community", 2)
	v2.Limits.Devices = NewLimit(40)
	catalog, err := NewPlanCatalog([]PlanDefinition{v2, v1})
	if err != nil {
		t.Fatal(err)
	}
	latest, ok := catalog.Latest("community")
	if !ok || latest.Version != 2 || latest.Limits.Devices.Value != 40 {
		t.Errorf("latest = %+v", latest)
	}
	if _, ok := catalog.Latest("trial"); ok {
		t.Error("latest of unknown plan found")
	}
	var empty PlanCatalog
	if _, ok := empty.Get("community", 1); ok || len(empty.Plans()) != 0 {
		t.Error("zero catalog is empty")
	}
}

func testPlan(key string, version int) PlanDefinition {
	return PlanDefinition{
		Key:         key,
		Version:     version,
		Name:        key,
		Permissions: []string{"live.*", "ptz.*", "media.playHD"},
		Limits:      fullLimits(),
	}
}

func TestPlanValidation(t *testing.T) {
	missingLimit := testPlan("trial", 1)
	missingLimit.Limits.WorkflowCredits = nil
	ungated := testPlan("trial", 1)
	ungated.Permissions = []string{"media.read"}
	billing := testPlan("trial", 1)
	billing.Permissions = []string{"billing.*"}
	unknownDomain := testPlan("trial", 1)
	unknownDomain.Permissions = []string{"liveview.*"}
	lookupA, lookupB := testPlan("community", 1), testPlan("enterprise", 1)
	lookupA.StripeLookupKeys, lookupB.StripeLookupKeys = []string{"monthly"}, []string{"monthly"}
	sameKeyA, sameKeyB := testPlan("community", 1), testPlan("community", 2)
	sameKeyA.StripeLookupKeys, sameKeyB.StripeLookupKeys = []string{"community_monthly"}, []string{"community_monthly"}
	emptyLookup := testPlan("community", 1)
	emptyLookup.StripeLookupKeys = []string{""}

	tests := []struct {
		name  string
		plans []PlanDefinition
		want  string
	}{
		{"valid", []PlanDefinition{testPlan("trial", 1), testPlan("trial", 2)}, ""},
		{"same lookup key across versions", []PlanDefinition{sameKeyA, sameKeyB}, ""},
		{"missing limit", []PlanDefinition{missingLimit}, "misses limits [workflowCredits]"},
		{"duplicate", []PlanDefinition{testPlan("trial", 1), testPlan("trial", 1)}, "duplicate plan trial v1"},
		{"no key", []PlanDefinition{testPlan("", 1)}, "without key"},
		{"version 0", []PlanDefinition{testPlan("trial", 0)}, "version 0"},
		{"ungated permission", []PlanDefinition{ungated}, `"media.read" matches no plan gated permission`},
		{"never gated domain", []PlanDefinition{billing}, `"billing.*" matches no plan gated permission`},
		{"unknown domain", []PlanDefinition{unknownDomain}, `"liveview.*"`},
		{"lookup key on two plans", []PlanDefinition{lookupA, lookupB}, `lookup key "monthly" maps to plans`},
		{"empty lookup key", []PlanDefinition{emptyLookup}, "empty Stripe lookup key"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPlanCatalog(test.plans)
			if test.want == "" {
				if err != nil {
					t.Fatalf("error = %v", err)
				}
				return
			}
			if !errors.Is(err, ErrPlanInvalid) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestParsePlansErrors(t *testing.T) {
	const valid = `
plans:
  - key: trial
    version: 1
    name: Trial
    permissions: [live.*]
    limits: {devices: 3, uploadGBPerDay: 1, uploadGBPerMonth: 30, retentionDays: 7, workflowCredits: 0, projects: unlimited, members: unlimited}
    stripeLookupKeys: []
    public: true
    archived: false
`
	if _, err := ParsePlans([]byte(valid)); err != nil {
		t.Fatalf("valid file: %v", err)
	}
	tests := []struct {
		name, file, want string
	}{
		{"empty", "", "empty plan file"},
		{"no plans", "plans: []", "no plans"},
		{"unknown field", strings.Replace(valid, "public: true", "pubilc: true", 1), "pubilc"},
		{"unknown limit", strings.Replace(valid, "devices: 3", "cameras: 3", 1), "cameras"},
		{"missing limit", strings.Replace(valid, "devices: 3, ", "", 1), "misses limits [devices]"},
		{"invalid limit", strings.Replace(valid, "devices: 3", "devices: -3", 1), "invalid limit"},
		{"quoted number", strings.Replace(valid, "devices: 3", `devices: "3"`, 1), "invalid limit"},
		{"unknown permission", strings.Replace(valid, "[live.*]", "[live.everything]", 1), `"live.everything"`},
		{"duplicate", valid + strings.TrimPrefix(valid, "\nplans:\n"), "duplicate plan trial v1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParsePlans([]byte(test.file))
			if !errors.Is(err, ErrPlanInvalid) || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadPlansMissingFile(t *testing.T) {
	if _, err := LoadPlans(filepath.Join(t.TempDir(), "plans.yaml")); err == nil {
		t.Fatal("missing file loaded")
	}
}

func TestLegacyPlansUnchanged(t *testing.T) {
	if len(DefaultPlans) != 12 || DefaultPlans["enterprise"].Level != 5 || DefaultPlans["basic"].DayLimit != 3 {
		t.Error("legacy DefaultPlans changed")
	}
}
