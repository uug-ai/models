package api

import "time"

type ErrorResponse struct {
	StatusCode int      `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	RequestId  string   `json:"requestId,omitempty" bson:"requestId,omitempty"`
	MetaData   Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
	ErrorCode  string   `json:"errorCode,omitempty" bson:"errorCode,omitempty"` // More specific custom error code or type
	Message    string   `json:"message,omitempty" bson:"message,omitempty"`
}

func CreateError(statusCode int, errorCode string, message string, metadata Metadata) ErrorResponse {
	metadata.Timestamp = time.Now().Unix()
	return ErrorResponse{
		StatusCode: statusCode,
		MetaData:   metadata,
		ErrorCode:  errorCode,
		Message:    message,
	}
}
