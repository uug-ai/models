package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	Id     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name   string             `json:"name,omitempty"`
	UserId string             `json:"user_id" bson:"user_id,omitempty"`
	Plan   string             `json:"stripe_plan" bson:"stripe_plan,omitempty"`
	Ends   string             `json:"ends_at" bson:"ends_at,omitempty"`
}