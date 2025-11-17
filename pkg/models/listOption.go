package models

type ListOption struct {
	Value string `bson:"value" json:"value"`
	Text  string `bson:"text" json:"text"`
}
