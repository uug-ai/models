package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	StatusCode int      `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	ErrorCode  int      `json:"errorCode,omitempty" bson:"errorCode,omitempty"` // More specific custom error code or type
	Message    string   `json:"message,omitempty" bson:"message,omitempty"`
	MetaData   Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
}

func CreateError(statusCode int, errorCode int, message string, metadata Metadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		StatusCode: statusCode,
		ErrorCode:  errorCode,
		Message:    message,
		MetaData:   metadata,
	}
}

func LogError(logger *logrus.Logger, message string, metadata Metadata) {
	logger.WithFields(logrus.Fields{
		"statusCode": 0,
		"errorCode":  StatusError,
		"message":    message,
		"metaData":   metadata,
	}).Error()
}
