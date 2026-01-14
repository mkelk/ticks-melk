# Instruction: Add Project Field to Ticks CLI

## Goal

Add a first-class `project` field to ticks that allows grouping tasks/epics by project code. This enables filtering, querying, and tracking work across a project.

## Requirements

### 1. Schema Change

Add `project` field to the tick struct/schema:

```go
type Tick struct {
    ID          string
    Title       string
    Description string
    Type        string // "task" or "epic"
    Status      string
    Priority    int
    Labels      []string
    Parent      string
    BlockedBy   []string
    Manual      bool
    Project     string  // NEW: project code (e.g., "2026-01-14-6453-auth")
    // ... other fields
}
```

**Properties:**
- Optional string field (empty string = no project)
- No format validation required
- Persisted in tick JSON files

### 2. Command: `tk create`

Add `-project` flag:

```bash
tk create "Title" -project "2026-01-14-6453-auth"
tk create "Title" -project "my-project" -parent <epic-id>
```

**Inheritance behavior:**
- If `-project` is specified, use it
- If `-project` is NOT specified AND `-parent` is specified:
  - Load parent tick
  - If parent has a project, inherit it
  - If parent has no project, task has no project
- If neither specified, no project

### 3. Command: `tk update`

Add `--project` flag:

```bash
tk update <id> --project "2026-01-14-6453-auth"  # Set project
tk update <id> --project ""                       # Clear project
```

### 4. Command: `tk list`

Add `--project` filter flag:

```bash
tk list --project "2026-01-14-6453-auth"
tk list -t epic --project "..."
tk list -s open --project "..."
```

**Behavior:**
- Only return ticks where `project` matches exactly
- Ticks with no project (empty string) are excluded when filter is set
- Filter combines with existing filters (type, status, parent, etc.)

### 5. Command: `tk ready`

Add `--project` filter flag:

```bash
tk ready --project "2026-01-14-6453-auth"
```

**Behavior:**
- Filter ready (unblocked) tasks to only those matching project
- Combines with existing ready logic

### 6. Command: `tk next`

Add `--project` filter flag:

```bash
tk next --project "2026-01-14-6453-auth"
tk next <epic-id> --project "..."
```

**Behavior:**
- When `--project` specified without epic: find next task across all epics matching project
- When `--project` specified with epic: validate epic matches project, then find next task
- Combines with existing next logic (priority, blockers, etc.)

### 7. Command: `tk show`

Display project in output:

```
ID:          abc123
Title:       Add JWT validation
Type:        task
Status:      open
Priority:    2 (Medium)
Project:     2026-01-14-6453-auth    # NEW LINE
Parent:      def456
Description: Implement JWT signing...
```

**JSON output** (`tk show <id> --json`):
```json
{
  "id": "abc123",
  "title": "Add JWT validation",
  "project": "2026-01-14-6453-auth",
  ...
}
```

### 8. JSON Output for `tk list`

Include project in `--json` output:

```json
{
  "ticks": [
    {
      "id": "abc123",
      "title": "Add JWT",
      "project": "2026-01-14-6453-auth",
      ...
    }
  ]
}
```

## Test Cases

```bash
# 1. Create with project
tk create "Test task" -project "test-proj-123"
# Verify: tk show <id> shows Project: test-proj-123

# 2. Create without project
tk create "No project task"
# Verify: tk show <id> shows Project: (empty or not displayed)

# 3. Inheritance from parent
tk create "Parent epic" -t epic -project "epic-proj"
tk create "Child task" -parent <epic-id>
# Verify: child task has Project: epic-proj

# 4. Override inherited project
tk create "Child task" -parent <epic-id> -project "different-proj"
# Verify: child task has Project: different-proj

# 5. List filter
tk create "Task A" -project "proj-a"
tk create "Task B" -project "proj-b"
tk create "Task C"  # no project
tk list --project "proj-a"
# Verify: only Task A returned

# 6. Ready filter
tk ready --project "proj-a"
# Verify: only ready tasks with proj-a returned

# 7. Next filter
tk next --project "proj-a"
# Verify: returns next task from proj-a only

# 8. Update project
tk update <id> --project "new-proj"
# Verify: tk show <id> shows Project: new-proj

# 9. Clear project
tk update <id> --project ""
# Verify: tk show <id> shows no project
```

## Files Likely to Change

Based on typical Go CLI structure:

- `internal/tick/tick.go` - Add Project field to struct
- `internal/storage/...` - Ensure project persisted/loaded
- `cmd/create.go` - Add -project flag, inheritance logic
- `cmd/update.go` - Add --project flag
- `cmd/list.go` - Add --project filter
- `cmd/ready.go` - Add --project filter
- `cmd/next.go` - Add --project filter
- `cmd/show.go` - Display project field

## Notes

- No migration needed for existing ticks - missing field = empty project
- Project is just a string, no special validation
- The recommended format `YYYY-MM-DD-XXXX-name` is a convention, not enforced
