#!/bin/bash
# Restore pool-2-6-2 test fixture
# Usage: ./test/fixtures/pool-2-6-2/restore.sh

set -e
cd "$(dirname "$0")/../../.."

FIXTURE_DIR="test/fixtures/pool-2-6-2"
EPIC_ID=$(cat "$FIXTURE_DIR/EPIC_ID")

echo "Restoring pool-2-6-2 fixture (epic: $EPIC_ID)..."

# Copy fixture files to .tick/issues/
cp "$FIXTURE_DIR"/*.json .tick/issues/

# Reset all tasks to open status
for f in .tick/issues/*.json; do
  id=$(basename "$f" .json)
  # Skip if not part of this epic
  if ! grep -q "\"parent\":\"$EPIC_ID\"" "$f" 2>/dev/null && [ "$id" != "$EPIC_ID" ]; then
    continue
  fi
  # Reset status to open (using jq if available, else sed)
  if command -v jq &> /dev/null; then
    jq '.status = "open" | del(.closed_at) | del(.closed_reason) | del(.started_at)' "$f" > "$f.tmp" && mv "$f.tmp" "$f"
  else
    sed -i '' 's/"status":"[^"]*"/"status":"open"/g' "$f"
  fi
done

echo "Fixture restored. Run with:"
echo "  ./tk graph $EPIC_ID"
echo "  ./tk run $EPIC_ID --pool --cloud --board"
