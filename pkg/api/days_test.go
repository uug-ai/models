package api

import (
	"encoding/json"
	"testing"
)

func TestGetDaysSuccessResponseJSON(t *testing.T) {
	response := GetDaysSuccessResponse{
		SuccessResponse: CreateSuccess(
			HttpStatusOK,
			ApplicationStatusGetSuccess,
			DaysFound,
			Metadata{},
		),
		Data: GetDaysResponse{Days: []string{"31-08-2026"}},
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var decoded GetDaysSuccessResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if decoded.EntityStatusCode != string(DaysFound) {
		t.Fatalf("entity status = %q, want %q", decoded.EntityStatusCode, DaysFound)
	}
	if len(decoded.Data.Days) != 1 || decoded.Data.Days[0] != "31-08-2026" {
		t.Fatalf("days = %v, want [31-08-2026]", decoded.Data.Days)
	}
}

func TestDaysStatusTranslation(t *testing.T) {
	if got := DaysRetrievalFailed.Translate("en"); got != "Failed to retrieve days" {
		t.Fatalf("translation = %q", got)
	}
	if got := DaysFound.Translate("unknown"); got != "Days found" {
		t.Fatalf("fallback translation = %q", got)
	}
}
