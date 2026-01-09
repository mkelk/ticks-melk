# Benchmark Summary

Dataset size: 100 items
Runs per op: 20

## Method Details

### Tick
- Repo: temp git repo with fake origin.
- Init: `tk init` (TICK_OWNER=bench).
- Data: 100 ticks created via `tk create "Tick N" --json`.
- Blockers: every 10th tick blocked by the first tick via `tk block`.
- Ops measured:
  - `list_open`: `tk list --json`
  - `ready`: `tk ready --json` (default limit=10)
  - `create`: `tk create "Bench create"`
  - `update`: `tk update <id> --status in_progress`
  - `note`: `tk note <id> "Bench note"`

### Beads
- Repo: temp git repo with fake origin.
- Init: `bd init` (BD_ACTOR=bench).
- Data: 100 issues via `bd create "Issue N" --description "Benchmark issue" --json`.
- Blockers: skipped (bd dep unstable; see script).
- Ops measured:
  - `list_open`: `bd list`
  - `ready`: `bd ready`
  - `create`: `bd create "Bench create"`
  - `update`: `bd update <id> --status in_progress`
  - `note`: `bd update <id> --notes "Bench note"`

## Tick (ms)
- list_open: median 16.39 / p95 36.95
- ready: median 16.40 / p95 155.41
- create: median 14.15 / p95 16.42
- update: median 24.12 / p95 39.93
- note: median 24.31 / p95 46.07

## Beads (ms)
- list_open: median 59.00 / p95 78.63
- ready: median 54.82 / p95 68.55
- create: median 74.97 / p95 78.98
- update: median 55.61 / p95 57.10
- note: median 55.13 / p95 74.51
