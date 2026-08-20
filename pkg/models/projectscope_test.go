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

	if _, relaxed := ProjectScopeFilter(organisationId.Hex(), projectId)["$or"]; relaxed {
		t.Fatal("a real project must not match documents that carry no project")
	}
}

func TestProjectScopeFilterRelaxesForTheDefaultProject(t *testing.T) {
	// Everything written before the project axis existed is unstamped, and the
	// default project is where it belongs. A strict clause here would erase an
	// organisation's entire pre-rollout history from the UI.
	organisationId := primitive.NewObjectID()
	defaultProjectId := DefaultProjectId(organisationId)

	got := ProjectScopeFilter(organisationId.Hex(), defaultProjectId)
	want := bson.M{"$or": []bson.M{
		{"projectId": defaultProjectId},
		{"projectId": bson.M{"$exists": false}},
		{"projectId": nil},
	}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ProjectScopeFilter() = %v, want %v", got, want)
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

	if _, relaxed := ProjectScopeFilter(upper, DefaultProjectId(organisationId))["$or"]; !relaxed {
		t.Fatal("an uppercase organisation id must still resolve to the default project")
	}
}
