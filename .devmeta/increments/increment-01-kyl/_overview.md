# Increment 01-kyl — Lean engine

**Status:** NOT STARTED · **Target:** —
**Engine:** synthesized ticks skill, user level (`~/.claude/skills/ticks/`). This increment edits the repo source `skills/ticks/`; the installed copy stays untouched during the run (skills-fe owns reinstall after merge).
**Spec:** `docs/thoughts/2026-11-12-lean-epics.md` (mkelk/skills#3)

## Goal
After this increment, a run on the ticks engine pays one cold start per chain rather than per tick. It runs slow and shared checks once per wave rather than per tick. Implementers get a criteria-shaped brief and a short always-read learnings head. And the retro retires per-tick checks that catch nothing, on evidence. The harness checks stay untouched: the wave gate, the frontier final review, the DoDs and the checkpoint.

## Produces
### On screen
- Implementer briefs shaped as criteria, gates, boundaries and report shape, with a per-check line in every report.
- Retro reports with the orchestrator-context ledger line and a steering-sunset tally.
- `tk status` / `tk close` warnings per learnings file (head and body budgets).
### Under the hood
- Engine docs: `skills/ticks/SKILL.md`, `references/agent-runner.md`, `tick-patterns.md`, `claude-runner.md`, `codex-runner.md`, `pi-runner.md` (limitations only).
- Go: `internal/tick/learnings.go` (+ tests), the warning call sites in `cmd/tk/cmd/status.go` and `close.go`.

## Not included
| Deferred | Why |
|---|---|
| dmtix bootstrap two-tier `## Testing` skeleton | Lives in mkelk/skills; skills-fe writes it after E1 lands, from E1's final Testing wording |
| fh E2 numbers as acceptance | Per Morten: not this increment's DoD; the spec's numeric acceptance is settled separately |
| Pi extension code (`extensions/ticks-runner/`) | Executable adapter; only a limitations note here |
| Lowering the final-review tier | Spec "Keep": the frontier review is the harness's best finder |
| Reinstalling/pinning the user-level ticks skill | skills-fe holds it; six fh streams run on the installed copy |

## Roadmap (tk)
Project `s07` — the checkpoint boundary; the run stops there for human review.
| Epic | tk id | Scope | Dep |
|---|---|---|---|
| Lean brief and two-tier verification | pjz | Prompt template + report shape, gate tiers, regenerate-once, planted faults on logic only, by-path rule | — |
| Lean dispatch and context | z5t | Warm-chain default + keep-warm-through-review + rate window + static prefix + orchestrator ledger; learnings split (docs + Go); steering-sunset retro step — one wave of three groups | blocked-by pjz (same prompt template) |

**Dispatch policy (both epics):** implementers at the **strong** tier, dispatch-ledger reason: "edits the engine's own rules; consistency across references". Final review at **frontier**.

**Ratified decisions:** two epics. Learnings layout: head `.tick/learnings.md` (≤4096 B) + body `.tick/learnings-ref.md` (≤16384 B, ≤800 B/entry); engine owns the format, project the content. Sunset demotions applied by the retro and recorded, reversible on evidence.

## Definitions of done
### Lean brief and two-tier verification (pjz)
1. agent-runner.md names two verification tiers. Per tick, in the worktree: typecheck, the tests the change reaches, and the build. Per wave or epic, by the integrator: the full suite, slow suites, specs, and shared generated outputs. The loop (steps 4e-4f), "Post-merge verification", and SKILL.md's `## Testing` row use those names. No file under skills/ticks/ tells an implementer to run the full suite or a slow tier per tick.
2. Shared generated outputs (screenshots, baselines, built assets) are regenerated once by the integrator after the merge, written into the wave-end guidance. The implementer prompt tells implementers not to regenerate them.
3. A planted fault (watch the test fail) is required only where the logic can be wrong in a way the test should catch. tick-patterns.md "Test-Driven Development", the prompt template, and both adapters agree; none mandates one on every new test.
4. The by-path rule is in the engine's testing guidance: a test that reads a source by path escapes changed-file selection, so after editing such a file the implementer runs that package's tests whole.
5. The implementer prompt template (and its warm-chain variant) is shaped as criteria, gates, boundaries, and report shape. The base check, read-inputs, and commit protocol and the Boundaries section stay. The report includes one line per mandated check: ran clean / went red and changed the code / not applicable.
6. claude-runner.md and codex-runner.md agree with the new template; grep finds no adapter describing the old per-tick gate set.
7. The `.tick/config.md` Testing commands pass: Pi runner `node --test` and Go `go test -short`. The Pi tests read tick-patterns.md.

### Lean dispatch and context (z5t)
1. The economic gate makes parallel warm-chains the default. Chains are grouped by subsystem and tier and are file-disjoint. They are balanced so the longest chain is about the longest single tick; a tick that is itself the longest runs solo. Target: 2-3 implementers per epic unless the work is large; any other choice is a noted one. agent-runner.md, tick-patterns.md partitioning Step 4, and SKILL.md "Group small cohesive ticks" agree.
2. Implementers are retired after the final review, not after their tick. The review's should-fix items go back to the chain or implementer that wrote the code as a warm turn. claude-runner.md and codex-runner.md each name the mechanism and the fallback when the warm worker is gone: a fresh dispatch, noted.
3. The rate window caps dispatch width as a gate input. The identical-static-prefix ordering is documented, with the cache-accounting caveat marked unverified.
4. A fresh orchestrator session per epic stays. The retro report records the orchestrator's context next to the dispatch-mode and review-depth ledgers: its size at close, or the best available proxy, named.
5. The learnings split:
   - `.tick/learnings.md` is the always-read head: hard rules plus an index of the body's categories, at most 4096 bytes.
   - `.tick/learnings-ref.md` is the looked-up body: Problem/Cause/Rule entries, at most 16384 bytes, at most 800 bytes per entry.
   - The prompt template reads the head and looks the body up by category.
   - Retro step 3 compacts both files and migrates an over-budget single file.
   - `tk status` and `tk close` warn per file. Go tests cover the head budget, the body budget, the entry limit, and an absent body file (no warning).
6. A steering-sunset retro step:
   - It tallies, from the implementer reports' per-check lines, what each mandated per-tick check caught this epic. "Caught" means it went red and the code changed.
   - A check silent for two consecutive epics is demoted in `.tick/config.md` `## Testing` from every tick to where the logic warrants. The retro report records the demotion with its evidence.
   - Harness checks (wave gate, final review, DoD, checkpoint) are never candidates.
   - Rework at the wave gate or final review in a demoted check's domain restores the check.
7. pi-runner.md "Current limitations" names which of these the Pi extension does not yet implement.
8. The `.tick/config.md` Testing commands pass: Pi runner `node --test` and Go `go test -short`.

## Exit criteria (verification gates)
- [ ] Pi runner: `node --test --no-warnings extensions/ticks-runner/*.test.ts`
- [ ] Go: `go test -short -count=1 ./...` (internal/worktree git-identity flake is a known local baseline)
- [ ] Cross-reference consistency: final frontier review finds no file under `skills/ticks/` contradicting the new tiers, brief, chain default, learnings layout or sunset step
- [ ] Spec coverage: every "Engine, …" row of the spec's Synthesis table is implemented or listed above under Not included

## Human items
- None blocking. Note: unrelated older epics 7ab, qr4, weo show as active in `tk roadmap`; the run scopes to project s07.

## How to run
`/dmtix go` on branch `2026-11-12-lean-engine` (no services needed). Resume: `/dmtix go` again.
