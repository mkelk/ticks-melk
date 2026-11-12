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

**Runtime tripwire (2026-11-12, load avg ~42 on 16 cores from other sessions):** the baseline at 403d086a (before any 01-kyl change) was red, and the failing set was not stable:
- Pi: 9 of 158 failed (lease/heartbeat/SIGKILL/concurrent-controller tests in runner.test.ts and planning.test.ts). A serial re-run of those two files failed a *different* 5 of 72.
- Go: `TestDetailStatusEditGolden` (internal/tui) failed in the full run and passed alone.

These are timing flakes under host load, not code faults. Seen again in z5t (load 20-48): `internal/tui` WaitFor-5s goldens (TestSmokeGolden, TestBoardGolden, TestDetailStatusEditGolden) and the planning.test.ts crash-recovery family (:369 SIGKILL, :499 recovery cases, partial-apply) — red at base with a different set each run. Wave-gate rule while load is high: a failure counts only if it reproduces in an isolated re-run of that test. CI on the epic PR is the authoritative gate (config Rules).

## Docs-by-path readers
- `extensions/ticks-runner/planning.ts` reads `skills/ticks/references/tick-patterns.md` by path at runtime; edits to that file can affect Pi planning tests even though no test imports it.

## Warm/cold ratio
Observed in z5t (2026-11-12, opus implementers, load 20-48), not like-for-like:
- Cold solo tick (zg4, Go TDD, 5 files): 4.6 min, 78k tokens.
- Warm-chain of 6 docs ticks in one file (chain P): 4.4 min total, ≈0.7 min per tick after one cold start; chain of 4 docs ticks across 5 files (S): 10.7 min.
- Warm turns after the final review: 51 s-154 s, +10k-20k tokens each (vs a fresh cold start ≈4-5 min, ≈80k).
Working ratio for the gate: a warm docs tick ≈ 1/5 of a cold one; a warm review turn ≈ 1/3. Re-measure on a code-heavy chain.
