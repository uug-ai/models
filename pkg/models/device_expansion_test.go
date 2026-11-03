package models

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type fakeDeviceSource struct {
	sites  map[string]Site
	groups map[string]Group
	err    error
	reads  []string
}

func (s *fakeDeviceSource) Sites(_ context.Context, ids []string) ([]Site, error) {
	s.reads = append(s.reads, "sites")
	var found []Site
	for _, id := range ids {
		if site, ok := s.sites[id]; ok {
			found = append(found, site)
		}
	}
	return found, s.err
}

func (s *fakeDeviceSource) Groups(_ context.Context, ids []string) ([]Group, error) {
	s.reads = append(s.reads, "groups")
	var found []Group
	for _, id := range ids {
		if group, ok := s.groups[id]; ok {
			found = append(found, group)
		}
	}
	return found, s.err
}

func expansionSource() *fakeDeviceSource {
	return &fakeDeviceSource{
		sites: map[string]Site{
			"site-a": {Devices: []string{"a1", "a2"}, Groups: []string{"cameras-x"}},
			"site-b": {Devices: []string{"b1"}},
			// Points back at a site group containing itself.
			"site-c": {Devices: []string{"c1"}, Groups: []string{"site-group", "missing"}},
		},
		groups: map[string]Group{
			"cameras-x":  {GroupType: GroupTypeCamera, Devices: []string{"x1", "a1"}},
			"site-group": {GroupType: GroupTypeSite, Sites: []string{"site-b", "site-c"}, Devices: []string{"ignored"}},
			"untyped":    {Devices: []string{"ignored"}},
		},
	}
}

func TestExpandDeviceKeys(t *testing.T) {
	tests := []struct {
		name          string
		sites, groups []string
		cameras, want []string
	}{
		{"site devices", []string{"site-b"}, nil, nil, []string{"b1"}},
		{"camera groups inside a site", []string{"site-a"}, nil, nil, []string{"a1", "a2", "x1"}},
		{"camera group", nil, []string{"cameras-x"}, nil, []string{"a1", "x1"}},
		{"site group sites, nested both ways", nil, []string{"site-group"}, nil, []string{"b1", "c1"}},
		{"site group listed inside a site", []string{"site-c"}, nil, nil, []string{"b1", "c1"}},
		{"cameras", nil, nil, []string{"k2", "k1", "k1", ""}, []string{"k1", "k2"}},
		{"union", []string{"site-b"}, []string{"cameras-x"}, []string{"k1"}, []string{"a1", "b1", "k1", "x1"}},
		{"untyped groups and unknown ids select nothing", []string{"nowhere"}, []string{"untyped", "nowhere"}, nil, []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ExpandDeviceKeys(context.Background(), expansionSource(), test.sites, test.groups, test.cameras)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestExpandDeviceKeysReadsEachOnce(t *testing.T) {
	source := expansionSource()
	if _, err := ExpandDeviceKeys(context.Background(), source, []string{"site-c", "site-c"}, []string{"site-group"}, nil); err != nil {
		t.Fatal(err)
	}
	// site-c and site-group, then site-b, then nothing new.
	if want := []string{"sites", "groups", "sites"}; !slices.Equal(source.reads, want) {
		t.Errorf("reads %v, want %v", source.reads, want)
	}

	source = expansionSource()
	if _, err := ExpandDeviceKeys(context.Background(), source, nil, nil, []string{"k1"}); err != nil || len(source.reads) > 0 {
		t.Errorf("cameras only: reads %v, err %v", source.reads, err)
	}
}

func TestExpandDeviceKeysFailsOnReadError(t *testing.T) {
	source := expansionSource()
	source.err = errors.New("down")
	if keys, err := ExpandDeviceKeys(context.Background(), source, []string{"site-a"}, nil, []string{"k1"}); err == nil || keys != nil {
		t.Errorf("keys %v, err %v; want no keys and the error", keys, err)
	}
}
