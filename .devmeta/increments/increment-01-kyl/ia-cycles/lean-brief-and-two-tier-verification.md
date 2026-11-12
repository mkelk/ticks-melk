# E1 pjz — Lean brief and two-tier verification: retro

**Closed:** 2026-11-12 · **Diff:** 403d086a..af06c92a, `skills/ticks/` (6 files, +72/−51) · **PR:** mkelk/ticks-melk#5 (draft, increment-wide)

**One line:** the engine now names a tick gate and a wave gate. It briefs implementers by criteria, gates, boundaries and report. Each report has a three-outcome line per mandated check. One warm-chain landed it, and the review's six findings were fixed as a warm turn.

## What landed
- `agent-runner.md`:
  - "The tick gate and the wave gate" is defined once, with its relation to venues and the by-path rule. Regenerate-once is in loop 4f, with a K-chains clause.
  - The names are used across the loop, post-merge verification, the status table, integration, conflicts and chains.
  - The implementer template is reshaped as Criteria / Gates / Boundaries / Report, with per-check lines `<check>: ran clean | went red and changed the code | not applicable`. The warm-chain variant matches.
  - The template states the tick-gate commands for an ungrouped `Testing` (review fix).
- `tick-patterns.md`: the TDD section requires a planted fault only where the logic can be wrong; "existing tests pass" means the reached tests.
- `SKILL.md`: the Testing row names both gates. The recommended two-group `## Testing` shape is in prose, with no new syntax. The Quick Reference now closes ticks only after the wave gate (a pre-existing slip fixed in passing).
- Adapters point to the shared template. `pi-runner.md` notes that the Pi package does not split the gates yet.

## Verification (DoD, outside-in)
| # | Item | Verified | Evidence / gap action |
|---|---|---|---|
| 1 | Two gates named, used; no per-tick full suite | yes | frontier review: met; sweep grep clean outside field notes |
| 2 | Regenerate once by integrator; implementers told not to | yes | loop 4f, "Verify the wave", template Boundaries |
| 3 | Planted fault conditional; TDD, template, adapters agree | yes | review: met (tick-patterns, template, codex, claude) |
| 4 | By-path rule in testing guidance | yes | tick-gate definition and template Gates |
| 5 | Template shape and per-check line | yes | review: met; three literal outcomes |
| 6 | Adapters agree, no old gate set | yes | review: met; `claude-runner.md:66` renamed in the fix |
| 7 | Testing commands pass | **yes, with a stated basis** | Go: CI on PR #5 (vet/build/test). Pi: CI does not run it. Local runs are red on the baseline too under host load (~40–54 on 16 cores). An A/B of the failing set (base / HEAD / base) failed 4 / 1 / 2, and the TUI golden flipped both ways on both commits: parity, nothing attributable to the diff. Logs: `../.ticks-worktrees/logs/pjz/`. |

## Review (frontier, one reviewer, axes: spec compliance and doc consistency, earned by a docs-only diff)
- 0 blockers, 2 should-fix, 4 nits; all fixed in 7ee8b156.
- Should-fix: the loop's step 5 was fused into 4h (an edit slip in the chain); the tick-gate command source was ambiguous for an ungrouped `Testing` (it read as "run all of Testing", the full suite per tick).
- Nits: old gate names in six places; the "two parts" count; K parallel chains; the Quick Reference close order.

## Dispatch-mode ledger
| Unit | Mode | Tier (model) | Reason | Tokens | Wall |
|---|---|---|---|---|---|
| Planning | one planner, no scouts (5 files, small surface) | frontier (fable) | engine rule: plan at frontier | 70k | 2 min |
| chain1 ada→tln→lhr→23c→1dw | warm-chain, 1 implementer | strong (opus) | "edits the engine's own rules; consistency across references" (human-set) | 160k | 16 min |
| Review fixes | **warm turn to the same implementer** | strong (opus) | review routes back to the writer | **7.5k** | 37 s |
| Final review | read-only, overlapped with the wave gate (noted) | frontier (fable) | engine rule | 69k | 2 min |

Implementer cold starts this epic: **1**. A fresh fix tick would have cost a second cold start; the warm turn cost 7.5k tokens. That is the first data point for z5t's "retire implementers after the review".

## Review-depth ledger
One reviewer. Spec compliance, plus doc accuracy and cross-reference consistency. No other axis was earned (no code, no security surface).

## Orchestrator context
Not measured. This session also carried bootstrap, start and planning. z5t adds the ledger line; by the engine's own advice, E2 starts in a fresh session.

## Check outcomes (a first look at the sunset tally)
- Tick-gate grep checks: ran clean for all five ticks.
- The Pi suite on tln (by-path run): went red, **code not changed**, red on the baseline too. The three outcomes have no slot for this; it is the reviewer's verdict (i).

## Learnings promoted
- `.tick/learnings.md` (repo): the Pi suite needs mise Go on PATH (`planning.test.ts` runs `go build ./cmd/tk`). 10,404 bytes, under budget.
- `skills/ticks/references/agent-runner.md` (runner-neutral): field note "a red wave gate on a loaded host, settled by A/B".
- `.tick/profile.md`: runtime tripwire for host-load flakes; the Go recipe through mise.
- Orchestrator slip, already covered by the existing claude-runner field note on `tk create -d`: an unquoted heredoc ran backtick spans while I wrote tick descriptions. Caught by a grep before `tk create`; two spans restored.

## Drift
None: no workaround flags or TODO/HACK in the diff. The one self-inflicted edit slip (the fused line) was caught by the review.

## Inputs for z5t planning (within its ratified scope; not roadmap changes)
1. The sunset tally needs fixed check labels (the command verbatim, plus the literals `by-path run` and `planted fault`) and a fourth outcome, "went red, not from this change" (flaky or pre-existing). This epic hit exactly that case.
2. Checks a project mandates through `## Rules` are outside the per-check list today; decide whether the sunset step counts them.
3. Once warm-chains are the default, the chain's per-tick status line needs DONE_WITH_CONCERNS (and NEEDS_CONTEXT), or those statuses drop out of most runs.
4. skills-fe's two points (already in z5t's DoD): the Go budget tick has its own file set, and the engine text names the body's readers.

## Proposed roadmap adjustments
None.
