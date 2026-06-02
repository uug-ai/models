package api

import (
	"encoding/json"
	"strings"

	"github.com/uug-ai/models/pkg/models"
)

// AnalysisStatus represents specific status codes for analysis operations
type AnalysisStatus string

const (
	AnalysisFaceRedactionBindingFailed AnalysisStatus = "analysis_face_redaction_binding_failed"
	AnalysisSaveRedactionSuccess       AnalysisStatus = "analysis_save_redaction_success"
	AnalysisSaveRedactionFailed        AnalysisStatus = "analysis_save_redaction_failed"
	AnalysisSubmitRedactionSuccess     AnalysisStatus = "analysis_submit_redaction_success"
	AnalysisSubmitRedactionFailed      AnalysisStatus = "analysis_submit_redaction_failed"

	AnalysisDetectionsBindingFailed     AnalysisStatus = "analysis_detections_binding_failed"
	AnalysisDetectionsStored            AnalysisStatus = "analysis_detections_stored"
	AnalysisDetectionsPartial           AnalysisStatus = "analysis_detections_partial"
	AnalysisDetectionsAllInvalid        AnalysisStatus = "analysis_detections_all_invalid"
	AnalysisDetectionsFailed            AnalysisStatus = "analysis_detections_failed"
	AnalysisDetectionsUnsupportedSchema AnalysisStatus = "analysis_detections_unsupported_schema_version"
	AnalysisDetectionsTooLarge          AnalysisStatus = "analysis_detections_too_large"
	AnalysisDetectionsTargetMissing     AnalysisStatus = "analysis_detections_target_missing"
	AnalysisDetectionsFound             AnalysisStatus = "analysis_detections_found"
	AnalysisDetectionsNotFound          AnalysisStatus = "analysis_detections_not_found"
	AnalysisDetectionsDeleted           AnalysisStatus = "analysis_detections_deleted"
	AnalysisDetectionRunIdMissing       AnalysisStatus = "analysis_detection_run_id_missing"

	AnalysisFileNameMissing            AnalysisStatus = "analysis_file_name_missing"
	AnalysisSignedUrlMissing           AnalysisStatus = "analysis_signed_url_missing"
	AnalysisAllFrameCoordinatesMissing AnalysisStatus = "analysis_all_frame_coordinates_missing"

	AnalysisNotFound            AnalysisStatus = "analysis_not_found"
	AnalysisFound               AnalysisStatus = "analysis_found"
	AnalysisIdMissing           AnalysisStatus = "analysisId_missing"
	AnalysisStarted             AnalysisStatus = "analysis_started"
	AnalysisQueueSubscribed     AnalysisStatus = "analysis_queue_subscribed"
	AnalysisStageMonitorMissing AnalysisStatus = "analysis_stage_monitor_missing"
	AnalysisCompleted           AnalysisStatus = "analysis_completed"

	AnalysisDecodeFailed             AnalysisStatus = "analysis_decode_failed"
	AnalysisInsertFailed             AnalysisStatus = "analysis_insert_failed"
	AnalysisUpdateFailed             AnalysisStatus = "analysis_update_failed"
	AnalysisNotificationUpdateFailed AnalysisStatus = "analysis_notification_update_failed"
	AnalysisSequenceUpdateFailed     AnalysisStatus = "analysis_sequence_update_failed"
	AnalysisTaskUpdateFailed         AnalysisStatus = "analysis_task_update_failed"
)

// String returns the string representation of the analysis status
func (as AnalysisStatus) String() string {
	return string(as)
}

// Translate returns the translated string representation of the analysis status in the specified language
func (as AnalysisStatus) Translate(lang string) string {
	translations := map[string]map[AnalysisStatus]string{
		"en": {
			AnalysisFaceRedactionBindingFailed:  "Face redaction binding failed",
			AnalysisSaveRedactionSuccess:        "Face redaction saved successfully",
			AnalysisSaveRedactionFailed:         "Failed to save face redaction",
			AnalysisSubmitRedactionSuccess:      "Face redaction submitted successfully",
			AnalysisSubmitRedactionFailed:       "Failed to submit face redaction",
			AnalysisDetectionsBindingFailed:     "Detections payload binding failed",
			AnalysisDetectionsStored:            "Detections stored successfully",
			AnalysisDetectionsPartial:           "Detections stored with some boxes rejected",
			AnalysisDetectionsAllInvalid:        "Detections rejected: every box is invalid",
			AnalysisDetectionsFailed:            "Failed to store detections",
			AnalysisDetectionsUnsupportedSchema: "Unsupported detections schema version",
			AnalysisDetectionsTooLarge:          "Detections payload too large",
			AnalysisDetectionsTargetMissing:     "A mediaId or analysisId is required",
			AnalysisDetectionsFound:             "Detections found",
			AnalysisDetectionsNotFound:          "Detection run not found",
			AnalysisDetectionsDeleted:           "Detection run deleted",
			AnalysisDetectionRunIdMissing:       "Detection run ID is missing",
			AnalysisSignedUrlMissing:            "Signed URL is missing",
			AnalysisNotFound:                    "Analysis not found",
			AnalysisFound:                       "Analysis found",
			AnalysisIdMissing:                   "Analysis ID is missing",
			AnalysisAllFrameCoordinatesMissing:  "All frame coordinates map are missing",
			AnalysisFileNameMissing:             "File name is missing",
			AnalysisStageMonitorMissing:         "Stage monitor is missing",
			AnalysisStarted:                     "Analysis started",
			AnalysisQueueSubscribed:             "Analysis queue subscribed",
			AnalysisCompleted:                   "Analysis completed",
			AnalysisDecodeFailed:                "Failed to decode analysis",
			AnalysisInsertFailed:                "Failed to insert analysis",
			AnalysisUpdateFailed:                "Failed to update analysis",
			AnalysisNotificationUpdateFailed:    "Failed to update analysis notification settings",
			AnalysisSequenceUpdateFailed:        "Failed to update analysis sequence information",
			AnalysisTaskUpdateFailed:            "Failed to update analysis task information",
		},
		"es": {
			AnalysisFaceRedactionBindingFailed:  "Error al vincular la redacción facial",
			AnalysisSaveRedactionSuccess:        "Redacción facial guardada con éxito",
			AnalysisSaveRedactionFailed:         "Error al guardar la redacción facial",
			AnalysisSubmitRedactionSuccess:      "Redacción facial enviada con éxito",
			AnalysisSubmitRedactionFailed:       "Error al enviar la redacción facial",
			AnalysisDetectionsBindingFailed:     "Error al vincular la carga de detecciones",
			AnalysisDetectionsStored:            "Detecciones almacenadas con éxito",
			AnalysisDetectionsPartial:           "Detecciones almacenadas con algunas cajas rechazadas",
			AnalysisDetectionsAllInvalid:        "Detecciones rechazadas: todas las cajas son inválidas",
			AnalysisDetectionsFailed:            "Error al almacenar las detecciones",
			AnalysisDetectionsUnsupportedSchema: "Versión de esquema de detecciones no compatible",
			AnalysisDetectionsTooLarge:          "Carga de detecciones demasiado grande",
			AnalysisDetectionsTargetMissing:     "Se requiere un mediaId o analysisId",
			AnalysisDetectionsFound:             "Detecciones encontradas",
			AnalysisDetectionsNotFound:          "Ejecución de detección no encontrada",
			AnalysisDetectionsDeleted:           "Ejecución de detección eliminada",
			AnalysisDetectionRunIdMissing:       "Falta el ID de la ejecución de detección",
			AnalysisSignedUrlMissing:            "Falta la URL firmada",
			AnalysisNotFound:                    "Análisis no encontrado",
			AnalysisFound:                       "Análisis encontrado",
			AnalysisIdMissing:                   "Falta el ID del análisis",
			AnalysisFileNameMissing:             "Falta el nombre del archivo",
			AnalysisAllFrameCoordinatesMissing:  "Faltan el mapa de coordenadas de todos los fotogramas",
			AnalysisStageMonitorMissing:         "Falta el monitor de etapa",
			AnalysisStarted:                     "Análisis iniciado",
			AnalysisQueueSubscribed:             "Cola de análisis suscrita",
			AnalysisCompleted:                   "Análisis completado",
			AnalysisDecodeFailed:                "Error al decodificar el análisis",
			AnalysisInsertFailed:                "Error al insertar el análisis",
			AnalysisUpdateFailed:                "Error al actualizar el análisis",
			AnalysisNotificationUpdateFailed:    "Error al actualizar la configuración de notificaciones del análisis",
			AnalysisSequenceUpdateFailed:        "Error al actualizar la información de la secuencia del análisis",
			AnalysisTaskUpdateFailed:            "Error al actualizar la información de la tarea del análisis",
		},
		"fr": {
			AnalysisFaceRedactionBindingFailed:  "Échec de la liaison de la rédaction faciale",
			AnalysisSaveRedactionSuccess:        "Rédaction du visage enregistrée avec succès",
			AnalysisSaveRedactionFailed:         "Échec de l'enregistrement de la rédaction du visage",
			AnalysisSubmitRedactionSuccess:      "Rédaction du visage soumise avec succès",
			AnalysisSubmitRedactionFailed:       "Échec de la soumission de la rédaction du visage",
			AnalysisDetectionsBindingFailed:     "Échec de la liaison de la charge de détections",
			AnalysisDetectionsStored:            "Détections enregistrées avec succès",
			AnalysisDetectionsPartial:           "Détections enregistrées avec certaines boîtes rejetées",
			AnalysisDetectionsAllInvalid:        "Détections rejetées : toutes les boîtes sont invalides",
			AnalysisDetectionsFailed:            "Échec de l'enregistrement des détections",
			AnalysisDetectionsUnsupportedSchema: "Version de schéma de détections non prise en charge",
			AnalysisDetectionsTooLarge:          "Charge de détections trop volumineuse",
			AnalysisDetectionsTargetMissing:     "Un mediaId ou analysisId est requis",
			AnalysisDetectionsFound:             "Détections trouvées",
			AnalysisDetectionsNotFound:          "Exécution de détection introuvable",
			AnalysisDetectionsDeleted:           "Exécution de détection supprimée",
			AnalysisDetectionRunIdMissing:       "ID de l'exécution de détection manquant",
			AnalysisSignedUrlMissing:            "URL signée manquante",
			AnalysisNotFound:                    "Analyse non trouvée",
			AnalysisFound:                       "Analyse trouvée",
			AnalysisIdMissing:                   "ID d'analyse manquant",
			AnalysisFileNameMissing:             "Nom de fichier manquant",
			AnalysisAllFrameCoordinatesMissing:  "Map de toutes les coordonnées des images sont manquantes",
			AnalysisStageMonitorMissing:         "Moniteur de stade manquant",
			AnalysisStarted:                     "Analyse démarrée",
			AnalysisQueueSubscribed:             "File d'attente d'analyse souscrite",
			AnalysisCompleted:                   "Analyse terminée",
			AnalysisDecodeFailed:                "Échec du décodage de l'analyse",
			AnalysisInsertFailed:                "Échec de l'insertion de l'analyse",
			AnalysisUpdateFailed:                "Échec de la mise à jour de l'analyse",
			AnalysisNotificationUpdateFailed:    "Échec de la mise à jour des paramètres de notification de l'analyse",
			AnalysisSequenceUpdateFailed:        "Échec de la mise à jour des informations de séquence d'analyse",
			AnalysisTaskUpdateFailed:            "Échec de la mise à jour des informations de tâche d'analyse",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[as]; exists {
			return translation
		}
	}

	// Default to English if language not found or translation doesn't exist
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[as]; exists {
			return translation
		}
	}

	// Fallback to the string representation
	return as.String()
}

// SaveFaceRedaction
// @Router /analysis/save-face-redaction [patch]
type SaveFaceRedactionRequest struct {
	AnalysisId    string               `json:"analysisId"`
	FaceRedaction models.FaceRedaction `json:"faceRedaction"`
}
type SaveFaceRedactionResponse struct {
	AnalysisId    string               `json:"analysisId"`
	FaceRedaction models.FaceRedaction `json:"faceRedaction"`
}
type SaveFaceRedactionSuccessResponse struct {
	SuccessResponse
	Data SaveFaceRedactionResponse `json:"data"`
}
type SaveFaceRedactionErrorResponse struct {
	ErrorResponse
}

// PostDetections
//
// Wire format for POST /detections. Producers send detection runs (e.g. a
// bring-your-own model) which the server normalises and stores in the dedicated
// "detections" collection, keyed by the recording. Exactly one of MediaId
// (recording/media key) or AnalysisId must identify the target recording. Runs
// are upserted by (recording key, Source.RunId).
// @Router /detections [post]
type PostDetectionsRequest struct {
	// MediaId is the recording/media key the run belongs to. Provide this or
	// AnalysisId (MediaId wins when both are present).
	MediaId string `json:"mediaId,omitempty"`
	// AnalysisId targets the recording via its analysis document id, as an
	// alternative to MediaId.
	AnalysisId string `json:"analysisId,omitempty"`
	// Task is an optional run discriminator; defaults to "detection".
	Task            string                     `json:"task,omitempty"`
	SchemaVersion   string                     `json:"schemaVersion,omitempty"`
	Source          models.DetectionSource     `json:"source"`
	CoordinateSpace string                     `json:"coordinateSpace"` // "pixel" | "normalized"
	Media           models.DetectionMedia      `json:"media,omitempty"`
	Categories      []models.DetectionCategory `json:"categories,omitempty"`
	Tracks          []DetectionTrackInput      `json:"tracks"`
}

// DetectionTrackInput is one track on the wire. It mirrors the editor's track
// shape; boxes carry the raw geometry the server normalises.
type DetectionTrackInput struct {
	Id            FlexibleString         `json:"id"`
	Label         string                 `json:"label,omitempty"`
	ClassId       *int                   `json:"classId,omitempty"`
	Confidence    float64                `json:"confidence,omitempty"`
	Color         string                 `json:"color,omitempty"`
	Shape         string                 `json:"shape,omitempty"`
	DeletedFrames []int64                `json:"deletedFrames,omitempty"`
	Meta          map[string]interface{} `json:"meta,omitempty"`
	Boxes         []DetectionBoxInput    `json:"boxes"`
}

// DetectionBoxInput is one detection of a subject at one frame. It accepts both
// the preferred {x, y, w, h} (top-left + size) and the legacy {x1, y1, x2, y2}
// forms; pointers let the server detect which form was sent.
type DetectionBoxInput struct {
	Frame       int64                  `json:"frame"`
	TimestampMs int64                  `json:"timestampMs,omitempty"`
	X           *float64               `json:"x,omitempty"`
	Y           *float64               `json:"y,omitempty"`
	W           *float64               `json:"w,omitempty"`
	H           *float64               `json:"h,omitempty"`
	X1          *float64               `json:"x1,omitempty"`
	Y1          *float64               `json:"y1,omitempty"`
	X2          *float64               `json:"x2,omitempty"`
	Y2          *float64               `json:"y2,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
	Label       string                 `json:"label,omitempty"`
	ClassId     *int                   `json:"classId,omitempty"`
	Edited      bool                   `json:"edited,omitempty"`
	Smoothed    bool                   `json:"smoothed,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// PostDetectionsResponse echoes what was stored plus any per-box rejections.
type PostDetectionsResponse struct {
	RunId        string               `json:"runId"`
	TracksStored int                  `json:"tracksStored"`
	BoxesStored  int                  `json:"boxesStored"`
	Rejected     []DetectionRejection `json:"rejected"`
	Warnings     []string             `json:"warnings"`
}

// DetectionRejection identifies a single box that failed validation.
type DetectionRejection struct {
	TrackId string `json:"trackId"`
	Frame   int64  `json:"frame"`
	Reason  string `json:"reason"`
}

type PostDetectionsSuccessResponse struct {
	SuccessResponse
	Data PostDetectionsResponse `json:"data"`
}
type PostDetectionsErrorResponse struct {
	ErrorResponse
}

// GetDetectionsResponse is the list of detection runs stored for a recording.
type GetDetectionsResponse struct {
	Key  string                `json:"key"`
	Runs []models.DetectionRun `json:"runs"`
}

type GetDetectionsSuccessResponse struct {
	SuccessResponse
	Data GetDetectionsResponse `json:"data"`
}
type GetDetectionsErrorResponse struct {
	ErrorResponse
}

// GetDetectionRunSuccessResponse returns a single detection run.
type GetDetectionRunSuccessResponse struct {
	SuccessResponse
	Data models.DetectionRun `json:"data"`
}
type GetDetectionRunErrorResponse struct {
	ErrorResponse
}

// DeleteDetectionRunResponse echoes the deleted run id.
type DeleteDetectionRunResponse struct {
	RunId   string `json:"runId"`
	Deleted bool   `json:"deleted"`
}

type DeleteDetectionRunSuccessResponse struct {
	SuccessResponse
	Data DeleteDetectionRunResponse `json:"data"`
}
type DeleteDetectionRunErrorResponse struct {
	ErrorResponse
}

// FlexibleString unmarshals from either a JSON string or a JSON number,
// coercing both to a string. Producers commonly send track ids as integers.
type FlexibleString string

func (fs *FlexibleString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*fs = ""
		return nil
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*fs = FlexibleString(s)
		return nil
	}
	// Not a string: keep the raw token (e.g. a number) as-is.
	*fs = FlexibleString(strings.TrimSpace(string(data)))
	return nil
}

func (fs FlexibleString) String() string {
	return string(fs)
}

// SubmitFaceRedaction is the worker-bound payload published by hub-api
// to the redaction queue. The user-facing endpoint is the generic
// POST /tasks/{taskId}/media-edits — this struct is the internal queue
// message format consumed by hub-pipeline-redaction.
//
// TaskId / CaseMediaId tie the job back to the CaseMedia entry that
// hub-api created when accepting the request. DestinationKey /
// DestinationProvider are computed server-side and tell the worker
// exactly where to upload the rendered artefact so the storage layout
// stays under hub-api's control.
type SubmitFaceRedactionRequest struct {
	AnalysisId          string                        `json:"analysisId"`
	TaskId              string                        `json:"taskId"`
	CaseMediaId         string                        `json:"caseMediaId"`
	SignedUrl           string                        `json:"signedUrl"`
	FileName            string                        `json:"fileName"`
	DestinationKey      string                        `json:"destinationKey"`
	DestinationProvider string                        `json:"destinationProvider"`
	EditType            models.CaseMediaEditType      `json:"editType,omitempty"`
	Mode                models.RedactionMode          `json:"mode,omitempty"`
	AllFrameCoordinates map[string][]*models.TrackBox `json:"allFrameCoordinates,omitempty"`
	FaceRedaction       *models.FaceRedaction         `json:"faceRedaction,omitempty"`
}
type SubmitFaceRedactionResponse struct {
	AnalysisId    string                `json:"analysisId"`
	FaceRedaction *models.FaceRedaction `json:"faceRedaction"`
	Status        AnalysisStatus        `json:"status"`
}
type SubmitFaceRedactionSuccessResponse struct {
	SuccessResponse
	Data SubmitFaceRedactionResponse `json:"data"`
}
type SubmitFaceRedactionErrorResponse struct {
	ErrorResponse
}

// FaceRedactionMessage is the structured payload published by hub-api
// onto the redaction queue when a face redaction is submitted. It carries
// the user context and a free-form Data map that is wrapped by the worker
// into a concrete request struct.
type FaceRedactionMessage struct {
	Events []string               `json:"events,omitempty"`
	User   models.User            `json:"user,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// CaseMediaStatusEvent is the structured payload published by edit
// workers (hub-pipeline-redaction first; trim/composite later) onto the
// analysis queue to drive lifecycle transitions of a CaseMedia entry.
// hub-pipeline-analysis owns the corresponding MongoDB writes against
// the case_media collection.
//
// File / Provider carry the storage location of the rendered artefact
// and are populated on the terminal Completed event so consumers can
// resolve the produced media without an additional lookup.
type CaseMediaStatusEvent struct {
	TaskId         string                 `json:"taskId"`
	CaseMediaId    string                 `json:"caseMediaId"`
	OrganisationId string                 `json:"organisationId"`
	Status         models.CaseMediaStatus `json:"status"`
	StatusError    string                 `json:"statusError,omitempty"`
	File           string                 `json:"file,omitempty"`
	Provider       string                 `json:"provider,omitempty"`
}

// GetAnalysis
// @Router /analysis [get]
type GetAnalysisRequest struct {
	AnalysisId string `json:"analysisId"`
}
type GetAnalysisResponse struct {
	Analysis models.Analysis `json:"analysis"`
}
type GetAnalysisSuccessResponse struct {
	SuccessResponse
	Data GetAnalysisResponse `json:"data"`
}
type GetAnalysisErrorResponse struct {
	ErrorResponse
}
