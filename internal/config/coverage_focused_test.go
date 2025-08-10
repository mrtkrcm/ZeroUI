package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestLoader_CoverageEnhancement(t *testing.T) {
	// Test NewLoader error path
	originalHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	
	_, err := NewLoader()
	assert.Error(t, err, "Should fail when HOME is not set")
	
	// Restore HOME
	os.Setenv("HOME", originalHome)
}

func TestLoader_ConfigDirectoryAccess(t *testing.T) {
	loader, err := NewLoader()
	require.NoError(t, err)

	// Test ListApps with empty directory  
	apps, err := loader.ListApps()
	require.NoError(t, err)
	assert.IsType(t, []string{}, apps)

	// Test LoadAppConfig with non-existent app
	_, err = loader.LoadAppConfig("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app config not found")
}

func TestLoader_TargetConfigFormats(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-format-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	loader, err := NewLoader()
	require.NoError(t, err)
	loader.SetConfigDir(tmpDir)

	testCases := []struct {
		name    string
		format  string
		content string
		setup   func(string) error
		verify  func(*testing.T, *koanf.Koanf)
	}{
		{
			name:    "JSON_Format",
			format:  "json",
			content: `{"theme": "dark", "size": 14, "enabled": true}`,
			setup: func(path string) error {
				return ioutil.WriteFile(path, []byte(`{"theme": "dark", "size": 14, "enabled": true}`), 0644)
			},
			verify: func(t *testing.T, k *koanf.Koanf) {
				assert.Equal(t, "dark", k.String("theme"))
				assert.Equal(t, float64(14), k.Float64("size"))
				assert.True(t, k.Bool("enabled"))
			},
		},
		{
			name:    "YAML_Format",
			format:  "yaml",
			content: `theme: light\nsize: 16\nenabled: false`,
			setup: func(path string) error {
				return ioutil.WriteFile(path, []byte("theme: light\nsize: 16\nenabled: false"), 0644)
			},
			verify: func(t *testing.T, k *koanf.Koanf) {
				assert.Equal(t, "light", k.String("theme"))
				assert.Equal(t, float64(16), k.Float64("size"))
				assert.False(t, k.Bool("enabled"))
			},
		},
		{
			name:    "TOML_Format",
			format:  "toml",
			content: `theme = "auto"\nsize = 18\nenabled = true`,
			setup: func(path string) error {
				return ioutil.WriteFile(path, []byte(`theme = "auto"`+"\n"+`size = 18`+"\n"+`enabled = true`), 0644)
			},
			verify: func(t *testing.T, k *koanf.Koanf) {
				assert.Equal(t, "auto", k.String("theme"))
				assert.Equal(t, float64(18), k.Float64("size"))
				assert.True(t, k.Bool("enabled"))
			},
		},
		{
			name:   "Custom_Format",
			format: "custom",
			setup: func(path string) error {
				content := `theme = dark
font-family = Monaco
font-size = 14
window-decoration = false
keybind = shift+super+k=reload_config
keybind = super+equal=increase_font_size`
				return ioutil.WriteFile(path, []byte(content), 0644)
			},
			verify: func(t *testing.T, k *koanf.Koanf) {
				assert.Equal(t, "dark", k.String("theme"))
				assert.Equal(t, "Monaco", k.String("font-family"))
				assert.Equal(t, "14", k.String("font-size"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configFile := filepath.Join(tmpDir, "test_"+tc.name+".conf")

			// Setup test file
			err := tc.setup(configFile)
			require.NoError(t, err)

			// Create app config
			appConfig := &AppConfig{
				Name:   "test-app",
				Path:   configFile,
				Format: tc.format,
				Fields: map[string]FieldConfig{
					"theme": {Type: "string"},
					"size":  {Type: "number"},
				},
			}

			// Test loading
			data, err := loader.LoadTargetConfig(appConfig)
			require.NoError(t, err)

			// Verify data
			tc.verify(t, data)

			// Test saving (create new koanf instance)
			newData := koanf.New(".")
			newData.Set("theme", "updated")
			newData.Set("new_setting", "test_value")

			err = loader.SaveTargetConfig(appConfig, newData)
			require.NoError(t, err)

			// Verify saved data can be reloaded
			reloaded, err := loader.LoadTargetConfig(appConfig)
			require.NoError(t, err)
			assert.Equal(t, "updated", reloaded.String("theme"))
		})
	}
}

func TestLoader_PathExpansionCoverage(t *testing.T) {
	loader, err := NewLoader()
	require.NoError(t, err)

	// Test with absolute path (no expansion needed)
	tmpDir, err := ioutil.TempDir("", "configtoggle-path-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	configFile := filepath.Join(tmpDir, "absolute.json")
	configData := `{"key": "absolute_value"}`
	err = ioutil.WriteFile(configFile, []byte(configData), 0644)
	require.NoError(t, err)

	appConfig := &AppConfig{
		Name:   "path-test",
		Path:   configFile, // Absolute path
		Format: "json",
	}

	data, err := loader.LoadTargetConfig(appConfig)
	require.NoError(t, err)
	assert.Equal(t, "absolute_value", data.String("key"))

	// Test home directory expansion
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	homeConfigFile := filepath.Join(home, "test_home_config.json")
	err = ioutil.WriteFile(homeConfigFile, []byte(`{"key": "home_value"}`), 0644)
	require.NoError(t, err)
	defer os.Remove(homeConfigFile)

	homeAppConfig := &AppConfig{
		Name:   "home-test",
		Path:   "~/test_home_config.json", // Tilde expansion
		Format: "json",
	}

	homeData, err := loader.LoadTargetConfig(homeAppConfig)
	require.NoError(t, err)
	assert.Equal(t, "home_value", homeData.String("key"))
}

func TestLoader_ErrorConditions(t *testing.T) {
	loader, err := NewLoader()
	require.NoError(t, err)

	tmpDir, err := ioutil.TempDir("", "configtoggle-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	loader.SetConfigDir(tmpDir)

	t.Run("UnsupportedFormat", func(t *testing.T) {
		configFile := filepath.Join(tmpDir, "unsupported.xml")
		err = ioutil.WriteFile(configFile, []byte(`<config>test</config>`), 0644)
		require.NoError(t, err)

		appConfig := &AppConfig{
			Name:   "unsupported-format",
			Path:   configFile,
			Format: "xml", // Unsupported format
		}

		_, err := loader.LoadTargetConfig(appConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported config format")
	})

	t.Run("InvalidJSONFile", func(t *testing.T) {
		invalidJsonFile := filepath.Join(tmpDir, "invalid.json")
		err = ioutil.WriteFile(invalidJsonFile, []byte(`{"invalid": json syntax`), 0644)
		require.NoError(t, err)

		appConfig := &AppConfig{
			Name:   "invalid-json",
			Path:   invalidJsonFile,
			Format: "json",
		}

		_, err := loader.LoadTargetConfig(appConfig)
		assert.Error(t, err)
	})

	t.Run("InvalidYAMLFile", func(t *testing.T) {
		invalidYamlFile := filepath.Join(tmpDir, "invalid.yaml")
		err = ioutil.WriteFile(invalidYamlFile, []byte("invalid: yaml: [unclosed"), 0644)
		require.NoError(t, err)

		appConfig := &AppConfig{
			Name:   "invalid-yaml",
			Path:   invalidYamlFile,
			Format: "yaml",
		}

		_, err := loader.LoadTargetConfig(appConfig)
		assert.Error(t, err)
	})

	t.Run("InvalidTOMLFile", func(t *testing.T) {
		invalidTomlFile := filepath.Join(tmpDir, "invalid.toml")
		err = ioutil.WriteFile(invalidTomlFile, []byte("[invalid toml syntax"), 0644)
		require.NoError(t, err)

		appConfig := &AppConfig{
			Name:   "invalid-toml",
			Path:   invalidTomlFile,
			Format: "toml",
		}

		_, err := loader.LoadTargetConfig(appConfig)
		assert.Error(t, err)
	})
}

func TestLoader_SaveTargetConfigCoverage(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-save-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	loader, err := NewLoader()
	require.NoError(t, err)
	loader.SetConfigDir(tmpDir)

	// Test saving different formats
	formats := []struct {
		name   string
		format string
		verify func(t *testing.T, filePath string)
	}{
		{
			name:   "JSON_Save",
			format: "json",
			verify: func(t *testing.T, filePath string) {
				content, err := ioutil.ReadFile(filePath)
				require.NoError(t, err)
				
				var data map[string]interface{}
				err = json.Unmarshal(content, &data)
				require.NoError(t, err)
				
				assert.Equal(t, "test_value", data["test_key"])
				assert.Equal(t, float64(123), data["number_key"])
			},
		},
		{
			name:   "YAML_Save",
			format: "yaml",
			verify: func(t *testing.T, filePath string) {
				content, err := ioutil.ReadFile(filePath)
				require.NoError(t, err)
				
				var data map[string]interface{}
				err = yamlv3.Unmarshal(content, &data)
				require.NoError(t, err)
				
				assert.Equal(t, "test_value", data["test_key"])
				assert.Equal(t, 123, data["number_key"])
			},
		},
		{
			name:   "Custom_Save",
			format: "custom",
			verify: func(t *testing.T, filePath string) {
				content, err := ioutil.ReadFile(filePath)
				require.NoError(t, err)
				
				contentStr := string(content)
				assert.Contains(t, contentStr, "test_key = test_value")
				assert.Contains(t, contentStr, "number_key = 123")
			},
		},
	}

	for _, format := range formats {
		t.Run(format.name, func(t *testing.T) {
			configFile := filepath.Join(tmpDir, "save_test_"+format.name)
			
			appConfig := &AppConfig{
				Name:   "save-test",
				Path:   configFile,
				Format: format.format,
			}

			// Create data to save
			data := koanf.New(".")
			data.Set("test_key", "test_value")
			data.Set("number_key", 123)
			data.Set("bool_key", true)

			// Test saving
			err := loader.SaveTargetConfig(appConfig, data)
			require.NoError(t, err)

			// Verify file was created and has expected content
			assert.FileExists(t, configFile)
			format.verify(t, configFile)
		})
	}
}

func TestLoader_LoadAppConfigCoverage(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-load-app-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	loader, err := NewLoader()
	require.NoError(t, err)
	loader.SetConfigDir(tmpDir)

	// Create app config directory
	appsDir := filepath.Join(tmpDir, "apps")
	err = os.MkdirAll(appsDir, 0755)
	require.NoError(t, err)

	// Create a sample app config file
	appConfigPath := filepath.Join(appsDir, "test-app.yaml")
	appConfigContent := `name: test-app
path: /path/to/config.json
format: json
description: Test application
fields:
  theme:
    type: choice
    values: [dark, light]
    default: dark
    description: Application theme
  font_size:
    type: number
    default: 12
    description: Font size setting
`

	err = ioutil.WriteFile(appConfigPath, []byte(appConfigContent), 0644)
	require.NoError(t, err)

	// Test loading the app config
	appConfig, err := loader.LoadAppConfig("test-app")
	require.NoError(t, err)

	assert.Equal(t, "test-app", appConfig.Name)
	assert.Equal(t, "/path/to/config.json", appConfig.Path)
	assert.Equal(t, "json", appConfig.Format)
	assert.Equal(t, "Test application", appConfig.Description)

	// Verify fields
	assert.Contains(t, appConfig.Fields, "theme")
	assert.Contains(t, appConfig.Fields, "font_size")

	themeField := appConfig.Fields["theme"]
	assert.Equal(t, "choice", themeField.Type)
	assert.Equal(t, []string{"dark", "light"}, themeField.Values)
	assert.Equal(t, "dark", themeField.Default)

	fontSizeField := appConfig.Fields["font_size"]
	assert.Equal(t, "number", fontSizeField.Type)
	assert.Equal(t, 12, fontSizeField.Default)

	// Test ListApps includes this app
	apps, err := loader.ListApps()
	require.NoError(t, err)
	assert.Contains(t, apps, "test-app")
}

func TestLoader_InvalidAppConfigFile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "configtoggle-invalid-app-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	loader, err := NewLoader()
	require.NoError(t, err)
	loader.SetConfigDir(tmpDir)

	// Create app config directory
	appsDir := filepath.Join(tmpDir, "apps")
	err = os.MkdirAll(appsDir, 0755)
	require.NoError(t, err)

	// Create an invalid YAML app config file
	invalidAppConfigPath := filepath.Join(appsDir, "invalid-app.yaml")
	invalidContent := `name: invalid-app
path: /path/to/config.json
format: json
fields: [invalid yaml structure`

	err = ioutil.WriteFile(invalidAppConfigPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	// Test loading invalid app config
	_, err = loader.LoadAppConfig("invalid-app")
	assert.Error(t, err)
}