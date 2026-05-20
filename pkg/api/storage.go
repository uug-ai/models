package api

// StorageStatus represents specific status codes for storage / file
// operations (uploads, downloads, provider/account resolution, ...).
type StorageStatus string

const (
	// File-level statuses
	FileUploaded            StorageStatus = "file_uploaded"
	FileUploadFailed        StorageStatus = "file_upload_failed"
	FileEmpty               StorageStatus = "file_empty"
	FileTooLarge            StorageStatus = "file_too_large"
	FileMissingName         StorageStatus = "file_missing_name"
	FileInvalidName         StorageStatus = "file_invalid_name"
	FileMissingExtension    StorageStatus = "file_missing_extension"
	FileExtensionNotAllowed StorageStatus = "file_extension_not_allowed"
	FileReadFailed          StorageStatus = "file_read_failed"

	// Credential statuses
	FileMissingAccessKey       StorageStatus = "file_missing_access_key"
	FileMissingSecretAccessKey StorageStatus = "file_missing_secret_access_key"

	// Provider / account / directory statuses
	StorageAccountNotFound  StorageStatus = "storage_account_not_found"
	StorageProviderNotFound StorageStatus = "storage_provider_not_found"
	StorageDirectoryMissing StorageStatus = "storage_directory_missing"
)

// String returns the string representation of the storage status.
func (ss StorageStatus) String() string {
	return string(ss)
}

// Translate returns the translated message for the storage status in the
// specified language. Falls back to English when the language or status
// is not found.
func (ss StorageStatus) Translate(lang string) string {
	translations := map[string]map[StorageStatus]string{
		"en": {
			FileUploaded:               "File uploaded successfully.",
			FileUploadFailed:           "File upload failed.",
			FileEmpty:                  "File is empty.",
			FileTooLarge:               "File exceeds the maximum allowed size.",
			FileMissingName:            "File name is missing.",
			FileInvalidName:            "File name is invalid.",
			FileMissingExtension:       "File extension is required.",
			FileExtensionNotAllowed:    "File extension is not allowed.",
			FileReadFailed:             "Failed to read the uploaded file.",
			FileMissingAccessKey:       "Access key is missing.",
			FileMissingSecretAccessKey: "Secret access key is missing.",
			StorageAccountNotFound:     "Storage account not found for the provided credentials.",
			StorageProviderNotFound:    "Storage provider could not be found.",
			StorageDirectoryMissing:    "Storage directory is missing.",
		},
	}

	if langTranslations, exists := translations[lang]; exists {
		if translation, exists := langTranslations[ss]; exists {
			return translation
		}
	}
	if enTranslations, exists := translations["en"]; exists {
		if translation, exists := enTranslations[ss]; exists {
			return translation
		}
	}
	return ss.String()
}

// PublishFileResponse is the payload returned on a successful upload via
// the /api/storage/file endpoint.
type PublishFileResponse struct {
	FileName  string `json:"fileName,omitempty" bson:"fileName,omitempty"`
	FileSize  int64  `json:"fileSize,omitempty" bson:"fileSize,omitempty"`
	Directory string `json:"directory,omitempty" bson:"directory,omitempty"`
	Provider  string `json:"provider,omitempty" bson:"provider,omitempty"`
	// SignedURL is a vault-signed URL that can be used to fetch the file
	// after upload. It carries an HMAC signature and a TTL.
	SignedURL string `json:"signedUrl,omitempty" bson:"signedUrl,omitempty"`
}

// PublishFileSuccessResponse is the wrapper success response for
// PublishFile, embedding the standard SuccessResponse envelope.
type PublishFileSuccessResponse struct {
	SuccessResponse
	Data PublishFileResponse `json:"data" bson:"data"`
}

// PublishFileErrorResponse is the wrapper error response for PublishFile.
type PublishFileErrorResponse struct {
	ErrorResponse
}
