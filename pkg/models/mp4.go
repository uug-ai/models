package models

type FragmentedBytesRangeOnTime struct {
	Duration string `json:"duration" bson:"duration"`
	Time     string `json:"time" bson:"time"`
	Range    string `json:"range" bson:"range"`
}
