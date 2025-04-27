package helper

import (
	"encoding/json"
	"fmt"
)

func MapToStruct[T any](data map[string]interface{}) (*T, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal map to json: %w", err)
	}

	var result T
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("unmarshal json to struct: %w", err)
	}

	return &result, nil
}
