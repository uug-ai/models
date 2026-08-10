package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOrganisationOwnedResourceBSONContracts(t *testing.T) {
	organisationId := primitive.NewObjectID().Hex()

	tests := []struct {
		name     string
		resource any
	}{
		{name: "IO", resource: IO{OrganisationId: organisationId}},
		{name: "Sequence", resource: Sequence{OrganisationId: organisationId, UserId: "actor"}},
		{name: "Task", resource: Task{OrganisationId: organisationId, UserId: "legacy-tenant", ReporterId: "reporter"}},
		{name: "Videowall", resource: Videowall{OrganisationId: organisationId, UserId: "creator", MasterUserId: "legacy-tenant"}},
		{name: "AnalysisWrapper", resource: AnalysisWrapper{OrganisationId: organisationId, UserId: "legacy-tenant"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := bson.Marshal(test.resource)
			if err != nil {
				t.Fatalf("marshal resource: %v", err)
			}

			var document bson.M
			if err := bson.Unmarshal(encoded, &document); err != nil {
				t.Fatalf("unmarshal resource: %v", err)
			}
			if document["organisationId"] != organisationId {
				t.Fatalf("organisationId = %v, want %s", document["organisationId"], organisationId)
			}
		})
	}
}
