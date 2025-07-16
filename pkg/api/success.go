package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

type SuccessResponse struct {
	StatusCode  int      `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	SuccessCode int      `json:"successCode,omitempty" bson:"successCode,omitempty"`
	Message     string   `json:"message,omitempty" bson:"message,omitempty"`
	MetaData    Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
}

func CreateSuccess(statusCode int, successCode int, message string, metadata Metadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		StatusCode:  statusCode,
		SuccessCode: successCode,
		Message:     message,
		MetaData:    metadata,
	}
}

func LogSuccess(logger *logrus.Logger, message string, metadata Metadata) {
	logger.WithFields(logrus.Fields{
		"statusCode":  0,
		"successCode": StatusSuccess,
		"message":     message,
		"metaData":    metadata,
	}).Info()
}
