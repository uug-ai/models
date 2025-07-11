package models

type Location struct {
	Geometry         LocationGeometry `json:"geometry" bson:"geometry,omitempty"`
	Description      string           `json:"description" bson:"description,omitempty"`
	FormattedAddress string           `json:"formatted_address" bson:"formatted_address,omitempty"`
	Country          string           `json:"country" bson:"country,omitempty"`
	CountryShort     string           `json:"country_short" bson:"country_short,omitempty"`
	PostalCode       string           `json:"postal_code" bson:"postal_code,omitempty"`
	Region           string           `json:"region" bson:"region,omitempty"`
	City             string           `json:"city" bson:"city,omitempty"`
	Street           string           `json:"street" bson:"street,omitempty"`
	StreetNumber     string           `json:"street_number" bson:"street_number,omitempty"`
}

type LocationShort struct {
	Geometry         LocationGeometry `json:"geometry" bson:"geometry,omitempty"`
	Description      string           `json:"description" bson:"description,omitempty"`
	FormattedAddress string           `json:"formatted_address" bson:"formatted_address,omitempty"`
}

type LocationGeometry struct {
	Location LocationGeometryLocation `json:"location" bson:"location,omitempty"`
}

type LocationGeometryLocation struct {
	Lat float64 `json:"lat" bson:"lat,omitempty"`
	Lng float64 `json:"lng" bson:"lng,omitempty"`
}
