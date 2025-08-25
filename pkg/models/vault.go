package models

type Storage struct {
	Uri       string `json:"uri,omitempty" bson:"uri,omitempty"`
	AccessKey string `json:"access_key,omitempty" bson:"access_key,omitempty"`
	Secret    string `json:"secret_key,omitempty" bson:"secret_key,omitempty"`
}