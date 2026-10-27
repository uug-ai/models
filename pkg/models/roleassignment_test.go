package models

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRoleValidate(t *testing.T) {
	valid := func(mutate func(*Role)) Role {
		role := Role{
			Name:         "Operator",
			Kind:         RoleKindCustom,
			AssignableAt: PermissionScopeProject,
			Permissions:  []Permission{PermissionDevicesRead, PermissionPTZControl},
		}
		mutate(&role)
		return role
	}

	tests := []struct {
		name    string
		role    Role
		wantErr error
	}{
		{"custom project role", valid(func(*Role) {}), nil},
		{"system organisation role", valid(func(r *Role) {
			r.Kind, r.Key, r.AssignableAt = RoleKindSystem, RoleKeyBillingAdmin, PermissionScopeOrganisation
			r.Permissions = []Permission{PermissionBillingRead, PermissionOrganisationRead}
		}), nil},
		{"legacy role", valid(func(r *Role) {
			r.Kind, r.Key, r.Origin = RoleKindLegacy, RoleKeyLegacyGuest, &RoleOrigin{BaseRole: BaseRoleGuest}
		}), nil},
		{"empty name", valid(func(r *Role) { r.Name = "" }), ErrRoleInvalid},
		{"unknown kind", valid(func(r *Role) { r.Kind = "builtin" }), ErrRoleInvalid},
		{"system role without key", valid(func(r *Role) { r.Kind = RoleKindSystem }), ErrRoleInvalid},
		{"custom role with key", valid(func(r *Role) { r.Key = RoleKeyEditor }), ErrRoleInvalid},
		{"unknown level", valid(func(r *Role) { r.AssignableAt = "tenant" }), ErrRoleInvalid},
		{"organisation permission on project role", valid(func(r *Role) {
			r.Permissions = append(r.Permissions, PermissionMembersRead)
		}), ErrRoleInvalid},
		{"project permission on organisation role", valid(func(r *Role) { r.AssignableAt = PermissionScopeOrganisation }), ErrRoleInvalid},
		{"unknown permission", valid(func(r *Role) { r.Permissions = []Permission{"devices.unknown"} }), ErrPermissionUnknown},
		{"application permission", valid(func(r *Role) { r.Permissions = []Permission{PermissionApplication} }), ErrPermissionUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.role.Validate(); !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestRoleRequiresAllProjects(t *testing.T) {
	delegate := Role{Permissions: []Permission{PermissionMembersRead, PermissionMembersInvite, PermissionProjectsManageAccess}}
	if delegate.RequiresAllProjects() {
		t.Fatal("a role without tenant-wide permissions must allow selected projects")
	}
	manager := Role{Permissions: []Permission{PermissionMembersRead, PermissionRolesUpdate}}
	if !manager.RequiresAllProjects() {
		t.Fatal("a role holding roles.update must require all projects")
	}
}

func TestRoleAssignmentScopeValidate(t *testing.T) {
	project := primitive.NewObjectID()
	site := primitive.NewObjectID()

	tests := []struct {
		name    string
		scope   RoleAssignmentScope
		wantErr error
	}{
		{"zero value", RoleAssignmentScope{}, ErrRoleAssignmentInvalid},
		{"organisation, all projects", RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionAll}, nil},
		{"organisation, selected projects", RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionSelected, ProjectIds: []primitive.ObjectID{project}}, nil},
		{"organisation, no project selection", RoleAssignmentScope{Level: PermissionScopeOrganisation}, ErrRoleAssignmentInvalid},
		{"organisation, selected without projects", RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionSelected}, ErrRoleAssignmentInvalid},
		{"organisation, all with projects", RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionAll, ProjectIds: []primitive.ObjectID{project}}, ErrRoleAssignmentInvalid},
		{"organisation, duplicate project", RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionSelected, ProjectIds: []primitive.ObjectID{project, project}}, ErrRoleAssignmentInvalid},
		{"organisation with resources", RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionAll, SiteIds: []primitive.ObjectID{site}}, ErrRoleAssignmentInvalid},
		{"organisation with project id", RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionAll, ProjectId: project}, ErrRoleAssignmentInvalid},
		{"project, all resources", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionAll}, nil},
		{"project, selected resources", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionSelected, SiteIds: []primitive.ObjectID{site}}, nil},
		{"project, pending", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionSelected}, nil},
		{"project without project id", RoleAssignmentScope{Level: PermissionScopeProject, Resources: ScopeSelectionAll}, ErrRoleAssignmentInvalid},
		{"project without resource selection", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project}, ErrRoleAssignmentInvalid},
		{"project, all with resources", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionAll, DeviceIds: []primitive.ObjectID{site}}, ErrRoleAssignmentInvalid},
		{"project, duplicate site", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionSelected, SiteIds: []primitive.ObjectID{site, site}}, ErrRoleAssignmentInvalid},
		{"project, empty group id", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionSelected, GroupIds: []primitive.ObjectID{{}}}, ErrRoleAssignmentInvalid},
		{"project with project selection", RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionAll, Projects: ScopeSelectionAll}, ErrRoleAssignmentInvalid},
		{"unknown level", RoleAssignmentScope{Level: "site"}, ErrRoleAssignmentInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.scope.Validate(); !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestRoleAssignmentValidate(t *testing.T) {
	organisation := primitive.NewObjectID()
	project := primitive.NewObjectID()
	editor := Role{Id: primitive.NewObjectID(), OrganisationId: organisation, AssignableAt: PermissionScopeProject, Permissions: []Permission{PermissionMediaRead}}
	manager := Role{Id: primitive.NewObjectID(), OrganisationId: organisation, AssignableAt: PermissionScopeOrganisation, Permissions: []Permission{PermissionMembersRead, PermissionRolesUpdate}}
	delegate := Role{Id: primitive.NewObjectID(), OrganisationId: organisation, AssignableAt: PermissionScopeOrganisation, Permissions: []Permission{PermissionMembersInvite, PermissionProjectsManageAccess}}

	assign := func(role Role, scope RoleAssignmentScope, mutate func(*RoleAssignment)) RoleAssignment {
		assignment := RoleAssignment{
			OrganisationId: organisation,
			Principal:      Principal{Type: PrincipalTypeUser, Id: primitive.NewObjectID()},
			RoleId:         role.Id,
			Scope:          scope,
			Source:         AssignmentSourceManual,
			IsActive:       true,
		}
		mutate(&assignment)
		return assignment
	}
	projectScope := RoleAssignmentScope{Level: PermissionScopeProject, ProjectId: project, Resources: ScopeSelectionAll}
	allProjects := RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionAll}
	selectedProjects := RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionSelected, ProjectIds: []primitive.ObjectID{project}}
	none := func(*RoleAssignment) {}

	tests := []struct {
		name       string
		role       Role
		assignment RoleAssignment
		wantErr    error
	}{
		{"project role in a project", editor, assign(editor, projectScope, none), nil},
		{"tenant-wide role on all projects", manager, assign(manager, allProjects, none), nil},
		{"delegate on selected projects", delegate, assign(delegate, selectedProjects, none), nil},
		{"tenant-wide role on selected projects", manager, assign(manager, selectedProjects, none), ErrRoleAssignmentInvalid},
		{"project role at organisation level", editor, assign(editor, allProjects, none), ErrRoleAssignmentInvalid},
		{"organisation role in a project", delegate, assign(delegate, projectScope, none), ErrRoleAssignmentInvalid},
		{"zero scope", editor, assign(editor, RoleAssignmentScope{}, none), ErrRoleAssignmentInvalid},
		{"other role", editor, assign(manager, projectScope, none), ErrRoleAssignmentInvalid},
		{"role of another organisation", editor, assign(editor, projectScope, func(a *RoleAssignment) { a.OrganisationId = primitive.NewObjectID() }), ErrRoleAssignmentInvalid},
		{"missing principal", editor, assign(editor, projectScope, func(a *RoleAssignment) { a.Principal.Id = primitive.NilObjectID }), ErrRoleAssignmentInvalid},
		{"unknown principal type", editor, assign(editor, projectScope, func(a *RoleAssignment) { a.Principal.Type = "team" }), ErrRoleAssignmentInvalid},
		{"unknown source", editor, assign(editor, projectScope, func(a *RoleAssignment) { a.Source = "" }), ErrRoleAssignmentInvalid},
		{"invalid time window", editor, assign(editor, projectScope, func(a *RoleAssignment) { a.TimeWindow = &TimeWindow{Timezone: "Mars/Olympus"} }), ErrRoleAssignmentInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.assignment.Validate(test.role); !errors.Is(err, test.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

var officeHours = TimeWindow{
	Timezone: "Europe/Brussels",
	Ranges: []TimeWindowRange{{
		Days:        []Weekday{WeekdayMonday, WeekdayTuesday, WeekdayWednesday, WeekdayThursday, WeekdayFriday, WeekdaySaturday, WeekdaySunday},
		StartMinute: 9 * 60,
		EndMinute:   17 * 60,
	}},
}

func TestTimeWindowAllows(t *testing.T) {
	tests := []struct {
		name string
		at   string
		want bool
	}{
		{"inside, winter time", "2026-01-14T08:30:00Z", true},
		{"before start, winter time", "2026-01-14T07:59:00Z", false},
		{"end is exclusive", "2026-01-14T16:00:00Z", false},
		// 29 March 2026 is the spring DST change in Brussels. Counting elapsed
		// seconds since midnight puts these an hour off; wall-clock time does not.
		{"DST day, 09:30 local", "2026-03-29T07:30:00Z", true},
		{"DST day, 17:30 local", "2026-03-29T15:30:00Z", false},
		{"summer time, 08:30 local", "2026-07-01T06:30:00Z", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			at, err := time.Parse(time.RFC3339, test.at)
			if err != nil {
				t.Fatal(err)
			}
			got, err := officeHours.Allows(at)
			if err != nil {
				t.Fatalf("Allows() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("Allows(%s) = %v, want %v", test.at, got, test.want)
			}
		})
	}
}

func TestTimeWindowMatchesDays(t *testing.T) {
	weekend := TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{Days: []Weekday{WeekdaySaturday, WeekdaySunday}, StartMinute: 0, EndMinute: minutesPerDay}}}
	saturday := time.Date(2026, 10, 31, 23, 59, 0, 0, time.UTC)
	if allowed, err := weekend.Allows(saturday); err != nil || !allowed {
		t.Fatalf("Allows(saturday) = %v, %v; want true", allowed, err)
	}
	if allowed, err := weekend.Allows(saturday.AddDate(0, 0, 2)); err != nil || allowed {
		t.Fatalf("Allows(monday) = %v, %v; want false", allowed, err)
	}
}

func TestTimeWindowValidate(t *testing.T) {
	everyDay := []Weekday{WeekdayMonday}
	tests := []struct {
		name   string
		window TimeWindow
	}{
		{"no timezone", TimeWindow{Ranges: []TimeWindowRange{{Days: everyDay, EndMinute: 60}}}},
		{"unknown timezone", TimeWindow{Timezone: "Mars/Olympus", Ranges: []TimeWindowRange{{Days: everyDay, EndMinute: 60}}}},
		{"no ranges", TimeWindow{Timezone: "UTC"}},
		{"no days", TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{EndMinute: 60}}}},
		{"unknown day", TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{Days: []Weekday{"funday"}, EndMinute: 60}}}},
		{"duplicate day", TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{Days: []Weekday{WeekdayMonday, WeekdayMonday}, EndMinute: 60}}}},
		{"empty range", TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{Days: everyDay, StartMinute: 60, EndMinute: 60}}}},
		{"past midnight", TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{Days: everyDay, StartMinute: 22 * 60, EndMinute: minutesPerDay + 60}}}},
		{"negative start", TimeWindow{Timezone: "UTC", Ranges: []TimeWindowRange{{Days: everyDay, StartMinute: -1, EndMinute: 60}}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.window.Validate(); !errors.Is(err, ErrRoleAssignmentInvalid) {
				t.Fatalf("Validate() error = %v, want ErrRoleAssignmentInvalid", err)
			}
			if allowed, err := test.window.Allows(time.Now()); err == nil || allowed {
				t.Fatalf("Allows() = %v, %v; want denied with error", allowed, err)
			}
		})
	}
}

func TestRoleAssignmentIsEffectivelyActive(t *testing.T) {
	at := time.Date(2026, 1, 14, 8, 30, 0, 0, time.UTC) // 09:30 in Brussels
	tests := []struct {
		name       string
		assignment RoleAssignment
		want       bool
	}{
		{"active", RoleAssignment{IsActive: true}, true},
		{"disabled", RoleAssignment{}, false},
		{"expired", RoleAssignment{IsActive: true, ExpiresAt: at.Add(-time.Minute)}, false},
		{"not yet expired", RoleAssignment{IsActive: true, ExpiresAt: at.Add(time.Minute)}, true},
		{"inside time window", RoleAssignment{IsActive: true, TimeWindow: &officeHours}, true},
		{"outside time window", RoleAssignment{IsActive: true, TimeWindow: &TimeWindow{Timezone: "America/New_York", Ranges: officeHours.Ranges}}, false},
		{"invalid time window", RoleAssignment{IsActive: true, TimeWindow: &TimeWindow{Timezone: "Mars/Olympus", Ranges: officeHours.Ranges}}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.assignment.isEffectivelyActiveAt(at); got != test.want {
				t.Fatalf("isEffectivelyActiveAt() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestOrganisationInvitationKeepsTokenOutOfStorage(t *testing.T) {
	invitation := OrganisationInvitation{Token: "secret-token", TokenHash: HashInvitationToken("secret-token")}

	stored, err := bson.Marshal(invitation)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(stored), "secret-token") {
		t.Fatal("stored invitation contains the raw token")
	}
	var document bson.M
	if err := bson.Unmarshal(stored, &document); err != nil {
		t.Fatal(err)
	}
	if document["tokenHash"] != HashInvitationToken("secret-token") {
		t.Fatalf("stored tokenHash = %v", document["tokenHash"])
	}

	encoded, err := json.Marshal(invitation)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), invitation.TokenHash) {
		t.Fatal("JSON invitation exposes the token hash")
	}
}

func TestHashInvitationToken(t *testing.T) {
	hash := HashInvitationToken("secret-token")
	if hash == "secret-token" || len(hash) != 64 {
		t.Fatalf("HashInvitationToken() = %q, want a SHA-256 hex digest", hash)
	}
	if HashInvitationToken("secret-token") != hash || HashInvitationToken("other-token") == hash {
		t.Fatal("HashInvitationToken() must be deterministic and distinct per token")
	}
}

func TestOrganisationIsPrimaryOwner(t *testing.T) {
	owner := primitive.NewObjectID()
	organisation := Organisation{OwnerId: owner}
	if !organisation.IsPrimaryOwner(owner) || organisation.IsPrimaryOwner(primitive.NewObjectID()) {
		t.Fatal("IsPrimaryOwner must match only the primary owner")
	}
	if (Organisation{}).IsPrimaryOwner(primitive.NilObjectID) {
		t.Fatal("an organisation without owner has no primary owner")
	}
}

func TestOwnerRole(t *testing.T) {
	owner := Role{Name: "Owner", Kind: RoleKindSystem, Key: RoleKeyOwner, AssignableAt: PermissionScopeOrganisation}
	if err := owner.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !owner.IsOwner() || !owner.RequiresAllProjects() {
		t.Fatal("Owner must be recognised and require all projects")
	}

	for name, role := range map[string]Role{
		"stored permissions": {Name: "Owner", Kind: RoleKindSystem, Key: RoleKeyOwner, AssignableAt: PermissionScopeOrganisation, Permissions: []Permission{PermissionMembersRead}},
		"project level":      {Name: "Owner", Kind: RoleKindSystem, Key: RoleKeyOwner, AssignableAt: PermissionScopeProject},
		"legacy kind":        {Name: "Owner", Kind: RoleKindLegacy, Key: RoleKeyOwner, AssignableAt: PermissionScopeOrganisation},
	} {
		t.Run(name, func(t *testing.T) {
			if err := role.Validate(); !errors.Is(err, ErrRoleInvalid) {
				t.Fatalf("Validate() error = %v, want ErrRoleInvalid", err)
			}
		})
	}
}

func TestOwnerAssignmentValidate(t *testing.T) {
	organisation := primitive.NewObjectID()
	owner := Role{Id: primitive.NewObjectID(), OrganisationId: organisation, Kind: RoleKindSystem, Key: RoleKeyOwner, AssignableAt: PermissionScopeOrganisation}
	assign := func(mutate func(*RoleAssignment)) RoleAssignment {
		assignment := RoleAssignment{
			OrganisationId: organisation,
			Principal:      Principal{Type: PrincipalTypeUser, Id: primitive.NewObjectID()},
			RoleId:         owner.Id,
			Scope:          RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionAll},
			Source:         AssignmentSourceMigration,
			IsActive:       true,
		}
		mutate(&assignment)
		return assignment
	}

	if err := assign(func(*RoleAssignment) {}).Validate(owner); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	tests := map[string]func(*RoleAssignment){
		"selected projects": func(a *RoleAssignment) {
			a.Scope = RoleAssignmentScope{Level: PermissionScopeOrganisation, Projects: ScopeSelectionSelected, ProjectIds: []primitive.ObjectID{primitive.NewObjectID()}}
		},
		"time window": func(a *RoleAssignment) { a.TimeWindow = &officeHours },
		"expiry":      func(a *RoleAssignment) { a.ExpiresAt = time.Now().Add(time.Hour) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			if err := assign(mutate).Validate(owner); !errors.Is(err, ErrRoleAssignmentInvalid) {
				t.Fatalf("Validate() error = %v, want ErrRoleAssignmentInvalid", err)
			}
		})
	}
}
