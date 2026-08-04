package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
