# Project Field for Ticks

Add a first-class `project` field to ticks that allows grouping tasks/epics by project code. This enables filtering, querying, and tracking work across a project.

## Schema Change

Add `project` to the Tick struct:

```go
type Tick struct {
    // ... existing fields ...
    Project string `json:"project,omitempty"` // project code (e.g., "2026-01-14-6453-auth")
}
```

**Properties:**
- Optional string field (empty string = no project)
- No format validation required
- Persisted in tick JSON files
- The recommended format `YYYY-MM-DD-XXXX-name` is convention, not enforced

**Migration:**
- No migration needed for existing ticks
- Missing field = empty project (Go zero value)

## CLI Changes

### `tk create`

Add `-project` flag with parent inheritance:

```bash
tk create "Title" -project "2026-01-14-6453-auth"
tk create "Title" -project "my-project" -parent <epic-id>
```

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | | Project code to assign |

**Inheritance behavior:**
1. If `-project` is specified, use it
2. If `-project` is NOT specified AND `-parent` is specified:
   - Load parent tick
   - If parent has a project, inherit it
   - If parent has no project, task has no project
3. If neither specified, no project

### `tk update`

Add `--project` flag:

```bash
tk update <id> --project "2026-01-14-6453-auth"  # Set project
tk update <id> --project ""                       # Clear project
```

| Flag | Description |
|------|-------------|
| `--project` | Set or clear project code |

### `tk list`

Add `--project` filter:

```bash
tk list --project "2026-01-14-6453-auth"
tk list -t epic --project "..."
tk list -s open --project "..."
```

| Flag | Description |
|------|-------------|
| `--project` | Filter by exact project match |

**Behavior:**
- Only return ticks where `project` matches exactly
- Ticks with no project (empty string) are excluded when filter is set
- Combines with existing filters (type, status, parent, etc.)

### `tk ready`

Add `--project` filter:

```bash
tk ready --project "2026-01-14-6453-auth"
```

**Behavior:**
- Filter ready (unblocked) tasks to only those matching project
- Combines with existing ready logic

### `tk next`

Add `--project` filter:

```bash
tk next --project "2026-01-14-6453-auth"
tk next <epic-id> --project "..."
```

**Behavior:**
- When `--project` specified without epic: find next task across all epics matching project
- When `--project` specified with epic: validate epic matches project, then find next task
- Combines with existing next logic (priority, blockers, etc.)

### `tk show`

Display project in output:

```
ID:          abc123
Title:       Add JWT validation
Type:        task
Status:      open
Priority:    2 (Medium)
Project:     2026-01-14-6453-auth
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

### `tk list --json`

Include project in JSON output:

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

## Files to Change

| File | Change |
|------|--------|
| `internal/tick/tick.go` | Add `Project` field to struct |
| `cmd/create.go` | Add `-project` flag, inheritance logic |
| `cmd/update.go` | Add `--project` flag |
| `cmd/list.go` | Add `--project` filter |
| `cmd/ready.go` | Add `--project` filter |
| `cmd/next.go` | Add `--project` filter |
| `cmd/show.go` | Display project field |

## Test Cases

### Create with project
```bash
tk create "Test task" -project "test-proj-123"
tk show <id>  # Verify: Project: test-proj-123
```

### Create without project
```bash
tk create "No project task"
tk show <id>  # Verify: Project: (empty or not displayed)
```

### Inheritance from parent
```bash
tk create "Parent epic" -t epic -project "epic-proj"
tk create "Child task" -parent <epic-id>
tk show <child-id>  # Verify: Project: epic-proj
```

### Override inherited project
```bash
tk create "Child task" -parent <epic-id> -project "different-proj"
tk show <child-id>  # Verify: Project: different-proj
```

### List filter
```bash
tk create "Task A" -project "proj-a"
tk create "Task B" -project "proj-b"
tk create "Task C"  # no project
tk list --project "proj-a"  # Verify: only Task A returned
```

### Ready filter
```bash
tk ready --project "proj-a"  # Verify: only ready tasks with proj-a
```

### Next filter
```bash
tk next --project "proj-a"  # Verify: returns next task from proj-a only
```

### Update project
```bash
tk update <id> --project "new-proj"
tk show <id>  # Verify: Project: new-proj
```

### Clear project
```bash
tk update <id> --project ""
tk show <id>  # Verify: no project
```
