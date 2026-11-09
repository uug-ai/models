package models

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPagePermissionAliases(t *testing.T) {
	if len(legacyPageAliases) != len(PagePermissions())+len(OrganisationPagePermissions()) {
		t.Fatal("every page must have exactly one compatibility alias")
	}
	for legacy, canonical := range legacyPageAliases {
		t.Run(string(legacy), func(t *testing.T) {
			parsed, err := ParsePermission(string(legacy))
			if err != nil || parsed != canonical {
				t.Fatalf("ParsePermission(%q) = %q, %v", legacy, parsed, err)
			}
			oldDefinition, oldOK := LookupPermission(legacy)
			definition, ok := LookupPermission(canonical)
			if !oldOK || !ok || oldDefinition.Permission != canonical ||
				oldDefinition.Level != definition.Level || oldDefinition.OwnerOnly != definition.OwnerOnly ||
				oldDefinition.TenantWide != definition.TenantWide {
				t.Fatal("alias changed the permission's policy")
			}
			for _, input := range [][]Permission{{legacy}, {canonical}, {legacy, canonical}} {
				if got := ExpandPermissions(input); !slices.Equal(got, []Permission{canonical}) {
					t.Errorf("ExpandPermissions(%v) = %v", input, got)
				}
				if got := canonicalPermissions(input); !slices.Equal(got, []Permission{canonical}) {
					t.Errorf("canonicalPermissions(%v) = %v", input, got)
				}
			}
			if err := ValidateAccessTokenScopes([]Permission{canonical}); !errors.Is(err, ErrAccessTokenScopeNotGrantable) {
				t.Errorf("page permission became a token scope: %v", err)
			}
		})
	}
	for _, permission := range AllPermissions() {
		if strings.HasPrefix(string(permission), "pages.") {
			t.Errorf("canonical catalog contains legacy identifier %q", permission)
		}
	}
	if _, err := ParsePermission("pages.unknown"); !errors.Is(err, ErrPermissionUnknown) {
		t.Fatalf("unknown legacy page was accepted: %v", err)
	}
	for _, permission := range ExpandPermissions([]Permission{PermissionDevicesRead, PermissionMediaRead}) {
		if strings.HasSuffix(string(permission), ".page") {
			t.Fatalf("action permission implied page access: %s", permission)
		}
	}
}

func TestStoredPageGrantsWorkWithoutMigration(t *testing.T) {
	organisation, user, project := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	membership := OrganisationUser{OrganisationId: organisation, UserId: user, Status: MembershipStatusActive}
	for legacy, canonical := range legacyPageAliases {
		t.Run(string(legacy), func(t *testing.T) {
			definition, _ := LookupPermission(canonical)
			for _, stored := range []Permission{legacy, canonical} {
				role := Role{
					Id: primitive.NewObjectID(), OrganisationId: organisation, Name: "Page only",
					Kind: RoleKindCustom, AssignableAt: definition.Level, Grants: SelectionSelected,
					Permissions: []Permission{stored}, IsActive: true,
				}
				scope := RoleAssignmentScope{Level: definition.Level, ProjectId: project, Resources: SelectionAll}
				if definition.Level == LevelOrganisation {
					scope = RoleAssignmentScope{Level: LevelOrganisation, Projects: SelectionAll}
				}
				assignment := RoleAssignment{
					Id: primitive.NewObjectID(), OrganisationId: organisation, RoleId: role.Id,
					Principal: Principal{Type: PrincipalTypeUser, Id: user}, Scope: scope,
					Source: AssignmentSourceManual, IsActive: true,
				}
				if err := role.Validate(); err != nil {
					t.Fatal(err)
				}
				access := NewAccess(membership, nil, []Role{role}, []RoleAssignment{assignment}, time.Now())
				for _, requested := range []Permission{legacy, canonical} {
					allowed := access.Project(requested, project)
					if definition.Level == LevelOrganisation {
						allowed = access.Organisation(requested, project)
					}
					if !allowed {
						t.Errorf("stored %q did not grant %q", stored, requested)
					}
				}
				if got := access.Permissions(project); !slices.Equal(got, []Permission{canonical}) {
					t.Errorf("page-only role granted %v", got)
				}
				if access.Project(PermissionDevicesRead, project) || access.Project(PermissionMediaRead, project) {
					t.Fatal("page permission granted resource access")
				}
				if definition.Level == LevelProject && access.Project(canonical, primitive.NewObjectID()) {
					t.Fatal("page permission leaked into another project")
				}
			}
		})
	}
}
