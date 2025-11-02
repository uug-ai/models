# Go Types Best Practices

This document outlines the best practices for defining Go types in the models repository. Following these guidelines ensures consistency, maintainability, and proper integration with our TypeScript generation pipeline.

## Table of Contents

- [General Principles](#general-principles)
- [Struct Definition](#struct-definition)
- [Field Naming and Tags](#field-naming-and-tags)
- [Comments and Documentation](#comments-and-documentation)
- [Type Organization](#type-organization)
- [Embedded Types and Composition](#embedded-types-and-composition)
- [Constants and Enums](#constants-and-enums)
- [Validation and Constraints](#validation-and-constraints)
- [Database Integration](#database-integration)
- [TypeScript Generation](#typescript-generation)
- [Common Patterns](#common-patterns)
- [Anti-Patterns to Avoid](#anti-patterns-to-avoid)

## General Principles

### 1. Consistency First
- Follow established patterns in the existing codebase
- Use consistent naming conventions across all models
- Maintain uniform field ordering and grouping

### 2. Clear Intent
- Type names should clearly indicate their purpose
- Field names should be descriptive and unambiguous
- Use meaningful comments to explain business logic

### 3. Future-Proof Design
- Design for extensibility without breaking changes
- Use pointers for optional complex types
- Consider backward compatibility when modifying existing types

### 4. Integer Type Selection


Use `int` by default, use `int64` when you need large or cross-platform stable numbers, and use fixed-size types when you're interfacing with external systems.

- **`int`** - Default choice for counters, lengths, and general numeric values
- **`int64`** - For timestamps, large numbers, or when you need consistent size across platforms
- **`int32`, `int16`, `int8`** - When interfacing with external APIs or systems that require specific sizes
- **`uint` variants** - Only when you specifically need unsigned values (avoid unless necessary)

## Struct Definition

### Basic Structure
```go
// ModelName represents a clear description of what this model does
type ModelName struct {
    // Core identifier (always first)
    Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    
    // Required fields (group logically)
    Name        string `json:"name" bson:"name"`
    DeviceId    string `json:"deviceId" bson:"deviceId"`
    
    // Optional fields with omitempty
    Description string `json:"description,omitempty" bson:"description,omitempty"`
    
    // Nested objects (use pointers for complex types)
    Metadata *ModelMetadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
    
    // Audit fields (always last)
    Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
```

### Field Ordering Guidelines
1. **ID field** - Always first
2. **Core identifying fields** - Name, DeviceId, etc.
3. **RBAC fields** - SiteId, GroupId, OrganisationId
4. **Business logic fields** - Group by functionality
5. **Metadata objects** - Complex nested structures
6. **Runtime metadata** - Data generated at runtime
7. **Audit fields** - Always last

## Field Naming and Tags

### Naming Conventions
- Use **PascalCase** for exported fields
- Use **camelCase** for JSON tags
- Use **snake_case** for BSON tags only when necessary
- Be consistent with abbreviations (Id vs ID, Url vs URL)

### JSON and BSON Tags
```go
type Example struct {
    // Required field - no omitempty
    Name string `json:"name" bson:"name"`
    
    // Optional field - always use omitempty
    Description string `json:"description,omitempty" bson:"description,omitempty"`
    
    // ID fields - special BSON handling
    Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    
    // Arrays/slices - always use omitempty
    Tags []string `json:"tags,omitempty" bson:"tags,omitempty"`
    
    // Pointers to structs - always use omitempty
    Metadata *Metadata `json:"metadata,omitempty" bson:"metadata,omitempty"`
}
```

### Tag Best Practices
- **Always include both JSON and BSON tags**
- **Use `omitempty`** for optional fields, arrays, and pointer fields
- **Match JSON and BSON names** unless there's a specific reason not to
- **Use `_id` for BSON** when dealing with MongoDB primary keys

## Comments and Documentation

### Struct Comments
```go
// Device represents a physical or virtual device in the UUG AI system.
// It contains identification, configuration, and runtime metadata for
// devices such as cameras, sensors, and access control systems.
type Device struct {
    // ... fields
}
```

### Field Comments
```go
type Device struct {
    // DeviceId is a unique identifier for the device across the system
    DeviceId string `json:"deviceId" bson:"deviceId"`
    
    // LastSeenTimestamp indicates when the device last communicated
    // with the system (timestamp in milliseconds since Unix epoch)
    LastSeenTimestamp int64 `json:"lastSeenTimestamp,omitempty" bson:"lastSeenTimestamp,omitempty"`
}
```

### Comment Guidelines
- **Document purpose**, not implementation
- **Explain business context** for non-obvious fields
- **Include units** for numeric fields (milliseconds, bytes, etc.)
- **Describe relationships** between fields when relevant
- **Use present tense** and be concise

## Type Organization

### File Structure
- **One main type per file** (e.g., `device.go` for `Device`)
- **Related types in the same file** (e.g., `DeviceMetadata`, `DeviceStatus`)
- **Group by domain**, not by technical concerns

### Logical Grouping
```go
// Main entity
type Device struct { ... }

// Supporting metadata types
type DeviceMetadata struct { ... }
type DeviceCameraMetadata struct { ... }
type DeviceLocationMetadata struct { ... }

// Runtime-specific types
type DeviceAtRuntimeMetadata struct { ... }

// Permission-related types
type DeviceFeaturePermissions struct { ... }
```

## Embedded Types and Composition

### Use Composition Over Inheritance
```go
// Good: Composition with explicit fields
type Media struct {
    DeviceId string `json:"deviceId" bson:"deviceId"`
    Audit    *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}

// Avoid: Anonymous embedding (makes TypeScript generation complex)
type Media struct {
    Device // Anonymous embedding - avoid
    Audit  // Anonymous embedding - avoid
}
```

### Common Embedded Patterns
```go
// Audit information - used across many models
type Audit struct {
    CreatedBy string `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
    CreatedAt int64  `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
    UpdatedBy string `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
    UpdatedAt int64  `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

// Include as pointer in main types
type SomeModel struct {
    // ... other fields
    Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
```

## Constants and Enums

### Define Constants for Status Values
```go
const (
    DEVICE_STATUS_ONLINE     = "online"
    DEVICE_STATUS_OFFLINE    = "offline"
    DEVICE_STATUS_MAINTENANCE = "maintenance"
)

const (
    USER_FOUND     = "One or more users were found"
    USER_NOT_FOUND = "One or more users not found, returning empty list"
)
```

### Enum-like Patterns
```go
// Use string constants for enum-like values
type DeviceType string

const (
    DeviceTypeCamera        DeviceType = "camera"
    DeviceTypeSensor        DeviceType = "sensor"
    DeviceTypeAccessControl DeviceType = "access_control"
)

// Use in struct
type Device struct {
    DeviceType string `json:"deviceType,omitempty" bson:"deviceType,omitempty"`
}
```

## Validation and Constraints

### Field Validation Comments
```go
type User struct {
    // Email must be a valid email address format
    Email string `json:"email,omitempty" bson:"email,omitempty"`
    
    // Username must be 3-50 characters, alphanumeric and underscore only
    Username string `json:"username,omitempty" bson:"username,omitempty"`
    
    // CustomUsageLimit must be >= 0, 0 means no custom limit
    CustomUsageLimit int `json:"custom_usage_limit,omitempty" bson:"custom_usage_limit,omitempty"`
}
```

### Required vs Optional Fields
```go
type Device struct {
    // Required fields - no omitempty
    DeviceId string `json:"deviceId" bson:"deviceId"`
    Name     string `json:"name" bson:"name"`
    
    // Optional fields - with omitempty
    Description string `json:"description,omitempty" bson:"description,omitempty"`
    Version     string `json:"version,omitempty" bson:"version,omitempty"`
}
```

## Database Integration

### MongoDB Best Practices
```go
type Model struct {
    // Use primitive.ObjectID for MongoDB _id
    Id primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    
    // Index-worthy fields should be documented
    // @index: unique
    DeviceId string `json:"deviceId" bson:"deviceId"`
    
    // Timestamps as int64 (milliseconds since epoch)
    CreatedAt int64 `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}
```

### Indexing Considerations
- **Document fields used for queries** in comments
- **Consider compound indexes** for multi-field queries
- **Use consistent timestamp formats** (int64 milliseconds)

## TypeScript Generation

### Swagger-Compatible Types
```go
// Use basic Go types that translate well to TypeScript
type Example struct {
    // Primitives - translate directly
    Name        string  `json:"name"`
    Count       int     `json:"count,omitempty"`
    Price       float64 `json:"price,omitempty"`
    IsActive    bool    `json:"isActive,omitempty"`
    
    // Arrays - translate to TypeScript arrays
    Tags        []string `json:"tags,omitempty"`
    
    // Objects - become TypeScript interfaces
    Metadata    *ExampleMetadata `json:"metadata,omitempty"`
    
    // Maps - become TypeScript Records
    Settings    map[string]interface{} `json:"settings,omitempty"`
}
```

### Avoiding TypeScript Issues
- **Avoid anonymous structs** - define named types instead
- **Avoid complex interface{}** usage - be specific where possible
- **Use consistent field naming** - camelCase in JSON tags
- **Document any custom types** that might not translate obviously

## Common Patterns

### The Metadata Pattern
```go
// Main entity with core fields
type Device struct {
    Id       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    DeviceId string            `json:"deviceId" bson:"deviceId"`
    
    // Metadata for additional information
    Metadata         *DeviceMetadata         `json:"metadata,omitempty" bson:"metadata,omitempty"`
    CameraMetadata   *DeviceCameraMetadata   `json:"cameraMetadata,omitempty" bson:"cameraMetadata,omitempty"`
    LocationMetadata *DeviceLocationMetadata `json:"locationMetadata,omitempty" bson:"locationMetadata,omitempty"`
    
    // Runtime metadata for computed values
    AtRuntimeMetadata *DeviceAtRuntimeMetadata `json:"atRuntimeMetadata,omitempty" bson:"atRuntimeMetadata,omitempty"`
}
```

### The RBAC Pattern
```go
type Entity struct {
    // RBAC fields - consistent across entities
    SiteId         string `json:"siteId,omitempty" bson:"siteId,omitempty"`
    GroupId        string `json:"groupId,omitempty" bson:"groupId,omitempty"`
    OrganisationId string `json:"organisationId,omitempty" bson:"organisationId,omitempty"`
}
```

### The Audit Pattern
```go
type Entity struct {
    // Other fields...
    
    // Audit trail - always as pointer at the end
    Audit *Audit `json:"audit,omitempty" bson:"audit,omitempty"`
}
```

### The Timestamp Pattern
```go
type Entity struct {
    // Use int64 for timestamps (milliseconds since epoch)
    StartTimestamp    int64 `json:"startTimestamp,omitempty" bson:"startTimestamp,omitempty"`
    EndTimestamp      int64 `json:"endTimestamp,omitempty" bson:"endTimestamp,omitempty"`
    LastSeenTimestamp int64 `json:"lastSeenTimestamp,omitempty" bson:"lastSeenTimestamp,omitempty"`
}
```

## Anti-Patterns to Avoid

### ❌ Don't Do This

```go
// Don't use anonymous structs
type Bad struct {
    Config struct {
        Name string `json:"name"`
    } `json:"config"`
}

// Don't mix naming conventions
type Bad struct {
    user_name string `json:"userName"` // Mixed snake_case and camelCase
    UserID    string `json:"userid"`   // Inconsistent casing
}

// Don't forget omitempty for optional fields
type Bad struct {
    OptionalField string `json:"optionalField"` // Missing omitempty
}

// Don't use interface{} without good reason
type Bad struct {
    Data interface{} `json:"data"` // Too generic
}

// Don't embed anonymously for TypeScript generation
type Bad struct {
    SomeStruct      // Anonymous embedding
    AnotherStruct   // Makes TypeScript generation complex
}
```

### ✅ Do This Instead

```go
// Define named types
type ConfigStruct struct {
    Name string `json:"name"`
}

type Good struct {
    Config *ConfigStruct `json:"config,omitempty"`
}

// Use consistent naming
type Good struct {
    Username string `json:"username"`
    UserID   string `json:"userId"`
}

// Always use omitempty for optional fields
type Good struct {
    OptionalField string `json:"optionalField,omitempty"`
}

// Be specific about data types
type Good struct {a
    Settings map[string]string `json:"settings,omitempty"`
    Tags     []string         `json:"tags,omitempty"`
}

// Use explicit composition
type Good struct {
    ConfigID string        `json:"configId"`
    Config   *ConfigStruct `json:"config,omitempty"`
}
```

## Checklist for New Types

Before adding a new type to the repository, ensure:

- [ ] **Struct name** clearly describes its purpose
- [ ] **Fields are logically ordered** (ID, core, RBAC, business, metadata, audit)
- [ ] **JSON tags use camelCase** and match BSON tags where appropriate
- [ ] **Optional fields use `omitempty`**
- [ ] **Complex nested types use pointers**
- [ ] **Comments document purpose and business context**
- [ ] **Constants are defined** for any enum-like string values
- [ ] **Audit fields included** if the entity tracks changes
- [ ] **RBAC fields included** if the entity has access control
- [ ] **Type added to `cmd/main.go`** for TypeScript generation
- [ ] **TypeScript types generated** and tested

## Examples

See the following files for examples of well-structured types:
- `pkg/models/device.go` - Complex entity with multiple metadata types
- `pkg/models/media.go` - Media handling with runtime metadata
- `pkg/models/audit.go` - Simple reusable audit type
- `pkg/models/user.go` - User entity with settings and permissions

Following these best practices ensures that our Go types are maintainable, consistent, and integrate well with our TypeScript generation pipeline while providing clear, self-documenting APIs for the UUG AI ecosystem.