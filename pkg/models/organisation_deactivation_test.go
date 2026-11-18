package models

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOrganisationDeactivation(t *testing.T) {
	at := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	deleteAfter := at.AddDate(0, 0, 14)
	organisation := Organisation{Id: primitive.NewObjectID(), Name: "Acme"}

	// An open organisation stores no deactivation field.
	encoded, err := bson.Marshal(organisation)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bson.Raw(encoded).LookupErr("deactivation"); err == nil {
		t.Fatal("an open organisation stores a deactivation field")
	}

	organisation.Deactivation = &Deactivation{At: at, Reason: DeactivationReasonDeletion, DeleteAfter: &deleteAfter}
	encoded, err = bson.Marshal(organisation)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Organisation
	if err := bson.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	lock := decoded.Deactivation
	if lock == nil || lock.Reason != "deletion" || !lock.At.Equal(at) || lock.DeleteAfter == nil || !lock.DeleteAfter.Equal(deleteAfter) || lock.PurgeStartedAt != nil {
		t.Fatalf("decoded lock = %#v", lock)
	}
}
