# DRY_RUN HANDLER
if [ "${DRY_RUN:-0}" != "0" ]; then
  echo "(DRY-RUN) $0: DRY_RUN enabled, skipping destructive actions."
fi

#!/bin/bash

# TUI Testing Automation Script
# Comprehensive TUI testing with automated rendering correctness verification

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TUI_DIR="$PROJECT_ROOT/internal/tui"
REPORT_DIR="$TUI_DIR/testdata/reports/$(date +%Y%m%d-%H%M%S)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
UPDATE_BASELINES="${UPDATE_BASELINES:-false}"
GENERATE_IMAGES="${GENERATE_IMAGES:-true}"
RUN_BENCHMARKS="${RUN_BENCHMARKS:-true}"
VERBOSE="${VERBOSE:-false}"
TERMINAL_SIZES="${TERMINAL_SIZES:-80x24,120x40,160x50}"
PARALLEL="${PARALLEL:-true}"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Usage information
show_usage() {
    cat << EOF
TUI Testing Automation Script

Usage: $0 [OPTIONS] [TEST_TYPE]

OPTIONS:
    -h, --help              Show this help message
    -u, --update-baselines  Update visual baselines
    -i, --generate-images   Generate image representations
    -b, --benchmarks        Run performance benchmarks
    -v, --verbose           Enable verbose output
    -p, --parallel          Run tests in parallel
    -s, --sizes SIZES       Terminal sizes to test (default: 80x24,120x40,160x50)

TEST_TYPE:
    all                     Run all TUI tests (default)
    snapshot                Run snapshot tests only
    visual                  Run visual regression tests only
    automation              Run automation framework tests only
    performance             Run performance tests only
    correctness             Run rendering correctness tests only
    ci                      Run CI-specific tests only

EXAMPLES:
    $0                              # Run all tests with default settings
    $0 -u visual                    # Update baselines and run visual tests
    $0 -b performance               # Run performance tests with benchmarks
    $0 -v -s "80x24,120x40" all     # Verbose mode with specific terminal sizes

ENVIRONMENT VARIABLES:
    UPDATE_BASELINES=true           Update visual baselines
    GENERATE_IMAGES=true            Generate image representations
    RUN_BENCHMARKS=true             Run performance benchmarks
    VERBOSE=true                    Enable verbose output
    TERMINAL_SIZES="80x24,120x40"   Specify terminal sizes
    PARALLEL=true                   Run tests in parallel

EOF
}

# Parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -u|--update-baselines)
                UPDATE_BASELINES=true
                shift
                ;;
            -i|--generate-images)
                GENERATE_IMAGES=true
                shift
                ;;
            -b|--benchmarks)
                RUN_BENCHMARKS=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -p|--parallel)
                PARALLEL=true
                shift
                ;;
            -s|--sizes)
                TERMINAL_SIZES="$2"
                shift 2
                ;;
            all|snapshot|visual|automation|performance|correctness|ci)
                TEST_TYPE="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Setup test environment
setup_environment() {
    log_info "Setting up test environment..."
    
    # Ensure we're in the right directory
    cd "$PROJECT_ROOT"
    
    # Create necessary directories
    mkdir -p "$TUI_DIR/testdata"/{snapshots,visual,baseline,diffs,automated,reports,baseline_images,diff_images}
    
    # Set environment variables
    export TERM=xterm-256color
    export CI=false
    export UPDATE_TUI_BASELINES="$UPDATE_BASELINES"
    export GENERATE_TUI_IMAGES="$GENERATE_IMAGES"
    
    # Create report directory
    mkdir -p "$REPORT_DIR"
    
    log_success "Environment setup complete"
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check Go version
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    GO_VERSION=$(go version | cut -d' ' -f3 | sed 's/go//')
    log_info "Go version: $GO_VERSION"
    
    # Check if we're in a Go module
    if [[ ! -f "go.mod" ]]; then
        log_error "Not in a Go module directory"
        exit 1
    fi
    
    # Download dependencies
    log_info "Downloading Go dependencies..."
    go mod download
    
    log_success "Dependencies check complete"
}

# Run specific test type
run_test_type() {
    local test_type="$1"
    local verbose_flag=""
    local timeout="10m"
    
    if [[ "$VERBOSE" == "true" ]]; then
        verbose_flag="-v"
    fi
    
    cd "$TUI_DIR"
    
    case "$test_type" in
        snapshot)
            log_info "Running TUI snapshot tests..."
            go test $verbose_flag -run TestSnapshot -timeout=5m
            ;;
        visual)
            log_info "Running TUI visual regression tests..."
            go test $verbose_flag -run TestVisualRegression -timeout=$timeout
            ;;
        automation)
            log_info "Running TUI automation framework tests..."
            go test $verbose_flag -run TestAutomatedTUIRendering -timeout=15m
            ;;
        performance)
            log_info "Running TUI performance tests..."
            go test $verbose_flag -run TestTUIRenderingCorrectness -timeout=5m
            if [[ "$RUN_BENCHMARKS" == "true" ]]; then
                log_info "Running performance benchmarks..."
                go test $verbose_flag -bench=BenchmarkTUI -benchmem -run=^$ -timeout=10m | tee "$REPORT_DIR/benchmarks.txt"
            fi
            ;;
        correctness)
            log_info "Running TUI rendering correctness tests..."
            go test $verbose_flag -run TestTUIRenderingCorrectness -timeout=5m
            ;;
        ci)
            log_info "Running CI-specific tests..."
            export CI=true
            go test $verbose_flag -run TestContinuousIntegration -timeout=5m
            ;;
        all)
            log_info "Running all TUI tests..."
            run_test_type snapshot
            run_test_type visual
            run_test_type automation
            run_test_type correctness
            if [[ "$RUN_BENCHMARKS" == "true" ]]; then
                run_test_type performance
            fi
            ;;
        *)
            log_error "Unknown test type: $test_type"
            exit 1
            ;;
    esac
}

# Run tests for multiple terminal sizes
run_multi_size_tests() {
    local test_type="$1"
    
    IFS=',' read -ra SIZES <<< "$TERMINAL_SIZES"
    
    for size in "${SIZES[@]}"; do
        IFS='x' read -ra DIMENSIONS <<< "$size"
        local width="${DIMENSIONS[0]}"
        local height="${DIMENSIONS[1]}"
        
        log_info "Testing terminal size: ${width}x${height}"
        
        export COLUMNS="$width"
        export LINES="$height"
        
        # Create size-specific report directory
        local size_report_dir="$REPORT_DIR/size_${width}x${height}"
        mkdir -p "$size_report_dir"
        
        # Run tests with size-specific output
        if run_test_type "$test_type" 2>&1 | tee "$size_report_dir/test_output.log"; then
            log_success "Tests passed for ${width}x${height}"
        else
            log_error "Tests failed for ${width}x${height}"
            return 1
        fi
    done
}

# Generate comprehensive report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="$REPORT_DIR/tui_test_report.md"
    
    cat > "$report_file" << EOF
# TUI Testing Report

Generated: $(date)
Script: $0
Arguments: $*

## Environment
- Go Version: $(go version)
- Terminal: $TERM
- Project Root: $PROJECT_ROOT
- Update Baselines: $UPDATE_BASELINES
- Generate Images: $GENERATE_IMAGES
- Run Benchmarks: $RUN_BENCHMARKS

## Terminal Sizes Tested
EOF
    
    IFS=',' read -ra SIZES <<< "$TERMINAL_SIZES"
    for size in "${SIZES[@]}"; do
        echo "- $size" >> "$report_file"
    done
    
    cat >> "$report_file" << EOF

## Test Results

EOF
    
    # Add test results for each size
    for size in "${SIZES[@]}"; do
        local size_report_dir="$REPORT_DIR/size_$size"
        if [[ -f "$size_report_dir/test_output.log" ]]; then
            echo "### Terminal Size: $size" >> "$report_file"
            echo "" >> "$report_file"
            
            # Extract test results
            if grep -q "PASS" "$size_report_dir/test_output.log"; then
                echo "✅ Tests passed" >> "$report_file"
            elif grep -q "FAIL" "$size_report_dir/test_output.log"; then
                echo "❌ Tests failed" >> "$report_file"
            else
                echo "⚠️ Tests status unknown" >> "$report_file"
            fi
            echo "" >> "$report_file"
        fi
    done
    
    # Add file listings
    cat >> "$report_file" << EOF
## Generated Files

### Snapshots
EOF
    
    if [[ -d "$TUI_DIR/testdata/snapshots" ]]; then
        ls -la "$TUI_DIR/testdata/snapshots" | tail -n +2 | while read -r line; do
            echo "- $line" >> "$report_file"
        done
    fi
    
    cat >> "$report_file" << EOF

### Visual Regression
EOF
    
    if [[ -d "$TUI_DIR/testdata/visual" ]]; then
        ls -la "$TUI_DIR/testdata/visual" | tail -n +2 | while read -r line; do
            echo "- $line" >> "$report_file"
        done
    fi
    
    # Add performance results if available
    if [[ -f "$REPORT_DIR/benchmarks.txt" ]]; then
        cat >> "$report_file" << EOF

## Performance Benchmarks

\`\`\`
$(cat "$REPORT_DIR/benchmarks.txt")
\`\`\`
EOF
    fi
    
    log_success "Report generated: $report_file"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    
    # Reset environment variables
    unset UPDATE_TUI_BASELINES
    unset GENERATE_TUI_IMAGES
    unset COLUMNS
    unset LINES
    
    # Return to original directory
    cd "$PROJECT_ROOT"
}

# Main execution function
main() {
    local test_type="${1:-all}"
    
    # Setup
    setup_environment
    check_dependencies
    
    # Register cleanup function
    trap cleanup EXIT
    
    # Run tests
    log_info "Starting TUI testing automation..."
    log_info "Test type: $test_type"
    log_info "Terminal sizes: $TERMINAL_SIZES"
    
    if run_multi_size_tests "$test_type"; then
        log_success "All tests completed successfully"
    else
        log_error "Some tests failed"
        exit 1
    fi
    
    # Generate report
    generate_report
    
    # Summary
    log_info "Testing complete!"
    log_info "Report directory: $REPORT_DIR"
    
    # Show quick stats
    local total_snapshots=$(find "$TUI_DIR/testdata/snapshots" -name "*.txt" 2>/dev/null | wc -l)
    local total_visual=$(find "$TUI_DIR/testdata/visual" -name "*.txt" 2>/dev/null | wc -l)
    local total_diffs=$(find "$TUI_DIR/testdata/diffs" -name "*.txt" 2>/dev/null | wc -l)
    
    log_info "Generated $total_snapshots snapshots, $total_visual visual files, $total_diffs diffs"
    
    if [[ $total_diffs -gt 0 ]]; then
        log_warning "Found $total_diffs visual differences - review recommended"
    fi
}

# Parse arguments and run
TEST_TYPE="all"
parse_arguments "$@"
main "$TEST_TYPE"
