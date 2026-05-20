package api

// OrganisationStatus represents specific status codes for organisation operations.
type OrganisationStatus string

const (
	OrganisationBindingFailed  OrganisationStatus = "organisation_binding_failed"
	OrganisationMissingInfo    OrganisationStatus = "organisation_missing_info"
	OrganisationNameExists     OrganisationStatus = "organisation_name_exists"
	OrganisationFound          OrganisationStatus = "organisation_found"
	OrganisationNotFound       OrganisationStatus = "organisation_not_found"
	OrganisationGetAllSuccess  OrganisationStatus = "organisation_get_all_success"
	OrganisationGetAllFailed   OrganisationStatus = "organisation_get_all_failed"
	OrganisationCreateSuccess  OrganisationStatus = "organisation_create_success"
	OrganisationCreateFailed   OrganisationStatus = "organisation_create_failed"
	OrganisationUpdateSuccess  OrganisationStatus = "organisation_update_success"
	OrganisationUpdateFailed   OrganisationStatus = "organisation_update_failed"
	OrganisationDeleteSuccess  OrganisationStatus = "organisation_delete_success"
	OrganisationDeleteFailed   OrganisationStatus = "organisation_delete_failed"
	OrganisationValidationFail OrganisationStatus = "organisation_validation_failed"
)

// String returns the string representation of the organisation status.
func (s OrganisationStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the organisation
// status in the specified language.
func (s OrganisationStatus) Translate(lang string) string {
	translations := map[string]map[OrganisationStatus]string{
		"en": {
			OrganisationBindingFailed:  "Organisation binding failed",
			OrganisationMissingInfo:    "Organisation is missing required information",
			OrganisationNameExists:     "Organisation name already exists",
			OrganisationFound:          "Organisation found",
			OrganisationNotFound:       "Organisation not found",
			OrganisationGetAllSuccess:  "Organisations retrieved successfully",
			OrganisationGetAllFailed:   "Failed to retrieve organisations",
			OrganisationCreateSuccess:  "Organisation created successfully",
			OrganisationCreateFailed:   "Failed to create organisation",
			OrganisationUpdateSuccess:  "Organisation updated successfully",
			OrganisationUpdateFailed:   "Failed to update organisation",
			OrganisationDeleteSuccess:  "Organisation deleted successfully",
			OrganisationDeleteFailed:   "Failed to delete organisation",
			OrganisationValidationFail: "Organisation validation failed",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[s]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[s]; exists {
			return translation
		}
	}

	return s.String()
}
