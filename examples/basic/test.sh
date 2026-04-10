#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="basic"

# Happy path with all args.
out=$("$BIN" --config app.yaml -v --port 3000 -t a,b,c 42 2>&1)
rc=$?
assert_exit 0 "$rc" "happy path exits 0"
assert_contains "Config:  app.yaml" "$out" "config set"
assert_contains "Verbose: true" "$out" "verbose set"
assert_contains "Port:    3000" "$out" "port set via CLI"
assert_contains "Output:  text" "$out" "output default"
assert_contains "Tags:    [a b c]" "$out" "delimiter splits tags"
assert_contains "Offset:  42" "$out" "positional captured"

# ArgRequiredElseHelp: empty args produces help output.
out=$("$BIN" 2>&1)
rc=$?
assert_exit 1 "$rc" "empty args exits 1"
assert_contains "USAGE:" "$out" "empty args prints help"

# Missing required --config.
out=$("$BIN" --port 3000 2>&1)
rc=$?
assert_exit 1 "$rc" "missing --config exits 1"
assert_contains "config" "$out" "error mentions config"

# Invalid possible value.
out=$("$BIN" --config app.yaml --output invalid 2>&1)
rc=$?
assert_exit 1 "$rc" "invalid possible value exits 1"
assert_contains "invalid" "$out" "error mentions bad value"

test_summary
