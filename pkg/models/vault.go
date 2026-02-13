package models

type Storage struct {
	Uri       string `json:"uri,omitempty" bson:"uri,omitempty"`
	AccessKey string `json:"access_key,omitempty" bson:"access_key,omitempty"`
	Provider  string `json:"provider,omitempty" bson:"provider,omitempty"`
	Secret    string `json:"secret_key,omitempty" bson:"secret_key,omitempty"`
}

type MediaUrlRequest struct {
	Provider      string `json:"provider" bson:"provider"`
	Filename      string `json:"filename" bson:"filename"`
	UriExpiryTime string `json:"uriExpiryTime" bson:"uriExpiryTime"`
}

type MediaUrlResponse struct {
	Data string `json:"data" bson:"data"`
}
