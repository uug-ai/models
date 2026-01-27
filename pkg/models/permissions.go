package models

type Permissions struct {
	Pages    []string           `json:"pages" bson:"pages,omitempty"`
	Features FeaturePermissions `json:"featurePermissions" bson:"featurePermissions"`
}
