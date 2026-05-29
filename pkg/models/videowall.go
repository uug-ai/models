package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Videowall struct {
	Id           primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Name         string             `json:"name" bson:"name,omitempty"`
	Description  string             `json:"description" bson:"description,omitempty"`
	Sites        []string           `json:"sites" bson:"sites"`
	Groups       []string           `json:"groups" bson:"groups"`
	Cameras      []string           `json:"cameras" bson:"cameras"`
	IsActive     int                `json:"isActive" bson:"isActive"`
	UserId       string             `json:"user_id" bson:"user_id"`
	Username     string             `json:"username" bson:"username,omitempty"`
	MasterUserId string             `json:"master_user_id" bson:"master_user_id"`
	PassCode     string             `json:"pass_code" bson:"pass_code"`
	Fingerprint  string             `json:"fingerprint" bson:"fingerprint"`
	ShortLink    string             `json:"short_link" bson:"short_link,omitempty"`
	Header       int                `json:"header" bson:"header"`
	Expiration   int64              `json:"expiration" bson:"expiration"`
	//ForceMFA    int      `json:"force_mfa" bson:"force_mfa"`
	PTZ            int               `json:"ptz" bson:"ptz"`
	Liveview       int               `json:"liveview" bson:"liveview"`
	IO             int               `json:"io" bson:"io"`
	AssignedUsers  []string          `json:"assigned_users" bson:"assigned_users"`
	WeeklySchedule []*WeeklySchedule `json:"weeklySchedule" bson:"weeklySchedule"`
}

// IsScheduledAt reports whether unixTs falls within the videowall's weekly
// schedule. When no schedule is configured the videowall is considered always
// active, preserving the behavior of walls created before scheduling was
// introduced. When a schedule IS configured but every day is disabled, the
// videowall is considered inactive — same semantics as CustomAlert.
func (v *Videowall) IsScheduledAt(unixTs int64) bool {
	if v == nil || len(v.WeeklySchedule) == 0 {
		return true
	}
	ts := time.Unix(unixTs, 0)
	for _, ws := range v.WeeklySchedule {
		if ws == nil || !ws.Enabled {
			continue
		}
		if ws.IsActiveAt(ts) {
			return true
		}
	}
	return false
}

// Input/Output types for repository operations

type GetVideowallsInput struct {
	User User `json:"user"`
}

type GetVideowallsOutput struct {
	Videowalls []Videowall `json:"videowalls"`
}

type GetVideowallInput struct {
	User        User   `json:"user"`
	VideowallId string `json:"videowall_id"`
}

type GetVideowallOutput struct {
	Videowall *Videowall `json:"videowall"`
}

type CreateVideowallInput struct {
	User      User      `json:"user"`
	Videowall Videowall `json:"videowall"`
}

type CreateVideowallOutput struct {
	Videowall *Videowall `json:"videowall"`
}

type UpdateVideowallInput struct {
	User        User      `json:"user"`
	VideowallId string    `json:"videowall_id"`
	Videowall   Videowall `json:"videowall"`
}

type UpdateVideowallOutput struct {
	Videowall *Videowall `json:"videowall"`
}

type PatchVideowallInput struct {
	User        User                   `json:"user"`
	VideowallId string                 `json:"videowall_id"`
	Updates     map[string]interface{} `json:"updates"`
}

type PatchVideowallOutput struct {
	Videowall *Videowall `json:"videowall"`
}

type DeleteVideowallInput struct {
	User        User   `json:"user"`
	VideowallId string `json:"videowall_id"`
}

type DeleteVideowallOutput struct{}
