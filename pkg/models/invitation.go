package models

import (
	"errors"
	"fmt"
	"net/mail"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvitationInvalid = errors.New("invalid invitation")

// NormalizeInvitationEmail returns the form invitations store and compare an
// address in: trimmed and lower-cased, so the invitee's account matches
// whatever case either side typed.
func NormalizeInvitationEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// Validate checks that the invitation belongs to an organisation, names one
// normalised email address, has a known status, that every assignment has a
// role and a complete scope, and that each team is listed once. Whether the
// inviter may grant each assignment, and each team's assignments, is
// Access.CanGrant, checked when the invitation is created; that every team
// exists in the organisation and is managed in Hub is the writer's check.
func (i OrganisationInvitation) Validate() error {
	if i.OrganisationId.IsZero() {
		return fmt.Errorf("%w: missing organisation", ErrInvitationInvalid)
	}
	if i.Email == "" || i.Email != NormalizeInvitationEmail(i.Email) {
		return fmt.Errorf("%w: email must be set and normalised", ErrInvitationInvalid)
	}
	if address, err := mail.ParseAddress(i.Email); err != nil || address.Address != i.Email {
		return fmt.Errorf("%w: email %q is not a plain address", ErrInvitationInvalid, i.Email)
	}
	switch i.Status {
	case InvitationStatusPending, InvitationStatusAccepted, InvitationStatusExpired, InvitationStatusRevoked:
	default:
		return fmt.Errorf("%w: unknown status %q", ErrInvitationInvalid, i.Status)
	}
	for index, assignment := range i.Assignments {
		if assignment.RoleId.IsZero() {
			return fmt.Errorf("%w: assignment %d without role", ErrInvitationInvalid, index)
		}
		if err := assignment.Scope.Validate(); err != nil {
			return fmt.Errorf("%w: assignment %d: %v", ErrInvitationInvalid, index, err)
		}
		if assignment.TimeWindow != nil {
			if err := assignment.TimeWindow.Validate(); err != nil {
				return fmt.Errorf("%w: assignment %d: %v", ErrInvitationInvalid, index, err)
			}
		}
		for _, earlier := range i.Assignments[:index] {
			if earlier.RoleId == assignment.RoleId && earlier.Scope.Level == assignment.Scope.Level && earlier.Scope.ProjectId == assignment.Scope.ProjectId {
				return fmt.Errorf("%w: assignment %d repeats a role in the same project", ErrInvitationInvalid, index)
			}
		}
	}
	for index, teamId := range i.TeamIds {
		if teamId.IsZero() {
			return fmt.Errorf("%w: empty team id", ErrInvitationInvalid)
		}
		if slices.Contains(i.TeamIds[:index], teamId) {
			return fmt.Errorf("%w: team %s listed twice", ErrInvitationInvalid, teamId.Hex())
		}
	}
	return nil
}

// IsOpenAt reports whether the invitation can still be accepted: pending and
// not past its expiry. An invitation without an expiry never expires.
func (i OrganisationInvitation) IsOpenAt(at time.Time) bool {
	return i.Status == InvitationStatusPending && (i.ExpiresAt.IsZero() || at.Before(i.ExpiresAt))
}

// RoleAssignments returns the assignments to create for the invitee when the
// invitation is accepted: one per InvitationAssignment, for the user, with
// source invitation, enabled. The writer validates each against its role.
func (i OrganisationInvitation) RoleAssignments(userId primitive.ObjectID) []RoleAssignment {
	assignments := make([]RoleAssignment, 0, len(i.Assignments))
	for _, assignment := range i.Assignments {
		assignments = append(assignments, RoleAssignment{
			OrganisationId: i.OrganisationId,
			Principal:      Principal{Type: PrincipalTypeUser, Id: userId},
			RoleId:         assignment.RoleId,
			Scope:          assignment.Scope,
			TimeWindow:     assignment.TimeWindow,
			ExpiresAt:      assignment.ExpiresAt,
			Source:         AssignmentSourceInvitation,
			IsActive:       true,
		})
	}
	return assignments
}
