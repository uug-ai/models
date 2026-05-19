package api

import "testing"

func TestSubscriptionStatus_String(t *testing.T) {
	if got := SubscriptionFound.String(); got != "subscription_found" {
		t.Errorf("String() = %q, want %q", got, "subscription_found")
	}
}

func TestSubscriptionStatus_Translate(t *testing.T) {
	tests := []struct {
		name   string
		status SubscriptionStatus
		lang   string
		want   string
	}{
		{"english found", SubscriptionFound, "en", "Subscription found"},
		{"english not found", SubscriptionNotFound, "en", "Subscription not found"},
		{"english get all success", SubscriptionGetAllSuccess, "en", "Subscriptions retrieved successfully"},
		{"english get all failed", SubscriptionGetAllFailed, "en", "Failed to retrieve subscriptions"},
		{"english get settings failed", SubscriptionGetSettingsFailed, "en", "Failed to retrieve subscription settings"},
		{"english get user failed", SubscriptionGetUserFailed, "en", "Failed to retrieve user subscription"},
		{"english update success", SubscriptionUpdateSuccess, "en", "Subscription updated successfully"},
		{"english update failed", SubscriptionUpdateFailed, "en", "Failed to update subscription"},
		{"english binding failed", SubscriptionBindingFailed, "en", "Subscription binding failed"},
		{"unknown lang falls back to english", SubscriptionFound, "xx", "Subscription found"},
		{"unknown status returns raw value", SubscriptionStatus("subscription_unknown"), "en", "subscription_unknown"},
		{"unknown lang and status returns raw value", SubscriptionStatus("subscription_unknown"), "xx", "subscription_unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.Translate(tc.lang); got != tc.want {
				t.Errorf("Translate(%q) = %q, want %q", tc.lang, got, tc.want)
			}
		})
	}
}
