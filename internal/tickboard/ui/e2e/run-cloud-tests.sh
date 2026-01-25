#!/bin/bash
#
# Unified Cloud Integration Test Runner
#
# Runs the full cloud integration test suite:
# 1. Starts wrangler dev
# 2. Sets up test auth
# 3. Starts test rig as local agent
# 4. Runs vitest integration tests
# 5. Runs E2E browser tests
# 6. Cleans up
#
# Usage:
#   ./run-cloud-tests.sh [--skip-e2e] [--skip-vitest]
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
UI_DIR="$PROJECT_ROOT/internal/tickboard/ui"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# State
WRANGLER_STARTED=false
TESTRIG_PID=""
VITEST_RESULT=0
E2E_RESULT=0

# Flags
SKIP_E2E=false
SKIP_VITEST=false

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --skip-e2e)
      SKIP_E2E=true
      shift
      ;;
    --skip-vitest)
      SKIP_VITEST=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--skip-e2e] [--skip-vitest]"
      exit 1
      ;;
  esac
done

log() {
  echo -e "${BLUE}[cloud-tests]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[cloud-tests]${NC} $1"
}

log_error() {
  echo -e "${RED}[cloud-tests]${NC} $1"
}

log_warn() {
  echo -e "${YELLOW}[cloud-tests]${NC} $1"
}

cleanup() {
  log "Cleaning up..."

  # Stop test rig
  if [ -n "$TESTRIG_PID" ]; then
    log "Stopping test rig (PID: $TESTRIG_PID)..."
    kill "$TESTRIG_PID" 2>/dev/null || true
    wait "$TESTRIG_PID" 2>/dev/null || true
  fi

  # Stop wrangler dev
  if [ "$WRANGLER_STARTED" = true ]; then
    log "Stopping wrangler dev..."
    "$SCRIPT_DIR/wrangler-ctl.sh" stop 2>/dev/null || true
  fi

  log "Cleanup complete"
}

# Always cleanup on exit
trap cleanup EXIT

check_prerequisites() {
  log "Checking prerequisites..."

  # Check wrangler
  if [ ! -f "$PROJECT_ROOT/cloud/worker/node_modules/.bin/wrangler" ]; then
    log_error "wrangler not installed in cloud/worker. Run: cd cloud/worker && pnpm install"
    exit 1
  fi
  log "  wrangler: OK"

  # Check test rig is built
  if ! command -v go &> /dev/null; then
    log_error "go not installed"
    exit 1
  fi
  log "  go: OK"

  # Check agent-browser (only if running E2E)
  if [ "$SKIP_E2E" = false ]; then
    if ! command -v agent-browser &> /dev/null; then
      log_error "agent-browser not installed (required for E2E tests)"
      exit 1
    fi
    log "  agent-browser: OK"
  fi

  log_success "Prerequisites OK"
}

start_wrangler() {
  log "Starting wrangler dev..."

  WRANGLER_OUTPUT=$("$SCRIPT_DIR/wrangler-ctl.sh" start 2>&1)
  WRANGLER_STARTED=true

  # Extract URL from output
  WRANGLER_URL=$(echo "$WRANGLER_OUTPUT" | grep -o 'http://localhost:[0-9]*' | head -1)

  if [ -z "$WRANGLER_URL" ]; then
    log_error "Failed to get wrangler dev URL"
    echo "$WRANGLER_OUTPUT"
    exit 1
  fi

  log_success "wrangler dev running at $WRANGLER_URL"
  export WRANGLER_URL
}

setup_auth() {
  log "Setting up test authentication..."

  AUTH_OUTPUT=$("$SCRIPT_DIR/setup-cloud-auth.sh" 2>&1)

  # Extract credentials from output (match actual variable names from setup script)
  export TEST_TOKEN=$(echo "$AUTH_OUTPUT" | grep 'TEST_AUTH_TOKEN=' | sed 's/.*TEST_AUTH_TOKEN="//' | sed 's/"$//')
  export TEST_PROJECT=$(echo "$AUTH_OUTPUT" | grep 'TEST_PROJECT_ID=' | sed 's/.*TEST_PROJECT_ID="//' | sed 's/"$//')
  export TEST_USER_ID=$(echo "$AUTH_OUTPUT" | grep 'TEST_USER_ID=' | sed 's/.*TEST_USER_ID="//' | sed 's/"$//')

  if [ -z "$TEST_TOKEN" ] || [ -z "$TEST_PROJECT" ]; then
    log_error "Failed to get auth credentials"
    echo "$AUTH_OUTPUT"
    exit 1
  fi

  log_success "Auth setup complete (project: $TEST_PROJECT)"
}

start_testrig() {
  log "Starting test rig as local agent..."

  # Build test rig
  cd "$PROJECT_ROOT"
  go build -o /tmp/testrig ./cmd/testrig

  # Convert HTTP URL to WebSocket URL
  WS_URL=$(echo "$WRANGLER_URL" | sed 's/http:/ws:/')

  # Start test rig with upstream connection
  /tmp/testrig \
    --port 18788 \
    --upstream "$WS_URL" \
    --project "$TEST_PROJECT" \
    --token "$TEST_TOKEN" \
    > /tmp/testrig-cloud.log 2>&1 &

  TESTRIG_PID=$!

  # Wait for test rig to be ready
  sleep 2

  if ! kill -0 "$TESTRIG_PID" 2>/dev/null; then
    log_error "Test rig failed to start"
    cat /tmp/testrig-cloud.log
    exit 1
  fi

  log_success "Test rig running (PID: $TESTRIG_PID)"
}

run_vitest() {
  if [ "$SKIP_VITEST" = true ]; then
    log_warn "Skipping vitest integration tests (--skip-vitest)"
    return
  fi

  log "Running vitest integration tests..."

  cd "$UI_DIR"

  # Run only cloud-do tests
  if npm test -- --run src/comms/cloud-do.test.ts 2>&1; then
    log_success "Vitest tests passed"
    VITEST_RESULT=0
  else
    log_error "Vitest tests failed"
    VITEST_RESULT=1
  fi
}

run_e2e() {
  if [ "$SKIP_E2E" = true ]; then
    log_warn "Skipping E2E browser tests (--skip-e2e)"
    return
  fi

  log "Running E2E browser tests..."

  if "$SCRIPT_DIR/cloud-e2e.sh" 2>&1; then
    log_success "E2E tests passed"
    E2E_RESULT=0
  else
    log_error "E2E tests failed"
    E2E_RESULT=1
  fi
}

print_summary() {
  echo ""
  echo "========================================"
  echo "Cloud Integration Test Results"
  echo "========================================"

  if [ "$SKIP_VITEST" = false ]; then
    if [ $VITEST_RESULT -eq 0 ]; then
      echo -e "  Vitest:  ${GREEN}PASSED${NC}"
    else
      echo -e "  Vitest:  ${RED}FAILED${NC}"
    fi
  else
    echo -e "  Vitest:  ${YELLOW}SKIPPED${NC}"
  fi

  if [ "$SKIP_E2E" = false ]; then
    if [ $E2E_RESULT -eq 0 ]; then
      echo -e "  E2E:     ${GREEN}PASSED${NC}"
    else
      echo -e "  E2E:     ${RED}FAILED${NC}"
    fi
  else
    echo -e "  E2E:     ${YELLOW}SKIPPED${NC}"
  fi

  echo "========================================"
}

main() {
  echo "========================================"
  echo "Cloud Integration Test Suite"
  echo "========================================"

  check_prerequisites
  start_wrangler
  setup_auth
  start_testrig

  run_vitest
  run_e2e

  print_summary

  # Return failure if any tests failed
  if [ $VITEST_RESULT -ne 0 ] || [ $E2E_RESULT -ne 0 ]; then
    exit 1
  fi
}

main "$@"
