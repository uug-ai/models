package models

import (
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDefaultProjectIdReusesOrganisationId(t *testing.T) {
	organisationId := primitive.NewObjectID()
	if got := DefaultProjectId(organisationId); got != organisationId {
		t.Fatalf("DefaultProjectId() = %s, want %s", got.Hex(), organisationId.Hex())
	}
}

func TestDefaultProjectIdIsStable(t *testing.T) {
	// Two independently deployed services calling this for the same
	// organisation must agree. Nothing here may depend on state, so calling it
	// repeatedly is the same as calling it from anywhere else.
	organisationId := primitive.NewObjectID()
	first := DefaultProjectId(organisationId)
	for range 100 {
		if got := DefaultProjectId(organisationId); got != first {
			t.Fatalf("DefaultProjectId() = %s, want %s", got.Hex(), first.Hex())
		}
	}
}

func TestResolveProjectIdPrefersStoredAssignment(t *testing.T) {
	organisationId := primitive.NewObjectID()
	stored := primitive.NewObjectID()

	if got := ResolveProjectId(organisationId, &stored); got != stored {
		t.Fatalf("ResolveProjectId() = %s, want stored %s", got.Hex(), stored.Hex())
	}
}

func TestResolveProjectIdFallsBackForUnassignedResources(t *testing.T) {
	organisationId := primitive.NewObjectID()
	zero := primitive.NilObjectID

	// A nil pointer and a stored zero both mean "no project assigned". Neither
	// may be treated as a real project, or the resource is stamped with
	// NilObjectID and disappears from project-scoped reads.
	for name, stored := range map[string]*primitive.ObjectID{
		"nil":  nil,
		"zero": &zero,
	} {
		if got := ResolveProjectId(organisationId, stored); got != organisationId {
			t.Fatalf("ResolveProjectId(%s) = %s, want organisation default %s", name, got.Hex(), organisationId.Hex())
		}
	}
}

func TestProjectScopeFilterIsStrictForARealProject(t *testing.T) {
	organisationId := primitive.NewObjectID()
	projectId := primitive.NewObjectID()

	got := ProjectScopeFilter(organisationId.Hex(), projectId)
	want := bson.M{"projectId": projectId}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProjectScopeFilter() = %v, want %v", got, want)
	}
}

func TestProjectScopeFilterDoesNotRelaxForARealProject(t *testing.T) {
	// This is the regression an unconditionally tolerant predicate would pass
	// every other test in this file and still cause. Today every project is its
	// organisation's default, so a null arm here is invisible; the day a second
	// project exists, that arm makes this project's reads match every unstamped
	// document in the organisation — documents that belong to the default
	// project, not to this one.
	organisationId := primitive.NewObjectID()
	projectId := primitive.NewObjectID()

	if toleratesUnstampedDocuments(ProjectScopeFilter(organisationId.Hex(), projectId)) {
		t.Fatal("a real project must not match documents that carry no project")
	}
}

// toleratesUnstampedDocuments reports whether a clause would select a document
// that carries no projectId at all.
//
// It reads the predicate's meaning rather than its spelling. The tolerant form
// has been an $or over an $exists arm and is now a two-element $in; both select
// the same documents, and it is that selection — not the syntax — these tests
// exist to pin down.
func toleratesUnstampedDocuments(filter bson.M) bool {
	// A bare equality never matches an absent field.
	operators, ok := filter["projectId"].(bson.M)
	if !ok {
		return false
	}
	candidates, ok := operators["$in"].(bson.A)
	if !ok {
		return false
	}
	for _, candidate := range candidates {
		// null matches both an explicit null and a missing field.
		if candidate == nil {
			return true
		}
	}
	return false
}

func TestProjectScopeFilterRelaxesForTheDefaultProject(t *testing.T) {
	// Everything written before the project axis existed is unstamped, and the
	// default project is where it belongs. A strict clause here would erase an
	// organisation's entire pre-rollout history from the UI.
	organisationId := primitive.NewObjectID()
	defaultProjectId := DefaultProjectId(organisationId)

	got := ProjectScopeFilter(organisationId.Hex(), defaultProjectId)
	want := bson.M{"projectId": bson.M{"$in": bson.A{defaultProjectId, nil}}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProjectScopeFilter() = %v, want %v", got, want)
	}
}

// The tolerant clause must stay a single key holding point-matchable values.
// An $exists arm, or any other operator the planner cannot turn into index
// bounds, selects exactly the same documents but forces the reader that ANDs
// this clause onto an organisation scope into a full collection scan.
func TestProjectScopeFilterTolerantFormIsIndexBounded(t *testing.T) {
	organisationId := primitive.NewObjectID()

	filter := ProjectScopeFilter(organisationId.Hex(), DefaultProjectId(organisationId))
	if len(filter) != 1 {
		t.Fatalf("tolerant filter must be a single key, got %v", filter)
	}

	candidates, ok := filter["projectId"].(bson.M)["$in"].(bson.A)
	if !ok {
		t.Fatalf("tolerant filter must narrow projectId with $in, got %v", filter)
	}
	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}
		if _, isOperator := candidate.(bson.M); isOperator {
			t.Fatalf("$in must hold plain values, got operator %v", candidate)
		}
	}
}

func TestProjectScopeFilterDoesNotNarrowWithoutAResolvableProject(t *testing.T) {
	// Both inputs mean the caller could not resolve a project. Returning nil
	// degrades the read to organisation-wide, which the caller's own
	// organisation clause still bounds. The alternative — a predicate that
	// matches nothing — would blank a tenant's screen over a resolution glitch.
	organisationId := primitive.NewObjectID()

	for name, filter := range map[string]bson.M{
		"zero project":              ProjectScopeFilter(organisationId.Hex(), primitive.NilObjectID),
		"non-ObjectID organisation": ProjectScopeFilter("legacy-tenant", primitive.NewObjectID()),
	} {
		if filter != nil {
			t.Fatalf("ProjectScopeFilter(%s) = %v, want nil", name, filter)
		}
	}
}

func TestProjectScopeFilterAcceptsAnUppercaseOrganisationId(t *testing.T) {
	// The default-project test is a comparison against a parsed id rather than
	// against the hex string, so a differently-cased id still resolves to the
	// tolerant form. Comparing strings would silently fall through to the strict
	// clause and hide every unstamped document.
	organisationId := primitive.NewObjectID()
	upper := strings.ToUpper(organisationId.Hex())

	if !toleratesUnstampedDocuments(ProjectScopeFilter(upper, DefaultProjectId(organisationId))) {
		t.Fatal("an uppercase organisation id must still resolve to the default project")
	}
}
