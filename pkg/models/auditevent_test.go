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
		SchemaVersion:  AuditEventSchemaVersion,
		OrganisationId: organisationId,
		ProjectId:      &projectId,
		AuthorizationInfo: []AuditAuthorizationInfo{{
			Permission: "media.read", ResourceType: "media", ResourceId: "recording-1", Granted: true,
		}},
		Status:  AuditEventStatus{Outcome: "success"},
		Request: &AuditRequestInfo{Id: "request-1", TraceId: "trace-1"},
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
	if got := document["schemaVersion"]; got != int32(AuditEventSchemaVersion) {
		t.Fatalf("schemaVersion = %#v, want %d", got, AuditEventSchemaVersion)
	}
	decisions := document["authorizationInfo"].(primitive.A)
	decision := decisions[0].(bson.M)
	if decision["permission"] != "media.read" || decision["granted"] != true {
		t.Fatalf("authorizationInfo = %#v", decisions)
	}
	status := document["status"].(bson.M)
	if status["outcome"] != "success" {
		t.Fatalf("status = %#v", status)
	}
	request := document["request"].(bson.M)
	if request["id"] != "request-1" || request["traceId"] != "trace-1" {
		t.Fatalf("request = %#v", request)
	}

	encoded, err = bson.Marshal(AuditEvent{SchemaVersion: AuditEventSchemaVersion, OrganisationId: organisationId})
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
	if _, exists := document["authorizationInfo"]; exists {
		t.Fatalf("audit event without decisions persisted authorizationInfo: %#v", document["authorizationInfo"])
	}
}
