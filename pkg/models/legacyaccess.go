package models

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MapLegacyAccess turns a user's legacy access (tier or custom role, plus the
// sites, groups and cameras on the user) into the membership, roles and
// assignments it corresponds to. It writes nothing: services evaluate the
// result at runtime until an organisation is migrated, and the migration
// stores exactly the same documents. Every id is derived from its source, so
// the runtime result and the stored one are identical and a rerun changes
// nothing.

var ErrLegacyAccessInvalid = errors.New("invalid legacy access")

// LegacyRole is a legacy custom role document, as stored in the legacy roles
// collection.
type LegacyRole struct {
	Id                 primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	Name               string              `json:"roleName" bson:"roleName,omitempty"`
	ParentRole         string              `json:"role" bson:"role,omitempty"`
	Pages              []string            `json:"pages" bson:"pages"`
	TimeWindow         LegacyTimeWindow    `json:"timeWindow" bson:"timeWindow"`
	UserId             string              `json:"user_id" bson:"user_id,omitempty"` // The organisation's main account
	IsActive           int                 `json:"isActive" bson:"isActive"`
	FeaturePermissions LegacyFeatureLevels `json:"featurePermissions" bson:"featurePermissions"`
	TimeWindowActive   int                 `json:"timeWindowActive" bson:"timeWindowActive"`
}

// LegacyFeatureLevels are a legacy custom role's feature levels: 0 unset
// (the highest grade), 1 none, higher values a grade.
type LegacyFeatureLevels struct {
	PTZ                int `json:"ptz" bson:"ptz"`                                 // 2 zoom, 3 pan/tilt, 4 both, 5 presets
	Liveview           int `json:"liveview" bson:"liveview"`                       // 2 preview, 3 live
	RemoteConfig       int `json:"remote_config" bson:"remote_config"`             // 2 allowed
	IO                 int `json:"io" bson:"io"`                                   // Not editable; ignored
	FloorPlans         int `json:"floorPlans" bson:"floorPlans"`                   // Never read; ignored
	DownloadRecordings int `json:"download_recordings" bson:"download_recordings"` // 2 allowed
}

// LegacyTimeWindow holds up to two ranges per day, in hours after local
// midnight with both ends inclusive.
type LegacyTimeWindow struct {
	TimeRange1MinMonday    float64 `json:"timeRange1MinMonday" bson:"timeRange1MinMonday"`
	TimeRange1MaxMonday    float64 `json:"timeRange1MaxMonday" bson:"timeRange1MaxMonday"`
	TimeRange2MinMonday    float64 `json:"timeRange2MinMonday" bson:"timeRange2MinMonday"`
	TimeRange2MaxMonday    float64 `json:"timeRange2MaxMonday" bson:"timeRange2MaxMonday"`
	TimeRange1MinTuesday   float64 `json:"timeRange1MinTuesday" bson:"timeRange1MinTuesday"`
	TimeRange1MaxTuesday   float64 `json:"timeRange1MaxTuesday" bson:"timeRange1MaxTuesday"`
	TimeRange2MinTuesday   float64 `json:"timeRange2MinTuesday" bson:"timeRange2MinTuesday"`
	TimeRange2MaxTuesday   float64 `json:"timeRange2MaxTuesday" bson:"timeRange2MaxTuesday"`
	TimeRange1MinWednesday float64 `json:"timeRange1MinWednesday" bson:"timeRange1MinWednesday"`
	TimeRange1MaxWednesday float64 `json:"timeRange1MaxWednesday" bson:"timeRange1MaxWednesday"`
	TimeRange2MinWednesday float64 `json:"timeRange2MinWednesday" bson:"timeRange2MinWednesday"`
	TimeRange2MaxWednesday float64 `json:"timeRange2MaxWednesday" bson:"timeRange2MaxWednesday"`
	TimeRange1MinThursday  float64 `json:"timeRange1MinThursday" bson:"timeRange1MinThursday"`
	TimeRange1MaxThursday  float64 `json:"timeRange1MaxThursday" bson:"timeRange1MaxThursday"`
	TimeRange2MinThursday  float64 `json:"timeRange2MinThursday" bson:"timeRange2MinThursday"`
	TimeRange2MaxThursday  float64 `json:"timeRange2MaxThursday" bson:"timeRange2MaxThursday"`
	TimeRange1MinFriday    float64 `json:"timeRange1MinFriday" bson:"timeRange1MinFriday"`
	TimeRange1MaxFriday    float64 `json:"timeRange1MaxFriday" bson:"timeRange1MaxFriday"`
	TimeRange2MinFriday    float64 `json:"timeRange2MinFriday" bson:"timeRange2MinFriday"`
	TimeRange2MaxFriday    float64 `json:"timeRange2MaxFriday" bson:"timeRange2MaxFriday"`
	TimeRange1MinSaturday  float64 `json:"timeRange1MinSaturday" bson:"timeRange1MinSaturday"`
	TimeRange1MaxSaturday  float64 `json:"timeRange1MaxSaturday" bson:"timeRange1MaxSaturday"`
	TimeRange2MinSaturday  float64 `json:"timeRange2MinSaturday" bson:"timeRange2MinSaturday"`
	TimeRange2MaxSaturday  float64 `json:"timeRange2MaxSaturday" bson:"timeRange2MaxSaturday"`
	TimeRange1MinSunday    float64 `json:"timeRange1MinSunday" bson:"timeRange1MinSunday"`
	TimeRange1MaxSunday    float64 `json:"timeRange1MaxSunday" bson:"timeRange1MaxSunday"`
	TimeRange2MinSunday    float64 `json:"timeRange2MinSunday" bson:"timeRange2MinSunday"`
	TimeRange2MaxSunday    float64 `json:"timeRange2MaxSunday" bson:"timeRange2MaxSunday"`
}

// LegacyUserAccess is one user's legacy access in one organisation. The
// caller resolves the user's legacy references: site and group ids to
// ObjectIDs and camera keys to device ids, failing on any it cannot resolve
// rather than dropping it. Repeated ids are merged.
type LegacyUserAccess struct {
	OrganisationId primitive.ObjectID
	UserId         primitive.ObjectID
	// Role is the user's legacy role: a tier name, or the name of the custom
	// role in CustomRole. Tier names win, as they do today.
	Role       string
	CustomRole *LegacyRole
	SiteIds    []primitive.ObjectID
	GroupIds   []primitive.ObjectID
	DeviceIds  []primitive.ObjectID
	// Timezone is the organisation owner's IANA timezone, which custom role
	// time windows are evaluated in. Only needed when a window is active.
	Timezone string
}

// LegacyAccess is what a user's legacy access maps to: an active membership,
// the roles it references and the assignments granting them.
type LegacyAccess struct {
	Membership  OrganisationUser
	Roles       []Role
	Assignments []RoleAssignment
}

// MapLegacyAccess maps a user's legacy access. An owner gets the Owner role;
// the other tiers and custom roles get their legacy project role in the
// organisation's default project, on the user's sites, groups and cameras
// (none means no resources), plus their organisation part on all projects.
//
// A custom role keeps granting while it is marked inactive, as it does
// today. Its active time window moves to its assignments in Timezone; a
// window that is never open makes the assignments inactive.
func MapLegacyAccess(in LegacyUserAccess) (LegacyAccess, error) {
	if in.OrganisationId.IsZero() || in.UserId.IsZero() {
		return LegacyAccess{}, fmt.Errorf("%w: missing organisation or user", ErrLegacyAccessInvalid)
	}
	access := LegacyAccess{Membership: OrganisationUser{
		Id:             derivedObjectId("membership", in.OrganisationId.Hex(), in.UserId.Hex()),
		UserId:         in.UserId,
		OrganisationId: in.OrganisationId,
		Status:         MembershipStatusActive,
	}}

	tier := BaseRole(in.Role)
	switch tier {
	case "application":
		return LegacyAccess{}, fmt.Errorf("%w: the application user is not migrated", ErrLegacyAccessInvalid)
	case BaseRoleOwner:
		owner := organisationBuiltInRole(in.OrganisationId, RoleKeyOwner)
		access.add(in, owner, allProjects(), nil, true)
		return access, nil
	case BaseRoleGuest, BaseRoleEditor, BaseRoleAdmin:
		project, organisation := legacyTierRoleKeys(tier)
		access.add(in, organisationBuiltInRole(in.OrganisationId, project), in.projectScope(), nil, true)
		access.add(in, organisationBuiltInRole(in.OrganisationId, organisation), allProjects(), nil, true)
		return access, nil
	}

	role := in.CustomRole
	if role == nil || role.Name != in.Role {
		return LegacyAccess{}, fmt.Errorf("%w: custom role %q not found", ErrLegacyAccessInvalid, in.Role)
	}
	parent := BaseRole(role.ParentRole)
	features, err := role.FeaturePermissions.levels()
	if err != nil {
		return LegacyAccess{}, err
	}
	projectPermissions, organisationPermissions, ok := LegacyRolePermissions(parent, role.Pages, features)
	if !ok {
		return LegacyAccess{}, fmt.Errorf("%w: custom role %q has parent %q", ErrLegacyAccessInvalid, role.Name, role.ParentRole)
	}
	window, open, err := role.timeWindow(in.Timezone)
	if err != nil {
		return LegacyAccess{}, err
	}
	origin := &RoleOrigin{BaseRole: parent, LegacyRoleId: role.Id}
	projectRole := Role{
		Id:             derivedObjectId("role", in.OrganisationId.Hex(), "custom", role.Id.Hex(), string(LevelProject)),
		OrganisationId: in.OrganisationId,
		Name:           role.Name,
		Kind:           RoleKindCustom,
		AssignableAt:   LevelProject,
		Grants:         SelectionSelected,
		Permissions:    projectPermissions,
		Origin:         origin,
		IsActive:       true,
	}
	access.add(in, projectRole, in.projectScope(), window, open)
	access.add(in, legacyOrganisationRole(in.OrganisationId, role, organisationPermissions, origin), allProjects(), window, open)
	return access, nil
}

// BuiltInRoleId returns the id of an organisation's copy of a built-in role.
// It is derived from the organisation and the key, so every service and the
// migration agree on it without a lookup.
func BuiltInRoleId(organisationId primitive.ObjectID, key RoleKey) primitive.ObjectID {
	return derivedObjectId("role", organisationId.Hex(), string(key))
}

func (a *LegacyAccess) add(in LegacyUserAccess, role Role, scope RoleAssignmentScope, window *TimeWindow, open bool) {
	if !slices.ContainsFunc(a.Roles, func(existing Role) bool { return existing.Id == role.Id }) {
		a.Roles = append(a.Roles, role)
	}
	a.Assignments = append(a.Assignments, RoleAssignment{
		Id:             derivedObjectId("assignment", in.OrganisationId.Hex(), in.UserId.Hex(), role.Id.Hex(), scope.ProjectId.Hex()),
		OrganisationId: in.OrganisationId,
		Principal:      Principal{Type: PrincipalTypeUser, Id: in.UserId},
		RoleId:         role.Id,
		Scope:          scope,
		TimeWindow:     window,
		Source:         AssignmentSourceMigration,
		IsActive:       open,
	})
}

func (in LegacyUserAccess) projectScope() RoleAssignmentScope {
	return RoleAssignmentScope{
		Level:     LevelProject,
		ProjectId: DefaultProjectId(in.OrganisationId),
		Resources: SelectionSelected,
		SiteIds:   appendMissing(nil, in.SiteIds),
		GroupIds:  appendMissing(nil, in.GroupIds),
		DeviceIds: appendMissing(nil, in.DeviceIds),
	}
}

func allProjects() RoleAssignmentScope {
	return RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionAll}
}

func legacyTierRoleKeys(tier BaseRole) (project, organisation RoleKey) {
	switch tier {
	case BaseRoleGuest:
		return RoleKeyLegacyGuest, RoleKeyLegacyMemberOrganisation
	case BaseRoleEditor:
		return RoleKeyLegacyEditor, RoleKeyLegacyMemberOrganisation
	}
	return RoleKeyLegacyAdmin, RoleKeyLegacyAdminOrganisation
}

// organisationBuiltInRole returns an organisation's copy of a built-in role.
func organisationBuiltInRole(organisationId primitive.ObjectID, key RoleKey) Role {
	role, _ := BuiltInRole(key)
	role.Id = BuiltInRoleId(organisationId, key)
	role.OrganisationId = organisationId
	return role
}

// legacyOrganisationRole returns the organisation part of a custom role: the
// built-in legacy organisation role with the same permissions when there is
// one, otherwise a custom organisation role.
func legacyOrganisationRole(organisationId primitive.ObjectID, role *LegacyRole, permissions []Permission, origin *RoleOrigin) Role {
	for _, key := range []RoleKey{RoleKeyLegacyMemberOrganisation, RoleKeyLegacyAdminOrganisation} {
		builtIn := organisationBuiltInRole(organisationId, key)
		if slices.Equal(builtIn.Permissions, permissions) {
			return builtIn
		}
	}
	return Role{
		Id:             derivedObjectId("role", organisationId.Hex(), "custom", role.Id.Hex(), string(LevelOrganisation)),
		OrganisationId: organisationId,
		Name:           role.Name + " (organisation)",
		Kind:           RoleKindCustom,
		AssignableAt:   LevelOrganisation,
		Grants:         SelectionSelected,
		Permissions:    permissions,
		Origin:         origin,
		IsActive:       true,
	}
}

// levels validates the feature levels against their legacy scales.
func (f LegacyFeatureLevels) levels() (FeaturePermissions, error) {
	for _, level := range []struct {
		name       string
		value, max int
	}{{"ptz", f.PTZ, 5}, {"liveview", f.Liveview, 3}, {"remote_config", f.RemoteConfig, 2}, {"download_recordings", f.DownloadRecordings, 2}} {
		if level.value < 0 || level.value > level.max {
			return FeaturePermissions{}, fmt.Errorf("%w: feature %s level %d", ErrLegacyAccessInvalid, level.name, level.value)
		}
	}
	return FeaturePermissions{
		PTZ:                AccessLevel(f.PTZ),
		Liveview:           AccessLevel(f.Liveview),
		RemoteConfig:       AccessLevel(f.RemoteConfig),
		DownloadRecordings: AccessLevel(f.DownloadRecordings),
	}, nil
}

// timeWindow converts the role's active time window. Legacy ends are
// inclusive in fractional minutes; whole minutes are kept only when they are
// entirely inside, so the window never widens. open is false when an active
// window has no range left, which legacy evaluates as always closed.
func (r LegacyRole) timeWindow(timezone string) (window *TimeWindow, open bool, err error) {
	if r.TimeWindowActive != 1 {
		return nil, true, nil
	}
	w := r.TimeWindow
	days := []struct {
		day                    Weekday
		min1, max1, min2, max2 float64
	}{
		{WeekdayMonday, w.TimeRange1MinMonday, w.TimeRange1MaxMonday, w.TimeRange2MinMonday, w.TimeRange2MaxMonday},
		{WeekdayTuesday, w.TimeRange1MinTuesday, w.TimeRange1MaxTuesday, w.TimeRange2MinTuesday, w.TimeRange2MaxTuesday},
		{WeekdayWednesday, w.TimeRange1MinWednesday, w.TimeRange1MaxWednesday, w.TimeRange2MinWednesday, w.TimeRange2MaxWednesday},
		{WeekdayThursday, w.TimeRange1MinThursday, w.TimeRange1MaxThursday, w.TimeRange2MinThursday, w.TimeRange2MaxThursday},
		{WeekdayFriday, w.TimeRange1MinFriday, w.TimeRange1MaxFriday, w.TimeRange2MinFriday, w.TimeRange2MaxFriday},
		{WeekdaySaturday, w.TimeRange1MinSaturday, w.TimeRange1MaxSaturday, w.TimeRange2MinSaturday, w.TimeRange2MaxSaturday},
		{WeekdaySunday, w.TimeRange1MinSunday, w.TimeRange1MaxSunday, w.TimeRange2MinSunday, w.TimeRange2MaxSunday},
	}
	window = &TimeWindow{Timezone: timezone}
	for _, day := range days {
		for _, hours := range [][2]float64{{day.min1, day.max1}, {day.min2, day.max2}} {
			for _, value := range hours {
				if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
					return nil, false, fmt.Errorf("%w: custom role %q time window value %v", ErrLegacyAccessInvalid, r.Name, value)
				}
			}
			start := int(math.Ceil(hours[0] * 60))
			end := int(math.Min(math.Floor(hours[1]*60), minutesPerDay))
			if start < end {
				window.Ranges = append(window.Ranges, TimeWindowRange{Days: []Weekday{day.day}, StartMinute: start, EndMinute: end})
			}
		}
	}
	if len(window.Ranges) == 0 {
		return nil, false, nil
	}
	if err := window.Validate(); err != nil {
		return nil, false, err
	}
	return window, true, nil
}

// derivedObjectId returns an ObjectID derived from its parts, so the same
// source always maps to the same document.
func derivedObjectId(parts ...string) primitive.ObjectID {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	var id primitive.ObjectID
	copy(id[:], sum[:len(id)])
	return id
}
