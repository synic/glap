# Shared bash helpers for example test scripts.
# Sourced by examples/*/test.sh. Not meant to be run directly.
#
# Usage in a test.sh:
#   #!/usr/bin/env bash
#   set -euo pipefail
#   SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
#   source "$SCRIPT_DIR/../test-lib.sh"
#   BIN="${1:?usage: test.sh <binary>}"
#   EXAMPLE_NAME="myexample"
#
#   out=$("$BIN" --flag 2>&1) || true
#   assert_contains "expected output" "$out"
#   pass "scenario-name"
#
#   test_summary

if [ -t 1 ]; then
    _RED=$'\033[31m'
    _GREEN=$'\033[32m'
    _DIM=$'\033[2m'
    _RESET=$'\033[0m'
else
    _RED=""
    _GREEN=""
    _DIM=""
    _RESET=""
fi

_PASSED=0
_FAILED=0

pass() {
    _PASSED=$((_PASSED + 1))
    printf "  %sok%s  %s\n" "$_GREEN" "$_RESET" "$1"
}

fail() {
    _FAILED=$((_FAILED + 1))
    printf "  %sFAIL%s %s\n" "$_RED" "$_RESET" "$1"
    if [ $# -ge 2 ]; then
        printf "       %s%s%s\n" "$_DIM" "$2" "$_RESET"
    fi
}

assert_contains() {
    local needle="$1"
    local haystack="$2"
    local name="${3:-assert_contains}"
    if printf '%s' "$haystack" | grep -qF -- "$needle"; then
        pass "$name"
    else
        fail "$name" "expected to find: $needle"
        printf "       %sgot: %s%s\n" "$_DIM" "$haystack" "$_RESET"
    fi
}

assert_not_contains() {
    local needle="$1"
    local haystack="$2"
    local name="${3:-assert_not_contains}"
    if printf '%s' "$haystack" | grep -qF -- "$needle"; then
        fail "$name" "expected to NOT find: $needle"
        printf "       %sgot: %s%s\n" "$_DIM" "$haystack" "$_RESET"
    else
        pass "$name"
    fi
}

assert_exit() {
    local expected="$1"
    local actual="$2"
    local name="${3:-assert_exit}"
    if [ "$expected" = "$actual" ]; then
        pass "$name"
    else
        fail "$name" "expected exit $expected, got $actual"
    fi
}

test_summary() {
    local name="${EXAMPLE_NAME:-example}"
    printf "%s: %d passed, %d failed\n" "$name" "$_PASSED" "$_FAILED"
    if [ "$_FAILED" -gt 0 ]; then
        exit 1
    fi
}
