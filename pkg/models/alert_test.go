package models

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func TestAlertPrivacyCompatibility(t *testing.T) {
	for _, private := range []bool{false, true} {
		alert := CustomAlert{Private: private, Enabled: true, UserId: "creator"}
		data, err := bson.Marshal(alert)
		if err != nil {
			t.Fatal(err)
		}
		var decoded CustomAlert
		if err := bson.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Private != private || !decoded.Enabled || decoded.UserId != alert.UserId {
			t.Fatalf("alert round trip = %#v", decoded)
		}
	}
	var legacy CustomAlert
	if err := json.Unmarshal([]byte(`{"enabled":true,"user_id":"creator"}`), &legacy); err != nil {
		t.Fatal(err)
	}
	if legacy.Private {
		t.Fatal("legacy alerts must retain shared visibility")
	}
}

func TestAlertPrivacyPatchPreservesExplicitFalse(t *testing.T) {
	for _, test := range []struct {
		payload string
		want    *bool
	}{
		{payload: `{}`},
		{payload: `{"private":true}`, want: boolPointer(true)},
		{payload: `{"private":false}`, want: boolPointer(false)},
	} {
		var patch AlertPatch
		if err := json.Unmarshal([]byte(test.payload), &patch); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(patch.Private, test.want) {
			t.Fatalf("%s: private = %v, want %v", test.payload, patch.Private, test.want)
		}
		data, err := bson.Marshal(patch)
		if err != nil {
			t.Fatal(err)
		}
		var document bson.M
		if err := bson.Unmarshal(data, &document); err != nil {
			t.Fatal(err)
		}
		value, present := document["private"]
		if present != (test.want != nil) || present && value != *test.want {
			t.Fatalf("%s: BSON = %#v", test.payload, document)
		}
	}
}

func boolPointer(value bool) *bool {
	return &value
}

func TestWeeklyScheduleIsActiveAt(t *testing.T) {
	ts := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC) // Wednesday
	seconds := int64(ts.Hour()*3600 + ts.Minute()*60 + ts.Second())

	tests := []struct {
		name     string
		schedule *WeeklySchedule
		want     bool
	}{
		{
			name: "Active",
			schedule: &WeeklySchedule{
				Day:      int(ts.Weekday()),
				Enabled:  true,
				Timezone: "",
				Segments: []DayTimeRange{
					{Start: seconds - 60, End: seconds + 60},
				},
			},
			want: true,
		},
		{
			name: "MismatchedWeekday",
			schedule: &WeeklySchedule{
				Day:      int(ts.Add(24 * time.Hour).Weekday()),
				Enabled:  true,
				Timezone: "",
				Segments: []DayTimeRange{
					{Start: seconds - 60, End: seconds + 60},
				},
			},
			want: false,
		},
		{
			name: "Disabled",
			schedule: &WeeklySchedule{
				Day:      int(ts.Weekday()),
				Enabled:  false,
				Timezone: "",
				Segments: []DayTimeRange{
					{Start: seconds - 60, End: seconds + 60},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.schedule.IsActiveAt(ts); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestDateRangeScheduleIsActiveAt(t *testing.T) {
	ts := time.Date(2026, 2, 4, 10, 30, 0, 0, time.UTC)
	seconds := int64(ts.Hour()*3600 + ts.Minute()*60 + ts.Second())
	unixTs := ts.Unix()

	tests := []struct {
		name string
		dr   *DateRangeSchedule
		want bool
	}{
		{
			name: "Active",
			dr: &DateRangeSchedule{
				StartDate: ts.Add(-time.Hour).Unix(),
				EndDate:   ts.Add(time.Hour).Unix(),
				Enabled:   true,
				Segments: []DayTimeRange{
					{Start: seconds - 120, End: seconds + 120},
				},
			},
			want: true,
		},
		{
			name: "OutsideSegments",
			dr: &DateRangeSchedule{
				StartDate: ts.Add(-time.Hour).Unix(),
				EndDate:   ts.Add(time.Hour).Unix(),
				Enabled:   true,
				Segments:  []DayTimeRange{{Start: seconds + 1, End: seconds + 120}},
			},
			want: false,
		},
		{
			name: "EndExclusiveBoundary",
			dr: &DateRangeSchedule{
				StartDate: ts.Add(-time.Hour).Unix(),
				EndDate:   ts.Add(time.Hour).Unix(),
				Enabled:   true,
				Segments:  []DayTimeRange{{Start: seconds, End: seconds}},
			},
			want: false,
		},
		{
			name: "InvalidSegmentIgnored",
			dr: &DateRangeSchedule{
				StartDate: ts.Add(-time.Hour).Unix(),
				EndDate:   ts.Add(time.Hour).Unix(),
				Enabled:   true,
				Segments:  []DayTimeRange{{Start: seconds + 120, End: seconds - 120}},
			},
			want: false,
		},
		{
			name: "OutsideDateRange",
			dr: &DateRangeSchedule{
				StartDate: ts.Add(2 * time.Hour).Unix(),
				EndDate:   ts.Add(3 * time.Hour).Unix(),
				Enabled:   true,
				Segments:  []DayTimeRange{{Start: seconds - 120, End: seconds + 120}},
			},
			want: false,
		},
		{
			name: "Disabled",
			dr: &DateRangeSchedule{
				StartDate: ts.Add(-time.Hour).Unix(),
				EndDate:   ts.Add(time.Hour).Unix(),
				Enabled:   false,
				Segments:  []DayTimeRange{{Start: seconds - 120, End: seconds + 120}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dr.IsActiveAt(unixTs); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCustomAlertIsScheduledAtDateRangeOverride(t *testing.T) {
	ts := time.Date(2026, 2, 4, 9, 0, 0, 0, time.UTC)
	seconds := int64(ts.Hour()*3600 + ts.Minute()*60 + ts.Second())
	unixTs := ts.Unix()

	tests := []struct {
		name  string
		alert *CustomAlert
		want  bool
	}{
		{
			name: "OverrideInactive",
			alert: &CustomAlert{
				WeeklySchedule: []*WeeklySchedule{
					{
						Day:      int(ts.Weekday()),
						Enabled:  true,
						Segments: []DayTimeRange{{Start: seconds - 60, End: seconds + 60}},
					},
				},
				DateRangeSchedule: []*DateRangeSchedule{
					{
						StartDate: ts.Add(-time.Hour).Unix(),
						EndDate:   ts.Add(time.Hour).Unix(),
						Enabled:   true,
						Segments:  []DayTimeRange{{Start: seconds + 60, End: seconds + 120}},
					},
				},
			},
			want: false,
		},
		{
			name: "OverrideActive",
			alert: &CustomAlert{
				WeeklySchedule: []*WeeklySchedule{
					{
						Day:      int(ts.Weekday()),
						Enabled:  true,
						Segments: []DayTimeRange{{Start: seconds - 60, End: seconds + 60}},
					},
				},
				DateRangeSchedule: []*DateRangeSchedule{
					{
						StartDate: ts.Add(-time.Hour).Unix(),
						EndDate:   ts.Add(time.Hour).Unix(),
						Enabled:   true,
						Segments:  []DayTimeRange{{Start: seconds - 30, End: seconds + 30}},
					},
				},
			},
			want: true,
		},
		{
			name: "WeeklyOnly",
			alert: &CustomAlert{
				WeeklySchedule: []*WeeklySchedule{
					{
						Day:      int(ts.Weekday()),
						Enabled:  true,
						Segments: []DayTimeRange{{Start: seconds - 60, End: seconds + 60}},
					},
				},
			},
			want: true,
		},
		{
			name:  "NoSchedules",
			alert: &CustomAlert{},
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.alert.IsScheduledAt(unixTs); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCustomAlertIsScheduledAtNoSchedules(t *testing.T) {
	alert := &CustomAlert{}
	if !alert.IsScheduledAt(time.Now().Unix()) {
		t.Fatalf("expected alert with no schedules to be active")
	}
}

func TestDateRangeScheduleTimezonePolicy(t *testing.T) {
	fixed := time.FixedZone("TEST-0500", -5*3600)
	localMidnight := time.Date(2026, 2, 4, 0, 0, 0, 0, fixed)
	endMidnight := time.Date(2026, 2, 5, 0, 0, 0, 0, fixed)

	dr := &DateRangeSchedule{
		StartDate: localMidnight.Unix(),
		EndDate:   endMidnight.Unix(),
		Enabled:   true,
		Timezone:  fixed.String(),
		Segments: []DayTimeRange{
			{Start: 0, End: 86400},
		},
	}

	ts := time.Date(2026, 2, 4, 23, 0, 0, 0, fixed)
	if !dr.IsActiveAt(ts.Unix()) {
		t.Fatalf("expected date range schedule to be active within local day")
	}

	ts = time.Date(2026, 2, 5, 0, 0, 0, 0, fixed)
	if dr.IsActiveAt(ts.Unix()) {
		t.Fatalf("expected date range schedule to be inactive at end boundary")
	}
}

func TestDateRangeScheduleIanaTimezone(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Brussels")
	if err != nil {
		t.Skipf("timezone data not available: %v", err)
	}

	localMidnight := time.Date(2026, 2, 4, 0, 0, 0, 0, loc)
	endMidnight := time.Date(2026, 2, 5, 0, 0, 0, 0, loc)

	dr := &DateRangeSchedule{
		StartDate: localMidnight.Unix(),
		EndDate:   endMidnight.Unix(),
		Enabled:   true,
		Timezone:  "Europe/Brussels",
		Segments: []DayTimeRange{
			{Start: 23 * 3600, End: 24 * 3600},
		},
	}

	ts := time.Date(2026, 2, 4, 22, 30, 0, 0, time.UTC) // 23:30 in Brussels
	if !dr.IsActiveAt(ts.Unix()) {
		t.Fatalf("expected schedule to be active for Brussels local time")
	}

	ts = time.Date(2026, 2, 4, 21, 30, 0, 0, time.UTC) // 22:30 in Brussels
	if dr.IsActiveAt(ts.Unix()) {
		t.Fatalf("expected schedule to be inactive outside Brussels segment")
	}
}

func TestWeeklyScheduleFromDeprecatedTimeRanges(t *testing.T) {
	tests := []struct {
		name  string
		alert *CustomAlert
		tz    string
		want  []*WeeklySchedule
	}{
		{
			name: "FullDayBothRanges",
			alert: &CustomAlert{
				TimeRange1Min: 0,
				TimeRange1Max: 24,
				TimeRange2Min: 0,
				TimeRange2Max: 24,
			},
			tz: "",
			want: expectedWeeklySchedules([]DayTimeRange{
				{Start: 0, End: 24 * 3600},
				{Start: 0, End: 24 * 3600},
			}, "Europe/Brussels"),
		},
		{
			name: "CustomRanges",
			alert: &CustomAlert{
				TimeRange1Min: 6,
				TimeRange1Max: 11,
				TimeRange2Min: 12,
				TimeRange2Max: 20,
			},
			tz: "UTC",
			want: expectedWeeklySchedules([]DayTimeRange{
				{Start: 6 * 3600, End: 11 * 3600},
				{Start: 12 * 3600, End: 20 * 3600},
			}, "UTC"),
		},
		{
			name: "InvalidRangesSkipped",
			alert: &CustomAlert{
				TimeRange1Min: -1,
				TimeRange1Max: 10,
				TimeRange2Min: 18,
				TimeRange2Max: 18,
			},
			tz:   "Europe/Brussels",
			want: nil,
		},
		{
			name: "InvalidTimezoneFallsBack",
			alert: &CustomAlert{
				TimeRange1Min: 8,
				TimeRange1Max: 12,
			},
			tz: "Not/AZone",
			want: expectedWeeklySchedules([]DayTimeRange{
				{Start: 8 * 3600, End: 12 * 3600},
			}, "Europe/Brussels"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.alert.WeeklyScheduleFromDeprecatedTimeRanges(tt.tz)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected schedule %+v, got %+v", tt.want, got)
			}
		})
	}
}

func expectedWeeklySchedules(segments []DayTimeRange, timezone string) []*WeeklySchedule {
	if len(segments) == 0 {
		return nil
	}
	schedules := make([]*WeeklySchedule, 0, 7)
	for day := 0; day < 7; day++ {
		schedules = append(schedules, &WeeklySchedule{
			Day:      day,
			Segments: append([]DayTimeRange(nil), segments...),
			Enabled:  true,
			Timezone: timezone,
		})
	}
	return schedules
}
