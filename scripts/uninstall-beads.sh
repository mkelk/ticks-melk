#!/usr/bin/env bash
set -euo pipefail

# Uninstall beads from system (global config) and optionally current repo.
# Run with --repo to also clean per-repo config.
# Run with --dry-run to preview what would be removed.

CLEAN_REPO=false
DRY_RUN=false

for arg in "$@"; do
  case $arg in
    --repo) CLEAN_REPO=true ;;
    --dry-run) DRY_RUN=true ;;
    -h|--help)
      echo "Usage: $0 [--repo] [--dry-run]"
      echo "  --repo     Also clean beads config from current git repo"
      echo "  --dry-run  Preview what would be removed without removing"
      exit 0
      ;;
  esac
done

remove() {
  if [[ -e "$1" ]]; then
    if $DRY_RUN; then
      echo "[dry-run] Would remove: $1"
    else
      rm -rf "$1"
      echo "Removed: $1"
    fi
  fi
}

remove_line() {
  local file="$1"
  local pattern="$2"
  if [[ -f "$file" ]] && grep -q "$pattern" "$file" 2>/dev/null; then
    if $DRY_RUN; then
      echo "[dry-run] Would remove lines matching '$pattern' from $file"
    else
      # Use temp file for portability (macOS sed -i requires extension)
      grep -v "$pattern" "$file" > "$file.tmp" && mv "$file.tmp" "$file"
      echo "Removed lines matching '$pattern' from $file"
    fi
  fi
}

echo "=== Beads Uninstaller ==="
echo

# --- Global cleanup ---
echo "--- Global Config ---"

# Binary
remove "$HOME/.local/bin/bd"

# Plugin cache
remove "$HOME/.claude/plugins/cache/beads-marketplace"

# Claude settings.json - remove beads hooks and plugin
CLAUDE_SETTINGS="$HOME/.claude/settings.json"
if [[ -f "$CLAUDE_SETTINGS" ]]; then
  if command -v jq &>/dev/null; then
    if $DRY_RUN; then
      echo "[dry-run] Would remove beads from $CLAUDE_SETTINGS:"
      echo "  - hooks.PreCompact entries with 'bd prime'"
      echo "  - hooks.SessionStart entries with 'bd prime'"
      echo "  - enabledPlugins.beads@beads-marketplace"
    else
      # Remove beads hooks and plugin entry
      jq '
        # Remove hooks that contain "bd" commands
        .hooks.PreCompact = ((.hooks.PreCompact // []) | map(select(.hooks | all(.command | test("^bd ") | not)))) |
        .hooks.SessionStart = ((.hooks.SessionStart // []) | map(select(.hooks | all(.command | test("^bd ") | not)))) |
        # Remove empty hook arrays
        if .hooks.PreCompact == [] then del(.hooks.PreCompact) else . end |
        if .hooks.SessionStart == [] then del(.hooks.SessionStart) else . end |
        if .hooks == {} then del(.hooks) else . end |
        # Remove beads plugin
        del(.enabledPlugins["beads@beads-marketplace"])
      ' "$CLAUDE_SETTINGS" > "$CLAUDE_SETTINGS.tmp" && mv "$CLAUDE_SETTINGS.tmp" "$CLAUDE_SETTINGS"
      echo "Cleaned beads entries from $CLAUDE_SETTINGS"
    fi
  else
    echo "Warning: jq not found, cannot clean $CLAUDE_SETTINGS automatically"
    echo "Manually remove beads hooks and enabledPlugins entry"
  fi
fi

# Potential config directories
remove "$HOME/.beads"
remove "$HOME/.config/bd"
remove "$HOME/.config/beads"

# Plugin registry - remove beads entry
PLUGIN_REGISTRY="$HOME/.claude/plugins/installed_plugins.json"
if [[ -f "$PLUGIN_REGISTRY" ]]; then
  if command -v jq &>/dev/null; then
    if grep -q "beads" "$PLUGIN_REGISTRY" 2>/dev/null; then
      if $DRY_RUN; then
        echo "[dry-run] Would remove beads entries from $PLUGIN_REGISTRY"
      else
        jq 'del(.["beads@beads-marketplace"])' "$PLUGIN_REGISTRY" > "$PLUGIN_REGISTRY.tmp" && mv "$PLUGIN_REGISTRY.tmp" "$PLUGIN_REGISTRY"
        echo "Cleaned beads from $PLUGIN_REGISTRY"
      fi
    fi
  fi
fi

echo

# --- Per-repo cleanup ---
if $CLEAN_REPO; then
  echo "--- Per-Repo Config (current directory) ---"

  if [[ ! -d .git ]]; then
    echo "Warning: Not in a git repository root, skipping repo cleanup"
  else
    # .beads directory
    remove ".beads"

    # Project-level Claude settings - remove beads hooks
    if [[ -f ".claude/settings.local.json" ]]; then
      if command -v jq &>/dev/null; then
        if grep -q "bd " ".claude/settings.local.json" 2>/dev/null; then
          if $DRY_RUN; then
            echo "[dry-run] Would remove beads hooks from .claude/settings.local.json"
          else
            jq '
              .hooks.PreCompact = ((.hooks.PreCompact // []) | map(select(.hooks | all(.command | test("^bd ") | not)))) |
              .hooks.SessionStart = ((.hooks.SessionStart // []) | map(select(.hooks | all(.command | test("^bd ") | not)))) |
              if .hooks.PreCompact == [] then del(.hooks.PreCompact) else . end |
              if .hooks.SessionStart == [] then del(.hooks.SessionStart) else . end |
              if .hooks == {} then del(.hooks) else . end
            ' ".claude/settings.local.json" > ".claude/settings.local.json.tmp" && mv ".claude/settings.local.json.tmp" ".claude/settings.local.json"
            echo "Cleaned beads hooks from .claude/settings.local.json"
          fi
        fi
      fi
    fi

    # .gitattributes - remove beads merge driver line
    if [[ -f .gitattributes ]]; then
      remove_line ".gitattributes" "merge=beads"
    fi

    # Local git config - remove merge driver
    if git config --local merge.beads.driver &>/dev/null; then
      if $DRY_RUN; then
        echo "[dry-run] Would remove git config merge.beads.driver"
      else
        git config --local --unset merge.beads.driver 2>/dev/null || true
        git config --local --remove-section merge.beads 2>/dev/null || true
        echo "Removed git config merge.beads"
      fi
    fi

    # Git hooks - remove beads hooks or beads lines from hooks
    for hook in .git/hooks/pre-commit .git/hooks/post-merge; do
      if [[ -f "$hook" ]]; then
        if grep -q "bd " "$hook" 2>/dev/null; then
          # Check if hook is purely beads or has other content
          other_lines=$(grep -v "^#" "$hook" | grep -v "bd " | grep -v "^$" | wc -l | tr -d ' ')
          if [[ "$other_lines" -eq 0 ]]; then
            remove "$hook"
          else
            if $DRY_RUN; then
              echo "[dry-run] Would remove 'bd' lines from $hook (hook has other content)"
            else
              remove_line "$hook" "bd "
            fi
          fi
        fi
      fi
      # Remove hook backups created by beads
      remove "$hook.old"
      for backup in "$hook".backup-*; do
        [[ -e "$backup" ]] && remove "$backup"
      done
    done

    # Check AGENTS.md for bd prime references (just warn, don't auto-edit docs)
    for doc in AGENTS.md @AGENTS.md CLAUDE.md .claude/CLAUDE.md; do
      if [[ -f "$doc" ]] && grep -q "bd " "$doc" 2>/dev/null; then
        echo "Note: $doc contains beads references - manual review recommended"
      fi
    done
  fi
else
  echo "Run with --repo to also clean beads from current git repository"
fi

echo
if $DRY_RUN; then
  echo "Dry run complete. Run without --dry-run to apply changes."
else
  echo "Beads uninstall complete."
fi
