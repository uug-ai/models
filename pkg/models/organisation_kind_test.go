package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestOrganisationKind(t *testing.T) {
	tests := []struct {
		kind  OrganisationKind
		want  OrganisationKind
		valid bool
	}{
		{"", OrganisationKindBusiness, true},
		{OrganisationKindBusiness, OrganisationKindBusiness, true},
		{OrganisationKindPersonal, OrganisationKindPersonal, true},
		{"family", OrganisationKindBusiness, false},
	}
	for _, test := range tests {
		organisation := Organisation{Kind: test.kind}
		if got := organisation.EffectiveKind(); got != test.want {
			t.Errorf("%q: effective %q, want %q", test.kind, got, test.want)
		}
		if err := organisation.ValidateKind(); (err == nil) != test.valid {
			t.Errorf("%q: validate %v", test.kind, err)
		}
	}
	// An organisation without a kind stores none, so existing documents and
	// new business organisations look the same.
	data, _ := bson.Marshal(Organisation{})
	if _, err := bson.Raw(data).LookupErr("kind"); err == nil {
		t.Error("an unset kind is stored")
	}
	data, _ = bson.Marshal(Organisation{Kind: OrganisationKindPersonal})
	if value, err := bson.Raw(data).LookupErr("kind"); err != nil || value.StringValue() != "personal" {
		t.Errorf("personal kind stored as %v, %v", value, err)
	}
}
