#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="subcommands"

unset PORT

# serve subcommand.
out=$("$BIN" serve --port 9000 2>&1)
rc=$?
assert_exit 0 "$rc" "serve exits 0"
assert_contains "Serving on localhost:9000" "$out" "serve output"

# serve defaults.
out=$("$BIN" serve 2>&1)
rc=$?
assert_exit 0 "$rc" "serve defaults exit 0"
assert_contains "Serving on localhost:8080" "$out" "serve defaults"

# init with required positional.
out=$("$BIN" init myproject --template minimal 2>&1)
rc=$?
assert_exit 0 "$rc" "init exits 0"
assert_contains "Initializing project \"myproject\"" "$out" "init output"
assert_contains "minimal" "$out" "template set"

# init with invalid template.
out=$("$BIN" init myproject --template bogus 2>&1)
rc=$?
assert_exit 1 "$rc" "bad template exits 1"

# init missing required positional.
out=$("$BIN" init 2>&1)
rc=$?
assert_exit 1 "$rc" "init missing name exits 1"

# Nested subcommand: remote add.
out=$("$BIN" remote add origin https://example.com 2>&1)
rc=$?
assert_exit 0 "$rc" "remote add exits 0"
assert_contains "Adding remote \"origin\"" "$out" "remote add output"
assert_contains "https://example.com" "$out" "remote URL"

# Global arg propagation: -v before subcommand.
out=$("$BIN" -v serve 2>&1)
rc=$?
assert_exit 0 "$rc" "global verbose exits 0"
assert_contains "Verbose: true" "$out" "global verbose propagated"

test_summary
