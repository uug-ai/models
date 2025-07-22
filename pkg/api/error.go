package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	HttpCode        int      `json:"httpCode,omitempty" bson:"httpCode,omitempty"`               // HTTP status code for the error
	ApplicationCode int      `json:"applicationCode,omitempty" bson:"applicationCode,omitempty"` // Application-specific error code
	Message         string   `json:"message,omitempty" bson:"message,omitempty"`                 // Error message describing the issue
	MetaData        Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`               // Additional metadata about the error, such as timestamps and request IDs
}

func CreateError(httpStatusCode int, applicationStatusCode int, message string, metadata Metadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		HttpCode:        httpStatusCode,
		ApplicationCode: applicationStatusCode,
		Message:         message,
		MetaData:        metadata,
	}
}

func LogError(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpCode":        0,
		"applicationCode": ApplicationStatusError,
		"message":         message,
		"metaData":        metadata,
	}).Error()
}
