package api

import (
	"net/http"
	"runtime"
	"strings"
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
	ApplicationName    string         `json:"applicationName,omitempty" bson:"applicationName,omitempty"`       // Name of the application
	ApplicationVersion string         `json:"applicationVersion,omitempty" bson:"applicationVersion,omitempty"` // Version of the application
	Timestamp          int64          `json:"timestamp,omitempty" bson:"timestamp,omitempty"`                   // Timestamp of the request or response
	TraceId            string         `json:"traceId,omitempty" bson:"traceId,omitempty"`                       // Trace ID for tracking requests
	OrganisationId     string         `json:"organisationId,omitempty" bson:"organisationId,omitempty"`         // Organisation ID for the request
	UserId             string         `json:"userId,omitempty" bson:"userId,omitempty"`                         // User ID of the user making the request
	MediaFileName      string         `json:"mediaFileName,omitempty" bson:"mediaFileName,omitempty"`           // Name of the media file involved in the request
	DeviceKey          string         `json:"deviceKey,omitempty" bson:"deviceKey,omitempty"`                   // Device key involved in the request
	Path               string         `json:"path,omitempty" bson:"path,omitempty"`                             // Path of the request
	Function           string         `json:"function,omitempty" bson:"function,omitempty"`                     // Function name where the response was generated
	Line               int            `json:"line,omitempty" bson:"line,omitempty"`                             // Line number in the code where the response was generated
	Error              string         `json:"error,omitempty" bson:"error,omitempty"`                           // Error message if any
	MissingFields      []string       `json:"missingFields,omitempty" bson:"missingFields,omitempty"`           // List of missing fields in the request
	Language           string         `json:"language,omitempty" bson:"language,omitempty"`                     // Language of the response, if applicable
	Data               map[string]any `json:"data,omitempty" bson:"data,omitempty"`
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

func CreateSuccess(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata, skipFrames ...int) SuccessResponse {
	level := 2 // default
	if len(skipFrames) > 0 {
		level = skipFrames[0]
	}
	metadata.Timestamp = time.Now().Unix()
	callerInfo := GetCallerInfo(level)
	metadata.Path = callerInfo.File
	metadata.Function = callerInfo.Function
	metadata.Line = callerInfo.Line
	return SuccessResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateSuccessLog(logger *logrus.Logger, successResponse SuccessResponse, data ...any) logrus.Fields {
	var payload any
	if len(data) > 0 {
		payload = redactDataForLogging(data[0])
	}
	return logrus.Fields{
		"httpStatusCode":        successResponse.HttpStatusCode,
		"applicationStatusCode": successResponse.ApplicationStatusCode,
		"entityStatusCode":      successResponse.EntityStatusCode,
		"message":               successResponse.Message,
		"metadata":              successResponse.Metadata,
		"data":                  payload,
	}
}

// Logging functions for success and error responses
func LogInfo(logger *logrus.Logger, successResponse SuccessResponse, data ...any) {
	logger.WithFields(CreateSuccessLog(logger, successResponse, data...)).Info()
}

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateError(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata, skipFrames ...int) ErrorResponse {
	level := 4 // default
	if len(skipFrames) > 0 {
		level = skipFrames[0]
	}
	metadata.Timestamp = time.Now().Unix()
	callerInfo := GetCallerInfo(level)
	metadata.Path = callerInfo.File
	metadata.Function = callerInfo.Function
	metadata.Line = callerInfo.Line
	return ErrorResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateErrorLog(logger *logrus.Logger, errorResponse ErrorResponse, data ...any) logrus.Fields {
	var payload any
	if len(data) > 0 {
		payload = redactDataForLogging(data[0])
	}
	return logrus.Fields{
		"httpStatusCode":        errorResponse.HttpStatusCode,
		"applicationStatusCode": ApplicationStatusError,
		"entityStatusCode":      errorResponse.EntityStatusCode,
		"message":               errorResponse.Message,
		"metadata":              errorResponse.Metadata,
		"data":                  payload,
	}
}

func LogError(logger *logrus.Logger, errorResponse ErrorResponse, data ...any) {
	logger.WithFields(CreateErrorLog(logger, errorResponse, data...)).Error()
}

// Debug
type DebugResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateDebug(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata, skipFrames ...int) DebugResponse {
	level := 4 // default
	if len(skipFrames) > 0 {
		level = skipFrames[0]
	}
	metadata.Timestamp = time.Now().Unix()
	callerInfo := GetCallerInfo(level)
	metadata.Path = callerInfo.File
	metadata.Function = callerInfo.Function
	metadata.Line = callerInfo.Line

	return DebugResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateDebugLog(logger *logrus.Logger, debugResponse DebugResponse, data ...any) logrus.Fields {
	var payload any
	if len(data) > 0 {
		payload = redactDataForLogging(data[0])
	}
	return logrus.Fields{
		"httpStatusCode":        debugResponse.HttpStatusCode,
		"applicationStatusCode": debugResponse.ApplicationStatusCode,
		"entityStatusCode":      debugResponse.EntityStatusCode,
		"message":               debugResponse.Message,
		"metadata":              debugResponse.Metadata,
		"data":                  payload,
	}
}

func LogDebug(logger *logrus.Logger, debugResponse DebugResponse, data ...any) {
	logger.WithFields(CreateDebugLog(logger, debugResponse, data...)).Debug()
}

// Trace

type TraceResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateTrace(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata, skipFrames ...int) TraceResponse {
	level := 4 // default
	if len(skipFrames) > 0 {
		level = skipFrames[0]
	}
	metadata.Timestamp = time.Now().Unix()
	callerInfo := GetCallerInfo(level)
	metadata.Path = callerInfo.File
	metadata.Function = callerInfo.Function
	metadata.Line = callerInfo.Line
	return TraceResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateTraceLog(logger *logrus.Logger, traceResponse TraceResponse, data ...any) logrus.Fields {
	var payload any
	if len(data) > 0 {
		payload = redactDataForLogging(data[0])
	}
	return logrus.Fields{
		"httpStatusCode":        traceResponse.HttpStatusCode,
		"applicationStatusCode": traceResponse.ApplicationStatusCode,
		"entityStatusCode":      traceResponse.EntityStatusCode,
		"message":               traceResponse.Message,
		"metadata":              traceResponse.Metadata,
		"data":                  payload,
	}
}

func LogTrace(logger *logrus.Logger, traceResponse TraceResponse, data ...any) {
	logger.WithFields(CreateTraceLog(logger, traceResponse, data...)).Trace()
}

// Warning

type WarningResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateWarning(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata, skipFrames ...int) WarningResponse {
	level := 4 // default
	if len(skipFrames) > 0 {
		level = skipFrames[0]
	}
	metadata.Timestamp = time.Now().Unix()
	callerInfo := GetCallerInfo(level)
	metadata.Path = callerInfo.File
	metadata.Function = callerInfo.Function
	metadata.Line = callerInfo.Line
	return WarningResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateWarningLog(logger *logrus.Logger, warningResponse WarningResponse, data ...any) logrus.Fields {
	var payload any
	if len(data) > 0 {
		payload = redactDataForLogging(data[0])
	}
	return logrus.Fields{
		"httpStatusCode":        warningResponse.HttpStatusCode,
		"applicationStatusCode": warningResponse.ApplicationStatusCode,
		"entityStatusCode":      warningResponse.EntityStatusCode,
		"message":               warningResponse.Message,
		"metadata":              warningResponse.Metadata,
		"data":                  payload,
	}
}

func LogWarning(logger *logrus.Logger, warningResponse WarningResponse, data ...any) {
	logger.WithFields(CreateWarningLog(logger, warningResponse, data...)).Warning()
}

// Fatal

type FatalResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateFatal(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata, skipFrames ...int) FatalResponse {
	level := 4 // default
	if len(skipFrames) > 0 {
		level = skipFrames[0]
	}
	metadata.Timestamp = time.Now().Unix()
	callerInfo := GetCallerInfo(level)
	metadata.Path = callerInfo.File
	metadata.Function = callerInfo.Function
	metadata.Line = callerInfo.Line
	return FatalResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreateFatalLog(logger *logrus.Logger, fatalResponse FatalResponse, data ...any) logrus.Fields {
	var payload any
	if len(data) > 0 {
		payload = redactDataForLogging(data[0])
	}
	return logrus.Fields{
		"httpStatusCode":        fatalResponse.HttpStatusCode,
		"applicationStatusCode": fatalResponse.ApplicationStatusCode,
		"entityStatusCode":      fatalResponse.EntityStatusCode,
		"message":               fatalResponse.Message,
		"metadata":              fatalResponse.Metadata,
		"data":                  payload,
	}
}

func LogFatal(logger *logrus.Logger, fatalResponse FatalResponse, data ...any) {
	logger.WithFields(CreateFatalLog(logger, fatalResponse, data...)).Fatal()
}

// Panic

type PanicResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode string   `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	EntityStatusCode      string   `json:"entityStatusCode,omitempty" bson:"entityStatusCode,omitempty"`           // Entity-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	Metadata              Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreatePanic(httpStatusCode int, applicationStatusCode string, entityStatusCode EntityStatus, metadata Metadata, skipFrames ...int) PanicResponse {
	level := 4 // default
	if len(skipFrames) > 0 {
		level = skipFrames[0]
	}
	metadata.Timestamp = time.Now().Unix()
	callerInfo := GetCallerInfo(level)
	metadata.Path = callerInfo.File
	metadata.Function = callerInfo.Function
	metadata.Line = callerInfo.Line
	return PanicResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		EntityStatusCode:      entityStatusCode.String(),
		Message:               entityStatusCode.Translate(metadata.Language),
		Metadata:              metadata,
	}
}

func CreatePanicLog(logger *logrus.Logger, panicResponse PanicResponse, data ...any) logrus.Fields {
	var payload any
	if len(data) > 0 {
		payload = redactDataForLogging(data[0])
	}
	return logrus.Fields{
		"httpStatusCode":        panicResponse.HttpStatusCode,
		"applicationStatusCode": panicResponse.ApplicationStatusCode,
		"entityStatusCode":      panicResponse.EntityStatusCode,
		"message":               panicResponse.Message,
		"metadata":              panicResponse.Metadata,
		"data":                  payload,
	}
}

func LogPanic(logger *logrus.Logger, panicResponse PanicResponse, data ...any) {
	logger.WithFields(CreatePanicLog(logger, panicResponse, data...)).Panic()
}

// redactDataForLogging truncates strings longer than 50 characters to prevent excessive logging
func redactDataForLogging(data any) interface{} {
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

type CallerInfo struct {
	File     string
	Line     int
	Function string
}

func GetCallerInfo(skipFrames int) CallerInfo {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(skipFrames, pcs)
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		// Filter out internal logging helpers (adjust package path as needed)
		if !strings.Contains(frame.Function, "github.com/uug-ai/models/pkg/api.Log") &&
			!strings.Contains(frame.Function, "github.com/uug-ai/models/pkg/api.Create") {
			return CallerInfo{
				File:     frame.File,
				Line:     frame.Line,
				Function: frame.Function,
			}
		}
		if !more {
			break
		}
	}
	return CallerInfo{}
}
