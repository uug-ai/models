package models

import (
	"reflect"
	"testing"
	"time"

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

func TestActiveSubscriptionFilterCoversOpenEndedSubscriptions(t *testing.T) {
	now := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)

	want := bson.M{"$or": bson.A{
		bson.M{"ends_at": bson.M{"$gt": now}},
		bson.M{"ends_at": nil},
	}}

	if got := ActiveSubscriptionFilter(now); !reflect.DeepEqual(got, want) {
		t.Fatalf("ActiveSubscriptionFilter() = %#v, want %#v", got, want)
	}
}

func TestActiveSubscriptionFilterMatchesAbsentEndsAtViaNil(t *testing.T) {
	// The nil arm is load-bearing in a way an $exists rewrite would break in one
	// direction only. MongoDB matches nil against both an explicit null and a
	// missing key, so this one arm covers the never-cancelled documents that
	// carry no ends_at and the resumed ones that carry null. {$exists: false}
	// would silently drop the latter, and every fixture built by resuming a
	// subscription would stop matching.
	arms, ok := ActiveSubscriptionFilter(time.Now())["$or"].(bson.A)
	if !ok || len(arms) != 2 {
		t.Fatalf("filter is not a two-armed $or: %#v", ActiveSubscriptionFilter(time.Now()))
	}

	openEnded, ok := arms[1].(bson.M)
	if !ok {
		t.Fatalf("open-ended arm = %#v, want bson.M", arms[1])
	}
	endsAt, present := openEnded["ends_at"]
	if !present {
		t.Fatalf("open-ended arm does not test ends_at: %#v", openEnded)
	}
	if endsAt != nil {
		t.Fatalf("open-ended arm ends_at = %#v, want a literal nil rather than an operator document", endsAt)
	}
}

func TestActiveSubscriptionFilterUsesTheCallersInstant(t *testing.T) {
	// The instant is a parameter so a query assembled from several parts compares
	// all of them against one now. Reading the clock here would also make the
	// filter untestable against a fixture's ends_at.
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	arms := ActiveSubscriptionFilter(past)["$or"].(bson.A)
	active := arms[0].(bson.M)["ends_at"].(bson.M)
	if active["$gt"] != past {
		t.Fatalf("active arm compares against %#v, want the caller's instant %v", active["$gt"], past)
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
