# Models

TypeScript type definitions for Go models.

## Installation

```bash
npm install @uug-ai/models
```

## Usage

```typescript
import { Device, Media, Location } from '@uug-ai/models';

const device: Device = {
  id: "507f1f77bcf86cd799439011",
  name: "Front Door Camera",
  deviceId: "camera-001",
  deviceType: "camera"
  // ... other properties
};
```

## API Response Types

```typescript
import { ApiResponse, Metadata } from '@uug-ai/models';

const response: ApiResponse<Device[]> = {
  httpStatusCode: 200,
  applicationStatusCode: 100,
  message: "Devices retrieved successfully",
  data: devices
};
```