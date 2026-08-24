package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SubscriptionOwnershipFilter selects the subscriptions an organisation owns.
//
// Ownership is recorded two ways. Canonical documents carry organisation_id.
// Documents predating the organisation model carry only user_id, holding the
// hex of what is now the organisation id — organisations were minted at the
// owning user's id, so the two strings coincide. The Hub API backfills
// organisation_id opportunistically, on cancel, resume and upgrade, so the
// migration is lazy and incomplete by construction: both arms stay load-bearing
// until a backfill retires the second one.
//
// The legacy arm keeps the organisation_id $exists test, and that test is a
// tenant guard rather than a narrowing optimisation. Drop it and a subscription
// owned by organisation B whose legacy user_id happens to be A's hex matches a
// read for A. Only a backfill that leaves no document without organisation_id
// makes the arm safe to remove.
//
// Both arms are index-bounded on purpose. MongoDB serves an $or from indexes
// only when every arm applies its bounds at index level; one unindexed arm
// collapses the whole $or to a collection scan, so the arms are a pair rather
// than two independent predicates. The legacy arm rides a {user_id} index with
// the $exists as a cheap residual, and the canonical arm needs an index leading
// with organisation_id.
func SubscriptionOwnershipFilter(organisationId primitive.ObjectID) bson.M {
	return bson.M{"$or": bson.A{
		bson.M{"organisation_id": organisationId},
		bson.M{
			"organisation_id": bson.M{"$exists": false},
			"user_id":         organisationId.Hex(),
		},
	}}
}

// ActiveSubscriptionFilter selects the subscriptions that have not lapsed.
//
// A subscription is active while ends_at lies in the future, and indefinitely
// while ends_at is unset — an open-ended subscription is the normal state, and
// cancelling is what writes the field. Resuming clears it again, so a document
// moves between the two arms over its life rather than belonging to one.
//
// The nil comparison is doing more work than it looks. MongoDB matches both an
// explicit null and a missing key, which is what makes the second arm cover the
// never-cancelled documents that carry no ends_at at all. Rewriting it as an
// $exists test would quietly drop the ones holding an explicit null.
//
// Callers pass the instant rather than letting this read the clock, so a query
// built from several parts compares every one of them against the same now, and
// so a test can pick an instant either side of a fixture's ends_at.
//
// Pair this with SubscriptionOwnershipFilter and the {organisation_id, ends_at}
// and {user_id, ends_at} indexes cover the result. Note this is a read-side
// filter only: a write that has to find a cancelled subscription in order to
// resume it must not apply it, or the document it exists to revive is exactly
// the one it excludes.
func ActiveSubscriptionFilter(now time.Time) bson.M {
	return bson.M{"$or": bson.A{
		bson.M{"ends_at": bson.M{"$gt": now}},
		bson.M{"ends_at": nil},
	}}
}

// LatestSubscriptionSort orders subscriptions newest-first.
//
// An organisation legitimately owns more than one: an upgrade or a
// cancel-then-resume leaves the superseded documents in place, and they remain
// active by the ends_at test. Without an order a single-document read resolves
// to whichever one the query plan reaches first, which makes the answer a
// function of the index set rather than of the data — adding or dropping an
// index then silently changes which subscription a tenant gets.
//
// created_at breaks the tie between documents written in the same instant, and
// _id between documents old enough to predate created_at. Missing values sort
// lowest in MongoDB, so descending puts them last: a document no modern write
// path has touched loses to one that has.
//
// These are the trailing keys of the {organisation_id, updated_at, created_at,
// _id} index, so a reader pairing this with SubscriptionOwnershipFilter is
// ordered from the index rather than by a blocking in-memory sort.
func LatestSubscriptionSort() bson.D {
	return bson.D{
		{Key: "updated_at", Value: -1},
		{Key: "created_at", Value: -1},
		{Key: "_id", Value: -1},
	}
}
