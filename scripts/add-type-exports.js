#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Read the generated types file
const typesPath = path.join(__dirname, '..', 'src', 'types.ts');
let content = fs.readFileSync(typesPath, 'utf8');

// Extract schema names from the components interface
// Look for the schemas section more carefully
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
        
        // Look for schema name pattern: "models.SchemaName": { or SchemaName: {
        const quotedMatch = line.trim().match(/^"([^"]+)":\s*{/);
        const unquotedMatch = line.trim().match(/^(\w+):\s*{/);
        
        if (quotedMatch) {
            const fullName = quotedMatch[1];
            // Extract just the schema name (remove models. prefix if present)
            const schemaName = fullName.replace(/^models\./, '');
            schemaNames.push({ fullName, schemaName });
        } else if (unquotedMatch) {
            const schemaName = unquotedMatch[1];
            schemaNames.push({ fullName: schemaName, schemaName });
        }
    }
}

console.log(`Found schemas: ${schemaNames.map(s => s.schemaName).join(', ')}`);

// Generate direct exports
const directExports = schemaNames
    .map(s => `export type ${s.schemaName} = components['schemas']['${s.fullName}'];`)
    .join('\n');

// Generate namespace exports
const namespaceExports = schemaNames
    .map(s => `    export type ${s.schemaName} = components['schemas']['${s.fullName}'];`)
    .join('\n');

// Add convenient type exports at the end
const typeExports = `
// Convenient type exports for easier access
${directExports}

// Export namespace for organized access
export namespace models {
${namespaceExports}
}
`;

// Append the exports
content += typeExports;

// Write back to file
fs.writeFileSync(typesPath, content);

console.log(`Added convenient type exports for ${schemaNames.length} schemas: ${schemaNames.map(s => s.schemaName).join(', ')}`);
