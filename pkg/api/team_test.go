package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

var teamStatuses = []TeamStatus{
	TeamBindingFailed, TeamValidationFailed, TeamFound, TeamNotFound,
	TeamGetAllSuccess, TeamGetAllFailed, TeamCreateSuccess, TeamCreateFailed,
	TeamUpdateSuccess, TeamUpdateFailed, TeamDeleteSuccess, TeamDeleteFailed,
	TeamNameExists, TeamGroupLinked, TeamsDisabled,
	TeamMembersGetSuccess, TeamMembersGetFailed, TeamMembersAddSuccess, TeamMembersAddFailed,
	TeamMemberRemoveSuccess, TeamMemberRemoveFailed, TeamMemberNotFound,
	TeamMemberNotInOrganisation, TeamMembersManagedByProvider, TeamMemberSelf, TeamMemberExceedsAccess,
}

func TestTeamStatusTranslate(t *testing.T) {
	seen := map[TeamStatus]bool{}
	for _, status := range teamStatuses {
		if seen[status] {
			t.Errorf("%s listed twice", status)
		}
		seen[status] = true
		if !strings.HasPrefix(status.String(), "team") {
			t.Errorf("%s: status codes start with team", status)
		}
		if got := status.Translate("en"); got == status.String() || got == "" {
			t.Errorf("%s has no English translation", status)
		}
		if status.Translate("xx") != status.Translate("en") {
			t.Errorf("%s: unknown language does not fall back to English", status)
		}
	}
	if got := TeamStatus("team_unknown").Translate("en"); got != "team_unknown" {
		t.Errorf("unknown status = %q, want the raw value", got)
	}
}

func TestTeamMemberUserCarriesNoCredentials(t *testing.T) {
	body, err := json.Marshal(TeamMemberDetails{User: TeamMemberUser{Id: "u1", Username: "kim", Email: "kim@example.com"}})
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
		case "id", "username", "email", "nickname":
		default:
			t.Errorf("member identity exposes %q", field)
		}
	}
}

func TestCreateTeamRequestDecodesSyncedTeam(t *testing.T) {
	var request CreateTeamRequest
	body := `{"name":"Control room","source":"identityProvider","identityProvider":{"provider":"entra-id","externalGroupId":"g-1"}}`
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatal(err)
	}
	if request.Source != models.TeamSourceIdentityProvider || request.IdentityProvider == nil || request.IdentityProvider.ExternalGroupId != "g-1" {
		t.Errorf("decoded %+v", request)
	}
}
