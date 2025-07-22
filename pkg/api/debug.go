package api

import (
	"time"

	"github.com/sirupsen/logrus"
)

func LogDebug(logger *logrus.Logger, message string, metadata Metadata) {
	metadata.Timestamp = time.Now().Unix()
	logger.WithFields(logrus.Fields{
		"httpCode":        0,
		"applicationCode": ApplicationStatusSuccess,
		"message":         message,
		"metaData":        metadata,
	}).Debug()
}
