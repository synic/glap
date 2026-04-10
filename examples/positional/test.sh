#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="positional"

# Happy path: two positionals.
out=$("$BIN" a.txt b.txt 2>&1)
rc=$?
assert_exit 0 "$rc" "happy path exits 0"
assert_contains "Copying a.txt -> b.txt" "$out" "happy path output"

# --force flag alters prefix.
out=$("$BIN" -f a.txt b.txt 2>&1)
rc=$?
assert_exit 0 "$rc" "force path exits 0"
assert_contains "Force copying a.txt -> b.txt" "$out" "force path output"

# Missing required positional.
out=$("$BIN" a.txt 2>&1)
rc=$?
assert_exit 1 "$rc" "missing positional exits 1"

# --help shows ARGS: section with positional help text.
out=$("$BIN" --help 2>&1) || true
assert_contains "ARGS:" "$out" "help has ARGS section"
assert_contains "<SRC>" "$out" "help shows SRC placeholder"
assert_contains "<DST>" "$out" "help shows DST placeholder"
assert_contains "Source file" "$out" "help shows SRC help text"
assert_contains "Destination file" "$out" "help shows DST help text"

test_summary
