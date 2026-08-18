package projectscope

import (
	"testing"

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
