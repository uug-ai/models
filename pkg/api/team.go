package api

import "github.com/uug-ai/models/pkg/models"

// TeamStatus represents specific status codes for team operations.
type TeamStatus string

const (
	TeamBindingFailed    TeamStatus = "team_binding_failed"
	TeamValidationFailed TeamStatus = "team_validation_failed"
	TeamFound            TeamStatus = "team_found"
	TeamNotFound         TeamStatus = "team_not_found"
	TeamGetAllSuccess    TeamStatus = "team_get_all_success"
	TeamGetAllFailed     TeamStatus = "team_get_all_failed"
	TeamCreateSuccess    TeamStatus = "team_create_success"
	TeamCreateFailed     TeamStatus = "team_create_failed"
	TeamUpdateSuccess    TeamStatus = "team_update_success"
	TeamUpdateFailed     TeamStatus = "team_update_failed"
	TeamDeleteSuccess    TeamStatus = "team_delete_success"
	TeamDeleteFailed     TeamStatus = "team_delete_failed"
	TeamNameExists       TeamStatus = "team_name_exists"
	// TeamGroupLinked: another team of the organisation already mirrors the
	// identity provider group.
	TeamGroupLinked TeamStatus = "team_group_linked"
	// TeamsDisabled: the organisation has not been cut over to RBAC yet, so
	// teams cannot be used (spec 7.3).
	TeamsDisabled TeamStatus = "teams_disabled"

	TeamMembersGetSuccess   TeamStatus = "team_members_get_success"
	TeamMembersGetFailed    TeamStatus = "team_members_get_failed"
	TeamMembersAddSuccess   TeamStatus = "team_members_add_success"
	TeamMembersAddFailed    TeamStatus = "team_members_add_failed"
	TeamMemberRemoveSuccess TeamStatus = "team_member_remove_success"
	TeamMemberRemoveFailed  TeamStatus = "team_member_remove_failed"
	TeamMemberNotFound      TeamStatus = "team_member_not_found"
	// TeamMemberNotInOrganisation: a user to add is not an effectively
	// active member of the organisation.
	TeamMemberNotInOrganisation TeamStatus = "team_member_not_in_organisation"
	// TeamMembersManagedByProvider: the team mirrors an identity provider
	// group, whose members only change through the provider's sync.
	TeamMembersManagedByProvider TeamStatus = "team_members_managed_by_provider"
	// TeamMemberSelf: callers never change their own access, so they cannot
	// add or remove themselves (spec 4.5).
	TeamMemberSelf TeamStatus = "team_member_self"
	// TeamMemberOwnerOnly: the team is assigned a role holding an owner-only
	// permission, so only an owner changes its members (spec 4.5).
	TeamMemberOwnerOnly TeamStatus = "team_member_owner_only"
)

// String returns the string representation of the team status.
func (s TeamStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the team status
// in the specified language, falling back to English and then to the raw
// value.
func (s TeamStatus) Translate(lang string) string {
	translations := map[string]map[TeamStatus]string{
		"en": {
			TeamBindingFailed:            "Team binding failed",
			TeamValidationFailed:         "Team validation failed",
			TeamFound:                    "Team found",
			TeamNotFound:                 "Team not found",
			TeamGetAllSuccess:            "Teams retrieved successfully",
			TeamGetAllFailed:             "Failed to retrieve teams",
			TeamCreateSuccess:            "Team created successfully",
			TeamCreateFailed:             "Failed to create team",
			TeamUpdateSuccess:            "Team updated successfully",
			TeamUpdateFailed:             "Failed to update team",
			TeamDeleteSuccess:            "Team deleted successfully",
			TeamDeleteFailed:             "Failed to delete team",
			TeamNameExists:               "A team with this name already exists",
			TeamGroupLinked:              "Another team is already linked to this identity provider group",
			TeamsDisabled:                "Teams are not available for this organisation yet",
			TeamMembersGetSuccess:        "Team members retrieved successfully",
			TeamMembersGetFailed:         "Failed to retrieve team members",
			TeamMembersAddSuccess:        "Team members added successfully",
			TeamMembersAddFailed:         "Failed to add team members",
			TeamMemberRemoveSuccess:      "Team member removed successfully",
			TeamMemberRemoveFailed:       "Failed to remove team member",
			TeamMemberNotFound:           "Team member not found",
			TeamMemberNotInOrganisation:  "The user is not an active member of this organisation",
			TeamMembersManagedByProvider: "This team's members are managed by your identity provider",
			TeamMemberSelf:               "You cannot change your own team membership",
			TeamMemberOwnerOnly:          "Only an owner can change the members of this team",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[s]; exists {
			return translation
		}
	}
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[s]; exists {
			return translation
		}
	}
	return s.String()
}

// ---------- Request / Response DTOs (team domain) ----------
//
// Teams belong to the caller's current organisation; no request names it.
// Reading needs teams.read, creating teams.create, renaming and changing
// members teams.update, deleting teams.delete (organisation level). Granting
// roles to a team goes through the role assignment endpoints with
// principal {type: "team", id}.

// TeamFilter narrows a team listing.
type TeamFilter struct {
	// Name matches teams whose name contains it, ignoring case.
	Name *string `json:"name,omitempty" bson:"name,omitempty"`
	// Source limits the listing to manual or identity provider teams.
	Source *models.TeamSource `json:"source,omitempty" bson:"source,omitempty"`
	// UserId limits the listing to the teams this user belongs to.
	UserId *string `json:"userId,omitempty" bson:"userId,omitempty"`
}

// TeamSummary is a team in a listing, with its member count.
type TeamSummary struct {
	Team        models.Team `json:"team"`
	MemberCount int64       `json:"memberCount"`
}

// TeamMemberUser is the identity of a team member: enough to show and find
// them, never credentials or profile settings.
type TeamMemberUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname,omitempty"`
}

// TeamMemberDetails is one member of a team with their identity.
type TeamMemberDetails struct {
	Member models.TeamMember `json:"member"`
	User   TeamMemberUser    `json:"user"`
}

// GetTeams lists the organisation's teams.
// @Router /teams [get]
type GetTeamsRequest struct {
	Filter     *TeamFilter       `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetTeamsResponse struct {
	Teams      []TeamSummary     `json:"teams"`
	Pagination *CursorPagination `json:"pagination,omitempty"`
}
type GetTeamsSuccessResponse struct {
	SuccessResponse
	Data GetTeamsResponse `json:"data"`
}
type GetTeamsErrorResponse struct {
	ErrorResponse
}

// GetTeam returns one team with what it may do: its role assignments and the
// roles they grant. Members are listed separately (GetTeamMembers).
// @Router /teams/{id} [get]
type GetTeamRequest struct{}
type GetTeamResponse struct {
	Team            models.Team             `json:"team"`
	MemberCount     int64                   `json:"memberCount"`
	RoleAssignments []models.RoleAssignment `json:"roleAssignments"`
	Roles           []models.Role           `json:"roles"`
}
type GetTeamSuccessResponse struct {
	SuccessResponse
	Data GetTeamResponse `json:"data"`
}
type GetTeamErrorResponse struct {
	ErrorResponse
}

// CreateTeam creates a manual team, or one mirroring an identity provider
// group. The server sets the id, organisation and audit; a synced team's
// members then come from the provider only. Members of a manual team are
// added afterwards with AddTeamMembers.
// @Router /teams [post]
type CreateTeamRequest struct {
	Name             string                       `json:"name"`
	Description      string                       `json:"description,omitempty"`
	Source           models.TeamSource            `json:"source"`
	IdentityProvider *models.TeamIdentityProvider `json:"identityProvider,omitempty"` // Source=identityProvider only
}
type CreateTeamResponse struct {
	Team models.Team `json:"team"`
}
type CreateTeamSuccessResponse struct {
	SuccessResponse
	Data CreateTeamResponse `json:"data"`
}
type CreateTeamErrorResponse struct {
	ErrorResponse
}

// UpdateTeam renames a team or changes its description; only the fields set
// on the patch change.
// @Router /teams/{id} [patch]
type UpdateTeamRequest struct {
	Team models.TeamPatch `json:"team"`
}
type UpdateTeamResponse struct {
	Team models.Team `json:"team"`
}
type UpdateTeamSuccessResponse struct {
	SuccessResponse
	Data UpdateTeamResponse `json:"data"`
}
type UpdateTeamErrorResponse struct {
	ErrorResponse
}

// DeleteTeam deletes a team with its members and role assignments; members
// keep only the access they hold themselves or through other teams.
// @Router /teams/{id} [delete]
type DeleteTeamRequest struct{}
type DeleteTeamResponse struct{}
type DeleteTeamSuccessResponse struct {
	SuccessResponse
	Data DeleteTeamResponse `json:"data"`
}
type DeleteTeamErrorResponse struct {
	ErrorResponse
}

// GetTeamMembers lists a team's members with their identity.
// @Router /teams/{id}/members [get]
type GetTeamMembersRequest struct {
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetTeamMembersResponse struct {
	Members    []TeamMemberDetails `json:"members"`
	Pagination *CursorPagination   `json:"pagination,omitempty"`
}
type GetTeamMembersSuccessResponse struct {
	SuccessResponse
	Data GetTeamMembersResponse `json:"data"`
}
type GetTeamMembersErrorResponse struct {
	ErrorResponse
}

// AddTeamMembers adds organisation members to a manual team, all or none:
// every user must be an effectively active member of the organisation and not
// the caller. Adding a member grants them every assignment of the team, so a
// team assigned a role holding an owner-only permission is changed by owners
// only (models.Access.CanChangeTeamMembers). Users already in the team are
// left as they are and reported as unchanged.
// @Router /teams/{id}/members [post]
type AddTeamMembersRequest struct {
	UserIds []string `json:"userIds"`
}
type AddTeamMembersResponse struct {
	Added     []TeamMemberDetails `json:"added"`
	Unchanged []string            `json:"unchanged,omitempty"` // User ids already in the team
}
type AddTeamMembersSuccessResponse struct {
	SuccessResponse
	Data AddTeamMembersResponse `json:"data"`
}
type AddTeamMembersErrorResponse struct {
	ErrorResponse
}

// RemoveTeamMember removes one member from a manual team. The caller cannot
// remove themselves, and needs the same access as for adding (removing
// narrows the member's access, which is a change to their grants).
// @Router /teams/{id}/members/{userId} [delete]
type RemoveTeamMemberRequest struct{}
type RemoveTeamMemberResponse struct{}
type RemoveTeamMemberSuccessResponse struct {
	SuccessResponse
	Data RemoveTeamMemberResponse `json:"data"`
}
type RemoveTeamMemberErrorResponse struct {
	ErrorResponse
}
