#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="groups"

# Happy path: pick one format.
out=$("$BIN" -i file.txt --json 2>&1)
rc=$?
assert_exit 0 "$rc" "json format exits 0"
assert_contains "Processing file.txt with json output" "$out" "json output"

out=$("$BIN" -i file.txt --yaml 2>&1)
rc=$?
assert_exit 0 "$rc" "yaml format exits 0"
assert_contains "yaml output" "$out" "yaml output"

# Mutual exclusion: two formats.
out=$("$BIN" -i file.txt --json --text 2>&1)
rc=$?
assert_exit 1 "$rc" "two formats exits 1"
assert_contains "mutually exclusive" "$out" "error mentions exclusivity"

# Required group: no format at all.
out=$("$BIN" -i file.txt 2>&1)
rc=$?
assert_exit 1 "$rc" "no format exits 1"
assert_contains "required" "$out" "error mentions required"

# Required --input.
out=$("$BIN" --json 2>&1)
rc=$?
assert_exit 1 "$rc" "missing input exits 1"

# Help output shows group annotation.
out=$("$BIN" --help 2>&1) || true
assert_contains "[group: format]" "$out" "help shows group annotation"

test_summary
