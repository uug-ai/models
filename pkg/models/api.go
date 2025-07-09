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
	UserId    string `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Error     error  `json:"error,omitempty" bson:"error,omitempty"`
	Path      string `json:"path,omitempty" bson:"path,omitempty"`
}

type ErrorResponse struct {
	StatusCode int         `json:"status_code,omitempty" bson:"status_code,omitempty"`
	RequestId  string      `json:"request_id,omitempty" bson:"request_id,omitempty"`
	MetaData   APIMetadata `json:"meta_data,omitempty" bson:"meta_data,omitempty"`
	ErrorCode  string      `json:"error_code,omitempty" bson:"error_code,omitempty"` // More specific custom error code or type
	Message    string      `json:"message,omitempty" bson:"message,omitempty"`
}

type SuccessResponse struct {
	StatusCode  int         `json:"status_code,omitempty" bson:"status_code,omitempty"`
	RequestId   string      `json:"request_id,omitempty" bson:"request_id,omitempty"`
	MetaData    APIMetadata `json:"meta_data,omitempty" bson:"meta_data,omitempty"`
	SuccessCode string      `json:"success_code,omitempty" bson:"success_code,omitempty"`
	Message     string      `json:"message,omitempty" bson:"message,omitempty"`
}

// Example of Error reponse
//
//	{
//		"statuscode": 400,
//		"requestId": "123456",
//		"meta_data": {
//			"user_id": "123456",
//			"timestamp": 1234567890,
//			"error": "error"
//			"path": "/api/v1/xxx"
//		},
//		"errorCode": "error_code",
//		"message": "Error message"
//	}
func CreateError(statusCode int, errorCode string, message string, metadata APIMetadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		StatusCode: statusCode,
		MetaData:   metadata,
		ErrorCode:  errorCode,
		Message:    message,
	}
}

// Example of Success reponse
//
//	{
//			"statuscode": 200,
//			"requestId": "123456",
//			"meta_data": {
//				"user_id": "123456",
//				"timestamp": 1234567890,
//				"error": "error"
//				"path": "/api/v1/xxx"
//			},
//			"successCode": "success_code",
//			"message": "Success message",
//			"data": {}
//		}
func CreateSuccess(statusCode int, successCode string, message string, metadata APIMetadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		StatusCode:  statusCode,
		MetaData:    metadata,
		SuccessCode: successCode,
		Message:     message,
	}
}
