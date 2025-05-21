package tool

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// Define test structs
type TestParamStringStruct struct {
	Name string `param:"name"`
}

type TestParamIntStruct struct {
	Age int `param:"age"`
}

type TestParamBoolStruct struct {
	IsActive bool `param:"is_active"`
}

type TestMultiParamStruct struct {
	Name     string `param:"name"`
	Age      int    `param:"age"`
	IsActive bool   `param:"is_active"`
}

type TestNoParamStruct struct {
	Name string
}

//nolint:nestif // Test function
func TestMcpWith(t *testing.T) {
	tests := []struct {
		name          string
		paramName     string
		expectedFunc  string
		structType    any
		propertyOpts  []mcp.PropertyOption
		expectNoMatch bool
	}{
		{
			name:         "string parameter",
			paramName:    "name",
			expectedFunc: "WithString",
			structType:   TestParamStringStruct{},
			propertyOpts: []mcp.PropertyOption{
				mcp.DefaultString("default"),
				mcp.Description("description"),
			},
		},
		{
			name:         "int parameter",
			paramName:    "age",
			expectedFunc: "WithNumber",
			structType:   TestParamIntStruct{},
			propertyOpts: []mcp.PropertyOption{
				mcp.DefaultNumber(0),
				mcp.Description("age description"),
			},
		},
		{
			name:         "bool parameter",
			paramName:    "is_active",
			expectedFunc: "WithBoolean",
			structType:   TestParamBoolStruct{},
			propertyOpts: []mcp.PropertyOption{
				mcp.DefaultBool(false),
				mcp.Description("active status"),
			},
		},
		{
			name:         "multi param struct with string match",
			paramName:    "name",
			expectedFunc: "WithString",
			structType:   TestMultiParamStruct{},
			propertyOpts: []mcp.PropertyOption{
				mcp.Description("name description"),
			},
		},
		{
			name:         "multi param struct with int match",
			paramName:    "age",
			expectedFunc: "WithNumber",
			structType:   TestMultiParamStruct{},
			propertyOpts: []mcp.PropertyOption{
				mcp.Description("age description"),
			},
		},
		{
			name:         "multi param struct with bool match",
			paramName:    "is_active",
			expectedFunc: "WithBoolean",
			structType:   TestMultiParamStruct{},
			propertyOpts: []mcp.PropertyOption{
				mcp.Description("active status"),
			},
		},
		{
			name:          "no param in struct",
			paramName:     "name",
			expectedFunc:  "",
			structType:    TestNoParamStruct{},
			expectNoMatch: true,
		},
		{
			name:          "param name not in struct",
			paramName:     "unknown",
			expectedFunc:  "",
			structType:    TestMultiParamStruct{},
			expectNoMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // Create a tool to apply the option to
			tool := &mcp.Tool{
				Name:        "test-tool",
				Description: "Test tool",
				InputSchema: mcp.ToolInputSchema{
					Type:       "object",
					Properties: make(map[string]any),
				},
			}

			// Convert the concrete type to the generic type and call mcpWith
			var toolOption mcp.ToolOption

			// Call mcpWith based on the concrete type
			switch tt.structType.(type) {
			case TestParamStringStruct:
				toolOption = mcpWith[TestParamStringStruct](tt.paramName, tt.propertyOpts...)
			case TestParamIntStruct:
				toolOption = mcpWith[TestParamIntStruct](tt.paramName, tt.propertyOpts...)
			case TestParamBoolStruct:
				toolOption = mcpWith[TestParamBoolStruct](tt.paramName, tt.propertyOpts...)
			case TestMultiParamStruct:
				toolOption = mcpWith[TestMultiParamStruct](tt.paramName, tt.propertyOpts...)
			case TestNoParamStruct:
				toolOption = mcpWith[TestNoParamStruct](tt.paramName, tt.propertyOpts...)
			}

			// Apply the tool option to our test tool
			toolOption(tool)

			if tt.expectNoMatch {
				// If we expect no match, the property should not be added to the tool
				if _, ok := tool.InputSchema.Properties[tt.paramName]; ok {
					t.Errorf("Expected no property to be added for %s, but property was added", tt.paramName)
				}
			} else {
				// Check that the property was added to the tool
				prop, ok := tool.InputSchema.Properties[tt.paramName]
				if !ok {
					t.Fatalf("Property %s was not added to tool schema", tt.paramName)
				}

				// Check the type of the property
				propMap, ok := prop.(map[string]any)
				if !ok {
					t.Fatalf("Property %s is not a map", tt.paramName)
				}

				// Verify type matches expected function
				var expectedType string
				switch tt.expectedFunc {
				case "WithString":
					expectedType = "string"
				case "WithNumber":
					expectedType = "number"
				case "WithBoolean":
					expectedType = "boolean"
				}

				propType, ok := propMap["type"].(string)
				if !ok {
					t.Fatalf("Property %s does not have a type", tt.paramName)
				}

				if propType != expectedType {
					t.Errorf("Property %s has type %s, expected %s", tt.paramName, propType, expectedType)
				}

				// Check that property options were applied
				if len(tt.propertyOpts) > 0 {
					// We can't check specific options directly, but we can check if properties
					// were added to the schema by verifying keys exist in the map
					if _, ok := propMap["description"]; tt.expectedFunc != "" && !ok {
						t.Errorf("Expected properties not found in schema for %s", tt.paramName)
					}
				}
			}
		})
	}
}

func TestDecodeArguments(t *testing.T) {
	tests := []struct {
		name     string
		args     map[string]any
		expected interface{}
	}{
		{
			name: "decode string",
			args: map[string]any{
				"name": "John",
			},
			expected: TestParamStringStruct{
				Name: "John",
			},
		},
		{
			name: "decode int",
			args: map[string]any{
				"age": 25,
			},
			expected: TestParamIntStruct{
				Age: 25,
			},
		},
		{
			name: "decode bool",
			args: map[string]any{
				"is_active": true,
			},
			expected: TestParamBoolStruct{
				IsActive: true,
			},
		},
		{
			name: "decode multiple fields",
			args: map[string]any{
				"name":      "John",
				"age":       30,
				"is_active": true,
			},
			expected: TestMultiParamStruct{
				Name:     "John",
				Age:      30,
				IsActive: true,
			},
		},
		{
			name: "decode with missing fields",
			args: map[string]any{
				"name": "John",
				// age and is_active missing
			},
			expected: TestMultiParamStruct{
				Name:     "John",
				Age:      0,     // default int value
				IsActive: false, // default bool value
			},
		},
		{
			name: "decode with string values for non-string types",
			args: map[string]any{
				"name":      "John",
				"age":       "25",   // string instead of int
				"is_active": "true", // string instead of bool
			},
			expected: TestMultiParamStruct{
				Name:     "John",
				Age:      25,   // should convert string to int
				IsActive: true, // should convert string to bool
			},
		},
		{
			name: "decode with no matching fields",
			args: map[string]any{
				"unknown": "value",
			},
			expected: TestMultiParamStruct{}, // all default values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			switch expected := tt.expected.(type) {
			case TestParamStringStruct:
				var decoded TestParamStringStruct
				decoded, err = decodeArguments[TestParamStringStruct](tt.args)
				if err != nil {
					t.Fatalf("Failed to decode arguments: %v", err)
				}
				if decoded != expected {
					t.Errorf("Expected %+v, got %+v", expected, decoded)
				}

			case TestParamIntStruct:
				var decoded TestParamIntStruct
				decoded, err = decodeArguments[TestParamIntStruct](tt.args)
				if err != nil {
					t.Fatalf("Failed to decode arguments: %v", err)
				}
				if decoded != expected {
					t.Errorf("Expected %+v, got %+v", expected, decoded)
				}

			case TestParamBoolStruct:
				var decoded TestParamBoolStruct
				decoded, err = decodeArguments[TestParamBoolStruct](tt.args)
				if err != nil {
					t.Fatalf("Failed to decode arguments: %v", err)
				}
				if decoded != expected {
					t.Errorf("Expected %+v, got %+v", expected, decoded)
				}

			case TestMultiParamStruct:
				var decoded TestMultiParamStruct
				decoded, err = decodeArguments[TestMultiParamStruct](tt.args)
				if err != nil {
					t.Fatalf("Failed to decode arguments: %v", err)
				}
				if decoded != expected {
					t.Errorf("Expected %+v, got %+v", expected, decoded)
				}
			}
		})
	}
}
