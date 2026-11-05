package api

import (
	"time"

	"github.com/uug-ai/models/pkg/models"
)

// InvitationResponseStatus represents specific status codes for organisation
// invitation operations. (models.InvitationStatus is an invitation's own
// lifecycle state.)
type InvitationResponseStatus string

const (
	InvitationBindingFailed    InvitationResponseStatus = "invitation_binding_failed"
	InvitationValidationFailed InvitationResponseStatus = "invitation_validation_failed"
	InvitationFound            InvitationResponseStatus = "invitation_found"
	InvitationNotFound         InvitationResponseStatus = "invitation_not_found"
	InvitationGetAllSuccess    InvitationResponseStatus = "invitation_get_all_success"
	InvitationGetAllFailed     InvitationResponseStatus = "invitation_get_all_failed"
	InvitationCreateSuccess    InvitationResponseStatus = "invitation_create_success"
	InvitationCreateFailed     InvitationResponseStatus = "invitation_create_failed"
	InvitationResendSuccess    InvitationResponseStatus = "invitation_resend_success"
	InvitationResendFailed     InvitationResponseStatus = "invitation_resend_failed"
	InvitationRevokeSuccess    InvitationResponseStatus = "invitation_revoke_success"
	InvitationRevokeFailed     InvitationResponseStatus = "invitation_revoke_failed"
	InvitationAcceptSuccess    InvitationResponseStatus = "invitation_accept_success"
	InvitationAcceptFailed     InvitationResponseStatus = "invitation_accept_failed"
	// InvitationPending: the address already has an open invitation to the
	// organisation; resend or revoke that one.
	InvitationPending InvitationResponseStatus = "invitation_pending"
	// InvitationAlreadyMember: the address belongs to a member already; grant
	// them access with role assignments instead.
	InvitationAlreadyMember InvitationResponseStatus = "invitation_already_member"
	// InvitationClosed: the invitation was accepted, revoked or has expired.
	InvitationClosed InvitationResponseStatus = "invitation_closed"
	// InvitationEmailMismatch: the signed-in account's email is not the
	// invited address.
	InvitationEmailMismatch InvitationResponseStatus = "invitation_email_mismatch"
	// InvitationExceedsAccess: an assignment goes beyond the inviter's own
	// access (spec 4.5).
	InvitationExceedsAccess InvitationResponseStatus = "invitation_exceeds_access"
	// InvitationRoleNotFound: an assignment names a role that is not one of
	// the organisation's active roles.
	InvitationRoleNotFound InvitationResponseStatus = "invitation_role_not_found"
	// InvitationLegacyMode: the organisation has not been cut over to RBAC;
	// accounts are still created through the legacy account flow.
	InvitationLegacyMode InvitationResponseStatus = "invitation_legacy_mode"
)

// String returns the string representation of the invitation status.
func (s InvitationResponseStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the invitation
// status in the specified language, falling back to English and then to the
// raw value.
func (s InvitationResponseStatus) Translate(lang string) string {
	translations := map[string]map[InvitationResponseStatus]string{
		"en": {
			InvitationBindingFailed:    "Invitation binding failed",
			InvitationValidationFailed: "Invitation validation failed",
			InvitationFound:            "Invitation found",
			InvitationNotFound:         "Invitation not found",
			InvitationGetAllSuccess:    "Invitations retrieved successfully",
			InvitationGetAllFailed:     "Failed to retrieve invitations",
			InvitationCreateSuccess:    "Invitation sent successfully",
			InvitationCreateFailed:     "Failed to send invitation",
			InvitationResendSuccess:    "Invitation sent again",
			InvitationResendFailed:     "Failed to send the invitation again",
			InvitationRevokeSuccess:    "Invitation revoked successfully",
			InvitationRevokeFailed:     "Failed to revoke invitation",
			InvitationAcceptSuccess:    "Invitation accepted",
			InvitationAcceptFailed:     "Failed to accept invitation",
			InvitationPending:          "This address already has an open invitation",
			InvitationAlreadyMember:    "This address already belongs to a member of the organisation",
			InvitationClosed:           "This invitation is no longer valid",
			InvitationEmailMismatch:    "This invitation was sent to another email address",
			InvitationExceedsAccess:    "You cannot grant access you do not have",
			InvitationRoleNotFound:     "Role not found",
			InvitationLegacyMode:       "Members of this organisation are still added through the legacy account flow",
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

// ---------- Request / Response DTOs (invitation domain) ----------
//
// Invitations belong to the caller's current organisation. Inviting needs
// members.invite; each assignment is bounded by the inviter's own access
// (Access.CanGrant), as granting it directly would be. Accepting creates the
// membership and every assignment (source invitation) together.
//
// The raw token is returned once, by CreateInvitation and ResendInvitation,
// in Invitation.Token, for the email link; only its hash is stored.

// InvitationFilter narrows an invitation listing.
type InvitationFilter struct {
	Status *models.InvitationStatus `json:"status,omitempty" bson:"status,omitempty"`
	// Email matches addresses containing it, ignoring case.
	Email *string `json:"email,omitempty" bson:"email,omitempty"`
}

// InvitationInviter is who sent an invitation: identity only.
type InvitationInviter struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// InvitationDetails is an invitation with its inviter and the roles of its
// assignments.
type InvitationDetails struct {
	Invitation models.OrganisationInvitation `json:"invitation"`
	InvitedBy  InvitationInviter             `json:"invitedBy"`
	Roles      []models.Role                 `json:"roles"`
}

// GetInvitations lists the organisation's invitations.
// @Router /invitations [get]
type GetInvitationsRequest struct {
	Filter     *InvitationFilter `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetInvitationsResponse struct {
	Invitations []InvitationDetails `json:"invitations"`
	Pagination  *CursorPagination   `json:"pagination,omitempty"`
}
type GetInvitationsSuccessResponse struct {
	SuccessResponse
	Data GetInvitationsResponse `json:"data"`
}
type GetInvitationsErrorResponse struct {
	ErrorResponse
}

// CreateInvitation invites an address with the organisation and project
// assignments to grant on acceptance; with none, accepting only creates the
// membership. The server normalises the email, sets the inviter, status,
// expiry (ExpiresAt, or the deployment default when unset) and audit, and
// sends the email.
// @Router /invitations [post]
type CreateInvitationRequest struct {
	Email       string                        `json:"email"`
	Assignments []models.InvitationAssignment `json:"assignments,omitempty"`
	ExpiresAt   *time.Time                    `json:"expiresAt,omitempty"`
}
type CreateInvitationResponse struct {
	Invitation InvitationDetails `json:"invitation"`
}
type CreateInvitationSuccessResponse struct {
	SuccessResponse
	Data CreateInvitationResponse `json:"data"`
}
type CreateInvitationErrorResponse struct {
	ErrorResponse
}

// ResendInvitation replaces an open invitation's token, extends its expiry
// and sends the email again; the previous link stops working.
// @Router /invitations/{id}/resend [post]
type ResendInvitationRequest struct {
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}
type ResendInvitationResponse struct {
	Invitation InvitationDetails `json:"invitation"`
}
type ResendInvitationSuccessResponse struct {
	SuccessResponse
	Data ResendInvitationResponse `json:"data"`
}
type ResendInvitationErrorResponse struct {
	ErrorResponse
}

// RevokeInvitation closes an open invitation; its link stops working. Like
// removing access, it is bounded by the inviter rule: the caller must be able
// to grant every assignment it carries.
// @Router /invitations/{id} [delete]
type RevokeInvitationRequest struct{}
type RevokeInvitationResponse struct{}
type RevokeInvitationSuccessResponse struct {
	SuccessResponse
	Data RevokeInvitationResponse `json:"data"`
}
type RevokeInvitationErrorResponse struct {
	ErrorResponse
}

// GetInvitationByToken lets the invitee see what they are invited to before
// signing in or accepting. It needs only the token and returns nothing that
// is not in the email already.
// @Router /invitations/token/{token} [get]
type GetInvitationByTokenRequest struct{}
type GetInvitationByTokenResponse struct {
	OrganisationName string                  `json:"organisationName"`
	InvitedByName    string                  `json:"invitedByName"`
	Email            string                  `json:"email"`
	Status           models.InvitationStatus `json:"status"`
	ExpiresAt        time.Time               `json:"expiresAt"`
}
type GetInvitationByTokenSuccessResponse struct {
	SuccessResponse
	Data GetInvitationByTokenResponse `json:"data"`
}
type GetInvitationByTokenErrorResponse struct {
	ErrorResponse
}

// AcceptInvitation redeems an open invitation for the signed-in user, whose
// email must be the invited address: it creates the membership and every
// assignment together and closes the invitation.
// @Router /invitations/accept [post]
type AcceptInvitationRequest struct {
	Token string `json:"token"`
}
type AcceptInvitationResponse struct {
	Membership      models.OrganisationUser `json:"membership"`
	RoleAssignments []models.RoleAssignment `json:"roleAssignments"`
}
type AcceptInvitationSuccessResponse struct {
	SuccessResponse
	Data AcceptInvitationResponse `json:"data"`
}
type AcceptInvitationErrorResponse struct {
	ErrorResponse
}
