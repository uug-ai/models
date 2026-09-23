package api

// DaysStatus represents status codes for days operations.
type DaysStatus string

const (
	DaysFound           DaysStatus = "days_found"
	DaysRetrievalFailed DaysStatus = "days_retrieval_failed"
)

// String returns the string representation of the days status.
func (status DaysStatus) String() string {
	return string(status)
}

// Translate returns the translated days status.
func (status DaysStatus) Translate(lang string) string {
	translations := map[string]map[DaysStatus]string{
		"en": {
			DaysFound:           "Days found",
			DaysRetrievalFailed: "Failed to retrieve days",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[status]; exists {
			return translation
		}
	}
	if translation, exists := translations["en"][status]; exists {
		return translation
	}
	return status.String()
}

// GetDays
// @Router /media/days [get]
type GetDaysRequest struct{}

type GetDaysResponse struct {
	Days []string `json:"days"`
}

type GetDaysSuccessResponse struct {
	SuccessResponse
	Data GetDaysResponse `json:"data"`
}

type GetDaysErrorResponse struct {
	ErrorResponse
}
