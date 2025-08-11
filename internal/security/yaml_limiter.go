package security

import (
	"fmt"
	"io"
	"os"
	"time"
)

// YAMLLimits defines security limits for YAML parsing
type YAMLLimits struct {
	MaxFileSize    int64         // Maximum file size in bytes
	MaxDepth       int           // Maximum nesting depth
	MaxKeys        int           // Maximum number of keys
	ParseTimeout   time.Duration // Maximum parse time
	MaxMemoryUsage int64         // Maximum memory usage in bytes
}

// DefaultYAMLLimits returns sensible default limits for YAML parsing
func DefaultYAMLLimits() *YAMLLimits {
	return &YAMLLimits{
		MaxFileSize:    10 * 1024 * 1024,  // 10MB
		MaxDepth:       50,                // 50 levels deep
		MaxKeys:        10000,             // 10,000 keys
		ParseTimeout:   30 * time.Second,  // 30 second timeout
		MaxMemoryUsage: 100 * 1024 * 1024, // 100MB memory limit
	}
}

// YAMLValidator provides secure YAML validation and parsing
type YAMLValidator struct {
	limits *YAMLLimits
}

// NewYAMLValidator creates a new YAML validator with specified limits
func NewYAMLValidator(limits *YAMLLimits) *YAMLValidator {
	if limits == nil {
		limits = DefaultYAMLLimits()
	}
	return &YAMLValidator{
		limits: limits,
	}
}

// ValidateFile validates a YAML file before parsing
func (v *YAMLValidator) ValidateFile(filePath string) error {
	// Check file size first (prevents loading huge files)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.Size() > v.limits.MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum allowed size of %d bytes",
			fileInfo.Size(), v.limits.MaxFileSize)
	}

	return nil
}

// ValidateContent validates YAML content for complexity and resource usage
func (v *YAMLValidator) ValidateContent(content []byte) error {
	// Check content size
	if int64(len(content)) > v.limits.MaxFileSize {
		return fmt.Errorf("content size %d bytes exceeds maximum allowed size of %d bytes",
			len(content), v.limits.MaxFileSize)
	}

	// Basic structure validation to prevent YAML bombs
	if err := v.validateStructuralComplexity(content); err != nil {
		return fmt.Errorf("YAML structural validation failed: %w", err)
	}

	return nil
}

// validateStructuralComplexity performs basic validation to detect potential YAML bombs
func (v *YAMLValidator) validateStructuralComplexity(content []byte) error {
	depth := 0
	maxDepthSeen := 0
	keyCount := 0
	inString := false
	escaped := false

	for i := 0; i < len(content); i++ {
		char := content[i]

		// Handle string content (skip structural analysis inside strings)
		if char == '"' && !escaped {
			inString = !inString
			continue
		}

		if inString {
			escaped = (char == '\\' && !escaped)
			continue
		}

		escaped = false

		switch char {
		case '{', '[':
			// Opening braces/brackets increase depth
			depth++
			if depth > maxDepthSeen {
				maxDepthSeen = depth
			}

			// Check depth limit
			if depth > v.limits.MaxDepth {
				return fmt.Errorf("nesting depth %d exceeds maximum allowed depth of %d",
					depth, v.limits.MaxDepth)
			}

		case '}', ']':
			// Closing braces/brackets decrease depth
			if depth > 0 {
				depth--
			}

		case ':':
			// Count potential keys (colon usually indicates key-value pair)
			if !inString {
				keyCount++
				if keyCount > v.limits.MaxKeys {
					return fmt.Errorf("key count %d exceeds maximum allowed keys of %d",
						keyCount, v.limits.MaxKeys)
				}
			}
		}
	}

	return nil
}

// SafeReadFile reads and validates a file with size and timeout limits
func (v *YAMLValidator) SafeReadFile(filePath string) ([]byte, error) {
	// Pre-validate file
	if err := v.ValidateFile(filePath); err != nil {
		return nil, err
	}

	// Create a channel for the result
	type readResult struct {
		data []byte
		err  error
	}

	resultChan := make(chan readResult, 1)

	// Read file with timeout
	go func() {
		data, err := os.ReadFile(filePath)
		resultChan <- readResult{data: data, err: err}
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			return nil, fmt.Errorf("failed to read file: %w", result.err)
		}

		// Validate content before returning
		if err := v.ValidateContent(result.data); err != nil {
			return nil, err
		}

		return result.data, nil

	case <-time.After(v.limits.ParseTimeout):
		return nil, fmt.Errorf("file read timeout after %v", v.limits.ParseTimeout)
	}
}

// LimitedReader creates an io.Reader with size limits
func (v *YAMLValidator) LimitedReader(r io.Reader) io.Reader {
	return io.LimitReader(r, v.limits.MaxFileSize)
}

// GetLimits returns the current limits (for testing/debugging)
func (v *YAMLValidator) GetLimits() *YAMLLimits {
	return v.limits
}
