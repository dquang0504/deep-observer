package validation

import "github.com/santhosh-tekuri/jsonschema/v5"

type ValidationError struct{
	Message string `json:"message"`
}

func FormatSchemaError(err error) any{
	if ve, ok := err.(*jsonschema.ValidationError); ok{
		return map[string]any{
			"message": ve.Message,
			"key": ve.KeywordLocation,
			"instanceLocation": ve.InstanceLocation,
			"schemaLocation": ve.AbsoluteKeywordLocation,
		}
	}
	return map[string]any{"message": err.Error()}
}
