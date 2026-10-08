package models

import (
	"reflect"
	"sort"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type membershipFixture struct {
	groups map[string]Group
	sites  map[string]Site
}

func (f membershipFixture) lists() ([]Group, []Site) {
	var groups []Group
	var sites []Site
	for _, group := range f.groups {
		groups = append(groups, group)
	}
	for _, site := range f.sites {
		sites = append(sites, site)
	}
	return groups, sites
}

// hubAPIGroupDevices mirrors hub-api's GetDevicesFromGroup expansion.
func (f membershipFixture) hubAPIGroupDevices(name string) map[string]bool {
	group := f.groups[name]
	devices := map[string]bool{}
	for _, key := range group.Devices {
		devices[key] = true
	}
	if group.GroupType == GroupTypeSite {
		for _, siteId := range group.Sites {
			for _, site := range f.sites {
				if site.Id.Hex() != siteId {
					continue
				}
				for _, key := range site.Devices {
					devices[key] = true
				}
				for _, nestedId := range site.Groups {
					for _, nested := range f.groups {
						if nested.Id.Hex() == nestedId {
							for _, key := range nested.Devices {
								devices[key] = true
							}
						}
					}
				}
			}
		}
	}
	return devices
}

func newMembershipFixture() membershipFixture {
	id := func() primitive.ObjectID { return primitive.NewObjectID() }
	f := membershipFixture{groups: map[string]Group{}, sites: map[string]Site{}}
	f.groups["lobby"] = Group{Id: id(), GroupType: "camera", Devices: []string{"cam-a", "cam-b"}}
	f.groups["parking"] = Group{Id: id(), GroupType: "camera", Devices: []string{"cam-c"}}
	f.groups["deep"] = Group{Id: id(), GroupType: "camera", Devices: []string{"cam-e"}}
	f.sites["north"] = Site{Id: id(), Devices: []string{"cam-d"}, Groups: []string{f.groups["parking"].Id.Hex()}}
	f.sites["south"] = Site{Id: id(), Devices: []string{"cam-a"}}
	f.groups["region"] = Group{Id: id(), GroupType: GroupTypeSite, Sites: []string{f.sites["north"].Id.Hex()}}
	f.groups["everything"] = Group{Id: id(), GroupType: GroupTypeSite, Devices: []string{"cam-x"},
		Sites: []string{f.sites["north"].Id.Hex(), f.sites["south"].Id.Hex()}}
	// Sites listed on a camera group do not count, and nested groups are ignored.
	f.groups["mislabelled"] = Group{Id: id(), GroupType: "camera", Sites: []string{f.sites["north"].Id.Hex()}}
	f.groups["nested"] = Group{Id: id(), GroupType: "camera", Groups: []string{f.groups["lobby"].Id.Hex()}}
	// Only one level: a site group reached through another site group's site does not count.
	f.sites["east"] = Site{Id: id(), Devices: []string{"cam-e"}}
	f.groups["east-region"] = Group{Id: id(), GroupType: GroupTypeSite, Sites: []string{f.sites["east"].Id.Hex()}}
	f.sites["west"] = Site{Id: id(), Groups: []string{f.groups["east-region"].Id.Hex()}}
	f.groups["west-region"] = Group{Id: id(), GroupType: GroupTypeSite, Sites: []string{f.sites["west"].Id.Hex()}}
	return f
}

func TestEffectiveDeviceGroupIds(t *testing.T) {
	f := newMembershipFixture()
	groups, sites := f.lists()
	hex := func(names ...string) []string {
		ids := []string{}
		for _, name := range names {
			ids = append(ids, f.groups[name].Id.Hex())
		}
		sort.Strings(ids)
		return ids
	}
	for device, want := range map[string][]string{
		"cam-a":   hex("lobby", "everything"),             // direct, and via site south
		"cam-c":   hex("parking", "region", "everything"), // via site north's nested group
		"cam-d":   hex("region", "everything"),            // via site north
		"cam-e":   hex("deep", "east-region"),             // not west-region: two levels
		"cam-x":   hex("everything"),                      // direct on a site group
		"unknown": {},
		"":        {},
	} {
		if got := EffectiveDeviceGroupIds(device, groups, sites); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v, want %v", device, got, want)
		}
	}
}

// The reverse lookup agrees with hub-api's forward expansion for every device.
func TestEffectiveDeviceGroupIdsMatchesHubAPIExpansion(t *testing.T) {
	f := newMembershipFixture()
	groups, sites := f.lists()
	for _, device := range []string{"cam-a", "cam-b", "cam-c", "cam-d", "cam-e", "cam-x", "unknown"} {
		var want []string
		for name, group := range f.groups {
			if f.hubAPIGroupDevices(name)[device] {
				want = append(want, group.Id.Hex())
			}
		}
		sort.Strings(want)
		if want == nil {
			want = []string{}
		}
		if got := EffectiveDeviceGroupIds(device, groups, sites); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: reverse %v, hub-api expansion %v", device, got, want)
		}
	}
}
