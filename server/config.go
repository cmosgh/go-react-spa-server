package server

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration.
type Config struct {
	StaticDir string `json:"static_dir"`
}

// LoadConfig loads the configuration from a .go-spa-server-config.json file.
// It searches for the file in the current working directory.
func LoadConfig() (*Config, error) {
	configPath := ".go-spa-server-config.json"
	
	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil // File does not exist, return nil config and no error
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
