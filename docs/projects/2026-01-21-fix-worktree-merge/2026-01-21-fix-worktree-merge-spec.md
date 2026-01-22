# Fix Worktree Merge Target

**Created:** 2026-01-21
**Status:** Draft

## Problem

When a worktree is created from a feature branch and later merged, it merges to `main`/`master` instead of the originating branch.

**Steps to reproduce:**
1. Checkout `feature/my-feature`
2. Run `tk run --worktree <epic-id>`
3. Work completes, commits made to `tick/<epic-id>` branch
4. Merge happens → **targets `main` instead of `feature/my-feature`**

## Root Cause

`MergeManager.Merge()` at `internal/worktree/merge.go:55-104` always merges to `m.mainBranch`:

```go
func (m *MergeManager) Merge(wt *Worktree) (*MergeResult, error) {
    if err := m.checkoutMain(); err != nil {  // Always checks out main
        // ...
    }
    // ...
}
```

The `Worktree` struct has no record of which branch it was created from.

## Solution

Two changes:

1. **Record parent branch on worktree creation** - Store in `.tk-metadata` file in the worktree directory
2. **Merge to parent branch** - `MergeManager.Merge()` uses `Worktree.ParentBranch` instead of `mainBranch`

## Changes

### `internal/worktree/worktree.go`

**Add field to struct:**
```go
type Worktree struct {
    Path         string
    Branch       string
    EpicID       string
    Created      time.Time
    ParentBranch string    // NEW: Branch worktree was created from
}
```

**Add metadata file constants and helpers:**
```go
const metadataFileName = ".tk-metadata"

type worktreeMetadata struct {
    ParentBranch string    `json:"parentBranch"`
    CreatedAt    time.Time `json:"createdAt"`
}
```

**In `Create()`:** Before creating the worktree, get current branch. After creating, write metadata file. Set `ParentBranch` on returned struct.

**In `List()`:** After parsing worktrees, read metadata file for each to populate `ParentBranch`.

### `internal/worktree/merge.go`

**In `Merge()`:** Replace `m.checkoutMain()` with checkout of `wt.ParentBranch`. Fall back to `m.mainBranch` if parent is empty or doesn't exist.

```go
func (m *MergeManager) Merge(wt *Worktree) (*MergeResult, error) {
    target := m.mainBranch
    if wt.ParentBranch != "" && m.branchExists(wt.ParentBranch) {
        target = wt.ParentBranch
    }

    if err := m.checkoutBranch(target); err != nil {
        // ...
    }
    // ... rest unchanged
}
```

## Tests

### `internal/worktree/worktree_test.go`

1. `TestManager_Create_RecordsParentBranch` - Worktree created from feature branch has correct `ParentBranch`
2. `TestManager_Create_DetachedHead` - `ParentBranch` is empty when HEAD is detached
3. `TestManager_List_ReadsParentBranch` - `List()` populates `ParentBranch` from metadata

### `internal/worktree/merge_test.go`

4. `TestMergeManager_MergesToParentBranch` - Merge targets parent branch, not main
5. `TestMergeManager_FallbackToMain` - Falls back to main if parent branch deleted

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Parent branch deleted | Fall back to main/master |
| Detached HEAD | `ParentBranch` empty, merge to main |
| Legacy worktree (no metadata) | `ParentBranch` empty, merge to main |
| Metadata file corrupted | `ParentBranch` empty, merge to main |

## Files Modified

| File | Change |
|------|--------|
| `internal/worktree/worktree.go` | Add `ParentBranch` field, metadata read/write |
| `internal/worktree/merge.go` | Use `ParentBranch` as merge target |
| `internal/worktree/worktree_test.go` | Tests for parent branch tracking |
| `internal/worktree/merge_test.go` | Tests for parent branch merge |
