#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Read the generated types file
const typesPath = path.join(__dirname, '..', 'src', 'types.ts');
let content = fs.readFileSync(typesPath, 'utf8');

// Add convenient type exports at the end
const typeExports = `
// Convenient type exports for easier access
export type Media = components['schemas']['Media'];
export type MediaMetadata = components['schemas']['MediaMetadata'];
export type APIMetadata = components['schemas']['APIMetadata'];
export type ErrorResponse = components['schemas']['ErrorResponse'];
export type SuccessResponse = components['schemas']['SuccessResponse'];

// Export namespace for organized access
export namespace models {
    export type Media = components['schemas']['Media'];
    export type MediaMetadata = components['schemas']['MediaMetadata'];
    export type APIMetadata = components['schemas']['APIMetadata'];
    export type ErrorResponse = components['schemas']['ErrorResponse'];
    export type SuccessResponse = components['schemas']['SuccessResponse'];
}
`;

// Append the exports
content += typeExports;

// Write back to file
fs.writeFileSync(typesPath, content);

console.log('Added convenient type exports to types.ts');
