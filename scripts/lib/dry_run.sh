#!/bin/bash
# Dry run utility for ZeroUI scripts
# Simple implementation without external dependencies

DRY_RUN="${DRY_RUN:-false}"
VERBOSE="${VERBOSE:-false}"

dry_run_init() {
    if [[ "$DRY_RUN" == "true" ]]; then
        echo "🔍 DRY RUN MODE - No actual changes will be made"
    fi
}

dry_run_log() {
    local message="$1"
    if [[ "$DRY_RUN" == "true" ]] || [[ "$VERBOSE" == "true" ]]; then
        echo "🔍 $message"
    fi
}

dry_run_exec() {
    local cmd="$1"
    if [[ "$DRY_RUN" == "true" ]]; then
        dry_run_log "Would execute: $cmd"
        return 0
    else
        dry_run_log "Executing: $cmd"
        eval "$cmd"
        return $?
    fi
}
