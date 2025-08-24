#!/bin/bash
# ZeroUI Scripts Dispatcher
# Unified interface for all ZeroUI maintenance, testing, and generation scripts

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load common utilities
source "$SCRIPT_DIR/lib/common.sh"
source "$SCRIPT_DIR/lib/dry_run.sh" && dry_run_init

# Script version
VERSION="2.1.0"

# Available commands
declare -A COMMANDS=(
    ["config:fast"]="Fast configuration extraction with parallel processing"
    ["config:update"]="Update all application configurations"
    ["deps:update"]="Update Go dependencies safely"
    ["test:run"]="Run comprehensive TUI tests"
    ["gen:ghostty"]="Generate Ghostty reference configuration"
    ["gen:zed"]="Generate Zed reference configuration" 
    ["hooks:install"]="Install Git hooks"
    ["hooks:uninstall"]="Uninstall Git hooks"
    ["help"]="Show this help message"
)

show_header() {
    log_header "ZeroUI Scripts v$VERSION"
    echo ""
}

show_help() {
    show_header
    echo -e "${BLUE}Available Commands:${NC}"
    echo ""
    
    for cmd in "${!COMMANDS[@]}"; do
        printf "  ${GREEN}%-15s${NC} %s\n" "$cmd" "${COMMANDS[$cmd]}"
    done
    
    echo ""
    echo -e "${YELLOW}Usage:${NC}"
    echo "  $0 <command> [args...]"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  $0 config:fast --rebuild"
    echo "  $0 deps:update"
    echo "  $0 test:run --verbose"
    echo "  $0 gen:ghostty"
    echo "  $0 hooks:install"
}

run_command() {
    local cmd="$1"
    shift

    # Validate project structure
    if ! validate_project_structure "$PROJECT_ROOT"; then
        log_error "Invalid project structure"
        exit 1
    fi

    # Check Go version for Go-related commands
    case "$cmd" in
        config:*|deps:*|test:*)
            if ! check_go_version; then
                log_warning "Continuing with available Go version"
            fi
            ;;
    esac

    local start_time
    start_time=$(date +%s%N)

    case "$cmd" in
        "config:fast")
            log_info "🚀 Running fast configuration extraction..."
            cd "$PROJECT_ROOT" && bash "$SCRIPT_DIR/maintenance/fast-update-configs.sh" "$@"
            ;;
        "config:update")
            log_info "📦 Updating all configurations..."
            log_info "Note: config:update was consolidated into config:fast"
            cd "$PROJECT_ROOT" && bash "$SCRIPT_DIR/maintenance/fast-update-configs.sh" "$@"
            ;;
        "deps:update")
            log_info "📦 Updating dependencies..."
            if ! ensure_clean_git; then
                log_error "Cannot update dependencies with uncommitted changes"
                exit 1
            fi
            cd "$PROJECT_ROOT" && bash "$SCRIPT_DIR/maintenance/update-dependencies.sh" "$@"
            ;;
        "test:run")
            log_info "🧪 Running TUI tests..."
            cd "$PROJECT_ROOT" && bash "$SCRIPT_DIR/testing/run_tui_tests.sh" "$@"
            ;;
        "gen:ghostty")
            log_info "👻 Generating Ghostty reference..."
            if ! check_command "ghostty"; then
                log_error "Ghostty is not installed"
                exit 1
            fi
            cd "$PROJECT_ROOT" && bash "$SCRIPT_DIR/generator/generate_ghostty_reference.sh" "$@"
            ;;
        "gen:zed")
            log_info "⚡ Generating Zed reference..."
            if ! check_command "python3"; then
                log_error "Python 3 is not installed"
                exit 1
            fi
            cd "$PROJECT_ROOT" && python3 "$SCRIPT_DIR/generator/generate_zed_reference.py" "$@"
            ;;
        "hooks:install")
            log_info "🔗 Installing Git hooks..."
            cd "$PROJECT_ROOT" && bash "$SCRIPT_DIR/install-git-hooks.sh" --install "$@"
            ;;
        "hooks:uninstall")
            log_info "🔗 Uninstalling Git hooks..."
            cd "$PROJECT_ROOT" && bash "$SCRIPT_DIR/install-git-hooks.sh" --uninstall "$@"
            ;;
        *)
            log_error "Unknown command: $cmd"
            echo ""
            show_help
            exit 1
            ;;
    esac

    local duration=$(end_timer "$start_time")
    log_success "Command completed in $(format_duration "$duration")"
}

# Main execution
main() {
    if [[ $# -eq 0 ]] || [[ "$1" == "help" ]] || [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
        show_help
        exit 0
    fi
    
    local cmd="$1"
    shift
    run_command "$cmd" "$@"
}

main "$@"
