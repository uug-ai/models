#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Function to extract Go struct names from a file
function extractStructNames(filePath) {
    const content = fs.readFileSync(filePath, 'utf8');
    const structNames = [];
    
    // Match struct declarations: "type StructName struct"
    const structRegex = /type\s+([A-Z]\w*)\s+struct/g;
    let match;
    
    while ((match = structRegex.exec(content)) !== null) {
        structNames.push(match[1]);
    }
    
    // Match empty structs and type aliases: "type StructName struct{}" or "type StructName TypeAlias"
    const emptyStructRegex = /type\s+([A-Z]\w*)\s+(?:struct\s*\{\s*\}|[A-Z]\w*)/g;
    let emptyMatch;
    
    while ((emptyMatch = emptyStructRegex.exec(content)) !== null) {
        if (!structNames.includes(emptyMatch[1])) {
            structNames.push(emptyMatch[1]);
        }
    }
    
    // Match interface declarations: "type InterfaceName interface"
    const interfaceRegex = /type\s+([A-Z]\w*)\s+interface/g;
    let interfaceMatch;
    
    while ((interfaceMatch = interfaceRegex.exec(content)) !== null) {
        if (!structNames.includes(interfaceMatch[1])) {
            structNames.push(interfaceMatch[1]);
        }
    }
    
    return structNames;
}

// Function to scan all Go files in pkg/models for structs
function findAllModelStructs(package) {
    const packageDir = path.join(__dirname, '..', 'pkg', package);

    let allStructs = [];
    
    // Scan package directory
    if (fs.existsSync(packageDir)) {
        const files = fs.readdirSync(packageDir).filter(f => f.endsWith('.go'));
        
        for (const file of files) {
            const filePath = path.join(packageDir, file);
            const structs = extractStructNames(filePath);
            allStructs = allStructs.concat(structs);
            console.log(`Found in ${package}/${file}: ${structs.join(', ')}`);
        }
    }
    
    return [...new Set(allStructs)]; // Remove duplicates
}

// Function to generate the cmd/main.go file
function generateMainFile(apiStructNames, modelsStructNames) {
    // Function to generate dummy endpoints for models not used in real endpoints
    function generateDummyEndpoints(models, packageName) {
        // Models that are already used in real endpoints don't need dummy endpoints
        const usedInEndpoints = new Set(['Media', 'ErrorResponse', 'SuccessResponse']);
        
        const dummyEndpoints = models
            .filter(model => !usedInEndpoints.has(model))
            .map(model => `
// Get${model} godoc
// @Summary Get ${model} (schema generation only)
// @Description Internal endpoint used only to ensure ${model} schema is generated in OpenAPI spec
// @Tags internal
// @Accept json
// @Produce json
// @Success 200 {object} ${packageName}.${model}
// @Router /internal/${model.toLowerCase()} [get]
func Get${model}() {}`).join('\n');
        
        return dummyEndpoints;
    }
    
    const mainContent = `package main

import (
\t"fmt"

\t"github.com/uug-ai/models/pkg/models"
\t"github.com/uug-ai/models/pkg/api"
)

// @title Models API
// @version 1.0
// @description API for media models and related types
// @host localhost
// @BasePath /

func main() {
\t// This main function exists to allow swag to generate OpenAPI spec
\t// from the models in pkg/models
\tfmt.Println("Models API")
\t
\t// Keep the import valid - models are used in the API endpoint annotations below
\t_ = models.Media{}
\t_ = api.SuccessResponse{}
}

// GetMedia godoc
// @Summary Get a media item
// @Description Get a media item by ID
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /media/{id} [get]
func GetMedia() {}

// CreateMedia godoc
// @Summary Create a new media item
// @Description Create a new media item
// @Tags media
// @Accept json
// @Produce json
// @Param media body models.Media true "Media object"
// @Success 201 {object} api.SuccessResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /media [post]
func CreateMedia() {}

// Dummy endpoints to ensure all models are included in OpenAPI spec
// These endpoints exist only to force swag to generate schemas for all models

// API package models
${generateDummyEndpoints(apiStructNames, 'api')}

// Models package models
${generateDummyEndpoints(modelsStructNames, 'models')}
`;

    const mainFilePath = path.join(__dirname, '..', 'cmd', 'main.go');
    fs.writeFileSync(mainFilePath, mainContent);
    console.log(`Generated cmd/main.go with ${apiStructNames.length + modelsStructNames.length} model references`);
}

// Main execution
console.log('Auto-discovering Go structs in pkg/models...');
const apiStructNames = findAllModelStructs('api');
const modelsStructNames = findAllModelStructs('models');
console.log(`Found ${apiStructNames.length} API structs: ${apiStructNames.join(', ')}`);
console.log(`Found ${modelsStructNames.length} Models structs: ${modelsStructNames.join(', ')}`);

// Generate single main.go file with both API and Models structs
generateMainFile(apiStructNames, modelsStructNames);
console.log('cmd/main.go updated automatically!');