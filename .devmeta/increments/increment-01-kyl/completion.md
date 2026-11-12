# Increment 01-kyl — Lean engine: completion report

**Checkpoint:** 2026-11-12 · **Branch:** `2026-11-12-lean-engine` (integration stops here; not merged to main/master) · **PR:** mkelk/ticks-melk#5 (draft) · **Project tick:** `s07` (open, for the human to close)

## Features
After this increment, a run on the ticks engine:
- **pays one cold start per chain, not per tick.** Parallel warm-chains are the default (2-3 implementers per epic: file-disjoint, balanced against the longest tick, width capped by the rate window), and implementers stay warm through the final review to fix its findings (E2 z5t).
- **runs slow and shared checks once per wave.** The tick gate (typecheck, reached tests, build) runs in the worktree. The wave gate (full suite, slow suites, specs, generated outputs regenerated once) is the integrator's (E1 pjz).
- **briefs implementers by criteria.** Criteria, gates, boundaries and report, with a per-check line in four fixed outcomes and the full chain status set (pjz + z5t).
- **keeps the always-read learnings short.** A head of at most 4 KB plus a body looked up by category, with per-file `tk status`/`tk close` warnings (z5t). This repo is migrated: head 1.8 KB, body 10.3 KB.
- **retires per-tick checks on evidence.** The steering-sunset retro step; harness checks are never candidates (z5t).
- The harness checks are untouched: the wave gate, the frontier final review, the DoDs and the checkpoint.

| Epic | Retro |
|---|---|
| E1 pjz — Lean brief and two-tier verification | `ia-cycles/lean-brief-and-two-tier-verification.md` |
| E2 z5t — Lean dispatch and context | `ia-cycles/lean-dispatch-and-context.md` |

## Gates status
- [x] Go `go test -short -count=1 ./...`: CI on PR #5 (see below). Local runs are green except the `internal/tui` WaitFor goldens, which are red at base too (load flake, A/B).
- [x] Pi runner `node --test …`: 152/158 locally on both the merged tree and base, the same crash-recovery family; parity by A/B. CI does not run the Pi suite.
- [x] Cross-reference consistency: two frontier final reviews (one per epic). E2's 13 prose findings are all fixed; the re-check greps are clean.
- [x] Spec coverage: every "Engine, …" row of the spec's Synthesis table is implemented here or listed under the overview's "Not included".
- [x] CI on PR #5 at the close-out commit `1dad2e3b`: green (go, ui, worker), run 43225736911.

## Outstanding human items
- Review the increment and smoke-test, then close project tick `s07`. Decide whether and when to merge PR #5 (shipping is the human's decision).
- skills-fe reinstalls the user-level ticks skill after the merge (the six fh streams are pinned on the installed copy).
- skills-fe acked the checkpoint and ratified the sunset scope (demote only `## Testing` tick-gate commands; `by-path run`/`planted fault` tallied for the record). It owns both follow-ups: the two-gate `## Testing` skeleton (mkelk/skills#4, dmtix bootstrap), and the engine-text gap on `planted fault` outcomes (its catch is "went red"; "stayed green" means the test changed).
- Unrelated older epics 7ab, qr4 and weo still show as active in `tk roadmap`.

## Postmortems
- **Host load (20-54) all run.** Both epics' local suites were red at base with an unstable failing set. Every gate was settled by base/HEAD A/B and recorded; CI is the authority for Go. Cost: about four extra suite runs per epic.
- **Spec ambiguity resolved without skills-fe.** The sunset scope for `by-path run`/`planted fault` was decided from spec line 144 and reported at close-out, not asked mid-run.
- **The review found stale neighbours, not wrong rules.** Most of E2's 17 findings were older sentences next to the new rules that the parallel chains could not see. Two parallel prose chains need a cross-chain consistency review; it earned its axis.
