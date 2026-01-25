#!/bin/bash
#
# Setup test authentication for wrangler dev's local D1 database
#
# This script creates test user, token, and board for e2e cloud integration tests.
# It is idempotent - safe to run multiple times.
#
# Prerequisites:
#   - wrangler installed
#   - Running from the ticks repo root or cloud/worker directory
#
# Usage:
#   ./internal/tickboard/ui/e2e/setup-cloud-auth.sh
#
# Output:
#   Exports TEST_AUTH_TOKEN and TEST_PROJECT_ID for use in tests
#

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test credentials
TEST_USER_ID="test-user-001"
TEST_USER_EMAIL="test@example.com"
TEST_USER_PASSWORD="testpassword123"
TEST_TOKEN_ID="test-token-001"
TEST_TOKEN_NAME="e2e-test-token"
TEST_BOARD_ID="test-board-001"
TEST_BOARD_NAME="test-project"

# SHA-256 hash function (matches auth.ts hashPassword)
# Uses openssl for portability
sha256_hash() {
  echo -n "$1" | openssl dgst -sha256 | awk '{print $2}'
}

# Find the cloud/worker directory
find_worker_dir() {
  local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

  # Try relative from script location (internal/tickboard/ui/e2e -> cloud/worker)
  if [ -f "$script_dir/../../../../cloud/worker/wrangler.toml" ]; then
    echo "$script_dir/../../../../cloud/worker"
    return
  fi

  # Try from current directory
  if [ -f "cloud/worker/wrangler.toml" ]; then
    echo "cloud/worker"
    return
  fi

  # Try if we're already in cloud/worker
  if [ -f "wrangler.toml" ] && grep -q "ticks-cloud" wrangler.toml 2>/dev/null; then
    echo "."
    return
  fi

  echo ""
}

# Execute SQL against local D1
d1_execute() {
  local sql="$1"
  wrangler d1 execute tickboard --local --command "$sql" 2>&1
}

# Execute SQL and return JSON result
d1_execute_json() {
  local sql="$1"
  wrangler d1 execute tickboard --local --command "$sql" --json 2>&1
}

# Check if wrangler is available
check_wrangler() {
  if ! command -v wrangler &> /dev/null; then
    echo -e "${RED}ERROR: wrangler not found${NC}"
    echo "Install with: npm install -g wrangler"
    exit 1
  fi
}

# Initialize the database schema if needed
init_schema() {
  echo -e "${BLUE}Initializing database schema...${NC}"

  # Check if users table exists
  local result=$(d1_execute_json "SELECT name FROM sqlite_master WHERE type='table' AND name='users'" 2>&1)

  if echo "$result" | grep -q '"users"'; then
    echo -e "  Schema already initialized"
    return 0
  fi

  # Apply schema
  if [ -f "schema.sql" ]; then
    echo -e "  Applying schema.sql..."
    wrangler d1 execute tickboard --local --file schema.sql 2>&1 | head -5
  else
    echo -e "${RED}ERROR: schema.sql not found${NC}"
    exit 1
  fi
}

# Create or update test user (idempotent)
setup_test_user() {
  echo -e "${BLUE}Setting up test user...${NC}"

  local password_hash=$(sha256_hash "$TEST_USER_PASSWORD")

  # Delete existing test user (cascade deletes tokens and boards)
  d1_execute "DELETE FROM users WHERE id = '$TEST_USER_ID'" > /dev/null 2>&1 || true

  # Insert test user
  local result=$(d1_execute "INSERT INTO users (id, email, password_hash) VALUES ('$TEST_USER_ID', '$TEST_USER_EMAIL', '$password_hash')" 2>&1)

  if echo "$result" | grep -q "error\|Error\|ERROR"; then
    echo -e "${RED}  Failed to create user: $result${NC}"
    return 1
  fi

  echo -e "${GREEN}  Created user: $TEST_USER_EMAIL (id: $TEST_USER_ID)${NC}"
}

# Create test token (idempotent)
setup_test_token() {
  echo -e "${BLUE}Setting up test token...${NC}"

  # Generate a deterministic test token (for testing - in production this would be random)
  # We use a fixed token for reproducibility in tests
  TEST_AUTH_TOKEN="e2e_test_token_0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
  local token_hash=$(sha256_hash "$TEST_AUTH_TOKEN")

  # Delete existing test token
  d1_execute "DELETE FROM tokens WHERE id = '$TEST_TOKEN_ID'" > /dev/null 2>&1 || true

  # Insert test token
  local result=$(d1_execute "INSERT INTO tokens (id, user_id, name, token_hash) VALUES ('$TEST_TOKEN_ID', '$TEST_USER_ID', '$TEST_TOKEN_NAME', '$token_hash')" 2>&1)

  if echo "$result" | grep -q "error\|Error\|ERROR"; then
    echo -e "${RED}  Failed to create token: $result${NC}"
    return 1
  fi

  echo -e "${GREEN}  Created token: $TEST_TOKEN_NAME (id: $TEST_TOKEN_ID)${NC}"
}

# Create test board/project (idempotent)
setup_test_board() {
  echo -e "${BLUE}Setting up test board...${NC}"

  # Delete existing test board
  d1_execute "DELETE FROM boards WHERE id = '$TEST_BOARD_ID'" > /dev/null 2>&1 || true

  # Insert test board
  local result=$(d1_execute "INSERT INTO boards (id, user_id, name, machine_id) VALUES ('$TEST_BOARD_ID', '$TEST_USER_ID', '$TEST_BOARD_NAME', 'test-machine')" 2>&1)

  if echo "$result" | grep -q "error\|Error\|ERROR"; then
    echo -e "${RED}  Failed to create board: $result${NC}"
    return 1
  fi

  echo -e "${GREEN}  Created board: $TEST_BOARD_NAME (id: $TEST_BOARD_ID)${NC}"
}

# Verify the setup by querying the database
verify_setup() {
  echo -e "${BLUE}Verifying setup...${NC}"

  # Check user exists
  local user_check=$(d1_execute_json "SELECT id, email FROM users WHERE id = '$TEST_USER_ID'" 2>&1)
  if ! echo "$user_check" | grep -q "$TEST_USER_EMAIL"; then
    echo -e "${RED}  User verification failed${NC}"
    return 1
  fi
  echo -e "  User: OK"

  # Check token exists
  local token_check=$(d1_execute_json "SELECT id, name FROM tokens WHERE id = '$TEST_TOKEN_ID'" 2>&1)
  if ! echo "$token_check" | grep -q "$TEST_TOKEN_NAME"; then
    echo -e "${RED}  Token verification failed${NC}"
    return 1
  fi
  echo -e "  Token: OK"

  # Check board exists
  local board_check=$(d1_execute_json "SELECT id, name FROM boards WHERE id = '$TEST_BOARD_ID'" 2>&1)
  if ! echo "$board_check" | grep -q "$TEST_BOARD_NAME"; then
    echo -e "${RED}  Board verification failed${NC}"
    return 1
  fi
  echo -e "  Board: OK"

  echo -e "${GREEN}Setup verified successfully${NC}"
}

# Output environment variables for tests
output_env_vars() {
  echo ""
  echo -e "${YELLOW}========================================${NC}"
  echo -e "${YELLOW}Test Authentication Credentials${NC}"
  echo -e "${YELLOW}========================================${NC}"
  echo ""
  echo "# Add these to your test environment or source this output:"
  echo ""
  echo "export TEST_AUTH_TOKEN=\"$TEST_AUTH_TOKEN\""
  echo "export TEST_PROJECT_ID=\"$TEST_BOARD_NAME\""
  echo "export TEST_USER_ID=\"$TEST_USER_ID\""
  echo "export TEST_USER_EMAIL=\"$TEST_USER_EMAIL\""
  echo "export TEST_BOARD_ID=\"$TEST_BOARD_ID\""
  echo ""
  echo "# For curl testing:"
  echo "# curl -H \"Authorization: Bearer \$TEST_AUTH_TOKEN\" http://localhost:8787/api/boards"
  echo ""
}

# Write env file for easy sourcing
write_env_file() {
  local env_file="${WORKER_DIR}/.test-auth.env"

  cat > "$env_file" << EOF
# Auto-generated by setup-cloud-auth.sh
# Source this file before running tests: source .test-auth.env

export TEST_AUTH_TOKEN="$TEST_AUTH_TOKEN"
export TEST_PROJECT_ID="$TEST_BOARD_NAME"
export TEST_USER_ID="$TEST_USER_ID"
export TEST_USER_EMAIL="$TEST_USER_EMAIL"
export TEST_BOARD_ID="$TEST_BOARD_ID"
export TEST_CLOUD_URL="http://localhost:8787"
EOF

  echo -e "${GREEN}Wrote test credentials to: $env_file${NC}"
  echo "Source with: source $env_file"
}

# Main
main() {
  echo "========================================"
  echo "Cloud Test Auth Setup"
  echo "========================================"
  echo ""

  check_wrangler

  WORKER_DIR=$(find_worker_dir)
  if [ -z "$WORKER_DIR" ]; then
    echo -e "${RED}ERROR: Could not find cloud/worker directory${NC}"
    echo "Run this script from the ticks repo root or cloud/worker directory"
    exit 1
  fi

  echo -e "Working directory: ${BLUE}$WORKER_DIR${NC}"
  cd "$WORKER_DIR"

  # Setup steps
  init_schema
  setup_test_user
  setup_test_token
  setup_test_board
  verify_setup

  # Output results
  output_env_vars
  write_env_file

  echo ""
  echo -e "${GREEN}Setup complete!${NC}"
  echo ""
  echo "Next steps:"
  echo "  1. Start wrangler dev: cd cloud/worker && wrangler dev"
  echo "  2. Source credentials: source cloud/worker/.test-auth.env"
  echo "  3. Run tests with the exported environment variables"
}

main "$@"
