package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_FileExistsAndValid(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
	err := os.WriteFile(configPath, []byte(`{"static_dir": "./test_static"}`), 0644)
	assert.NoError(t, err)

	// Change to the temporary directory to simulate running from there
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "./test_static", config.StaticDir)
}

func TestLoadConfig_FileDoesNotExist(t *testing.T) {
	// Change to a temporary directory where the file won't exist
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err := os.Chdir(tempDir)
	assert.NoError(t, err)

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.Nil(t, config) // Expect nil config and no error
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create a temporary config file with invalid JSON
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
	err := os.WriteFile(configPath, []byte(`{"static_dir": "./test_static"`), 0644)
	assert.NoError(t, err)

	// Change to the temporary directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	config, err := LoadConfig()
	assert.Error(t, err) // Expect an error
	assert.Nil(t, config)
}
