package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Sequence struct {
	Id       primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	Start    int64                  `json:"start,omitempty"`
	End      int64                  `json:"end,omitempty"`
	UserId   string                 `json:"user_id" bson:"user_id,omitempty"`
	Images   []Media                `json:"images,omitempty"`
	Analysis map[string]interface{} `bson:"analysis,omitempty"`
	Devices  []string               `json:"devices,omitempty"`
	Notified bool                   `json:"notified,omitempty"`
}
