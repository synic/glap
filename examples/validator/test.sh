#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="validator"

# Happy path with explicit values.
out=$("$BIN" --port 8080 --bind 127.0.0.1 2>&1)
rc=$?
assert_exit 0 "$rc" "happy path exits 0"
assert_contains "Listening on 127.0.0.1:8080" "$out" "happy output"

# Defaults are valid.
out=$("$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "defaults exit 0"
assert_contains "Listening on 127.0.0.1:8080" "$out" "defaults output"

# Port out of range.
out=$("$BIN" --port 99999 2>&1)
rc=$?
assert_exit 1 "$rc" "out-of-range port exits 1"
assert_contains "between 1 and 65535" "$out" "port range error"

# Port not a number.
out=$("$BIN" --port notanumber 2>&1)
rc=$?
assert_exit 1 "$rc" "non-numeric port exits 1"
assert_contains "must be a number" "$out" "port number error"

# Invalid IP.
out=$("$BIN" --bind not.an.ip 2>&1)
rc=$?
assert_exit 1 "$rc" "bad IP exits 1"
assert_contains "valid IP address" "$out" "IP error"

test_summary
