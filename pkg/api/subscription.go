package api

// SubscriptionStatus represents specific status codes for subscription operations.
type SubscriptionStatus string

const (
	SubscriptionFound             SubscriptionStatus = "subscription_found"
	SubscriptionNotFound          SubscriptionStatus = "subscription_not_found"
	SubscriptionGetAllSuccess     SubscriptionStatus = "subscription_get_all_success"
	SubscriptionGetAllFailed      SubscriptionStatus = "subscription_get_all_failed"
	SubscriptionGetSettingsFailed SubscriptionStatus = "subscription_get_settings_failed"
	SubscriptionGetUserFailed     SubscriptionStatus = "subscription_get_user_failed"
	SubscriptionUpdateSuccess     SubscriptionStatus = "subscription_update_success"
	SubscriptionUpdateFailed      SubscriptionStatus = "subscription_update_failed"
	SubscriptionBindingFailed     SubscriptionStatus = "subscription_binding_failed"
)

// String returns the string representation of the subscription status.
func (s SubscriptionStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the subscription
// status in the specified language.
func (s SubscriptionStatus) Translate(lang string) string {
	translations := map[string]map[SubscriptionStatus]string{
		"en": {
			SubscriptionFound:             "Subscription found",
			SubscriptionNotFound:          "Subscription not found",
			SubscriptionGetAllSuccess:     "Subscriptions retrieved successfully",
			SubscriptionGetAllFailed:      "Failed to retrieve subscriptions",
			SubscriptionGetSettingsFailed: "Failed to retrieve subscription settings",
			SubscriptionGetUserFailed:     "Failed to retrieve user subscription",
			SubscriptionUpdateSuccess:     "Subscription updated successfully",
			SubscriptionUpdateFailed:      "Failed to update subscription",
			SubscriptionBindingFailed:     "Subscription binding failed",
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
