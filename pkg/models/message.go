package models

type Message struct {
	Type              string            `json:"type,omitempty" bson:"type,omitempty"`
	Id                string            `json:"id,omitempty" bson:"id,omitempty"`
	AlertId           string            `json:"alert_id,omitempty" bson:"alert_id,omitempty"`
	AlertName         string            `json:"alert_name,omitempty" bson:"alert_name,omitempty"`
	AlertUser         string            `json:"alert_user,omitempty" bson:"alert_user,omitempty"`
	AlertMasterUser   string            `json:"alert_master_user,omitempty" bson:"alert_master_user,omitempty"`
	Timestamp         int64             `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	NotificationType  string            `json:"notification_type,omitempty" bson:"notification_type,omitempty"` // generic, counting, region
	Title             string            `json:"title,omitempty" bson:"title,omitempty"`
	Body              string            `json:"body,omitempty" bson:"body,omitempty"`
	Unread            bool              `json:"unread,omitempty" bson:"unread,omitempty"`
	User              string            `json:"user,omitempty" bson:"user,omitempty"`
	UserId            string            `json:"userid,omitempty" bson:"userid,omitempty"`
	Timezone          string            `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Email             string            `json:"email,omitempty" bson:"email,omitempty"`
	SequenceId        string            `json:"sequence_id,omitempty" bson:"sequence_id,omitempty"`
	Thumbnail         string            `json:"thumbnail,omitempty" bson:"thumbnail,omitempty"`
	ThumbnailFile     string            `json:"thumbnailFile,omitempty" bson:"thumbnailFile,omitempty"`
	ThumbnailProvider string            `json:"thumbnailProvider,omitempty" bson:"thumbnailProvider,omitempty"`
	SpriteFile        string            `json:"spriteFile,omitempty" bson:"spriteFile,omitempty"`
	SpriteInterval    int               `json:"spriteInterval,omitempty" bson:"spriteInterval,omitempty"`
	SpriteProvider    string            `json:"spriteProvider,omitempty" bson:"spriteProvider,omitempty"`
	DeviceId          string            `json:"device_id,omitempty" bson:"device_id,omitempty"`
	DeviceName        string            `json:"device_name,omitempty" bson:"device_name,omitempty"`
	Classifications   []string          `json:"classifications,omitempty" bson:"classifications,omitempty"`
	Sites             []Site            `json:"sites,omitempty" bson:"sites,omitempty"`
	Groups            []Group           `json:"groups,omitempty" bson:"groups,omitempty"`
	MediaKey          string            `json:"media_key,omitempty" bson:"media_key,omitempty"`
	MediaProvider     string            `json:"media_provider,omitempty" bson:"media_provider,omitempty"`
	MediaSource       string            `json:"media_source,omitempty" bson:"media_source,omitempty"`
	Media             []Media           `json:"media,omitempty" bson:"media,omitempty"`
	NumberOfMedia     string            `json:"number_of_media,omitempty" bson:"number_of_media,omitempty"`
	DataUsage         string            `json:"data_usage,omitempty" bson:"data_usage,omitempty"`
	Data              map[string]string `json:"data,omitempty" bson:"data,omitempty"`
}
type Media struct {
	Timestamp    int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Type         string `json:"type,omitempty" bson:"type,omitempty"`
	Url          string `json:"url,omitempty" bson:"url,omitempty"`
	ThumbnailUrl string `json:"thumbnail_url,omitempty" bson:"thumbnail_url,omitempty"`
	SpriteUrl    string `json:"sprite_url,omitempty" bson:"sprite_url,omitempty"`
}
