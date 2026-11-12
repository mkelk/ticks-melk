# Learnings — head (always read)

Hard rules for every implementer, then an index of `.tick/learnings-ref.md` (the body).
Budget: at most 4096 bytes. The engine owns the format; this repo owns the content.

## Hard rules
- Go is a mise install not active here: run `mise exec go@1.24 -- go …`; run the Pi suite with `PATH=$HOME/.local/share/mise/installs/go/1.24.11/bin:$PATH`. Never write a mise config.
- pnpm only — never npm or yarn.
- A new cobra subcommand is registered in `cmd/tk/cmd/*.go` AND the switch + usage in `cmd/tk/main.go`.
- Drive commands in tests only via `ExecuteArgs`/`ExecuteArgsContext`; a new persistent flag var is also reset in `ResetFlags()` (root.go).
- Any `schemas/` edit runs `make codegen-go` and `scripts/build-ui.sh`; commit all generated output together.
- UI source changes run `scripts/build-ui.sh` and commit the regenerated `internal/tickboard/server/static/`.
- Every `tk` command written into docs/UI/copy is checked against `cmd/tk/cmd/*.go` (`Use:`/`Args:`).
- The host is often loaded: a red test counts only if the same test is not also red at the base commit (A/B).

## Index (`.tick/learnings-ref.md`)
- Tick authoring — before specifying a flag/field another tool consumes.
- This repo's build — cobra routing, in-process command tests, UI embed, CI pnpm setup, worker tests, Pi suite needing Go.
- Schema codegen — before touching `schemas/` (two codegen consumers; date fields).
- Docs & marketing copy — before writing `tk` examples into user-facing text (known CLI traps).
- Orchestration — orchestrator only: committing tick state, stale worktree bases, `model=`, shell quoting in `tk -d`, half-applied merges.
- TUI (internal/tui) — before changing shared chrome or view order (golden contamination, positional hotkeys).
