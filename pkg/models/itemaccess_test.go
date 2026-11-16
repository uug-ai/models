package models

import (
	"encoding/json"
	"testing"
)

func TestItemAccessJSON(t *testing.T) {
	cases := []struct {
		access ItemAccess
		want   string
	}{
		// An item never carries a mode; nothing held is an empty list.
		{ItemAccess{Permissions: []Permission{}}, `{"permissions":[]}`},
		{ItemAccess{Mode: AuthorizationModeRBAC, Permissions: []Permission{PermissionDevicesRead}}, `{"mode":"rbac","permissions":["devices.read"]}`},
	}
	for _, tc := range cases {
		got, err := json.Marshal(tc.access)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != tc.want {
			t.Errorf("got %s, want %s", got, tc.want)
		}
	}
}
