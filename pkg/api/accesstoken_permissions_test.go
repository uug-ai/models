package api

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/uug-ai/models/pkg/models"
)

func TestGetAccessTokenPermissionsSuccessResponseJSON(t *testing.T) {
	response := GetAccessTokenPermissionsSuccessResponse{
		SuccessResponse: CreateSuccess(
			HttpStatusOK,
			ApplicationStatusGetSuccess,
			AccessTokenPermissionsFound,
			Metadata{},
		),
		Data: GetAccessTokenPermissionsResponse{
			Permissions: models.AccessTokenScopes(),
		},
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var decoded GetAccessTokenPermissionsSuccessResponse
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if decoded.EntityStatusCode != string(AccessTokenPermissionsFound) {
		t.Fatalf("entity status = %q, want %q", decoded.EntityStatusCode, AccessTokenPermissionsFound)
	}
	if decoded.Message != "Access token permissions found" {
		t.Fatalf("message = %q, want %q", decoded.Message, "Access token permissions found")
	}
	if !reflect.DeepEqual(decoded.Data.Permissions, models.AccessTokenScopes()) {
		t.Fatalf("permissions = %v, want %v", decoded.Data.Permissions, models.AccessTokenScopes())
	}
}
