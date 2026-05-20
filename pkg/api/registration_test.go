package api

import "testing"

func TestRegistrationStatus_String(t *testing.T) {
	if got := RegistrationSuccess.String(); got != "registration_success" {
		t.Errorf("String() = %q, want %q", got, "registration_success")
	}
}

func TestRegistrationStatus_Translate(t *testing.T) {
	tests := []struct {
		name   string
		status RegistrationStatus
		lang   string
		want   string
	}{
		{"english known", RegistrationBindingFailed, "en", "Registration binding failed"},
		{"english missing info", RegistrationMissingInfo, "en", "Registration is missing required information"},
		{"english passwords no match", RegistrationPasswordsNoMatch, "en", "Passwords do not match"},
		{"english password too weak", RegistrationPasswordTooWeak, "en", "Password is too weak"},
		{"english username exists", RegistrationUsernameExists, "en", "Username already exists"},
		{"english email exists", RegistrationEmailExists, "en", "Email already exists"},
		{"english create user failed", RegistrationCreateUserFailed, "en", "Failed to create user"},
		{"english create org failed", RegistrationCreateOrgFailed, "en", "Failed to create organisation for user"},
		{"english hash password failed", RegistrationHashPasswordFailed, "en", "Failed to hash password"},
		{"english success", RegistrationSuccess, "en", "Registration successful"},
		{"english generate key success", RegistrationGenerateKeySuccess, "en", "User key generated successfully"},
		{"english generate key failed", RegistrationGenerateKeyFailed, "en", "Failed to generate user key"},
		{"english update password ok", RegistrationUpdatePasswordOK, "en", "Password updated successfully"},
		{"english update password fail", RegistrationUpdatePasswordFail, "en", "Failed to update password"},
		{"english user id required", RegistrationUserIdRequired, "en", "User ID is required"},
		{"unknown lang falls back to english", RegistrationSuccess, "xx", "Registration successful"},
		{"unknown status returns raw value", RegistrationStatus("registration_unknown"), "en", "registration_unknown"},
		{"unknown lang and status returns raw value", RegistrationStatus("registration_unknown"), "xx", "registration_unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.Translate(tc.lang); got != tc.want {
				t.Errorf("Translate(%q) = %q, want %q", tc.lang, got, tc.want)
			}
		})
	}
}
