package models

// Plan represents a subscription plan with its limits and features
type Plan struct {
	// Level indicates the tier of the plan (1-5, higher is better)
	Level int `json:"level,omitempty" bson:"level,omitempty"`
	// UploadLimit is the maximum number of uploads allowed
	UploadLimit int `json:"uploadLimit,omitempty" bson:"uploadLimit,omitempty"`
	// VideoLimit is the maximum number of videos allowed
	VideoLimit int `json:"videoLimit,omitempty" bson:"videoLimit,omitempty"`
	// Usage is the storage limit in MB
	Usage int `json:"usage,omitempty" bson:"usage,omitempty"`
	// AnalysisLimit is the maximum number of analysis operations allowed
	AnalysisLimit int `json:"analysisLimit,omitempty" bson:"analysisLimit,omitempty"`
	// DayLimit is the retention period in days
	DayLimit int `json:"dayLimit,omitempty" bson:"dayLimit,omitempty"`
}

// Plans is a map of plan names to their configurations
type Plans map[string]Plan

// DefaultPlans returns the default plan configurations
var DefaultPlans = Plans{
	"basic": {
		Level:         1,
		UploadLimit:   100,
		VideoLimit:    100,
		Usage:         500,
		AnalysisLimit: 0,
		DayLimit:      3,
	},
	"premium": {
		Level:         2,
		UploadLimit:   500,
		VideoLimit:    500,
		Usage:         1000,
		AnalysisLimit: 0,
		DayLimit:      7,
	},
	"gold": {
		Level:         3,
		UploadLimit:   1000,
		VideoLimit:    1000,
		Usage:         3000,
		AnalysisLimit: 500,
		DayLimit:      30,
	},
	"business": {
		Level:         4,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         10000,
		AnalysisLimit: 2000,
		DayLimit:      30,
	},
	"enterprise": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         25000,
		AnalysisLimit: 2000,
		DayLimit:      30,
	},
	"corporate": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         500000,
		AnalysisLimit: 0,
		DayLimit:      30,
	},
	"global": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         1000000,
		AnalysisLimit: 1000,
		DayLimit:      60,
	},
	"demo": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         50000,
		AnalysisLimit: 2000,
		DayLimit:      3,
	},
	"unlimited": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         99999999,
		AnalysisLimit: 0,
		DayLimit:      30,
	},
	"50gb": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         50000,
		AnalysisLimit: 0,
		DayLimit:      30,
	},
	"100gb": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         100000,
		AnalysisLimit: 1000,
		DayLimit:      30,
	},
	"200gb": {
		Level:         5,
		UploadLimit:   99999999,
		VideoLimit:    99999999,
		Usage:         200000,
		AnalysisLimit: 0,
		DayLimit:      30,
	},
}
