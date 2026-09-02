package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAuditEventProjectIdBSON(t *testing.T) {
	organisationId := primitive.NewObjectID()
	projectId := primitive.NewObjectID()

	encoded, err := bson.Marshal(AuditEvent{
		OrganisationId:     organisationId,
		ProjectId:          &projectId,
		RequiredPermission: "media.read",
	})
	if err != nil {
		t.Fatalf("marshal project audit event: %v", err)
	}

	var document bson.M
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal project audit event: %v", err)
	}
	if got := document["organisationId"]; got != organisationId {
		t.Fatalf("organisationId = %#v, want %s", got, organisationId.Hex())
	}
	if got := document["projectId"]; got != projectId {
		t.Fatalf("projectId = %#v, want %s", got, projectId.Hex())
	}
	if got := document["requiredPermission"]; got != "media.read" {
		t.Fatalf("requiredPermission = %#v, want media.read", got)
	}

	encoded, err = bson.Marshal(AuditEvent{OrganisationId: organisationId})
	if err != nil {
		t.Fatalf("marshal organisation audit event: %v", err)
	}
	document = nil
	if err := bson.Unmarshal(encoded, &document); err != nil {
		t.Fatalf("unmarshal organisation audit event: %v", err)
	}
	if _, exists := document["projectId"]; exists {
		t.Fatalf("organisation-only audit event persisted projectId: %#v", document["projectId"])
	}
	if _, exists := document["requiredPermission"]; exists {
		t.Fatalf("audit event without a permission persisted requiredPermission: %#v", document["requiredPermission"])
	}
}