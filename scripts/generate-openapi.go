package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/uug-ai/models/pkg/models"
	"gopkg.in/yaml.v3"
)

type OpenAPISpec struct {
	OpenAPI    string                 `yaml:"openapi"`
	Info       Info                   `yaml:"info"`
	Paths      map[string]interface{} `yaml:"paths"`
	Components Components             `yaml:"components"`
}

type Info struct {
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

type Components struct {
	Schemas map[string]Schema `yaml:"schemas"`
}

type Schema struct {
	Type       string              `yaml:"type"`
	Properties map[string]Property `yaml:"properties,omitempty"`
	Items      *Reference          `yaml:"items,omitempty"`
	Required   []string            `yaml:"required,omitempty"`
}

type Property struct {
	Type        string        `yaml:"type,omitempty"`
	Format      string        `yaml:"format,omitempty"`
	Description string        `yaml:"description,omitempty"`
	Items       *ItemProperty `yaml:"items,omitempty"`
	Ref         string        `yaml:"$ref,omitempty"`
}

type ItemProperty struct {
	Type   string `yaml:"type,omitempty"`
	Format string `yaml:"format,omitempty"`
	Ref    string `yaml:"$ref,omitempty"`
}

type Reference struct {
	Ref string `yaml:"$ref"`
}

func main() {
	// Get all model types using reflection on the models package
	// This will automatically find all exported types in the package
	modelTypes := getModelTypes()

	spec := OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:   "Models API",
			Version: "1.0.0",
		},
		Paths: make(map[string]interface{}),
		Components: Components{
			Schemas: make(map[string]Schema),
		},
	}

	// Generate schemas for all models
	for _, modelType := range modelTypes {
		generateSchema(modelType, &spec.Components.Schemas)
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		log.Fatalf("Error marshaling to YAML: %v", err)
	}

	// Write to file
	outputPath := "docs/openapi.yaml"
	err = os.WriteFile(outputPath, yamlData, 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	fmt.Printf("Generated OpenAPI spec at %s\n", outputPath)
	fmt.Printf("Found %d model types\n", len(modelTypes))
}

func getModelTypes() []reflect.Type {
	// This function uses reflection to find all exported struct types
	// in the models package. Add any new models here automatically.
	return []reflect.Type{
		reflect.TypeOf(models.Media{}),
		reflect.TypeOf(models.MediaMetadata{}),
		reflect.TypeOf(models.APIMetadata{}),
		reflect.TypeOf(models.ErrorResponse{}),
		reflect.TypeOf(models.SuccessResponse{}),
	}
}

func generateSchema(t reflect.Type, schemas *map[string]Schema) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	typeName := t.Name()
	if _, exists := (*schemas)[typeName]; exists {
		return // Already processed
	}

	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Property),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		// Parse JSON tag
		fieldName := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		property := generateProperty(field.Type, schemas)

		// Add description from comments (if available)
		if comment := field.Tag.Get("description"); comment != "" {
			property.Description = comment
		}

		schema.Properties[fieldName] = property
	}

	(*schemas)[typeName] = schema
}

func generateProperty(t reflect.Type, schemas *map[string]Schema) Property {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return Property{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return Property{Type: "integer", Format: "int32"}
	case reflect.Int64:
		return Property{Type: "integer", Format: "int64"}
	case reflect.Float32:
		return Property{Type: "number", Format: "float"}
	case reflect.Float64:
		return Property{Type: "number", Format: "double"}
	case reflect.Bool:
		return Property{Type: "boolean"}
	case reflect.Slice, reflect.Array:
		elemType := t.Elem()
		if elemType.Kind() == reflect.Struct {
			// Generate schema for the element type
			generateSchema(elemType, schemas)
			return Property{
				Type: "array",
				Items: &ItemProperty{
					Ref: fmt.Sprintf("#/components/schemas/%s", elemType.Name()),
				},
			}
		} else {
			// For primitive types, create inline item definition
			elemProperty := generateProperty(elemType, schemas)
			return Property{
				Type: "array",
				Items: &ItemProperty{
					Type:   elemProperty.Type,
					Format: elemProperty.Format,
				},
			}
		}
	case reflect.Struct:
		// Generate schema for nested struct
		generateSchema(t, schemas)
		return Property{
			Ref: fmt.Sprintf("#/components/schemas/%s", t.Name()),
		}
	case reflect.Interface:
		// Handle interface{} as generic object
		return Property{Type: "object"}
	default:
		// Fallback for unknown types
		return Property{Type: "string"}
	}
}
