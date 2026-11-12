# Project execution profile (inferred — maintained by the orchestrator)

Characterized 2026-11-12 (increment 01-kyl run start). Inputs: go.mod (go 1.24.11), root package.json (no dependencies), extensions/ticks-runner/*.test.ts, .tick/config.md Testing.

## Provisioning recipe
- Docs and Pi runner: none. Node 26 runs `.test.ts` directly; the root package.json has no dependencies. Confidence: high (baseline run).
- Go: the toolchain is installed via mise but not active in this directory, so `go` is not on PATH. Prefix every Go command with `mise exec go@1.24 --` (e.g. `mise exec go@1.24 -- go test -short -count=1 ./...`). Do not write a mise config to activate it. Confidence: high.
- UI (`internal/tickboard/ui/`) and worker (`cloud/worker/`): `pnpm install --frozen-lockfile` in the package dir; not needed unless a tick touches them.

## Test-tier venues
| Tier | Command | Venue | Evidence |
|---|---|---|---|
| Pi runner unit | `node --test --no-warnings extensions/ticks-runner/*.test.ts` | in-worktree | 11 test files use mkdtemp/tmpdir; no fixed ports |
| Go unit | `mise exec go@1.24 -- go test -short -count=1 ./...` | in-worktree | 23 Go test files use t.TempDir/MkdirTemp; no shared services |
| Go vet/build | `mise exec go@1.24 -- go build ./...` | in-worktree | pure compile |

Known baseline: `internal/worktree` can fail locally when temp repos lack git identity (passes in CI) — config Testing note.

## Docs-by-path readers
- `extensions/ticks-runner/planning.ts` reads `skills/ticks/references/tick-patterns.md` by path at runtime; edits to that file can affect Pi planning tests even though no test imports it.

## Warm/cold ratio
Not yet observed in this repo.
