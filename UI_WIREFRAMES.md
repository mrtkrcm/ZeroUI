# ZeroUI Application Grid - ASCII Wireframes

## Main Grid View (Default - No Arguments)
```
┌────────────────────────────────────────────────────────────────────────────┐
│                                                                            │
│  🚀 ZeroUI Applications                                                   │
│  Select an application to configure (9 apps)                              │
│                                                                            │
│  ╭──────────────────╮  ╭──────────────────╮  ╭──────────────────╮       │
│  │        👻        │  │        🖥️        │  │        🪟        │       │
│  │                  │  │                  │  │                  │       │
│  │     Ghostty      │  │    Alacritty     │  │     WezTerm      │       │
│  │     Terminal     │  │     Terminal     │  │     Terminal     │       │
│  │                  │  │                  │  │                  │       │
│  │   ✓ Installed    │  │   ✓ Installed    │  │  ✗ Not Installed │       │
│  │   ⚙ Configured   │  │   📄 Has Config  │  │                  │       │
│  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯       │
│                                                                            │
│  ╭──────────────────╮  ╭──────────────────╮  ╭──────────────────╮       │
│  │        📝        │  │        📜        │  │        ⚡        │       │
│  │                  │  │                  │  │                  │       │
│  │     VS Code      │  │      Neovim      │  │        Zed       │       │
│  │      Editor      │  │      Editor      │  │      Editor      │       │
│  │                  │  │                  │  │                  │       │
│  │   ✓ Installed    │  │   ✓ Installed    │  │  ✗ Not Installed │       │
│  │   ⚙ Configured   │  │   📄 Has Config  │  │                  │       │
│  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯       │
│                                                                            │
│  ╭──────────────────╮  ╭──────────────────╮  ╭──────────────────╮       │
│  │        🔲        │  │        🌳        │  │        🚀        │       │
│  │                  │  │                  │  │                  │       │
│  │       Tmux       │  │        Git       │  │     Starship     │       │
│  │   Multiplexer    │  │   Development    │  │       Shell      │       │
│  │                  │  │                  │  │                  │       │
│  │   ✓ Installed    │  │   ✓ Installed    │  │   ✓ Installed    │       │
│  │                  │  │   ⚙ Configured   │  │   📄 Has Config  │       │
│  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯       │
│                                                                            │
│  ↑↓←→ Navigate • ⏎ Select • [a] Show All • [q] Quit                      │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

## Selected Card State
```
  ╭══════════════════╮  ╭──────────────────╮  ╭──────────────────╮
  ║        👻        ║  │        🖥️        │  │        🪟        │
  ║                  ║  │                  │  │                  │
  ║     Ghostty      ║  │    Alacritty     │  │     WezTerm      │
  ║     Terminal     ║  │     Terminal     │  │     Terminal     │
  ║                  ║  │                  │  │                  │
  ║   ✓ Installed    ║  │   ✓ Installed    │  │  ✗ Not Installed │
  ║   ⚙ Configured   ║  │   📄 Has Config  │  │                  │
  ╰══════════════════╯  ╰──────────────────╯  ╰──────────────────╯
        ^
    SELECTED (thick border, highlighted)
```

## Dimmed State (Not Installed)
```
  ╭──────────────────╮  ╭──────────────────╮  ╭┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈╮
  │        👻        │  │        🖥️        │  ┊        🪟        ┊
  │                  │  │                  │  ┊                  ┊
  │     Ghostty      │  │    Alacritty     │  ┊     WezTerm      ┊
  │     Terminal     │  │     Terminal     │  ┊     Terminal     ┊
  │                  │  │                  │  ┊                  ┊
  │   ✓ Installed    │  │   ✓ Installed    │  ┊  ✗ Not Installed ┊
  │   ⚙ Configured   │  │   📄 Has Config  │  ┊                  ┊
  ╰──────────────────╯  ╰──────────────────╯  ╰┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈╯
                                                       ^
                                                 DIMMED (gray, dotted border)
```

## Card Status Indicators
```
✓ Installed      - App executable found in PATH
✗ Not Installed  - App executable not found
⚙ Configured     - ZeroUI has config for this app
📄 Has Config    - App's native config file exists
```

## Filtered View (Show Only Installed)
```
┌────────────────────────────────────────────────────────────────────────────┐
│                                                                            │
│  🚀 ZeroUI Applications                                                   │
│  Select an application to configure (6 apps)                              │
│                                                                            │
│  ╭──────────────────╮  ╭──────────────────╮  ╭──────────────────╮       │
│  │        👻        │  │        🖥️        │  │        📝        │       │
│  │                  │  │                  │  │                  │       │
│  │     Ghostty      │  │    Alacritty     │  │     VS Code      │       │
│  │     Terminal     │  │     Terminal     │  │      Editor      │       │
│  │                  │  │                  │  │                  │       │
│  │   ✓ Installed    │  │   ✓ Installed    │  │   ✓ Installed    │       │
│  │   ⚙ Configured   │  │   📄 Has Config  │  │   ⚙ Configured   │       │
│  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯       │
│                                                                            │
│  ╭──────────────────╮  ╭──────────────────╮  ╭──────────────────╮       │
│  │        📜        │  │        🔲        │  │        🌳        │       │
│  │                  │  │                  │  │                  │       │
│  │      Neovim      │  │       Tmux       │  │        Git       │       │
│  │      Editor      │  │   Multiplexer    │  │   Development    │       │
│  │                  │  │                  │  │                  │       │
│  │   ✓ Installed    │  │   ✓ Installed    │  │   ✓ Installed    │       │
│  │   📄 Has Config  │  │                  │  │   ⚙ Configured   │       │
│  ╰──────────────────╯  ╰──────────────────╯  ╰──────────────────╯       │
│                                                                            │
│  ↑↓←→ Navigate • ⏎ Select • [a] Show All • [q] Quit                      │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```

## Flow: Selecting Unconfigured but Installed App
```
1. User selects Alacritty (installed but no ZeroUI config)
   ╭══════════════════╮
   ║        🖥️        ║
   ║    Alacritty     ║  <-- User presses Enter
   ║   ✓ Installed    ║
   ║   📄 Has Config  ║
   ╰══════════════════╯

2. Creates ZeroUI config automatically
   ┌─────────────────────────────────────┐
   │  Creating configuration...          │
   │  ~/.config/zeroui/apps/alacritty.yaml │
   └─────────────────────────────────────┘

3. Opens config editor
   ┌────────────────────────────────────────────────┐
   │  Alacritty Configuration                      │
   │  ────────────────────────                     │
   │  Theme        [dark] ▼                        │
   │  Font Size    [14]                            │
   │  Font Family  [JetBrains Mono]                │
   │  ...                                           │
   └────────────────────────────────────────────────┘
```

## Responsive Layout (Narrow Terminal)
```
When width < 80 chars:
┌──────────────────────────┐
│  🚀 ZeroUI Applications  │
│                          │
│  ╭──────────────────╮   │
│  │        👻        │   │
│  │     Ghostty      │   │
│  │   ✓ Installed    │   │
│  ╰──────────────────╯   │
│                          │
│  ╭──────────────────╮   │
│  │        🖥️        │   │
│  │    Alacritty     │   │
│  │   ✓ Installed    │   │
│  ╰──────────────────╯   │
│                          │
│  ↑↓ Nav • ⏎ Select      │
└──────────────────────────┘
```

## Empty State (No Apps Found)
```
┌────────────────────────────────────────────────────────────────────────────┐
│                                                                            │
│  🚀 ZeroUI Applications                                                   │
│                                                                            │
│                                                                            │
│                                                                            │
│                       No applications found                               │
│                                                                            │
│                   Check your installation or filters                      │
│                                                                            │
│                                                                            │
│  [a] Show All • [q] Quit                                                  │
│                                                                            │
└────────────────────────────────────────────────────────────────────────────┘
```