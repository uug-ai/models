package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// https://pkg.go.dev/net/http#pkg-constants
const (
	HttpNoStatus                  int = 0
	HttpStatusOK                  int = http.StatusOK
	HttpStatusCreated             int = http.StatusCreated
	HttpStatusUnauthorized        int = http.StatusUnauthorized
	HttpStatusBadRequest          int = http.StatusBadRequest
	HttpStatusInternalServerError int = http.StatusInternalServerError
	HttpStatusNotFound            int = http.StatusNotFound
	HttpStatusServiceUnavailable  int = http.StatusServiceUnavailable
	HttpStatusGone                int = http.StatusGone
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
	Timestamp      int64          `json:"timestamp,omitempty" bson:"timestamp,omitempty"`           // Timestamp of the request or response
	TraceId        string         `json:"traceId,omitempty" bson:"traceId,omitempty"`               // Trace ID for tracking requests
	OrganisationId string         `json:"organisationId,omitempty" bson:"organisationId,omitempty"` // Organisation ID for the request
	UserId         string         `json:"userId,omitempty" bson:"userId,omitempty"`                 // User ID of the user making the request
	Path           string         `json:"path,omitempty" bson:"path,omitempty"`                     // Path of the request
	Function       string         `json:"function,omitempty" bson:"function,omitempty"`             // Function name where the response was generated
	Error          string         `json:"error,omitempty" bson:"error,omitempty"`                   // Error message if any
	MissingFields  []string       `json:"missingFields,omitempty" bson:"missingFields,omitempty"`   // List of missing fields in the request
	Language       string         `json:"language,omitempty" bson:"language,omitempty"`             // Language of the response, if applicable
	Data           map[string]any `json:"data,omitempty" bson:"data,omitempty"`
	// Additional data relevant to the request or response, this can be free-format
	Pagination *CursorPagination `json:"pagination,omitempty" bson:"pagination,omitempty"`
}

// SuccessResponse represents a standard success response structure.
type SuccessResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the response
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific status code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific status code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Success message describing the operation
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the response, such as timestamps and request IDs
}

func CreateSuccess(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateSuccessLog(logger *logrus.Logger, successResponse SuccessResponse, data interface{}) logrus.Fields {
	data = redactDataForLogging(data)
	return logrus.Fields{
		"httpStatusCode":        successResponse.HttpStatusCode,
		"applicationStatusCode": successResponse.ApplicationStatusCode,
		"entityStatusCode":      successResponse.EntityStatusCode,
		"message":               successResponse.Message,
		"metadata":              successResponse.Metadata,
		"data":                  data,
	}
}

// Logging functions for success and error responses
func LogInfo(logger *logrus.Logger, successResponse SuccessResponse, data interface{}) {
	logger.WithFields(CreateSuccessLog(logger, successResponse, data)).Info()
}

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateError(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateErrorLog(logger *logrus.Logger, errorResponse ErrorResponse) logrus.Fields {
	return logrus.Fields{
		"httpStatusCode":        errorResponse.HttpStatusCode,
		"applicationStatusCode": ApplicationStatusError,
		"entityStatusCode":      errorResponse.EntityStatusCode,
		"message":               errorResponse.Message,
		"metadata":              errorResponse.Metadata,
	}
}

func LogError(logger *logrus.Logger, errorResponse ErrorResponse) {
	logger.WithFields(CreateErrorLog(logger, errorResponse)).Error()
}

// Debug
type DebugResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateDebug(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) DebugResponse {
	metadata.Timestamp = time.Now().Unix()
	return DebugResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateDebugLog(logger *logrus.Logger, debugResponse DebugResponse) logrus.Fields {
	return logrus.Fields{
		"httpStatusCode":        debugResponse.HttpStatusCode,
		"applicationStatusCode": debugResponse.ApplicationStatusCode,
		"entityStatusCode":      debugResponse.EntityStatusCode,
		"message":               debugResponse.Message,
		"metadata":              debugResponse.Metadata,
	}
}

func LogDebug(logger *logrus.Logger, debugResponse DebugResponse, data interface{}) {
	logger.WithFields(CreateDebugLog(logger, debugResponse)).Debug()
}

// Trace

type TraceResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateTrace(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) TraceResponse {
	metadata.Timestamp = time.Now().Unix()
	return TraceResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateTraceLog(logger *logrus.Logger, traceResponse TraceResponse) logrus.Fields {
	return logrus.Fields{
		"httpStatusCode":        traceResponse.HttpStatusCode,
		"applicationStatusCode": traceResponse.ApplicationStatusCode,
		"entityStatusCode":      traceResponse.EntityStatusCode,
		"message":               traceResponse.Message,
		"metadata":              traceResponse.Metadata,
	}
}

func LogTrace(logger *logrus.Logger, traceResponse TraceResponse, data interface{}) {
	logger.WithFields(CreateTraceLog(logger, traceResponse)).Trace()
}

// Warning

type WarningResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateWarning(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) WarningResponse {
	metadata.Timestamp = time.Now().Unix()
	return WarningResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateWarningLog(logger *logrus.Logger, warningResponse WarningResponse) logrus.Fields {
	return logrus.Fields{
		"httpStatusCode":        warningResponse.HttpStatusCode,
		"applicationStatusCode": warningResponse.ApplicationStatusCode,
		"entityStatusCode":      warningResponse.EntityStatusCode,
		"message":               warningResponse.Message,
		"metadata":              warningResponse.Metadata,
	}
}

func LogWarning(logger *logrus.Logger, warningResponse WarningResponse, data interface{}) {
	logger.WithFields(CreateWarningLog(logger, warningResponse)).Warning()
}

// Fatal

type FatalResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateFatal(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) FatalResponse {
	metadata.Timestamp = time.Now().Unix()
	return FatalResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateFatalLog(logger *logrus.Logger, fatalResponse FatalResponse) logrus.Fields {
	return logrus.Fields{
		"httpStatusCode":        fatalResponse.HttpStatusCode,
		"applicationStatusCode": fatalResponse.ApplicationStatusCode,
		"entityStatusCode":      fatalResponse.EntityStatusCode,
		"message":               fatalResponse.Message,
		"metadata":              fatalResponse.Metadata,
	}
}

func LogFatal(logger *logrus.Logger, fatalResponse FatalResponse, data interface{}) {
	logger.WithFields(CreateFatalLog(logger, fatalResponse)).Fatal()
}

// Panic

type PanicResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreatePanic(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata) PanicResponse {
	metadata.Timestamp = time.Now().Unix()
	return PanicResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreatePanicLog(logger *logrus.Logger, panicResponse PanicResponse) logrus.Fields {
	return logrus.Fields{
		"httpStatusCode":        panicResponse.HttpStatusCode,
		"applicationStatusCode": panicResponse.ApplicationStatusCode,
		"entityStatusCode":      panicResponse.EntityStatusCode,
		"message":               panicResponse.Message,
		"metadata":              panicResponse.Metadata,
	}
}

func LogPanic(logger *logrus.Logger, panicResponse PanicResponse, data interface{}) {
	logger.WithFields(CreatePanicLog(logger, panicResponse)).Panic()
}

// redactDataForLogging truncates strings longer than 50 characters to prevent excessive logging
func redactDataForLogging(data interface{}) interface{} {
	const maxLength = 50

	switch v := data.(type) {
	case string:
		if len(v) > maxLength {
			return v[:maxLength] + "... [REDACTED]"
		}
		return v
	case map[string]interface{}:
		redacted := make(map[string]interface{})
		for key, val := range v {
			redacted[key] = redactDataForLogging(val)
		}
		return redacted
	case []interface{}:
		redacted := make([]interface{}, len(v))
		for i, val := range v {
			redacted[i] = redactDataForLogging(val)
		}
		return redacted
	default:
		return v
	}
}
