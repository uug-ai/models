package models

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccessTokenScope = Permission

// ErrAccessTokenScopeNotGrantable is returned for a catalog permission that may
// not be granted to an access token.
var ErrAccessTokenScopeNotGrantable = errors.New("permission cannot be granted to an access token")

// accessTokenScopes is an explicit allow-list rather than the whole catalog.
// Access tokens are bound to one project, so organisation permissions are never
// grantable; a project permission joins this list only once hub-api enforces it
// on the routes a token can call.
var accessTokenScopes = appendPermissions(
	markerPermissions,
	mediaPermissions,
	casePermissions,
	workflowPermissions,
	workflowRunPermissions,
	dayPermissions,
	channelPermissions,
)

var accessTokenScopeSet = func() map[AccessTokenScope]struct{} {
	scopes := make(map[AccessTokenScope]struct{}, len(accessTokenScopes))
	for _, scope := range accessTokenScopes {
		scopes[scope] = struct{}{}
	}
	return scopes
}()

// AccessTokenScopes returns every canonical scope that may be granted to an
// access token.
func AccessTokenScopes() []AccessTokenScope {
	return append([]AccessTokenScope(nil), accessTokenScopes...)
}

// ValidateAccessTokenScopes verifies that every scope is canonical, appears at
// most once, and may be granted to an access token.
func ValidateAccessTokenScopes(scopes []AccessTokenScope) error {
	if err := ValidatePermissions(scopes); err != nil {
		return err
	}
	for _, scope := range scopes {
		if _, ok := accessTokenScopeSet[scope]; !ok {
			return fmt.Errorf("%w %q", ErrAccessTokenScopeNotGrantable, scope)
		}
	}
	return nil
}

func appendPermissions(catalogs ...[]Permission) []Permission {
	var permissions []Permission
	for _, catalog := range catalogs {
		permissions = append(permissions, catalog...)
	}
	return permissions
}

type AccessToken struct {
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`

	Name        string             `json:"name" bson:"name,omitempty"`                         // Name of the access token
	Description string             `json:"description,omitempty" bson:"description,omitempty"` // Description of the access token
	Expiration  int64              `json:"expiration,omitempty" bson:"expiration,omitempty"`   // Expiration timestamp of the access token in seconds since epoch
	Scopes      []AccessTokenScope `json:"scopes,omitempty" bson:"scopes,omitempty"`           // List of scopes associated with the access token
	Token       string             `json:"token,omitempty" bson:"token,omitempty"`             // The actual access token value, should be kept secret (will be trimmed before saving)

	// RBAC information
	OrganisationId string              `json:"organisationId" bson:"organisationId,omitempty"` // OrganisationId is used to identify the organisation that owns the marker
	UserId         string              `json:"userId" bson:"userId,omitempty"`                 // UserId is used to identify the user that created the marker
	ProjectId      *primitive.ObjectID `json:"-" bson:"projectId,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the access token
}
