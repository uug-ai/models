package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// https://pkg.go.dev/net/http#pkg-constants
const (
	HttpStatusOK                  int = http.StatusOK
	HttpStatusCreated             int = http.StatusCreated
	HttpStatusUnauthorized        int = http.StatusUnauthorized
	HttpStatusBadRequest          int = http.StatusBadRequest
	HttpStatusInternalServerError int = http.StatusInternalServerError
	HttpStatusNotFound            int = http.StatusNotFound
)

// Custom status codes for specific operations
const (
	ApplicationStatusSuccess         string = "success"
	ApplicationStatusError           string = "error"
	ApplicationStatusGetSuccess      string = "get_success"
	ApplicationStatusGetFailed       string = "get_failed"
	ApplicationStatusGetSuccessEmpty string = "get_success_empty"
	ApplicationStatusAddSuccess      string = "add_success"
	ApplicationStatusAddFailed       string = "add_failed"
	ApplicationStatusUpdateSuccess   string = "update_success"
	ApplicationStatusUpdateFailed    string = "update_failed"
	ApplicationStatusDeleteSuccess   string = "delete_success"
	ApplicationStatusDeleteFailed    string = "delete_failed"
)

// EntityStatus should be a high-level status type that can be used across different entities
// Idea is that we have specific statuses for each entity type, but they all derive from a common EntityStatus type
// e.g. marker, user, etc. can all have their own specific statuses but also share common statuses like success, error, not found, etc.
type EntityStatus interface {
	String() string               // Returns the string representation of the status
	Translate(lang string) string // Returns the translated string representation of the status in the specified language
}

// Metadata holds additional information about the request or response.
// It can include timestamps, trace IDs, organisation IDs, user IDs, and other relevant data
type Metadata struct {
	Timestamp      int64    `json:"timestamp,omitempty" bson:"timestamp,omitempty"`           // Timestamp of the request or response
	TraceId        string   `json:"traceId,omitempty" bson:"traceId,omitempty"`               // Trace ID for tracking requests
	OrganisationId string   `json:"organisationId,omitempty" bson:"organisationId,omitempty"` // Organisation ID for the request
	UserId         string   `json:"userId,omitempty" bson:"userId,omitempty"`                 // User ID of the user making the request
	Path           string   `json:"path,omitempty" bson:"path,omitempty"`                     // Path of the request
	Function       string   `json:"function,omitempty" bson:"function,omitempty"`             // Function name where the response was generated
	Error          string   `json:"error,omitempty" bson:"error,omitempty"`                   // Error message if any
	MissingFields  []string `json:"missingFields,omitempty" bson:"missingFields,omitempty"`   // List of missing fields in the request
	Language       string   `json:"language,omitempty" bson:"language,omitempty"`             // Language of the response, if applicable
}

// SuccessResponse represents a standard success response structure.
type SuccessResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the response
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific status code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific status code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Success message describing the operation
	MetaData              Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`                           // Additional metadata about the response, such as timestamps and request IDs
}

func CreateSuccess(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		MetaData:              metadata,
	}
}

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	MetaData              Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateError(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		MetaData:              metadata,
	}
}

// Logging functions for success and error responses
func LogSuccess(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpStatusCode":        0,
		"applicationStatusCode": ApplicationStatusSuccess,
		"entityStatusCode":      "",
		"message":               message,
		"metaData":              metadata,
	}).Info()
}

func LogError(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpStatusCode":        0,
		"applicationStatusCode": ApplicationStatusError,
		"entityStatusCode":      "",
		"message":               message,
		"metaData":              metadata,
	}).Error()
}

func LogDebug(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpStatusCode":        0,
		"applicationStatusCode": ApplicationStatusSuccess,
		"entityStatusCode":      "",
		"message":               message,
		"metaData":              metadata,
	}).Debug()
}
