package api

import "github.com/uug-ai/models/pkg/models"

// WorkflowRunOverview is an allowlisted operational overview, never a worker
// envelope. Source fields are populated only after source authorization.
type WorkflowRunOverview struct {
	RunId              string                              `json:"runId"`
	WorkflowId         string                              `json:"workflowId,omitempty"`
	WorkflowName       string                              `json:"workflowName,omitempty"`
	Origin             models.WorkflowRunOrigin            `json:"origin,omitempty"`
	State              models.WorkflowRunState             `json:"state"`
	Start              int64                               `json:"start"`
	End                int64                               `json:"end,omitempty"`
	DurationMs         int64                               `json:"durationMs,omitempty"`
	Dispatched         int                                 `json:"dispatched,omitempty"`
	Resolved           int                                 `json:"resolved,omitempty"`
	HasResults         bool                                `json:"hasResults,omitempty"`
	Operations         []models.WorkflowRunOperationStatus `json:"operations,omitempty"`
	StageExecutions    []models.WorkflowRunStageExecution  `json:"stageExecutions,omitempty"`
	SourceAccess       string                              `json:"sourceAccess" enums:"available,restricted,unavailable"`
	SourceType         string                              `json:"sourceType" enums:"media,case"`
	SourceLabel        string                              `json:"sourceLabel,omitempty"`
	SourceRef          string                              `json:"sourceRef,omitempty"`
	CaseMediaId        string                              `json:"caseMediaId,omitempty"`
	MediaId            string                              `json:"mediaId,omitempty"`
	Key                string                              `json:"key,omitempty"`
	DeviceKey          string                              `json:"deviceKey,omitempty"`
	DeviceName         string                              `json:"deviceName,omitempty"`
	RecordingTimestamp int64                               `json:"recordingTimestamp,omitempty"`
}
