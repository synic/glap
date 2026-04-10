#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="completions"

# Bash completion script dispatch.
out=$(COMPLETE=bash "$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "bash completion exits 0"
assert_contains "complete -F" "$out" "bash script has 'complete -F'"

# Zsh completion.
out=$(COMPLETE=zsh "$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "zsh completion exits 0"
assert_contains "#compdef" "$out" "zsh script has '#compdef'"

# Fish completion.
out=$(COMPLETE=fish "$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "fish completion exits 0"
assert_contains "complete -c myapp" "$out" "fish script has 'complete -c myapp'"

# PowerShell completion.
out=$(COMPLETE=powershell "$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "powershell completion exits 0"
assert_contains "Register-ArgumentCompleter" "$out" "powershell script has Register-ArgumentCompleter"

# Normal invocation without COMPLETE env var.
out=$("$BIN" serve 2>&1)
rc=$?
assert_exit 0 "$rc" "normal serve exits 0"
assert_contains "Command: serve" "$out" "serve subcommand runs"

test_summary
