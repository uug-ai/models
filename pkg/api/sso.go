package api

// GroupStatus represents specific status codes for group operations
type SingleSignOnStatus string

const (
	SingleSignOnDomainsSuccess SingleSignOnStatus = "sso_domains_retrieval_success"
)

// String returns the string representation of the group status
func (ds SingleSignOnStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the group status in the specified language
func (ds SingleSignOnStatus) Translate(lang string) string {
	translations := map[string]map[SingleSignOnStatus]string{
		"en": {
			SingleSignOnDomainsSuccess: "Single sign-on domains retrieval successful",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ds]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ds]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ds.String()
}

type SingleSignOnDomainsRequest struct {
}

type SingleSignOnDomainsResponse struct {
	Domains         []string `json:"domains" bson:"domains"`
	ForceSSODomains []string `json:"force_sso_domains" bson:"force_sso_domains"`
}

type GetSingleSignOnDomainsSuccessResponse struct {
	SuccessResponse
	Data SingleSignOnDomainsResponse `json:"data" bson:"data"`
}
type GetSingleSignOnDomainsErrorResponse struct {
	ErrorResponse
}
