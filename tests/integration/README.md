# ZeroUI Integration Tests

This directory contains focused integration tests for ZeroUI's core functionality. These tests verify that components work together correctly without unnecessary complexity.

## Test Structure

### Core Test Files

- **`core_integration_test.go`** - Essential CLI and system integration tests
- **`plugin_integration_test.go`** - RPC plugin system integration tests  
- **`engine_integration_test.go`** - Configuration engine and file operation tests

### Test Categories

#### 1. Core Integration Tests (`TestCoreIntegration`)
- **CLI Commands**: `list`, `toggle`, `extract`, `--help`
- **Configuration Engine**: Config detection, parsing, validation
- **Plugin System**: Plugin discovery, RPC communication
- **Error Handling**: Invalid inputs, missing files, graceful degradation

#### 2. Plugin System Tests (`TestPluginSystemIntegration`)
- **Plugin Discovery**: PATH-based plugin detection
- **RPC Communication**: gRPC protocol verification
- **Plugin Lifecycle**: Start, stop, crash recovery
- **Plugin API**: Interface compliance testing

#### 3. Engine Integration Tests (`TestEngineIntegration`)
- **Configuration Detection**: Multi-format config file support
- **Toggle Operations**: Field modification and validation
- **File Operations**: Safe config writing with backups
- **Validation Engine**: Schema and type validation

## Running Tests

### Quick Start
```bash
# Run core integration tests (essential functionality)
make test-integration

# Run all integration tests
make test-integration-all

# Run only essential tests (fast)
make test-integration-short
```

### Specific Test Suites
```bash
# Test CLI commands only
make test-cli

# Test configuration engine only  
make test-engine

# Test plugin system only
make test-plugins

# Test TUI interface
make test-integration-tui
```

### Development Workflow
```bash
# Watch for changes and run tests
make test-watch

# Run with coverage reporting
make test-integration-coverage

# Run tests in parallel (faster)
make test-integration-parallel
```

## Test Philosophy

### Focus on Core Functionality
- **Happy Path Testing**: Verify essential workflows work correctly
- **Critical Error Handling**: Test important failure scenarios gracefully
- **No Edge Case Bloat**: Avoid testing unlikely or complex edge cases

### Essential Test Scenarios

#### CLI Integration
- ✅ `zeroui list apps` shows available applications
- ✅ `zeroui toggle app field value --dry-run` processes requests
- ✅ `zeroui extract app --dry-run` extracts configuration
- ✅ Invalid commands show helpful error messages

#### Configuration Engine
- ✅ Detects existing configuration files
- ✅ Handles missing configuration gracefully
- ✅ Validates field names and values
- ✅ Preserves files in dry-run mode

#### Plugin System
- ✅ Discovers plugins in PATH
- ✅ Communicates via RPC protocol
- ✅ Handles plugin crashes gracefully
- ✅ Manages plugin lifecycle correctly

#### Error Recovery
- ✅ Missing configuration files → helpful guidance
- ✅ Plugin communication failure → fallback behavior
- ✅ Invalid field values → clear error messages
- ✅ File permission issues → graceful degradation

## Test Environment

### Isolated Testing
- Each test creates temporary directories
- No modification of user's actual config files
- Plugin tests use isolated plugin directories
- Cleanup after each test completion

### Test Dependencies
- **Go testing framework**: Standard `testing` package
- **Testify**: Assertions and test structure (`github.com/stretchr/testify`)
- **Real binaries**: Tests build and use actual ZeroUI binary
- **No mocks**: Integration tests use real components

### Build Requirements
```bash
# Ensure ZeroUI builds successfully
go build -o build/zeroui .

# Ensure plugin builds successfully
cd plugins/ghostty-rpc && go build -o zeroui-plugin-ghostty-rpc
```

## Test Data

### Sample Configurations
Tests create realistic configuration files:

```bash
# Ghostty config
~/.config/ghostty/config

# Alacritty config  
~/.config/alacritty/alacritty.yml
```

### Plugin Environment
Tests set up isolated plugin directories:

```bash
/tmp/test-plugins/
├── zeroui-plugin-ghostty-rpc
└── zeroui-plugin-test (for error testing)
```

## Expected Results

### Successful Test Run
```
🧪 Running core integration tests...
=== RUN   TestCoreIntegration
=== RUN   TestCoreIntegration/CLI_Core_Functionality
=== RUN   TestCoreIntegration/Configuration_Engine_Core
=== RUN   TestCoreIntegration/Plugin_System_Core
=== RUN   TestCoreIntegration/Error_Handling_Core
--- PASS: TestCoreIntegration (2.34s)
PASS
```

### What Tests Verify
- ✅ All core CLI commands execute without crashing
- ✅ Configuration files are detected and parsed correctly
- ✅ Plugin system initializes and communicates successfully
- ✅ Error conditions are handled gracefully without panics
- ✅ Dry-run mode preserves original files
- ✅ Help and usage information is displayed correctly

## Troubleshooting

### Common Issues

**Plugin Tests Fail**
- Ensure `plugins/ghostty-rpc` builds successfully
- Check that plugin binary has execute permissions
- Verify plugin is discoverable in test PATH

**Configuration Tests Fail**
- Check file permissions in test directories
- Ensure sample config files are created correctly
- Verify HOME environment variable in tests

**TUI Tests Hang**
- TUI tests have timeouts to prevent hanging
- Set `testing.Short()` to skip interactive tests
- Check terminal compatibility for headless testing

### Debug Mode
```bash
# Run with verbose output
make test-integration-verbose

# Check test environment
make test-env

# Clean test artifacts
make clean-test
```

## Contributing

### Adding New Tests
1. **Focus on core functionality** - avoid edge cases
2. **Use real components** - no mocks in integration tests
3. **Clean up after tests** - remove temporary files
4. **Test error handling** - verify graceful degradation
5. **Keep tests fast** - aim for sub-5-minute execution

### Test Naming Convention
```go
func TestCoreIntegration(t *testing.T) {
    t.Run("CLI Core Functionality", func(t *testing.T) {
        // Test essential CLI operations
    })
}
```

### Best Practices
- Use `require` for critical setup operations
- Use `assert` for test assertions
- Create isolated test environments
- Test both success and failure scenarios
- Focus on user-facing functionality
- Avoid testing internal implementation details