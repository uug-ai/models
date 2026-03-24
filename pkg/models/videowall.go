package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Videowall struct {
	Id           primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name         string             `json:"name" bson:"name,omitempty"`
	Description  string             `json:"description" bson:"description,omitempty"`
	Sites        []string           `json:"sites" bson:"sites"`
	Groups       []string           `json:"groups" bson:"groups"`
	Cameras      []string           `json:"cameras" bson:"cameras"`
	IsActive     int                `json:"isActive" bson:"isActive"`
	UserId       string             `json:"user_id" bson:"user_id"`
	MasterUserId string             `json:"master_user_id" bson:"master_user_id"`
	PassCode     string             `json:"pass_code" bson:"pass_code"`
	Fingerprint  string             `json:"fingerprint" bson:"fingerprint"`
	ShortLink    string             `json:"short_link" bson:"short_link,omitempty"`
	Header       int                `json:"header" bson:"header"`
	Expiration   int64              `json:"expiration" bson:"expiration"`
	//ForceMFA    int      `json:"force_mfa" bson:"force_mfa"`
	PTZ           int      `json:"ptz" bson:"ptz"`
	Liveview      int      `json:"liveview" bson:"liveview"`
	IO            int      `json:"io" bson:"io"`
	AssignedUsers []string `json:"assigned_users" bson:"assigned_users"`
}
