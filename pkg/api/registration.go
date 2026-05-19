package api

// RegistrationStatus represents specific status codes for user registration
// performed by an admin (POST /admin/user).
type RegistrationStatus string

const (
	RegistrationBindingFailed      RegistrationStatus = "registration_binding_failed"
	RegistrationMissingInfo        RegistrationStatus = "registration_missing_info"
	RegistrationPasswordsNoMatch   RegistrationStatus = "registration_passwords_no_match"
	RegistrationPasswordTooWeak    RegistrationStatus = "registration_password_too_weak"
	RegistrationUsernameExists     RegistrationStatus = "registration_username_exists"
	RegistrationEmailExists        RegistrationStatus = "registration_email_exists"
	RegistrationCreateUserFailed   RegistrationStatus = "registration_create_user_failed"
	RegistrationCreateOrgFailed    RegistrationStatus = "registration_create_organisation_failed"
	RegistrationHashPasswordFailed RegistrationStatus = "registration_hash_password_failed"
	RegistrationSuccess            RegistrationStatus = "registration_success"
	RegistrationGenerateKeySuccess RegistrationStatus = "registration_generate_key_success"
	RegistrationGenerateKeyFailed  RegistrationStatus = "registration_generate_key_failed"
	RegistrationUpdatePasswordOK   RegistrationStatus = "registration_update_password_success"
	RegistrationUpdatePasswordFail RegistrationStatus = "registration_update_password_failed"
	RegistrationUserIdRequired     RegistrationStatus = "registration_user_id_required"
)

// String returns the string representation of the registration status.
func (s RegistrationStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the registration
// status in the specified language.
func (s RegistrationStatus) Translate(lang string) string {
	translations := map[string]map[RegistrationStatus]string{
		"en": {
			RegistrationBindingFailed:      "Registration binding failed",
			RegistrationMissingInfo:        "Registration is missing required information",
			RegistrationPasswordsNoMatch:   "Passwords do not match",
			RegistrationPasswordTooWeak:    "Password is too weak",
			RegistrationUsernameExists:     "Username already exists",
			RegistrationEmailExists:        "Email already exists",
			RegistrationCreateUserFailed:   "Failed to create user",
			RegistrationCreateOrgFailed:    "Failed to create organisation for user",
			RegistrationHashPasswordFailed: "Failed to hash password",
			RegistrationSuccess:            "Registration successful",
			RegistrationGenerateKeySuccess: "User key generated successfully",
			RegistrationGenerateKeyFailed:  "Failed to generate user key",
			RegistrationUpdatePasswordOK:   "Password updated successfully",
			RegistrationUpdatePasswordFail: "Failed to update password",
			RegistrationUserIdRequired:     "User ID is required",
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
