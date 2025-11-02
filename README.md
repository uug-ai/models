# UUG AI Model

A Go package defining data models for media file management and metadata handling.

## Overview

This package provides a unified `Media` struct that represents media files (primarily video) along with their associated metadata, storage information, and processed assets like thumbnails and sprites.

📋 **[Best Practices Guide](./BEST_PRACTICES.md)** - Comprehensive guidelines for defining Go types in this repository

## TypeScript Generation

This project automatically generates TypeScript types from Go models using **Swagger/OpenAPI** as an intermediate format.

### How it works

1. **`swag`** scans Go code and generates Swagger 2.0 spec from struct definitions
2. **`swagger2openapi`** converts Swagger 2.0 to OpenAPI 3.x
3. **`openapi-typescript`** generates TypeScript types from OpenAPI 3.x
4. **Post-processing script** adds convenient direct exports

### Usage

To generate TypeScript types from your Go models:

```bash
# Generate both OpenAPI spec and TypeScript types
npm run generate

# Or run steps individually:
npm run generate:openapi  # Go models → OpenAPI 3.x YAML
npm run generate:types    # OpenAPI YAML → TypeScript types
```

### Generated Files

- `docs/swagger.yaml` - Swagger 2.0 specification (intermediate)
- `docs/openapi.yaml` - OpenAPI 3.x specification
- `src/typescript/types.ts` - TypeScript type definitions

### Adding New Models

1. **Add your Go struct** to `pkg/models/`
2. **Reference it in `cmd/main.go`** (add to the variable declarations or API endpoints)
3. **Run `npm run generate`** to update the TypeScript types

**No manual script updates needed!** The `swag` tool automatically discovers all referenced models.

### Example Usage in TypeScript

```typescript
import { Media, models } from './src/types';

// Direct import
const media: Media = {
  deviceId: "camera-001",
  startTimestamp: 1640995200,
  endTimestamp: 1640998800,
  duration: 3600,
  videoUrl: "https://example.com/video.mp4"
};

// Namespace import
const media2: models.Media = { ... };
```

## Media Model

The `Media` struct is designed to handle comprehensive media file information including:

### Core Properties
- **FileName**: The original name of the media file
- **StartTimestamp**: When the media recording started (Unix timestamp)
- **EndTimestamp**: When the media recording ended (Unix timestamp)
- **Duration**: Total duration of the media in milliseconds
- **Provider**: The service/provider that handled the media
- **Storage**: Storage backend identifier where the media is stored

### Identification
- **DeviceId**: Unique identifier for the device that created the media
- **GroupId**: Group/organization identifier for access control
- **UserId**: User identifier who owns/created the media

### Media Assets
- **VideoUrl**: Direct URL to access the video file
- **ThumbnailUrl**: URL to the generated thumbnail image
- **ThumbnailFile**: File path/name of the thumbnail
- **ThumbnailProvider**: Service that generated the thumbnail
- **SpriteUrl**: URL to the sprite sheet for video scrubbing
- **SpriteFile**: File path/name of the sprite sheet
- **SpriteProvider**: Service that generated the sprite sheet
- **SpriteInterval**: Time interval (in seconds) between sprite frames

## Use Cases

This model is typically used for:

- **Video Management Systems**: Storing metadata for uploaded videos
- **Media Processing Pipelines**: Tracking processed assets and their locations
- **Content Delivery**: Managing URLs and access to media files
- **Analytics**: Tracking media duration, creation times, and user associations
- **UI Components**: Providing data for video players with thumbnail previews and scrubbing

## JSON/BSON Support

The struct includes JSON and BSON tags for seamless serialization with:
- REST APIs (JSON)
- MongoDB databases (BSON)
- Other NoSQL databases

All fields use `omitempty` to exclude empty values from serialization.

## Package Information

- **Language**: Go
- **Package**: main
- **Dependencies**: Standard library only
- **Database Support**: MongoDB via BSON tags

## Example Usage

```go
media := Media{
    FileName:       "video_001.mp4",
    StartTimestamp: 1640995200000,
    EndTimestamp:   1640995800000,
    Duration:       600000, // 10 minutes
    Provider:       "aws-s3",
    Storage:        "media-bucket",
    DeviceId:       "device_123",
    UserId:         "user_456",
    VideoUrl:       "https://cdn.example.com/videos/video_001.mp4",
    ThumbnailUrl:   "https://cdn.example.com/thumbs/video_001.jpg",
    SpriteUrl:      "https://cdn.example.com/sprites/video_001.jpg",
    SpriteInterval: 10,
}
```

## License

This project is part of the UUG AI ecosystem.