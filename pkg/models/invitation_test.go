package models

import (
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func invitationFixture() OrganisationInvitation {
	project := primitive.NewObjectID()
	return OrganisationInvitation{
		OrganisationId: primitive.NewObjectID(),
		Email:          "kim@example.com",
		Status:         InvitationStatusPending,
		Assignments: []InvitationAssignment{
			{RoleId: primitive.NewObjectID(), Scope: RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionAll}},
			{RoleId: primitive.NewObjectID(), Scope: RoleAssignmentScope{Level: LevelProject, ProjectId: project, Resources: SelectionSelected},
				TimeWindow: &TimeWindow{Timezone: "Europe/Brussels", Ranges: []TimeWindowRange{{Days: []Weekday{WeekdayMonday}, StartMinute: 0, EndMinute: 60}}},
				ExpiresAt:  time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
}

func TestNormalizeInvitationEmail(t *testing.T) {
	if got := NormalizeInvitationEmail("  Kim@Example.COM "); got != "kim@example.com" {
		t.Errorf("got %q", got)
	}
}

func TestOrganisationInvitationValidate(t *testing.T) {
	if err := invitationFixture().Validate(); err != nil {
		t.Fatalf("valid invitation: %v", err)
	}
	if err := (OrganisationInvitation{OrganisationId: primitive.NewObjectID(), Email: "kim@example.com", Status: InvitationStatusPending}).Validate(); err != nil {
		t.Fatalf("membership-only invitation: %v", err)
	}
	for name, mutate := range map[string]func(*OrganisationInvitation){
		"missing organisation":       func(i *OrganisationInvitation) { i.OrganisationId = primitive.NilObjectID },
		"missing email":              func(i *OrganisationInvitation) { i.Email = "" },
		"email not normalised":       func(i *OrganisationInvitation) { i.Email = "Kim@example.com" },
		"email with a display name":  func(i *OrganisationInvitation) { i.Email = "kim <kim@example.com>" },
		"not an address":             func(i *OrganisationInvitation) { i.Email = "kim" },
		"unknown status":             func(i *OrganisationInvitation) { i.Status = "sent" },
		"assignment without role":    func(i *OrganisationInvitation) { i.Assignments[0].RoleId = primitive.NilObjectID },
		"assignment with bad scope":  func(i *OrganisationInvitation) { i.Assignments[1].Scope.ProjectId = primitive.NilObjectID },
		"assignment with bad window": func(i *OrganisationInvitation) { i.Assignments[1].TimeWindow.Timezone = "Mars/Olympus" },
		"role repeated in a project": func(i *OrganisationInvitation) {
			repeat := i.Assignments[1]
			repeat.Scope.SiteIds = []primitive.ObjectID{primitive.NewObjectID()}
			i.Assignments = append(i.Assignments, repeat)
		},
	} {
		t.Run(name, func(t *testing.T) {
			invitation := invitationFixture()
			mutate(&invitation)
			if err := invitation.Validate(); !errors.Is(err, ErrInvitationInvalid) {
				t.Errorf("err = %v, want ErrInvitationInvalid", err)
			}
		})
	}
}

func TestOrganisationInvitationIsOpenAt(t *testing.T) {
	now := time.Now()
	invitation := invitationFixture()
	if !invitation.IsOpenAt(now) {
		t.Error("pending invitation without expiry is closed")
	}
	invitation.ExpiresAt = now.Add(time.Hour)
	if !invitation.IsOpenAt(now) {
		t.Error("pending invitation before expiry is closed")
	}
	if invitation.IsOpenAt(now.Add(time.Hour)) {
		t.Error("invitation open at its expiry")
	}
	for _, status := range []InvitationStatus{InvitationStatusAccepted, InvitationStatusExpired, InvitationStatusRevoked} {
		invitation.Status = status
		if invitation.IsOpenAt(now) {
			t.Errorf("%s invitation is open", status)
		}
	}
}

func TestOrganisationInvitationRoleAssignments(t *testing.T) {
	invitation := invitationFixture()
	userId := primitive.NewObjectID()
	assignments := invitation.RoleAssignments(userId)
	if len(assignments) != len(invitation.Assignments) {
		t.Fatalf("%d assignments, want %d", len(assignments), len(invitation.Assignments))
	}
	for index, assignment := range assignments {
		source := invitation.Assignments[index]
		if assignment.OrganisationId != invitation.OrganisationId || assignment.Principal != (Principal{Type: PrincipalTypeUser, Id: userId}) ||
			assignment.RoleId != source.RoleId || assignment.Scope.Level != source.Scope.Level || assignment.Scope.ProjectId != source.Scope.ProjectId ||
			assignment.TimeWindow != source.TimeWindow || !assignment.ExpiresAt.Equal(source.ExpiresAt) ||
			assignment.Source != AssignmentSourceInvitation || !assignment.IsActive || !assignment.Id.IsZero() {
			t.Errorf("assignment %d = %+v", index, assignment)
		}
	}
	if got := (OrganisationInvitation{}).RoleAssignments(userId); len(got) != 0 {
		t.Errorf("membership-only invitation created %d assignments", len(got))
	}
}
