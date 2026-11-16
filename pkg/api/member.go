package api

import "github.com/uug-ai/models/pkg/models"

// MemberStatus represents specific status codes for organisation membership
// operations.
type MemberStatus string

const (
	MemberBindingFailed    MemberStatus = "member_binding_failed"
	MemberValidationFailed MemberStatus = "member_validation_failed"
	MemberFound            MemberStatus = "member_found"
	MemberNotFound         MemberStatus = "member_not_found"
	MemberGetAllSuccess    MemberStatus = "member_get_all_success"
	MemberGetAllFailed     MemberStatus = "member_get_all_failed"
	MemberRemoveSuccess    MemberStatus = "member_remove_success"
	MemberRemoveFailed     MemberStatus = "member_remove_failed"
	// MemberSelf: callers never remove themselves; leaving an organisation is
	// a separate action.
	MemberSelf MemberStatus = "member_self"
	// MemberOwnerOnly: the member holds Owner or a role with an owner-only
	// permission, directly or through a team, so only an owner removes them
	// (spec 4.5).
	MemberOwnerOnly MemberStatus = "member_owner_only"
	// MemberLastOwner: removing the member would leave the organisation
	// without an effectively active owner.
	MemberLastOwner MemberStatus = "member_last_owner"
	// MemberPrimaryOwner: the member is the organisation's primary owner
	// (Organisation.OwnerId); name another primary owner first.
	MemberPrimaryOwner MemberStatus = "member_primary_owner"
	// MemberLegacyMode: the organisation has not been cut over to RBAC, so
	// its members are still the legacy accounts (spec 7.3).
	MemberLegacyMode MemberStatus = "member_legacy_mode"
)

// String returns the string representation of the member status.
func (s MemberStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the member
// status in the specified language, falling back to English and then to the
// raw value.
func (s MemberStatus) Translate(lang string) string {
	translations := map[string]map[MemberStatus]string{
		"en": {
			MemberBindingFailed:    "Member binding failed",
			MemberValidationFailed: "Member validation failed",
			MemberFound:            "Member found",
			MemberNotFound:         "Member not found",
			MemberGetAllSuccess:    "Members retrieved successfully",
			MemberGetAllFailed:     "Failed to retrieve members",
			MemberRemoveSuccess:    "Member removed successfully",
			MemberRemoveFailed:     "Failed to remove member",
			MemberSelf:             "You cannot remove yourself",
			MemberOwnerOnly:        "Only an owner can remove this member",
			MemberLastOwner:        "An organisation needs at least one owner",
			MemberPrimaryOwner:     "The primary owner cannot be removed",
			MemberLegacyMode:       "Members of this organisation are still managed as legacy accounts",
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

// ---------- Request / Response DTOs (member domain) ----------
//
// Members are the organisation_users memberships of the caller's current
// organisation; no request names it. They are only served for organisations
// in rbac mode (otherwise member_legacy_mode); legacy organisations keep
// their accounts on /user/accounts.
//
// Reading needs members.read; removing needs members.remove. What a member
// may do is read through the role assignment endpoints (principal user,
// includeTeams), and invitations through the invitation endpoints.

// MemberFilter narrows a member listing. Set fields combine.
type MemberFilter struct {
	// Search matches members whose username, email, first or last name
	// contains it, ignoring case.
	Search *string                  `json:"search,omitempty" bson:"search,omitempty"`
	Status *models.MembershipStatus `json:"status,omitempty" bson:"status,omitempty"`
	// TeamId limits the listing to the members of this team.
	TeamId *string `json:"teamId,omitempty" bson:"teamId,omitempty"`
}

// MemberSummary is one member with their membership and teams.
type MemberSummary struct {
	User       UserIdentity            `json:"user"`
	Membership models.OrganisationUser `json:"membership"`
	TeamIds    []string                `json:"teamIds"`
	// Owner is set when the member effectively holds the Owner role, and
	// PrimaryOwner when they are Organisation.OwnerId; both guard removal.
	Owner        bool `json:"owner"`
	PrimaryOwner bool `json:"primaryOwner"`
}

// GetMembers lists the organisation's members.
// @Router /members [get]
type GetMembersRequest struct {
	Filter     *MemberFilter     `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetMembersResponse struct {
	Members    []MemberSummary   `json:"members"`
	Pagination *CursorPagination `json:"pagination,omitempty"`
}
type GetMembersSuccessResponse struct {
	SuccessResponse
	Data GetMembersResponse `json:"data"`
}
type GetMembersErrorResponse struct {
	ErrorResponse
}

// GetMember returns one member.
// @Router /members/{userId} [get]
type GetMemberRequest struct{}
type GetMemberResponse struct {
	Member MemberSummary `json:"member"`
}
type GetMemberSuccessResponse struct {
	SuccessResponse
	Data GetMemberResponse `json:"data"`
}
type GetMemberErrorResponse struct {
	ErrorResponse
}

// RemoveMember removes a member from the organisation: the membership is
// revoked, their own role assignments are deleted and they leave the
// organisation's Hub-managed teams. The account itself is kept; without
// another organisation it starts at onboarding on its next sign-in. Nobody
// removes themselves, only owners remove an owner (or a member holding an
// owner-only permission), and the last and the primary owner stay.
// @Router /members/{userId} [delete]
type RemoveMemberRequest struct{}
type RemoveMemberResponse struct {
	RemovedAssignments     int64 `json:"removedAssignments"`
	RemovedTeamMemberships int64 `json:"removedTeamMemberships"`
}
type RemoveMemberSuccessResponse struct {
	SuccessResponse
	Data RemoveMemberResponse `json:"data"`
}
type RemoveMemberErrorResponse struct {
	ErrorResponse
}
