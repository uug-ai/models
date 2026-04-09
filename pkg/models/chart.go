package models

type TimeSeriesDataPoint struct {
	Time  int64 `json:"time" bson:"time,omitempty"`   // x-axis timestamp in milliseconds
	Value int64 `json:"value" bson:"value,omitempty"` // y-axis value
}

type TimeSeries struct {
	Key   string                `json:"key" bson:"key,omitempty"`
	Label string                `json:"label" bson:"label,omitempty"`
	Color string                `json:"color,omitempty" bson:"color,omitempty"`
	Data  []TimeSeriesDataPoint `json:"data" bson:"data,omitempty"`
}

type TimeSeriesChart struct {
	Metric      string       `json:"metric,omitempty" bson:"metric,omitempty"`
	Granularity string       `json:"granularity,omitempty" bson:"granularity,omitempty"` // "hour", "day"
	Timezone    string       `json:"timezone,omitempty" bson:"timezone,omitempty"`
	Series      []TimeSeries `json:"series" bson:"series"`
}
