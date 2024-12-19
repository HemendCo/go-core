package helpers

import (
	"encoding/json"
	"fmt"
)

func CastValue[T any](value interface{}) T {
	var result T
	if v, ok := value.(T); ok {
		return v
	}
	return result
}

func DecodeJSON(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}

func StringInArray(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}
