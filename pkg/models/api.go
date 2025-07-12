package models

import (
	"net/http"
	"time"
)

// https://pkg.go.dev/net/http#pkg-constants
const (
	StatusOK                  int = http.StatusOK
	StatusCreated             int = http.StatusCreated
	StatusUnauthorized        int = http.StatusUnauthorized
	StatusBadRequest          int = http.StatusBadRequest
	StatusInternalServerError int = http.StatusInternalServerError
	StatusNotFound            int = http.StatusNotFound
)

type APIMetadata struct {
	TraceId        string `json:"traceId,omitempty" bson:"traceId,omitempty"`
	OrganisationId string `json:"organisationId,omitempty" bson:"organisationId,omitempty"`
	UserId         string `json:"userId,omitempty" bson:"userId,omitempty"`
	Timestamp      int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Error          error  `json:"error,omitempty" bson:"error,omitempty"`
	Path           string `json:"path,omitempty" bson:"path,omitempty"`
}

type ErrorResponse struct {
	StatusCode int         `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	RequestId  string      `json:"requestId,omitempty" bson:"requestId,omitempty"`
	MetaData   APIMetadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
	ErrorCode  string      `json:"errorCode,omitempty" bson:"errorCode,omitempty"` // More specific custom error code or type
	Message    string      `json:"message,omitempty" bson:"message,omitempty"`
}

type SuccessResponse struct {
	StatusCode  int         `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	RequestId   string      `json:"requestId,omitempty" bson:"requestId,omitempty"`
	MetaData    APIMetadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
	SuccessCode string      `json:"successCode,omitempty" bson:"successCode,omitempty"`
	Message     string      `json:"message,omitempty" bson:"message,omitempty"`
}

func CreateError(statusCode int, errorCode string, message string, metadata APIMetadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		StatusCode: statusCode,
		MetaData:   metadata,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

func CreateSuccess(statusCode int, successCode string, message string, metadata APIMetadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		StatusCode:  statusCode,
		MetaData:    metadata,
		SuccessCode: successCode,
		Message:     message,
	}
}
