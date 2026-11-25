package models

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validConnection(t *testing.T) IdentityProviderConnection {
	t.Helper()
	domain, err := NewIdentityProviderDomain(" Example.COM. ")
	if err != nil {
		t.Fatal(err)
	}
	return IdentityProviderConnection{
		Id: primitive.NewObjectID(), OrganisationId: primitive.NewObjectID(),
		Name: "Entra ID", Protocol: IdentityProviderProtocolOIDC,
		Issuer: "https://login.example.com/tenant/v2.0", ClientId: "hub", ClientSecret: "s3cret",
		Domains: []IdentityProviderDomain{domain}, Enabled: true,
	}
}

func TestIdentityProviderConnectionValidate(t *testing.T) {
	if err := validConnection(t).Validate(); err != nil {
		t.Fatalf("valid connection: %v", err)
	}
	cases := map[string]func(*IdentityProviderConnection){
		"no name":             func(c *IdentityProviderConnection) { c.Name = "" },
		"padded name":         func(c *IdentityProviderConnection) { c.Name = " Entra" },
		"long name":           func(c *IdentityProviderConnection) { c.Name = strings.Repeat("x", 65) },
		"saml":                func(c *IdentityProviderConnection) { c.Protocol = "saml" },
		"http issuer":         func(c *IdentityProviderConnection) { c.Issuer = "http://login.example.com" },
		"issuer with query":   func(c *IdentityProviderConnection) { c.Issuer = "https://login.example.com/?x=1" },
		"not a url":           func(c *IdentityProviderConnection) { c.Issuer = "login" },
		"no client id":        func(c *IdentityProviderConnection) { c.ClientId = " " },
		"claim with space":    func(c *IdentityProviderConnection) { c.Claims.Groups = "my groups" },
		"unnormalised domain": func(c *IdentityProviderConnection) { c.Domains[0].Name = "Example.com" },
		"bad domain":          func(c *IdentityProviderConnection) { c.Domains[0].Name = "localhost" },
		"no token":            func(c *IdentityProviderConnection) { c.Domains[0].VerificationToken = "" },
		"duplicate domain":    func(c *IdentityProviderConnection) { c.Domains = append(c.Domains, c.Domains[0]) },
	}
	for name, change := range cases {
		connection := validConnection(t)
		change(&connection)
		if err := connection.Validate(); !errors.Is(err, ErrIdentityProviderInvalid) {
			t.Errorf("%s: got %v, want ErrIdentityProviderInvalid", name, err)
		}
	}
	// A local test provider may use http on a loopback host.
	local := validConnection(t)
	local.Issuer = "http://localhost:8080/realms/hub"
	if err := local.Validate(); err != nil {
		t.Errorf("loopback http issuer: %v", err)
	}
}

func TestIdentityProviderDomains(t *testing.T) {
	domain, err := NewIdentityProviderDomain(" Example.COM. ")
	if err != nil {
		t.Fatal(err)
	}
	if domain.Name != "example.com" || len(domain.VerificationToken) != 32 || domain.Verified() {
		t.Fatalf("domain %+v", domain)
	}
	if domain.VerificationRecordName() != "_hub-verification.example.com" ||
		domain.VerificationRecordValue() != "hub-verification="+domain.VerificationToken {
		t.Fatalf("record %s = %s", domain.VerificationRecordName(), domain.VerificationRecordValue())
	}
	other, _ := NewIdentityProviderDomain("example.com")
	if other.VerificationToken == domain.VerificationToken {
		t.Fatal("tokens repeat")
	}
	for _, bad := range []string{"", "com", "-example.com", "exa mple.com", "example..com"} {
		if _, err := NewIdentityProviderDomain(bad); !errors.Is(err, ErrIdentityProviderInvalid) {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestIdentityProviderAnswersForVerifiedDomainsOnly(t *testing.T) {
	connection := validConnection(t)
	if connection.AnswersFor("kim@example.com") {
		t.Fatal("unverified domain answers")
	}
	connection.Domains[0].VerifiedAt = time.Now()
	for email, want := range map[string]bool{
		"kim@example.com": true, "Kim@EXAMPLE.com": true, "kim@other.com": false,
		"kim@sub.example.com": false, "kim": false, "kim@": false,
	} {
		if got := connection.AnswersFor(email); got != want {
			t.Errorf("AnswersFor(%q) = %v, want %v", email, got, want)
		}
	}
	connection.Enabled = false
	if connection.AnswersFor("kim@example.com") {
		t.Fatal("disabled connection answers")
	}
}

func TestIdentityProviderClaimsDefault(t *testing.T) {
	var claims IdentityProviderClaims
	if claims.SubjectClaim() != "sub" || claims.EmailClaim() != "email" || claims.GroupsClaim() != "groups" {
		t.Fatalf("defaults %s %s %s", claims.SubjectClaim(), claims.EmailClaim(), claims.GroupsClaim())
	}
	claims.Groups = "roles"
	if claims.GroupsClaim() != "roles" {
		t.Fatal("configured claim ignored")
	}
}

func TestIdentityProviderSecretIsNeverSent(t *testing.T) {
	connection := validConnection(t)
	data, err := json.Marshal(connection)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "s3cret") || strings.Contains(string(data), "clientSecret") {
		t.Fatalf("secret in JSON: %s", data)
	}
	stored, err := bson.Marshal(connection)
	if err != nil {
		t.Fatal(err)
	}
	var back IdentityProviderConnection
	if err := bson.Unmarshal(stored, &back); err != nil || back.ClientSecret != "s3cret" {
		t.Fatalf("secret not stored: %v %q", err, back.ClientSecret)
	}
}

func TestIdentityLinkValidate(t *testing.T) {
	link := IdentityLink{UserId: primitive.NewObjectID(), OrganisationId: primitive.NewObjectID(), ConnectionId: primitive.NewObjectID(), Subject: "abc"}
	if err := link.Validate(); err != nil {
		t.Fatal(err)
	}
	link.Subject = " "
	if err := link.Validate(); !errors.Is(err, ErrIdentityProviderInvalid) {
		t.Fatalf("blank subject: %v", err)
	}
}

func TestManagingIdentityProvidersIsOwnerOnlyAndSessionOnly(t *testing.T) {
	definition, ok := LookupPermission(PermissionOrganisationManageIdentityProvider)
	if !ok || definition.Level != LevelOrganisation || !definition.TenantWide || !definition.OwnerOnly {
		t.Fatalf("definition %+v", definition)
	}
	if slices.Contains(AccessTokenScopes(), PermissionOrganisationManageIdentityProvider) {
		t.Fatal("grantable to access tokens")
	}
}

func TestSourcesAreOptionalOnStoredDocuments(t *testing.T) {
	var membership OrganisationUser
	var member TeamMember
	for _, document := range []any{&membership, &member} {
		if err := bson.Unmarshal(mustBSON(t, bson.M{"userId": primitive.NewObjectID()}), document); err != nil {
			t.Fatal(err)
		}
	}
	if membership.Source != "" || member.Source != "" {
		t.Fatalf("sources %q %q", membership.Source, member.Source)
	}
	data, _ := json.Marshal(OrganisationUser{Source: MembershipSourceIdentityProvider})
	if !strings.Contains(string(data), `"source":"identityProvider"`) {
		t.Fatalf("membership JSON %s", data)
	}
}

func mustBSON(t *testing.T, document bson.M) []byte {
	t.Helper()
	data, err := bson.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
