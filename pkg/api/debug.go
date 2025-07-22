package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

func LogDebug(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpStatusCode":        0,
		"applicationStatusCode": ApplicationStatusSuccess,
		"message":               message,
		"metaData":              metadata,
	}).Debug()
}
