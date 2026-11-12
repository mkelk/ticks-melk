# Lean epics: what a run spends, and where the checks pay

2026-11-12. Observations from the fh project's `look` side stream (increment 11s128-kqr, a design
increment: colour tokens, small UI components, one page, a logo, then a frame). Not authoritative: a
proposal for dmtix and the ticks engine, to settle after one more epic gives numbers.

## Where it came from

Morten, mid-run, after E1: *"can we make this increment efficient in tokens and not wasting time on
overdoing testing"*, then: *"are they something that we should consider for general improvement of
dmtix?"*. From E2 on the run went lean (below). This note records what E1 cost, what the checks found,
and which changes look general.

## What E1 cost

E1 was four read-only scouts for planning, one wave of four implementer ticks, a final review, and two
follow-up ticks from the review. Numbers are the harness's per-subagent totals; minutes rounded down.

| Tick | Model | Subagent tokens | Minutes | Result |
|---|---|---|---|---|
| four planning scouts | (default) | 208,970 | <1 each | maps of the code to touch |
| tokens and colour rule | sonnet | 110,860 | 15 | done, but merged red (see below) |
| UI components | sonnet | 106,536 | 21 | done |
| one page (Kurzia's home) | sonnet | 120,239 | 17 | done, unfiled four flaky pictures to redraw them |
| logo drawing | opus | 188,218 | 18 | done |
| final review | opus | 162,758 | 7 | 0 blockers, 4 should-fix |
| review fixes, page | sonnet | 107,995 | 12 | done |
| review fixes, rules | sonnet | 109,615 | 10 | done |
| **total** | | **~1.11M** (implementers and review ~906k) | | |

Every implementer ran the same per-tick gates: typecheck, the tests its change reaches (`--changed`),
the scripts suite, three pin tests by name, the web build, its own UI specs, a redraw of its pictures,
and a planted fault per new test ("watch it fail"). The orchestrator then ran the full suite per wave.

## What the checks actually found

- **The one real break came from the wave gate, not a tick.** A tick added a product colour block;
  `tokens.test.ts` held the old "one look" rule and failed. The tick's `--changed` run never reached it,
  because that test reads `tokens.css` by path, not by import. Every tick-level gate passed; the full
  suite at the wave caught it.
- **The final review found four real issues** no gate could: a rule table naming the wrong colours, six
  avatar colours doubling as chip colours, a greeting that named readers by their email's local part,
  and held or dropped issues counting as new. Each would have reached readers. Best value in the epic.
- **Pictures were redrawn three times.** Two ticks redrew their own; the orchestrator redrew the shared
  style guide after the merge; the fix ticks redrew again. Only the last redraw counted.
- **The planted-fault ritual** found nothing on word tables and class maps; it earned its cost on the
  hash, the clock edges and the host rules.
- **The fixed cost per implementer is large.** Each one reads the brief, the rules, the learnings file
  (16 KB), the project's CLAUDE.md and the code it touches before writing a line. A small tick (a few
  word changes and two tests) still cost ~108k tokens.

## What the run changed from E2 on

1. 2-3 ticks per epic, split by files that do not overlap, not one tick per deliverable.
2. A tick runs typecheck, the tests its change reaches, and the build. No UI specs, no redraws, no pin
   or docs-number runs, planted faults only on real logic.
3. The orchestrator runs the full suite, the scripts suite, diagrams and the changed pages' specs once
   at the epic gate, and redraws the pictures once (or has one cheap agent redraw and flag).
4. Shorter briefs: the task, the gates, the boundaries, the report shape. No look pages unless a human
   must decide.
5. The final review drops to Sonnet when the diff is Sonnet's and the logic is plain; Opus only for
   tricky logic. This note argues against that change (see "Keep").

First E2 data point: a sonnet tick on a lean brief (proxy rewrites, built icons, a share card, a header,
tests) cost 105,755 tokens in 13 minutes, about the same as E1's sonnet ticks. The saving is not in the
tick's own run; it is in fewer ticks, gates run once, and redraws done once. E2's totals come at its close.

## What looks general (proposal)

| Change | Where it would live | Why general |
|---|---|---|
| Two gate tiers: per tick (typecheck, reached tests, build) and per wave or epic (full suite and everything slow) | ticks engine, `references/agent-runner.md` (per-tick verification vs post-wave gate) | The wave gate is where cross-tick breaks show; repeating slow gates per tick buys little. |
| Size *dispatches* against the fixed context cost, not ticks: keep ticks small, run same-subsystem ticks as a warm-chain, aim for 2-3 implementers per epic unless the work is large | ticks engine, the economic gate's warm-chain default for this shape of work | The cold start is per implementer, not per tick; chains pay it once and keep per-tick resume (see "Second look"). |
| Shared generated files (screenshots, baselines, built assets) are regenerated once, by the integrator, after the merge | ticks engine, "At wave end" guidance | Ticks regenerating shared outputs duplicate work and collide. |
| Planted-fault ("watch it fail") only where the logic can be wrong in a way a test should catch | ticks engine, implementer prompt | It costs a run per test; on data tables it proves nothing. |
| A test that reads a source by path escapes `--changed`; run a package's tests whole after editing a file that tests read by path | project learnings, and a line in the engine's testing guidance | Cheap rule; it was the only break in E1. |
| The project's own gate commands and redraw procedure | the project's `.tick/config.md` `## Testing` | Project-specific; dmtix bootstrap could offer a two-tier Testing skeleton without commands. |

dmtix itself stays a door ("doors, not engines"): the only dmtix-side change would be the bootstrap
template offering an optional two-tier `## Testing` skeleton, with no commands in it.

## Keep

- **The final review at the top model.** E1's review was the cheapest finder of reader-facing bugs. Do
  not lower its tier to save tokens; scope it to the epic's diff.
- **The human gate on taste** (the logo, the colour rules): one look page, through the coordinator.

## Second look (same day): fewer cold starts, not fewer ticks

Morten's question, later the same day: dmtix's point is that the ticks hold the state, so compaction
or a stop and restart is never dangerous. Are fatter epics in conflict with that?

Only if "fatter" means fewer ticks. The property rests on the **orchestrator** resuming from `tk next`,
where every ready item maps to one action. The implementer is a fresh subagent either way and has no
memory to lose. So the tick is the unit of orchestrator state and of resume granularity; it was never
the unit of implementer context. E1's cost finding is about cold starts, one per implementer (brief,
rules, learnings, CLAUDE.md, the code: ~100k before the first line). The engine names this, the economic
gate's "cold-start tax", and already has the construct for it: the **warm-chain**. One implementer,
one warm context, an ordered chain of small ticks, a commit per tick, "a dead chain resumes from its
last committed tick". That is the unification:

- ticks stay small: resume granularity, audit trail and `tk next` routing are untouched;
- implementers get few: the cold start is paid once per chain;
- the gates split by themselves: in-worktree tiers per tick, shared tiers "post-merge per chain, not
  per tick" (already the engine's venue rule), redraws once per chain.

So the proposal reads "2-3 implementers per epic", not "2-3 ticks per epic". A fat tick with no
intermediate commit resumes from zero; a chain of small ticks resumes from the last one that landed.
Same token bill, better durability. The table above is corrected accordingly.

**The table in "What E1 cost" is missing the biggest line.** It counts subagent tokens only. The
engine's own cost ordering (`agent-runner.md`, "What a run actually costs") puts the orchestrator's
context first: it lives for the whole epic and grows with the square of session length. E2's
accounting must include it, and the lever there is a fresh session per epic, which is the opposite of
"fatter".

## Harness vs steering

Two further observations from Morten. Models get stronger and need less steering; dmtix is three to six
months old and should take that into account. And the best practitioners agree that predictability
comes from model *plus* harness; Morten's own experience is that dmtix is what gives predictable results.
Both hold, once harness and steering are told apart.

What the harness supplies that no model supplies for itself, however strong:

| Kind | In dmtix | Ages how |
|---|---|---|
| **Memory it does not have** | ticks, worktrees, records tree, `tk next` | Not at all. Orchestrator-only, zero per-agent tax. |
| **A definition of done it did not write** | ratified DoDs, the full suite at the wave gate, the frontier review, the checkpoint | Not at all. The model checked against something outside itself. E1: the one break and all four real bugs were caught here. |
| **Human intent, held constant** | epic partition, exclusions, taste gates, the boundaries in the brief | Grows in value. Stronger models over-deliver; scope steering stays. |

Beside these a fourth thing crept in: **the model re-verifying itself at every step**. Planted faults on
word tables, the full gate set on every tick, three redraws. Those are the model doing its own work twice.
They passed every time and caught nothing, so they neither added predictability nor could have. That is
the part that ages with model strength, and it is where E1's tokens went.

The test for any ritual in `.tick/config.md` or the implementer prompt: *is this the harness checking
the model, or the model checking itself?* The first stays constant as models change. The second shrinks.
Shrinking the second is not thinning the first.

One structural gap: dmtix only accretes steering. The retro harvests learnings into the file; nothing
ever retires a ritual, and the 150-line cap is the lone pruning mechanism. Proposed retro step: every
mandated check reports what it caught this epic; nothing for two epics demotes it from "every tick" to
"where the logic warrants" (what this run did by hand for planted faults). That makes "steer less as
models improve" a cadence with evidence rather than a mid-run judgment call.

## Cross-check: Google's "The New SDLC With Vibe Coding" (Osmani, Saboo, Kartakis; May 2026)

Read in full on the same day. It does not change the analysis; it sharpens it in three places and
disagrees in one.

- **Agent = Model + Harness is its central equation**, and its inventory of a harness is useful: rule
  files and context, tools, sandboxes, orchestration logic, *hooks* ("deterministic code that runs at
  specific lifecycle points... the place for things the agent should never forget but often does"),
  observability. The model re-verifying its own work is not on the list. Verification is tests for the
  deterministic parts and evals / LM judges for the trajectory, run by the harness. That is the line
  above in the paper's vocabulary: the wave-gate suite is a hook; a planted fault run by the
  implementer is not.
- **Static vs dynamic context.** "Static context is expensive because every token is present in every
  interaction, regardless of relevance... The best systems treat this boundary as a first-class
  architectural decision, reviewed and versioned like any other configuration." Progressive disclosure
  (metadata, then instructions on match, then deep reference on demand). This moves the learnings-file
  split from "open" to the recommended default; the engine's config sections already work this way.
- **"Success comes from giving agents success criteria rather than step-by-step instructions."** That is
  the lean brief and the DoD. It is the precise reading of "steer less" the paper endorses, while it
  calls the model-is-the-system intuition "wrong, and it leads to the wrong investments".
- **Disagreement: model routing.** The paper routes "Test Generation, Code Review, and CI/CD monitoring"
  to cheaper models. E1 says the opposite for the epic review. The paper's "code review" is the
  first-pass lint review; ours is the LM-judge eval of the semantic outcome, which its own 80%-problem
  section says is where errors "look right and may even pass basic tests". Keep the frontier review.
- **Silent** on checkpoint granularity (thin vs fat ticks, warm-chains) and on the orchestrator's own
  context as a cost; those remain this note's observations.
- Its statistics are weak (an independent critique traces "41% of new code", "25-39% gains" and
  "3-10x cost" to misreadings or no source). Use the framework, do not cite the numbers.

## Synthesis: the next iteration of dmtix

One principle. dmtix v1 was built for a model that needed *telling*. The next iteration assumes a model
that needs *criteria* and a harness that *checks*. The harness is the fixed part: external state,
external checks, human intent. Steering is the variable part: the smallest static context that still
carries the criteria, pruned on evidence. Same method, same records, same walk-away property; less of
the model doing its own work twice.

The changes, by where they live. Most are engine changes; the door does not move.

| Where | Change | Status |
|---|---|---|
| Engine, dispatch | Parallel warm-chains are the default: K chains in K worktrees, grouped by subsystem and tier, balanced so the longest chain ≈ the longest single tick, implementers kept warm through the review. Ticks stay small (resume granularity); implementers get few (cold-start count). Target: 2-3 implementers per epic. See "Tokens and wall time". | Decision needed |
| Engine, context | Learnings split: a short always-read head (hard rules) and a looked-up body. Progressive disclosure. The one change that cuts every cold start. | Decision needed |
| Engine, retro | Steering-sunset step: every mandated check reports what it caught; silent for two epics demotes it from "every tick" to "where the logic warrants". The only new process step. | Decision needed |
| Engine, verification | Two tiers in the `## Testing` shape: in-worktree per tick (typecheck, reached tests, build); shared per wave (full suite, slow suites, specs, shared artifacts regenerated once at integration). Planted faults only on logic. The by-path test rule. | Done by hand in fh; write it down |
| Engine, the brief | Criteria, gates, boundaries, report shape. No step lists. Boundaries stay: stronger models over-deliver. | Done by hand in fh; write it down |
| Engine, orchestrator | Fresh session per epic; orchestrator context in the ledger. | Free |
| dmtix | Bootstrap offers the two-tier `## Testing` skeleton, no commands. Everything else unchanged: human-ratified partition, DoDs, frontier review, taste gates, checkpoint. | Small |

Run it as an increment of its own, with a DoD, and let E2's numbers be the acceptance test:

- **Efficiency:** tokens per delivered DoD item down against E1, orchestrator context counted.
- **Predictability, the invariant:** rework found at the wave gate and the final review does *not*
  rise. If it rises, the pruning went too far and the sunset step runs in reverse.

### What it would gain, honestly

From E1's numbers, not a forecast. Six implementer cold starts (four wave ticks, two fix ticks) at
roughly 100k each before the first productive line. Chains of 2-3 would cut three or four of those:
~200-300k, a quarter to a third of the implementer spend. Gates and redraws run once instead of per
tick: the three redraws become one; the per-tick gate runs that found nothing go. The learnings split
saves little in tokens (~4k per read) and more in signal; its value is that the head stays short as the
body grows. The orchestrator's context is unmeasured and, by the engine's own ordering, the largest
line; a fresh session per epic is the one change that might save more than all the others, and the
only one with no data behind it yet.

What it does not gain. A tick does not get cheaper: E2's first lean tick cost the same as E1's. The
review's *finding* cost stays; that is quality work, not waste. What the fix costs after the finding
is a dispatch question (next section). Wall time does not improve either, but it need not worsen if
the chains are cut against the critical path.

### Tokens and wall time: the forces, and a balance

A single chain of everything serializes what a wave ran in parallel (E1's wave of four finished in 21
minutes; one chain of four would take an hour). Wall time only matters while a human waits, and dmtix
is walk-away by design, so tokens bind harder than minutes; but minutes count too. The forces:

| Force | Pushes toward | Note |
|---|---|---|
| Cold start (~100k per fresh implementer) | fewer implementers | Two parts: a *generic* part (learnings, CLAUDE.md, rules, brief template), the same for every tick, and a *task-specific* part (the code it touches). Chaining same-subsystem ticks saves both; chaining unrelated ticks saves only the generic part. |
| Parallelism | more implementers | E1's wave finished in the time of its longest tick. |
| Serial gate per merge unit | fewer merge units | Engine: parallel-wave wall ≈ max(tick) + width × gate time. A chain merges as one unit. |
| Orchestrator turns | fewer units | Every dispatch, merge, gate and report is a turn, and turns compound in its context. |
| Rework | keep implementers warm | E1's two fix ticks cost ~218k, mostly cold start, for small changes to code the original implementers had just read. |
| Tier homogeneity | breaks chains | One chain is one model; the opus logo tick cannot chain with sonnet ticks. |
| Rate window (subscription) | caps width | A burst of four can hit the window and stall everything; width can *lengthen* wall time. |
| Resume granularity | neutral | Chains commit per tick; holds either way. |

The balance, four rules:

1. **Group by subsystem and tier, run the groups wide.** K chains in K worktrees, K = the number of
   cohesive, file-disjoint, same-tier groups. The engine names this ("parallel warm-chains"); make it
   the default rather than something the gate has to be argued into.
2. **Balance chains against the critical path.** Chain the short ticks; a tick that is itself the
   longest runs solo. Target: longest chain ≈ longest single tick. Wall time stays at the wave's;
   cold starts drop to K.
3. **Retire implementers after the review, not after the tick.** An idle subagent costs nothing. The
   review's should-fix items go back to the chain that wrote the code, as a warm turn.
4. **Width capped by the rate window**, as a gate input, not a surprise.

Two levers that make cold starts cheaper and so shift the balance back toward width: an identical
static prefix across implementers (system prompt, CLAUDE.md, learnings head, rules, in that order) so
the generic part caches across agents launched together (how the subscription's rate accounting
treats cache reads is unverified; check before counting on it); and the learnings split, which
shrinks the generic part outright.

E1 replayed under these rules: wave ticks of 15, 21, 17 and 18 minutes (tokens, components, page,
logo). Tokens and components share a subsystem; the logo is opus. So one sonnet chain [tokens →
components], page solo, logo solo. With the engine's own warm/cold estimate (a warm tick in ⅓–½ the
time), the chain is ~15 + 8 ≈ 23 min. Wall ~23 vs 21; cold starts 3 vs 4 for the wave, then 0 vs 2
for the fixes: six becomes three; merge units 3 vs 4. The rule in one line: *chain only what would
not lengthen the critical path; keep everyone warm until the reviewer is done.*

The gain that is not a number: v1 only accretes. Every retro adds a rule; nothing removes one; in a
year the static context is all rules and the briefs are all steps. The sunset step is what makes the
method stay lean as models change, instead of a cut made once and then grown back.

## Open

- E2's totals vs E1's (tokens per delivered DoD item, minutes, rework found at the gate and the review),
  **including the orchestrator's own context**. Settle the proposal after that.
- Split the learnings file (read in full by every implementer, 16 KB cap) into a short always-read part
  and a looked-up part; the whitepaper's static/dynamic rule makes this the default, not a question. What
  remains open is the cut and whether the engine or the project owns it.
- Whether the warm-chain should be the economic gate's default for design-increment ticks (small,
  same-subsystem, shared context), rather than something the gate has to be argued into per wave.
- The steering-sunset retro step above: where it lives (engine retro, or a dmtix close-out hook) and
  what counts as "caught something".
