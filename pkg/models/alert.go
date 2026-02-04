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
	// StartDate/EndDate are unix seconds for local midnight in Timezone (inclusive bounds).
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

var scheduleTzCache sync.Map

// IsScheduledAt reports whether ts falls within the alert's schedules.
// Date ranges take precedence when the date is within any configured range.
func (a *CustomAlert) IsScheduledAt(ts time.Time) bool {
	hasWeekly := len(a.WeeklySchedule) > 0
	hasDateRange := len(a.DateRangeSchedule) > 0

	if hasDateRange {
		activeRange := false
		for _, dr := range a.DateRangeSchedule {
			if dr == nil || !dr.Enabled {
				continue
			}
			if dr.DateInRange(ts) {
				activeRange = true
				if dr.IsActiveAt(ts) {
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

// DateInRange reports whether ts (as a unix second) is within the inclusive range.
func (d *DateRangeSchedule) DateInRange(ts time.Time) bool {
	if d == nil || !d.Enabled {
		return false
	}
	tsUnix := ts.Unix()
	return tsUnix >= d.StartDate && tsUnix <= d.EndDate
}

// IsActiveAt reports whether ts falls within any enabled date-range segment.
func (d *DateRangeSchedule) IsActiveAt(ts time.Time) bool {
	if d == nil || !d.Enabled {
		return false
	}
	if !d.DateInRange(ts) {
		return false
	}
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
