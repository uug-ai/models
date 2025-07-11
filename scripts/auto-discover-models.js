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
    
    return structNames;
}

// Function to scan all Go files in pkg/models for structs
function findAllModelStructs() {
    const modelsDir = path.join(__dirname, '..', 'pkg', 'models');
    const files = fs.readdirSync(modelsDir).filter(f => f.endsWith('.go'));
    
    let allStructs = [];
    
    for (const file of files) {
        const filePath = path.join(modelsDir, file);
        const structs = extractStructNames(filePath);
        allStructs = allStructs.concat(structs);
        console.log(`Found in ${file}: ${structs.join(', ')}`);
    }
    
    return [...new Set(allStructs)]; // Remove duplicates
}

// Function to generate the cmd/main.go file
function generateMainFile(structNames) {
    // Function to generate dummy endpoints for models not used in real endpoints
    function generateDummyEndpoints(models) {
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
// @Success 200 {object} models.${model}
// @Router /internal/${model.toLowerCase()} [get]
func Get${model}() {}`).join('\n');
        
        return dummyEndpoints;
    }
    
    const mainContent = `package main

import (
\t"fmt"

\t"github.com/uug-ai/models/pkg/models"
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
}

// GetMedia godoc
// @Summary Get a media item
// @Description Get a media item by ID
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} models.Media
// @Failure 400 {object} models.ErrorResponse
// @Router /media/{id} [get]
func GetMedia() {}

// CreateMedia godoc
// @Summary Create a new media item
// @Description Create a new media item
// @Tags media
// @Accept json
// @Produce json
// @Param media body models.Media true "Media object"
// @Success 201 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Router /media [post]
func CreateMedia() {}

// Dummy endpoints to ensure all models are included in OpenAPI spec
// These endpoints exist only to force swag to generate schemas for all models
${generateDummyEndpoints(structNames)}
`;

    const mainFilePath = path.join(__dirname, '..', 'cmd', 'main.go');
    fs.writeFileSync(mainFilePath, mainContent);
    console.log(`Generated cmd/main.go with ${structNames.length} model references`);
}

// Main execution
console.log('Auto-discovering Go structs in pkg/models...');
const structNames = findAllModelStructs();
console.log(`Found ${structNames.length} structs: ${structNames.join(', ')}`);

generateMainFile(structNames);
console.log('cmd/main.go updated automatically!');
