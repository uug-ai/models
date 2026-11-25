package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// SCIM provisioning (RBAC spec 4.16, decision 8). SCIM never creates or takes
// over an account: it records the people and groups of a connection, invites
// people who have not joined, and acts at once only on people already linked
// to the connection.

// IdentityProviderScimUser is a person a connection's provider provisioned
// (collection identity_provider_scim_users): one per (connection, external
// id) and per (connection, user name).
type IdentityProviderScimUser struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId"`
	ConnectionId   primitive.ObjectID `json:"connectionId" bson:"connectionId"`
	// ExternalId is the person's id at the provider; it equals the subject
	// of their identity link once they are linked.
	ExternalId  string `json:"externalId,omitempty" bson:"externalId,omitempty"`
	UserName    string `json:"userName" bson:"userName"`
	Email       string `json:"email,omitempty" bson:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty" bson:"displayName,omitempty"`
	Active      bool   `json:"active" bson:"active"`
	// UserId is the linked account, zero until the person joined.
	UserId primitive.ObjectID `json:"userId,omitempty" bson:"userId,omitempty"`
	// InvitationId is the invitation SCIM sent while the person had not
	// joined.
	InvitationId primitive.ObjectID `json:"invitationId,omitempty" bson:"invitationId,omitempty"`
	// SuspendedMembership is set while SCIM suspended the membership, so
	// reactivating lifts only a suspension SCIM made.
	SuspendedMembership bool  `json:"suspendedMembership,omitempty" bson:"suspendedMembership,omitempty"`
	Audit               Audit `json:"audit" bson:"audit"`
}

// Linked reports whether the person joined: an account is linked.
func (u IdentityProviderScimUser) Linked() bool {
	return !u.UserId.IsZero()
}

// IdentityProviderScimGroup is a group a connection's provider provisioned
// (collection identity_provider_scim_groups), with its members (SCIM user
// ids). A synced team whose externalGroupId is the group's external id
// follows its linked members.
type IdentityProviderScimGroup struct {
	Id             primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID   `json:"organisationId" bson:"organisationId"`
	ConnectionId   primitive.ObjectID   `json:"connectionId" bson:"connectionId"`
	ExternalId     string               `json:"externalId,omitempty" bson:"externalId,omitempty"`
	DisplayName    string               `json:"displayName" bson:"displayName"`
	Members        []primitive.ObjectID `json:"members" bson:"members"`
	Audit          Audit                `json:"audit" bson:"audit"`
}

// GroupKey is what synced teams match: the external id, else the display
// name.
func (g IdentityProviderScimGroup) GroupKey() string {
	if g.ExternalId != "" {
		return g.ExternalId
	}
	return g.DisplayName
}
