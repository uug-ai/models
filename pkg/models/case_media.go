package models

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CaseMedia is a case-scoped media entry stored in the dedicated
// `case_media` MongoDB collection. Each entry is either a snapshot of a
// source media attached to the case (Role = "source") or a derivative
// produced by an edit pipeline (Role = "edit"). Edits are versioned per
// (TaskId, ParentId, Action, EditType) and may supersede a previous entry
// via SupersedesId so we keep a full history of edits within the case.
//
// The collection is intentionally separate from the `tasks` collection so
// that worker-driven status updates do not require positional array
// updates on the task document, and so the per-case media inventory can
// grow independently from the task list view.
type CaseMedia struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TaskId         primitive.ObjectID `json:"taskId" bson:"task_id"`
	OrganisationId string             `json:"organisationId" bson:"organisation_id"`

	// Role describes whether this entry represents a source attached to
	// the case or an edit derived from another entry. See CaseMediaRole.
	Role CaseMediaRole `json:"role" bson:"role"`

	// ParentId links an edit (Role = "edit") to the case_media entry it
	// was produced from. For composite edits the parent is still a single
	// source — the individual operations are described in Params.
	ParentId *primitive.ObjectID `json:"parentId,omitempty" bson:"parent_id,omitempty"`

	// Action describes what the edit does. Only meaningful when
	// Role = "edit". See CaseMediaAction.
	Action CaseMediaAction `json:"action,omitempty" bson:"action,omitempty"`

	// EditType is a sub-variant of Action (for example "face_blur",
	// "face_pixelate", "face_mask" when Action = "redaction"). Display
	// labels are derived from this in the UI. Canonical values are
	// declared as CaseMediaEditType constants below.
	EditType CaseMediaEditType `json:"editType,omitempty" bson:"edit_type,omitempty"`

	// Version is monotonically increasing within
	// (TaskId, ParentId, Action, EditType). The most recent completed
	// version is the one the export pipeline picks up.
	Version int `json:"version,omitempty" bson:"version,omitempty"`

	// SupersedesId points at the previous version this entry replaces,
	// preserving the full edit chain for auditing / rollback.
	SupersedesId *primitive.ObjectID `json:"supersedesId,omitempty" bson:"supersedes_id,omitempty"`

	// Params carries kind-specific parameters used to produce the edit.
	// For Action = "composite" it is expected to contain an
	// "operations" array whose elements have an "op" discriminator
	// matching one of the single-op CaseMediaAction values.
	Params map[string]interface{} `json:"params,omitempty" bson:"params,omitempty"`

	// Source snapshot fields (populated on Role = "source"; mirrored from
	// Media at attach time so the case stays self-contained even after
	// the original media document is cleaned up).
	SourceMediaId     *primitive.ObjectID `json:"sourceMediaId,omitempty" bson:"source_media_id,omitempty"`
	Timestamp         int64               `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	EndTimestamp      int64               `json:"endTimestamp,omitempty" bson:"end_timestamp,omitempty"`
	CameraId          string              `json:"cameraId,omitempty" bson:"camera_id,omitempty"`
	StorageSolution   string              `json:"storageSolution,omitempty" bson:"storage_solution,omitempty"`
	VideoFile         string              `json:"videoFile,omitempty" bson:"video_file,omitempty"`
	VideoProvider     string              `json:"videoProvider,omitempty" bson:"video_provider,omitempty"`
	ThumbnailFile     string              `json:"thumbnailFile,omitempty" bson:"thumbnail_file,omitempty"`
	ThumbnailProvider string              `json:"thumbnailProvider,omitempty" bson:"thumbnail_provider,omitempty"`
	SpriteFile        string              `json:"spriteFile,omitempty" bson:"sprite_file,omitempty"`
	SpriteProvider    string              `json:"spriteProvider,omitempty" bson:"sprite_provider,omitempty"`
	SpriteInterval    int                 `json:"spriteInterval,omitempty" bson:"sprite_interval,omitempty"`

	// Media is a full snapshot of the source Media document captured at
	// attach time. It is only populated on Role = "source" and is what
	// the media-detail page and edit modal consume — the flat fields
	// above stay in place because they are queried directly by the
	// export pipeline and the case-level summarisers. The snapshot
	// makes the case self-contained: edits to or deletions of the
	// original media row do not affect what the case shows.
	Media *Media `json:"media,omitempty" bson:"media,omitempty"`

	// Analysis is a snapshot of the AnalysisWrapper associated with
	// the source media at attach time, including the per-operation
	// lifecycle fields (AsyncOperations / RequiredOperations /
	// ResolvedOperations) plus the analysis data (classify, counting,
	// faceRedaction, …) so the detail page and the face-redaction
	// edit modal can render without an extra lookup against the
	// `analysis` collection. Only populated on Role = "source".
	Analysis *AnalysisWrapper `json:"analysis,omitempty" bson:"analysis,omitempty"`

	// Produced artefact (populated on Role = "edit" once the worker
	// completes; for Role = "source" these mirror VideoFile/VideoProvider).
	File     string `json:"file,omitempty" bson:"file,omitempty"`
	Provider string `json:"provider,omitempty" bson:"provider,omitempty"`

	// Lifecycle (only meaningful on Role = "edit"). For sources this is
	// implicitly Completed at creation time.
	Status      CaseMediaStatus `json:"status,omitempty" bson:"status,omitempty"`
	StatusError string          `json:"statusError,omitempty" bson:"status_error,omitempty"`

	// Optional job linkage — back-pointers to the analysis / face
	// redaction documents that drove the edit, when applicable.
	AnalysisId      *primitive.ObjectID `json:"analysisId,omitempty" bson:"analysis_id,omitempty"`
	FaceRedactionId *primitive.ObjectID `json:"faceRedactionId,omitempty" bson:"face_redaction_id,omitempty"`

	CreatedAt int64  `json:"createdAt,omitempty" bson:"created_at,omitempty"`
	CreatedBy string `json:"createdBy,omitempty" bson:"created_by,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updated_at,omitempty"`

	// Url is signed by the API at fetch time and not persisted.
	Url string `json:"url,omitempty" bson:"-"`
	// ThumbnailUrl is a signed playback URL for ThumbnailFile and is
	// populated by the API at fetch time. Not persisted.
	ThumbnailUrl string `json:"thumbnailUrl,omitempty" bson:"-"`
	// SpriteUrl is a signed playback URL for SpriteFile and is populated
	// by the API at fetch time. Not persisted.
	SpriteUrl string `json:"spriteUrl,omitempty" bson:"-"`
}

// CaseMediaRole distinguishes source attachments from edit derivatives
// inside the case_media collection.
type CaseMediaRole string

const (
	CaseMediaRoleSource CaseMediaRole = "source"
	CaseMediaRoleEdit   CaseMediaRole = "edit"
)

// CaseMediaAction enumerates the supported edit actions. Only meaningful
// when CaseMedia.Role == "edit". The "composite" action is a sequence of
// single-action operations described in CaseMedia.Params["operations"].
type CaseMediaAction string

const (
	CaseMediaActionRedaction CaseMediaAction = "redaction"
	CaseMediaActionTrim      CaseMediaAction = "trim"
	CaseMediaActionComposite CaseMediaAction = "composite"
)

// CaseMediaStatus mirrors the worker-driven lifecycle of an edit entry.
//
//   - "queued":       hub-api accepted the request and enqueued the job.
//   - "processing":   a worker has picked the job up and is rendering.
//   - "completed":    the artefact is available at File/Provider.
//   - "failed":       the worker aborted; see StatusError.
type CaseMediaStatus string

const (
	CaseMediaStatusQueued     CaseMediaStatus = "queued"
	CaseMediaStatusProcessing CaseMediaStatus = "processing"
	CaseMediaStatusCompleted  CaseMediaStatus = "completed"
	CaseMediaStatusFailed     CaseMediaStatus = "failed"
)

// CaseMediaEditType enumerates the supported sub-variants of an edit
// action. Values are grouped by their parent Action: redaction variants
// (face_blur / face_pixelate / face_mask) are validated by hub-api when
// Action = redaction; trim / composite do not currently use sub-variants.
type CaseMediaEditType string

const (
	CaseMediaEditTypeFaceBlur     CaseMediaEditType = "face_blur"
	CaseMediaEditTypeFacePixelate CaseMediaEditType = "face_pixelate"
	CaseMediaEditTypeFaceMask     CaseMediaEditType = "face_mask"
)

// --- Contract notes ---------------------------------------------------------
//
// The flat snapshot fields on a Role=source row (VideoFile, VideoProvider,
// ThumbnailFile, ThumbnailProvider, SpriteFile, SpriteProvider,
// SpriteInterval, CameraId, Timestamp, EndTimestamp, StorageSolution)
// are the *contract surface*. They are queried directly by the export
// pipeline and by case-level summarisers without unmarshalling the
// embedded Media / Analysis blobs.
//
// The embedded Media / Analysis fields are the *display payload* —
// consumed only by the media-detail page and the face-redaction edit
// modal. They are written exactly once (at attach time) and MUST NOT be
// mutated piecemeal: replace the whole row if an upstream change has to
// be propagated.
//
// Validate is intentionally cheap (no I/O); callers are expected to run
// it before persisting.
func (cm *CaseMedia) Validate() error {
	if cm == nil {
		return fmt.Errorf("case_media: nil")
	}

	switch cm.Role {
	case CaseMediaRoleSource:
		if cm.SourceMediaId == nil || cm.SourceMediaId.IsZero() {
			return fmt.Errorf("case_media: source row requires sourceMediaId")
		}
		if cm.VideoFile == "" {
			return fmt.Errorf("case_media: source row requires videoFile")
		}
		if cm.ParentId != nil {
			return fmt.Errorf("case_media: source row must not have parentId")
		}
		if cm.Action != "" || cm.EditType != "" {
			return fmt.Errorf("case_media: source row must not set action/editType")
		}
		if cm.Status == "" {
			// Sources are inherently completed; reject implicit empty
			// state so list/filter callers can rely on a populated
			// status field.
			return fmt.Errorf("case_media: source row must set status (typically Completed)")
		}
	case CaseMediaRoleEdit:
		if cm.ParentId == nil || cm.ParentId.IsZero() {
			return fmt.Errorf("case_media: edit row requires parentId")
		}
		if cm.Action == "" {
			return fmt.Errorf("case_media: edit row requires action")
		}
		switch cm.Action {
		case CaseMediaActionRedaction:
			switch cm.EditType {
			case CaseMediaEditTypeFaceBlur, CaseMediaEditTypeFacePixelate, CaseMediaEditTypeFaceMask:
			default:
				return fmt.Errorf("case_media: redaction edit requires a face_* editType (got %q)", cm.EditType)
			}
		case CaseMediaActionTrim, CaseMediaActionComposite:
			if cm.EditType != "" {
				return fmt.Errorf("case_media: action %q does not accept editType (got %q)", cm.Action, cm.EditType)
			}
		default:
			return fmt.Errorf("case_media: unknown action %q", cm.Action)
		}
		if cm.Version <= 0 {
			return fmt.Errorf("case_media: edit row requires version > 0")
		}
		if cm.Status == "" {
			return fmt.Errorf("case_media: edit row requires status")
		}
	default:
		return fmt.Errorf("case_media: unknown role %q", cm.Role)
	}
	return nil
}
