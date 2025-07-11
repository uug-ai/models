package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type IO struct {
	Id                primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Hash              string             `json:"hash" bson:"hash,omitempty"`
	DeviceId          string             `json:"deviceId" bson:"deviceId,omitempty"` // device identifier
	Type              string             `json:"type" bson:"type,omitempty"`         // input or output
	Key               string             `json:"key" bson:"key,omitempty"`
	Value             string             `json:"value" bson:"value,omitempty"`
	LastSeenTimestamp int64              `json:"lastSeenTimestamp" bson:"lastSeenTimestamp,omitempty"` // last time the IO was seen
	External          bool               `json:"external" bson:"external,omitempty"`
}
