#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Read the generated types file
const typesPath = path.join(__dirname, '..', 'src', 'typescript', 'types.ts');
let content = fs.readFileSync(typesPath, 'utf8');

// Extract schema names from the components interface
const schemaNames = [];

// Simple approach: find all lines that look like schema definitions
const lines = content.split('\n');
let inSchemasSection = false;
let braceCount = 0;

for (const line of lines) {
    if (line.includes('schemas: {')) {
        inSchemasSection = true;
        braceCount = 1;
        continue;
    }
    
    if (inSchemasSection) {
        // Count braces to know when we exit the schemas section
        braceCount += (line.match(/{/g) || []).length;
        braceCount -= (line.match(/}/g) || []).length;
        
        if (braceCount <= 0) {
            break; // We've exited the schemas section
        }
        
        // Look for schema name pattern: "models.SchemaName": { or "api.SchemaName": { or SchemaName: {
        const quotedMatch = line.trim().match(/^"([^"]+)":\s*{/);
        const unquotedMatch = line.trim().match(/^(\w+):\s*{/);
        
        if (quotedMatch) {
            const fullName = quotedMatch[1];
            // Extract just the schema name based on prefix
            let schemaName, namespace;
            if (fullName.startsWith('models.')) {
                schemaName = fullName.replace(/^models\./, '');
                namespace = 'models';
            } else if (fullName.startsWith('api.')) {
                schemaName = fullName.replace(/^api\./, '');
                namespace = 'api';
            } else {
                schemaName = fullName;
                namespace = 'models'; // default namespace
            }
            schemaNames.push({ fullName, schemaName, namespace });
        } else if (unquotedMatch) {
            const schemaName = unquotedMatch[1];
            schemaNames.push({ fullName: schemaName, schemaName, namespace: 'models' });
        }
    }
}

// Extract API paths from the paths interface
const apiPaths = [];
let inPathsSection = false;
braceCount = 0;

for (const line of lines) {
    if (line.includes('paths: {')) {
        inPathsSection = true;
        braceCount = 1;
        continue;
    }
    
    if (inPathsSection) {
        // Count braces to know when we exit the paths section
        braceCount += (line.match(/{/g) || []).length;
        braceCount -= (line.match(/}/g) || []).length;
        
        if (braceCount <= 0) {
            break; // We've exited the paths section
        }
        
        // Look for path pattern: "/api/path": { or "path": {
        const pathMatch = line.trim().match(/^"([^"]+)":\s*{/);
        
        if (pathMatch) {
            const pathName = pathMatch[1];
            // Convert path to a valid TypeScript identifier
            const apiName = pathName
                .replace(/^\/+/, '') // Remove leading slashes
                .replace(/\/+$/, '') // Remove trailing slashes
                .replace(/[\/\-\.]/g, '_') // Replace special chars with underscore
                .replace(/[^a-zA-Z0-9_]/g, '') // Remove invalid chars
                .replace(/^(\d)/, '_$1') // Prefix with underscore if starts with number
                .replace(/_+/g, '_') // Replace multiple underscores with single
                .replace(/^_|_$/g, ''); // Remove leading/trailing underscores
            
            if (apiName) {
                apiPaths.push({ fullName: pathName, apiName });
            }
        }
    }
}

// Group schemas by namespace
const modelSchemas = schemaNames.filter(s => s.namespace === 'models');
const apiSchemas = schemaNames.filter(s => s.namespace === 'api');

console.log(`Found model schemas: ${modelSchemas.map(s => s.schemaName).join(', ')}`);
console.log(`Found API schemas: ${apiSchemas.map(s => s.schemaName).join(', ')}`);
console.log(`Found API paths: ${apiPaths.map(p => p.apiName).join(', ')}`);

// Generate namespace exports for models
const namespaceModelExports = modelSchemas
    .map(s => `    export type ${s.schemaName} = components['schemas']['${s.fullName}'];`)
    .join('\n');

// Generate namespace exports for API schemas and paths
const namespaceApiExports = [
    ...apiSchemas.map(s => `    export type ${s.schemaName} = components['schemas']['${s.fullName}'];`),
    ...apiPaths.map(p => `    export type ${p.apiName} = paths['${p.fullName}'];`)
].join('\n');

// Add convenient type exports at the end - only namespaces
const typeExports = `
// Export namespaces for organized access
export namespace models {
${namespaceModelExports}
}

export namespace api {
${namespaceApiExports}
}
`;

// Append the exports
content += typeExports;

// Write back to file
fs.writeFileSync(typesPath, content);

console.log(`Added convenient type exports for ${modelSchemas.length} model schemas, ${apiSchemas.length} API schemas, and ${apiPaths.length} API paths`);