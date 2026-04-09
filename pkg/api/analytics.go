package api

import "github.com/uug-ai/models/pkg/models"

// AnalyticsStatus represents specific status codes for analytics operations.
type AnalyticsStatus string

const (
	AnalyticsBindingFailed AnalyticsStatus = "analytics_binding_failed"
	AnalyticsMissingInfo   AnalyticsStatus = "analytics_missing_info"
	AnalyticsFound         AnalyticsStatus = "analytics_found"
	AnalyticsNotFound      AnalyticsStatus = "analytics_not_found"
)

// String returns the string representation of the analytics status.
func (as AnalyticsStatus) String() string {
	return string(as)
}

// Translate returns the translated string representation of the analytics status in the specified language.
func (as AnalyticsStatus) Translate(lang string) string {
	translations := map[string]map[AnalyticsStatus]string{
		"en": {
			AnalyticsBindingFailed: "Analytics binding failed",
			AnalyticsMissingInfo:   "Analytics missing information",
			AnalyticsFound:         "Analytics found",
			AnalyticsNotFound:      "Analytics not found",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[as]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[as]; exists {
			return translation
		}
	}

	return as.String()
}

// GetAnalyticsDashboard
// @Router /analytics/dashboard [post]
type GetAnalyticsDashboardRequest struct {
	Filter *models.AnalyticsFilter `json:"filter,omitempty" bson:"filter,omitempty"`
}

type GetAnalyticsDashboardResponse struct {
	Analytics models.AnalyticsDashboard `json:"analytics" bson:"analytics"`
}

type GetAnalyticsDashboardSuccessResponse struct {
	SuccessResponse
	Data GetAnalyticsDashboardResponse `json:"data" bson:"data"`
}

type GetAnalyticsDashboardErrorResponse struct {
	ErrorResponse
}
