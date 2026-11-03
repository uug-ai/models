package models

import (
	"context"
	"slices"
)

// GroupTypeCamera marks a group whose members are the devices it lists.
const GroupTypeCamera = "camera"

// DeviceSource reads the sites and groups a device expansion walks. An
// implementation limits every read to one organisation (and project) and
// leaves out ids it does not find: they select nothing.
type DeviceSource interface {
	Sites(ctx context.Context, ids []string) ([]Site, error)
	Groups(ctx context.Context, ids []string) ([]Group, error)
}

// ExpandDeviceKeys returns the keys of the devices selected by sites, groups
// and device keys, sorted. It is the one rule every service uses to compile
// device access, and the same rule hub-api's GetAllowedDeviceKeys applies:
//
//   - a site selects its devices and whatever its groups select;
//   - a camera group selects its devices;
//   - a site group selects whatever its sites select; devices listed on a
//     site group itself are ignored.
//
// Nesting is followed to any depth, reading each site and group once. A read
// error fails the expansion, because a partial result shows or hides the
// wrong devices.
func ExpandDeviceKeys(ctx context.Context, source DeviceSource, siteIds, groupIds, deviceKeys []string) ([]string, error) {
	keys := map[string]bool{}
	addKeys(keys, deviceKeys)
	seen := map[string]bool{}
	sites, groups := siteIds, groupIds
	for len(sites) > 0 || len(groups) > 0 {
		var err error
		if sites = unseen(sites, "site:", seen); len(sites) > 0 {
			var found []Site
			if found, err = source.Sites(ctx, sites); err != nil {
				return nil, err
			}
			sites = nil
			for _, site := range found {
				addKeys(keys, site.Devices)
				groups = append(groups, site.Groups...)
			}
		}
		if groups = unseen(groups, "group:", seen); len(groups) > 0 {
			var found []Group
			if found, err = source.Groups(ctx, groups); err != nil {
				return nil, err
			}
			groups = nil
			for _, group := range found {
				switch group.GroupType {
				case GroupTypeCamera:
					addKeys(keys, group.Devices)
				case GroupTypeSite:
					sites = append(sites, group.Sites...)
				}
			}
		}
	}
	expanded := make([]string, 0, len(keys))
	for key := range keys {
		expanded = append(expanded, key)
	}
	slices.Sort(expanded)
	return expanded, nil
}

func addKeys(keys map[string]bool, more []string) {
	for _, key := range more {
		if key != "" {
			keys[key] = true
		}
	}
}

// unseen returns the ids not read yet, once each, and marks them read.
func unseen(ids []string, kind string, seen map[string]bool) []string {
	var fresh []string
	for _, id := range ids {
		if id != "" && !seen[kind+id] {
			seen[kind+id] = true
			fresh = append(fresh, id)
		}
	}
	return fresh
}
