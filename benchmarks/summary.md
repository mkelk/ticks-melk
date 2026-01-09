# Benchmark Summary

Dataset size: 1000 items
Cold runs: 1
Warm runs per op: 20

## Method Details

### Tick
- Repo: temp git repo with fake origin.
- Init: `tk init` (TICK_OWNER=bench).
- Data: 1000 ticks. First 20 seeded with labels/notes for filtering; rest minimal.
- Blockers: every 10th tick blocked by the first tick via `tk block`.
- Ops measured:
  - `list_open`: `tk list --json`
  - `list_label`: `tk list --label bench-hit --all --json`
  - `ready`: `tk ready --json` (default limit=10)
  - `create`: `tk create "Bench create"`
  - `update`: `tk update <id> --status in_progress`
  - `note`: `tk note <id> "Bench note"`

### Beads
- Repo: temp git repo with fake origin.
- Init: `bd init` (BD_ACTOR=bench).
- Data: 1000 issues. First 20 seeded with labels/notes for filtering; rest minimal.
- Blockers: skipped (bd dep unstable; see script).
- Ops measured:
  - `list_open`: `bd list`
  - `list_label`: `bd list --label bench-hit`
  - `ready`: `bd ready`
  - `create`: `bd create "Bench create"`
  - `update`: `bd update <id> --status in_progress`
  - `note`: `bd update <id> --notes "Bench note"`

## Tick (ms)
- list_open: cold 78.60 / warm median 35.53 / warm p95 41.81
- list_label: cold 34.52 / warm median 34.04 / warm p95 35.25
- ready: cold 34.53 / warm median 34.24 / warm p95 38.87
- create: cold 14.83 / warm median 14.97 / warm p95 15.67
- update: cold 24.87 / warm median 25.07 / warm p95 27.11
- note: cold 26.29 / warm median 25.53 / warm p95 31.39

## Beads (ms)
- list_open: cold 76.11 / warm median 68.48 / warm p95 72.51
- list_label: cold 61.68 / warm median 67.43 / warm p95 132.24
- ready: cold 159.49 / warm median 66.59 / warm p95 222.09
- create: cold 86.45 / warm median 91.45 / warm p95 118.03
- update: cold 75.66 / warm median 67.65 / warm p95 84.50
- note: cold 66.92 / warm median 63.82 / warm p95 67.43

