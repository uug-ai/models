package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestProjectResourceScopeIsOrganisationWide(t *testing.T) {
	projectId := primitive.NewObjectID()

	tests := []struct {
		name  string
		scope ProjectResourceScope
		want  bool
	}{
		{name: "unassigned", scope: ProjectResourceScope{}, want: true},
		{name: "assigned", scope: ProjectResourceScope{ProjectId: &projectId}, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.scope.IsOrganisationWide(); got != test.want {
				t.Fatalf("IsOrganisationWide() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRoleAssignmentScopeWithProjectsIsNotOrganisationWide(t *testing.T) {
	scope := RoleAssignmentScope{ProjectIds: []primitive.ObjectID{primitive.NewObjectID()}}
	if scope.IsOrganisationWide() {
		t.Fatal("project-scoped role assignment must not be organisation-wide")
	}
}
