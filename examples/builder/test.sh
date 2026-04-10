#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="builder"

# Root command with required config.
out=$("$BIN" --config foo.yaml 2>&1)
rc=$?
assert_exit 0 "$rc" "root exits 0 with config"
assert_contains "Config:    foo.yaml" "$out" "config set"
assert_contains "Verbosity: 0" "$out" "verbosity defaults to 0"

# Count action.
out=$("$BIN" --config foo.yaml -vvv 2>&1)
rc=$?
assert_exit 0 "$rc" "-vvv exits 0"
assert_contains "Verbosity: 3" "$out" "count action"

# Append action.
out=$("$BIN" --config foo.yaml -t one -t two 2>&1)
rc=$?
assert_exit 0 "$rc" "append exits 0"
assert_contains "[one two]" "$out" "append accumulates"

# Subcommand with possible values.
out=$("$BIN" --config foo.yaml deploy --target staging 2>&1)
rc=$?
assert_exit 0 "$rc" "deploy staging exits 0"
assert_contains "Deploying to staging" "$out" "deploy target"
assert_contains "dry-run: false" "$out" "dry-run default"

out=$("$BIN" --config foo.yaml deploy --target production -n 2>&1)
rc=$?
assert_exit 0 "$rc" "deploy production dry-run exits 0"
assert_contains "dry-run: true" "$out" "dry-run set"

# Invalid subcommand target.
out=$("$BIN" --config foo.yaml deploy --target bogus 2>&1)
rc=$?
assert_exit 1 "$rc" "invalid target exits 1"

# Missing required config on root.
out=$("$BIN" 2>&1)
rc=$?
assert_exit 1 "$rc" "missing config exits 1"

test_summary
