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
	err := os.WriteFile(configPath, []byte(`{"static_dir": "./test_static", "spa_fallback_file": "app.html"}`), 0644)
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
	assert.Equal(t, "app.html", config.SpaFallbackFile)
}

func TestLoadConfig_FileDoesNotExist(t *testing.T) {
	// Change to a temporary directory where the file won't exist
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err := os.Chdir(tempDir)
	assert.NoError(t, err)

	config, err := LoadConfig()
	assert.NoError(t, err) // No error expected if file doesn't exist
	assert.NotNil(t, config)
	assert.Equal(t, "", config.StaticDir) // Should be empty if not set
	assert.Equal(t, "index.html", config.SpaFallbackFile) // Should be default
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

func TestLoadConfig_EnvVarPrecedence(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
	err := os.WriteFile(configPath, []byte(`{"static_dir": "./config_static", "spa_fallback_file": "config.html"}`), 0644)
	assert.NoError(t, err)

	// Set environment variables
	t.Setenv("STATIC_DIR", "./env_static")
	t.Setenv("SPA_FALLBACK_FILE", "env.html")

	// Change to the temporary directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "./env_static", config.StaticDir)
	assert.Equal(t, "env.html", config.SpaFallbackFile)
}

func TestLoadConfig_DefaultSpaFallbackFile(t *testing.T) {
	// Ensure no config file or env var is set
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err := os.Chdir(tempDir)
	assert.NoError(t, err)

	t.Setenv("SPA_FALLBACK_FILE", "") // Unset for this test

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "index.html", config.SpaFallbackFile)
}

func TestLoadConfig_InvalidSpaFallbackFile(t *testing.T) {
	// Test empty string in config file
	t.Run("empty string in config file", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
		err := os.WriteFile(configPath, []byte(`{"spa_fallback_file": ""}`), 0644)
		assert.NoError(t, err)

		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		err = os.Chdir(tempDir)
		assert.NoError(t, err)

		config, err := LoadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})

	// Test with path separator in config file
	t.Run("path separator in config file", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
		err := os.WriteFile(configPath, []byte(`{"spa_fallback_file": "path/to/file.html"}`), 0644)
		assert.NoError(t, err)

		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		err = os.Chdir(tempDir)
		assert.NoError(t, err)

		config, err := LoadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestLoadConfig_DefaultPort(t *testing.T) {
	// Ensure no config file or env var is set
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err := os.Chdir(tempDir) // First assignment of err in this scope
	assert.NoError(t, err)

	t.Setenv("PORT", "") // Unset for this test

	config, err := LoadConfig() // First assignment of config, subsequent assignment of err
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 8080, config.Port) // Should be default
}

func TestLoadConfig_PortFromEnvVar(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err := os.Chdir(tempDir) // First assignment of err in this scope
	assert.NoError(t, err)

	t.Setenv("PORT", "9000")

	config, err := LoadConfig() // First assignment of config, subsequent assignment of err
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 9000, config.Port);
}

func TestLoadConfig_InvalidPortEnvVar(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err := os.Chdir(tempDir) // First assignment of err in this scope
	assert.NoError(t, err)

	t.Setenv("PORT", "invalid")

	config, err := LoadConfig() // First assignment of config, subsequent assignment of err
	assert.Error(t, err) // Expect an error
	assert.Nil(t, config);
}

func TestLoadConfig_PortFromFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
	err := os.WriteFile(configPath, []byte(`{"port": 9090}`), 0644) // First assignment of err in this scope
	assert.NoError(t, err)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err = os.Chdir(tempDir) // Subsequent assignment of err
	assert.NoError(t, err)

	config, err := LoadConfig() // First assignment of config, subsequent assignment of err
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 9090, config.Port);
}

func TestLoadConfig_PortEnvVarPrecedence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".go-spa-server-config.json")
	err := os.WriteFile(configPath, []byte(`{"port": 9090}`), 0644) // First assignment of err in this scope
	assert.NoError(t, err)

	t.Setenv("PORT", "9000")

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	err = os.Chdir(tempDir) // Subsequent assignment of err
	assert.NoError(t, err)

	config, err := LoadConfig() // First assignment of config, subsequent assignment of err
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 9000, config.Port); // Env var should take precedence
}