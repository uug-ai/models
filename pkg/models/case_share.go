package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type CaseShare struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TaskId         string             `json:"task_id" bson:"task_id"`
	UserId         string             `json:"user_id" bson:"user_id"`
	UserEmail      string             `json:"user_email" bson:"user_email"`
	OrganisationId string             `json:"organisation_id" bson:"organisation_id"`
	Email          string             `json:"email" bson:"email"`
	Token          string             `json:"token" bson:"token"`
	Permissions    []string           `json:"permissions" bson:"permissions"` // e.g. ["view"]
	IsActive       bool               `json:"is_active" bson:"is_active"`
	ExpiresAt      int64              `json:"expires_at" bson:"expires_at"`
	CreatedAt      int64              `json:"created_at" bson:"created_at"`
	OTPCode        string             `json:"-" bson:"otp_code,omitempty"`
	OTPExpiry      int64              `json:"-" bson:"otp_expiry,omitempty"`
	OTPAttempts    int                `json:"-" bson:"otp_attempts,omitempty"`

	// Selection is the per-share snapshot of the case_media ids the
	// recipient is allowed to browse, captured at CreateShare time.
	// It is the source of truth for what this specific token resolves
	// to — task.ShareSelection is only used as the owner-side
	// template that pre-fills the next share modal. Storing the
	// allow-list here decouples each recipient's view from later
	// edits to the task and from subsequent shares of the same case.
	//
	// nil  = legacy / unsnapshotted share — resolvers fall back to
	//        the task-level selection so old rows keep working.
	// []   = "include all" (same convention as the export pipeline).
	// [..] = literal allow-list of case_media ids.
	Selection []primitive.ObjectID `json:"selection,omitempty" bson:"selection,omitempty"`

	// AttachmentSelection is the per-share snapshot of the
	// task.Attachments[] ids the recipient is allowed to browse.
	// Same nil/empty/non-empty semantics as Selection.
	AttachmentSelection []primitive.ObjectID `json:"attachment_selection,omitempty" bson:"attachment_selection,omitempty"`
}

// Input/Output types for repository operations

type CreateCaseShareInput struct {
	Share CaseShare `json:"share"`
}
type CreateCaseShareOutput struct {
	Share *CaseShare `json:"share"`
}

type GetCaseShareByTokenInput struct {
	Token string `json:"token"`
}
type GetCaseShareByTokenOutput struct {
	Share *CaseShare `json:"share"`
}

type GetCaseSharesForTaskInput struct {
	TaskId string `json:"task_id"`
	UserId string `json:"user_id"`
}
type GetCaseSharesForTaskOutput struct {
	Shares []CaseShare `json:"shares"`
}

type DeleteCaseShareInput struct {
	ShareId string `json:"share_id"`
	UserId  string `json:"user_id"`
}
type DeleteCaseShareOutput struct{}

type UpdateCaseShareOTPInput struct {
	Token   string `json:"token"`
	OTPCode string `json:"otp_code"`
	Expiry  int64  `json:"otp_expiry"`
}
type UpdateCaseShareOTPOutput struct{}
