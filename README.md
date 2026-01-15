# Models

Shared data models and TypeScript type definitions for the UUG AI ecosystem.

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://go.dev/)
[![codecov](https://codecov.io/gh/uug-ai/models/graph/badge.svg?token=GD113W0PCL)](https://codecov.io/gh/uug-ai/models)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/uug-ai/models?status.svg)](https://godoc.org/github.com/uug-ai/models)
[![Release](https://img.shields.io/github/release/uug-ai/models.svg)](https://github.com/uug-ai/models/releases/latest)

A comprehensive Go package providing type-safe data models for media management, device control, authentication, analytics, and more. Includes automatic TypeScript type generation for seamless cross-language development.

📋 **[Best Practices Guide](./BEST_PRACTICES.md)** - Comprehensive guidelines for defining Go types in this repository

## Features

- **Comprehensive Type System**: 30+ models covering media, devices, authentication, analytics, and more
- **Cross-Language Support**: Automatic TypeScript type generation from Go structs
- **Database Ready**: Full MongoDB support with BSON tags and validation
- **API Optimized**: JSON serialization with proper omitempty handling
- **Type Safety**: Compile-time type checking in both Go and TypeScript
- **Auto-Discovery**: Automatic model detection and TypeScript generation
- **Production Ready**: Battle-tested models used in production systems

## Installation

### Go
```bash
go get github.com/uug-ai/models
```

### TypeScript
```bash
npm install @uug-ai/models
# or
yarn add @uug-ai/models
```

## Quick Start

### Go Usage

```go
package main

import (
    "github.com/uug-ai/models/pkg/models"
    "time"
)

func main() {
    // Create a media record
    media := models.Media{
        FileName:       "video_001.mp4",
        StartTimestamp: time.Now().Unix(),
        Duration:       600000, // 10 minutes in milliseconds
        DeviceId:       "camera-001",
        UserId:         "user-123",
        VideoUrl:       "https://cdn.example.com/videos/video_001.mp4",
        ThumbnailUrl:   "https://cdn.example.com/thumbs/video_001.jpg",
        SpriteUrl:      "https://cdn.example.com/sprites/video_001.jpg",
        SpriteInterval: 10,
    }
    
    // Create a device
    device := models.Device{
        Name:        "Front Door Camera",
        DeviceId:    "camera-001",
        Type:        "camera",
        Status:      "online",
        GroupId:     "group-001",
        Coordinates: &models.Coordinates{
            Latitude:  37.7749,
            Longitude: -122.4194,
        },
    }
}
```

### TypeScript Usage

```typescript
import { Media, Device, models } from '@uug-ai/models';

// Direct import
const media: Media = {
  fileName: "video_001.mp4",
  startTimestamp: 1640995200,
  duration: 600000,
  deviceId: "camera-001",
  userId: "user-123",
  videoUrl: "https://cdn.example.com/videos/video_001.mp4"
};

// Namespace import
const device: models.Device = {
  name: "Front Door Camera",
  deviceId: "camera-001",
  type: "camera",
  status: "online"
};
```

## Core Concepts

### Automatic Type Generation

This project bridges Go and TypeScript using an automated pipeline:

1. **Go Models** - Define structs in `pkg/models/` with proper tags
2. **Swagger Generation** - `swag` scans code and generates API spec
3. **OpenAPI Conversion** - Swagger 2.0 → OpenAPI 3.x format
4. **TypeScript Generation** - OpenAPI spec → TypeScript types
5. **Post-Processing** - Add convenient exports and build

This ensures type safety across your entire stack with a single source of truth.

### Model Categories

The models package organizes types into logical categories:

- **Media Models** - Video, thumbnails, sprites, and media metadata
- **Device Models** - Cameras, sensors, and IoT device management
- **Authentication** - Users, tokens, permissions, and access control
- **Analytics** - Metrics, timelines, and data analysis
- **Integration** - Third-party service connections and webhooks
- **Infrastructure** - Health checks, configurations, and system status

## Usage Examples

### Generating TypeScript Types

The primary workflow for maintaining cross-language type safety:

```bash
# Generate both OpenAPI spec and TypeScript types
npm run generate

# Or run steps individually:
npm run generate:openapi  # Go models → OpenAPI 3.x YAML
npm run generate:types    # OpenAPI YAML → TypeScript types
```

**Generated Files:**
- `docs/swagger.yaml` - Swagger 2.0 specification (intermediate)
- `docs/openapi.yaml` - OpenAPI 3.x specification  
- `src/typescript/types.ts` - TypeScript type definitions

### Adding New Models

1. **Create Go struct** in `pkg/models/`:

```go
package models

// User represents a system user
type User struct {
    ID        string `json:"id" bson:"_id,omitempty"`
    Email     string `json:"email" bson:"email" validate:"required,email"`
    Name      string `json:"name" bson:"name" validate:"required"`
    CreatedAt int64  `json:"createdAt" bson:"created_at"`
}
```

2. **Reference in `cmd/main.go`**:

```go
// Just declare a variable or add to API endpoint
var _ = models.User{}
```

3. **Generate TypeScript types**:

```bash
npm run generate
```

**No manual script updates needed!** The generation pipeline automatically discovers all referenced models.

### Media Management Example

```go
package main

import (
    "context"
    "time"
    "github.com/uug-ai/models/pkg/models"
    "go.mongodb.org/mongo-driver/mongo"
)

func saveMedia(collection *mongo.Collection, deviceId string) error {
    media := models.Media{
        FileName:         "recording_2024.mp4",
        StartTimestamp:   time.Now().Add(-1 * time.Hour).Unix(),
        EndTimestamp:     time.Now().Unix(),
        Duration:         3600000, // 1 hour in milliseconds
        Provider:         "aws-s3",
        Storage:          "media-bucket",
        DeviceId:         deviceId,
        VideoUrl:         "https://cdn.example.com/videos/recording_2024.mp4",
        ThumbnailUrl:     "https://cdn.example.com/thumbs/recording_2024.jpg",
        ThumbnailFile:    "recording_2024_thumb.jpg",
        SpriteUrl:        "https://cdn.example.com/sprites/recording_2024.jpg",
        SpriteFile:       "recording_2024_sprite.jpg",
        SpriteInterval:   10,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err := collection.InsertOne(ctx, media)
    return err
}
```

### Device Registration Example

```go
package main

import (
    "github.com/uug-ai/models/pkg/models"
    "time"
)

func registerDevice(name, deviceId string, lat, lng float64) models.Device {
    return models.Device{
        Name:        name,
        DeviceId:    deviceId,
        Type:        "camera",
        Status:      "online",
        GroupId:     "group-001",
        CreatedAt:   time.Now().Unix(),
        UpdatedAt:   time.Now().Unix(),
        Coordinates: &models.Coordinates{
            Latitude:  lat,
            Longitude: lng,
        },
        Capabilities: []string{"video", "audio", "motion-detection"},
    }
}
```

### API Response Handling

```go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/uug-ai/models/pkg/api"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    response := api.HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now().Unix(),
        Version:   "1.0.0",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func errorHandler(w http.ResponseWriter, message string, code int) {
    response := api.ErrorResponse{
        Error:      true,
        Message:    message,
        StatusCode: code,
        Timestamp:  time.Now().Unix(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(response)
}
```

## Project Structure

```
.
├── pkg/
│   ├── models/              # Core data models
│   │   ├── media.go        # Media and video models
│   │   ├── device.go       # Device management models
│   │   ├── user.go         # User and authentication models
│   │   ├── authentication.go # Auth tokens and sessions
│   │   ├── analysis.go     # Analytics and metrics
│   │   ├── pipeline.go     # Processing pipelines
│   │   ├── integration.go  # Third-party integrations
│   │   └── ...             # 30+ additional models
│   └── api/                 # API response structures
│       ├── api.go          # Common API types
│       ├── media.go        # Media API responses
│       ├── device.go       # Device API responses
│       └── ...
├── cmd/
│   └── main.go             # Model auto-discovery entry point
├── scripts/
│   ├── auto-discover-models.js # Automatic model detection
│   └── add-type-exports.js     # TypeScript export generation
├── src/
│   └── typescript/
│       ├── types.ts        # Generated TypeScript types
│       └── package.json
├── docs/
│   ├── swagger.yaml        # Swagger 2.0 spec
│   └── openapi.yaml        # OpenAPI 3.x spec
├── go.mod
├── package.json
├── BEST_PRACTICES.md       # Type definition guidelines
└── README.md
```

## Configuration

### Development Setup

1. **Clone the repository**:
```bash
git clone https://github.com/uug-ai/models.git
cd models
```

2. **Install Go dependencies**:
```bash
go mod download
```

3. **Install Node.js dependencies**:
```bash
npm install
```

4. **Install Swagger CLI**:
```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Environment Variables

TypeScript generation can be configured via environment:

```bash
# Skip OpenAPI generation (use existing spec)
SKIP_OPENAPI=true npm run generate

# Verbose output
DEBUG=true npm run generate
```

## Available Models

### Core Models (`pkg/models/`)

- `Media` - Video files, thumbnails, sprites, and metadata
- `Device` - Cameras, sensors, and IoT devices
- `User` - User accounts and profiles
- `Authentication` - Auth tokens, sessions, and credentials
- `AccessToken` - API access tokens and permissions
- `Group` - User groups and organizations
- `Site` - Physical locations and sites
- `Analysis` - Analytics data and metrics
- `Pipeline` - Data processing pipelines
- `Strategy` - Processing strategies and configurations
- `Marker` - Timeline markers and annotations
- `Activity` - User activity logs
- `Alert` - System alerts and notifications
- `Audit` - Audit logs and compliance
- `Comment` - User comments and feedback
- `Config` - System configurations
- `FloorPlan` - Site floor plans and layouts
- `Health` - Health check data
- `Integration` - Third-party integrations
- `Location` - Geographic locations
- `Message` - System messages and notifications
- `Object` - Object detection results
- `Permission` - Access permissions
- `Provider` - Service providers
- `Role` - User roles and capabilities
- `Subscription` - Subscription and billing
- `Vault` - Secure credential storage

### API Models (`pkg/api/`)

- `ErrorResponse` - Standard error responses

### Selective Fields

Clients can request optional fields on certain API payloads using the `include` parameter in request models. For example, site option queries support `include: ["metadata"]` to return site metadata alongside the default fields.
- `HealthResponse` - Health check responses
- `PaginationResponse` - Paginated list responses
- `MediaResponse` - Media query responses
- `DeviceResponse` - Device query responses
- `AnalysisResponse` - Analytics responses

## Validation

Models use [go-playground/validator](https://github.com/go-playground/validator) for field validation:

```go
type User struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Age      int    `json:"age" validate:"gte=0,lte=150"`
    Role     string `json:"role" validate:"required,oneof=admin user guest"`
}

// Validate in your code
import "github.com/go-playground/validator/v10"

validate := validator.New()
err := validate.Struct(user)
if err != nil {
    // Handle validation errors
}
```

Common validation tags:
- `required` - Field must be present
- `email` - Valid email format
- `min=n` - Minimum length/value
- `max=n` - Maximum length/value
- `oneof=a b c` - Must be one of specified values
- `gte=n,lte=n` - Range validation

## Testing

Run the test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Run tests for specific packages:

```bash
# Model tests
go test ./pkg/models -v

# Pipeline tests
go test ./pkg/models -run TestPipeline
```

## Contributing

Contributions are welcome! Please follow the guidelines in [BEST_PRACTICES.md](./BEST_PRACTICES.md) when adding new models.

### Development Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-model`)
3. Add your model to `pkg/models/` or `pkg/api/`
4. Include proper JSON, BSON, and validation tags
5. Reference the model in `cmd/main.go`
6. Generate TypeScript types: `npm run generate`
7. Add tests for your model
8. Ensure all tests pass: `go test ./...`
9. Commit following [Conventional Commits](https://www.conventionalcommits.org/)
10. Push to your branch (`git push origin feat/amazing-model`)
11. Open a Pull Request

### Commit Message Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `types`

**Scopes:**
- `models` - Changes to Go model structs in `pkg/models/`
- `api` - Changes to API response structures in `pkg/api/`
- `types` - TypeScript type generation or definitions
- `scripts` - Build/generation scripts
- `docs` - Documentation updates

**Examples:**

```
feat(models): add sprite interval field to Media struct
fix(api): correct JSON tags for ErrorResponse message field
docs(readme): update TypeScript generation workflow
refactor(models): standardize BSON tag formatting across all structs
types: regenerate TypeScript definitions from updated Go models
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Dependencies

This project uses the following key libraries:

- [swaggo/swag](https://github.com/swaggo/swag) - Swagger/OpenAPI generation from Go
- [mongo-driver](https://github.com/mongodb/mongo-go-driver) - Official MongoDB Go driver
- [go-playground/validator](https://github.com/go-playground/validator) - Struct validation
- [openapi-typescript](https://github.com/drwpow/openapi-typescript) - TypeScript generation from OpenAPI
- [swagger2openapi](https://github.com/Mermade/oas-kit) - Swagger to OpenAPI conversion

See [go.mod](go.mod) and [package.json](package.json) for complete dependency lists.

## Support

- **Issues**: [GitHub Issues](https://github.com/uug-ai/models/issues)
- **Discussions**: [GitHub Discussions](https://github.com/uug-ai/models/discussions)
- **Documentation**: See [BEST_PRACTICES.md](./BEST_PRACTICES.md) and inline code comments