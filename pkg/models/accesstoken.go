package models

const (
	ACCESSTOKEN_BINDING_FAILED string = "Failed to bind access token from request body"
	ACCESSTOKEN_NAME_EXISTS    string = "Access token with the same name already exists"
	ACCESSTOKEN_MISSING_INFO   string = "Access token is missing required information"
	ACCESSTOKEN_FOUND          string = "One or more access tokens where found"
	ACCESSTOKEN_NOT_FOUND      string = "One or more access tokens not found, returning empty list"
	ACCESSTOKEN_ADD_SUCCESS    string = "Access token added successfully"
	ACCESSTOKEN_ADD_FAILED     string = "Failed to add access token"
	ACCESSTOKEN_UPDATE_SUCCESS string = "Access token updated successfully"
	ACCESSTOKEN_UPDATE_FAILED  string = "Failed to update access token"
	ACCESSTOKEN_DELETE_SUCCESS string = "Access token deleted successfully"
	ACCESSTOKEN_DELETE_FAILED  string = "Failed to delete access token"
)

type AccessToken struct {
	Name        string   `json:"name" bson:"name,omitempty"`
	Description string   `json:"description,omitempty" bson:"description,omitempty"` // Description of the access token
	Expiration  int64    `json:"expiration,omitempty" bson:"expiration,omitempty"`   // Expiration timestamp of the access token in seconds since epoch
	Scopes      []string `json:"scopes,omitempty" bson:"scopes,omitempty"`           // List of scopes associated with the access token
	Token       string   `json:"token,omitempty" bson:"token,omitempty"`             // The actual access token value, should be kept secret (will be trimmed before saving)

	// RBAC information
	OrganisationId string `json:"organisationId" bson:"organisationId,omitempty"` // OrganisationId is used to identify the organisation that owns the marker
	UserId         string `json:"userId" bson:"userId,omitempty"`                 // UserId is used to identify the user that created the marker

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the access token
}
