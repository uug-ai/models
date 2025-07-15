package api

type Metadata struct {
	Timestamp      int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	TraceId        string `json:"traceId,omitempty" bson:"traceId,omitempty"`
	OrganisationId string `json:"organisationId,omitempty" bson:"organisationId,omitempty"`
	UserId         string `json:"userId,omitempty" bson:"userId,omitempty"`
	Path           string `json:"path,omitempty" bson:"path,omitempty"`
	Error          string `json:"error,omitempty" bson:"error,omitempty"` // Error message if any
}
