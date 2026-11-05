package models

import (
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func patchFixture() RoleAssignment {
	expires := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	return RoleAssignment{
		Id: primitive.NewObjectID(), OrganisationId: primitive.NewObjectID(),
		Principal: Principal{Type: PrincipalTypeUser, Id: primitive.NewObjectID()}, RoleId: primitive.NewObjectID(),
		Scope:      RoleAssignmentScope{Level: LevelProject, ProjectId: primitive.NewObjectID(), Resources: SelectionSelected, DeviceKeys: []string{"camera-1"}},
		TimeWindow: &TimeWindow{Timezone: "Europe/Brussels", Ranges: []TimeWindowRange{{Days: []Weekday{WeekdayMonday}, StartMinute: 480, EndMinute: 1020}}},
		Source:     AssignmentSourceManual, ExpiresAt: expires, IsActive: true,
	}
}

func TestRoleAssignmentPatchApply(t *testing.T) {
	assignment := patchFixture()
	if !(RoleAssignmentPatch{}).IsEmpty() {
		t.Fatal("empty patch not empty")
	}
	unchanged, err := RoleAssignmentPatch{}.Apply(assignment)
	if err != nil || unchanged.TimeWindow != assignment.TimeWindow || !unchanged.ExpiresAt.Equal(assignment.ExpiresAt) || len(unchanged.Scope.DeviceKeys) != 1 {
		t.Fatalf("empty patch changed the assignment: %+v, %v", unchanged, err)
	}

	widened := assignment.Scope
	widened.DeviceKeys = []string{"camera-1", "camera-2"}
	inactive := false
	expires := time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC)
	patched, err := RoleAssignmentPatch{Scope: &widened, ExpiresAt: &expires, IsActive: &inactive}.Apply(assignment)
	if err != nil {
		t.Fatal(err)
	}
	if len(patched.Scope.DeviceKeys) != 2 || !patched.ExpiresAt.Equal(expires) || patched.IsActive || patched.TimeWindow == nil {
		t.Errorf("patched %+v", patched)
	}
	if patched.Principal != assignment.Principal || patched.RoleId != assignment.RoleId || patched.Id != assignment.Id || patched.Source != assignment.Source {
		t.Error("patch changed a fixed field")
	}

	removed, err := RoleAssignmentPatch{RemoveTimeWindow: true, RemoveExpiry: true}.Apply(assignment)
	if err != nil || removed.TimeWindow != nil || !removed.ExpiresAt.IsZero() {
		t.Errorf("remove: %+v, %v", removed, err)
	}
	if assignment.TimeWindow == nil || assignment.ExpiresAt.IsZero() {
		t.Error("Apply changed its input")
	}
}

func TestRoleAssignmentPatchRefuses(t *testing.T) {
	assignment := patchFixture()
	otherProject := assignment.Scope
	otherProject.ProjectId = primitive.NewObjectID()
	organisation := RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionAll}
	expires := time.Now()
	for name, patch := range map[string]RoleAssignmentPatch{
		"another project":        {Scope: &otherProject},
		"another level":          {Scope: &organisation},
		"window set and removed": {TimeWindow: &TimeWindow{Timezone: "UTC"}, RemoveTimeWindow: true},
		"expiry set and removed": {ExpiresAt: &expires, RemoveExpiry: true},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := patch.Apply(assignment); !errors.Is(err, ErrRoleAssignmentInvalid) {
				t.Errorf("err = %v, want ErrRoleAssignmentInvalid", err)
			}
		})
	}
}

func TestRoleAssignmentSameSlot(t *testing.T) {
	assignment := patchFixture()
	widened := assignment
	widened.Id = primitive.NewObjectID()
	widened.Scope.DeviceKeys = []string{"camera-9"}
	if !assignment.SameSlot(widened) {
		t.Error("same principal, role and project on other resources is the same slot")
	}
	for name, mutate := range map[string]func(*RoleAssignment){
		"another principal":      func(a *RoleAssignment) { a.Principal.Id = primitive.NewObjectID() },
		"another principal type": func(a *RoleAssignment) { a.Principal.Type = "team" },
		"another role":           func(a *RoleAssignment) { a.RoleId = primitive.NewObjectID() },
		"another project":        func(a *RoleAssignment) { a.Scope.ProjectId = primitive.NewObjectID() },
		"another organisation":   func(a *RoleAssignment) { a.OrganisationId = primitive.NewObjectID() },
		"organisation level": func(a *RoleAssignment) {
			a.Scope = RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionAll}
		},
	} {
		t.Run(name, func(t *testing.T) {
			other := assignment
			mutate(&other)
			if assignment.SameSlot(other) {
				t.Error("same slot")
			}
		})
	}
}
