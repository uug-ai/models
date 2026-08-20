package api

import (
	"encoding/json"
	"testing"

	"github.com/uug-ai/models/pkg/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMediaEditContractsCarryProjectOwnership(t *testing.T) {
	projectId := primitive.NewObjectID()
	contracts := []any{
		SubmitFaceRedactionRequest{OrganisationId: primitive.NewObjectID().Hex(), ProjectId: &projectId},
		CaseMediaStatusEvent{OrganisationId: primitive.NewObjectID().Hex(), ProjectId: &projectId, Status: models.CaseMediaStatusProcessing},
	}
	for _, contract := range contracts {
		data, err := json.Marshal(contract)
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(data, &document); err != nil {
			t.Fatal(err)
		}
		if document["projectId"] != projectId.Hex() {
			t.Fatalf("projectId = %#v, want %s for %T", document["projectId"], projectId.Hex(), contract)
		}
	}
}
