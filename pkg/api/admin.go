package api

// AdminStatus represents specific status codes for admin dashboard operations.
type AdminStatus string

const (
	AdminStatsSuccess AdminStatus = "admin_stats_success"
	AdminStatsFailed  AdminStatus = "admin_stats_failed"
)

// String returns the string representation of the admin status.
func (s AdminStatus) String() string {
	return string(s)
}

// Translate returns the translated string representation of the admin status
// in the specified language.
func (s AdminStatus) Translate(lang string) string {
	translations := map[string]map[AdminStatus]string{
		"en": {
			AdminStatsSuccess: "Admin stats retrieved successfully",
			AdminStatsFailed:  "Admin stats retrieval failed",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[s]; exists {
			return translation
		}
	}

	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[s]; exists {
			return translation
		}
	}

	return s.String()
}

// ----------------------------------------------------------------------------
// Admin response wrappers
//
// These envelopes are shared by the hub-api admin endpoints (versioned under
// /admin/v20260101). They embed the canonical SuccessResponse / ErrorResponse
// envelope so the wire format matches device / task / alert endpoints.
//
// Entity-typed payload fields use `any` for now because several of the
// underlying entity types (DeviceShort, UserSubscription, SiteShort,
// OrganisationType, OrganisationRole, SubscriptionPlan, …) still live in
// hub-api/api/models. Once those entities migrate to models/pkg/models the
// `any` fields can be tightened to their concrete types without changing the
// wire format.
// ----------------------------------------------------------------------------

// ---------- Stats ----------

// AdminRecentSubscription is a lightweight projection of a user subscription
// enriched with the owning user's display name (used by the admin dashboard).
//
// The fields use `any` for values whose representation differs across the
// underlying entities (UserSubscription vs SubscriptionPlan). Once those
// entities migrate to models/pkg/models the types can be tightened.
type AdminRecentSubscription struct {
	Id           any    `json:"id"`
	Name         string `json:"name"`
	UserId       string `json:"user_id"`
	StripePlan   string `json:"stripe_plan"`
	StripeActive any    `json:"stripe_active"`
	Quantity     any    `json:"quantity"`
	EndsAt       any    `json:"ends_at"`
	CreatedAt    any    `json:"created_at"`
	Username     string `json:"username"`
}

type GetAdminStatsResponse struct {
	DeviceCount         int64                     `json:"device_count"`
	SubscriptionCount   int64                     `json:"subscription_count"`
	RecordingCount      int64                     `json:"recording_count"`
	UserCount           int64                     `json:"user_count"`
	OrganisationCount   int64                     `json:"organisation_count"`
	RecentUsers         any                       `json:"recent_users"`
	RecentOrganisations any                       `json:"recent_organisations"`
	RecentSubscriptions []AdminRecentSubscription `json:"recent_subscriptions"`
}

type GetAdminStatsSuccessResponse struct {
	SuccessResponse
	Data GetAdminStatsResponse `json:"data,omitempty"`
}

type GetAdminStatsErrorResponse struct {
	ErrorResponse
}

// ---------- Organisations (admin variant) ----------

type GetAdminOrganisationsMeta struct {
	TotalOrganisationCount int64 `json:"total_organisation_count"`
	OrganisationCount      int64 `json:"organisation_count"`
	TotalPages             int   `json:"total_pages"`
	OrganisationTypes      any   `json:"organisation_types"`
	OrganisationRoles      any   `json:"organisation_roles"`
}

type GetAdminOrganisationsResponse struct {
	Organisations any                       `json:"organisations"`
	Meta          GetAdminOrganisationsMeta `json:"meta"`
}

type GetAdminOrganisationsSuccessResponse struct {
	SuccessResponse
	Data GetAdminOrganisationsResponse `json:"data,omitempty"`
}

type GetAdminOrganisationsErrorResponse struct {
	ErrorResponse
}

type CreateAdminOrganisationResponse struct {
	Organisation any `json:"organisation"`
}

type CreateAdminOrganisationSuccessResponse struct {
	SuccessResponse
	Data CreateAdminOrganisationResponse `json:"data,omitempty"`
}

type CreateAdminOrganisationErrorResponse struct {
	ErrorResponse
}

// ---------- Users (admin variants) ----------

type CreateAdminUserResponse struct {
	User any `json:"user"`
}

type CreateAdminUserSuccessResponse struct {
	SuccessResponse
	Data CreateAdminUserResponse `json:"data,omitempty"`
}

type CreateAdminUserErrorResponse struct {
	ErrorResponse
}

type GetAdminUsersMeta struct {
	TotalUserCount int64 `json:"total_user_count"`
	UserCount      int64 `json:"user_count"`
	TotalPages     int   `json:"total_pages"`
}

type GetAdminUsersResponse struct {
	Users any               `json:"users"`
	Meta  GetAdminUsersMeta `json:"meta"`
}

type GetAdminUsersSuccessResponse struct {
	SuccessResponse
	Data GetAdminUsersResponse `json:"data,omitempty"`
}

type GetAdminUsersErrorResponse struct {
	ErrorResponse
}

type GetAdminUserProfileResponse struct {
	User any `json:"user"`
}

type GetAdminUserProfileSuccessResponse struct {
	SuccessResponse
	Data GetAdminUserProfileResponse `json:"data,omitempty"`
}

type GetAdminUserProfileErrorResponse struct {
	ErrorResponse
}

type GenerateAdminUserKeyResponse struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key,omitempty"`
}

type GenerateAdminUserKeySuccessResponse struct {
	SuccessResponse
	Data GenerateAdminUserKeyResponse `json:"data,omitempty"`
}

type GenerateAdminUserKeyErrorResponse struct {
	ErrorResponse
}

type UpdateAdminUserResponse struct {
	User any `json:"user"`
}

type UpdateAdminUserSuccessResponse struct {
	SuccessResponse
	Data UpdateAdminUserResponse `json:"data,omitempty"`
}

type UpdateAdminUserErrorResponse struct {
	ErrorResponse
}

type UpdateAdminUserPasswordSuccessResponse struct {
	SuccessResponse
}

type UpdateAdminUserPasswordErrorResponse struct {
	ErrorResponse
}

// ---------- User devices (admin variants) ----------

type GetAdminUserDevicesMeta struct {
	TotalDeviceCount int `json:"total_device_count"`
	DeviceCount      int `json:"device_count"`
	TotalPages       int `json:"total_pages"`
}

type GetAdminUserDevicesInformationResponse struct {
	Devices any                     `json:"devices"`
	Meta    GetAdminUserDevicesMeta `json:"meta"`
}

type GetAdminUserDevicesInformationSuccessResponse struct {
	SuccessResponse
	Data GetAdminUserDevicesInformationResponse `json:"data,omitempty"`
}

type GetAdminUserDevicesInformationErrorResponse struct {
	ErrorResponse
}

type GetAdminUserDevicesResponse struct {
	Devices any `json:"devices"`
}

type GetAdminUserDevicesSuccessResponse struct {
	SuccessResponse
	Data GetAdminUserDevicesResponse `json:"data,omitempty"`
}

type GetAdminUserDevicesErrorResponse struct {
	ErrorResponse
}

// ---------- Subscriptions (admin variants) ----------

type GetAdminSubscriptionSettingsResponse struct {
	Settings any `json:"settings"`
}

type GetAdminSubscriptionSettingsSuccessResponse struct {
	SuccessResponse
	Data GetAdminSubscriptionSettingsResponse `json:"data,omitempty"`
}

type GetAdminSubscriptionSettingsErrorResponse struct {
	ErrorResponse
}

type GetAdminUserSubscriptionResponse struct {
	Subscription any `json:"subscription"`
}

type GetAdminUserSubscriptionSuccessResponse struct {
	SuccessResponse
	Data GetAdminUserSubscriptionResponse `json:"data,omitempty"`
}

type GetAdminUserSubscriptionErrorResponse struct {
	ErrorResponse
}

type UpdateAdminUserSubscriptionResponse struct {
	Subscription any `json:"subscription"`
}

type UpdateAdminUserSubscriptionSuccessResponse struct {
	SuccessResponse
	Data UpdateAdminUserSubscriptionResponse `json:"data,omitempty"`
}

type UpdateAdminUserSubscriptionErrorResponse struct {
	ErrorResponse
}
