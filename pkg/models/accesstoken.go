package models

type AccessTokenScope string

const (
	MarkersWrite AccessTokenScope = "markers:write"
	MarkersRead  AccessTokenScope = "markers:read"
)

type AccessToken struct {
	Name        string             `json:"name" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"` // Description of the access token
	Expiration  int64              `json:"expiration,omitempty" bson:"expiration,omitempty"`   // Expiration timestamp of the access token in seconds since epoch
	Scopes      []AccessTokenScope `json:"scopes,omitempty" bson:"scopes,omitempty"`           // List of scopes associated with the access token
	Token       string             `json:"token,omitempty" bson:"token,omitempty"`             // The actual access token value, should be kept secret (will be trimmed before saving)

	// RBAC information
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"` // OrganisationId is used to identify the organisation that owns the marker
	UserId         string `json:"userId" bson:"userId,omitempty"`                 // UserId is used to identify the user that created the marker

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the access token
}
