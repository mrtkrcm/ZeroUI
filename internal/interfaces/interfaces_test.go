package interfaces

import (
	"reflect"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/mrtkrcm/ZeroUI/internal/appconfig"
)

// mockConfigLoader is a mock implementation of ConfigLoader for testing
type mockConfigLoader struct{}

func (m *mockConfigLoader) LoadAppConfig(appName string) (*appconfig.AppConfig, error) {
	return &appconfig.AppConfig{Name: appName}, nil
}

func (m *mockConfigLoader) LoadTargetConfig(app *appconfig.AppConfig) (*koanf.Koanf, error) {
	return koanf.New("."), nil
}

func (m *mockConfigLoader) SaveTargetConfig(app *appconfig.AppConfig, target *koanf.Koanf) error {
	return nil
}

func (m *mockConfigLoader) ListApps() ([]string, error) {
	return []string{"app1", "app2"}, nil
}

func (m *mockConfigLoader) SetConfigDir(dir string) {
	// Mock implementation
}

// mockToggleEngine is a mock implementation of ToggleEngine for testing
type mockToggleEngine struct{}

func (m *mockToggleEngine) Toggle(appName, key, value string) error {
	return nil
}

func (m *mockToggleEngine) Cycle(appName, key string) error {
	return nil
}

func (m *mockToggleEngine) ApplyPreset(appName, presetName string) error {
	return nil
}

func (m *mockToggleEngine) GetAppConfig(appName string) (*appconfig.AppConfig, error) {
	return &appconfig.AppConfig{Name: appName}, nil
}

func (m *mockToggleEngine) GetCurrentValues(appName string) (map[string]interface{}, error) {
	return map[string]interface{}{"key": "value"}, nil
}

// mockLogger is a mock implementation of Logger for testing
type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields ...map[string]interface{}) {
	// Mock implementation
}

func (m *mockLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	// Mock implementation
}

func (m *mockLogger) Debug(msg string, fields ...map[string]interface{}) {
	// Mock implementation
}

func TestConfigLoaderInterface(t *testing.T) {
	// Test that mockConfigLoader implements ConfigLoader
	var _ ConfigLoader = &mockConfigLoader{}

	// Test that the interface has the expected methods
	loaderType := reflect.TypeOf((*ConfigLoader)(nil)).Elem()

	expectedMethods := []string{
		"LoadAppConfig",
		"LoadTargetConfig",
		"SaveTargetConfig",
		"ListApps",
		"SetConfigDir",
	}

	for _, methodName := range expectedMethods {
		method, found := loaderType.MethodByName(methodName)
		if !found {
			t.Errorf("ConfigLoader interface missing method: %s", methodName)
		} else {
			t.Logf("Found method: %s %s", methodName, method.Type)
		}
	}
}

func TestToggleEngineInterface(t *testing.T) {
	// Test that mockToggleEngine implements ToggleEngine
	var _ ToggleEngine = &mockToggleEngine{}

	// Test that the interface has the expected methods
	engineType := reflect.TypeOf((*ToggleEngine)(nil)).Elem()

	expectedMethods := []string{
		"Toggle",
		"Cycle",
		"ApplyPreset",
		"GetAppConfig",
		"GetCurrentValues",
	}

	for _, methodName := range expectedMethods {
		method, found := engineType.MethodByName(methodName)
		if !found {
			t.Errorf("ToggleEngine interface missing method: %s", methodName)
		} else {
			t.Logf("Found method: %s %s", methodName, method.Type)
		}
	}
}

func TestLoggerInterface(t *testing.T) {
	// Test that mockLogger implements Logger
	var _ Logger = &mockLogger{}

	// Test that the interface has the expected methods
	loggerType := reflect.TypeOf((*Logger)(nil)).Elem()

	expectedMethods := []string{
		"Info",
		"Error",
		"Debug",
	}

	for _, methodName := range expectedMethods {
		method, found := loggerType.MethodByName(methodName)
		if !found {
			t.Errorf("Logger interface missing method: %s", methodName)
		} else {
			t.Logf("Found method: %s %s", methodName, method.Type)
		}
	}
}

func TestInterfaceCompliance(t *testing.T) {
	// Test that all interfaces can be implemented
	var loader ConfigLoader = &mockConfigLoader{}
	var engine ToggleEngine = &mockToggleEngine{}
	var logger Logger = &mockLogger{}

	// Interfaces are implemented by concrete types
	_ = loader
	_ = engine
	_ = logger
}

func TestInterfaceMethodSignatures(t *testing.T) {
	// Test that interface methods have correct signatures
	loaderType := reflect.TypeOf((*ConfigLoader)(nil)).Elem()

	// Check LoadAppConfig signature
	loadAppMethod, _ := loaderType.MethodByName("LoadAppConfig")
	if loadAppMethod.Type.NumIn() != 1 { // appName (receiver is separate)
		t.Errorf("LoadAppConfig should have 1 input, got %d", loadAppMethod.Type.NumIn())
	}
	if loadAppMethod.Type.NumOut() != 2 { // (*config.AppConfig, error)
		t.Errorf("LoadAppConfig should have 2 outputs, got %d", loadAppMethod.Type.NumOut())
	}
}
