# Tick Run Configuration

## Testing

- Pi runner: `node --test --no-warnings extensions/ticks-runner/*.test.ts`
- Go: `go test -short -count=1 ./...`
- Go note: internal/worktree can fail locally when temporary repositories lack git identity; it passes in CI. Do not chase that environmental baseline.
- UI hint: when UI source changes, run pnpm install frozen, TypeScript noEmit, and targeted Vitest files; the full suite has pre-existing failures.
- Worker hint: when worker source changes, run pnpm install frozen, TypeScript noEmit, and targeted Vitest files; full pnpm test has a known workerd boot crash.

## Closeout Evidence Commands

- Package RPC discovery: `node --no-warnings scripts/verify-pi-ticks-qfs.ts package-rpc`
- Four-command dry/status smoke: `node --no-warnings scripts/verify-pi-ticks-qfs.ts command-smoke`
- Dry-run wave plan: `node --no-warnings scripts/verify-pi-ticks-qfs.ts dry-run-wave-plan`
- Real live scenario: `node --no-warnings scripts/verify-pi-ticks-qfs.ts live-scenario`
- Dashboard demo/dump: `node --no-warnings scripts/verify-pi-ticks-qfs.ts dashboard-dump`
- Docs/install content: `node --no-warnings scripts/verify-pi-ticks-qfs.ts docs-install`

## Acceptance Evidence

- A1: `node --no-warnings scripts/verify-pi-ticks-qfs.ts package-rpc`
- A2: `node --no-warnings scripts/verify-pi-ticks-qfs.ts command-smoke`
- A3: `node --no-warnings scripts/verify-pi-ticks-qfs.ts dry-run-wave-plan`
- A4: `node --no-warnings scripts/verify-pi-ticks-qfs.ts live-scenario`
- A5: `node --no-warnings scripts/verify-pi-ticks-qfs.ts dashboard-dump`
- A6: `node --no-warnings scripts/verify-pi-ticks-qfs.ts docs-install`

## Pi Orchestrator

- planner_model: openai-codex/gpt-5.6-sol:xhigh
- scout_model: openai-codex/gpt-5.6-sol:low
- implement_economy_model: openai-codex/gpt-5.6-sol:low
- implement_balanced_model: openai-codex/gpt-5.6-sol:medium
- implement_strong_model: openai-codex/gpt-5.6-sol:high
- review_model: openai-codex/gpt-5.6-sol:xhigh
- max_parallel: 4

## Environment

- `which go` — Go toolchain on PATH
- `which pnpm` — pnpm on PATH (never npm/yarn in this repo)
- `git config user.email` — git identity configured

## Rules

- Epic integration goes through a PR + CI gate: the orchestrator pushes the epic branch and opens a PR; the epic close-out may not complete until the CI workflow (`.github/workflows/ci.yml`) is green on that PR. No direct merges of epic branches to the default branch.
- Package management is pnpm only — never npm or yarn.
- After UI source changes, run `scripts/build-ui.sh` and commit regenerated `internal/tickboard/server/static/`.
- Any edit under `schemas/` must run both `make codegen-go` and `make codegen-ts` and commit all regenerated output together.

## At run start
- Branch guard (verify, don't ask): `git rev-parse --abbrev-ref HEAD` must not be
  `main`/`master` — devmeta runs only on an increment base branch. On main/master, stop
  and tell the human (or switch to the active increment's branch per
  `.devmeta/current-increment.md` → its `_overview.md` "How to run").

## At epic close-out
Resolve the ACTIVE increment from `.devmeta/current-increment.md`: its directory
is <DIR>, its id is <ID>. Then:
1. Write the full epic-close retro report to `<DIR>/ia-cycles/<epic-title>.md`
   (one-line summary as the tk close --reason). Include the dispatch-mode and
   review-depth ledgers.
2. Append a narrative entry to `.devmeta/project-history.md`.
3. Update `<DIR>/_overview.md` (mark this epic DONE).
4. Living-docs audit: fix any doc the epic proved wrong.
5. Tag: `git tag <ID>.<epic-slug> -m "<summary>"`.

## At project checkpoint
Same resolution as above. Then:
1. Write the increment completion report to `<DIR>/completion.md`
   (features, gates status, outstanding human items, postmortems).
2. Integration stops at the increment's base branch: never merge it to `main`/`master` —
   shipping the increment is the human's decision at this checkpoint.
3. STOP for human review.
