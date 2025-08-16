#!/bin/bash
# Dry-run wrapper for zeroui binary calls used by scripts
echo "[DRY-RUN WRAPPER] called: $0 $@"
# Print simulated output depending on subcommand
case "$1" in
  batch-extract)
    echo "Simulating batch-extract with args: $@"
    mkdir -p "$PWD/tmp-configs"
    echo "settings: {}" > "$PWD/tmp-configs/example1.yaml"
    echo "settings: {}" > "$PWD/tmp-configs/example2.yaml"
    exit 0
    ;;
  batch-extract*)
    echo "Simulating batch-extract"
    exit 0
    ;;
  ref|reference)
    echo "Simulating reference subcommand: $@"
    exit 0
    ;;
  *)
    echo "Simulating command: $@"
    exit 0
    ;;
esac
