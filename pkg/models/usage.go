package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UsageKind is what a usage counter counts.
type UsageKind string

const (
	// UsageKindUploadBytes counts uploaded recording bytes, against
	// uploadGBPerDay (day periods) and uploadGBPerMonth (month periods).
	UsageKindUploadBytes UsageKind = "uploadBytes"
	// UsageKindWorkflowCredits counts workflow credits per month, against
	// workflowCredits.
	UsageKindWorkflowCredits UsageKind = "workflowCredits"
)

// UsageCounter counts one kind of usage of an organisation, or of one of its
// projects, in one period. Collection usage, unique
// {organisationId, projectId, kind, period}. Counters are incremented with
// $inc; a new period is a new document, so nothing resets.
type UsageCounter struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OrganisationId primitive.ObjectID `json:"organisationId" bson:"organisationId"`
	// ProjectId is nil for the organisation total.
	ProjectId *primitive.ObjectID `json:"projectId,omitempty" bson:"projectId,omitempty"`
	Kind      UsageKind           `json:"kind" bson:"kind"`
	// Period is a day ("2026-11-16", UsageDayPeriod) or a month ("2026-11",
	// UsageMonthPeriod) in the organisation's timezone.
	Period string `json:"period" bson:"period"`
	Value  int64  `json:"value" bson:"value"`
	// Reserved are credits reserved by running workflows.
	Reserved  int64     `json:"reserved,omitempty" bson:"reserved,omitempty"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

const (
	// UsageDayPeriodLayout formats a day period.
	UsageDayPeriodLayout = "2006-01-02"
	// UsageMonthPeriodLayout formats a month period.
	UsageMonthPeriodLayout = "2006-01"
	// DefaultOrganisationTimezone applies when an organisation has no
	// settings.timezone.
	DefaultOrganisationTimezone = "Europe/Brussels"
)

// OrganisationLocation loads an organisation's IANA timezone
// (OrganisationSettings.Timezone); an empty timezone is
// DefaultOrganisationTimezone.
func OrganisationLocation(timezone string) (*time.Location, error) {
	if timezone == "" {
		timezone = DefaultOrganisationTimezone
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("organisation timezone %q: %w", timezone, err)
	}
	return location, nil
}

// UsageDayPeriod returns the day period ("2006-01-02") that at falls in, in
// the given timezone (empty: DefaultOrganisationTimezone).
func UsageDayPeriod(at time.Time, timezone string) (string, error) {
	location, err := OrganisationLocation(timezone)
	if err != nil {
		return "", err
	}
	return at.In(location).Format(UsageDayPeriodLayout), nil
}

// UsageMonthPeriod returns the month period ("2006-01") that at falls in, in
// the given timezone (empty: DefaultOrganisationTimezone).
func UsageMonthPeriod(at time.Time, timezone string) (string, error) {
	location, err := OrganisationLocation(timezone)
	if err != nil {
		return "", err
	}
	return at.In(location).Format(UsageMonthPeriodLayout), nil
}

// NextUsageDay returns the start of the day after the one at falls in
// (midnight in the given timezone), e.g. for Retry-After on a capped upload.
func NextUsageDay(at time.Time, timezone string) (time.Time, error) {
	location, err := OrganisationLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	local := at.In(location)
	return time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, location), nil
}

// NextUsageMonth returns the start of the month after the one at falls in
// (midnight on the first, in the given timezone).
func NextUsageMonth(at time.Time, timezone string) (time.Time, error) {
	location, err := OrganisationLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	local := at.In(location)
	return time.Date(local.Year(), local.Month()+1, 1, 0, 0, 0, 0, location), nil
}
