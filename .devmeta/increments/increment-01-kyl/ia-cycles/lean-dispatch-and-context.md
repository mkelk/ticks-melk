# E2 z5t — Lean dispatch and context: retro

**Closed:** 2026-11-12 · **Diff:** 441b6020..HEAD, `skills/ticks/` (6 files) + Go learnings budgets (`internal/tick/`, `cmd/tk/cmd/`, CHANGELOG) · **PR:** mkelk/ticks-melk#5 (draft, increment-wide)

**One line:** the engine now defaults to 2-3 parallel warm-chains per epic. Implementers are kept warm until the final review and fix its findings as warm turns. The learnings are a 4 KB head plus a looked-up body, and the retro has a steering-sunset step. The epic ran that way itself: three chains, three cold starts for eleven ticks, and seventeen review findings fixed warm.

## What landed
- `agent-runner.md` (chain P):
  - The economic gate makes parallel warm-chains the default: grouped by subsystem and tier, file-disjoint, balanced to the longest tick, 2-3 implementers per epic, any other choice noted. The rate window caps width as a gate input. The static-prefix order is documented, with its cache benefit marked unverified.
  - Implementers are retired after the final review. Findings, including repair ticks, go back as a warm turn; a noted fresh dispatch is the fallback. Cleanup, pooling and resume reconciliation all wait for the final review.
  - The learnings split: head and body formats, budgets, and readers (the planner at partition time, the retro at compaction, implementers through the head's index). Retro step 3 compacts both files and migrates a single over-budget file.
  - The template report has fixed check labels, a fourth outcome (`went red, not from this change`) and the full chain status set.
  - The retro has a new step 6, "Steering sunset" (8 steps now). The retro report records the dispatch-mode and review-depth ledgers, with the orchestrator's context at close beside them. A fresh session per epic is reconciled with auto-continue.
- `tick-patterns.md` and `SKILL.md` (chain S) state the same chain default. The planner reads both learnings files at Step 0.
- `claude-runner.md` and `codex-runner.md` (chain S) name the warm-turn mechanism and the noted fallback. `pi-runner.md` lists the seven rules the extension does not implement yet.
- Go (chain G): `tk status` and `tk close` (epics) warn per file — head over 4096 B, body over 16384 B, body entry over 800 B with its line number. Absent or unreadable files are silent. There are boundary, unreadable-file, hint-text and command-level tests.

## Verification (DoD, outside-in)
`.tick/config.md` `Acceptance Evidence` A1-A6 map to the earlier qfs epic's Pi scripts, not to z5t's items, so they authorize nothing here. z5t is verified by reads, greps and the Testing commands.

| # | Item | Verified | Evidence / gap action |
|---|---|---|---|
| 1 | Parallel warm-chains are the default, 2-3 implementers; AR, tick-patterns Step 4 and SKILL agree | yes | greps hit in all three; review: met after fixes (SKILL.md max_parallel, Step 5) |
| 2 | Retire after the final review; warm turn; adapters name the mechanism and fallback | yes | AR loop 4g/5, Reviewing the work; claude/codex Continuation; review fixes for AR:115, :678 |
| 3 | Rate window as a gate input; static-prefix order, unverified | yes | AR "Dispatch modes and the economic gate"; "same static prefix order" (review nit fixed) |
| 4 | Fresh session per epic; context line in the retro report | yes | AR "What a run actually costs" item 1 and retro step 7; this report's context line |
| 5 | Learnings head and body, budgets, readers, compaction and migration, per-file warnings with tests | yes | AR format section and step 3; TP Step 0; Go tests at the boundaries; live: the repo's `tk status` warned per file on the migrated body, then went silent after two entries were tightened |
| 6 | Steering-sunset step (tally, two silent epics, harness checks never candidates, restore) | yes | AR step 6. Scope fixed in review: demotion candidates are `## Testing` commands; the two template checks are tallied for the record |
| 7 | pi-runner "Current limitations" | yes | 7 "Not implemented" rows, with code evidence |
| 8 | Testing commands pass | **yes, with a stated basis** | Go `-short`: green except `internal/tui` WaitFor goldens, red at base too (A/B). Pi: 152/158 at the merged tree and at base, the same crash-recovery family. Re-check after the review fixes: Go packages ok; `tick-patterns.md` unchanged since the gate. CI on PR #5 is authoritative for Go. Logs: `../.ticks-worktrees/logs/z5t/` |

## Drift
None. No TODO/HACK/`|| true`/`--force` in the diff. The final review's consistency findings (13) were stale wording next to the new rules, and all are fixed.

## Review (frontier, two read-only reviewers; axes earned by the plan)
- Prose (spec compliance and cross-reference consistency, earned by three parallel prose chains): 1 blocker downgraded to should-fix (SKILL.md "close and clean up" vs retire-after-review), 6 should-fix, 5 nits (AR wave gate once per wave, capability table per unit, repair-tick routing, session vs fresh-session wording, sunset scope, resume cleanup; pooling, step 3 heading, "identical" prefix, adapter wording). All fixed.
- Go (correctness and test quality, earned by boundary and absent-file logic): 1 should-fix (no command-level test of the warnings), 3 nits (unreadable file, hint text, contradicting CHANGELOG bullets). All fixed.
- The orchestrator's own lead (AR:196 "per chain") was confirmed and fixed.

## Dispatch-mode ledger
| Unit | Mode | Tier (model) | Reason | Tokens | Wall |
|---|---|---|---|---|---|
| Planning (previous session) | frontier planner, ratified | frontier (fable) | engine rule | — | — |
| Chain P 6na→uy4→nx1→c7q→344→30z | parallel warm-chain | strong (opus) | "edits the engine's own rules; consistency across references" (human-set) | 127k | 4.4 min |
| Chain S vt7→eh5→8bu→q0f | parallel warm-chain | strong (opus) | same | 128k | 10.7 min |
| Chain G zg4 | solo (longest-tick rule) | strong (opus) | same | 78k | 4.6 min |
| Mid-chain message to S (P's wording, three stale lines) | SendMessage to a running chain | — | cross-chain consistency | — | — |
| Review fixes G / S / P | **warm turns to the writers** | strong (opus) | review routes back to the writer | +20k / +10k / +17k | 154 s / 51 s / 105 s |
| Final review | two read-only reviewers in parallel | frontier (fable) | engine rule; two earned axes | 62k + 116k | 1.2 / 2.6 min |

Noted choices: orchestrator-created worktrees, not harness isolation (non-default branch, claude-runner field note); the Go wave gate started before S merged (S was docs-only, and no Go test reads `skills/`); chain P ran longer than the balance target because agent-runner.md is one seam file (planner's note). Cold starts: **3** for 11 ticks plus 17 findings; per-tick dispatch would have been about 14.

## Review-depth ledger
Two reviewers, as the plan's kff tick named. Spec compliance and cross-reference consistency were earned by prose written in parallel by chains that could not see each other. Correctness and test quality were earned by new boundary logic whose tests are the only proof. No security, performance or concurrency surface.

## Orchestrator context
At close: this session's transcript, **2.42 MB, 162 assistant turns** (proxy: the JSONL transcript size, which includes tool output). The session covered z5t only: fresh after `/clear`, with one context compaction mid-run. This is the first measured line; pjz's was unmeasured.

## Steering-sunset tally (first live run)
| Check label | pjz | z5t | Demotion |
|---|---|---|---|
| `## Testing` Go `go test -short` (reached packages) | not applicable (docs only) | **caught**: zg4 went red and changed the code (TDD, plus one wrong fixture) | none |
| Go build / vet | — | ran clean | none (one epic of data) |
| `## Testing` Pi runner suite | not run per tick (wave gate only) | not run per tick | none: never a tick-gate check here |
| `by-path run` (planning.test.ts) | went red, not from this change | went red, not from this change (vt7, eh5) | for the record (template rule) |
| `planted fault` | not applicable | zg4: 5 faults + 4 in the review fix, each turned its test red | for the record (template rule) |
| Tick criteria greps (docs ticks) | ran clean (5) | ran clean (10 + review fixes) | not a `## Testing` command |

**No demotion, no restore.** Two gaps found by running it:
- `.tick/config.md` `## Testing` is not yet grouped by gate (the dmtix two-tier skeleton is deferred), so the tick-gate labels map to it loosely.
- The `planted fault` outcomes read backwards: a planted fault *should* go red, and the implementer reported that as "went red and changed the code", although no code changed. Proposal for skills-fe: define a catch for a planted fault as "stayed green, so the test changed".

## Learnings promoted
- `.tick/learnings.md` → migrated (step 3): head 1,795 B (8 hard rules plus a 6-category index), body `.tick/learnings-ref.md` 10,262 B. Two entries were tightened under 800 B, flagged by the new `tk` warning itself.
- `.tick/profile.md`: the first observed warm/cold ratio, and the z5t flake set (TUI goldens, planning recovery family).
- `skills/ticks/references/claude-runner.md`: field note "three parallel warm-chains kept warm through the review" (measured).

## Proposed roadmap adjustments
None. Two inputs for skills-fe (not roadmap changes): the planted-fault outcome semantics above, and the deferred dmtix `## Testing` skeleton, which the sunset tally now depends on.
