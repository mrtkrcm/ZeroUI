# ZeroUI

ZeroUI is a zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It provides both a CLI and an interactive TUI built with Charm libraries (Bubble Tea, Huh, Lipgloss).

This README focuses on a concise onboarding and developer quickstart so you can be productive in minutes.

---

## 1-minute quickstart

Prerequisites:

- A recent Go toolchain (see `go.mod` for the minimum required version).
- Basic Unix tooling (make, git).

From the repository root:

1. Prepare test stubs (idempotent):

   ```bash
   make test-setup
   ```

2. Build the binary:

   ```bash
   make build
   ```

3. Run the interactive TUI:

   ```bash
   ./build/zeroui
   ```

4. Run fast unit tests:
   ```bash
   make test-fast
   ```

That's it — you should be able to explore the app and run tests locally within a few minutes.

---

## Developer quickstart

This section expands the most common developer tasks.

### Build & run

- Build the binary:

  ```bash
  make build
  ```

  Output: `./build/zeroui`

- Install to `GOBIN` / `GOPATH/bin`:

  ```bash
  make install
  ```

- Run in development:
  ```bash
  make run
  ```

### Tests

- Prepare deterministic test stubs (ensures `testdata/bin` binaries are executable):

  ```bash
  make test-setup
  ```

- Fast tests (short unit tests only):

  ```bash
  make test-fast
  ```

- Deterministic full test run (CI-like; relaxes visual tests):

  ```bash
  make test-deterministic
  ```

- Full test run with coverage and HTML report:

  ```bash
  make test
  # coverage report is at build/coverage/coverage.html (see Makefile)
  ```

- Update TUI visual baselines (local-only; run and review diffs carefully):
  ```bash
  make test-update-baselines
  ```

Notes:

- The repo includes repo-local stub binaries under `testdata/bin/` (e.g., `ghostty`) used by tests to avoid relying on system-installed CLIs.
- Several packages include a package-level `TestMain` that will prepend `testdata/bin` to `PATH` and set an isolated `HOME` for reproducible tests.

### Formatting & linting

- Format code:

  ```bash
  make fmt
  ```

- Run linters:

  ```bash
  make lint
  ```

  (Requires `golangci-lint`; install it if needed.)

- Security checks:
  ```bash
  make security
  ```
  (Optional; some tools are required to be installed locally.)

### Development & iteration

- Watch / rebuild loop (requires `entr`):
  ```bash
  make dev
  ```
- Run tests in watch mode (requires `entr`):
  ```bash
  make test-watch
  ```

---

## Key UX notes / TUI controls

- Global
  - `q` / `Ctrl+C`: quit
  - `?`: toggle help
  - `/`: search where supported

- App List
  - `enter` / `space`: select app
  - `r`: refresh apps

- Form (config editor)
  - `tab` / `shift+tab`: navigate fields
  - `enter`: select/confirm
  - `ctrl+s`: save
  - `C`: changed-only view
  - `p`: open presets selector
  - `u`: undo last save (restore most recent backup)
  - `esc`: back to app list

---

## Testing environment details (short)

- Repo-local test stubs: `testdata/bin/` contains deterministic stubs used by tests.
- `make test-setup` ensures the executables are present and executable.
- Packages with many tests often include `TestMain` implementations that:
  - Prepend `testdata/bin` to `PATH`.
  - Create an isolated `HOME` directory for test execution.
  - Restore original environment after tests.

If you add tests that exec external tools, add a small deterministic stub under `testdata/bin` and prefer using the provided `test/helpers/testing_env.go` helpers.

---

## Contributing & PR checklist

When preparing a change, please:

- Run `make fmt` to format code.
- Run `make test-fast` locally; for CI parity run `make test-deterministic`.
- Run `make lint` and address critical linter warnings if possible.
- If you change TUI visuals, run `make test-update-baselines` locally and keep baseline changes focused and reviewed.
- Keep changes small and focused; update docs/README if you add user-visible CLI flags or flows.

---

## Troubleshooting

- Tests failing because of missing stub binaries:
  - Run `make test-setup` and ensure `testdata/bin/*` is executable.
- Linter command not found:
  - Install `golangci-lint` (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`) or run formatting locally.
- Visual test diffs:
  - Visual tests can be updated locally with `make test-update-baselines`. Always review diffs carefully before committing.

---

## Where to look in the codebase

- `cmd/` — CLI entrypoints and Cobra commands.
- `internal/tui/` — TUI implementation (Bubble Tea components, styles).
- `internal/config/` — config detection, parsing, provider integrations.
- `internal/plugins/rpc/` — plugin/rpc registry and helpers.
- `pkg/` — reusable public packages (extractors, references).
- `testdata/bin/` — deterministic test stubs used during testing.

---

If you'd like, I can:

- apply automated formatting across the repo (run `make fmt` and persist changes),
- triage and fix a couple of top linter warnings,
- or add a short `CONTRIBUTING.md` derived from this quickstart.

Pick one and I'll proceed.
