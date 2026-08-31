package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestProjectCredentialsPersistButDoNotAppearInJSON(t *testing.T) {
	project := Project{
		PublicKey:     "public-id",
		PrivateKey:    "private-secret",
		EncryptionKey: "encryption-secret",
	}

	document := marshalM(project)
	if document["publicKey"] != "public-id" || document["privateKey"] != "private-secret" || document["encryptionKey"] != "encryption-secret" {
		t.Fatalf("project credential BSON = %#v", document)
	}

	encoded, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("marshal project JSON: %v", err)
	}
	for _, secret := range []string{"public-id", "private-secret", "encryption-secret"} {
		if string(encoded) == secret || containsJSONValue(encoded, secret) {
			t.Fatalf("ordinary project JSON leaked credential %q: %s", secret, encoded)
		}
	}
}

func containsJSONValue(encoded []byte, value string) bool {
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		return true
	}
	for _, fieldValue := range document {
		if fieldValue == value {
			return true
		}
	}
	return false
}

func TestRoleAssignmentScopeWithProjectsIsNotOrganisationWide(t *testing.T) {
	scope := RoleAssignmentScope{ProjectIds: []primitive.ObjectID{primitive.NewObjectID()}}
	if scope.IsOrganisationWide() {
		t.Fatal("project-scoped role assignment must not be organisation-wide")
	}
}

// TestProjectScopedResourceBSONContracts pins the flat ProjectId field to a
// top-level projectId key on every project-scoped resource, so
// database.ProjectScope (which matches that exact top-level key) stays in
// lock-step with the model. It also asserts an unset ProjectId stays
// organisation-wide (no projectId key), which is why no historical backfill is
// required.
func TestProjectScopedResourceBSONContracts(t *testing.T) {
	projectId := primitive.NewObjectID()

	build := func(assign func(*primitive.ObjectID) any) (assigned bson.M, unset bson.M) {
		return marshalM(assign(&projectId)), marshalM(assign(nil))
	}

	tests := []struct {
		name   string
		assign func(*primitive.ObjectID) any
	}{
		{name: "Device", assign: func(p *primitive.ObjectID) any { return Device{ProjectId: p} }},
		{name: "Site", assign: func(p *primitive.ObjectID) any { return Site{ProjectId: p} }},
		{name: "Group", assign: func(p *primitive.ObjectID) any { return Group{ProjectId: p} }},
		{name: "Label", assign: func(p *primitive.ObjectID) any { return Label{ProjectId: p} }},
		{name: "IO", assign: func(p *primitive.ObjectID) any { return IO{ProjectId: p} }},
		{name: "Videowall", assign: func(p *primitive.ObjectID) any { return Videowall{ProjectId: p} }},
		{name: "Media", assign: func(p *primitive.ObjectID) any { return Media{ProjectId: p} }},
		{name: "Marker", assign: func(p *primitive.ObjectID) any { return Marker{ProjectId: p} }},
		{name: "AnalysisWrapper", assign: func(p *primitive.ObjectID) any { return AnalysisWrapper{ProjectId: p} }},
		{name: "DetectionRun", assign: func(p *primitive.ObjectID) any { return DetectionRun{ProjectId: p} }},
		{name: "Sequence", assign: func(p *primitive.ObjectID) any { return Sequence{ProjectId: p} }},
		{name: "Task", assign: func(p *primitive.ObjectID) any { return Task{ProjectId: p} }},
		{name: "Comment", assign: func(p *primitive.ObjectID) any { return Comment{ProjectId: p} }},
		{name: "CustomAlert", assign: func(p *primitive.ObjectID) any { return CustomAlert{ProjectId: p} }},
		{name: "State", assign: func(p *primitive.ObjectID) any { return State{ProjectId: p} }},
		{name: "Message", assign: func(p *primitive.ObjectID) any { return Message{ProjectId: p} }},
		{name: "NotificationEvent", assign: func(p *primitive.ObjectID) any { return NotificationEvent{Message: Message{ProjectId: p}} }},
		{name: "CaseMedia", assign: func(p *primitive.ObjectID) any { return CaseMedia{ProjectId: p} }},
		{name: "CaseAttachment", assign: func(p *primitive.ObjectID) any { return CaseAttachment{ProjectId: p} }},
		{name: "CaseShare", assign: func(p *primitive.ObjectID) any { return CaseShare{ProjectId: p} }},
		{name: "Workflow", assign: func(p *primitive.ObjectID) any { return Workflow{ProjectId: p} }},
		{name: "WorkflowRun", assign: func(p *primitive.ObjectID) any { return WorkflowRun{ProjectId: p} }},
		{name: "AnalyticsCount", assign: func(p *primitive.ObjectID) any { return AnalyticsCount{ProjectId: p} }},
		{name: "Heatmap", assign: func(p *primitive.ObjectID) any { return Heatmap{ProjectId: p} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assigned, unset := build(test.assign)

			if assigned["projectId"] != projectId {
				t.Fatalf("assigned projectId = %v, want %s", assigned["projectId"], projectId.Hex())
			}
			if _, exists := unset["projectId"]; exists {
				t.Fatal("organisation-wide resource must not carry a projectId")
			}
		})
	}
}

func marshalM(v any) bson.M {
	encoded, err := bson.Marshal(v)
	if err != nil {
		panic(err)
	}
	var document bson.M
	if err := bson.Unmarshal(encoded, &document); err != nil {
		panic(err)
	}
	return document
}
