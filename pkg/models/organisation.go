package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Organisation represents an organization entity that users can belong to.
type Organisation struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty"`
	Description string             `json:"description" bson:"description,omitempty"`
	OwnerId     string             `json:"ownerId" bson:"ownerId,omitempty"`
	IsActive    int                `json:"isActive" bson:"isActive"`

	// Credentials
	PublicKey  string `json:"publicKey" bson:"publicKey,omitempty"`
	PrivateKey string `json:"privateKey" bson:"privateKey,omitempty"`

	// Subscription plan
	Subscription Subscription `json:"subscription" bson:"subscription,omitempty"`

	// Company Details
	Company        CompanyDetails `json:"company" bson:"company,omitempty"`
	BillingAddress Address        `json:"billingAddress" bson:"billingAddress,omitempty"`

	// Additional metadata
	Metadata *OrganisationMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the organisation information, something that doesnt need to be stored in the database,
	// for example the roles assignments, role permissions
	AtRuntimeMetadata *OrganisationAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

type OrganisationMetadata struct {
	Domain           string             `json:"domain" bson:"domain,omitempty"`
	ForceMFA         bool               `json:"forceMFA" bson:"forceMFA,omitempty"`
	AllowedDomains   []string           `json:"allowedDomains" bson:"allowedDomains,omitempty"` // Email domains allowed for membership
	DefaultRoleId    primitive.ObjectID `json:"defaultRoleId" bson:"defaultRoleId,omitempty"`   // Default role for new members
	MaxMembers       int                `json:"maxMembers" bson:"maxMembers,omitempty"`
	AllowInvitations bool               `json:"allowInvitations" bson:"allowInvitations,omitempty"`

	// Subscription information
	ReachedLimit          bool  `json:"reachedLimit" bson:"reachedLimit,omitempty"`
	ReachedLimitTimestamp int64 `json:"reachedLimitTimestamp" bson:"reachedLimitTimestamp,omitempty"`

	// Contacts for different purposes
	FinancialContact Contact `json:"financialContact" bson:"financialContact,omitempty"` // Billing/finance contact
	TechnicalContact Contact `json:"technicalContact" bson:"technicalContact,omitempty"` // Technical/support contact
	PrimaryContact   Contact `json:"primaryContact" bson:"primaryContact,omitempty"`     // Main point of contact
}

type OrganisationAtRuntimeMetadata struct {
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

// UserOrganisation represents a user's membership in an organisation.
// This is the join table between users and organisations, allowing users
// to belong to multiple organisations. Role assignments are managed separately
// through the RoleAssignment model.
type UserOrganisation struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId         string             `json:"userId" bson:"userId,omitempty"`
	OrganisationId string             `json:"organisationId" bson:"organisationId,omitempty"`
	Status         string             `json:"status" bson:"status,omitempty"` // "pending", "active", "suspended", "revoked"
	InvitedBy      string             `json:"invitedBy" bson:"invitedBy,omitempty"`
	InvitedAt      time.Time          `json:"invitedAt" bson:"invitedAt,omitempty"`
	JoinedAt       time.Time          `json:"joinedAt" bson:"joinedAt,omitempty"`
	ExpiresAt      time.Time          `json:"expiresAt" bson:"expiresAt,omitempty"`     // Optional expiration for temporary access
	Permissions    UserOrgPermissions `json:"permissions" bson:"permissions,omitempty"` // Additional permissions specific to this membership
	Audit          Audit              `json:"audit" bson:"audit,omitempty"`
}

// UserOrgPermissions defines specific permissions a user has within an organisation.
type UserOrgPermissions struct {
	CanInviteUsers   bool     `json:"canInviteUsers" bson:"canInviteUsers,omitempty"`
	CanManageRoles   bool     `json:"canManageRoles" bson:"canManageRoles,omitempty"`
	CanManageDevices bool     `json:"canManageDevices" bson:"canManageDevices,omitempty"`
	CanManageSites   bool     `json:"canManageSites" bson:"canManageSites,omitempty"`
	CanManageGroups  bool     `json:"canManageGroups" bson:"canManageGroups,omitempty"`
	SiteIds          []string `json:"siteIds" bson:"siteIds,omitempty"`     // Specific sites user has access to
	GroupIds         []string `json:"groupIds" bson:"groupIds,omitempty"`   // Specific groups user has access to
	DeviceIds        []string `json:"deviceIds" bson:"deviceIds,omitempty"` // Specific devices user has access to
}

// OrganisationInvitation represents a pending invitation to join an organisation.
type OrganisationInvitation struct {
	Id             primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	OrganisationId string               `json:"organisationId" bson:"organisationId,omitempty"`
	Email          string               `json:"email" bson:"email,omitempty"`
	RoleIds        []primitive.ObjectID `json:"roleIds" bson:"roleIds,omitempty"` // Roles to assign upon acceptance
	Token          string               `json:"token" bson:"token,omitempty"`
	InvitedBy      primitive.ObjectID   `json:"invitedBy" bson:"invitedBy,omitempty"`
	Status         string               `json:"status" bson:"status,omitempty"` // "pending", "accepted", "expired", "revoked"
	ExpiresAt      time.Time            `json:"expiresAt" bson:"expiresAt,omitempty"`
	Audit          Audit                `json:"audit" bson:"audit,omitempty"`
}

// UserOrganisationDetails is a helper struct that includes full organisation and role details
type UserOrganisationDetails struct {
	UserId          string           `json:"userId" bson:"userId,omitempty"`
	Membership      UserOrganisation `json:"membership" bson:"membership,omitempty"`
	Organisation    Organisation     `json:"organisation" bson:"organisation,omitempty"`
	RoleAssignments []RoleAssignment `json:"roleAssignments" bson:"roleAssignments,omitempty"` // User's role assignments in this organisation
	Roles           []Role           `json:"roles" bson:"roles,omitempty"`                     // Populated role details
}

// OrganisationMember is a helper struct that includes full user details for an organisation member
type OrganisationMember struct {
	OrganisationId  string           `json:"organisationId" bson:"organisationId,omitempty"`
	Membership      UserOrganisation `json:"membership" bson:"membership,omitempty"`
	User            User             `json:"user" bson:"user,omitempty"`
	RoleAssignments []RoleAssignment `json:"roleAssignments" bson:"roleAssignments,omitempty"` // Member's role assignments
	Roles           []Role           `json:"roles" bson:"roles,omitempty"`                     // Populated role details
}
