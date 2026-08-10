package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserNotificationMailbox struct {
	Id     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId string             `json:"user_id" bson:"user_id"`
	Data   []Message          `json:"data" bson:"data,omitempty"`
}

type NotificationEvent struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Message `bson:",inline"`
}
