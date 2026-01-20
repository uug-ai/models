package api

// UserStatus represents specific status codes for user operations
type UserStatus string

const (
	UserBindingFailed         UserStatus = "user_binding_failed"
	UserDuplicateName         UserStatus = "user_duplicate_name"
	UserMissingInfo           UserStatus = "user_missing_info"
	UserFound                 UserStatus = "user_found"
	UserNotFound              UserStatus = "user_not_found"
	UserAddSuccess            UserStatus = "user_add_success"
	UserAddFailed             UserStatus = "user_add_failed"
	UserUpdateSuccess         UserStatus = "user_update_success"
	UserUpdateFailed          UserStatus = "user_update_failed"
	UserDeleteSuccess         UserStatus = "user_delete_success"
	UserDeleteFailed          UserStatus = "user_delete_failed"
	UserFetchByUsernameFailed UserStatus = "user_fetch_by_username_failed"
)

// String returns the string representation of the User status
func (ms UserStatus) String() string {
	return string(ms)
}

// Into returns the translated string representation of the User status in the specified language
func (ms UserStatus) Translate(lang string) string {
	translations := map[string]map[UserStatus]string{
		"en": {
			UserBindingFailed:         "User binding failed",
			UserDuplicateName:         "User duplicate name",
			UserMissingInfo:           "User missing information",
			UserFound:                 "User found",
			UserNotFound:              "User not found",
			UserAddSuccess:            "User added successfully",
			UserAddFailed:             "User failed to add",
			UserUpdateSuccess:         "User updated successfully",
			UserUpdateFailed:          "User failed to update",
			UserDeleteSuccess:         "User deleted successfully",
			UserDeleteFailed:          "User failed to delete",
			UserFetchByUsernameFailed: "User fetch by username failed",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ms]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ms]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return ms.String()
}
