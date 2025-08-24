#!/bin/bash
# Common utilities for ZeroUI scripts

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

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

log_header() {
    local msg="$1"
    local width=60
    local padding=$(( (width - ${#msg}) / 2 ))
    local line=$(printf '%*s' "$width" '' | tr ' ' '═')
    
    echo -e "${CYAN}╔${line}╗${NC}"
    printf "${CYAN}║%*s%s%*s║${NC}\n" "$padding" "" "$msg" "$padding" ""
    echo -e "${CYAN}╚${line}╝${NC}"
}

# Utility functions
check_command() {
    local cmd="$1"
    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "$cmd is not installed or not in PATH"
        return 1
    fi
    return 0
}

check_go_version() {
    if ! check_command "go"; then
        return 1
    fi
    
    local go_version=$(go version | cut -d' ' -f3 | sed 's/go//')
    local major=$(echo "$go_version" | cut -d'.' -f1)
    local minor=$(echo "$go_version" | cut -d'.' -f2)
    
    if [[ $major -lt 1 ]] || [[ $major -eq 1 && $minor -lt 24 ]]; then
        log_warning "Go version $go_version detected. Go 1.24+ recommended."
    else
        log_info "Go version $go_version detected"
    fi
}

get_project_root() {
    local script_dir="$1"
    local project_root=""
    
    # Try git first
    if command -v git >/dev/null 2>&1; then
        project_root=$(git -C "$script_dir" rev-parse --show-toplevel 2>/dev/null)
    fi
    
    # Fallback to directory traversal
    if [[ -z "$project_root" ]]; then
        local current="$script_dir"
        while [[ "$current" != "/" ]]; do
            if [[ -f "$current/go.mod" ]]; then
                project_root="$current"
                break
            fi
            current=$(dirname "$current")
        done
    fi
    
    echo "$project_root"
}

validate_project_structure() {
    local project_root="$1"
    
    if [[ ! -f "$project_root/go.mod" ]]; then
        log_error "Not in a Go project (go.mod not found)"
        return 1
    fi
    
    if [[ ! -d "$project_root/internal" ]]; then
        log_error "Project structure incomplete (internal/ directory missing)"
        return 1
    fi
    
    return 0
}

format_duration() {
    local milliseconds="$1"
    local seconds=$((milliseconds / 1000))
    local ms=$((milliseconds % 1000))
    
    if [[ $seconds -gt 0 ]]; then
        echo "${seconds}s ${ms}ms"
    else
        echo "${ms}ms"
    fi
}

# Configuration helpers
get_config_value() {
    local config_file="$1"
    local key="$2"
    local default="$3"
    
    if [[ -f "$config_file" ]]; then
        # Simple key=value parsing (not for complex configs)
        grep "^$key=" "$config_file" 2>/dev/null | cut -d'=' -f2- || echo "$default"
    else
        echo "$default"
    fi
}

set_config_value() {
    local config_file="$1"
    local key="$2"
    local value="$3"
    
    # Create directory if it doesn't exist
    mkdir -p "$(dirname "$config_file")"
    
    if [[ -f "$config_file" ]]; then
        # Update existing key
        sed -i.bak "s|^$key=.*|$key=$value|" "$config_file"
        # Add if key doesn't exist
        if ! grep -q "^$key=" "$config_file"; then
            echo "$key=$value" >> "$config_file"
        fi
    else
        echo "$key=$value" > "$config_file"
    fi
}

# File operations with safety
safe_backup() {
    local file="$1"
    if [[ -f "$file" ]]; then
        local backup="${file}.bak.$(date +%Y%m%d_%H%M%S)"
        cp "$file" "$backup"
        log_info "Created backup: $backup"
    fi
}

cleanup_backups() {
    local pattern="$1"
    local keep_days="${2:-7}"
    
    find . -name "$pattern" -mtime +$keep_days -delete 2>/dev/null || true
}

# Progress and status
show_progress() {
    local current="$1"
    local total="$2"
    local item="$3"
    
    local percentage=$((current * 100 / total))
    local progress=$((percentage / 2)) # 50 characters wide
    local bar=$(printf '%*s' "$progress" '' | tr ' ' '█')
    local spaces=$((50 - progress))
    local space_fill=$(printf '%*s' "$spaces" '')
    
    printf "\r${BLUE}Progress:${NC} [%s%s] %d%% (%d/%d) %s" \
           "$bar" "$space_fill" "$percentage" "$current" "$total" "$item"
}

clear_progress() {
    echo ""
}

# Error handling
trap_errors() {
    local func="${1:-}"
    if [[ -n "$func" ]]; then
        trap "$func" ERR
    fi
}

cleanup_on_error() {
    local exit_code="$?"
    if [[ $exit_code -ne 0 ]]; then
        log_error "Script failed with exit code $exit_code"
        # Add cleanup logic here
    fi
}

# Validation helpers
validate_directory() {
    local dir="$1"
    local description="$2"
    
    if [[ ! -d "$dir" ]]; then
        log_error "$description directory not found: $dir"
        return 1
    fi
    return 0
}

validate_file() {
    local file="$1"
    local description="$2"
    
    if [[ ! -f "$file" ]]; then
        log_error "$description file not found: $file"
        return 1
    fi
    return 0
}

# Network helpers
check_network() {
    if ! curl -s --connect-timeout 5 https://httpbin.org/status/200 >/dev/null 2>&1; then
        log_warning "Network connectivity issues detected"
        return 1
    fi
    return 0
}

# Git helpers
ensure_clean_git() {
    if ! git diff-index --quiet HEAD --; then
        log_error "Git working directory is not clean"
        log_info "Please commit or stash your changes first"
        return 1
    fi
    return 0
}

get_git_branch() {
    git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown"
}

get_git_commit() {
    git rev-parse --short HEAD 2>/dev/null || echo "unknown"
}

# Performance monitoring
start_timer() {
    START_TIME=$(date +%s%N)
}

end_timer() {
    local start_time="$1"
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # milliseconds
    echo "$duration"
}

# Export all functions
export -f log_info log_success log_warning log_error log_header
export -f check_command check_go_version get_project_root validate_project_structure
export -f format_duration get_config_value set_config_value safe_backup cleanup_backups
export -f show_progress clear_progress trap_errors cleanup_on_error
export -f validate_directory validate_file check_network ensure_clean_git
export -f get_git_branch get_git_commit start_timer end_timer
