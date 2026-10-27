package models

import "go.mongodb.org/mongo-driver/bson/primitive"

const (
	PermissionNotificationsRead   Permission = "notifications.read"
	PermissionNotificationsUpdate Permission = "notifications.update"
)

var notificationPermissions = []Permission{
	PermissionNotificationsRead,
	PermissionNotificationsUpdate,
}

// NotificationPermissions returns the canonical notification permission catalog.
func NotificationPermissions() []Permission {
	return append([]Permission(nil), notificationPermissions...)
}

type UserNotificationMailbox struct {
	Id     primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId string             `json:"user_id" bson:"user_id"`
	Data   []Message          `json:"data" bson:"data,omitempty"`
}

type NotificationEvent struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Message `bson:",inline"`
}
