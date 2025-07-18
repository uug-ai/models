package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

func LogDebug(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"statusCode":  0,
		"successCode": StatusSuccess,
		"message":     message,
		"metaData":    metadata,
	}).Debug()
}
