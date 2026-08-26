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
		OrganisationId: organisationId,
		ProjectId:      &projectId,
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
}