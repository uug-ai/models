package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Subscription struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisation_id,omitempty" bson:"organisation_id,omitempty"` // Organisation this subscription belongs to
	UserId         string             `json:"user_id,omitempty" bson:"user_id,omitempty"`                 // Legacy: user who created/owns the subscription
	StripeId       string             `json:"stripe_id,omitempty" bson:"stripe_id,omitempty"`
	StripePlan     string             `json:"stripe_plan,omitempty" bson:"stripe_plan,omitempty"`
	Quantity       int                `json:"quantity,omitempty" bson:"quantity,omitempty"`
	Status         string             `json:"status,omitempty" bson:"status,omitempty"` // active, cancelled, past_due, trialing, etc.
	TrialEndsAt    time.Time          `json:"trial_ends_at,omitempty" bson:"trial_ends_at,omitempty"`
	EndsAt         time.Time          `json:"ends_at,omitempty" bson:"ends_at,omitempty"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	CreatedAt      time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
}
