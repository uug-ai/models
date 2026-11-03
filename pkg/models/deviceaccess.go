package models

import (
	"context"
	"slices"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ItemAccess is what the caller may do, on the profile for the selected
// project and on each item for that item (RBAC spec 4.13). Permissions are
// catalog names.
type ItemAccess struct {
	Permissions []Permission `json:"permissions"`
}

// DeviceCoverage is the devices a permission covers in a project: all of
// them, or those with a key in Keys.
type DeviceCoverage struct {
	All  bool
	Keys []string
}

// Includes reports whether the coverage includes a device.
func (c DeviceCoverage) Includes(deviceKey string) bool {
	return c.All || slices.Contains(c.Keys, deviceKey)
}

// Empty reports whether the coverage includes no device.
func (c DeviceCoverage) Empty() bool {
	return !c.All && len(c.Keys) == 0
}

// DeviceAccess answers what a member may do on each device of one project:
// the device permissions of every grant whose selection covers the device,
// narrowed by the device's ceilings. Compile it once per request with
// Access.CompileDeviceAccess; answering is then a set lookup per device.
type DeviceAccess struct {
	owner  bool
	grants []deviceGrant
}

type deviceGrant struct {
	all         bool
	keys        map[string]bool
	permissions []Permission
}

// CompileDeviceAccess expands, once, which devices each of the member's grants
// in the project covers, with the shared site and group rule
// (ExpandDeviceKeys). Grants holding no device permission are skipped.
func (a Access) CompileDeviceAccess(ctx context.Context, source DeviceSource, projectId primitive.ObjectID) (DeviceAccess, error) {
	compiled := DeviceAccess{owner: a.owner}
	if a.owner || projectId.IsZero() {
		return compiled, nil
	}
	for _, grant := range a.grants {
		if grant.scope.Level != LevelProject || grant.scope.ProjectId != projectId {
			continue
		}
		permissions := deviceScoped(grant.permissions)
		if len(permissions) == 0 {
			continue
		}
		compiledGrant := deviceGrant{permissions: permissions}
		if grant.scope.Resources == SelectionAll {
			compiledGrant.all = true
		} else {
			keys, err := ExpandDeviceKeys(ctx, source, hexIds(grant.scope.SiteIds), hexIds(grant.scope.GroupIds), grant.scope.DeviceKeys)
			if err != nil {
				return DeviceAccess{}, err
			}
			compiledGrant.keys = make(map[string]bool, len(keys))
			for _, key := range keys {
				compiledGrant.keys[key] = true
			}
		}
		compiled.grants = append(compiled.grants, compiledGrant)
	}
	return compiled, nil
}

// Permissions returns the member's permissions on one device, in catalog
// order: those of every grant covering it, minus what the device's ceilings
// rule out. A device no grant covers gets none.
func (d DeviceAccess) Permissions(deviceKey string, ceilings *DeviceFeaturePermissions) []Permission {
	var held []Permission
	if d.owner {
		held = deviceScoped(AllPermissions())
	} else {
		for _, grant := range d.grants {
			if grant.all || grant.keys[deviceKey] {
				held = append(held, grant.permissions...)
			}
		}
	}
	return ApplyDeviceCeilings(ExpandPermissions(held), ceilings)
}

// Allows reports whether the member holds a permission on one device.
func (d DeviceAccess) Allows(permission Permission, deviceKey string, ceilings *DeviceFeaturePermissions) bool {
	return slices.Contains(d.Permissions(deviceKey, ceilings), permission)
}

// Covered returns the devices a permission covers, from the grants already
// compiled: all of them for an owner or a grant on all resources, otherwise
// the devices of the grants holding it. Device ceilings are not applied; use
// Allows for one device.
func (d DeviceAccess) Covered(permission Permission) DeviceCoverage {
	if d.owner {
		return DeviceCoverage{All: true}
	}
	var keys []string
	for _, grant := range d.grants {
		if !slices.Contains(grant.permissions, permission) {
			continue
		}
		if grant.all {
			return DeviceCoverage{All: true}
		}
		for key := range grant.keys {
			if !slices.Contains(keys, key) {
				keys = append(keys, key)
			}
		}
	}
	slices.Sort(keys)
	return DeviceCoverage{Keys: keys}
}

// ApplyDeviceCeilings removes the graded permissions a device's ceilings rule
// out, for every member, owners included, as the frontend always applied
// them. A ceiling of 0 (unset) limits nothing and 1 removes the feature;
// higher values keep the grades the legacy levels allow:
//
//   - PTZ: 2 zoom, 3 pan/tilt (no zoom), 4 zoom, pan/tilt and control,
//     5 and up everything including presets;
//   - live view: 2 preview, 3 and up preview and stream;
//   - remote config and IO: 2 and up allowed.
func ApplyDeviceCeilings(permissions []Permission, ceilings *DeviceFeaturePermissions) []Permission {
	if ceilings == nil {
		return permissions
	}
	var removed []Permission
	removed = append(removed, ceilingRemoves(ceilings.PTZ, ptzGrades)...)
	removed = append(removed, ceilingRemoves(ceilings.Liveview, liveGrades)...)
	removed = append(removed, ceilingRemoves(ceilings.RemoteConfig, remoteConfigGrades)...)
	removed = append(removed, ceilingRemoves(ceilings.IO, ioGrades)...)
	if len(removed) == 0 {
		return permissions
	}
	kept := make([]Permission, 0, len(permissions))
	for _, permission := range permissions {
		if !slices.Contains(removed, permission) {
			kept = append(kept, permission)
		}
	}
	return kept
}

// A feature's grades: the permissions each ceiling level allows, from level
// 1 (nothing) up; levels past the last allow what the last allows.
var (
	ptzGrades = [][]Permission{
		{},
		{PermissionPTZZoom},
		{PermissionPTZPanTilt},
		{PermissionPTZZoom, PermissionPTZPanTilt, PermissionPTZControl},
		{PermissionPTZZoom, PermissionPTZPanTilt, PermissionPTZControl, PermissionPTZPresets},
	}
	liveGrades         = [][]Permission{{}, {PermissionLivePreview}, {PermissionLivePreview, PermissionLiveStream}}
	remoteConfigGrades = [][]Permission{{}, {PermissionDevicesConfigure, PermissionDevicesManageRecording}}
	ioGrades           = [][]Permission{{}, {PermissionIOTrigger}}
)

// DeviceCeilingPermissions returns the permissions a device's ceilings can
// remove (ApplyDeviceCeilings), in catalog order.
func DeviceCeilingPermissions() []Permission {
	var graded []Permission
	for _, grades := range [][][]Permission{ptzGrades, liveGrades, remoteConfigGrades, ioGrades} {
		graded = append(graded, grades[len(grades)-1]...)
	}
	return ExpandPermissions(graded)
}

// ceilingRemoves returns the feature's permissions a ceiling level rules out.
func ceilingRemoves(level PermissionLevel, grades [][]Permission) []Permission {
	if level <= 0 {
		return nil
	}
	index := min(int(level), len(grades)) - 1
	allowed, all := grades[index], grades[len(grades)-1]
	var removes []Permission
	for _, permission := range all {
		if !slices.Contains(allowed, permission) {
			removes = append(removes, permission)
		}
	}
	return removes
}

// deviceScoped keeps the permissions that act on one device: the catalog's
// device and IO permissions.
func deviceScoped(permissions []Permission) []Permission {
	var kept []Permission
	for _, permission := range permissions {
		if slices.Contains(devicePermissions, permission) || slices.Contains(ioPermissions, permission) {
			kept = append(kept, permission)
		}
	}
	return kept
}

func hexIds(ids []primitive.ObjectID) []string {
	hexes := make([]string, 0, len(ids))
	for _, id := range ids {
		hexes = append(hexes, id.Hex())
	}
	return hexes
}
