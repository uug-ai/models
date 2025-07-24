package models

type Preset struct {
	Name  string  `json:"name" bson:"name"`
	Token string  `json:"token" bson:"token"`
	X     float64 `json:"x" bson:"x"`
	Y     float64 `json:"y" bson:"y"`
	Z     float64 `json:"z" bson:"z"`
}

type Tour struct {
	Name    string   `json:"name" bson:"name,omitempty"`
	Presets []Preset `json:"presets" bson:"presets,omitempty"`
	Current int      `json:"current" bson:"current,omitempty"`
	Running bool     `json:"running" bson:"running,omitempty"`
	Loop    bool     `json:"loop" bson:"loop,omitempty"`
	Speed   float64  `json:"speed" bson:"speed,omitempty"`
}
