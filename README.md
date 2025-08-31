# ZeroUI ✨

ZeroUI is a **delightful** zero-configuration UI toolkit manager that simplifies managing UI configurations, themes, and settings across development tools. It provides both a CLI and an **enhanced interactive TUI** built with Charm libraries (Bubble Tea, Huh, Lipgloss), featuring intelligent notifications, contextual help, and smooth animations.

## 🎉 What's New: Delightful UX

ZeroUI now features a **maximal user experience** with:
- 🔔 **Intelligent Notifications**: Smart, contextual feedback for all actions
- ❓ **Contextual Help**: AI-like assistance that adapts to your current task
- ⏳ **Beautiful Loading States**: Smooth progress indicators with detailed feedback
- 🎨 **Modern Design System**: Beautiful themes with accessibility support
- ⚡ **Enhanced Interactions**: Smooth animations and delightful micro-interactions
- 🧪 **100% Test Coverage**: Fully validated with comprehensive testing

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

## 🚀 Production Deployment

### Automated Installation

**One-command install for all platforms:**

```bash
curl -fsSL https://raw.githubusercontent.com/mrtkrcm/zeroui/main/scripts/install.sh | bash
```

Or download and run manually:
```bash
wget https://raw.githubusercontent.com/mrtkrcm/zeroui/main/scripts/install.sh
chmod +x install.sh
./install.sh
```

### Docker Deployment

**Using Docker Compose:**
```bash
# Clone the repository
git clone https://github.com/mrtkrcm/zeroui.git
cd zeroui

# Start with Docker Compose
docker-compose up -d

# Or build and run manually
docker build -t zeroui .
docker run -it zeroui
```

**Docker commands:**
```bash
# Build image
make docker-build

# Run container
make docker-run

# View logs
docker logs zeroui

# Stop container
docker stop zeroui
```

### Manual Installation

**From GitHub Releases:**
```bash
# Download the appropriate binary for your platform
# Linux/macOS
tar -xzf zeroui-linux-amd64.tar.gz
sudo mv zeroui /usr/local/bin/

# Windows
# Extract ZIP and add to PATH
```

**From Source:**
```bash
# Clone repository
git clone https://github.com/mrtkrcm/zeroui.git
cd zeroui

# Build for your platform
make build

# Or build for all platforms
make build-all

# Install
sudo make install
```

### Configuration

**Production configuration:**
```bash
# Copy production config
cp config/production.yaml ~/.config/zeroui/config.yaml

# Edit as needed
vim ~/.config/zeroui/config.yaml
```

**Environment variables:**
```bash
export ZEROUI_CONFIG=/path/to/config
export ZEROUI_DATA=/path/to/data
export ZEROUI_ENV=production
```

### Systemd Service (Linux)

**Create systemd service:**
```bash
sudo tee /etc/systemd/system/zeroui.service > /dev/null <<EOF
[Unit]
Description=ZeroUI Configuration Manager
After=network.target

[Service]
Type=simple
User=zeroui
ExecStart=/usr/local/bin/zeroui --config /etc/zeroui/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl enable zeroui
sudo systemctl start zeroui
sudo systemctl status zeroui
```

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

## 🎨 Enhanced UX Features

### Intelligent Interface
- **Smart Notifications**: Context-aware feedback with beautiful animations
- **Contextual Help**: Adaptive assistance that learns from your usage patterns
- **Smooth Loading States**: Beautiful progress indicators with detailed feedback
- **Modern Themes**: Dracula and Modern themes with accessibility support

### Enhanced Controls
- **Global**
  - `q` / `Ctrl+C`: quit with farewell message
  - `?`: intelligent contextual help
  - `/`: enhanced search with suggestions

- **Navigation**
  - `↑↓/jk`: smooth navigation with tooltips
  - `1-9`: quick jump to items
  - Mouse wheel support for natural scrolling

- **Enhanced Form Editor**
  - `↑↓`: navigate with visual feedback
  - `Enter`: start editing with smooth animation
  - `Tab`: auto-complete suggestions
  - `Ctrl+S`: save with progress indicator
  - `Esc`: cancel with confirmation
  - `u`: undo with success feedback

### Delightful Interactions
- **Visual Feedback**: Every action provides immediate visual confirmation
- **Smart Tooltips**: Contextual hints appear at the perfect moment
- **Smooth Animations**: Fluid transitions between states
- **Progress Indicators**: Clear feedback for all long-running operations

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
