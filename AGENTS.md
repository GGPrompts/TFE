# Repository Guidelines

## Project Structure & Module Organization
- Root Go modules (`main.go`, `update_keyboard.go`, `render_preview.go`, etc.) each own a single concern; extend an existing file or create a new focused module instead of piling logic into `main.go`.
- `docs/` stores narrative references—update `CLAUDE.md` whenever architecture or flows shift, and keep `HOTKEYS.md` aligned with UI changes.
- `Screenshots/` holds UI captures used in `README.md`; refresh them when visual tweaks alter the layout.
- `tfe-wrapper.sh` powers the Quick CD feature; update it alongside behavior changes in `command.go` or `main.go`.

## Build, Test, and Development Commands
- `go run .` starts TFE from the repo root for iterative work; use it after modifying Bubble Tea update loops.
- `go build -o tfe` produces a distributable binary in place—run before publishing releases or docs screenshots.
- `go test ./...` executes all package tests; add tests before new features so the command fails when coverage is missing.
- `go vet ./...` surfaces common mistakes (unused params, shadowed vars); run before opening PRs that touch core navigation or command execution.

## Coding Style & Naming Conventions
- Format with Go 1.24 tooling: `gofmt -w` (or `go fmt ./...`) and keep imports tidy using `goimports`.
- Stick to Go naming: exported symbols use PascalCase, internal helpers stay camelCase; follow existing file naming (e.g., `render_*`, `update_*`) to signal module scope.
- Keep UI copy and constants near their renderers to ease theming, and avoid hard-coding behavior in `main.go`.

## Testing Guidelines
- Place `_test.go` beside the source file (`render_preview_test.go`, `file_operations_test.go`) and prefer table-driven cases for state machines in `update_*.go`.
- Use temp directories and `t.Cleanup` when exercising filesystem helpers, and include golden snapshots for preview rendering where practical.
- Target coverage on navigation flows, preview loaders, and command execution—areas most likely to regress during UI refinements.

## Commit & Pull Request Guidelines
- Follow the Conventional Commit scheme visible in history (`feat:`, `fix:`, `docs:`) and keep commits scoped to one user-facing change.
- PR descriptions should outline behavior, list the manual commands you ran (`go run .`, `go test ./...`), and link issues or backlog entries.
- Attach screenshots/asciicasts for visual changes and note any documentation updates (`HOTKEYS.md`, `README.md`, `CLAUDE.md`) required by the patch.
- Ensure updates that affect Quick CD mention any necessary changes to `tfe-wrapper.sh` so shell integrations stay in sync.
