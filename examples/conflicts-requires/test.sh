#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="conflicts-requires"

# Happy path with --stdin and --json.
out=$("$BIN" --stdin --json 2>&1)
rc=$?
assert_exit 0 "$rc" "stdin + json exits 0"
assert_contains "Using JSON output" "$out" "json output selected"
assert_contains "Reading from stdin" "$out" "stdin path"

# Conflict: --json and --text.
out=$("$BIN" --input foo --json --text 2>&1)
rc=$?
assert_exit 1 "$rc" "json+text conflict exits 1"
assert_contains "conflicts" "$out" "error mentions conflict"

# Requires: --output without --format.
out=$("$BIN" --input foo -o out.txt 2>&1)
rc=$?
assert_exit 1 "$rc" "output without format exits 1"
assert_contains "format" "$out" "error mentions format"

# required_unless: satisfied by --stdin alone.
out=$("$BIN" --stdin 2>&1)
rc=$?
assert_exit 0 "$rc" "stdin alone exits 0"

# required_unless: --input missing without --stdin.
out=$("$BIN" 2>&1)
rc=$?
assert_exit 1 "$rc" "missing input without stdin exits 1"

test_summary
