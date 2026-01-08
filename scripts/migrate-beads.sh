#!/usr/bin/env bash
set -euo pipefail

# Migrate beads issues to tick format.
# Requires: jq

if [[ ! -d .tick/issues ]]; then
  echo "Missing .tick/issues. Run tk init first." >&2
  exit 1
fi

bd list --json | jq -c '.[]' | while read -r bead; do
  old_id=$(echo "$bead" | jq -r '.id')
  new_id=$(echo "$old_id" | sed 's/bd-//' | cut -c1-3)

  echo "$bead" | jq --arg id "$new_id" '
    .id = $id |
    del(.external_ref, .content_hash, .source_repo)
  ' > ".tick/issues/${new_id}.json"
done
