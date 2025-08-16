#!/usr/bin/env bash
#
# Dry-run helper library for ZeroUI scripts
#
# Usage:
#   source scripts/lib/dry_run.sh
#   dry_run_init               # initialize dry-run environment (if DRY_RUN=1)
#   dry_run_exec <cmd> [args] # run a command, simulated under DRY_RUN or executed normally
#   dry_run_safe_build <build_cmd...>
#
# This library is safe to source. When DRY_RUN=1 it will:
#   - Prefer an in-repo wrapper at scripts/_dry_run_wrapper.sh as the simulated BINARY
#   - Prepend scripts/drybin to PATH (if present) so fake 'go' / 'nproc' can be used
#   - Export SKIP_BUILD=1 to signal other scripts to avoid actual builds
#
# The helper functions provide consistent dry-run behavior across scripts and make it
# easier to implement safe, idempotent checks in CI or local testing without performing
# destructive actions.

# Do not make this script exit the caller; keep it safe to source.
# Avoid `set -e` here because we don't want to force the caller to adopt strict mode.
# But use `set -u` to avoid accidental use of unset variables in the library itself.
set -u

# Internal variables (set by dry_run_init)
DRY_RUN_PROJECT_ROOT=""
DRY_RUN_WRAPPER=""
DRY_RUN_DRYBIN_DIR=""

# Initialize dry-run environment. This should be called by scripts that source this file.
# If DRY_RUN is set to a non-empty, non-zero value, this will:
#  - set BINARY to scripts/_dry_run_wrapper.sh if present (and executable)
#  - prepend scripts/drybin to PATH if present
#  - export SKIP_BUILD=1
dry_run_init() {
  # Resolve the directory of this library file, even when sourced from another directory.
  local lib_dir
  # If BASH_SOURCE is not defined (rare outside bash), fall back to $0.
  if [ -n "${BASH_SOURCE[0]:-}" ]; then
    lib_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  else
    lib_dir="$(cd "$(dirname "$0")" && pwd)"
  fi

  # Project root is two levels up from scripts/lib -> (project)/scripts/lib
  DRY_RUN_PROJECT_ROOT="$(cd "$lib_dir/../.." >/dev/null 2>&1 && pwd || true)"
  DRY_RUN_WRAPPER="$DRY_RUN_PROJECT_ROOT/scripts/_dry_run_wrapper.sh"
  DRY_RUN_DRYBIN_DIR="$DRY_RUN_PROJECT_ROOT/scripts/drybin"

  # If DRY_RUN is enabled, configure environment
  if [ "${DRY_RUN:-0}" != "0" ]; then
    echo "(DRY-RUN) dry_run_init: DRY_RUN enabled — configuring dry-run environment"

    # Prefer explicit BINARY override, otherwise use wrapper if available
    if [ -z "${BINARY:-}" ] && [ -x "$DRY_RUN_WRAPPER" ]; then
      export BINARY="$DRY_RUN_WRAPPER"
      echo "(DRY-RUN) using BINARY=$BINARY"
    else
      # If BINARY already set, show it
      if [ -n "${BINARY:-}" ]; then
        echo "(DRY-RUN) BINARY already set to: ${BINARY}"
      fi
    fi

    # Prepend drybin to PATH if available
    if [ -d "$DRY_RUN_DRYBIN_DIR" ]; then
      export PATH="$DRY_RUN_DRYBIN_DIR:$PATH"
      echo "(DRY-RUN) prefixed PATH with $DRY_RUN_DRYBIN_DIR"
    fi

    # Mark that builds should be skipped by scripts that respect SKIP_BUILD
    export SKIP_BUILD=1
    echo "(DRY-RUN) SKIP_BUILD=1 exported"
  fi
}

# dry_run_exec: run a command under dry-run semantics.
# If DRY_RUN=1:
#   - If a BINARY wrapper exists and the command starts with the real binary path or 'zeroui',
#     the wrapper will be invoked with the subcommand arguments (best-effort simulation).
#   - Otherwise, prints a simulated command and returns 0.
# If DRY_RUN is not enabled, executes the provided command normally.
#
# Examples:
#   dry_run_exec "$BINARY" batch-extract --output-dir configs --workers 4 --update
#   dry_run_exec go test ./...
dry_run_exec() {
  if [ $# -eq 0 ]; then
    echo "dry_run_exec: no command provided" >&2
    return 2
  fi

  # Build command string for printing
  local cmd_quoted
  cmd_quoted="$*"

  if [ "${DRY_RUN:-0}" != "0" ]; then
    # If a wrapper (BINARY) exists and is executable, prefer calling it for simulation.
    if [ -n "${BINARY:-}" ] && [ -x "${BINARY}" ]; then
      # If the command invokes the built binary (e.g. /path/to/build/zeroui or 'zeroui'), call wrapper
      local first="$1"
      shift || true
      echo "(DRY-RUN) dry_run_exec: invoking wrapper ${BINARY} with args: $*"
      "${BINARY}" "$@" || true
      return 0
    fi

    # Otherwise, just print what would have been executed
    echo "(DRY-RUN) dry_run_exec: simulated -> ${cmd_quoted}"
    return 0
  fi

  # Normal execution path
  "$@"
}

# dry_run_safe_build: helper to run builds safely.
# If SKIP_BUILD or DRY_RUN is set, it will skip the build and print a message.
# Otherwise it runs the provided build command (e.g. go build -o build/zeroui .).
# A timeout will be used if available.
#
# Usage:
#   dry_run_safe_build go build -o build/zeroui .
dry_run_safe_build() {
  if [ "${DRY_RUN:-0}" != "0" ] || [ -n "${SKIP_BUILD:-}" ]; then
    echo "(DRY-RUN) dry_run_safe_build: skipping build (DRY_RUN or SKIP_BUILD enabled)"
    return 0
  fi

  if [ $# -eq 0 ]; then
    echo "dry_run_safe_build: no build command provided" >&2
    return 2
  fi

  # Use timeout if available to avoid long-running builds
  if command -v timeout >/dev/null 2>&1; then
    timeout 60s "$@"
    return $?
  else
    # Fallback: run normally
    "$@"
    return $?
  fi
}

# dry_run_confirm: prompt for confirmation unless DRY_RUN=1
# Returns 0 if confirmed or DRY_RUN is set, non-zero otherwise.
dry_run_confirm() {
  if [ "${DRY_RUN:-0}" != "0" ]; then
    echo "(DRY-RUN) dry_run_confirm: DRY_RUN enabled — auto-confirming"
    return 0
  fi

  # If stdin not a tty, don't block; return non-zero
  if [ ! -t 0 ]; then
    echo "dry_run_confirm: non-interactive shell; refusing to confirm" >&2
    return 1
  fi

  # Prompt the user
  local prompt="${1:-Proceed? [y/N]}"
  read -r -p "$prompt " response
  case "$response" in
    [yY]|[yY][eE][sS]) return 0 ;;
    *) return 1 ;;
  esac
}

# Export functions for convenience if sourced in bash
export -f dry_run_init 2>/dev/null || true
export -f dry_run_exec 2>/dev/null || true
export -f dry_run_safe_build 2>/dev/null || true
export -f dry_run_confirm 2>/dev/null || true

# If this file is sourced directly by a script, auto-initialize so scripts only need:
#   source scripts/lib/dry_run.sh && dry_run_init
# We do NOT auto-call dry_run_init here to avoid surprising behavior; caller should explicitly call it.
#
# End of dry_run.sh
