package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LegacyDefaultTimezone is the timezone legacy accounts fall back to when they
// have none, and so the one their custom role time windows are evaluated in.
const LegacyDefaultTimezone = "Europe/Brussels"

// LegacySubject is a user's legacy access as a service loaded it, for one
// request: who they are, which organisation is selected, and the legacy role
// and allow-lists stored on the user.
type LegacySubject struct {
	UserId primitive.ObjectID
	// HomeOrganisationId is the organisation the legacy role belongs to: the
	// master account for a sub-account, the user's own account otherwise.
	HomeOrganisationId primitive.ObjectID
	// OrganisationId is the organisation selected for the request.
	OrganisationId primitive.ObjectID
	// Role is the stored legacy role: a tier name or a custom role name.
	Role    string
	Sites   []string
	Groups  []string
	Cameras []string // Device keys
	// OwnerTimezone is the home organisation owner's timezone, which custom
	// role time windows are evaluated in. Empty means LegacyDefaultTimezone.
	OwnerTimezone string
	// IgnoreTimeWindows resolves custom roles as if they had no time window,
	// for checks legacy makes without one, such as device status alerts.
	IgnoreTimeWindows bool
}

// NewLegacySubject builds the subject from a user as stored, with Master set
// for a sub-account when the caller loaded it (for its timezone).
func NewLegacySubject(user User) LegacySubject {
	home := user.Id
	if masterId, err := primitive.ObjectIDFromHex(user.MasterAccount); err == nil && !masterId.IsZero() {
		home = masterId
	}
	if !user.HasHomeOrganisation() {
		// No home organisation: the subject has only the organisation it
		// selected, and none when it selected none.
		home = primitive.NilObjectID
	}
	organisationId := user.OrganisationId
	if organisationId.IsZero() {
		organisationId = home
	}
	timezone := user.Timezone
	if user.Master != nil {
		timezone = user.Master.Timezone
	}
	return LegacySubject{
		UserId:             user.Id,
		HomeOrganisationId: home,
		OrganisationId:     organisationId,
		Role:               user.Role,
		Sites:              user.Sites,
		Groups:             user.Groups,
		Cameras:            user.Cameras,
		OwnerTimezone:      timezone,
	}
}

// LegacyAccessStore reads the legacy data resolution needs beyond the
// subject itself.
type LegacyAccessStore interface {
	// LegacyRoles returns a home organisation's custom roles with a name, in
	// the order legacy picks them, at most two: the second only reveals a
	// duplicate.
	LegacyRoles(ctx context.Context, homeOrganisationId primitive.ObjectID, name string) ([]LegacyRole, error)
	// OrganisationOwner returns the organisation's primary owner.
	OrganisationOwner(ctx context.Context, organisationId primitive.ObjectID) (primitive.ObjectID, error)
}

// LegacyReason names a known, intentional difference between resolved access
// and the legacy decision. Shadow mismatches carry the reasons that apply to
// the subject, so a mismatch without one points at a mapping bug or a missed
// call site.
type LegacyReason string

const (
	// LegacyReasonTimeWindow: the custom role's time window is checked on
	// every request in the owner's timezone; legacy checks it at login only,
	// in the user's own profile timezone.
	LegacyReasonTimeWindow LegacyReason = "timeWindow"
	// LegacyReasonDuplicateCustomRole: several custom roles share the user's
	// role name. Both pick the first one; the migration reports a conflict.
	LegacyReasonDuplicateCustomRole LegacyReason = "duplicateCustomRole"
	// LegacyReasonOtherOrganisation: the request is for an organisation other
	// than the home one. Legacy applies the home role there; resolution grants
	// Owner when the subject owns it and nothing otherwise.
	LegacyReasonOtherOrganisation LegacyReason = "otherOrganisation"
	// LegacyReasonNoRole: the subject has no stored role. Resolution keeps the
	// membership without any project; legacy may still apply its allow-lists.
	LegacyReasonNoRole LegacyReason = "noRole"
	// LegacyReasonUnresolved: the legacy access could not be resolved, so it
	// grants nothing.
	LegacyReasonUnresolved LegacyReason = "unresolved"
)

// LegacyResolution is a subject's access and the reasons it may differ from
// the legacy decision.
type LegacyResolution struct {
	Access  Access
	Reasons []LegacyReason
}

// ResolveLegacyAccess returns the subject's access at the given instant, from
// legacy data. In the home organisation the legacy role applies, as today. In
// an organisation the subject owns they hold Owner. Anywhere else they have no
// access. Any failure grants nothing and is returned with the
// LegacyReasonUnresolved reason.
func ResolveLegacyAccess(ctx context.Context, store LegacyAccessStore, subject LegacySubject, at time.Time) (LegacyResolution, error) {
	in := LegacyUserAccess{
		OrganisationId: subject.OrganisationId,
		UserId:         subject.UserId,
		Timezone:       subject.OwnerTimezone,
		// Set before the role is known so every branch maps the same way.
		IgnoreTimeWindow: subject.IgnoreTimeWindows,
	}
	if in.Timezone == "" {
		in.Timezone = LegacyDefaultTimezone
	}
	if subject.OrganisationId != subject.HomeOrganisationId {
		reasons := []LegacyReason{LegacyReasonOtherOrganisation}
		owner, err := store.OrganisationOwner(ctx, subject.OrganisationId)
		if err != nil {
			return unresolvedLegacyAccess(reasons), err
		}
		if owner != subject.UserId {
			return LegacyResolution{Reasons: reasons}, nil
		}
		in.Role = string(BaseRoleOwner)
		return resolveLegacyUserAccess(in, at, reasons)
	}

	var reasons []LegacyReason
	in.Role = subject.Role
	if subject.Role == "" {
		reasons = append(reasons, LegacyReasonNoRole)
	}
	if subject.Role != "" && !isLegacyTier(subject.Role) {
		roles, err := store.LegacyRoles(ctx, subject.HomeOrganisationId, subject.Role)
		if err != nil {
			return unresolvedLegacyAccess(reasons), err
		}
		if len(roles) > 1 {
			reasons = append(reasons, LegacyReasonDuplicateCustomRole)
		}
		if len(roles) > 0 {
			in.CustomRole = &roles[0]
			if roles[0].TimeWindowActive == 1 && !subject.IgnoreTimeWindows {
				reasons = append(reasons, LegacyReasonTimeWindow)
			}
		}
	}
	in.SiteIds = legacyObjectIds(subject.Sites)
	in.GroupIds = legacyObjectIds(subject.Groups)
	in.DeviceKeys = subject.Cameras
	return resolveLegacyUserAccess(in, at, reasons)
}

func resolveLegacyUserAccess(in LegacyUserAccess, at time.Time, reasons []LegacyReason) (LegacyResolution, error) {
	mapped, err := MapLegacyAccess(in)
	if err != nil {
		return unresolvedLegacyAccess(reasons), err
	}
	return LegacyResolution{Access: NewAccess(mapped.Membership, nil, mapped.Roles, mapped.Assignments, at), Reasons: reasons}, nil
}

func unresolvedLegacyAccess(reasons []LegacyReason) LegacyResolution {
	return LegacyResolution{Reasons: append(reasons, LegacyReasonUnresolved)}
}

func isLegacyTier(role string) bool {
	switch BaseRole(role) {
	case BaseRoleGuest, BaseRoleEditor, BaseRoleAdmin, BaseRoleOwner:
		return true
	}
	return false
}

// legacyObjectIds parses legacy id strings. An id that does not parse selects
// nothing today, so leaving it out changes nothing.
func legacyObjectIds(values []string) []primitive.ObjectID {
	var ids []primitive.ObjectID
	for _, value := range values {
		if id, err := primitive.ObjectIDFromHex(value); err == nil && !id.IsZero() {
			ids = append(ids, id)
		}
	}
	return ids
}
