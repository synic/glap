#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../test-lib.sh"

BIN="${1:?usage: test.sh <binary>}"
EXAMPLE_NAME="env-override"

# Scrub any ambient env vars the user may have set.
unset APP_HOST APP_PORT APP_DEBUG APP_SECRET

# Defaults: no env, no flags.
out=$("$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "defaults exit 0"
assert_contains "Host:   localhost" "$out" "host default"
assert_contains "Port:   8080" "$out" "port default"
assert_contains "Debug:  false" "$out" "debug default"

# Env var overrides default.
out=$(APP_PORT=9090 APP_HOST=0.0.0.0 "$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "env override exits 0"
assert_contains "Host:   0.0.0.0" "$out" "host from env"
assert_contains "Port:   9090" "$out" "port from env"

# CLI beats env.
out=$(APP_PORT=9090 APP_HOST=0.0.0.0 "$BIN" --port 3000 2>&1)
rc=$?
assert_exit 0 "$rc" "cli+env exits 0"
assert_contains "Host:   0.0.0.0" "$out" "host still from env"
assert_contains "Port:   3000" "$out" "port from CLI wins over env"

# APP_DEBUG env var.
out=$(APP_DEBUG=true "$BIN" 2>&1)
rc=$?
assert_exit 0 "$rc" "debug env exits 0"
assert_contains "Debug:  true" "$out" "debug from env"

test_summary
