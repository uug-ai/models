package api

import "testing"

func TestAdminStatus_String(t *testing.T) {
	if got := AdminStatsSuccess.String(); got != "admin_stats_success" {
		t.Errorf("String() = %q, want %q", got, "admin_stats_success")
	}
}

func TestAdminStatus_Translate(t *testing.T) {
	tests := []struct {
		name   string
		status AdminStatus
		lang   string
		want   string
	}{
		{"english stats success", AdminStatsSuccess, "en", "Admin stats retrieved successfully"},
		{"english stats failed", AdminStatsFailed, "en", "Admin stats retrieval failed"},
		{"unknown lang falls back to english", AdminStatsSuccess, "xx", "Admin stats retrieved successfully"},
		{"unknown status returns raw value", AdminStatus("admin_unknown"), "en", "admin_unknown"},
		{"unknown lang and status returns raw value", AdminStatus("admin_unknown"), "xx", "admin_unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.status.Translate(tc.lang); got != tc.want {
				t.Errorf("Translate(%q) = %q, want %q", tc.lang, got, tc.want)
			}
		})
	}
}
