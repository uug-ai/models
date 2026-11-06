package api

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

var invitationStatuses = []InvitationResponseStatus{
	InvitationBindingFailed, InvitationValidationFailed, InvitationFound, InvitationNotFound,
	InvitationGetAllSuccess, InvitationGetAllFailed, InvitationCreateSuccess, InvitationCreateFailed,
	InvitationResendSuccess, InvitationResendFailed, InvitationRevokeSuccess, InvitationRevokeFailed,
	InvitationAcceptSuccess, InvitationAcceptFailed, InvitationPending, InvitationAlreadyMember,
	InvitationClosed, InvitationEmailMismatch, InvitationExceedsAccess, InvitationOwnerOnly, InvitationRoleNotFound,
	InvitationTeamNotFound, InvitationTeamManagedByProvider, InvitationLegacyMode,
}

func TestInvitationResponseStatusTranslate(t *testing.T) {
	seen := map[InvitationResponseStatus]bool{}
	for _, status := range invitationStatuses {
		if seen[status] {
			t.Errorf("%s listed twice", status)
		}
		seen[status] = true
		if !strings.HasPrefix(status.String(), "invitation_") {
			t.Errorf("%s: status codes start with invitation_", status)
		}
		if got := status.Translate("en"); got == status.String() || got == "" {
			t.Errorf("%s has no English translation", status)
		}
		if status.Translate("xx") != status.Translate("en") {
			t.Errorf("%s: unknown language does not fall back to English", status)
		}
	}
	if got := InvitationResponseStatus("invitation_unknown").Translate("en"); got != "invitation_unknown" {
		t.Errorf("unknown status = %q, want the raw value", got)
	}
}

func TestCreateInvitationRequestDecodes(t *testing.T) {
	body := `{"email": "kim@example.com", "assignments": [
		{"roleId": "6720f0a1b2c3d4e5f6a7b8ca", "scope": {"level": "project", "projectId": "6720f0a1b2c3d4e5f6a7b8cb", "resources": "selected"}}
	], "teamIds": ["6720f0a1b2c3d4e5f6a7b8cc"]}`
	var request CreateInvitationRequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatal(err)
	}
	if request.Email != "kim@example.com" || len(request.Assignments) != 1 || request.Assignments[0].Scope.Level != models.LevelProject || request.ExpiresAt != nil || len(request.TeamIds) != 1 {
		t.Errorf("decoded %+v", request)
	}
}

func TestInvitationDetailsKeepsTokenHashOut(t *testing.T) {
	details := InvitationDetails{Invitation: models.OrganisationInvitation{Token: "raw", TokenHash: models.HashInvitationToken("raw")}}
	body, err := json.Marshal(details)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), details.Invitation.TokenHash) {
		t.Error("response exposes the token hash")
	}
	if !strings.Contains(string(body), `"token":"raw"`) {
		t.Error("create and resend responses carry the raw token once")
	}
}
