package appconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
)

// IntegrityChecker provides file integrity verification
type IntegrityChecker struct {
	checksums map[string]string
}

// NewIntegrityChecker creates a new integrity checker
func NewIntegrityChecker() *IntegrityChecker {
	return &IntegrityChecker{
		checksums: make(map[string]string),
	}
}

// CalculateChecksum calculates SHA-256 checksum of a file
func (ic *IntegrityChecker) CalculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	ic.checksums[path] = checksum
	return checksum, nil
}

// VerifyChecksum verifies file against stored checksum
func (ic *IntegrityChecker) VerifyChecksum(path string, expectedChecksum string) (bool, error) {
	actualChecksum, err := ic.CalculateChecksum(path)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}

// ValidateFormat validates the format of a configuration file
func (ic *IntegrityChecker) ValidateFormat(path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Empty files are valid (new configs)
	if len(data) == 0 {
		return nil
	}

	switch ext {
	case ".json":
		return ic.validateJSON(data)
	case ".yaml", ".yml":
		return ic.validateYAML(data)
	case ".toml":
		return ic.validateTOML(data)
	case ".conf", ".config":
		// For custom formats, just check it's not binary
		return ic.validateText(data)
	default:
		// Unknown format, check if it's text
		return ic.validateText(data)
	}
}

// ValidateContent performs content-level validation
func (ic *IntegrityChecker) ValidateContent(path string, schema interface{}) error {
	// This would implement schema validation
	// For now, just ensure file is readable
	_, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("content validation failed: %w", err)
	}
	return nil
}

// CompareFiles compares two files for differences
func (ic *IntegrityChecker) CompareFiles(file1, file2 string) (bool, error) {
	checksum1, err := ic.CalculateChecksum(file1)
	if err != nil {
		return false, fmt.Errorf("failed to checksum file1: %w", err)
	}

	checksum2, err := ic.CalculateChecksum(file2)
	if err != nil {
		return false, fmt.Errorf("failed to checksum file2: %w", err)
	}

	return checksum1 == checksum2, nil
}

// validateJSON validates JSON format
func (ic *IntegrityChecker) validateJSON(data []byte) error {
	var js interface{}
	if err := json.Unmarshal(data, &js); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}
	return nil
}

// validateYAML validates YAML format
func (ic *IntegrityChecker) validateYAML(data []byte) error {
	parser := yaml.Parser()
	_, err := parser.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}
	return nil
}

// validateTOML validates TOML format
func (ic *IntegrityChecker) validateTOML(data []byte) error {
	parser := toml.Parser()
	_, err := parser.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("invalid TOML format: %w", err)
	}
	return nil
}

// validateText validates that content is valid text (not binary)
func (ic *IntegrityChecker) validateText(data []byte) error {
	// Check for null bytes which indicate binary content
	for i, b := range data {
		if b == 0 {
			return fmt.Errorf("binary content detected at byte %d", i)
		}
		// Check for other non-text bytes (excluding common whitespace)
		if b < 32 && b != '\t' && b != '\n' && b != '\r' {
			return fmt.Errorf("non-text character detected at byte %d", i)
		}
	}
	return nil
}

// CreateIntegrityReport generates a report of file integrity
func (ic *IntegrityChecker) CreateIntegrityReport(path string) (*IntegrityReport, error) {
	report := &IntegrityReport{
		FilePath:  path,
		Timestamp: fmt.Sprintf("%d", time.Now().Unix()),
	}

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		report.Error = err.Error()
		return report, nil
	}

	report.FileSize = info.Size()
	report.ModTime = info.ModTime().String()
	report.Permissions = info.Mode().String()

	// Calculate checksum
	checksum, err := ic.CalculateChecksum(path)
	if err != nil {
		report.Error = err.Error()
		return report, nil
	}
	report.Checksum = checksum

	// Validate format
	if err := ic.ValidateFormat(path); err != nil {
		report.FormatValid = false
		report.FormatError = err.Error()
	} else {
		report.FormatValid = true
	}

	report.Valid = report.FormatValid && report.Error == ""
	return report, nil
}

// IntegrityReport contains integrity check results
type IntegrityReport struct {
	FilePath    string `json:"file_path"`
	Checksum    string `json:"checksum"`
	FileSize    int64  `json:"file_size"`
	ModTime     string `json:"mod_time"`
	Permissions string `json:"permissions"`
	FormatValid bool   `json:"format_valid"`
	FormatError string `json:"format_error,omitempty"`
	Valid       bool   `json:"valid"`
	Error       string `json:"error,omitempty"`
	Timestamp   string `json:"timestamp"`
}
