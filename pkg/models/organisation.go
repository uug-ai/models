package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Organisation represents an organization entity that users can belong to.
type Organisation struct {
	Id          primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Name        string               `json:"name" bson:"name,omitempty"`
	Description string               `json:"description" bson:"description,omitempty"`
	Domain      string               `json:"domain" bson:"domain,omitempty"`
	OwnerId     primitive.ObjectID   `json:"owner_id" bson:"owner_id,omitempty"` // The user who owns this organisation
	Settings    OrganisationSettings `json:"settings" bson:"settings,omitempty"`
	IsActive    int                  `json:"is_active" bson:"is_active"`

	// Company Details
	Company CompanyDetails `json:"company" bson:"company,omitempty"`

	// Billing & Subscription
	Subscription   Subscription `json:"subscription" bson:"subscription,omitempty"`
	BillingAddress Address      `json:"billing_address" bson:"billing_address,omitempty"`

	Audit Audit `json:"audit" bson:"audit,omitempty"`
}

// CompanyDetails contains the legal and business information for an organisation.
type CompanyDetails struct {
	LegalName          string `json:"legal_name" bson:"legal_name,omitempty"`                   // Official registered company name
	TradingName        string `json:"trading_name" bson:"trading_name,omitempty"`               // Trading/DBA name if different
	RegistrationNumber string `json:"registration_number" bson:"registration_number,omitempty"` // Company registration number
	VATNumber          string `json:"vat_number" bson:"vat_number,omitempty"`                   // VAT/Tax ID number
	TaxId              string `json:"tax_id" bson:"tax_id,omitempty"`                           // Alternative tax identifier
	Industry           string `json:"industry" bson:"industry,omitempty"`                       // Industry/sector
	Website            string `json:"website" bson:"website,omitempty"`                         // Company website
	Phone              string `json:"phone" bson:"phone,omitempty"`                             // Main company phone
	Email              string `json:"email" bson:"email,omitempty"`                             // Main company email
	Logo               string `json:"logo" bson:"logo,omitempty"`                               // Company logo URL
}

// Address represents a physical address.
type Address struct {
	StreetNumber string `json:"street_number" bson:"street_number,omitempty"`
	Street       string `json:"street" bson:"street,omitempty"`
	Street2      string `json:"street2" bson:"street2,omitempty"` // Additional address line
	City         string `json:"city" bson:"city,omitempty"`
	PostalCode   string `json:"postal_code" bson:"postal_code,omitempty"`
	Region       string `json:"region" bson:"region,omitempty"`             // State/Province/Region
	Country      string `json:"country" bson:"country,omitempty"`           // ISO country code
	CountryName  string `json:"country_name" bson:"country_name,omitempty"` // Full country name
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
	ForceMFA         bool               `json:"force_mfa" bson:"force_mfa,omitempty"`
	AllowedDomains   []string           `json:"allowed_domains" bson:"allowed_domains,omitempty"` // Email domains allowed for membership
	DefaultRoleId    primitive.ObjectID `json:"default_role_id" bson:"default_role_id,omitempty"` // Default role for new members
	MaxMembers       int                `json:"max_members" bson:"max_members,omitempty"`
	AllowInvitations bool               `json:"allow_invitations" bson:"allow_invitations,omitempty"`

	// Contacts for different purposes
	FinancialContact Contact `json:"financial_contact" bson:"financial_contact,omitempty"` // Billing/finance contact
	TechnicalContact Contact `json:"technical_contact" bson:"technical_contact,omitempty"` // Technical/support contact
	PrimaryContact   Contact `json:"primary_contact" bson:"primary_contact,omitempty"`     // Main point of contact
}

// UserOrganisation represents a user's membership in an organisation.
// This is the join table between users and organisations, allowing users
// to belong to multiple organisations. Role assignments are managed separately
// through the RoleAssignment model.
type UserOrganisation struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId         primitive.ObjectID `json:"user_id" bson:"user_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisation_id" bson:"organisation_id,omitempty"`
	Status         string             `json:"status" bson:"status,omitempty"` // "pending", "active", "suspended", "revoked"
	InvitedBy      primitive.ObjectID `json:"invited_by" bson:"invited_by,omitempty"`
	InvitedAt      time.Time          `json:"invited_at" bson:"invited_at,omitempty"`
	JoinedAt       time.Time          `json:"joined_at" bson:"joined_at,omitempty"`
	ExpiresAt      time.Time          `json:"expires_at" bson:"expires_at,omitempty"`   // Optional expiration for temporary access
	Permissions    UserOrgPermissions `json:"permissions" bson:"permissions,omitempty"` // Additional permissions specific to this membership
	Audit          Audit              `json:"audit" bson:"audit,omitempty"`
}

// UserOrgPermissions defines specific permissions a user has within an organisation.
type UserOrgPermissions struct {
	CanInviteUsers   bool     `json:"can_invite_users" bson:"can_invite_users,omitempty"`
	CanManageRoles   bool     `json:"can_manage_roles" bson:"can_manage_roles,omitempty"`
	CanManageDevices bool     `json:"can_manage_devices" bson:"can_manage_devices,omitempty"`
	CanManageSites   bool     `json:"can_manage_sites" bson:"can_manage_sites,omitempty"`
	CanManageGroups  bool     `json:"can_manage_groups" bson:"can_manage_groups,omitempty"`
	SiteIds          []string `json:"site_ids" bson:"site_ids,omitempty"`     // Specific sites user has access to
	GroupIds         []string `json:"group_ids" bson:"group_ids,omitempty"`   // Specific groups user has access to
	DeviceIds        []string `json:"device_ids" bson:"device_ids,omitempty"` // Specific devices user has access to
}

// OrganisationInvitation represents a pending invitation to join an organisation.
type OrganisationInvitation struct {
	Id             primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID   `json:"organisation_id" bson:"organisation_id,omitempty"`
	Email          string               `json:"email" bson:"email,omitempty"`
	RoleIds        []primitive.ObjectID `json:"role_ids" bson:"role_ids,omitempty"` // Roles to assign upon acceptance
	Token          string               `json:"token" bson:"token,omitempty"`
	InvitedBy      primitive.ObjectID   `json:"invited_by" bson:"invited_by,omitempty"`
	Status         string               `json:"status" bson:"status,omitempty"` // "pending", "accepted", "expired", "revoked"
	ExpiresAt      time.Time            `json:"expires_at" bson:"expires_at,omitempty"`
	Audit          Audit                `json:"audit" bson:"audit,omitempty"`
}

// UserOrganisationDetails is a helper struct that includes full organisation and role details
type UserOrganisationDetails struct {
	UserId          primitive.ObjectID `json:"user_id" bson:"user_id,omitempty"`
	Membership      UserOrganisation   `json:"membership" bson:"membership,omitempty"`
	Organisation    Organisation       `json:"organisation" bson:"organisation,omitempty"`
	RoleAssignments []RoleAssignment   `json:"role_assignments" bson:"role_assignments,omitempty"` // User's role assignments in this organisation
	Roles           []Role             `json:"roles" bson:"roles,omitempty"`                       // Populated role details
}

// OrganisationMember is a helper struct that includes full user details for an organisation member
type OrganisationMember struct {
	OrganisationId  primitive.ObjectID `json:"organisation_id" bson:"organisation_id,omitempty"`
	Membership      UserOrganisation   `json:"membership" bson:"membership,omitempty"`
	User            User               `json:"user" bson:"user,omitempty"`
	RoleAssignments []RoleAssignment   `json:"role_assignments" bson:"role_assignments,omitempty"` // Member's role assignments
	Roles           []Role             `json:"roles" bson:"roles,omitempty"`                       // Populated role details
}
