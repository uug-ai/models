package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PermissionVideowallsRead   Permission = "videowalls.read"
	PermissionVideowallsCreate Permission = "videowalls.create"
	PermissionVideowallsUpdate Permission = "videowalls.update"
	PermissionVideowallsDelete Permission = "videowalls.delete"
	PermissionVideowallsAssign Permission = "videowalls.assign"
)

var videowallPermissions = []Permission{
	PermissionVideowallsRead,
	PermissionVideowallsCreate,
	PermissionVideowallsUpdate,
	PermissionVideowallsDelete,
	PermissionVideowallsAssign,
}

// VideowallPermissions returns the canonical videowall permission catalog.
func VideowallPermissions() []Permission {
	return append([]Permission(nil), videowallPermissions...)
}

type Videowall struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId string             `json:"organisationId" bson:"organisationId,omitempty"`

	// ProjectId optionally places the videowall in a project within its
	// organisation. A nil value places it in the organisation's default
	// project (see ResolveProjectId).
	ProjectId *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`

	Name         string   `json:"name" bson:"name,omitempty"`
	Description  string   `json:"description" bson:"description,omitempty"`
	Sites        []string `json:"sites" bson:"sites"`
	Groups       []string `json:"groups" bson:"groups"`
	Cameras      []string `json:"cameras" bson:"cameras"`
	IsActive     int      `json:"isActive" bson:"isActive"`
	UserId       string   `json:"user_id" bson:"user_id"`
	Username     string   `json:"username" bson:"username,omitempty"`
	MasterUserId string   `json:"master_user_id" bson:"master_user_id"`
	PassCode     string   `json:"pass_code" bson:"pass_code"`
	Fingerprint  string   `json:"fingerprint" bson:"fingerprint"`
	ShortLink    string   `json:"short_link" bson:"short_link,omitempty"`
	Header       int      `json:"header" bson:"header"`
	Expiration   int64    `json:"expiration" bson:"expiration"`
	//ForceMFA    int      `json:"force_mfa" bson:"force_mfa"`
	PTZ            int               `json:"ptz" bson:"ptz"`
	Liveview       int               `json:"liveview" bson:"liveview"`
	IO             int               `json:"io" bson:"io"`
	AssignedUsers  []string          `json:"assigned_users" bson:"assigned_users"`
	WeeklySchedule []*WeeklySchedule `json:"weeklySchedule" bson:"weeklySchedule"`
	// DefaultViewingMode chooses which stream quality the wall loads with:
	// "preview" (SD) or "live" (HD). Empty defaults to "live". Always clamped
	// to "preview" when Liveview only grants preview permission.
	DefaultViewingMode string `json:"default_viewing_mode" bson:"default_viewing_mode,omitempty"`
	// Layout stores a shareable custom arrangement of the wall's camera tiles
	// (order + per-tile column/row spans). When nil the wall falls back to the
	// default responsive grid. It travels with the wall so it is shared with
	// everyone the wall is shared with.
	Layout *VideowallLayout `json:"layout,omitempty" bson:"layout,omitempty"`
	// AllowLayoutEdits lets any signed-in viewer (not just the owner) persist
	// changes to Layout. When false only the owner may save.
	AllowLayoutEdits bool `json:"allow_layout_edits" bson:"allow_layout_edits"`
}

// VideowallLayout is the persisted, shareable tile arrangement of a videowall.
type VideowallLayout struct {
	// Columns is the grid column count the layout was authored against (1-5).
	Columns int `json:"columns" bson:"columns"`
	// Tiles lists each placed camera in display order with its span.
	Tiles []VideowallLayoutTile `json:"tiles" bson:"tiles"`
	// AspectRatio is the shape of a single grid cell as "W:H" (one of
	// VideowallLayoutAspectRatios). Empty means 16:9, so layouts saved before
	// this field existed keep rendering unchanged.
	AspectRatio string `json:"aspectRatio,omitempty" bson:"aspectRatio,omitempty"`
}

// VideowallLayoutAspectRatios lists the accepted VideowallLayout.AspectRatio
// values. The first entry is the default.
var VideowallLayoutAspectRatios = []string{"16:9", "4:3", "1:1"}

// VideowallLayoutTile is a single camera's position and size within a layout.
type VideowallLayoutTile struct {
	CameraKey string `json:"cameraKey" bson:"cameraKey"`
	ColSpan   int    `json:"colSpan" bson:"colSpan"`
	RowSpan   int    `json:"rowSpan" bson:"rowSpan"`
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
