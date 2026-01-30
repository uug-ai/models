package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	USER_FOUND     string = "One or more users were found"
	USER_NOT_FOUND string = "One or more users not found, returning empty list"
)

type UserProfile struct {
	User User `json:"user" bson:"user,omitempty"`
	Card Card `json:"card" bson:"card,omitempty"`
}

func GetUserIdFromAccountOrMaster(user User) string {
	if user.MasterAccount != "" {
		return user.Master.Id.Hex()
	}
	return user.Id.Hex()
}

func GetOrganisationId(user User) string {
	if user.MasterAccount != "" {
		return user.Master.Id.Hex()
	}
	return user.Id.Hex()
}

type User struct {
	Id       primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Email    string             `json:"email" bson:"email,omitempty"`
	Username string             `json:"username" bson:"username,omitempty"`
	Password string             `json:"password" bson:"password,omitempty"`

	// A User can belong to multiple organisations, and depending on the organisation he/she
	// might have different roleassignments. A user also has its own personal organisation id.
	PersonalOrganisationId string   `json:"personalOrganisationId" bson:"personalOrganisationId,omitempty"`
	OrganisationIds        []string `json:"organisationIds" bson:"organisationIds,omitempty"`

	// Domain information
	Domain string `json:"domain" bson:"domain,omitempty"`

	// Registration token
	RegistrationToken string `json:"registerToken" bson:"registerToken,omitempty"`

	// Forgot password
	ForgotPassword string `json:"forgotPassword" bson:"forgotPassword,omitempty"`

	// Roles and permissions
	Role        string      `json:"role" bson:"role,omitempty"`
	CustomRole  string      `json:"custom_role" bson:"custom_role,omitempty"`
	RoleLevel   int         `json:"role_level" bson:"role_level,omitempty"`
	Permissions Permissions `json:"permissions" bson:"permissions,omitempty"`

	Days     []string `json:"dates" bson:"dates,omitempty"`
	Timezone string   `json:"timezone" bson:"timezone,omitempty"`

	// MFA settings
	GoogleMFASecret  string `json:"google2fa_secret" bson:"google2fa_secret,omitempty"`
	GoogleMFAEnabled bool   `json:"google2fa_enabled" bson:"google2fa_enabled,omitempty"`
	Mfa              bool   `json:"mfa" bson:"mfa,omitempty"`
	ForceMFA         int    `json:"force_mfa" bson:"force_mfa"`

	// If user reached his limit
	ReachedLimit          bool  `json:"reachedLimit" bson:"reachedLimit,omitempty"`
	ReachedLimitTimestamp int64 `json:"reachedLimitTimestamp" bson:"reachedLimitTimestamp,omitempty"`

	// Personal information
	Nickname  string `json:"nickname" bson:"nickname,omitempty"`
	FirstName string `json:"firstname" bson:"firstname,omitempty"`
	LastName  string `json:"lastname" bson:"lastname,omitempty"`
	Address   string `json:"address" bson:"address,omitempty"`

	// Company information
	CompanyName         string `json:"company_name" bson:"company_name,omitempty"`
	CompanyNumber       string `json:"company_number" bson:"company_number,omitempty"`
	CompanyStreetNumber string `json:"company_street_number" bson:"company_street_number,omitempty"`
	CompanyStreet       string `json:"company_street" bson:"company_street,omitempty"`
	CompanyCity         string `json:"company_city" bson:"company_city,omitempty"`
	CompanyPostalCode   string `json:"company_postal" bson:"company_postal,omitempty"`
	CompanyRegion       string `json:"company_region" bson:"company_region,omitempty"`
	CompanyCountry      string `json:"company_country" bson:"company_country,omitempty"`
	CompanyCountryLong  string `json:"company_country_long" bson:"company_country_long,omitempty"`

	// Checks
	ProfileCompleted bool `json:"profileCompleted" bson:"profileCompleted,omitempty"`
	IsActive         int  `json:"isActive" bson:"isActive"`

	// Additional fields
	Livestream           Livestream           `json:"livestream" bson:"livestream,omitempty"`
	Devices              []Device             `json:"devices" bson:"devices,omitempty"`
	CreatedAt            time.Time            `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt            time.Time            `json:"updated_at" bson:"updated_at,omitempty"`
	Subscription         Subscription         `json:"subscription" bson:"subscription,omitempty"`
	Storage              Storage              `json:"storage" bson:"storage,omitempty"`
	ArchiveStorage       Storage              `json:"archive_storage" bson:"archive_storage,omitempty"`
	NotificationSettings NotificationSettings `json:"notificationSettings" bson:"notificationSettings,omitempty"`
	Channels             Channels             `json:"channels" bson:"channels,omitempty"`
	Activity             []Activity           `json:"activity" bson:"activity,omitempty"`
	Sites                []string             `json:"sites" bson:"sites,omitempty"`
	Groups               []string             `json:"groups" bson:"groups,omitempty"`
	Cameras              []string             `json:"cameras" bson:"cameras,omitempty"`
	CameraBrands         Settings             `json:"camera_brands" bson:"camera_brands"`
	ClassificationList   Settings             `json:"classification_list" bson:"classification_list"`
	ProfileSettings      ProfileSettings      `json:"profileSettings" bson:"profileSettings,omitempty"`

	// We can override the subscription settings if needed.
	CustomUsageLimit    int `json:"custom_usage_limit" bson:"custom_usage_limit,omitempty"`
	CustomDayLimit      int `json:"custom_day_limit" bson:"custom_day_limit,omitempty"`
	CustomAnalysisLimit int `json:"custom_analysis_limit" bson:"custom_analysis_limit,omitempty"`

	// Subscription - Credentials
	Plan       string `json:"plan" bson:"plan,omitempty"`
	PublicKey  string `json:"amazon_access_key_id" bson:"amazon_access_key_id,omitempty"`
	PrivateKey string `json:"amazon_secret_access_key" bson:"amazon_secret_access_key,omitempty"`
	Bucket     string `json:"bucket" bson:"bucket,omitempty"`
	Region     string `json:"region" bson:"region,omitempty"`

	// Master account to which this account was added.
	MasterAccount string `json:"user_id" bson:"user_id,omitempty"`
	Master        *User  `json:"master" bson:"master,omitempty"`

	// Should go away into Card struct
	StripeId          string   `json:"stripe_id" bson:"stripe_id,omitempty"`
	Coupons           []string `json:"coupons" bson:"coupons"`
	CardBrand         string   `json:"card_brand" bson:"card_brand,omitempty"`
	CardLastFour      string   `json:"card_last_four" bson:"card_last_four,omitempty"`
	CardStatus        string   `json:"card_status" bson:"card_status,omitempty"`
	CardStatusMessage string   `json:"card_status_message" bson:"card_status_message,omitempty"`

	// Encryption
	Encryption Encryption `json:"encryption" bson:"encryption,omitempty"`

	// Additional metadata
	Metadata *UserMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`

	// AtRuntimeMetadata contains metadata that is generated at runtime, which can include
	// more verbose information about the users information, something that doesnt need to be stored in the database,
	// for example the roles assignments, role permissions
	AtRuntimeMetadata *UserAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`

	// Audit information
	Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"` // Audit information for tracking changes to the marker
}

func (user *User) WithTimezone() {
	timezone := user.Timezone
	if timezone == "" {
		user.Timezone = "Europe/Brussels"
	}
}

type UserMetadata struct {
}

type UserAtRuntimeMetadata struct {
	Assignments []RoleAssignment `json:"assignments" bson:"assignments,omitempty"`
	Roles       []Role           `json:"roles" bson:"roles,omitempty"` // Populated role detail
	Devices     []Device         `json:"devices" bson:"devices,omitempty"`
	Groups      []Group          `json:"groups" bson:"groups,omitempty"`
}

type UserProfileSettings struct {
	Username            string          `json:"username" bson:"username,omitempty"`
	Domain              string          `json:"domain" bson:"domain,omitempty"`
	RegistrationToken   string          `json:"registerToken" bson:"registerToken,omitempty"`
	Nickname            string          `json:"nickname" bson:"nickname,omitempty"`
	FirstName           string          `json:"firstname" bson:"firstname,omitempty"`
	LastName            string          `json:"lastname" bson:"lastname,omitempty"`
	CompanyName         string          `json:"company_name" bson:"company_name,omitempty"`
	CompanyNumber       string          `json:"company_number" bson:"company_number,omitempty"`
	Address             string          `json:"address" bson:"address,omitempty"`
	CompanyStreetNumber string          `json:"company_street_number" bson:"company_street_number,omitempty"`
	CompanyStreet       string          `json:"company_street" bson:"company_street,omitempty"`
	CompanyCity         string          `json:"company_city" bson:"company_city,omitempty"`
	CompanyPostalCode   string          `json:"company_postal" bson:"company_postal,omitempty"`
	CompanyRegion       string          `json:"company_region" bson:"company_region,omitempty"`
	CompanyCountry      string          `json:"company_country" bson:"company_country,omitempty"`
	CompanyCountryLong  string          `json:"company_country_long" bson:"company_country_long,omitempty"`
	Timezone            string          `json:"timezone" bson:"timezone,omitempty"`
	ProfileCompleted    bool            `json:"profileCompleted" bson:"profileCompleted,omitempty"`
	IsActive            int             `json:"isActive" bson:"isActive"`
	ForceMFA            int             `json:"force_mfa" bson:"force_mfa"`
	ProfileSettings     ProfileSettings `json:"profileSettings" bson:"profileSettings,omitempty"`
}

type UserSettings struct {
	HLSCallbackURL          string `json:"hls_callback_url" bson:"hls_callback_url,omitempty"`
	OAuthClientID           string `json:"oauth_client_id" bson:"oauth_client_id,omitempty"`
	OAuthClientSecret       string `json:"oauth_client_secret" bson:"oauth_client_secret,omitempty"`
	OAuthClientName         string `json:"oauth_client_name" bson:"oauth_client_name,omitempty"`
	OAuthClientCreationDate int64  `json:"oauth_client_creation_date" bson:"oauth_client_creation_date,omitempty"`
}

// HubCredentials is included in the API documentation as a schema example.
// @name HubCredentials
type HubCredentials struct {
	PublicKey  string `json:"amazon_access_key_id" bson:"amazon_access_key_id,omitempty"`
	PrivateKey string `json:"amazon_secret_access_key" bson:"amazon_secret_access_key,omitempty"`
	Bucket     string `json:"bucket" bson:"bucket,omitempty"`
	Directory  string `json:"directory" bson:"directory,omitempty"`
	Region     string `json:"region" bson:"region,omitempty"`
	Active     bool   `json:"active" bson:"active,omitempty"`
}

type Card struct {
	StripeId          string `json:"stripe_id" bson:"stripe_id,omitempty"`
	CardBrand         string `json:"card_brand" bson:"card_brand,omitempty"`
	CardLastFour      string `json:"card_last_four" bson:"card_last_four,omitempty"`
	CardStatus        string `json:"card_status" bson:"card_status,omitempty"`
	CardStatusMessage string `json:"card_status_message" bson:"card_status_message,omitempty"`
}

type Account struct {
	Account AccountBody `json:"account" bson:"account,omitempty"`
}

type AccountBody struct {
	Domain    string   `json:"domain" bson:"domain,omitempty"`
	FirstName string   `json:"firstname" bson:"firstname,omitempty"`
	LastName  string   `json:"lastname" bson:"lastname,omitempty"`
	Username  string   `json:"username" bson:"username,omitempty"`
	Email     string   `json:"email" bson:"email,omitempty"`
	Password  string   `json:"password" bson:"password,omitempty"`
	Role      string   `json:"role" bson:"role,omitempty"`
	Sites     []string `json:"sites" bson:"sites"`
	Groups    []string `json:"groups" bson:"groups"`
	Cameras   []string `json:"cameras" bson:"cameras"`
	IsActive  int      `json:"isActive" bson:"isActive"`
	ForceMFA  int      `json:"force_mfa" bson:"force_mfa"`
}

type Credentials struct {
	CurrentPassword   string `json:"currentPassword" bson:"username,currentPassword"`
	NewPassword       string `json:"newPassword" bson:"password,newPassword"`
	NewPasswordRepeat string `json:"newPasswordRepeat" bson:"role,newPasswordRepeat"`
}

type Encryption struct {
	Enabled              bool   `json:"enabled" bson:"enabled"`
	HasPassphrase        bool   `json:"has_passphrase" bson:"has_passphrase,omitempty"`
	Fingerprint          string `json:"fingerprint" bson:"fingerprint,omitempty"`
	FingerprintEncrypted string `json:"fingerprint_encrypted" bson:"fingerprint_encrypted,omitempty"`
	PublicKey            string `json:"public_key" bson:"public_key,omitempty"`
	SymmetricKey         string `json:"symmetric_key" bson:"symmetric_key,omitempty"`
	CreationDate         int64  `json:"creation_date" bson:"creation_date,omitempty"`
}

type KeyPair struct {
	PublicKey  string `json:"amazon_access_key_id" bson:"amazon_access_key_id,omitempty"`
	PrivateKey string `json:"amazon_secret_access_key" bson:"amazon_secret_access_key,omitempty"`
}

type UserShort struct {
	Id               primitive.ObjectID `json:"id" bson:"_id,omitempty,omitempty"`
	Username         string             `json:"username" bson:"username,omitempty"`
	Password         string             `json:"password" bson:"password,omitempty"`
	Email            string             `json:"email" bson:"email,omitempty"`
	FirstName        string             `json:"firstname" bson:"firstname,omitempty"`
	LastName         string             `json:"lastname" bson:"lastname,omitempty"`
	Stripe_plan      string             `json:"stripe_plan" bson:"stripe_plan,omitempty"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at,omitempty"`
	GoogleMFAEnabled bool               `json:"googlemfa_enabled" bson:"google2fa_enabled,omitempty"`
	PublicKey        string             `json:"amazon_access_key_id,omitempty" bson:"amazon_access_key_id,omitempty"`
	PrivateKey       string             `json:"amazon_secret_access_key,omitempty" bson:"amazon_secret_access_key,omitempty"`
	HasKeys          bool               `json:"has_keys" bson:"has_keys,omitempty"`
	DeviceData       DeviceData         `json:"deviceData" bson:"deviceData,omitempty"`
}

type DeviceData struct {
	ActiveDevices int `json:"activeDevices" bson:"activeDevices,omitempty"`
	TotalDevices  int `json:"totalDevices" bson:"totalDevices,omitempty"`
}

type ProfileSettings struct {
	DefaultFloorPlanLabelsHidden bool `json:"defaultFloorPlanLabelsHidden" bson:"defaultFloorPlanLabelsHidden,omitempty"`
}

type UserUpdate struct {
	PublicKey       *string          `json:"amazon_access_key_id" bson:"amazon_access_key_id,omitempty"`
	PrivateKey      *string          `json:"amazon_secret_access_key" bson:"amazon_secret_access_key,omitempty"`
	UpdatedAt       *time.Time       `json:"updated_at" bson:"updated_at,omitempty"`
	Username        *string          `json:"username" bson:"username,omitempty"`
	Email           *string          `json:"email" bson:"email,omitempty"`
	FirstName       *string          `json:"firstname" bson:"firstname,omitempty"`
	LastName        *string          `json:"lastname" bson:"lastname,omitempty"`
	CompanyName     *string          `json:"company_name" bson:"company_name,omitempty"`
	CompanyNumber   *string          `json:"company_number" bson:"company_number,omitempty"`
	Address         *string          `json:"address" bson:"address,omitempty"`
	ProfileSettings *ProfileSettings `json:"profileSettings" bson:"profileSettings,omitempty"`
	// Add fields that are allowed to be updated here
}

type Livestream struct {
	Speech bool `json:"speech" bson:"speech"`
}
