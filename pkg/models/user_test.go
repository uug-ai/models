package models

import (
	"encoding/json"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserOrganisationIdSerialization(t *testing.T) {
	organisationId := primitive.NewObjectID()
	user := User{OrganisationId: organisationId}

	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	var jsonDocument map[string]any
	if err := json.Unmarshal(jsonData, &jsonDocument); err != nil {
		t.Fatal(err)
	}
	if got := jsonDocument["organisationId"]; got != organisationId.Hex() {
		t.Fatalf("JSON organisationId = %v, want %s", got, organisationId.Hex())
	}
	if _, exists := jsonDocument["organisation_id"]; exists {
		t.Fatal("legacy JSON organisation_id field was emitted")
	}

	bsonData, err := bson.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	var bsonDocument bson.M
	if err := bson.Unmarshal(bsonData, &bsonDocument); err != nil {
		t.Fatal(err)
	}
	if got := bsonDocument["organisationId"]; got != organisationId {
		t.Fatalf("BSON organisationId = %v, want %s", got, organisationId.Hex())
	}
	if _, exists := bsonDocument["organisation_id"]; exists {
		t.Fatal("legacy BSON organisation_id field was emitted")
	}
}

func TestGetOrganisationObjectIdPrefersExplicitOrganisation(t *testing.T) {
	activeOrganisationId := primitive.NewObjectID()
	masterId := primitive.NewObjectID()
	user := User{
		Id:             primitive.NewObjectID(),
		OrganisationId: activeOrganisationId,
		MasterAccount:  masterId.Hex(),
		Master:         &User{Id: masterId},
	}

	if got := GetOrganisationObjectId(user); got != activeOrganisationId {
		t.Fatalf("organisation id = %s, want %s", got.Hex(), activeOrganisationId.Hex())
	}
}

func TestGetOrganisationObjectIdFallsBackToMaster(t *testing.T) {
	masterId := primitive.NewObjectID()
	user := User{
		Id:            primitive.NewObjectID(),
		MasterAccount: masterId.Hex(),
		Master:        &User{Id: masterId},
	}

	if got := GetOrganisationObjectId(user); got != masterId {
		t.Fatalf("organisation id = %s, want %s", got.Hex(), masterId.Hex())
	}
}

func TestGetOrganisationObjectIdHandlesUnhydratedMaster(t *testing.T) {
	userId := primitive.NewObjectID()
	user := User{Id: userId, MasterAccount: primitive.NewObjectID().Hex()}

	if got := GetOrganisationObjectId(user); got != userId {
		t.Fatalf("organisation id = %s, want %s", got.Hex(), userId.Hex())
	}
}

func TestStableOwnerObjectIdIgnoresActiveOrganisation(t *testing.T) {
	userId := primitive.NewObjectID()
	masterId := primitive.NewObjectID()
	activeOrganisationId := primitive.NewObjectID()

	if got := StableOwnerObjectId(User{
		Id:             userId,
		MasterAccount:  masterId.Hex(),
		OrganisationId: activeOrganisationId,
	}); got != masterId {
		t.Fatalf("stable owner = %s, want master %s", got.Hex(), masterId.Hex())
	}
	if got := StableOwnerObjectId(User{Id: userId, OrganisationId: activeOrganisationId}); got != userId {
		t.Fatalf("stable owner = %s, want own ID %s", got.Hex(), userId.Hex())
	}
}

func TestStableOwnerObjectIdRejectsMalformedMaster(t *testing.T) {
	userId := primitive.NewObjectID()
	if got := StableOwnerObjectId(User{Id: userId, MasterAccount: "invalid"}); got != userId {
		t.Fatalf("stable owner = %s, want own ID %s", got.Hex(), userId.Hex())
	}
}
