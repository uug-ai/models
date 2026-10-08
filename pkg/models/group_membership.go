package models

import "sort"

// GroupTypeSite marks a group whose members are the devices of its sites.
const GroupTypeSite = "site"

// EffectiveDeviceGroupIds returns the IDs of every group deviceKey belongs to,
// sorted, for WorkflowDevice.GroupIds. It is the reverse of hub-api's group
// expansion (groups v20260101 GetDevicesFromGroup) and must stay in sync with
// it. A device belongs to a group when:
//
//   - the group lists it in Devices (any group type), or
//   - the group is a site group and one of its Sites lists the device, or lists
//     a group whose Devices contain it (one level; Group.Groups is ignored
//     because it is never written).
//
// groups and sites are candidates the caller loaded from trusted, tenant-scoped
// data: at least every group listing the device, every site listing the device
// or one of those groups, and every site group referencing those sites. Extra
// candidates are harmless.
//
// Known possible failure point: resolving these candidates costs several
// queries per workflow hand-off (groups by device, sites by device or group,
// site groups by site), and correctness depends on this rule matching hub-api.
func EffectiveDeviceGroupIds(deviceKey string, groups []Group, sites []Site) []string {
	if deviceKey == "" {
		return []string{}
	}
	direct := map[string]bool{}
	for _, group := range groups {
		if containsString(group.Devices, deviceKey) {
			direct[group.Id.Hex()] = true
		}
	}
	memberSites := map[string]bool{}
	for _, site := range sites {
		if containsString(site.Devices, deviceKey) {
			memberSites[site.Id.Hex()] = true
			continue
		}
		for _, groupId := range site.Groups {
			if direct[groupId] {
				memberSites[site.Id.Hex()] = true
				break
			}
		}
	}
	effective := map[string]bool{}
	for id := range direct {
		effective[id] = true
	}
	for _, group := range groups {
		if group.GroupType != GroupTypeSite {
			continue
		}
		for _, siteId := range group.Sites {
			if memberSites[siteId] {
				effective[group.Id.Hex()] = true
				break
			}
		}
	}
	ids := make([]string, 0, len(effective))
	for id := range effective {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
