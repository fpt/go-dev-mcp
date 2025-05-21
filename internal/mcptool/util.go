package tool

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
)

func mcpWith[T any](name string, opts ...mcp.PropertyOption) mcp.ToolOption {
	// Usage: mcpWith[ToolArguments]("arg_name", mcp.DefaultString("default"), mcp.Description("description"))
	// arg_name: argument name
	// default: default value of the parameter
	// description: description of the parameter
	var params T

	// Use reflection to get the type of the struct
	val := reflect.ValueOf(params)
	typ := val.Type()

	// If params is not a struct, there are no parameters to add
	if typ.Kind() != reflect.Struct {
		return func(t *mcp.Tool) {}
	}

	// Iterate through all fields of the struct
	for i := range typ.NumField() {
		field := typ.Field(i)

		// Get the param tag value, which contains the parameter name
		paramTag := field.Tag.Get("param")
		if paramTag == "" {
			continue // Skip fields without param tag
		}

		if name == paramTag {
			switch field.Type.Kind() {
			case reflect.String:
				// Add string parameter
				return mcp.WithString(paramTag, opts...)
			case reflect.Int:
				// Add number parameter for int
				return mcp.WithNumber(paramTag, opts...)
			case reflect.Bool:
				// Add boolean parameter
				return mcp.WithBoolean(paramTag, opts...)
			default:
				// Handle unsupported types
				// You can choose to log or return an error here
				fmt.Printf("Unsupported field type: %s\n", field.Type.Kind())
				continue
			}
		}
	}

	// If no matching parameter was found, return an empty ToolOption
	return func(t *mcp.Tool) {}
}

func decodeArguments[T any](args map[string]any) (T, error) {
	var params T

	// Use reflection to get the type and value of the struct
	val := reflect.ValueOf(&params).Elem()
	typ := val.Type()

	// Iterate through all fields of the struct
	for i := range typ.NumField() {
		field := typ.Field(i)

		// Get the param tag value, which contains the argument name
		paramTag := field.Tag.Get("param")
		if paramTag == "" {
			continue // Skip fields without param tag
		}

		// Check if the argument exists in the map
		if argVal, ok := args[paramTag]; ok {
			fieldVal := val.Field(i)

			// Make sure the field is settable
			if !fieldVal.CanSet() {
				continue
			}

			// Set the field value based on its type
			switch fieldVal.Kind() {
			case reflect.String:
				if strVal, ok := argVal.(string); ok {
					fieldVal.SetString(strVal)
				}
			case reflect.Int:
				switch v := argVal.(type) {
				case int:
					fieldVal.SetInt(int64(v))
				case string:
					if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
						fieldVal.SetInt(intVal)
					}
				}
			case reflect.Bool:
				switch v := argVal.(type) {
				case bool:
					fieldVal.SetBool(v)
				case string:
					if boolVal, err := strconv.ParseBool(v); err == nil {
						fieldVal.SetBool(boolVal)
					}
				}
			default:
				return params, fmt.Errorf("unsupported field type: %s", fieldVal.Kind())
			}
		}
	}

	return params, nil
}
