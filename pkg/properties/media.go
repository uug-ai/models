package properties

// Media property field names (BSON)
const (
	MediaId                    = "_id"
	MediaStartTimestamp        = "startTimestamp"
	MediaEndTimestamp          = "endTimestamp"
	MediaDuration              = "duration"
	MediaDeviceId              = "deviceId"
	MediaGroupId               = "groupId"
	MediaSiteId                = "siteId"
	MediaOrganisationId        = "organisationId"
	MediaStorageSolution       = "storageSolution"
	MediaVideoFile             = "videoFile"
	MediaVideoProvider         = "videoProvider"
	MediaThumbnailFile         = "thumbnailFile"
	MediaThumbnailProvider     = "thumbnailProvider"
	MediaSpriteFile            = "spriteFile"
	MediaSpriteProvider        = "spriteProvider"
	MediaRedactionFile         = "redactionFile"
	MediaRedactionProvider     = "redactionProvider"
	MediaClassificationSummary = "classificationSummary"
	MediaCountingSummary       = "countingSummary"
	MediaDeviceName            = "deviceName"
	MediaMetadata              = "metadata"
	MediaAtRuntimeMetadata     = "atRuntimeMetadata"
	MediaAudit                 = "audit"
)

// MediaMetadata property field names (BSON)
const (
	MediaMetadataContainer        = "containerType"
	MediaMetadataResolution       = "resolution"
	MediaMetadataWidth            = "width"
	MediaMetadataHeight           = "height"
	MediaMetadataCodec            = "codec"
	MediaMetadataBitrate          = "bitrate"
	MediaMetadataFPS              = "fps"
	MediaMetadataTags             = "tags"
	MediaMetadataSpriteInterval   = "spriteInterval"
	MediaMetadataMotionPixels     = "motionPixels"
	MediaMetadataMotionPercentage = "motionPercentage"
	MediaMetadataAnalysisId       = "analysisId"
	MediaMetadataClassifications  = "classifications"
	MediaMetadataDescription      = "description"
	MediaMetadataDetections       = "detections"
	MediaMetadataDominantColors   = "dominantColors"
	MediaMetadataCount            = "count"
	MediaMetadataEmbedding        = "embedding"
)

// MediaAtRuntimeMetadata property field names (BSON)
const (
	MediaAtRuntimeMetadataCachedTimestamp   = "cachedTimestamp"
	MediaAtRuntimeMetadataVideoUrl          = "videoUrl"
	MediaAtRuntimeMetadataThumbnailUrl      = "thumbnailUrl"
	MediaAtRuntimeMetadataSpriteUrl         = "spriteUrl"
	MediaAtRuntimeMetadataRedactionUrl      = "redactionUrl"
	MediaAtRuntimeMetadataAnalysis          = "analysis"
	MediaAtRuntimeMetadataDevice            = "device"
	MediaAtRuntimeMetadataDurationFormatted = "durationFormatted"
)

// VaultMedia property field names (BSON)
const (
	VaultMediaTimestamp         = "timestamp"
	VaultMediaFileName          = "filename"
	VaultMediaFileSize          = "filesize"
	VaultMediaDevice            = "device"
	VaultMediaAccount           = "account"
	VaultMediaProvider          = "provider"
	VaultMediaStatus            = "status"
	VaultMediaFinished          = "finished"
	VaultMediaTemporary         = "temporary"
	VaultMediaForwarded         = "forwarded"
	VaultMediaToBeForwarded     = "to_be_forwarded"
	VaultMediaUploaded          = "uploaded"
	VaultMediaForwarderId       = "forwarder_id"
	VaultMediaForwarderType     = "forwarder_type"
	VaultMediaForwarderWorker   = "forwarder_worker"
	VaultMediaForwardTimestamp  = "forward_timestamp"
	VaultMediaEvents            = "events"
	VaultMediaMainProvider      = "main_provider"
	VaultMediaSecondaryProvider = "secondary_provider"
	VaultMediaMetadata          = "metadata"
	VaultMediaUriExpiryTime     = "uriExpiryTime"
)

// VaultMediaMetadata property field names (BSON)
const (
	VaultMediaMetadataBytesRanges      = "bytes_ranges"
	VaultMediaMetadataBytesRangeOnTime = "bytes_range_on_time"
	VaultMediaMetadataIsFragmented     = "is_fragmented"
	VaultMediaMetadataDuration         = "duration"
	VaultMediaMetadataTimescale        = "timescale"
)

// VaultMediaFragmentCollection property field names (BSON)
const (
	VaultMediaFragmentCollectionKey              = "key"
	VaultMediaFragmentCollectionFileName         = "filename"
	VaultMediaFragmentCollectionCameraId         = "camera_id"
	VaultMediaFragmentCollectionTimestamp        = "timestamp"
	VaultMediaFragmentCollectionUrl              = "url"
	VaultMediaFragmentCollectionStart            = "start"
	VaultMediaFragmentCollectionEnd              = "end"
	VaultMediaFragmentCollectionDuration         = "duration"
	VaultMediaFragmentCollectionBytesRanges      = "bytes_ranges"
	VaultMediaFragmentCollectionBytesRangeOnTime = "bytes_range_on_time"
)
