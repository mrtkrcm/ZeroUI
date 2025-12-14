package errors

import (
	"fmt"
	"os"
	"path/filepath"
)

// HandleFileError wraps file system errors with context
func HandleFileError(err error, path string, operation string) error {
	if err == nil {
		return nil
	}

	// Check for specific file system errors
	if os.IsNotExist(err) {
		return New(ConfigNotFound, fmt.Sprintf("file not found: %s", path)).
			WithSuggestions(
				"Check if the file path is correct",
				"Ensure the file exists",
				fmt.Sprintf("Create the file at: %s", path),
			)
	}

	if os.IsPermission(err) {
		return NewPermissionError(path, operation)
	}

	// Generic file error
	return Wrap(SystemFileError, fmt.Sprintf("file operation failed: %s", operation), err)
}

// HandleConfigError wraps configuration errors with helpful context
func HandleConfigError(err error, app, path string) error {
	if err == nil {
		return nil
	}

	// Check if it's already a ZeroUIError
	if zErr, ok := GetZeroUIError(err); ok {
		return zErr.WithApp(app)
	}

	// Check for file errors
	if os.IsNotExist(err) {
		return NewConfigNotFoundError(path).WithApp(app)
	}

	if os.IsPermission(err) {
		return NewPermissionError(path, "read").WithApp(app)
	}

	// Generic config error
	return Wrap(ConfigParseError, fmt.Sprintf("failed to load config for %s", app), err)
}

// HandlePluginError wraps plugin errors with context
func HandlePluginError(err error, plugin string, operation string) error {
	if err == nil {
		return nil
	}

	return Wrap(PluginError, fmt.Sprintf("plugin %s failed during %s", plugin, operation), err).
		WithSuggestions(
			fmt.Sprintf("Check if the plugin '%s' is installed correctly", plugin),
			"Try reinstalling the plugin",
			"Check plugin logs for more details",
		)
}

// HandleValidationError wraps validation errors with field context
func HandleValidationError(err error, app, field string, value interface{}) error {
	if err == nil {
		return nil
	}

	valueStr := fmt.Sprintf("%v", value)
	return Wrap(ValidationError, fmt.Sprintf("validation failed for field '%s'", field), err).
		WithApp(app).
		WithField(field).
		WithValue(valueStr)
}

// Must panics if error is not nil (use sparingly, mainly for initialization)
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustGet returns the value or panics if error is not nil
func MustGet[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

// WrapIfError wraps an error only if it's not nil
func WrapIfError(err error, errorType ErrorType, message string) error {
	if err == nil {
		return nil
	}
	return Wrap(errorType, message, err)
}

// FirstError returns the first non-nil error from a list
func FirstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// CombineErrors combines multiple errors into a single error
func CombineErrors(errs ...error) error {
	var nonNilErrs []error
	for _, err := range errs {
		if err != nil {
			nonNilErrs = append(nonNilErrs, err)
		}
	}

	switch len(nonNilErrs) {
	case 0:
		return nil
	case 1:
		return nonNilErrs[0]
	default:
		messages := make([]string, len(nonNilErrs))
		for i, err := range nonNilErrs {
			messages[i] = err.Error()
		}
		return New(SystemCommand, fmt.Sprintf("multiple errors occurred: %v", messages))
	}
}

// RecoverAsError recovers from a panic and returns it as an error
func RecoverAsError() error {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case error:
			return Wrap(SystemCommand, "panic recovered", v)
		default:
			return New(SystemCommand, fmt.Sprintf("panic recovered: %v", v))
		}
	}
	return nil
}

// SafeExecute executes a function and recovers from panics
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = Wrap(SystemCommand, "execution failed", v)
			default:
				err = New(SystemCommand, fmt.Sprintf("execution failed: %v", v))
			}
		}
	}()
	return fn()
}

// IsFileNotFound checks if an error is a file not found error
func IsFileNotFound(err error) bool {
	if err == nil {
		return false
	}

	// Check OS error
	if os.IsNotExist(err) {
		return true
	}

	// Check our error type
	if zErr, ok := GetZeroUIError(err); ok {
		return zErr.Type == ConfigNotFound || zErr.Type == SystemFileError
	}

	return false
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	if err == nil {
		return false
	}

	// Check OS error
	if os.IsPermission(err) {
		return true
	}

	// Check our error type
	if zErr, ok := GetZeroUIError(err); ok {
		return zErr.Type == SystemPermission || zErr.Type == ConfigPermission
	}

	return false
}

// GetConfigDir returns the configuration directory for an app
func GetConfigDir(app string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", app)
}

// GetConfigPath returns the default configuration path for an app
func GetConfigPath(app string) string {
	return filepath.Join(GetConfigDir(app), "config.yml")
}
