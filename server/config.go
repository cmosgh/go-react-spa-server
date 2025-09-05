package server

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Config represents the application configuration.
type Config struct {
	StaticDir     string `json:"static_dir"`
	SpaFallbackFile string `json:"spa_fallback_file"`
}

// LoadConfig loads the configuration from environment variables and a .go-spa-server-config.json file.
// Environment variables take precedence over the config file.
func LoadConfig() (*Config, error) {
	config := &Config{
		SpaFallbackFile: "index.html", // Default fallback file
	}

	// Load from config file if it exists
	configPath := ".go-spa-server-config.json"
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	if staticDirEnv := os.Getenv("STATIC_DIR"); staticDirEnv != "" {
		config.StaticDir = staticDirEnv
	}
	if spaFallbackFileEnv := os.Getenv("SPA_FALLBACK_FILE"); spaFallbackFileEnv != "" {
		config.SpaFallbackFile = spaFallbackFileEnv
	}

	// Basic validation for SpaFallbackFile
	if config.SpaFallbackFile == "" || strings.ContainsAny(config.SpaFallbackFile, "/\\") {

		
		
		return nil, fmt.Errorf("invalid SPA_FALLBACK_FILE: %s", config.SpaFallbackFile)
	}

	return config, nil
}