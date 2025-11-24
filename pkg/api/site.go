package api

import "github.com/uug-ai/models/pkg/models"

// SiteStatus represents specific status codes for device operations
type SiteStatus string

const (
	SiteRetrievalSuccess SiteStatus = "site_retrieval_success"
	SiteBindingFailed    SiteStatus = "device_binding_failed"
	SiteDuplicateName    SiteStatus = "device_duplicate_name"
	SiteMissingInfo      SiteStatus = "device_missing_info"
	SiteRetrievalFailed  SiteStatus = "device_retrieval_failed"
	SiteFound            SiteStatus = "device_found"
	SiteNotFound         SiteStatus = "device_not_found"
	SiteAddSuccess       SiteStatus = "device_add_success"
	SiteAddFailed        SiteStatus = "device_add_failed"
	SiteUpdateSuccess    SiteStatus = "device_update_success"
	SiteUpdateFailed     SiteStatus = "device_update_failed"
	SiteDeleteSuccess    SiteStatus = "device_delete_success"
	SiteDeleteFailed     SiteStatus = "device_delete_failed"
)

// String returns the string representation of the device status
func (ds SiteStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the device status in the specified language
func (ds SiteStatus) Translate(lang string) string {
	translations := map[string]map[SiteStatus]string{
		"en": {
			SiteBindingFailed:    "Site binding failed",
			SiteDuplicateName:    "Site duplicate name",
			SiteMissingInfo:      "Site missing information",
			SiteRetrievalFailed:  "Site retrieval failed",
			SiteFound:            "Site found",
			SiteNotFound:         "Site not found",
			SiteAddSuccess:       "Site added successfully",
			SiteAddFailed:        "Site failed to add",
			SiteUpdateSuccess:    "Site updated successfully",
			SiteUpdateFailed:     "Site failed to update",
			SiteDeleteSuccess:    "Site deleted successfully",
			SiteDeleteFailed:     "Site failed to delete",
			SiteRetrievalSuccess: "Site retrieved successfully",
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

type SiteFilter struct {
	SiteIds []*string `json:"siteIds,omitempty" bson:"siteIds,omitempty"`
	Name    *string   `json:"name,omitempty" bson:"name,omitempty"`
}

type GetSiteOptionsRequest struct {
	Filter     *SiteFilter       `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetSiteOptionsResponse struct {
	Sites []models.SiteOption `json:"sites,omitempty" bson:"sites,omitempty"`
}
type GetSiteOptionsSuccessResponse struct {
	SuccessResponse
	Data GetSiteOptionsResponse `json:"data,omitempty" bson:"data,omitempty"`
}
type GetSiteOptionsErrorResponse struct {
	ErrorResponse
}
