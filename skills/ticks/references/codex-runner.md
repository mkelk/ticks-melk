# Codex Runner Adapter

Read [`agent-runner.md`](agent-runner.md) first. This file maps its capability contract onto Codex. It uses standard git worktrees plus `codex exec`, so it works even when the current Codex harness does not expose an in-process subagent tool.

## Capability mapping

| Shared capability | Codex primitive |
|---|---|
| Isolation | `git worktree add -b tick/<epic>/<tick> <path> <integration-commit>` |
| Parallel dispatch | One background `codex exec -C <worktree>` process per chain (parallel warm-chains are the default) or solo tick |
| Completion | Shell `wait` plus `--output-last-message <report>` |
| Continuation | `codex exec resume` when a session is available (mid-task corrections, the warm turn after review); otherwise a fresh dispatch in the same worktree, noted |
| Review | `codex exec review` or a read-only `codex exec` against the diff |

Set `TK_ACTOR=codex:orchestrator` before tracker writes.

## Dispatch

Create worktrees from the current integration commit, not from an assumed default branch:

```bash
repo=$(git rev-parse --show-toplevel)
integration_commit=$(git rev-parse HEAD)
codex_model="${CODEX_MODEL:-gpt-5.5}" # default is a dated example — resolve to the current strongest Codex model
branch="tick/<epic-id>/<tick-id>"
worktree="$(dirname "$repo")/.ticks-worktrees/<tick-id>"
report="$(dirname "$repo")/.ticks-worktrees/<tick-id>.report"
log="$(dirname "$repo")/.ticks-worktrees/<tick-id>.log"

git worktree add -b "$branch" "$worktree" "$integration_commit"
tk note <tick-id> "runner-state: runner=codex branch=$branch worktree=$worktree base=$integration_commit"

codex exec \
  -C "$worktree" \
  --model "$codex_model" \
  --config 'model_reasoning_effort="medium"' \
  --sandbox workspace-write \
  --output-last-message "$report" \
  - < "$prompt_file" > "$log" 2>&1 &
pid=$!
```

Prepare `prompt_file` from the shared implementer template (or its warm-chain variant); it already carries the tick gate, the planted-fault rule and the report shape, so add the tick's data, not a second gate list. Launch every ready chain or solo tick before waiting. Keep an in-memory `unit -> pid` map (and the unit's session ID, for its warm turn), then use shell `wait <pid>` for completion. This is blocking process supervision, not polling. Read the small report first; inspect the log only for failures or missing status.

Ticks still close at the wave gate, but the worktree and branch stay until after the epic's final review: a warm turn on the review's findings runs in them (shared protocol → "Reviewing the work"). Once the final review's findings are fixed or deferred, remove each manually created worktree before deleting its merged branch:

```bash
git worktree remove "$worktree"
git branch -d "$branch"
rm -f "$report" "$log" "$prompt_file"
```

Do not clean up a blocked or incomplete worktree; it is durable handoff state.

## Model and effort policy

Codex cost control is primarily an **effort** decision, not a small-model decision. Do not automatically downshift implementation to mini/nano models. Unless the user configures another mapping, use the current strongest generally available Codex model for every role and vary `model_reasoning_effort`:

| Work | Model | Reasoning effort |
|---|---|---|
| Mechanical implementation with precise acceptance | strongest Codex model | `low` |
| Normal implementation | strongest Codex model | `medium` |
| Subtle or integration-heavy implementation | strongest Codex model | `high` |
| Epic planning and decomposition | strongest Codex model | `high` |
| Foundation, security, and final review | strongest Codex model | `high` |

At the time this adapter was updated, the recommended strongest model is `gpt-5.5`; verify the current Codex model catalog when "latest" matters. Prefer project profiles for stable automation, for example `codex exec --profile implementation-low` and `--profile review-high`. Mini, nano, or other reduced-capability models are opt-in only.

## Continuation and handoff

`codex exec resume <session-id>` is useful within one live orchestration session, but session IDs are not durable runner state. It is the warm turn: an implementer is retired after the epic's final review, not after its tick, and the review's findings on its code go back to it by resuming its session. After a crash or cross-runner handoff, or when the warm worker is gone, run a fresh `codex exec -C <existing-worktree>` with the tick, prior report, review feedback, and current branch state — the fallback for a warm turn too, noted in the dispatch ledger. Never create a second branch for the tick.

Codex loads `AGENTS.md` automatically. The rest of the read order comes from the shared implementer template (instruction file, learnings head `.tick/learnings.md` with `.tick/learnings-ref.md` looked up by category when the head points there, then `.tick/config.md`); keep it in the prompt. Those files are execution inputs even though the implementer must not modify `.tick/`.

## Boundary hardening

Use Codex exec policies or hooks, when available, to deny `tk` and writes under `.tick/**`. `workspace-write` limits filesystem access but does not by itself enforce that logical boundary, so the orchestrator must still run the pre-merge `.tick/` diff check.
