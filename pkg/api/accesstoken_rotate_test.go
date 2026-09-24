package api

import (
	"encoding/json"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

func TestRotateAccessTokenSuccessResponseJSON(t *testing.T) {
	response := RotateAccessTokenSuccessResponse{
		SuccessResponse: CreateSuccess(
			HttpStatusOK,
			ApplicationStatusUpdateSuccess,
			AccessTokenRotateSuccess,
			Metadata{},
		),
		Data: RotateAccessTokenResponse{Token: models.AccessToken{
			Name:   "ci",
			Token:  "new-secret",
			Scopes: []models.AccessTokenScope{models.PermissionMediaRead},
		}},
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	var decoded RotateAccessTokenSuccessResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if decoded.EntityStatusCode != string(AccessTokenRotateSuccess) {
		t.Fatalf("entity status = %q, want %q", decoded.EntityStatusCode, AccessTokenRotateSuccess)
	}
	if decoded.Data.Token.Token != "new-secret" || decoded.Data.Token.Name != "ci" {
		t.Fatalf("token = %+v", decoded.Data.Token)
	}
}

func TestRotateAccessTokenStatusTranslation(t *testing.T) {
	if got := AccessTokenRotateFailed.Translate("en"); got != "Failed to rotate access token" {
		t.Fatalf("translation = %q", got)
	}
	if got := AccessTokenExpired.Translate("unknown"); got != "Access token has expired" {
		t.Fatalf("fallback translation = %q", got)
	}
	if got := AccessTokenForbidden.Translate("en"); got != "You are not allowed to manage this access token" {
		t.Fatalf("forbidden translation = %q", got)
	}
}
