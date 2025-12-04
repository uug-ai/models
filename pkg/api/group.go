package api

import "github.com/uug-ai/models/pkg/models"

// GroupStatus represents specific status codes for group operations
type GroupStatus string

const (
	GroupRetrievalSuccess GroupStatus = "group_retrieval_success"
	GroupBindingFailed    GroupStatus = "group_binding_failed"
	GroupDuplicateName    GroupStatus = "group_duplicate_name"
	GroupMissingInfo      GroupStatus = "group_missing_info"
	GroupRetrievalFailed  GroupStatus = "group_retrieval_failed"
	GroupFound            GroupStatus = "group_found"
	GroupNotFound         GroupStatus = "group_not_found"
	GroupAddSuccess       GroupStatus = "group_add_success"
	GroupAddFailed        GroupStatus = "group_add_failed"
	GroupUpdateSuccess    GroupStatus = "group_update_success"
	GroupUpdateFailed     GroupStatus = "group_update_failed"
	GroupDeleteSuccess    GroupStatus = "group_delete_success"
	GroupDeleteFailed     GroupStatus = "group_delete_failed"
)

// String returns the string representation of the group status
func (ds GroupStatus) String() string {
	return string(ds)
}

// Into returns the translated string representation of the group status in the specified language
func (ds GroupStatus) Translate(lang string) string {
	translations := map[string]map[GroupStatus]string{
		"en": {
			GroupBindingFailed:    "Groups binding failed",
			GroupDuplicateName:    "Groups duplicate name",
			GroupMissingInfo:      "Groups missing information",
			GroupRetrievalFailed:  "Groups retrieval failed",
			GroupFound:            "Groups found",
			GroupNotFound:         "Groups not found",
			GroupAddSuccess:       "Groups added successfully",
			GroupAddFailed:        "Groups failed to add",
			GroupUpdateSuccess:    "Groups updated successfully",
			GroupUpdateFailed:     "Groups failed to update",
			GroupDeleteSuccess:    "Groups deleted successfully",
			GroupDeleteFailed:     "Groups failed to delete",
			GroupRetrievalSuccess: "Groups retrieved successfully",
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

type GroupFilter struct {
	GroupIds []*string `json:"groupIds,omitempty" bson:"groupIds,omitempty"`
	Name     *string   `json:"name,omitempty" bson:"name,omitempty"`
}

type GetGroupOptionsRequest struct {
	Filter     *GroupFilter      `json:"filter,omitempty" bson:"filter,omitempty"`
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}
type GetGroupOptionsResponse struct {
	Groups []models.GroupOption `json:"groups,omitempty" bson:"groups,omitempty"`
}
type GetGroupOptionsSuccessResponse struct {
	SuccessResponse
	Data GetGroupOptionsResponse `json:"data,omitempty" bson:"data,omitempty"`
}
type GetGroupOptionsErrorResponse struct {
	ErrorResponse
}
