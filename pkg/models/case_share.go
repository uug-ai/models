package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type CaseShare struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	TaskId      string             `json:"task_id" bson:"task_id"`
	OwnerId     string             `json:"owner_id" bson:"owner_id"`
	OwnerEmail  string             `json:"owner_email" bson:"owner_email"`
	Email       string             `json:"email" bson:"email"`
	Token       string             `json:"token" bson:"token"`
	Permissions []string           `json:"permissions" bson:"permissions"` // e.g. ["view"]
	IsActive    bool               `json:"is_active" bson:"is_active"`
	ExpiresAt   int64              `json:"expires_at" bson:"expires_at"`
	CreatedAt   int64              `json:"created_at" bson:"created_at"`
	OTPCode     string             `json:"-" bson:"otp_code,omitempty"`
	OTPExpiry   int64              `json:"-" bson:"otp_expiry,omitempty"`
	OTPAttempts int                `json:"-" bson:"otp_attempts,omitempty"`
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
	TaskId  string `json:"task_id"`
	OwnerId string `json:"owner_id"`
}
type GetCaseSharesForTaskOutput struct {
	Shares []CaseShare `json:"shares"`
}

type DeleteCaseShareInput struct {
	ShareId string `json:"share_id"`
	OwnerId string `json:"owner_id"`
}
type DeleteCaseShareOutput struct{}
