#!/bin/bash
# Quick CI gate for subagentStop.
# Skips build and tests — build artifacts may not exist mid-loop, and the
# developer agent runs `task test ./...` internally before marking work done.
# `task lint` runs golangci-lint which already invokes `go vet` + type-check
# style passes, so this catches most mistakes early without paying for a full
# test run.
set -uo pipefail

output=$(task lint 2>&1) || {
  jq -n --arg r "Quick CI check failed (task lint). Fix before finishing:

$output" '{"decision":"block","reason":$r}'
  exit 0
}

echo '{"decision":"approve"}'
