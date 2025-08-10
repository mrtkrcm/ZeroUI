# Bubble Tea UI Implementation Plan - 2025-08-11

## Objective
Implement a full-featured Bubble Tea TUI that displays supported applications as cards, showing their installation and configuration status. The UI should launch when the program runs without arguments.

## Source Analysis
- **Source Type**: Feature requirements specification
- **Core Features**: 
  - Application card grid layout
  - Application logo display
  - Installation status detection
  - Configuration status detection
  - Interactive card selection
  - Auto-create config for installed apps
- **Dependencies**: Bubble Tea, Lipgloss (already in project)
- **Complexity**: Medium-High

## Target Integration
- **Integration Points**: 
  - Main command (`cmd/root.go`) - detect no args
  - Application discovery (`internal/toggle/engine.go`)
  - Configuration detection (`internal/config/loader.go`)
  - Executable detection (new utility)
- **Affected Files**:
  - `cmd/root.go` - Launch UI when no args
  - `internal/tui/app.go` - Main UI implementation
  - `internal/tui/components/app_card.go` - New card component
  - `internal/tui/components/app_grid.go` - New grid layout
  - `internal/tui/util/exec_detector.go` - New executable detection
- **Pattern Matching**: Use existing component patterns from `internal/tui/components/`

## Implementation Tasks

### Phase 1: Foundation (Setup)
- [x] Analyze existing TUI structure
- [ ] Create new state for app cards view
- [ ] Define app metadata structure (name, logo, executable)

### Phase 2: Core Components
- [ ] Create AppCard component with:
  - [ ] App logo/icon display
  - [ ] App name
  - [ ] Installation status indicator
  - [ ] Configuration status indicator
  - [ ] Dimmed state for unavailable apps
- [ ] Create AppGrid component for card layout
- [ ] Implement executable detection utility
- [ ] Add app metadata registry

### Phase 3: State Management
- [ ] Detect when program runs without arguments
- [ ] Load all supported app definitions
- [ ] Check executable existence for each app
- [ ] Check configuration existence for each app
- [ ] Update card states based on status

### Phase 4: Interactions
- [ ] Handle card selection (arrow keys/mouse)
- [ ] Create config on selection if executable exists
- [ ] Navigate to config editor after creation
- [ ] Handle back navigation to grid

### Phase 5: Visual Polish
- [ ] Add app logos/icons (Unicode or ASCII art)
- [ ] Implement responsive grid layout
- [ ] Add smooth transitions
- [ ] Style dimmed/active states

## Supported Applications Registry
```yaml
apps:
  - name: Ghostty
    executable: ghostty
    logo: "👻"
    config_path: ~/.config/ghostty/config
  - name: Alacritty
    executable: alacritty
    logo: "🖥️"
    config_path: ~/.config/alacritty/alacritty.yml
  - name: VS Code
    executable: code
    logo: "📝"
    config_path: ~/.config/Code/User/settings.json
  - name: Neovim
    executable: nvim
    logo: "📜"
    config_path: ~/.config/nvim/init.lua
  - name: Zed
    executable: zed
    logo: "⚡"
    config_path: ~/.config/zed/settings.json
  - name: WezTerm
    executable: wezterm
    logo: "🪟"
    config_path: ~/.config/wezterm/wezterm.lua
```

## Validation Checklist
- [ ] UI launches when no arguments provided
- [ ] All supported apps display as cards
- [ ] Installation status correctly detected
- [ ] Configuration status correctly shown
- [ ] Dimmed state for unavailable apps
- [ ] Card selection works (keyboard & mouse)
- [ ] Config creation on selection works
- [ ] Navigation between views works
- [ ] Tests written for new components
- [ ] No existing functionality broken

## Risk Mitigation
- **Potential Issues**: 
  - Executable detection across platforms
  - Performance with many apps
  - Grid layout responsiveness
- **Rollback Strategy**: Git checkpoint before implementation

## Architecture Notes
- Leverage existing Model/View/Update pattern
- Reuse existing components where possible
- Maintain consistency with current TUI structure
- Keep executable detection platform-agnostic