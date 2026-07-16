# Ticks upgrades — workshop agenda

This branch (`workshop-prep`) rebases the remaining ticks-skill upgrades from `mkelk/ticks-melk`
onto Peter's **current `main`**. Two of the five proposed pieces already landed; three remain,
and this doc is the agenda for reconciling them.

Full rationale and per-run evidence: [`design/ticks-upgrades-proposal.md`](design/ticks-upgrades-proposal.md)
and [`design/2026-07-05-devmeta-ticks-synthesis.md`](design/2026-07-05-devmeta-ticks-synthesis.md).

## Where we are

| Piece | Status |
|---|---|
| Editorial cleanup | ✅ merged (PR #34) |
| Constraint-surface partitioning | ✅ merged (PR #35) |
| **Config lifecycle hooks** | 🟡 reconciled & applied on this branch — confirm |
| **Runnable worktrees / execution-readiness profile** | 🔴 to reconcile with qfs |
| **Dispatch economics / warm-chains** | 🔴 to reconcile with qfs |

**The base moved.** `main` now carries the **qfs epic (#33)** — the Pi orchestrator plus a
rewritten wave loop (provisional-merge → post-wave gate), a fail-closed acceptance-evidence
model, and a rewritten close-out. Peter noted #33 "was just a draft, can be thrown away."
Whether it stays or is reverted changes each item below — so that's decision zero.

---

## Decision 0 — does qfs (#33) stay on `main`?

- **If it stays:** the three items reconcile *around* qfs (config already does; worktrees/dispatch
  weave into qfs's new loop).
- **If it's reverted:** the loop returns to the simpler pre-qfs form and all three items get
  simpler — config becomes purely additive; worktrees/dispatch layer onto the old loop without
  fighting the post-wave gate.

Settle this first; everything downstream depends on it.

---

## Item 1 — config lifecycle hooks  *(applied here: commit `4740d5ed`)*

**Proposal.** Treat `.tick/config.md` sections as *delivery addresses* wired to lifecycle points;
add three orchestrator-only hook sections — `At wave end`, `At epic close-out`, `At project
checkpoint` — so a project runs its own steps (a release tag, a durable report) at boundaries the
engine already reaches, without forking the skill. Additive; no `tk` change.

**Overlap with current main.** qfs added `Closeout Evidence Commands` + `Acceptance Evidence` to
the *same* config.md section list, and rewrote the loop/close-out the hooks attach to.

**Reconciled (done on this branch).** Unified section list; qfs's fail-closed evidence wording
preserved verbatim; `At wave end` fires *after* qfs's post-wave gate; `At epic close-out` runs as
retro step 7 *after* acceptance-evidence verification; authorization (evidence) vs hooks (side
effects) kept distinct; SKILL.md §5 reframed as an 8-row delivery-address table (Plan B).

**Decision.** (a) Confirm the Plan-B table reframe is acceptable — it rewrites qfs's freshly-merged
§5 (backward-compatible, but not textually additive). (b) If #33 is dropped, swap to the
pure-additive variant (the evidence weaving disappears; the three hooks just get appended).

---

## Item 2 — runnable worktrees / execution-readiness profile  *(not applied)*

**Proposal.** A characterization step at run start writes/maintains `.tick/profile.md`: an inferred
**provisioning recipe** (what makes a fresh worktree buildable/testable — gitignored
`node_modules`/venv/`target` mean a raw worktree can't build — validated by a solo probe before
wave 1), a **tier→venue map** (each test tier runs *in-worktree* only on positive evidence of
isolation, else *post-merge, serial*), and required services. Post-merge verification uses
candidate merges (`git merge --no-commit` → verify → commit/abort) so the integration branch is
never left broken. Conservative by design: "parallel-safe" only on proof, because a wrong
"parallel" corrupts silently while a wrong "serial" only costs time.
*Touches SKILL.md project-files section · agent-runner.md run-start + wave-end loop steps + retro.*

**Overlap with current main.** Largely **complementary** — qfs added no profile, provisioning, or
venue routing. But two collisions:
1. qfs rewrote the wave loop (steps d/e/f) and worktree reconciliation; the profile's wave-end
   venue routing and run-start provisioning must slot into qfs's new step order.
2. qfs's "**provisional merge until the post-wave gate passes**" and the proposal's "**candidate
   merge → verify → commit/abort**" are arguably the *same idea* reached independently. Reconcile
   into one mechanism, not two.

**Decision.** (a) Adopt `.tick/profile.md` + the provisioning probe? (b) Merge qfs's post-wave gate
and the candidate-merge verification into a single model. (c) Where does venue routing
(in-worktree vs post-merge-serial) attach in qfs's loop?

---

## Item 3 — dispatch economics / warm-chains  *(not applied)*

**Proposal.** Three dispatch modes — **solo**, **parallel-wave**, **warm-chain** — chosen per wave
by an economic gate weighing cold-start tax (~20–35 min fresh vs ~4–9 min warm, in the benchmark)
plus provisioning against a warm chain. A warm-chain runs an ordered chain of small related ticks
in one implementer/worktree, committing per tick; disjoint chains still run in parallel worktrees.
Orchestrator still owns all state; skeleton/retro unchanged.
*Touches agent-runner.md dispatch + gate + chain prompt · claude-runner.md (a chain = one agent).*

**Overlap with current main.** Mostly **new** — qfs added a post-wave *verification* gate but no
*dispatch-mode* selection. Two sync points:
1. The merged partitioning (#35) speaks of "constraint groups," but the warm-chain references were
   **stripped** to keep #35 self-contained. Adopting dispatch re-introduces the "group small
   cohesive ticks into a chain" step and its tick-patterns cross-reference.
2. Mode selection reshapes loop step 3b ("launch one implementer per tick"), which qfs left mostly
   intact — needs weaving with the post-wave gate.

**Decision.** (a) Adopt warm-chains + the economic gate, or keep parallel-wave only? (b) Re-link
partitioning ↔ dispatch (the stripped cross-reference). (c) Does the gate's arithmetic live in the
skill or the profile?

---

## Also parked

- **Review-depth ledger** (observability only): a retro line recording what the epic-final review
  ran, time spent, and findings by severity. No behavior change; small — adopt alongside whichever
  items land.
- **Two optional `tk` conveniences** from the proposal (`tk next --json` returning matching
  `At epic close-out` content; nothing required).

---

## Suggested order

1. **Decision 0** — keep or drop qfs (#33).
2. **Item 1 (config)** — smallest, already reconciled; confirm and land.
3. **Item 2 (worktrees)** — the "real bug"; resolve the candidate-merge ↔ post-wave-gate overlap.
4. **Item 3 (dispatch)** — most consequential; depends on Item 2's profile for its cost inputs.
