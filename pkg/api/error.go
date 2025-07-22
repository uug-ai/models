package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`               // HTTP status code for the error
	ApplicationStatusCode int      `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"` // Application-specific error code
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`                             // Error message describing the issue
	MetaData              Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`                           // Additional metadata about the error, such as timestamps and request IDs
}

func CreateError(httpStatusCode int, applicationStatusCode int, message string, metadata Metadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		Message:               message,
		MetaData:              metadata,
	}
}

func LogError(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpStatusCode":        0,
		"applicationStatusCode": ApplicationStatusError,
		"message":               message,
		"metaData":              metadata,
	}).Error()
}
