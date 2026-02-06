package models

import (
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomAlert struct {
	Id                  primitive.ObjectID   `json:"id" bson:"_id,omitempty"`
	Enabled             bool                 `json:"enabled" bson:"enabled"`
	Title               string               `json:"title" bson:"title,omitempty"`
	Description         string               `json:"description" bson:"description,omitempty"`
	ChannelsAll         bool                 `json:"channelsAll" bson:"channelsAll"`
	ChannelsList        []string             `json:"channelsList" bson:"channelsList"`
	DevicesAll          bool                 `json:"devicesAll" bson:"devicesAll"`
	DevicesList         []DeviceKey          `json:"devicesList" bson:"devicesList"`
	CountingDevicesAll  bool                 `json:"countingDevicesAll" bson:"countingDevicesAll"`
	CountingDevicesList []DeviceKey          `json:"countingDevicesList" bson:"countingDevicesList"`
	ClassificationAll   bool                 `json:"classificationAll" bson:"classificationAll"`
	ClassificationList  []string             `json:"classificationList" bson:"classificationList"`
	TimeAdvanced        bool                 `json:"timeAdvanced" bson:"timeAdvanced"`
	UserId              string               `json:"user_id" bson:"user_id,omitempty"`
	MasterUserId        string               `json:"master_user_id" bson:"master_user_id,omitempty"`
	EmailEmail          string               `json:"email_email" bson:"email_email,omitempty"`
	SlackHook           string               `json:"slack_hook" bson:"slack_hook,omitempty"`
	SlackBotname        string               `json:"slack_botname" bson:"slack_botname,omitempty"`
	PushbulletApikey    string               `json:"pushbullet_apikey" bson:"pushbullet_apikey,omitempty"`
	TelegramToken       string               `json:"telegram_token" bson:"telegram_token,omitempty"`
	TelegramChannel     string               `json:"telegram_channel" bson:"telegram_channel,omitempty"`
	AlexaToken          string               `json:"alexa_token" bson:"alexa_token,omitempty"`
	WebhookUrl          string               `json:"webhook_url" bson:"webhook_url,omitempty"`
	IftttToken          string               `json:"ifttt_token" bson:"ifttt_token,omitempty"`
	SMSAccountsid       string               `json:"sms_accountsid" bson:"sms_accountsid,omitempty"`
	SMSAuthtoken        string               `json:"sms_authtoken" bson:"sms_authtoken,omitempty"`
	SMSTelfrom          string               `json:"sms_telfrom" bson:"sms_telfrom,omitempty"`
	SMSTelto            string               `json:"sms_telto" bson:"sms_telto,omitempty"`
	PushoverApikey      string               `json:"pushover_apikey" bson:"pushover_apikey,omitempty"`
	PushoverSendto      string               `json:"pushover_sendto" bson:"pushover_sendto,omitempty"`
	MotionRegions       []Region             `json:"motionRegions" bson:"motionRegions,omitempty"`
	CountingLines       []Region             `json:"countingLines" bson:"countingLines,omitempty"`
	InputList           []string             `json:"inputList" bson:"inputList,omitempty"`
	InputsAND           bool                 `json:"inputsAND" bson:"inputsAND"`
	OutputList          []string             `json:"outputList" bson:"outputList,omitempty"`
	Features            *AlertFeatures       `json:"features,omitempty" bson:"features,omitempty"`
	WeeklySchedule      []*WeeklySchedule    `json:"weeklySchedule" bson:"weeklySchedule,omitempty"`
	DateRangeSchedule   []*DateRangeSchedule `json:"dateRangeSchedule" bson:"dateRangeSchedule,omitempty"`
	OrganisationId      string               `json:"organisationId" bson:"organisationId,omitempty"`

	// Deprecated: legacy time range fields. Use WeeklySchedule/DateRangeSchedule instead.
	TimeRange1Max int32 `json:"timeRange1Max" bson:"timeRange1Max"`
	TimeRange1Min int32 `json:"timeRange1Min" bson:"timeRange1Min"`
	TimeRange2Max int32 `json:"timeRange2Max" bson:"timeRange2Max"`
	TimeRange2Min int32 `json:"timeRange2Min" bson:"timeRange2Min"`
}

type AlertFeatures struct {
	CreateMarker bool `json:"createMarker,omitempty" bson:"createMarker,omitempty"`
}

type WeeklySchedule struct {
	Day      int            `json:"day" bson:"day"`
	Segments []DayTimeRange `json:"segments" bson:"segments"`
	Enabled  bool           `json:"enabled" bson:"enabled"`
	Timezone string         `json:"timezone" bson:"timezone"`
}

type DateRangeSchedule struct {
	// StartDate is inclusive; EndDate is exclusive. Both are unix seconds for local midnight in Timezone.
	StartDate int64          `json:"startDate" bson:"startDate"`
	EndDate   int64          `json:"endDate" bson:"endDate"`
	Segments  []DayTimeRange `json:"segments" bson:"segments"`
	Enabled   bool           `json:"enabled" bson:"enabled"`
	Timezone  string         `json:"timezone" bson:"timezone"`
}

type DayTimeRange struct {
	Start int64 `json:"start" bson:"start"` // seconds since midnight, 0..86400
	End   int64 `json:"end" bson:"end"`
}

type AlertPatch struct {
	Title               *string           `json:"title,omitempty" bson:"title,omitempty"`
	Enabled             *bool             `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Description         *string           `json:"description,omitempty" bson:"description,omitempty"`
	TimeRange1Min       *int32            `json:"timeRange1Min,omitempty" bson:"timeRange1Min,omitempty"`
	TimeRange1Max       *int32            `json:"timeRange1Max,omitempty" bson:"timeRange1Max,omitempty"`
	TimeRange2Min       *int32            `json:"timeRange2Min,omitempty" bson:"timeRange2Min,omitempty"`
	TimeRange2Max       *int32            `json:"timeRange2Max,omitempty" bson:"timeRange2Max,omitempty"`
	TimeAdvanced        *bool             `json:"timeAdvanced,omitempty" bson:"timeAdvanced,omitempty"`
	WeeklySchedule      []*WeeklySchedule `json:"weeklySchedule,omitempty" bson:"weeklySchedule,omitempty"`
	ChannelsAll         *bool             `json:"channelsAll,omitempty" bson:"channelsAll,omitempty"`
	ChannelsList        []string          `json:"channelsList,omitempty" bson:"channelsList,omitempty"`
	DevicesAll          *bool             `json:"devicesAll,omitempty" bson:"devicesAll,omitempty"`
	DevicesList         []DeviceKey       `json:"devicesList,omitempty" bson:"devicesList,omitempty"`
	CountingDevicesAll  *bool             `json:"countingDevicesAll,omitempty" bson:"countingDevicesAll,omitempty"`
	CountingDevicesList []DeviceKey       `json:"countingDevicesList,omitempty" bson:"countingDevicesList,omitempty"`
	ClassificationAll   *bool             `json:"classificationAll,omitempty" bson:"classificationAll,omitempty"`
	ClassificationList  []string          `json:"classificationList,omitempty" bson:"classificationList,omitempty"`
	MotionRegions       []Region          `json:"motionRegions,omitempty" bson:"motionRegions,omitempty"`
	CountingRegions     []Region          `json:"countingRegions,omitempty" bson:"countingRegions,omitempty"`
	CountingLines       []Region          `json:"countingLines,omitempty" bson:"countingLines,omitempty"`
	InputList           []string          `json:"inputList,omitempty" bson:"inputList,omitempty"`
	OutputList          []string          `json:"outputList,omitempty" bson:"outputList,omitempty"`
	InputsAND           *bool             `json:"inputsAND,omitempty" bson:"inputsAND,omitempty"`
	EmailEmail          *string           `json:"email_email,omitempty" bson:"email_email,omitempty"`
	SlackHook           *string           `json:"slack_hook,omitempty" bson:"slack_hook,omitempty"`
	SlackBotname        *string           `json:"slack_botname,omitempty" bson:"slack_botname,omitempty"`
	PushbulletApikey    *string           `json:"pushbullet_apikey,omitempty" bson:"pushbullet_apikey,omitempty"`
	TelegramToken       *string           `json:"telegram_token,omitempty" bson:"telegram_token,omitempty"`
	TelegramChannel     *string           `json:"telegram_channel,omitempty" bson:"telegram_channel,omitempty"`
	AlexaToken          *string           `json:"alexa_token,omitempty" bson:"alexa_token,omitempty"`
	WebhookUrl          *string           `json:"webhook_url,omitempty" bson:"webhook_url,omitempty"`
	IftttToken          *string           `json:"ifttt_token,omitempty" bson:"ifttt_token,omitempty"`
	SMSAccountsid       *string           `json:"sms_accountsid,omitempty" bson:"sms_accountsid,omitempty"`
	SMSAuthtoken        *string           `json:"sms_authtoken,omitempty" bson:"sms_authtoken,omitempty"`
	SMSTelfrom          *string           `json:"sms_telfrom,omitempty" bson:"sms_telfrom,omitempty"`
	SMSTelto            *string           `json:"sms_telto,omitempty" bson:"sms_telto,omitempty"`
	PushoverApikey      *string           `json:"pushover_apikey,omitempty" bson:"pushover_apikey,omitempty"`
	PushoverSendto      *string           `json:"pushover_sendto,omitempty" bson:"pushover_sendto,omitempty"`
}

type GetAlertsInput struct {
	User User `json:"user" bson:"user"`
}

type GetAlertsOutput struct {
	Alerts []CustomAlert `json:"alerts" bson:"alerts"`
}

type CreateAlertInput struct {
	User  User        `json:"user" bson:"user"`
	Alert CustomAlert `json:"alert" bson:"alert"`
}

type CreateAlertOutput struct {
	Alert *CustomAlert `json:"alert" bson:"alert"`
}

type UpdateAlertInput struct {
	User       User        `json:"user" bson:"user"`
	AlertId    string      `json:"alertId" bson:"alertId"`
	AlertPatch *AlertPatch `json:"alertPatch" bson:"alertPatch"`
}

type UpdateAlertOutput struct {
	Alert *CustomAlert `json:"alert" bson:"alert"`
}

type RemoveAlertInput struct {
	User    User   `json:"user" bson:"user"`
	AlertId string `json:"alertId" bson:"alertId"`
}

type RemoveAlertOutput struct{}

var scheduleTzCache sync.Map

// IsScheduledAt reports whether unixTs falls within the alert's schedules.
// Date ranges take precedence when the date is within any configured range.
func (a *CustomAlert) IsScheduledAt(unixTs int64) bool {
	ts := time.Unix(unixTs, 0)
	hasWeekly := len(a.WeeklySchedule) > 0
	hasDateRange := len(a.DateRangeSchedule) > 0

	if hasDateRange {
		activeRange := false
		for _, dr := range a.DateRangeSchedule {
			if dr == nil || !dr.Enabled {
				continue
			}
			if dr.DateInRange(unixTs) {
				activeRange = true
				if dr.IsActiveAt(unixTs) {
					return true
				}
			}
		}
		if activeRange {
			return false
		}
	}

	if hasWeekly {
		for _, ws := range a.WeeklySchedule {
			if ws == nil || !ws.Enabled {
				continue
			}
			if ws.IsActiveAt(ts) {
				return true
			}
		}
		return false
	}

	return !hasDateRange
}

// IsActiveAt reports whether ts falls within any enabled weekly segment for the given day.
func (w *WeeklySchedule) IsActiveAt(ts time.Time) bool {
	if w == nil || !w.Enabled {
		return false
	}
	loc := scheduleLocation(w.Timezone, ts)
	tsInLoc := ts.In(loc)
	if int(tsInLoc.Weekday()) != w.Day {
		return false
	}
	return inSegments(secondsOfDay(tsInLoc), w.Segments)
}

// DateInRange reports whether unixTs is within the range.
// EndDate is treated as an exclusive boundary to ensure the range
// covers full local days from StartDate up to, but not including, EndDate.
func (d *DateRangeSchedule) DateInRange(unixTs int64) bool {
	if d == nil || !d.Enabled {
		return false
	}
	return unixTs >= d.StartDate && unixTs < d.EndDate
}

// IsActiveAt reports whether unixTs falls within any enabled date-range segment.
func (d *DateRangeSchedule) IsActiveAt(unixTs int64) bool {
	if d == nil || !d.Enabled {
		return false
	}
	if !d.DateInRange(unixTs) {
		return false
	}
	ts := time.Unix(unixTs, 0)
	loc := scheduleLocation(d.Timezone, ts)
	tsInLoc := ts.In(loc)
	return inSegments(secondsOfDay(tsInLoc), d.Segments)
}

func scheduleLocation(timezone string, fallback time.Time) *time.Location {
	if timezone == "" {
		return fallback.Location()
	}
	if cached, ok := scheduleTzCache.Load(timezone); ok {
		if loc, ok := cached.(*time.Location); ok {
			return loc
		}
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return fallback.Location()
	}
	scheduleTzCache.Store(timezone, loc)
	return loc
}

func secondsOfDay(t time.Time) int {
	return t.Hour()*3600 + t.Minute()*60 + t.Second()
}

func inSegments(seconds int, segments []DayTimeRange) bool {
	for _, seg := range segments {
		if !isValidSegment(seg) {
			continue
		}
		if seg.Start <= int64(seconds) && int64(seconds) < seg.End {
			return true
		}
	}
	return false
}

func isValidSegment(seg DayTimeRange) bool {
	return seg.Start >= 0 && seg.End <= 86400 && seg.Start < seg.End
}

// WeeklyScheduleFromDeprecatedTimeRanges maps legacy TimeRange values to a weekly schedule.
// Each time range becomes a daily segment for every weekday, enabled in Europe/Brussels.
func (a *CustomAlert) WeeklyScheduleFromDeprecatedTimeRanges() []*WeeklySchedule {
	if a == nil {
		return nil
	}
	segments := deprecatedTimeRangeSegments(a.TimeRange1Min, a.TimeRange1Max, a.TimeRange2Min, a.TimeRange2Max)
	if len(segments) == 0 {
		return nil
	}
	schedules := make([]*WeeklySchedule, 0, 7)
	for day := 0; day < 7; day++ {
		schedules = append(schedules, &WeeklySchedule{
			Day:      day,
			Segments: append([]DayTimeRange(nil), segments...),
			Enabled:  true,
			Timezone: "Europe/Brussels",
		})
	}
	return schedules
}

func deprecatedTimeRangeSegments(min1, max1, min2, max2 int32) []DayTimeRange {
	segments := make([]DayTimeRange, 0, 2)
	if seg, ok := toDayTimeRange(min1, max1); ok {
		segments = append(segments, seg)
	}
	if seg, ok := toDayTimeRange(min2, max2); ok {
		segments = append(segments, seg)
	}
	return segments
}

func toDayTimeRange(minHour, maxHour int32) (DayTimeRange, bool) {
	if minHour < 0 || maxHour > 24 || minHour >= maxHour {
		return DayTimeRange{}, false
	}
	return DayTimeRange{
		Start: int64(minHour) * 3600,
		End:   int64(maxHour) * 3600,
	}, true
}
