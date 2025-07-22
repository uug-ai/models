package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

type SuccessResponse struct {
	HttpCode        int      `json:"httpCode,omitempty" bson:"httpCode,omitempty"`
	ApplicationCode int      `json:"applicationCode,omitempty" bson:"applicationCode,omitempty"`
	Message         string   `json:"message,omitempty" bson:"message,omitempty"`
	MetaData        Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
}

func CreateSuccess(statusCode int, successCode int, message string, metadata Metadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		HttpCode:        statusCode,
		ApplicationCode: successCode,
		Message:         message,
		MetaData:        metadata,
	}
}

func LogSuccess(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpCode":        0,
		"applicationCode": ApplicationSuccess,
		"message":         message,
		"metaData":        metadata,
	}).Info()
}
