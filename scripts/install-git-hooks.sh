#!/usr/bin/env bash
# Best-effort: ensure this script is marked executable in the git index so clones preserve the bit.
# This uses `git update-index --chmod=+x` and will not fail the script if git isn't available.
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [ -n "$REPO_ROOT" ] && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  # Path to this file relative to repository root
  # Calculate path relative to repository root
  SCRIPT_PATH="$(realpath "$0")"
  GIT_REL_PATH="${SCRIPT_PATH#$REPO_ROOT/}"
  # Try to set the executable bit in the index (harmless if it fails)
  git -C "$REPO_ROOT" update-index --add --chmod=+x "$GIT_REL_PATH" 2>/dev/null || true
fi
#
# Install / manage repository Git hooks path to use the in-repo `.githooks/` directory.
#
# Usage:
#   scripts/install-git-hooks.sh --install     # set git config core.hooksPath to .githooks
#   scripts/install-git-hooks.sh --uninstall   # restore previous core.hooksPath (if any) or unset it
#   scripts/install-git-hooks.sh --show        # show current core.hooksPath and saved backup
#   scripts/install-git-hooks.sh --help        # show this help
#
# Behavior:
#  - Detects repo top-level via `git rev-parse --show-toplevel`.
#  - Validates that `.githooks/` exists (relative to repo root). If missing, it will exit with a helpful error.
#  - When installing, saves the previous `core.hooksPath` (if any) in a local git config key:
#        hooks.install.previous
#    so it can be restored later on uninstall.
#  - Makes any scripts inside `.githooks/` executable by default.
#  - Safe by default; use `--force` to overwrite the saved backup key if you need to.
#
# Notes:
#  - This script updates the local repository git config (not global).
#  - After installation, `git` will look for hooks in the `.githooks/` directory at the repo root.
#  - To enable hooks immediately for your local clone, run:
#        git config core.hooksPath .githooks
#    (this script does that for you).
#

set -euo pipefail

PROG_NAME="$(basename "$0")"
HOOKS_DIR_REL=".githooks"
BACKUP_CONFIG_KEY="hooks.install.previous"

print_help() {
  cat <<EOF
$PROG_NAME - Install / manage in-repo Git hooks directory

Usage:
  $0 --install [--force]   Install .githooks as core.hooksPath (saves previous value)
  $0 --uninstall           Restore previous core.hooksPath (if saved) or unset it
  $0 --show                Show current core.hooksPath and any saved backup
  $0 --help                Show this help text

Options:
  --force                  Overwrite any existing saved backup when --install is used
  --dry-run                Don't make changes, only show what would be done
  --verbose                Print additional diagnostic information

Examples:
  # Install hooks for this repository
  $0 --install

  # Uninstall / restore previous hooksPath
  $0 --uninstall

EOF
}

die() {
  echo "$PROG_NAME: $*" >&2
  exit 1
}

info() {
  if [ "${VERBOSE:-0}" -ne 0 ]; then
    echo "$PROG_NAME: $*"
  fi
}

DRY_RUN=0
VERBOSE=0
FORCE=0

# parse args
if [ $# -eq 0 ]; then
  print_help
  exit 0
fi

ACTION=""
while [ $# -gt 0 ]; do
  case "$1" in
    --install) ACTION="install"; shift ;;
    --uninstall) ACTION="uninstall"; shift ;;
    --show) ACTION="show"; shift ;;
    --help|-h) print_help; exit 0 ;;
    --dry-run) DRY_RUN=1; shift ;;
    --verbose) VERBOSE=1; shift ;;
    --force) FORCE=1; shift ;;
    *) die "Unknown argument: $1";;
  esac
done

# ensure inside a git repo
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [ -z "$REPO_ROOT" ]; then
  die "Not inside a Git repository (git rev-parse failed). Run this from inside the repository."
fi

HOOKS_DIR="$REPO_ROOT/$HOOKS_DIR_REL"

# Ensure .githooks exists for install; for show/uninstall we can proceed without it
case "$ACTION" in
  install)
    if [ ! -d "$HOOKS_DIR" ]; then
      die "Hooks directory not found: $HOOKS_DIR_REL (expected at: $HOOKS_DIR). Create it or run from the repo root."
    fi

    current="$(git config --local core.hooksPath || true)"

    if [ -n "$current" ]; then
      info "Current core.hooksPath is: $current"
    else
      info "No core.hooksPath currently set (using default .git/hooks)"
    fi

    # Save existing value unless already saved (or if --force)
    saved="$(git config --local "$BACKUP_CONFIG_KEY" || true)"
    if [ -n "$saved" ] && [ "$FORCE" -eq 0 ]; then
      echo "A previous hooksPath backup is already stored in git config key '$BACKUP_CONFIG_KEY':"
      echo "  $saved"
      echo "Use --force to overwrite the saved backup, or run --show to inspect / --uninstall to restore."
      exit 1
    fi

    if [ "$DRY_RUN" -ne 0 ]; then
      echo "[DRY-RUN] Would save existing core.hooksPath (if any) to '$BACKUP_CONFIG_KEY': '$current'"
      echo "[DRY-RUN] Would set core.hooksPath to: $HOOKS_DIR_REL"
      exit 0
    fi

    # Save previous value (may be empty string)
    if [ -n "$current" ]; then
      git config --local "$BACKUP_CONFIG_KEY" "$current"
      info "Saved previous core.hooksPath to git config key '$BACKUP_CONFIG_KEY'."
    else
      # Save empty marker to indicate we replaced the default
      git config --local "$BACKUP_CONFIG_KEY" ""
      info "Saved empty previous core.hooksPath marker to git config key '$BACKUP_CONFIG_KEY'."
    fi

    # Make files in .githooks executable (common expectation for hook scripts)
    # This is benign if files are already executable.
    info "Ensuring hook scripts in '$HOOKS_DIR_REL' are executable..."
    find "$HOOKS_DIR" -type f -maxdepth 2 -print0 | while IFS= read -r -d '' f; do
      if [ -x "$f" ]; then
        info "Already executable: $f"
      else
        chmod +x "$f"
        info "Made executable: $f"
      fi
    done

    # Set core.hooksPath to the relative path (relative paths are preferred)
    git config --local core.hooksPath "$HOOKS_DIR_REL"
    echo "Installed .githooks as repository hooks path (core.hooksPath = $HOOKS_DIR_REL)."
    echo "Saved previous value (may be empty) to git config key '$BACKUP_CONFIG_KEY'."
    echo "To undo: $0 --uninstall"

    ;;

  uninstall)
    saved="$(git config --local "$BACKUP_CONFIG_KEY" || true)"
    current="$(git config --local core.hooksPath || true)"

    if [ "$DRY_RUN" -ne 0 ]; then
      echo "[DRY-RUN] Current core.hooksPath: '$current'"
      echo "[DRY-RUN] Saved backup value ('$BACKUP_CONFIG_KEY'): '$saved'"
      if [ -z "$saved" ]; then
        echo "[DRY-RUN] Would unset core.hooksPath (restore to default .git/hooks)."
      else
        echo "[DRY-RUN] Would restore core.hooksPath to: '$saved'"
      fi
      exit 0
    fi

    if [ -z "$saved" ]; then
      # No saved backup - remove core.hooksPath entirely (fallback to .git/hooks)
      git config --local --unset core.hooksPath 2>/dev/null || true
      git config --local --unset "$BACKUP_CONFIG_KEY" 2>/dev/null || true
      echo "No saved backup found. core.hooksPath unset; Git will use default .git/hooks."
    else
      # Restore previous value (may be empty string marker)
      # If saved is an empty string we interpret that as 'no value' originally
      if [ -z "$saved" ]; then
        git config --local --unset core.hooksPath 2>/dev/null || true
        echo "Restored original state: core.hooksPath unset (using default .git/hooks)."
      else
        git config --local core.hooksPath "$saved"
        echo "Restored core.hooksPath to saved value: $saved"
      fi
      git config --local --unset "$BACKUP_CONFIG_KEY" 2>/dev/null || true
    fi

    ;;

  show)
    current="$(git config --local core.hooksPath || true)"
    saved="$(git config --local "$BACKUP_CONFIG_KEY" || true)"

    echo "Repository: $REPO_ROOT"
    if [ -n "$current" ]; then
      echo "Current core.hooksPath: $current"
    else
      echo "Current core.hooksPath: (unset) -> using default .git/hooks"
    fi

    if [ -n "$saved" ]; then
      # Show marker if empty string intentionally stored
      if [ -z "$saved" ]; then
        echo "Saved backup ($BACKUP_CONFIG_KEY): (empty marker) -> original was unset"
      else
        echo "Saved backup ($BACKUP_CONFIG_KEY): $saved"
      fi
    else
      echo "No saved backup under git config key: $BACKUP_CONFIG_KEY"
    fi

    echo ""
    echo "Local .githooks directory exists at: $HOOKS_DIR"
    if [ -d "$HOOKS_DIR" ]; then
      echo "Hooks found (listing up to 20 entries):"
      find "$HOOKS_DIR" -maxdepth 2 -type f -printf "  %p (%m)\n" | head -n 20 || true
    else
      echo "Warning: $HOOKS_DIR_REL not present in repository."
    fi
    ;;

  *)
    die "Unknown action. Use --help to see available commands."
    ;;
esac

exit 0
