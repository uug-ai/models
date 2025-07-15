package api

import "time"

type SuccessResponse struct {
	StatusCode  int      `json:"statusCode,omitempty" bson:"statusCode,omitempty"`
	RequestId   string   `json:"requestId,omitempty" bson:"requestId,omitempty"`
	MetaData    Metadata `json:"metaData,omitempty" bson:"metaData,omitempty"`
	SuccessCode int      `json:"successCode,omitempty" bson:"successCode,omitempty"`
	Message     string   `json:"message,omitempty" bson:"message,omitempty"`
}

func CreateSuccess(statusCode int, successCode int, message string, metadata Metadata) SuccessResponse {
	metadata.Timestamp = time.Now().Unix()
	return SuccessResponse{
		StatusCode:  statusCode,
		MetaData:    metadata,
		SuccessCode: successCode,
		Message:     message,
	}
}
