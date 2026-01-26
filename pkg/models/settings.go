package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Settings struct {
	Id  primitive.ObjectID `json:"id" bson:"_id"`
	Key string             `json:"key" bson:"key"`
	Map map[string]any     `json:"map" bson:"map"` // @TODO replace this
}
