package models

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSubscriptionOwnershipFilterUsesCanonicalBeforeLegacyFallback(t *testing.T) {
	organisationId := primitive.NewObjectID()

	want := bson.M{"$or": bson.A{
		bson.M{"organisation_id": organisationId},
		bson.M{
			"organisation_id": bson.M{"$exists": false},
			"user_id":         organisationId.Hex(),
		},
	}}

	if got := SubscriptionOwnershipFilter(organisationId); !reflect.DeepEqual(got, want) {
		t.Fatalf("SubscriptionOwnershipFilter() = %#v, want %#v", got, want)
	}
}

func TestSubscriptionOwnershipFilterGuardsTheLegacyArm(t *testing.T) {
	// The $exists test is what keeps the legacy arm from being a cross-tenant
	// leak: without it, a subscription owned by another organisation whose
	// legacy user_id happens to be this organisation's hex would match. It is
	// asserted on its own because a "simplification" that collapses the filter
	// to {$or: [{organisation_id: id}, {user_id: hex}]} selects the same
	// documents in every test fixture that has no such collision.
	organisationId := primitive.NewObjectID()

	arms, ok := SubscriptionOwnershipFilter(organisationId)["$or"].(bson.A)
	if !ok || len(arms) != 2 {
		t.Fatalf("filter is not a two-armed $or: %#v", SubscriptionOwnershipFilter(organisationId))
	}

	legacy, ok := arms[1].(bson.M)
	if !ok {
		t.Fatalf("legacy arm = %#v, want bson.M", arms[1])
	}
	if !reflect.DeepEqual(legacy["organisation_id"], bson.M{"$exists": false}) {
		t.Fatalf("legacy arm organisation_id guard = %#v, want {$exists: false}", legacy["organisation_id"])
	}
	if legacy["user_id"] != organisationId.Hex() {
		t.Fatalf("legacy arm user_id = %v, want %s", legacy["user_id"], organisationId.Hex())
	}
}

func TestLatestSubscriptionSortIsNewestFirst(t *testing.T) {
	// Order and direction are both load-bearing. The keys are the trailing keys
	// of the {organisation_id, updated_at, created_at, _id} index, so any
	// reordering costs the readers their index-served sort and gives them a
	// blocking in-memory one instead.
	want := bson.D{
		{Key: "updated_at", Value: -1},
		{Key: "created_at", Value: -1},
		{Key: "_id", Value: -1},
	}

	if got := LatestSubscriptionSort(); !reflect.DeepEqual(got, want) {
		t.Fatalf("LatestSubscriptionSort() = %#v, want %#v", got, want)
	}
}

func TestLatestSubscriptionSortIsNotAliased(t *testing.T) {
	// bson.D is a slice, so returning a package-level value would let one
	// caller's SetSort mutate every other caller's order.
	first := LatestSubscriptionSort()
	first[0].Key = "mutated"

	if got := LatestSubscriptionSort(); got[0].Key != "updated_at" {
		t.Fatalf("LatestSubscriptionSort() leading key = %q after a caller mutated an earlier result", got[0].Key)
	}
}
