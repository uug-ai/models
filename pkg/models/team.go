package models

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionTeamsRead   Permission = "teams.read"
	PermissionTeamsCreate Permission = "teams.create"
	// PermissionTeamsUpdate covers a team's name and description and its
	// members. Adding a member grants them every assignment of the team, so
	// writers bound it like a grant (spec 4.5).
	PermissionTeamsUpdate Permission = "teams.update"
	PermissionTeamsDelete Permission = "teams.delete"
)

var teamPermissions = []Permission{
	PermissionTeamsRead,
	PermissionTeamsCreate,
	PermissionTeamsUpdate,
	PermissionTeamsDelete,
}

// TeamPermissions returns the canonical team permission catalog.
func TeamPermissions() []Permission {
	return append([]Permission(nil), teamPermissions...)
}

var ErrTeamInvalid = errors.New("invalid team")

// maxTeamNameLength bounds a team's name, in characters.
const maxTeamNameLength = 100

// TeamSource says who decides a team's members.
type TeamSource string

const (
	// TeamSourceManual teams are managed in Hub: members are added and
	// removed there.
	TeamSourceManual TeamSource = "manual"
	// TeamSourceIdentityProvider teams mirror a group of an identity
	// provider. Only the provider's sync (SCIM or login claims) changes their
	// members; Hub still decides what the team may do through its
	// assignments.
	TeamSourceIdentityProvider TeamSource = "identityProvider"
)

// TeamIdentityProvider links a team to a group of an identity provider.
type TeamIdentityProvider struct {
	// Provider names the organisation's identity provider connection the
	// group comes from (e.g. its Entra ID or Okta tenant).
	Provider string `json:"provider" bson:"provider"`
	// ExternalGroupId is the provider's stable id for the group, never its
	// display name, which can change.
	ExternalGroupId string `json:"externalGroupId" bson:"externalGroupId"`
}

// Team is a named set of an organisation's members that role assignments can
// target (PrincipalTypeTeam), so access is granted once to the team instead of
// to each member. A team grants nothing by itself: its members receive the
// team's assignments, and only while their own membership is effectively
// active.
//
// Members are stored as TeamMember documents, one per member, not on the team.
// Within an organisation, team names are unique ignoring case, and there is at
// most one team per identity provider group. Source and IdentityProvider are
// fixed at creation (TeamPatch never changes them).
type Team struct {
	Id               primitive.ObjectID    `json:"id" bson:"_id,omitempty"`
	OrganisationId   primitive.ObjectID    `json:"organisationId" bson:"organisationId,omitempty"`
	Name             string                `json:"name" bson:"name,omitempty"`
	Description      string                `json:"description,omitempty" bson:"description,omitempty"`
	Source           TeamSource            `json:"source" bson:"source,omitempty"`
	IdentityProvider *TeamIdentityProvider `json:"identityProvider,omitempty" bson:"identityProvider,omitempty"` // Source=identityProvider only
	Audit            Audit                 `json:"audit" bson:"audit"`
}

// Validate checks that the team belongs to an organisation, has a name, and
// links an identity provider group exactly when its source says so.
func (t Team) Validate() error {
	if t.OrganisationId.IsZero() {
		return fmt.Errorf("%w: missing organisation", ErrTeamInvalid)
	}
	name := strings.TrimSpace(t.Name)
	if name == "" {
		return fmt.Errorf("%w: missing name", ErrTeamInvalid)
	}
	if name != t.Name {
		return fmt.Errorf("%w: name has leading or trailing spaces", ErrTeamInvalid)
	}
	if utf8.RuneCountInString(name) > maxTeamNameLength {
		return fmt.Errorf("%w: name longer than %d characters", ErrTeamInvalid, maxTeamNameLength)
	}
	switch t.Source {
	case TeamSourceManual:
		if t.IdentityProvider != nil {
			return fmt.Errorf("%w: manual team with an identity provider group", ErrTeamInvalid)
		}
	case TeamSourceIdentityProvider:
		if t.IdentityProvider == nil ||
			strings.TrimSpace(t.IdentityProvider.Provider) == "" ||
			strings.TrimSpace(t.IdentityProvider.ExternalGroupId) == "" {
			return fmt.Errorf("%w: identity provider team without provider and group id", ErrTeamInvalid)
		}
	default:
		return fmt.Errorf("%w: unknown source %q", ErrTeamInvalid, t.Source)
	}
	return nil
}

// ManagesMembersInHub reports whether members may be added and removed in
// Hub. A team synced from an identity provider only changes through the
// provider's sync.
func (t Team) ManagesMembersInHub() bool {
	return t.Source == TeamSourceManual
}

// TeamPatch is a partial update of a team: only the fields that are set
// change. A team's source and identity provider group never change; linking
// another group means creating another team.
type TeamPatch struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// IsEmpty reports whether the patch changes nothing.
func (p TeamPatch) IsEmpty() bool {
	return p.Name == nil && p.Description == nil
}

// Apply returns the team with the patch applied. The result still needs
// Validate.
func (p TeamPatch) Apply(team Team) Team {
	if p.Name != nil {
		team.Name = *p.Name
	}
	if p.Description != nil {
		team.Description = *p.Description
	}
	return team
}

// TeamMember puts one organisation member in one team. It carries no access
// of its own: the member receives the team's assignments while both the team
// membership and their organisation membership exist and the latter is
// effectively active. Removing a member deletes the document; suspending the
// organisation membership disables team-derived access with the rest.
//
// A team holds a user at most once.
type TeamMember struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"`
	TeamId         primitive.ObjectID `json:"teamId" bson:"teamId,omitempty"`
	UserId         primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	Audit          Audit              `json:"audit" bson:"audit"`
}

// Validate checks the member against its team: both ids are set and the
// member belongs to the team's organisation. Whether the writer may change
// the team's members at all is ManagesMembersInHub (or the provider sync).
func (m TeamMember) Validate(team Team) error {
	if m.UserId.IsZero() {
		return fmt.Errorf("%w: member without user", ErrTeamInvalid)
	}
	if m.TeamId.IsZero() || m.TeamId != team.Id {
		return fmt.Errorf("%w: member of team %s, not %s", ErrTeamInvalid, m.TeamId.Hex(), team.Id.Hex())
	}
	if m.OrganisationId.IsZero() || m.OrganisationId != team.OrganisationId {
		return fmt.Errorf("%w: member of another organisation", ErrTeamInvalid)
	}
	return nil
}

// teamIdsOf returns the teams a member belongs to in the membership's
// organisation, ignoring anything for another user or organisation.
func teamIdsOf(membership OrganisationUser, teams []TeamMember) []primitive.ObjectID {
	var ids []primitive.ObjectID
	for _, member := range teams {
		if member.UserId == membership.UserId && member.OrganisationId == membership.OrganisationId && !member.TeamId.IsZero() {
			ids = append(ids, member.TeamId)
		}
	}
	return ids
}
