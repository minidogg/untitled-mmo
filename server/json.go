package main

import (
	"encoding/json"
	"os"
)

func WriteJSONFile[T any](filename string, data T) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func ReadJSONFile[T any](filename string) (T, error) {
	var data T

	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}
