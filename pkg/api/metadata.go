package api

type Metadata struct {
	TraceId        string `json:"traceId,omitempty" bson:"traceId,omitempty"`
	OrganisationId string `json:"organisationId,omitempty" bson:"organisationId,omitempty"`
	UserId         string `json:"userId,omitempty" bson:"userId,omitempty"`
	Timestamp      int64  `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Path           string `json:"path,omitempty" bson:"path,omitempty"`
	Message        string `json:"message,omitempty" bson:"message,omitempty"`
}
