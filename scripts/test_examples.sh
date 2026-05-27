#!/usr/bin/env bash
# test_examples.sh — run all .din examples and report pass/fail
# Usage: ./scripts/test_examples.sh [timeout_seconds]
#   timeout_seconds defaults to 10
#
# Exit code: 0 if all non-skipped examples pass, 1 otherwise.
#
# Files listed in SKIP require external services (DB, network) or are
# intentional long-running servers — excluded from the test run.

BINARY="${BINARY:-./uddinlang}"
TIMEOUT="${1:-10}"

SKIP=(
  "examples/database/mysql_stream.din"
  "examples/database/postgres_stream.din"
  "examples/networking-and-http/http_response_return_demo.din"
  "examples/networking-and-http/persistent_http_server.din"
)

is_skipped() {
  local f="$1"
  for s in "${SKIP[@]}"; do
    [[ "$f" == "$s" ]] && return 0
  done
  return 1
}

if [[ ! -x "$BINARY" ]]; then
  echo "Binary not found: $BINARY — run 'make build' first"
  exit 1
fi

pass=0
fail=0
skip=0
timeout_count=0

while IFS= read -r f; do
  if is_skipped "$f"; then
    echo "SKIP: $f"
    skip=$((skip+1))
    continue
  fi

  output=$(timeout "$TIMEOUT" "$BINARY" "$f" 2>&1)
  ec=$?

  if [[ $ec -eq 124 ]]; then
    echo "TIMEOUT: $f"
    timeout_count=$((timeout_count+1))
    fail=$((fail+1))
  elif [[ $ec -ne 0 ]]; then
    echo "FAIL: $f"
    echo "$output" | grep -iE "(error at|panic:|type error|value error|runtime error)" 2>/dev/null | head -3 | sed 's/^/  /' || true
    fail=$((fail+1))
  else
    echo "OK: $f"
    pass=$((pass+1))
  fi
done < <(find examples/ -name "*.din" | sort)

echo ""
echo "Results: ${pass} passed, ${fail} failed, ${skip} skipped (${timeout_count} timeouts)"

[[ $fail -eq 0 ]]
