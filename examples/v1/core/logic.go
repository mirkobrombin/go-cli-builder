package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Data represents the structure of the data saved in the JSON file.
type Data struct {
	Items []string `json:"items"`
}

// GetDataPath returns the path of the JSON storage file.
func GetDataPath() string {
	return filepath.Join(os.TempDir(), "mycli_data.json")
}

// LoadData loads the data from the JSON file.
// If the file does not exist, it returns empty data without an error.
// If there is an error reading or unmarshalling the file, it returns an error.
func LoadData() (Data, error) {
	var data Data
	filePath := GetDataPath()

	file, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Data{}, nil // File doesn't exist, return empty data
		}
		return Data{}, fmt.Errorf("error reading data file: %w", err)
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		return Data{}, fmt.Errorf("error unmarshalling data: %w", err)
	}

	return data, nil
}

// SaveData saves the data to the JSON file.
// If there is an error marshalling or writing the file, it returns an error.
func SaveData(data Data) error {
	filePath := GetDataPath()

	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling data: %w", err)
	}

	err = os.WriteFile(filePath, file, 0644)
	if err != nil {
		return fmt.Errorf("error writing data file: %w", err)
	}

	return nil
}
