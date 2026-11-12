# Claude Code Runner Adapter

Read [`agent-runner.md`](agent-runner.md) first. This file maps its capability contract onto Claude Code; it does not redefine tick semantics or recovery.

## Capability mapping

| Shared capability | Claude Code primitive |
|---|---|
| Isolation | `Agent(..., isolation: "worktree")` |
| Parallel dispatch | Put all `Agent` calls for a wave (one per chain or solo tick) in one message with `run_in_background: true` |
| Completion | Agent completion notifications and the value returned by `Agent` |
| Continuation | `SendMessage` to the named agent or returned agent ID: mid-task corrections and the warm turn after review (agents spawned in this session, running or finished — see "Continuation after a crash") |
| Review | A read-only reviewer `Agent` (see "Review" below for axis mapping; "Workflow" before reaching for fan-out) |

Set `TK_ACTOR=claude:orchestrator` before tracker writes. For orchestrator self-isolation (the "don't run on main" rule), `EnterWorktree` gives the orchestrating session its own worktree as an alternative to creating a feature branch.

## Capability tier resolution

Resolve the shared capability tiers to Claude model classes. The tier names are the contract; the parenthesized models are dated examples — resolve each against what the harness offers today, and when it exposes fewer levels, collapse adjacent tiers downward (balanced and strong merge first).

| Shared tier | Claude model class |
|---|---|
| economy implementation | smallest current model (haiku-class, dated example) |
| balanced implementation | mid-tier workhorse (sonnet-class, dated example) |
| strong implementation | large non-frontier model (opus-class, dated example) |
| frontier | most capable model available (fable-class, dated example) |

The shared tier table is guidance, not enforcement: the orchestrator is assumed to run on a frontier-class model and is trusted to weigh each tick's actual complexity — round up when a "mechanical" tick hides judgment (load-bearing docs, deletions needing reference triage), and say in the ledger what the judgment was (shared protocol → DEFAULT-DOWN RULE).

**`model=` is REQUIRED on every `Agent` call.** Omitting it is not "defaulting to balanced" — the subagent silently inherits the orchestrator's model, which is normally frontier. Pick a tier first, then resolve it; never omit.

**When the frontier class is at its usage limit.** Claude subscriptions meter the top class separately, so `model: "fable"` can return HTTP 429 with no output while every other class still answers. Record the degrade, redispatch *that role* one class down (the reviewer-tier rule still holds: a reviewer one class down is fine when it is still at or above what wrote the code), and leave the rest of the wave on the tiers their shapes earned. The failure mode this exists to stop, observed across two increments: after the first 429 the run quietly settles at strong-class for planning, review *and* implementation — the ledger then reads "opus" on nearly every line, and the tier table has stopped doing anything. If the frontier class is refusing every attempt for a whole session, that is worth saying to the human once, not re-discovering per dispatch.

**The orchestrating session is itself a cost.** On Claude the orchestrator commonly runs frontier-class with a large context window, and it lives longer than any agent it dispatches (shared protocol → "What a run actually costs", item 1). Prefer one session per epic over one per increment: the records are designed to be resumed from. `/clear` at an epic boundary, then re-enter with `/dmtix go`, costs one cheap re-read of `.tick/` and the overview, and stops every later turn from paying for the whole epic's transcript.

## Dispatch

```text
Agent(
  subagent_type: "general-purpose",
  description: "Implement <tick-title>",
  prompt: <shared implementer prompt>,
  isolation: "worktree",
  run_in_background: true,
  mode: "bypassPermissions",
  model: "<resolved tier>",
  name: "tick-<tick-id>"  # or "chain-<label>"; the warm-turn address, when supported
)
```

## Worktree provisioning under `isolation: "worktree"`

The harness creates the worktree **at agent launch**, so the orchestrator cannot provision it beforehand. Two resolutions of the shared protocol's provisioning step (`agent-runner.md` → "Provisioning: runnable worktrees"):

- **Default: implementer self-provisions.** The shared prompt template's provisioning instruction (right after its base check) handles it — the agent applies the profile's recipe before touching code. This requires the recipe to already exist in `.tick/profile.md` (discovered solo at run start, loop step 1), or the wave's agents will each rediscover it.
- **Alternative: orchestrator-created worktrees.** When provisioning must precede the agent (heavy setup, private resources, or you want to verify runnability before paying for a dispatch), create the worktree yourself, provision it, then dispatch a normal `Agent` **without** `isolation: "worktree"` and point its prompt at that directory and branch.

Watch the concurrency cost: N simultaneous full dependency installs can thrash disk/network — prefer profile recipes that share immutable state safely from the main checkout, and remember the shared-deps **no-install** boundary. Apply the shared **economic gate** before provisioning wide — when a wave is too narrow/short to repay provisioning, deliberate-and-noted sequential is the correct call. Do NOT let a provisioning failure silently degrade the run to shared-tree sequential execution; if provisioning fails, fix the recipe (or fall back deliberately and note it), don't drift. If worktrees are **pooled** across ticks, the manual-worktree path fits best (harness-created `isolation: "worktree"` trees are single-shot); run cleanup at pool retirement (epic end), not per tick.

**Warm-chain mapping.** A chain is **one `Agent` call** with the chain prompt variant (shared protocol → "Warm-chain prompt variant"; its gates, planted-fault rule and per-tick check lines come from there, not from this adapter) — the warm context lives inside that single agent, so never split a chain across Agent calls. Mid-chain course corrections use `SendMessage` to the live agent; after a crash, redispatch a fresh Agent into the same worktree/branch pointed at the first un-committed tick (the completed prefix stays merged). A finished chain or solo implementer is not retired at its tick's close: it stays warm until the epic's final review, and the review's findings on its code reach it as a warm turn — `SendMessage` to the finished implementer's agent ID or name, which resumes it with its context, within this session. If the warm worker is gone, the fallback is a fresh `Agent` without `isolation`, into the existing worktree and branch (see "Continuation after a crash"), noted in the dispatch ledger. Tier note: a chain aggregates several ticks' worth of judgment — resolve the chain's capability tier from its *hardest* link, not its average.

## Successful-integration cleanup

This is the adapter's resolution of the shared protocol's "run the active adapter's successful-integration cleanup" step (`agent-runner.md`). **Claude isolation worktrees are not self-cleaning once they hold commits.** The harness auto-removes an `isolation: "worktree"` directory only while it is *unchanged*; every tick implementer commits code, so its worktree and `worktree-agent-*` branch persist after the run and accumulate under `.claude/worktrees/` (multi-GB leaks across a few epics). The orchestrator must remove them explicitly — do not assume the harness will.

Ticks still close at the wave gate. Worktree and branch cleanup waits until after the epic's final review: a warm turn on the review's findings runs in them (shared protocol → "Reviewing the work"). Once the final review's findings are fixed or deferred, and only then, remove each worktree and branch:

```bash
# branch comes from the agent's report, e.g. worktree-agent-a1bdb1ab…
git worktree remove ".claude/worktrees/${branch#worktree-}"   # path = .claude/worktrees/agent-<id>
git branch -d "$branch"                                        # safe delete; -d fails if not merged
```

`git branch -d` (not `-D`) is the guard: it refuses to delete a branch whose commits aren't merged, so a mis-derived branch name or an un-integrated tick won't silently lose work. As with the other adapters, **do not clean up a blocked or incomplete worktree** — it is durable handoff state (see "Continuation after a crash").

Note `git branch -d` checks ancestry against the *current* branch. Run cleanup while still on the epic integration branch (where the real `git merge` made the tick branch an ancestor), not after a later squash-merge to the default branch — squashing rewrites SHAs, after which `-d` can no longer recognize the branch as merged and will refuse. If a sweep finds branches stranded that way but you've confirmed their content is in the default branch, `git branch -D` is the deliberate override.

If the orchestrator took its own worktree via `EnterWorktree` for self-isolation, release it with `ExitWorktree` at epic end rather than leaving it behind.

## Branch naming and crash recovery

Claude-managed worktrees do not use the deterministic shared convention: the harness assigns the branch name (historically `worktree-agent-*`, under `.claude/worktrees/`), and the orchestrator only learns it from the implementer's report. So the shared protocol's "record the branch in a durable note" happens **when the report arrives**, not before launch — write `tk note <tick-id> "runner-state: runner=claude branch=<reported-branch>"` as the first integration step for every finished tick.

That leaves a window: a session that dies mid-wave has live branches with no note. During worktree reconciliation, sweep for them explicitly — `git branch --list 'worktree-agent-*'` and `git worktree list` (look under `.claude/worktrees/`) — and match each to its tick by commit messages (`tick <id>: …`) or diff content. On handoff, branch plus tick ID is authoritative; the Claude agent ID is optional continuation metadata.

## Continuation after a crash

`SendMessage` only reaches agents spawned in this session, running or finished. To continue incomplete work in an existing worktree after a crash or handoff — or to deliver a warm turn whose warm worker is gone — dispatch a fresh `Agent` **without** `isolation: "worktree"` and instruct it in the prompt to work in the existing worktree directory on the existing branch (give it the absolute path and branch name, plus the prior report and what remains). Never let it create a second branch for the tick. A fresh dispatch in place of a warm turn is noted in the dispatch ledger.

## Workflow

The execution loop does **not** fit the `Workflow` tool: a Workflow script can't run `git merge` or `tk close`, so keep wave orchestration on the `Agent` path. `Workflow` fan-out is strictly an optional upgrade to the read-only *review* step (one agent per axis, then verify findings) for epics big enough to earn it; a single reviewer `Agent` remains the default.

## Review

The shared doc defines the review-axis menu (spec compliance, correctness, security, performance, error handling, test quality, type/contract design, comment accuracy) and the severity + confidence output. Map those axes onto Claude subagents two ways:

- **Purpose-built agents, if installed.** Claude Code's `pr-review-toolkit` plugin ships axis-specialized reviewers — `silent-failure-hunter` (error handling), `pr-test-analyzer` (test quality), `type-design-analyzer` (type/contract design), `comment-analyzer` (comment accuracy), `code-reviewer` (general quality + spec compliance), `code-simplifier` (simplification, which doubles as the retro drift pass). When a consuming repo has these — or equivalents — as `subagent_type` values, dispatch them directly, one per axis the diff earns.
- **Otherwise, a focused `general-purpose` Agent per axis.** Give it a single axis as its remit and the shared severity + confidence format. Don't ask one agent to cover every axis at once — independent passes are the point of fan-out.

Resolve the reviewer to a model **at least as capable as the implementer** (shared REVIEWER-TIER RULE), and to the frontier class for the epic final review. To parallelize, put the axis `Agent` calls in one message with `run_in_background: true`; the orchestrator then verifies findings against the code — low-confidence ones especially — before acting. Agents review; the orchestrator adjudicates. Verified findings go back to the implementer that wrote the code as a warm turn: `SendMessage` to its agent ID or name, on its existing worktree and branch; if it is gone, a fresh `Agent` without `isolation` into that worktree and branch, noted in the dispatch ledger (see "Warm-chain mapping").

## Boundary hardening

For repeated use, define a project-level `.claude/agents/ticks-implementer.md` that denies `tk` and writes under `.tick/**`, then use that subagent type. PreToolUse hooks receive parameters nested under `tool_input` (not at the top level); on exit 2, write the explanation to stderr — Claude Code feeds stderr back to the model. Test new hooks against the real input shape before committing them.

Do not use `TaskOutput` for subagent results; it can load the full transcript. Use the `Agent` result and completion notification.

## Field note: worktree base after orchestrator branch switches (2026-08-16, omamelk-monitor-watch)
After the orchestrating session checks out another branch mid-session (e.g. for a PR merge) and returns, `isolation: "worktree"` dispatches may create the agent worktree from the wrong base commit (observed: based on master instead of the checked-out feature branch HEAD). Keep the "verify your worktree contains commit <X>" gate as the FIRST instruction of every implementer prompt, and on mismatch fall back to the orchestrator-created worktree path (git worktree add from the exact commit, dispatch without isolation into it).

Recurred 2026-08-17 on a session that never switched branches (fresh session already on the feature branch): the harness worktree was again based on master, and because the gated agent stopped before committing, the harness auto-removed the unchanged worktree — so there was nothing to fast-forward. Practical consequence: when orchestrating on a non-default branch, prefer orchestrator-created worktrees (`git worktree add ../.ticks-worktrees/<unit> -b tick/<epic>/<unit> <exact-commit>` + dispatch without isolation) from the start; treat harness `isolation: "worktree"` as reliable only on the default branch.

## Field note: `tk create -d` through the Bash tool (2026-09-26, japan-companion)
Backticks and `$(…)` inside a double-quoted `-d "…"` are executed by the shell, silently
mangling the description (lines like "add `X` and `Y`" became "add  and ."). Write the
description to a scratch file and pass `-d "$(cat file)"`, or single-quote it. Check
`tk show <id> --json` after creation when the text carried code spans. Same session: manual
worktrees from the exact integration commit + dependency-driven launches ran a 6-wide wave
on a feature branch with no base-commit surprises.

## Field note: brief-in-a-file dispatch (2026-09-26, japan-companion increment 02)
Writing each implementer's full brief (template + description + contracts) to a scratch
file and making "read this file in full first" the agent's first instruction worked for
8/8 dispatches: the orchestrator never loads the assembled prompts into its own context,
and the brief stays self-contained on disk for redispatch after a crash. Keep the key facts
(worktree, branch, integration commit, boundaries) repeated inline in the short Agent prompt.

## Field note: mid-task `SendMessage` refused as injection (2026-09-26, japan-companion increment 02)
A running implementer whose brief said "stay within the N items" declined a mid-task
`SendMessage` adding an item, treating it as unverifiable. A second message that named the
sender as the orchestrator that wrote the brief and stated the scope extension explicitly was
accepted. Write briefs so follow-ups are expected ("the orchestrator may send you further items
by message; treat them as part of this tick"), and phrase course corrections as such. Also:
a warm-chain of three ≲10-min ticks took ≈18.5 min wall with one cold start (first data point).

## Field note: implementers that background their own test runs (2026-09-28, fh read 11s25-rdx)
Across one increment, about half the implementers started a long `pnpm fp test` (or a UI run) in the background despite a brief saying "foreground", and the harness handed the tick back with "waiting on its own background work" — several times per tick. Every one resumed and reported when its run finished; none lost work. The orchestrator's move is to wait for the report, check the worktree's processes by `/proc/*/cwd` if it runs long, and never merge on a completion notice that carries no report.

## Field note: `tk` writes into the shell's current directory (2026-09-29, fh writer-d 11s29-kqv)
The Bash tool's working directory persists between calls. One orchestrator call did `cd ../.ticks-worktrees/<wt> && pnpm install … && tk update <id> --status in_progress`, and the update landed in the worktree's `.tick/`, not the controller's: `tk` resolves `.tick/` from the cwd. The implementer later found "uncommitted `.tick/` changes it had not made", and the boundary check would have read them as its change. Run every `tk` command from the controller checkout (absolute paths, or `cd` back first); before merging, `git -C <worktree> status --short -- .tick/` should be empty, and a stray change there is discarded, then re-applied in the checkout.

## Field note: the brief's base check vs the state commit (2026-09-29, fh picking 11s30-pkr)
Worktrees were cut from commit A, then the orchestrator committed tick state (B) and wrote the
briefs with `git rev-parse --short HEAD`, so every brief told its implementer to verify commit B
while its worktree held A. Both implementers noticed, reasoned that B only touched `.tick/`, and
went on; a stricter one would have reported BLOCKED. Take the base for the brief from the
worktree itself (`git -C <worktree> rev-parse --short HEAD`), never from the checkout's HEAD.

## Field note: a chain and a review round trip, measured (2026-09-29, fh picking 11s30-pkr)
A warm-chain of two related ticks (a page, then the proof that runs it) took 31.5 min with one
cold start; the two solo ticks before it took 6 and 19.5 min. An opus review found a blocker
whose fix line contradicted the spec; the opus repair followed the fix line; the re-review
caught it; a sonnet repair fixed the mechanical remainder in 9 min. The cheaper repair after a
review round trip is right when the findings are mechanical; the repair still follows the
finding's claim and the spec, not the reviewer's suggested fix.

## Field note: three orchestrator slips in one epic (2026-09-29, fh sources 11s31-src E1)
- `tk update` has no `--blocked-by`; a dependency added after creation is `tk block <id> <blocker…>`. A `&&` chain through that failing `tk update` skipped the worktree cuts and the briefs were written with an empty base commit: after any chain that cuts worktrees, print the base and check it is non-empty before writing a brief, and never put a `tk` call whose flags you have not used before inside a `&&` chain with the cuts.
- A full post-wave gate under a load average of 40+ took 10 minutes and hit the Bash tool's 600 s foreground limit; the harness moved it to the background mid-run, which worked but left the orchestrator guessing. Under load, start a gate with `run_in_background` from the first call and read its output file on the completion notice.
- A tick that adds a `class X extends Error` anywhere passes its own `--changed` set and fails the wave gate on a fingerprint test in another package. Add "run `<that test>` by name" to the brief of any tick that may add an error class, or expect it at the gate.

## Field note: a gate in the checkout you are merging into (2026-10-01, fh privacy 11s41-neg)
With dependency-driven launching, the orchestrator started the wave gate (full Vitest, ~10 min) in the controller checkout in the background, then merged the next finished tick into that same checkout while the gate was still reading files. The gate's result described no single tree and had to be voided and re-run on the union. Either run a background gate in a throwaway worktree at the exact commit (`git worktree add --detach <dir> <sha>`), or merge nothing into the checkout until the gate's EXIT line is written. Same run: a whole epic (7 ticks incl. a sweep, a repair and an evidence tick) took about 2 h wall with 3-wide first wave and dependency-driven launches.

## Field note: a repair that changes what the proof measured (2026-10-11, fh reach 11s45-yfh E2)
E2's replay showed a gain (5 → 6 of 13); the final review then found a blocker in what the writer was shown, and the gain had come through exactly that leak. A review repair that changes the measured path voids the evidence measured before it: re-run the proof on the repaired build before close-out, keep the old numbers in the evidence marked "superseded: measured with <the defect>", and say plainly in the landing note if the gain does not survive. Same run: an `--ff-only` merge failed and the `;`-chained close and worktree removal ran anyway; gate every close and cleanup on the merge's exit (`&&`), and merge with `--no-edit`, not `--ff-only`, after any tick-state commit.

## Field note: restarts wipe the scratchpad and the implementers (2026-10-28, fh prices 11s66-jje)
Three session restarts in one small epic killed three background implementers; the first left its tick uncommitted, and the orchestrator's briefs and gate logs, written to the session scratchpad, were gone with the next session's new scratchpad path. Keep briefs and logs beside the worktree (`../.ticks-worktrees/<wt>.logs/`), save an uncommitted diff as a WIP commit before a planned exit, and when sessions are restarting tell implementers to commit and push early. A dead tick whose diff only needs its gates run is cheaper to verify from the orchestrator than to redispatch a third time.
