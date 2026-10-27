package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transferring ownership is not a permission: only the owner can do it (see
// Organisation.OwnerId).
const (
	PermissionOrganisationRead   Permission = "organisation.read"
	PermissionOrganisationUpdate Permission = "organisation.update"
)

var organisationPermissions = []Permission{
	PermissionOrganisationRead,
	PermissionOrganisationUpdate,
}

// OrganisationPermissions returns the canonical organisation permission catalog.
func OrganisationPermissions() []Permission {
	return append([]Permission(nil), organisationPermissions...)
}

const (
	PermissionMembersRead   Permission = "members.read"
	PermissionMembersInvite Permission = "members.invite"
	PermissionMembersUpdate Permission = "members.update"
	PermissionMembersRemove Permission = "members.remove"
)

var memberPermissions = []Permission{
	PermissionMembersRead,
	PermissionMembersInvite,
	PermissionMembersUpdate,
	PermissionMembersRemove,
}

// MemberPermissions returns the canonical organisation member permission catalog.
func MemberPermissions() []Permission {
	return append([]Permission(nil), memberPermissions...)
}

// Organisation represents an organization entity that users can belong to.
type Organisation struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name string             `json:"name" bson:"name,omitempty"`
	// Slug is a stable, URL-friendly handle for the organisation (e.g. "acme-corp").
	// It is unique and decoupled from the display Name so it can be used in URLs
	// and references without breaking when the name changes.
	Slug        string `json:"slug" bson:"slug,omitempty"`
	Description string `json:"description" bson:"description,omitempty"`
	Domain      string `json:"domain" bson:"domain,omitempty"`
	// OwnerId is the primary owner: the accountable contact for billing and
	// the contract. Ownership itself is the Owner role (RoleKeyOwner), which
	// several members can hold; the primary owner must always hold it.
	// Changing the primary owner moves this field to another current owner.
	OwnerId  primitive.ObjectID   `json:"ownerId" bson:"ownerId,omitempty"`
	Settings OrganisationSettings `json:"settings" bson:"settings"`
	IsActive bool                 `json:"isActive" bson:"isActive"`

	// Company Details
	Company CompanyDetails `json:"company" bson:"company"`

	// Billing & Subscription
	Subscription   Subscription `json:"subscription" bson:"subscription"`
	BillingAddress Address      `json:"billingAddress" bson:"billingAddress"`

	Audit Audit `json:"audit" bson:"audit"`
}

// IsPrimaryOwner reports whether the user is the organisation's primary owner.
// Other members may also hold the Owner role.
func (o Organisation) IsPrimaryOwner(userId primitive.ObjectID) bool {
	return !o.OwnerId.IsZero() && o.OwnerId == userId
}

// CompanyDetails contains the legal and business information for an organisation.
type CompanyDetails struct {
	LegalName          string `json:"legalName" bson:"legalName,omitempty"`                   // Official registered company name
	TradingName        string `json:"tradingName" bson:"tradingName,omitempty"`               // Trading/DBA name if different
	RegistrationNumber string `json:"registrationNumber" bson:"registrationNumber,omitempty"` // Company registration number
	VATNumber          string `json:"vatNumber" bson:"vatNumber,omitempty"`                   // VAT/Tax ID number
	TaxId              string `json:"taxId" bson:"taxId,omitempty"`                           // Alternative tax identifier
	Industry           string `json:"industry" bson:"industry,omitempty"`                     // Industry/sector
	Website            string `json:"website" bson:"website,omitempty"`                       // Company website
	Phone              string `json:"phone" bson:"phone,omitempty"`                           // Main company phone
	Email              string `json:"email" bson:"email,omitempty"`                           // Main company email
	Logo               string `json:"logo" bson:"logo,omitempty"`                             // Company logo URL
}

// Address represents a physical address.
type Address struct {
	StreetNumber string `json:"streetNumber" bson:"streetNumber,omitempty"`
	Street       string `json:"street" bson:"street,omitempty"`
	Street2      string `json:"street2" bson:"street2,omitempty"` // Additional address line
	City         string `json:"city" bson:"city,omitempty"`
	PostalCode   string `json:"postalCode" bson:"postalCode,omitempty"`
	Region       string `json:"region" bson:"region,omitempty"`           // State/Province/Region
	Country      string `json:"country" bson:"country,omitempty"`         // ISO country code
	CountryName  string `json:"countryName" bson:"countryName,omitempty"` // Full country name
}

// Contact represents a contact person with their details.
type Contact struct {
	Name  string `json:"name" bson:"name,omitempty"`
	Email string `json:"email" bson:"email,omitempty"`
	Phone string `json:"phone" bson:"phone,omitempty"`
	Role  string `json:"role" bson:"role,omitempty"` // Job title/role
}

// OrganisationSettings contains configurable settings for an organisation.
type OrganisationSettings struct {
	ForceMFA         bool     `json:"forceMFA" bson:"forceMFA,omitempty"`
	AllowedDomains   []string `json:"allowedDomains" bson:"allowedDomains,omitempty"` // Email domains allowed for membership
	MaxMembers       int      `json:"maxMembers" bson:"maxMembers,omitempty"`
	AllowInvitations bool     `json:"allowInvitations" bson:"allowInvitations,omitempty"`

	// Regional defaults applied across the organisation for display, scheduling
	// and billing.
	Timezone string `json:"timezone" bson:"timezone,omitempty"` // IANA timezone, e.g. "Europe/Brussels"
	Locale   string `json:"locale" bson:"locale,omitempty"`     // BCP-47 locale tag, e.g. "en-US"
	Currency string `json:"currency" bson:"currency,omitempty"` // ISO 4217 currency code, e.g. "EUR"

	// Contacts for different purposes
	FinancialContact Contact `json:"financialContact" bson:"financialContact"` // Billing/finance contact
	TechnicalContact Contact `json:"technicalContact" bson:"technicalContact"` // Technical/support contact
	PrimaryContact   Contact `json:"primaryContact" bson:"primaryContact"`     // Main point of contact
}

// MembershipStatus enumerates the lifecycle states of an organisation membership.
type MembershipStatus string

const (
	MembershipStatusPending   MembershipStatus = "pending"   // Invited/awaiting acceptance
	MembershipStatusActive    MembershipStatus = "active"    // Full member
	MembershipStatusSuspended MembershipStatus = "suspended" // Temporarily disabled by an admin
	MembershipStatusRevoked   MembershipStatus = "revoked"   // Access permanently removed
)

// OrganisationUser represents a user's membership in an organisation.
// This is the join table between users and organisations, allowing users
// to belong to multiple organisations. Membership grants nothing by itself:
// access comes from RoleAssignment, and a membership that is not effectively
// active disables all of the member's assignments. There is no separate
// project membership; a member is in a project when they hold an assignment
// there.
type OrganisationUser struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId         primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"`
	Status         MembershipStatus   `json:"status" bson:"status,omitempty"`
	InvitedBy      primitive.ObjectID `json:"invitedBy" bson:"invitedBy,omitempty"`
	InvitedAt      time.Time          `json:"invitedAt" bson:"invitedAt,omitempty"`
	JoinedAt       time.Time          `json:"joinedAt" bson:"joinedAt,omitempty"`
	// ExpiresAt optionally bounds the membership in time. It is an overlay on
	// Status: a membership is only effectively active when Status is
	// MembershipStatusActive AND ExpiresAt is unset or in the future.
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt,omitempty"`
	Audit     Audit     `json:"audit" bson:"audit"`
}

// IsExpired reports whether the membership has a set expiry that is in the past.
func (u OrganisationUser) IsExpired() bool {
	return !u.ExpiresAt.IsZero() && u.ExpiresAt.Before(time.Now())
}

// IsEffectivelyActive reports whether the membership is currently in force:
// its status is active and it has not expired. This is the single source of
// truth consumers should use.
func (u OrganisationUser) IsEffectivelyActive() bool {
	return u.Status == MembershipStatusActive && !u.IsExpired()
}

// InvitationStatus enumerates the lifecycle states of an organisation invitation.
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"  // Awaiting acceptance
	InvitationStatusAccepted InvitationStatus = "accepted" // Redeemed by the invitee
	InvitationStatusExpired  InvitationStatus = "expired"  // Passed ExpiresAt without acceptance
	InvitationStatusRevoked  InvitationStatus = "revoked"  // Cancelled by an admin
)

// OrganisationInvitation represents a pending invitation to join an organisation.
type OrganisationInvitation struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"`
	Email          string             `json:"email" bson:"email,omitempty"`
	// Assignments are created together with the membership when the
	// invitation is accepted, each as a RoleAssignment for the invitee with
	// source invitation. They are checked against the inviter's own access
	// when the invitation is created. An invitation without assignments only
	// creates the membership.
	Assignments []InvitationAssignment `json:"assignments" bson:"assignments,omitempty"`
	// Token is returned once when the invitation is created and never stored;
	// TokenHash (see HashInvitationToken) is what lookups match on.
	Token     string             `json:"token,omitempty" bson:"-"`
	TokenHash string             `json:"-" bson:"tokenHash,omitempty"`
	InvitedBy primitive.ObjectID `json:"invitedBy" bson:"invitedBy,omitempty"`
	Status    InvitationStatus   `json:"status" bson:"status,omitempty"`
	ExpiresAt time.Time          `json:"expiresAt" bson:"expiresAt,omitempty"`
	Audit     Audit              `json:"audit" bson:"audit"`
}

// InvitationAssignment is a RoleAssignment to create when an invitation is
// accepted.
type InvitationAssignment struct {
	RoleId     primitive.ObjectID  `json:"roleId" bson:"roleId,omitempty"`
	Scope      RoleAssignmentScope `json:"scope" bson:"scope"`
	TimeWindow *TimeWindow         `json:"timeWindow,omitempty" bson:"timeWindow,omitempty"`
	ExpiresAt  time.Time           `json:"expiresAt" bson:"expiresAt,omitempty"`
}

// HashInvitationToken returns the stored form of an invitation token.
func HashInvitationToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// OrganisationUserDetails is a helper struct that includes full organisation and role details
type OrganisationUserDetails struct {
	UserId          primitive.ObjectID `json:"userId" bson:"userId,omitempty"`
	Membership      OrganisationUser   `json:"membership" bson:"membership,omitempty"`
	Organisation    Organisation       `json:"organisation" bson:"organisation,omitempty"`
	RoleAssignments []RoleAssignment   `json:"roleAssignments" bson:"roleAssignments,omitempty"` // User's role assignments in this organisation
	Roles           []Role             `json:"roles" bson:"roles,omitempty"`                     // Populated role details
}

// OrganisationMember is a helper struct that includes full user details for an organisation member
type OrganisationMember struct {
	OrganisationId  primitive.ObjectID `json:"organisationId" bson:"organisationId,omitempty"`
	Membership      OrganisationUser   `json:"membership" bson:"membership,omitempty"`
	User            User               `json:"user" bson:"user,omitempty"`
	RoleAssignments []RoleAssignment   `json:"roleAssignments" bson:"roleAssignments,omitempty"` // Member's role assignments
	Roles           []Role             `json:"roles" bson:"roles,omitempty"`                     // Populated role details
}

// Input / Output types for organisation repository operations.
//
// These mirror the repository-operation DTO convention used across the domain
// packages (see the workflow and videowall equivalents): each operation takes a
// single typed Input and returns a single typed Output, keeping the repository
// interface stable as fields are added.

// GetOrganisationsInput lists the organisations the caller can see (the ones
// they belong to or own).
type GetOrganisationsInput struct {
	User User `json:"user"`
}

type GetOrganisationsOutput struct {
	Organisations []Organisation `json:"organisations"`
}

// GetOrganisationInput fetches a single organisation by id, scoped to what the
// caller is allowed to access.
type GetOrganisationInput struct {
	User           User   `json:"user"`
	OrganisationId string `json:"organisationId"`
}

type GetOrganisationOutput struct {
	Organisation *Organisation `json:"organisation"`
}

// GetCurrentOrganisationInput resolves the caller's active organisation (the one
// whose canonical id equals the caller's organisation id).
type GetCurrentOrganisationInput struct {
	User User `json:"user"`
}

type GetCurrentOrganisationOutput struct {
	Organisation *Organisation `json:"organisation"`
}

// SetCurrentOrganisationInput selects the organisation used to scope subsequent
// requests for the caller.
type SetCurrentOrganisationInput struct {
	User           User   `json:"user"`
	OrganisationId string `json:"organisationId"`
}

type SetCurrentOrganisationOutput struct {
	Organisation *Organisation `json:"organisation"`
}

// CreateOrganisationInput creates a new organisation owned by the caller.
type CreateOrganisationInput struct {
	User         User         `json:"user"`
	Organisation Organisation `json:"organisation"`
}

type CreateOrganisationOutput struct {
	Organisation *Organisation `json:"organisation"`
}

// OrganisationUpdate is a partial-update (PATCH) payload for an organisation.
// Every field is a pointer so a nil value means "not provided" (leave the stored
// value untouched), which is distinct from an explicit zero value (e.g. clearing
// a string or setting IsActive to false). Identity, ownership and audit fields
// are intentionally absent because they are managed server-side.
type OrganisationUpdate struct {
	Name           *string               `json:"name,omitempty"`
	Slug           *string               `json:"slug,omitempty"`
	Description    *string               `json:"description,omitempty"`
	Domain         *string               `json:"domain,omitempty"`
	IsActive       *bool                 `json:"isActive,omitempty"`
	Settings       *OrganisationSettings `json:"settings,omitempty"`
	Company        *CompanyDetails       `json:"company,omitempty"`
	Subscription   *Subscription         `json:"subscription,omitempty"`
	BillingAddress *Address              `json:"billingAddress,omitempty"`
}

// UpdateOrganisationInput applies a partial update to an organisation the caller
// is allowed to modify. Only the fields set on the OrganisationUpdate payload
// are changed; ownership and audit stamps are managed server-side.
type UpdateOrganisationInput struct {
	User           User               `json:"user"`
	OrganisationId string             `json:"organisationId"`
	Organisation   OrganisationUpdate `json:"organisation"`
}

type UpdateOrganisationOutput struct {
	Organisation *Organisation `json:"organisation"`
}
