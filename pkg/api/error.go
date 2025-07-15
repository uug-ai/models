package api

import "time"

type ErrorResponse struct {
	StatusCode int      `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	ErrorCode  int      `json:"errorCode,omitempty" bson:"errorCode,omitempty"` // More specific custom error code or type
	Error      error    `json:"error,omitempty" bson:"error,omitempty"`
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
