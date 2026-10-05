#!/usr/bin/env bash
set -euo pipefail

# Resolution scoring script for issue branch PRs
# Weighted scoring (must hit >= 95%):
#   Tests pass:     30%
#   Benchmarks:     20%
#   Brief:          25%
#   Prototype:      25%

THRESHOLD=95
TOTAL=0

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log() { echo -e "${GREEN}[resolution]${NC} $*"; }
fail() { echo -e "${RED}[resolution]${NC} $*"; }

log "Running tests..."
if go test ./... > /dev/null 2>&1; then
    TESTS=30
    log "Tests: PASS (+30)"
else
    TESTS=0
    fail "Tests: FAIL (+0)"
fi
TOTAL=$((TOTAL + TESTS))

log "Running benchmarks..."
if go test -bench=. -benchmem -run='^$' ./... > /dev/null 2>&1; then
    BENCH=20
    log "Benchmarks: PASS (+20)"
else
    BENCH=0
    fail "Benchmarks: FAIL (+0)"
fi
TOTAL=$((TOTAL + BENCH))

log "Checking brief..."
BRIEF=0
BRIEF_PATH=""
if [ -d "discovery/briefs" ]; then
    BRIEF_PATH=$(find discovery/briefs -name "*.md" 2>/dev/null | head -1 || true)
fi
if [ -n "$BRIEF_PATH" ] && [ -f "$BRIEF_PATH" ]; then
    BRIEF=5
    for section in "Issue" "Research" "Architecture" "Recommended solution paths" "Implementation plan" "Risks" "Goals" "Success Criteria"; do
        if grep -qF "## $section" "$BRIEF_PATH" 2>/dev/null; then
            BRIEF=$((BRIEF + 3))
        fi
    done
    [ $BRIEF -gt 25 ] && BRIEF=25
    log "Brief: $BRIEF/25"
else
    fail "Brief: missing (+0)"
fi
TOTAL=$((TOTAL + BRIEF))

log "Checking prototype..."
PROTO=0
PROTO_DIR=""
if [ -d "discovery/prototypes" ]; then
    PROTO_DIR=$(find discovery/prototypes -type d -maxdepth 1 2>/dev/null | tail -n +2 | head -1 || true)
fi
if [ -n "$PROTO_DIR" ] && [ -d "$PROTO_DIR" ]; then
    [ -f "$PROTO_DIR/main.go" ] && PROTO=$((PROTO + 8))
    [ -f "$PROTO_DIR/main_test.go" ] && PROTO=$((PROTO + 7))
    [ -f "$PROTO_DIR/README.md" ] && PROTO=$((PROTO + 5))
    if go vet "$PROTO_DIR/..." > /dev/null 2>&1; then
        PROTO=$((PROTO + 5))
    fi
    [ $PROTO -gt 25 ] && PROTO=25
    log "Prototype: $PROTO/25"
else
    fail "Prototype: missing (+0)"
fi
TOTAL=$((TOTAL + PROTO))

log "Total score: $TOTAL/100"
if [ $TOTAL -ge $THRESHOLD ]; then
    log "PASSED (>= $THRESHOLD%)"
    printf '{"passed":true,"score":%d,"details":{"tests":%d,"benchmarks":%d,"brief":%d,"code":%d}}\n' "$TOTAL" "$TESTS" "$BENCH" "$BRIEF" "$PROTO"
else
    fail "FAILED (< $THRESHOLD%)"
    printf '{"passed":false,"score":%d,"details":{"tests":%d,"benchmarks":%d,"brief":%d,"code":%d}}\n' "$TOTAL" "$TESTS" "$BENCH" "$BRIEF" "$PROTO"
    exit 1
fi
