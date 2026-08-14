package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRoleAssignmentScopeWithProjectsIsNotOrganisationWide(t *testing.T) {
	scope := RoleAssignmentScope{ProjectIds: []primitive.ObjectID{primitive.NewObjectID()}}
	if scope.IsOrganisationWide() {
		t.Fatal("project-scoped role assignment must not be organisation-wide")
	}
}

// TestProjectScopedResourceBSONContracts pins the flat ProjectId field to a
// top-level projectId key on every project-scoped resource, so
// database.ProjectScope (which matches that exact top-level key) stays in
// lock-step with the model. It also asserts an unset ProjectId stays
// organisation-wide (no projectId key), which is why no historical backfill is
// required.
func TestProjectScopedResourceBSONContracts(t *testing.T) {
	projectId := primitive.NewObjectID()

	build := func(assign func(*primitive.ObjectID) any) (assigned bson.M, unset bson.M) {
		return marshalM(assign(&projectId)), marshalM(assign(nil))
	}

	tests := []struct {
		name   string
		assign func(*primitive.ObjectID) any
	}{
		{name: "Device", assign: func(p *primitive.ObjectID) any { return Device{ProjectId: p} }},
		{name: "Site", assign: func(p *primitive.ObjectID) any { return Site{ProjectId: p} }},
		{name: "Group", assign: func(p *primitive.ObjectID) any { return Group{ProjectId: p} }},
		{name: "Label", assign: func(p *primitive.ObjectID) any { return Label{ProjectId: p} }},
		{name: "IO", assign: func(p *primitive.ObjectID) any { return IO{ProjectId: p} }},
		{name: "Videowall", assign: func(p *primitive.ObjectID) any { return Videowall{ProjectId: p} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assigned, unset := build(test.assign)

			if assigned["projectId"] != projectId {
				t.Fatalf("assigned projectId = %v, want %s", assigned["projectId"], projectId.Hex())
			}
			if _, exists := unset["projectId"]; exists {
				t.Fatal("organisation-wide resource must not carry a projectId")
			}
		})
	}
}

func marshalM(v any) bson.M {
	encoded, err := bson.Marshal(v)
	if err != nil {
		panic(err)
	}
	var document bson.M
	if err := bson.Unmarshal(encoded, &document); err != nil {
		panic(err)
	}
	return document
}
