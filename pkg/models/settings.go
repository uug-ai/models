package models

type Settings struct {
	Key string         `json:"key" bson:"key"`
	Map map[string]any `json:"map" bson:"map"` // @TODO replace this
}
