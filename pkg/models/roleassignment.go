package models

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrRoleAssignmentInvalid = errors.New("invalid role assignment")

// PrincipalType enumerates who an assignment can grant to. Teams, managed in
// Hub or synced from an identity provider, become another type later without
// changing assignments.
type PrincipalType string

const (
	PrincipalTypeUser PrincipalType = "user"
)

// Principal identifies who a RoleAssignment grants to.
type Principal struct {
	Type PrincipalType      `json:"type" bson:"type"`
	Id   primitive.ObjectID `json:"id" bson:"id"`
}

// AssignmentSource records how an assignment was created, so automated writers
// (migration, later identity provider sync) can tell their grants apart from
// manual ones.
type AssignmentSource string

const (
	AssignmentSourceManual     AssignmentSource = "manual"
	AssignmentSourceInvitation AssignmentSource = "invitation"
	AssignmentSourceMigration  AssignmentSource = "migration"
)

// RoleAssignment grants a role to a principal within an explicit scope. It is
// the only thing that grants access: membership grants nothing, and a
// suspended, revoked or expired membership disables all of the member's
// assignments.
//
// An organisation holds at most one assignment per principal, role and
// project (organisation-level assignments have no project). Granting a role on
// more resources widens that assignment's scope; different roles in the same
// project are separate assignments and add up.
type RoleAssignment struct {
	Id             primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID  `json:"organisationId" bson:"organisationId,omitempty"` // Organisation context for this assignment
	Principal      Principal           `json:"principal" bson:"principal"`
	RoleId         primitive.ObjectID  `json:"roleId" bson:"roleId,omitempty"` // Reference to the organisation's Role
	Scope          RoleAssignmentScope `json:"scope" bson:"scope"`
	// TimeWindow optionally limits when the assignment is in force.
	TimeWindow *TimeWindow      `json:"timeWindow,omitempty" bson:"timeWindow,omitempty"`
	Source     AssignmentSource `json:"source" bson:"source,omitempty"`
	// ExpiresAt optionally bounds the assignment in time. It is an overlay on
	// IsActive: an assignment is only effectively active when IsActive is true
	// AND ExpiresAt is unset or in the future. Use IsEffectivelyActive().
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt,omitempty"`
	IsActive  bool      `json:"isActive" bson:"isActive"` // Administrative enable/disable switch (independent of expiry)
	Audit     Audit     `json:"audit" bson:"audit"`
}

// IsExpired reports whether the assignment has a set expiry that is in the past.
func (a RoleAssignment) IsExpired() bool {
	return a.isExpiredAt(time.Now())
}

// IsEffectivelyActive reports whether the assignment is currently in force:
// administratively active, not past its expiry, and inside its time window.
// A time window that cannot be evaluated keeps the assignment out of force.
// This is the single source of truth consumers should use for access
// decisions.
func (a RoleAssignment) IsEffectivelyActive() bool {
	return a.isEffectivelyActiveAt(time.Now())
}

func (a RoleAssignment) isExpiredAt(at time.Time) bool {
	return !a.ExpiresAt.IsZero() && a.ExpiresAt.Before(at)
}

func (a RoleAssignment) isEffectivelyActiveAt(at time.Time) bool {
	if !a.IsActive || a.isExpiredAt(at) {
		return false
	}
	if a.TimeWindow == nil {
		return true
	}
	allowed, err := a.TimeWindow.Allows(at)
	return err == nil && allowed
}

// Validate checks the assignment against the role it grants: the principal is
// set, the role belongs to the same organisation, the scope is valid and at
// the role's level, and a role holding tenant-wide permissions covers all
// projects.
func (a RoleAssignment) Validate(role Role) error {
	if a.OrganisationId.IsZero() {
		return fmt.Errorf("%w: missing organisation", ErrRoleAssignmentInvalid)
	}
	if a.Principal.Type != PrincipalTypeUser {
		return fmt.Errorf("%w: unknown principal type %q", ErrRoleAssignmentInvalid, a.Principal.Type)
	}
	if a.Principal.Id.IsZero() {
		return fmt.Errorf("%w: missing principal", ErrRoleAssignmentInvalid)
	}
	if a.RoleId.IsZero() || a.RoleId != role.Id {
		return fmt.Errorf("%w: role %s does not match %s", ErrRoleAssignmentInvalid, a.RoleId.Hex(), role.Id.Hex())
	}
	if role.OrganisationId != a.OrganisationId {
		return fmt.Errorf("%w: role belongs to another organisation", ErrRoleAssignmentInvalid)
	}
	switch a.Source {
	case AssignmentSourceManual, AssignmentSourceInvitation, AssignmentSourceMigration:
	default:
		return fmt.Errorf("%w: unknown source %q", ErrRoleAssignmentInvalid, a.Source)
	}
	if err := a.Scope.ValidateFor(role); err != nil {
		return err
	}
	if a.TimeWindow != nil {
		if err := a.TimeWindow.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// ScopeSelection says whether a scope covers everything at its level or only
// the listed ids.
type ScopeSelection string

const (
	ScopeSelectionAll      ScopeSelection = "all"
	ScopeSelectionSelected ScopeSelection = "selected"
)

// RoleAssignmentScope states explicitly what an assignment covers; nothing is
// implied by an empty field, and a zero-value scope is invalid.
//
// Level must match the role's AssignableAt.
//
// Level=organisation (organisation roles): Projects says which projects the
// organisation permissions apply to, e.g. which projects a delegated manager
// may grant access in. It never gives access to project resources. A role
// holding tenant-wide permissions requires Projects=all.
//
// Level=project (project roles): one ProjectId, and Resources=all for every
// resource in it, or Resources=selected for the union of the devices in
// SiteIds, in GroupIds and in DeviceIds. Sites and groups resolve to their
// current devices, so devices added later are covered. Selected with no ids
// is valid: the principal is in the project but has no resources yet.
type RoleAssignmentScope struct {
	Level PermissionScope `json:"level" bson:"level"`

	// Level=organisation only.
	Projects   ScopeSelection       `json:"projects,omitempty" bson:"projects,omitempty"`
	ProjectIds []primitive.ObjectID `json:"projectIds,omitempty" bson:"projectIds,omitempty"` // Projects=selected only; at least one

	// Level=project only.
	ProjectId primitive.ObjectID   `json:"projectId,omitempty" bson:"projectId,omitempty"`
	Resources ScopeSelection       `json:"resources,omitempty" bson:"resources,omitempty"`
	SiteIds   []primitive.ObjectID `json:"siteIds,omitempty" bson:"siteIds,omitempty"`     // Resources=selected only
	GroupIds  []primitive.ObjectID `json:"groupIds,omitempty" bson:"groupIds,omitempty"`   // Resources=selected only
	DeviceIds []primitive.ObjectID `json:"deviceIds,omitempty" bson:"deviceIds,omitempty"` // Resources=selected only
}

// Validate checks that the scope is complete for its level and sets no field
// of the other level.
func (s RoleAssignmentScope) Validate() error {
	switch s.Level {
	case PermissionScopeOrganisation:
		if !s.ProjectId.IsZero() || s.Resources != "" || s.hasResourceIds() {
			return fmt.Errorf("%w: organisation scope with project resources", ErrRoleAssignmentInvalid)
		}
		switch s.Projects {
		case ScopeSelectionAll:
			if len(s.ProjectIds) > 0 {
				return fmt.Errorf("%w: all projects with project ids", ErrRoleAssignmentInvalid)
			}
		case ScopeSelectionSelected:
			if len(s.ProjectIds) == 0 {
				return fmt.Errorf("%w: selected projects without project ids", ErrRoleAssignmentInvalid)
			}
		default:
			return fmt.Errorf("%w: unknown project selection %q", ErrRoleAssignmentInvalid, s.Projects)
		}
		return validateScopeIds("project", s.ProjectIds)
	case PermissionScopeProject:
		if s.Projects != "" || len(s.ProjectIds) > 0 {
			return fmt.Errorf("%w: project scope with project selection", ErrRoleAssignmentInvalid)
		}
		if s.ProjectId.IsZero() {
			return fmt.Errorf("%w: project scope without project", ErrRoleAssignmentInvalid)
		}
		switch s.Resources {
		case ScopeSelectionAll:
			if s.hasResourceIds() {
				return fmt.Errorf("%w: all resources with resource ids", ErrRoleAssignmentInvalid)
			}
		case ScopeSelectionSelected:
		default:
			return fmt.Errorf("%w: unknown resource selection %q", ErrRoleAssignmentInvalid, s.Resources)
		}
		if err := validateScopeIds("site", s.SiteIds); err != nil {
			return err
		}
		if err := validateScopeIds("group", s.GroupIds); err != nil {
			return err
		}
		return validateScopeIds("device", s.DeviceIds)
	default:
		return fmt.Errorf("%w: unknown level %q", ErrRoleAssignmentInvalid, s.Level)
	}
}

// ValidateFor checks the scope and that it fits the role: the same level, and
// all projects for a role holding tenant-wide permissions.
func (s RoleAssignmentScope) ValidateFor(role Role) error {
	if err := s.Validate(); err != nil {
		return err
	}
	if s.Level != role.AssignableAt {
		return fmt.Errorf("%w: %s scope for %s role", ErrRoleAssignmentInvalid, s.Level, role.AssignableAt)
	}
	if role.RequiresAllProjects() && s.Projects != ScopeSelectionAll {
		return fmt.Errorf("%w: role with tenant-wide permissions must cover all projects", ErrRoleAssignmentInvalid)
	}
	return nil
}

func (s RoleAssignmentScope) hasResourceIds() bool {
	return len(s.SiteIds) > 0 || len(s.GroupIds) > 0 || len(s.DeviceIds) > 0
}

func validateScopeIds(kind string, ids []primitive.ObjectID) error {
	for index, id := range ids {
		if id.IsZero() {
			return fmt.Errorf("%w: empty %s id", ErrRoleAssignmentInvalid, kind)
		}
		if slices.Contains(ids[:index], id) {
			return fmt.Errorf("%w: duplicate %s id %s", ErrRoleAssignmentInvalid, kind, id.Hex())
		}
	}
	return nil
}

// Weekday names a day in a TimeWindowRange.
type Weekday string

const (
	WeekdayMonday    Weekday = "monday"
	WeekdayTuesday   Weekday = "tuesday"
	WeekdayWednesday Weekday = "wednesday"
	WeekdayThursday  Weekday = "thursday"
	WeekdayFriday    Weekday = "friday"
	WeekdaySaturday  Weekday = "saturday"
	WeekdaySunday    Weekday = "sunday"
)

var weekdays = map[time.Weekday]Weekday{
	time.Monday:    WeekdayMonday,
	time.Tuesday:   WeekdayTuesday,
	time.Wednesday: WeekdayWednesday,
	time.Thursday:  WeekdayThursday,
	time.Friday:    WeekdayFriday,
	time.Saturday:  WeekdaySaturday,
	time.Sunday:    WeekdaySunday,
}

const minutesPerDay = 24 * 60

// TimeWindow limits when an assignment is in force. It is evaluated in its own
// Timezone, set by whoever grants the assignment and never taken from the
// user's profile, and compares wall-clock time, so DST changes don't shift it.
// An unknown timezone denies.
type TimeWindow struct {
	Timezone string            `json:"timezone" bson:"timezone"` // IANA timezone, e.g. "Europe/Brussels"
	Ranges   []TimeWindowRange `json:"ranges" bson:"ranges"`     // In force when any range matches
}

// TimeWindowRange is a daily interval on the listed days, in minutes after local
// midnight: StartMinute inclusive, EndMinute exclusive. A range ends on the
// day it starts; an overnight shift is two ranges.
type TimeWindowRange struct {
	Days        []Weekday `json:"days" bson:"days"`
	StartMinute int       `json:"startMinute" bson:"startMinute"` // 0-1439
	EndMinute   int       `json:"endMinute" bson:"endMinute"`     // 1-1440
}

// Validate checks the timezone and every range.
func (w TimeWindow) Validate() error {
	_, err := w.location()
	return err
}

// Allows reports whether the window is open at the given instant. It returns
// an error when the window is invalid; callers must then deny.
func (w TimeWindow) Allows(at time.Time) (bool, error) {
	location, err := w.location()
	if err != nil {
		return false, err
	}
	local := at.In(location)
	day := weekdays[local.Weekday()]
	minute := local.Hour()*60 + local.Minute()
	for _, timeRange := range w.Ranges {
		if slices.Contains(timeRange.Days, day) && minute >= timeRange.StartMinute && minute < timeRange.EndMinute {
			return true, nil
		}
	}
	return false, nil
}

// location validates the window and loads its timezone.
func (w TimeWindow) location() (*time.Location, error) {
	if w.Timezone == "" {
		return nil, fmt.Errorf("%w: time window without timezone", ErrRoleAssignmentInvalid)
	}
	location, err := time.LoadLocation(w.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w: time window timezone %q: %v", ErrRoleAssignmentInvalid, w.Timezone, err)
	}
	if len(w.Ranges) == 0 {
		return nil, fmt.Errorf("%w: time window without ranges", ErrRoleAssignmentInvalid)
	}
	for _, timeRange := range w.Ranges {
		if len(timeRange.Days) == 0 {
			return nil, fmt.Errorf("%w: time range without days", ErrRoleAssignmentInvalid)
		}
		for index, day := range timeRange.Days {
			if !slices.Contains(weekdayNames, day) {
				return nil, fmt.Errorf("%w: unknown weekday %q", ErrRoleAssignmentInvalid, day)
			}
			if slices.Contains(timeRange.Days[:index], day) {
				return nil, fmt.Errorf("%w: duplicate weekday %q", ErrRoleAssignmentInvalid, day)
			}
		}
		if timeRange.StartMinute < 0 || timeRange.EndMinute > minutesPerDay || timeRange.StartMinute >= timeRange.EndMinute {
			return nil, fmt.Errorf("%w: time range %d-%d", ErrRoleAssignmentInvalid, timeRange.StartMinute, timeRange.EndMinute)
		}
	}
	return location, nil
}

var weekdayNames = []Weekday{
	WeekdayMonday,
	WeekdayTuesday,
	WeekdayWednesday,
	WeekdayThursday,
	WeekdayFriday,
	WeekdaySaturday,
	WeekdaySunday,
}

// UserRoleAssignments is a helper struct to include role details with assignments
type UserRoleAssignments struct {
	UserId      primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	Assignments []RoleAssignment   `json:"assignments" bson:"assignments,omitempty"`
	Roles       []Role             `json:"roles" bson:"roles,omitempty"` // Populated role details
}
