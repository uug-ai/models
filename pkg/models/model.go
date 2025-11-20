package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// This is an example model struct that illustrates how a model should be constructed.
// It includes an identifier, relevant information, metadata, runtime metadata, and audit information.
// The idea of this model is that its structure is inherited fro more specific models like Media, User, Device, etc.

type Model struct {
	// Unique identifier for the model, this is used to retrieve the model from the database by its unique ID.
	Id primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`

	// Important information about the model, that you would like to display everytime the model is retrieved (for example in a table or list).
	// e.g.
	// User model -> username, email, roles, etc.
	// Device model -> name, deviceId, status, etc.
	// Media model -> filename, mediaType, size, etc.
	// Marker model -> name, events type, timestamps, etc.

	// Metadata is additional information about the model that is not critical for its primary function,
	// but can provide useful context or details.
	// e.g.
	// User model -> profile information, preferences, etc.
	// Device model -> location, installation date, etc.
	// Media model -> tags, description, etc.
	// Marker model -> comments, tags, etc.
	Metadata *MediaMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// AtRuntimeMetadata are computed or dynamic information about the model that is relevant during its usage or processing.
	// This information is generated at run time and is not stored into the database. All other information of the model is persisted into the database.
	// and is not altered during runtime.
	// e.g.
	// Media model -> signed URLS for accessing the media, processing status, etc.
	// Device model -> current status (active, inactive), etc.
	AtRuntimeMetadata *MediaAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Audit information: every model should have audit information to track its creation and modification history.
	// This is important for maintaining data integrity and accountability.
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

type ModelOption struct {
	Value string `bson:"value" json:"value"`
	Text  string `bson:"text" json:"text"`
}
