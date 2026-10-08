# Models

Shared data models and TypeScript type definitions for the UUG AI ecosystem.

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://go.dev/)
[![codecov](https://codecov.io/gh/uug-ai/models/graph/badge.svg?token=GD113W0PCL)](https://codecov.io/gh/uug-ai/models)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/uug-ai/models?status.svg)](https://godoc.org/github.com/uug-ai/models)
[![Release](https://img.shields.io/github/release/uug-ai/models.svg)](https://github.com/uug-ai/models/releases/latest)

A comprehensive Go package providing type-safe data models for media management, device control, authentication, analytics, and more. Includes automatic TypeScript type generation for seamless cross-language development.

📋 **[Best Practices Guide](./BEST_PRACTICES.md)** - Comprehensive guidelines for defining Go types in this repository

## Features

- **Comprehensive Type System**: 30+ models covering media, devices, authentication, analytics, and more
- **Cross-Language Support**: Automatic TypeScript type generation from Go structs
- **Database Ready**: Full MongoDB support with BSON tags and validation
- **API Optimized**: JSON serialization with proper omitempty handling
- **Type Safety**: Compile-time type checking in both Go and TypeScript
- **Auto-Discovery**: Automatic model detection and TypeScript generation
- **Production Ready**: Battle-tested models used in production systems

### Alert configuration visibility

`CustomAlert.private` defaults to false (including historical documents without
the field). When true, configuration access is restricted to the creator in
`user_id`, including against other owners/admins. This is separate from
organisation/project ownership and does not change alert execution, notification
recipients, generated events, or auditing. Only the creator can change privacy;
`AlertPatch.private` supports explicit false to restore shared visibility.

## Installation

### Go
```bash
go get github.com/uug-ai/models
```

### TypeScript
```bash
npm install @uug-ai/models
# or
yarn add @uug-ai/models
```

## Quick Start

### Go Usage

```go
package main

import (
    "github.com/uug-ai/models/pkg/models"
    "time"
)

func main() {
    // Create a media record
    media := models.Media{
        FileName:       "video_001.mp4",
        StartTimestamp: time.Now().Unix(),
        Duration:       600000, // 10 minutes in milliseconds
        DeviceId:       "camera-001",
        UserId:         "user-123",
        VideoUrl:       "https://cdn.example.com/videos/video_001.mp4",
        ThumbnailUrl:   "https://cdn.example.com/thumbs/video_001.jpg",
        SpriteUrl:      "https://cdn.example.com/sprites/video_001.jpg",
        SpriteInterval: 10,
    }
    
    // Create a device
    device := models.Device{
        Name:        "Front Door Camera",
        DeviceId:    "camera-001",
        Type:        "camera",
        Status:      "online",
        GroupId:     "group-001",
        Coordinates: &models.Coordinates{
            Latitude:  37.7749,
            Longitude: -122.4194,
        },
    }
}
```

### TypeScript Usage

```typescript
import { Media, Device, models } from '@uug-ai/models';

// Direct import
const media: Media = {
  fileName: "video_001.mp4",
  startTimestamp: 1640995200,
  duration: 600000,
  deviceId: "camera-001",
  userId: "user-123",
  videoUrl: "https://cdn.example.com/videos/video_001.mp4"
};

// Namespace import
const device: models.Device = {
  name: "Front Door Camera",
  deviceId: "camera-001",
  type: "camera",
  status: "online"
};
```

## Core Concepts

### Pipeline transport compatibility

`MonitorStage` JSON serialization omits `user.audit` and the audit field of
every nested `user.master`. Pipeline workers do not use this audit history.
`json.Unmarshal` into `PipelineEvent` accepts and ignores absent, null, legacy
array, and current object audit values before queue handlers run; forwarding
and worker fanout through `json.Marshal` omit them for old and new readers.
All other user and pipeline fields retain their existing JSON representation.
Reusing a decode destination retains absent fields as usual; decoding a user
object clears its audit instead of retaining stale audit data.

This is a pipeline-only transport boundary: standalone `User`/`Audit` JSON
(including API responses) and all BSON persistence remain unchanged. No data
migration is required.

### VLM scene and stable states

`MediaMetadata.VLM` (`metadata.vlm` on media) stores the structured
vision-language model result: a `VLMScene` (summary, static elements, zones,
lighting) and `VLMObservation` entries (`marker`, `event`, `tags`) that map onto
markers in category `vlm`, for example marker `person`, event `walking`, tags
`red`, `sweater`.

`DeviceMetadata.VLM` (`metadata.vlm` on devices) holds up to
`VLMMaxStableStates` named `VLMStableState` entries promoted from a recording's
scene. Promoting the same recording again replaces its entry. A promotion or a
later update may rename a state and drop scene values, but never add them; a
state can also be deleted. VLM analysis uses these states as the device's "normal" scene so it reports changes instead of
repeating the scene. Both fields are optional and omitted when empty, so
existing media and devices need no migration.

### Workflow run explanations (models-only additions)

#### Trigger matches

`WorkflowRun.TriggerMatch` (`triggerMatch` in JSON, `triggermatch` in BSON) is an
optional, immutable run-level record of the first automatic trigger that opened
the run. It is not a stage or a stage-execution record. `Workflow.MatchAutomaticTrigger`
shares selection with `AutomaticMatches` and returns a detached snapshot containing:

- `index`: the zero-based position in the normalized trigger list at selection
  time, not a stable identifier into an edited workflow.
- `evaluatedAtMs`: the supplied schedule-evaluation instant in Unix milliseconds
  (usually recording time), not the engine's processing or run-open time.
- `trigger`: only the selected trigger definition, with its effective automatic
  type; no raw event envelope, credentials, or other triggers.
- `matchedEdgeIds`: every matching automatic Start edge's stable ID, in graph
  order; absent for legacy trigger lists. The first trigger also carries `edgeId`.

`WorkflowRun.MatchedStartEdgeIds` (`matchedStartEdgeIds` JSON, `matchedstartedgeids` BSON) stores
the same IDs separately for safe list projections. `WorkflowRunOverview` exposes
only this ID list, never the full trigger definition or its predicates.

Logical `WorkflowCondition` groups use
`{"op":"all"|"any","conditions":[...]}`. Children keep absolute envelope paths;
`anyMatch.match` still contains scalar predicates relative to one array object.
Groups require nonempty children and omit `path` (or use `""`), `value` (or use
`null`), and `match`. Scalar/`anyMatch` conditions cannot carry `conditions`.
Validation and evaluation cap logical nesting at eight levels and each top-level
set at 256 nodes, including groups and `anyMatch` relative predicates. Empty
top-level sets remain unconditional; empty nested groups are invalid. Automatic
Start validation rejects `results.*` throughout nested groups.

Scalar conditions and `anyMatch` predicates may carry an optional `field`: the
contract field ID they were authored with (for example `objectCount` from a Start
contract or `plate` from a stage contract). Editors use it to show a saved value
in that field's section and widget again. It is metadata only: evaluation reads
`path`, `op` and `value`, untagged conditions stay valid, and whether the ID
exists in a contract is checked by the authoring API. Logical groups and the
`anyMatch` condition itself never carry a `field`; IDs must match
`^[A-Za-z][A-Za-z0-9_-]{0,79}$`.

This permits migrating legacy Start predicates exactly: an `all` group containing
the old Start `any` group and an edge's `any` group preserves their independent
choices without expanding edges or retaining hidden settings in the editor.
Old documents that have not been edited still use the shared-condition adapter.

The helper returns nil when nothing matches and an error if snapshot encoding
fails. Callers must synchronize graph-derived triggers before selection.

The engine persists trigger matches when selecting automatic workflows; API
readers must use the matching models release. Writers must ignore inbound
`triggerMatch`, compute the match during authoritative automatic
selection, and preserve persisted data on replay, including an absent match.
Legacy runs and manual or explicitly targeted launches retain nil/unknown;
never backfill a match from `Origin` or the current workflow definition.
No migration is required, and existing workflow-run
millisecond timestamp normalization is unchanged.

#### Stage decisions

`WorkflowRunStageExecutionDetails.Decision` (`execution.decision` in JSON/BSON)
stores one display-only summary produced alongside the engine's existing routing
evaluation. It must never be read to drive dispatch. The summary records
`evaluatedAtMs` in Unix milliseconds, overall `eligible`, and per-edge `outcome`
entries referencing the containing stage's frozen `Needs[index]`.

Outcomes are `passed`, `failed`, `waiting` (upstream unavailable), or
`notEvaluated` (skipped, for example by short-circuiting). This adds no evaluator,
readiness flag, per-condition trace, raw inputs, observed values, or history.

Future writers retain only the latest pending summary and use existing run-state
guards to protect the final dispatch summary from stale overwrites. Ignore
summaries supplied by workers or replays. Keep the last summary for stages that
never dispatch.
Nil means unknown for legacy or uncaptured summaries; never backfill from
results, lifecycle state, or current definitions. Engine/API adoption follows
a models release.

#### Condition root and contract field paths

Conditions never read a stored `WorkflowRun` directly. They read its
credential-free projection, `WorkflowRun.ConditionRoot(mode)`, and
`WorkflowConditionRootSchema()` lists every path in it. That schema is the
reference for every field path a workflow contract declares.

```text
device                          envelope, from WorkflowRun.Device
  deviceKey, deviceName         string
  provider, storageSolution     string
  siteIds.*, groupIds.*         string (array elements)
user
  organisationId                string (projectId and storage are never exposed)
key, operation, traceId         string, run identity
runId                           string, run mode only
inputs.<operation>.…            open: the hand-off result that started the run
  classify                      platform-owned, described by WorkflowClassifyInput
    objectCount                 integer (automatic runs only)
    properties.*                string
    details.*                   one tracked object
      id, classified            string
      colorStr.*                string
      isStatic, valid           boolean
      distance, staticDistance  number
      x, y                      number
      frameWidth, frameHeight   integer
      frame, occurence          integer
results.<operation>.…           open, run mode only: accumulated stage outputs
  anpr                          platform-owned, described by WorkflowAnprResult
    detections.*.tracks.*.meta  one track per vehicle
      plate                     string: uppercase A-Z0-9, no separators; "" when unread
      unread                    boolean
      classifiedAs              string: copied from classify, not recognised by ANPR
    markers.*                   one marker per vehicle
      name                      string: the plate, or "anpr unread"
      duration                  integer seconds
      metadata.engine           string: fast, tesseract, tesseract+opencv, http (read plates only)
      metadata.confidence       number: OCR confidence (read plates only; not clamped for http)
```

| Mode | Used for | Contains |
| --- | --- | --- |
| `trigger` | automatic Start trigger matching, before a run exists | envelope, identity, `inputs` (omitted when empty) |
| `run` | conditions between stages during a run | everything in `trigger` plus `runId` and `results` |

Storage credentials, signed URLs, payloads, params, `user.storage`, the project
scope and lifecycle bookkeeping are never projected.

How contracts reference the root:

| Contract | Path form | Example | Resolves to |
| --- | --- | --- | --- |
| Start | absolute | `device.deviceName` | `device.deviceName` |
| Start | absolute, inside an open namespace | `inputs.classify.details.*.classified` | same |
| Stage (`stage: anpr`) | relative to `results.<stage>` | `detections.*.plate` | `results.anpr.detections.*.plate` |

Envelope paths and `inputs.classify` are fixed by this module. Automatic runs
carry the classifier result as analysis hands it off (also stored as
`data.classify`); `WorkflowClassifyInput` types it and the schema's classify
paths are generated from its JSON tags, skipping fields tagged
`condition:"-"` (coordinate and colour matrices, the unused `w`). Manual
launches seed `inputs.classify` from the stored analysis through `Classify`,
which has no `objectCount`; such paths are marked `automaticOnly`. Tests decode
a fixture shaped like the classifier's output strictly into
`WorkflowClassifyInput` and resolve every classify path against both the
automatic and the manual shape, so a classifier change fails here first.

`results.anpr` is the ANPR stage result the engine builds from hub-anpr's
ingest blocks (one detection run, one marker per vehicle, grouped by block type
with an `s` appended). `WorkflowAnprResult` mirrors the wire types it is built
from (`api.PostDetectionsRequest`, `api.DetectionTrackInput`, `Marker`,
`MarkerMetadata`); tests require identical JSON keys, decode a result built from
those types strictly, and resolve every `results.anpr` schema path. Constants
(track confidence 0.9, `recognisedBy`), coordinates and bookkeeping are tagged
`condition:"-"`. Plate and OCR confidence live in different lists (tracks and
markers), so conditions on both cannot be tied to the same vehicle; plate,
unread and class can, through `anyMatch` on the tracks. Condition validation
keeps `results.*` open at runtime, so conditions stored before this type existed
still validate.

The contents of other `inputs.<operation>` and all other `results.<operation>`
namespaces come from workers, so contracts describe them and consumers validate
those declarations against real worker output rather than against this schema.

Condition validation (`ValidateWorkflowCondition`) derives its device, user and
`inputs.classify` field lists from the same schema, so condition validation and
contract validation cannot disagree.

`AutomaticTriggerRootWithInputs` builds the same device/user envelope as trigger
mode, without the identity scalars. Round-trip tests require the projection and
the schema to match exactly in both modes, require every `WorkflowDevice` and
`WorkflowUser` field to be either projected or explicitly excluded, and assert
that credentials never appear. A new field therefore becomes matchable only by
adding it to the projection and the schema together.

### Automatic Type Generation

This project bridges Go and TypeScript using an automated pipeline:

1. **Go Models** - Define structs in `pkg/models/` with proper tags
2. **Swagger Generation** - `swag` scans code and generates API spec
3. **OpenAPI Conversion** - Swagger 2.0 → OpenAPI 3.x format
4. **TypeScript Generation** - OpenAPI spec → TypeScript types
5. **Post-Processing** - Add convenient exports and build

This ensures type safety across your entire stack with a single source of truth.

### Model Categories

The models package organizes types into logical categories:

- **Media Models** - Video, thumbnails, sprites, and media metadata
- **Device Models** - Cameras, sensors, and IoT device management
- **Authentication** - Users, tokens, permissions, and access control
- **Analytics** - Metrics, timelines, and data analysis
- **Integration** - Third-party service connections and webhooks
- **Infrastructure** - Health checks, configurations, and system status

## Usage Examples

### Generating TypeScript Types

The primary workflow for maintaining cross-language type safety:

```bash
# Generate both OpenAPI spec and TypeScript types
npm run generate

# Or run steps individually:
npm run generate:openapi  # Go models → OpenAPI 3.x YAML
npm run generate:types    # OpenAPI YAML → TypeScript types
```

**Generated Files:**
- `docs/swagger.yaml` - Swagger 2.0 specification (intermediate)
- `docs/openapi.yaml` - OpenAPI 3.x specification  
- `src/typescript/types.ts` - TypeScript type definitions

### Adding New Models

1. **Create Go struct** in `pkg/models/`:

```go
package models

// User represents a system user
type User struct {
    ID        string `json:"id" bson:"_id,omitempty"`
    Email     string `json:"email" bson:"email" validate:"required,email"`
    Name      string `json:"name" bson:"name" validate:"required"`
    CreatedAt int64  `json:"createdAt" bson:"createdAt"`
}
```

2. **Reference in `cmd/main.go`**:

```go
// Just declare a variable or add to API endpoint
var _ = models.User{}
```

3. **Generate TypeScript types**:

```bash
npm run generate
```

**No manual script updates needed!** The generation pipeline automatically discovers all referenced models.

### Media Management Example

```go
package main

import (
    "context"
    "time"
    "github.com/uug-ai/models/pkg/models"
    "go.mongodb.org/mongo-driver/mongo"
)

func saveMedia(collection *mongo.Collection, deviceId string) error {
    media := models.Media{
        FileName:         "recording_2024.mp4",
        StartTimestamp:   time.Now().Add(-1 * time.Hour).Unix(),
        EndTimestamp:     time.Now().Unix(),
        Duration:         3600000, // 1 hour in milliseconds
        Provider:         "aws-s3",
        Storage:          "media-bucket",
        DeviceId:         deviceId,
        VideoUrl:         "https://cdn.example.com/videos/recording_2024.mp4",
        ThumbnailUrl:     "https://cdn.example.com/thumbs/recording_2024.jpg",
        ThumbnailFile:    "recording_2024_thumb.jpg",
        SpriteUrl:        "https://cdn.example.com/sprites/recording_2024.jpg",
        SpriteFile:       "recording_2024_sprite.jpg",
        SpriteInterval:   10,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err := collection.InsertOne(ctx, media)
    return err
}
```

### Device Registration Example

```go
package main

import (
    "github.com/uug-ai/models/pkg/models"
    "time"
)

func registerDevice(name, deviceId string, lat, lng float64) models.Device {
    return models.Device{
        Name:        name,
        DeviceId:    deviceId,
        Type:        "camera",
        Status:      "online",
        GroupId:     "group-001",
        CreatedAt:   time.Now().Unix(),
        UpdatedAt:   time.Now().Unix(),
        Coordinates: &models.Coordinates{
            Latitude:  lat,
            Longitude: lng,
        },
        Capabilities: []string{"video", "audio", "motion-detection"},
    }
}
```

### API Response Handling

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/uug-ai/models/pkg/api"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    response := api.HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now().Unix(),
        Version:   "1.0.0",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func errorHandler(w http.ResponseWriter, message string, code int) {
    response := api.ErrorResponse{
        Error:      true,
        Message:    message,
        StatusCode: code,
        Timestamp:  time.Now().Unix(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(response)
}
```

## Consolidated workflow run stages

Cross-workflow history uses `api.WorkflowRunOverview`, not the worker-envelope
`WorkflowRun`. Organisation/project-scoped operational rows survive missing or
restricted sources. `sourceAccess` is `available`, `restricted`, or `unavailable`;
`sourceType` is `media` or `case`. Source identifiers, labels, and recording
metadata are optional and require source authorization. Inputs, results, routing
configuration and dispatch errors are never part of this list contract.
New manual writers persist `WorkflowRun.SourceType` to distinguish case IDs from
media launch-group references; legacy case readers still resolve `SourceRef`.

`WorkflowRun.Stages` contains `WorkflowRunStage` entries: immutable routing
(`operation`, `name`, `dispatch`, `queue`, `needs`, `needsMode`) alongside nested
`execution` facts (attempts, timestamps, last dispatch error). Run stages are
distinct from reusable `WorkflowStage` catalog/deployment definitions. Routing
never contains deployment environment, credentials, worker images or resources.

```json
{
  "stages": [{
    "operation": "anpr",
    "dispatch": "conditional",
    "queue": "hub-anpr-queue",
    "needsMode": "all",
    "needs": [{ "operation": "vlm" }],
    "execution": {
      "dispatchAttempts": 1,
      "dispatchedAtMs": 1790681732386
    }
  }]
}
```

Conditions remain on `needs`, including the shared plural conditions described
below. Snapshot all planned stages at creation, including stages that may never
dispatch. Only `execution` changes afterwards; its `state` and `durationMs` are
derived read fields and never persist. The dispatched/resolved operation sets
continue to support existing idempotency and finalization.

### Models-first adoption and legacy compatibility

This is a **Go source change**: `WorkflowRun.Stages` is now
`[]WorkflowRunStage`, not `[]WorkflowStage`. Existing deployments do not change
until they adopt this models release and update their readers/writers.

- `NewWorkflowRunStages(compiled)` makes a detached routing-only copy and starts
  each stage with `execution: {}`. The caller must validate the graph, compile it,
  normalize dispatch defaults and resolve deployment queues first.
- `run.RoutingStages()` produces a detached `[]WorkflowStage` projection for
  registry compilation and existing worker messages, without execution facts.
- `run.NormalizeStages()` explicitly joins legacy `stageExecutions` to `stages`
  by operation. Existing non-empty stage names and entire non-null nested
  executions win, including zero counters and cleared error strings. Matched
  legacy records are removed. Empty/duplicate operations return an error before
  any mutation.
- Legacy summaries without a matching routing stage remain in the deprecated
  field. In particular, old config runs may contain a timeline but no stage
  snapshot. We retain that history and do not invent rules from dependency names
  or today's workflow definition.
- `run.StageExecutionSummaries()` projects the old timeline shape for API
  clients, deriving dependency names from actual needs and retaining unmatched
  legacy history. It does not mutate the run.
- Normalize before `PopulateRuntimeFields` when reading split documents. The
  latter derives both nested and remaining legacy lifecycle states and timings.

Nil/absent/null stages mean no captured routing; an explicit `[]` means a
captured zero-stage plan. JSON and BSON preserve that distinction. Serialization
does **not** implicitly consolidate legacy lifecycle fields or change write paths.
No custom BSON decoder is used, so enclosing ownership-aware document wrappers
continue to decode their own fields.

Adopt readers and compatibility projections in the API/engine first, then
coordinate the writer switch across all engine replicas: legacy writers update
`stageexecutions`, whereas upgraded writers must update `stages.$.execution`.
Do not mix those authorities for the same run. Preserve old API/worker contracts
through projections while their consumers are upgraded. These service changes
are intentionally outside this models-only PR. No data backfill is required.

## Workflow activation and routing conditions

Triggers, graph edges, and compiled stage dependencies share `conditions` and
`conditionMode` (`all` by default, or `any`). An empty top-level group adds no
restriction, including in `any` mode. `WorkflowCondition` supports the existing
`path`/`op`/`value` comparisons and `anyMatch` for same-element matching:

```json
{
  "conditionMode": "all",
  "conditions": [
    {
      "path": "inputs.classify.details",
      "op": "anyMatch",
      "match": {
        "conditionMode": "all",
        "conditions": [
          {"path": "classified", "op": "eq", "value": "pedestrian"},
          {"path": "moving", "op": "eq", "value": true}
        ]
      }
    }
  ]
}
```

The fields and values above are illustrative: use the deployed classifier's
catalog/schema. Inside `match`, paths are relative to **one array element**;
predicates cannot contain another `anyMatch`. At least one element must match,
and missing, non-array, or empty values do not match. Legacy wildcard predicates
retain independent-candidate behavior; two `details.*` predicates can match two
different detections.

- **Triggers:** `type` defaults to `automatic`. `devices`, `siteIds`, and `groupIds`
  select recording scope: OR within each category, AND across populated
  categories. The scope, weekly active hours, and condition group must all hold.
  Schedules filter recording events; they are not scheduled jobs. Site/group IDs
  are stable strings, and callers must resolve effective memberships and enforce
  tenant/project access separately.
- **Manual triggers:** `surfaces` advertises `case`, `media`, and/or `redaction`.
  Validation requires a supported surface; automatic-only fields remain ignored
  for compatibility with existing documents.
- **Start roots:** new editor graphs use
  `{"id":"recording","type":"start","trigger":{"type":"automatic"}}`.
  Set `trigger.type` to `manual` and select `trigger.surfaces` for user-launched
  workflows. `NormalizeTriggers`/`SyncGraphTriggers` derive one automatic trigger
  per outgoing Start edge (or one manual trigger), overriding stale top-level
  triggers. Each outgoing edge owns an optional `trigger` object containing only
  `devices`, `siteIds`, `groupIds`, `classifications`, and `weeklySchedule`; its `conditions` and
  `conditionMode` remain on the edge, outside that object. A present `trigger`
  (including `{}`) completely replaces legacy node source/schedule scope; an
  absent/null trigger reads node `devices` and node-trigger site/group/schedule
  fields without migration. No field-by-field merging occurs. Legacy node
  conditions remain an independent AND gate through derived
  `sharedConditions`, preserving both groups' `all`/`any` modes. Inactive settings remain
  on the node/edges, but automatic runtime triggers omit manual surfaces and manual
  runtime triggers omit automatic settings. Edge triggers are invalid on ordinary
  stage and legacy device edges. Authored automatic edge schedules reject null
  entries, weekdays outside 0–6, unknown non-empty timezones, and segments outside
  `0 <= start < end <= 86400`, including on drafts. Manual edge schedules remain
  dormant; legacy schedule readers retain their compatibility behavior.
  Selected `classifications` require any selected label at
  `inputs.classify.details.*.classified` in the initial handoff. This filter is
  ANDed with source scope, schedule, and the outer predicate group—even with
  `conditionMode: "any"`. Empty means unrestricted; missing initial classifier
  input cannot satisfy a non-empty selection, and future results never count.
  There is no node-level classification fallback; existing predicates are unchanged.
  A graph may have only one Start or
  legacy device root, with no incoming edges. An explicit Start graph supersedes
  cached `stages`; disconnected nodes never start automatically. Automatic Start
  edges require stable unique IDs and compile to dependencies carrying
  `startEdgeId`. After `MatchAutomaticTrigger`, `BindStartEdges` freezes their
  `startMatched` decisions; later results cannot open an unmatched entry.
  Entry alternatives combine with OR, required alongside ordinary dependency
  joins. An empty edge condition group is unconditional within its effective scope and
  schedule; no outgoing runnable edges means no automatic activation.
  Manual Start edges retain their existing classifier-gated routing behavior.
  Missing/unknown Start modes and unknown
  surfaces are invalid; disabled manual drafts may have no selected surfaces.
  Legacy device roots still preserve their manual triggers; stage-only/config
  workflows remain unchanged. No data migration is required.
- **Manual launch surfaces:** `POST /tasks/{taskId}/workflows` accepts an optional
  `surface` of `case` (default) or `redaction`; the selected workflow must be
  enabled and expose that exact manual surface. Media launches use the existing
  `POST /media/workflows` endpoint and require the `media` surface. Surface
  selection does not grant access to another tenant, project, case, or recording.
- **Edges:** `source`/`target` are node IDs. Compilation derives the source
  operation as a dependency readiness gate. Predicate paths are absolute and
  independent of that source; port names do not rebase them. A stage dependency
  without predicates still waits for its operation to be available. Legacy
  device edges with predicates gate on `classify`; those without predicates
  start their target immediately.
- **Two distinct modes:** a dependency's `conditionMode` combines its predicates.
  A stage's `needsMode` combines incoming dependencies and still defaults to
  `any`.

Use `WorkflowTrigger.Validate`, `Workflow.ValidateTriggers`,
`Workflow.ValidateGraph`, and `StageDependency.ValidateConditions` before
accepting configurations. Use `EvaluateConditionSet` for shared evaluation and
`StageDependency.Matches` with the available operations from run inputs/results
for readiness plus predicates. `WorkflowTrigger.Matches` includes selectors and
schedule; `Workflow.AutomaticMatches` provides one boolean activation decision
across alternative triggers. `CompiledConditions` remains a deprecated flat
inspection helper and must not be used to evaluate `any` groups.

### Compatibility and rollout

`StageCondition` remains source-compatible with `WorkflowCondition`. Singular
edge/dependency `condition` remains readable without a migration: `ConditionSet`
normalizes it in memory to a one-item `all` group. Supplying both singular and
plural forms (including an explicit empty plural list) is rejected by validation.
New writers should emit only plural conditions. Existing JSON/BSON field names,
default modes, and workflow ownership normalization remain unchanged.

This is a **models-first, reader-first** change, not an enabled editor feature:

1. Adopt shared validation/evaluation in `hub-workflows` and `hub-api` before
   accepting new-format writes. Older engines ignore plural edge conditions and
   could dispatch unconditionally.
2. Supply sanitized, already-available operation inputs through
   `AutomaticTriggerRootWithInputs`; the original `AutomaticTriggerRoot` retains
   its input-free behavior. Trigger predicates cannot reference future `results`.
   Callers must also populate trusted `WorkflowDevice.GroupIds`; models do not
   resolve site/group relationships or scrub arbitrary worker payloads.
3. Author Start-edge source/schedule in `WorkflowEdge.Trigger` and predicates in
   the outer edge fields using the shared condition catalog. `Workflow.Triggers`
   is the derived runtime projection, not an editor write surface. Legacy device
   graphs retain their existing `SyncGraphTriggers` camera-only adapter.
4. Enable the new UI/writes only after all readers are updated. No data backfill
   is required for existing workflows.

## Project Structure

```
.
├── pkg/
│   ├── models/              # Core data models
│   │   ├── media.go        # Media and video models
│   │   ├── device.go       # Device management models
│   │   ├── user.go         # User and authentication models
│   │   ├── authentication.go # Auth tokens and sessions
│   │   ├── analysis.go     # Analytics and metrics
│   │   ├── pipeline.go     # Processing pipelines
│   │   ├── integration.go  # Third-party integrations
│   │   └── ...             # 30+ additional models
│   └── api/                 # API response structures
│       ├── api.go          # Common API types
│       ├── media.go        # Media API responses
│       ├── device.go       # Device API responses
│       └── ...
├── cmd/
│   └── main.go             # Model auto-discovery entry point
├── scripts/
│   ├── auto-discover-models.js # Automatic model detection
│   └── add-type-exports.js     # TypeScript export generation
├── src/
│   └── typescript/
│       ├── types.ts        # Generated TypeScript types
│       └── package.json
├── docs/
│   ├── swagger.yaml        # Swagger 2.0 spec
│   └── openapi.yaml        # OpenAPI 3.x spec
├── go.mod
├── package.json
├── BEST_PRACTICES.md       # Type definition guidelines
└── README.md
```

## Configuration

### Development Setup

1. **Clone the repository**:
```bash
git clone https://github.com/uug-ai/models.git
cd models
```

2. **Install Go dependencies**:
```bash
go mod download
```

3. **Install Node.js dependencies**:
```bash
npm install
```

4. **Install Swagger CLI**:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Environment Variables

TypeScript generation can be configured via environment:

```bash
# Skip OpenAPI generation (use existing spec)
SKIP_OPENAPI=true npm run generate

# Verbose output
DEBUG=true npm run generate
```

## Available Models

### Core Models (`pkg/models/`)

- `Media` - Video files, thumbnails, sprites, and metadata
- `Device` - Cameras, sensors, and IoT devices
- `User` - User accounts and profiles
- `Authentication` - Auth tokens, sessions, and credentials
- `AccessToken` - API access tokens and permissions
- `Group` - User groups and organizations
- `Site` - Physical locations and sites
- `Analysis` - Analytics data and metrics
- `Pipeline` - Data processing pipelines
- `Strategy` - Processing strategies and configurations
- `Marker` - Timeline markers and annotations
- `Activity` - User activity logs
- `Alert` - System alerts and notifications
- `Audit` - Audit logs and compliance
- `Comment` - User comments and feedback
- `Config` - System configurations
- `FloorPlan` - Site floor plans and layouts
- `Health` - Health check data
- `Integration` - Third-party integrations
- `Location` - Geographic locations
- `Message` - System messages and notifications
- `Object` - Object detection results
- `Permission` - Access permissions
- `Provider` - Service providers
- `Role` - User roles and capabilities
- `Subscription` - Subscription and billing
- `Vault` - Secure credential storage

### API Models (`pkg/api/`)

- `ErrorResponse` - Standard error responses

### Selective Fields

Clients can request optional fields on certain API payloads using the `include` parameter in request models. For example, site option queries support `include: ["metadata"]` to return site metadata alongside the default fields.
- `HealthResponse` - Health check responses
- `PaginationResponse` - Paginated list responses
- `MediaResponse` - Media query responses
- `DeviceResponse` - Device query responses
- `AnalysisResponse` - Analytics responses

## Validation

Models use [go-playground/validator](https://github.com/go-playground/validator) for field validation:

```go
type User struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Age      int    `json:"age" validate:"gte=0,lte=150"`
    Role     string `json:"role" validate:"required,oneof=admin user guest"`
}

// Validate in your code
import "github.com/go-playground/validator/v10"

validate := validator.New()
err := validate.Struct(user)
if err != nil {
    // Handle validation errors
}
```

Common validation tags:
- `required` - Field must be present
- `email` - Valid email format
- `min=n` - Minimum length/value
- `max=n` - Maximum length/value
- `oneof=a b c` - Must be one of specified values
- `gte=n,lte=n` - Range validation

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Run tests for specific packages:

```bash
# Model tests
go test ./pkg/models -v

# Pipeline tests
go test ./pkg/models -run TestPipeline
```

## Contributing

Contributions are welcome! Please follow the guidelines in [BEST_PRACTICES.md](./BEST_PRACTICES.md) when adding new models.

### Development Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-model`)
3. Add your model to `pkg/models/` or `pkg/api/`
4. Include proper JSON, BSON, and validation tags
5. Reference the model in `cmd/main.go`
6. Generate TypeScript types: `npm run generate`
7. Add tests for your model
8. Ensure all tests pass: `go test ./...`
9. Commit following [Conventional Commits](https://www.conventionalcommits.org/)
10. Push to your branch (`git push origin feat/amazing-model`)
11. Open a Pull Request

### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `types`

**Scopes:**
- `models` - Changes to Go model structs in `pkg/models/`
- `api` - Changes to API response structures in `pkg/api/`
- `types` - TypeScript type generation or definitions
- `scripts` - Build/generation scripts
- `docs` - Documentation updates

**Examples:**

```
feat(models): add sprite interval field to Media struct
fix(api): correct JSON tags for ErrorResponse message field
docs(readme): update TypeScript generation workflow
refactor(models): standardize BSON tag formatting across all structs
types: regenerate TypeScript definitions from updated Go models
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

This project uses the following key libraries:

- [swaggo/swag](https://github.com/swaggo/swag) - Swagger/OpenAPI generation from Go
- [mongo-driver](https://github.com/mongodb/mongo-go-driver) - Official MongoDB Go driver
- [go-playground/validator](https://github.com/go-playground/validator) - Struct validation
- [openapi-typescript](https://github.com/drwpow/openapi-typescript) - TypeScript generation from OpenAPI
- [swagger2openapi](https://github.com/Mermade/oas-kit) - Swagger to OpenAPI conversion

See [go.mod](go.mod) and [package.json](package.json) for complete dependency lists.

## Support

- **Issues**: [GitHub Issues](https://github.com/uug-ai/models/issues)
- **Discussions**: [GitHub Discussions](https://github.com/uug-ai/models/discussions)
- **Documentation**: See [BEST_PRACTICES.md](./BEST_PRACTICES.md) and inline code comments