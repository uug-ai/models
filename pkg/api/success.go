package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

type SuccessResponse struct {
	HttpStatusCode        int      `json:"httpStatusCode,omitempty" bson:"httpStatusCode,omitempty"`
	ApplicationStatusCode int      `json:"applicationStatusCode,omitempty" bson:"applicationStatusCode,omitempty"`
	Message               string   `json:"message,omitempty" bson:"message,omitempty"`
	MetaData              Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
}

func CreateSuccess(httpStatusCode int, applicationStatusCode int, message string, metadata Metadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		HttpStatusCode:        httpStatusCode,
		ApplicationStatusCode: applicationStatusCode,
		Message:               message,
		MetaData:              metadata,
	}
}

func LogSuccess(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpStatusCode":        0,
		"applicationStatusCode": ApplicationStatusSuccess,
		"message":               message,
		"metaData":              metadata,
	}).Info()
}
