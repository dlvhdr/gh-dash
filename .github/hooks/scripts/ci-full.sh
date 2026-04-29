#!/bin/bash
# Full CI gate for agentStop.
# Outputs {"decision":"block","reason":"..."} on failure so the agent
# receives the error output as context and continues working to fix it.
# Always exits 0 — a non-zero exit signals a hook infrastructure error,
# not a block; the decision field controls the gate.
set -uo pipefail

output=$(task fmt && \
         task lint && \
         go build ./... && \
         task test ./... 2>&1) || {
  jq -n --arg r "CI gate failed. Fix all issues before finishing:

$output" '{"decision":"block","reason":$r}'
  exit 0
}

echo '{"decision":"approve"}'
