package models

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
	RecordingsPerHour          TimeSeriesChart           `json:"recordingsPerHour" bson:"recordingsPerHour"`
	CountsPerHourByDevice      TimeSeriesChart           `json:"countsPerHourByDevice" bson:"countsPerHourByDevice"`
	CountsPerHourByAlert       DirectionalTimeSeriesChart `json:"countsPerHourByAlert" bson:"countsPerHourByAlert"`
	RegionDurationPerHourAlert TimeSeriesChart           `json:"regionDurationPerHourAlert" bson:"regionDurationPerHourAlert"`
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
