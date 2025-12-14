# Changelog

All notable changes to ZeroUI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Consolidated historical improvement summaries into this changelog
- Moved archived documentation to changelog for better organization

## [Rebranding] - 2024

### Changed
- **Brand Identity**: Renamed from ConfigToggle to ZeroUI
  - New tagline: "Zero-configuration UI toolkit manager for developers"
  - Updated binary name from `configtoggle` to `zeroui`
  - Updated configuration paths from `~/.config/configtoggle/` to `~/.config/zeroui/`
- **Repository Migration**: Moved from public repository to private `mrtkrcm/ZeroUI`
- **Go Module Path**: Updated to `github.com/mrtkrcm/ZeroUI`
- **CLI Commands**: All help text, examples, and branding updated
- **Error Types**: Renamed error types (e.g., ConfigToggleError → ZeroUIError)
- **Documentation**: Complete rebranding of README, docs, and all references
- **Build Files**: Updated Dockerfile, Makefile, GitHub Actions, and scripts
- **TUI Branding**: All interface titles, messages, and design system updated

### Technical Changes
- Updated all import paths across the codebase
- Maintained full functionality during rebranding
- Verified builds and tests pass with new branding

## [Iteration 2] - Code Quality Improvements

### Added
- **Enhanced Application Scanner (V2)**: Cleaner state management with `ScannerState` enum, improved progress tracking, better error handling, more efficient rendering
- **Concurrent Scanning**: Parallel application checking with worker pools, context-based cancellation, 5x faster scanning with 5 workers, timeout protection (10 seconds max)
- **Centralized Error Handling**: Unified error management with severity levels (Info, Warning, Error, Critical), panic recovery with stack traces, error history tracking, automatic error notification system
- **State Machine for UI Transitions**: Validated state transitions, state history with back navigation, prevents invalid state changes, clear transition rules
- **Configuration Validation**: Comprehensive validation rules, app definition validation, registry validation, path and format checking, warning system for non-critical issues

### Performance Improvements
- Concurrent scanning: <500ms for 15 apps (previously ~2 seconds sequential)
- Reduced cyclomatic complexity: Average 4 (previously 8)
- Improved error handling coverage: 95% of functions (previously 30%)
- Reduced code duplication: <5% (previously 15%)

### Architecture Improvements
- Separation of concerns with dedicated components for scanning, error handling, state management, and validation
- Dependency injection for easier testing and clear interfaces
- Enhanced test coverage with new tests for scanner, error handler, state machine, and concurrent operations

### Developer Experience
- Cleaner, more maintainable code with consistent patterns
- Better testing capabilities with mock implementations
- Clear architectural patterns and comprehensive documentation
- Improved stability and reliability

## [Initial Release] - Major Features Implementation

### Added
- **Robust Terminal UI Rendering**: Fixed terminal rendering issues and garbled output, proper cleanup on exit with terminal restoration, window size handling and responsive layout, eliminated UI flicker and misalignments
- **Application Scanning System**: `AppScanner` component with progress indicators, package manager style inspired by Bubble Tea examples, real-time progress with spinner and progress bar, categorized results by status (Ready/Not Configured/Error)
- **Apps Registry System**: YAML-based registry with 15+ pre-configured apps, support for custom apps via `~/.config/zeroui/apps.yaml`, override capability for existing apps, organization by categories (terminal, editor, shell, tools)
- **Configuration Safety**: Temporary file management with integrity checks, SHA-256 checksums for file verification, atomic operations with backup rotation, cross-platform support (Windows, macOS, Linux)
- **Performance Optimizations**: Render cache for improved performance, smart refresh with debounce control, parallel execution for scanning and operations, optimized memory management with buffer pools

### UI/UX Enhancements
- No more flickering or misalignment in TUI
- Guaranteed navigation with no dead states
- Snappy and responsive interface with progress indicators
- Clear status symbols (○ Ready, ○ Not Configured)
- ASCII icons instead of emojis for better compatibility

### Developer Experience
- Custom app registration via YAML configuration
- Automatic config file discovery
- Multiple config path support
- Category-based organization

### Reliability
- Safe file operations with backups and integrity verification
- Proper error handling and graceful fallbacks
- Comprehensive testing including unit, integration, and performance tests

### Configuration
- ZeroUI config priority: embedded registry → custom apps → full overrides
- Example custom app support with flexible YAML structure

---

**Note**: This changelog consolidates historical development iterations and rebranding efforts. Future releases will follow semantic versioning with detailed change descriptions.
