package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMonitorStageCarriesEntitlementThroughJSON(t *testing.T) {
	stage := MonitorStage{
		OrganisationId: "org",
		Entitlement: &Entitlement{
			State:          EntitlementStateActive,
			OrganisationId: primitive.NewObjectID(),
			PlanKey:        "community",
			PlanVersion:    1,
			Permissions:    []Permission{PermissionWorkflowRunsCreate},
			Limits: Limits{
				RetentionDays:   NewLimit(30),
				WorkflowCredits: NewUnlimitedLimit(),
			},
		},
	}
	data, err := json.Marshal(stage)
	if err != nil {
		t.Fatal(err)
	}
	var decoded MonitorStage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.Entitlement
	if got == nil {
		t.Fatalf("entitlement lost in %s", data)
	}
	if got.State != EntitlementStateActive || got.PlanKey != "community" || got.PlanVersion != 1 ||
		got.OrganisationId != stage.Entitlement.OrganisationId {
		t.Fatalf("entitlement = %+v", got)
	}
	if !got.Allows(PermissionWorkflowRunsCreate) {
		t.Fatalf("decoded entitlement does not allow %s", PermissionWorkflowRunsCreate)
	}
	if got.Limits.RetentionDays == nil || got.Limits.RetentionDays.Unlimited || got.Limits.RetentionDays.Value != 30 {
		t.Fatalf("retentionDays = %+v", got.Limits.RetentionDays)
	}
	if got.Limits.WorkflowCredits == nil || !got.Limits.WorkflowCredits.Unlimited {
		t.Fatalf("workflowCredits = %+v", got.Limits.WorkflowCredits)
	}
}

func TestMonitorStageWithoutEntitlementOmitsIt(t *testing.T) {
	data, err := json.Marshal(MonitorStage{OrganisationId: "org"})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["entitlement"]; ok {
		t.Fatalf("legacy stage carries an entitlement: %s", data)
	}
	var decoded MonitorStage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Entitlement != nil {
		t.Fatalf("entitlement = %+v, want nil", decoded.Entitlement)
	}
}
