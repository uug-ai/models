package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// AnalyticsFilter defines the shared filter contract for analytics dashboard data.
// It mirrors the analysis page selectors and keeps the request aligned with media filters.
type AnalyticsFilter struct {
	TimeRanges []*TimeRange `json:"timeRanges,omitempty" bson:"timeRanges,omitempty"`
	Sites      []*string    `json:"sites,omitempty" bson:"sites,omitempty"`
	Groups     []*string    `json:"groups,omitempty" bson:"groups,omitempty"`
	Devices    []*string    `json:"devices,omitempty" bson:"devices,omitempty"`
}

// AnalyticsHours contains recording counts per hour for the whole selection and per device/instance.
type AnalyticsHours struct {
	Total     []int            `json:"total,omitempty" bson:"total,omitempty"`
	Instances map[string][]int `json:"instances,omitempty" bson:"instances,omitempty"`
}

// AnalyticsDashboard groups the typed data needed by the analysis page.
// The charts section is designed around the internal chart component.
type AnalyticsDashboard struct {
	Summary AnalyticsSummary        `json:"summary" bson:"summary"`
	Alerts  []AnalyticsAlertSummary `json:"alerts,omitempty" bson:"alerts,omitempty"`
	Lists   AnalyticsLists          `json:"lists" bson:"lists"`
	Charts  AnalyticsCharts         `json:"charts" bson:"charts"`
}

// AnalyticsSummary contains the KPI data shown at the top of the analysis page.
type AnalyticsSummary struct {
	TotalRecordings            int64   `json:"totalRecordings,omitempty" bson:"totalRecordings,omitempty"`
	TotalCounts                int64   `json:"totalCounts,omitempty" bson:"totalCounts,omitempty"`
	TotalRegions               int64   `json:"totalRegions,omitempty" bson:"totalRegions,omitempty"`
	TotalRegionDurationSeconds float64 `json:"totalRegionDurationSeconds,omitempty" bson:"totalRegionDurationSeconds,omitempty"`
	TotalRegionDurationLabel   string  `json:"totalRegionDurationLabel,omitempty" bson:"totalRegionDurationLabel,omitempty"`
}

// AnalyticsCharts contains the time-series charts rendered on the analysis page.
type AnalyticsCharts struct {
	RecordingsPerHour          TimeSeriesChart            `json:"recordingsPerHour" bson:"recordingsPerHour"`
	CountsPerHourByDevice      TimeSeriesChart            `json:"countsPerHourByDevice" bson:"countsPerHourByDevice"`
	CountsPerHourByAlert       DirectionalTimeSeriesChart `json:"countsPerHourByAlert" bson:"countsPerHourByAlert"`
	RegionDurationPerHourAlert TimeSeriesChart            `json:"regionDurationPerHourAlert" bson:"regionDurationPerHourAlert"`
}

// DirectionalTimeSeriesChart stores parallel chart variants for all/in/out views.
type DirectionalTimeSeriesChart struct {
	All TimeSeriesChart `json:"all" bson:"all"`
	In  TimeSeriesChart `json:"in" bson:"in"`
	Out TimeSeriesChart `json:"out" bson:"out"`
}

// AnalyticsAlertSummary powers the per-alert KPI cards on the analysis page.
type AnalyticsAlertSummary struct {
	AlertId         string  `json:"alertId,omitempty" bson:"alertId,omitempty"`
	AlertLabel      string  `json:"alertLabel,omitempty" bson:"alertLabel,omitempty"`
	Type            string  `json:"type,omitempty" bson:"type,omitempty"`
	Total           int64   `json:"total,omitempty" bson:"total,omitempty"`
	Count           int64   `json:"count,omitempty" bson:"count,omitempty"`
	In              int64   `json:"in,omitempty" bson:"in,omitempty"`
	Out             int64   `json:"out,omitempty" bson:"out,omitempty"`
	RegionIn        int64   `json:"regionIn,omitempty" bson:"regionIn,omitempty"`
	DurationSeconds float64 `json:"durationSeconds,omitempty" bson:"durationSeconds,omitempty"`
	DurationLabel   string  `json:"durationLabel,omitempty" bson:"durationLabel,omitempty"`
}

// AnalyticsLists groups the recent event tables shown below the charts.
type AnalyticsLists struct {
	CountEvents  []AnalyticsEvent `json:"countEvents,omitempty" bson:"countEvents,omitempty"`
	RegionEvents []AnalyticsEvent `json:"regionEvents,omitempty" bson:"regionEvents,omitempty"`
}

// AnalyticsEvent represents a row in one of the analytics event lists.
type AnalyticsEvent struct {
	Key             string  `json:"key,omitempty" bson:"key,omitempty"`
	Timestamp       int64   `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	DeviceId        string  `json:"deviceId,omitempty" bson:"deviceId,omitempty"`
	DeviceLabel     string  `json:"deviceLabel,omitempty" bson:"deviceLabel,omitempty"`
	AlertId         string  `json:"alertId,omitempty" bson:"alertId,omitempty"`
	AlertLabel      string  `json:"alertLabel,omitempty" bson:"alertLabel,omitempty"`
	SequenceId      string  `json:"sequenceId,omitempty" bson:"sequenceId,omitempty"`
	ObjectId        string  `json:"objectId,omitempty" bson:"objectId,omitempty"`
	ObjectLabel     string  `json:"objectLabel,omitempty" bson:"objectLabel,omitempty"`
	Type            string  `json:"type,omitempty" bson:"type,omitempty"`
	Count           int64   `json:"count,omitempty" bson:"count,omitempty"`
	DurationSeconds float64 `json:"durationSeconds,omitempty" bson:"durationSeconds,omitempty"`
	DurationLabel   string  `json:"durationLabel,omitempty" bson:"durationLabel,omitempty"`
}

type AnalyticsCount struct {
	Id             primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	OrganisationId string              `json:"organisationId" bson:"organisationId,omitempty"`
	ProjectId      *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`
	Key            string              `json:"key,omitempty" bson:"key,omitempty"`
	Timestamp      int64               `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Username       string              `json:"username,omitempty" bson:"username,omitempty"`
	UserId         string              `json:"user_id,omitempty" bson:"user_id,omitempty"`
	DeviceId       string              `json:"device_id,omitempty" bson:"device_id,omitempty"`
	SegmentId      string              `json:"segment_id,omitempty" bson:"segment_id,omitempty"`
	ObjectId       string              `json:"object_id,omitempty" bson:"object_id,omitempty"`
	ObjectName     string              `json:"object_name,omitempty" bson:"object_name,omitempty"`
	Count          int                 `json:"count,omitempty" bson:"count,omitempty"`
	Duration       float64             `json:"duration,omitempty" bson:"duration,omitempty"`
	AlertId        string              `json:"alert_id,omitempty" bson:"alert_id,omitempty"`
	AlertName      string              `json:"alert_name,omitempty" bson:"alert_name,omitempty"`
	SequenceId     string              `json:"sequence_id,omitempty" bson:"sequence_id,omitempty"`
	Type           string              `json:"type,omitempty" bson:"type,omitempty"`
}

type Heatmap struct {
	Id             primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	OrganisationId string              `json:"organisationId" bson:"organisationId,omitempty"`
	ProjectId      *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`
	Key            string              `json:"key" bson:"key"`
	Timestamp      int64               `json:"timestamp" bson:"timestamp"`
	UserId         string              `json:"user_id" bson:"user_id"`
	DeviceId       string              `json:"device_id" bson:"device_id"`
	FrameWidth     int                 `json:"frame_width" bson:"frame_width"`
	FrameHeight    int                 `json:"frame_height" bson:"frame_height"`
	Coordinates    [][]int             `json:"coordinates" bson:"coordinates"`
}
