#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="trailing-var-arg"

# Runs echo with args.
out=$("$BIN" echo hello world 2>&1)
rc=$?
assert_exit 0 "$rc" "echo exits 0"
assert_contains "hello world" "$out" "echo output"
assert_not_contains "Running:" "$out" "no verbose prefix when --verbose absent"

# Verbose adds a prefix and still runs the command.
out=$("$BIN" --verbose echo hello 2>&1)
rc=$?
assert_exit 0 "$rc" "verbose echo exits 0"
assert_contains "Running: [echo hello]" "$out" "verbose prints command"
assert_contains "hello" "$out" "echo still runs"

# No command given -> error.
out=$("$BIN" 2>&1)
rc=$?
assert_exit 1 "$rc" "no command exits 1"
assert_contains "no command specified" "$out" "missing-command error"

test_summary
