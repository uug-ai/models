package api

import "github.com/uug-ai/models/pkg/models"

// ChartStatus represents specific status codes for chart operations.
type ChartStatus string

const (
	ChartBindingFailed ChartStatus = "chart_binding_failed"
	ChartMissingInfo   ChartStatus = "chart_missing_info"
	ChartFound         ChartStatus = "chart_found"
	ChartNotFound      ChartStatus = "chart_not_found"
)

// String returns the string representation of the chart status.
func (cs ChartStatus) String() string {
	return string(cs)
}

// Translate returns the translated string representation of the chart status in the specified language.
func (cs ChartStatus) Translate(lang string) string {
	translations := map[string]map[ChartStatus]string{
		"en": {
			ChartBindingFailed: "Chart binding failed",
			ChartMissingInfo:   "Chart missing information",
			ChartFound:         "Chart found",
			ChartNotFound:      "Chart not found",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[cs]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[cs]; exists {
			return translation
		}
	}

	return cs.String()
}

type GetTimeSeriesChartRequest struct {
	Filter *MediaFilter `json:"filter,omitempty" bson:"filter,omitempty"`
}

type GetTimeSeriesChartResponse struct {
	Chart models.TimeSeriesChart `json:"chart" bson:"chart"`
}

type GetTimeSeriesChartSuccessResponse struct {
	SuccessResponse
	Data GetTimeSeriesChartResponse `json:"data" bson:"data"`
}

type GetTimeSeriesChartErrorResponse struct {
	ErrorResponse
}
