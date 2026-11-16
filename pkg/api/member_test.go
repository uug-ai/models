package api

import (
	"encoding/json"
	"strings"
	"testing"
)

var memberStatuses = []MemberStatus{
	MemberBindingFailed, MemberValidationFailed, MemberFound, MemberNotFound,
	MemberGetAllSuccess, MemberGetAllFailed, MemberRemoveSuccess, MemberRemoveFailed,
	MemberSelf, MemberOwnerOnly, MemberLastOwner, MemberPrimaryOwner, MemberLegacyMode,
}

func TestMemberStatusTranslate(t *testing.T) {
	seen := map[MemberStatus]bool{}
	for _, status := range memberStatuses {
		if seen[status] {
			t.Errorf("%s listed twice", status)
		}
		seen[status] = true
		if !strings.HasPrefix(status.String(), "member_") {
			t.Errorf("%s: status codes start with member_", status)
		}
		if got := status.Translate("en"); got == status.String() || got == "" {
			t.Errorf("%s has no English translation", status)
		}
		if status.Translate("xx") != status.Translate("en") {
			t.Errorf("%s: unknown language does not fall back to English", status)
		}
	}
	if got := MemberStatus("member_unknown").Translate("en"); got != "member_unknown" {
		t.Errorf("unknown status = %q, want the raw value", got)
	}
}

func TestUserIdentityCarriesNoCredentials(t *testing.T) {
	body, err := json.Marshal(MemberSummary{User: UserIdentity{Id: "u1", Username: "kim", Email: "kim@example.com", FirstName: "Kim", LastName: "Lee", Nickname: "kimmy"}})
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		User map[string]any `json:"user"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	for field := range decoded.User {
		switch field {
		case "id", "username", "email", "firstName", "lastName", "nickname":
		default:
			t.Errorf("member identity exposes %q", field)
		}
	}
}
