# ZeroUI Scripts

A comprehensive collection of maintenance, testing, and utility scripts for the ZeroUI project.

## Quick Start

Use the unified script dispatcher for all operations:

```bash
# Show all available commands
./scripts/zeroui-scripts.sh help

# Fast configuration update
./scripts/zeroui-scripts.sh config:fast

# Update dependencies
./scripts/zeroui-scripts.sh deps:update

# Run tests
./scripts/zeroui-scripts.sh test:run

# Generate configurations
./scripts/zeroui-scripts.sh gen:ghostty
./scripts/zeroui-scripts.sh gen:zed

# Manage Git hooks
./scripts/zeroui-scripts.sh hooks:install
```

## Script Organization

### Core Scripts

- **`zeroui-scripts.sh`** - Unified command dispatcher for all scripts
- **`install-git-hooks.sh`** - Git hooks management

### Maintenance (`maintenance/`)

- **`fast-update-configs.sh`** - High-performance configuration extraction
- **`update-dependencies.sh`** - Safe dependency updates with rollback

### Testing (`testing/`)

- **`run_tui_tests.sh`** - Comprehensive TUI testing with multiple terminal sizes

### Generators (`generator/`)

- **`generate_ghostty_reference.sh`** - Ghostty configuration reference generator
- **`generate_zed_reference.py`** - Zed configuration reference generator

### Libraries (`lib/`)

- **`dry_run.sh`** - Shared dry-run functionality

## Environment Variables

All scripts support these environment variables:

- **`DRY_RUN=true`** - Show what would be done without making changes
- **`VERBOSE=true`** - Enable verbose output
- **`SKIP_BUILD=true`** - Skip binary building where applicable

## Examples

### Development Workflow

```bash
# Update configurations after adding new app support
./scripts/zeroui-scripts.sh config:fast --rebuild

# Update dependencies safely
./scripts/zeroui-scripts.sh deps:update

# Run comprehensive tests
./scripts/zeroui-scripts.sh test:run --verbose

# Generate new configuration references
./scripts/zeroui-scripts.sh gen:ghostty
```

### CI/CD Integration

```bash
# Dry run for CI validation
DRY_RUN=true ./scripts/zeroui-scripts.sh config:fast

# Automated dependency updates
./scripts/zeroui-scripts.sh deps:update
```

## Script Dependencies

Most scripts require:
- **Go 1.24+** - For building and running the main application
- **Python 3** - For Python-based generators
- **bash** - For shell scripts
- **Standard Unix tools** - `grep`, `awk`, `sed`, etc.

## Contributing

When adding new scripts:

1. **Add to dispatcher**: Update `zeroui-scripts.sh` with new commands
2. **Follow patterns**: Use consistent error handling and logging
3. **Document**: Add usage examples and environment variables
4. **Test**: Include both success and failure scenarios

## Troubleshooting

### Common Issues

**Script not found error:**
```bash
# Make sure you're in the project root
cd /path/to/zeroui
./scripts/zeroui-scripts.sh help
```

**Permission denied:**
```bash
# Fix script permissions
chmod +x scripts/*.sh scripts/**/*.sh
```

**Missing dependencies:**
```bash
# Install required tools
go version  # Should be 1.24+
python3 --version  # Should be 3.x+
```

## Advanced Usage

### Custom Terminal Sizes for Testing

```bash
TERMINAL_SIZES="80x24,120x40,160x50" ./scripts/zeroui-scripts.sh test:run
```

### Parallel Configuration Updates

```bash
WORKERS=8 ./scripts/zeroui-scripts.sh config:fast
```

### Selective Configuration Updates

```bash
# Update only specific applications
./scripts/zeroui-scripts.sh config:fast --apps ghostty,zed
```
