# TUI Layout Improvement Plan

## Current Layout Issues
1. **Help text overflow** on narrow terminals
2. **Massive vertical waste** - 20+ empty lines
3. **No visual hierarchy** - poor information organization
4. **No status feedback** - missing error/info display
5. **Poor responsiveness** - doesn't adapt to different screen sizes

## Proposed Enhanced Layout

### Option A: Compact Vertical Layout (Recommended)
```
┌─ ZeroUI - Select Application ──────────────────────────────────────────┐
│                                                                        │
│  Available Applications:                                               │
│  > ghostty                   [recently used]                          │
│    alacritty                [configured]                              │
│    wezterm                  [needs setup]                             │
│                                                                        │
│  Status: Ready • Apps: 3 • Theme: Default                            │
│                                                                        │
│  Navigation: ↑↓ select • Enter choose • ? help • q quit               │
└────────────────────────────────────────────────────────────────────────┘
```

### Option B: Multi-Panel Layout
```
┌─ ZeroUI ───────────────────┬─ Quick Info ───────────────────────────────┐
│                            │                                            │
│  Applications:             │  Selected: ghostty                        │
│  > ghostty                 │  Config: ~/.config/ghostty/config         │
│    alacritty               │  Status: ✓ Valid                          │
│    wezterm                 │  Fields: 12 configurable                  │
│                            │  Last modified: 2 hours ago               │
│                            │                                            │
│                            │  Available actions:                        │
│                            │  • e - Edit configuration                  │
│                            │  • p - Apply preset                       │
│                            │  • b - Create backup                      │
│                            │                                            │
├────────────────────────────┴────────────────────────────────────────────┤
│ Status: Ready               Help: ? toggle • Navigation: ↑↓ • q quit   │
└─────────────────────────────────────────────────────────────────────────┘
```

### Option C: Horizontal Split with Context
```
┌─ ZeroUI - Configuration Manager ─────────────────────────────────────────┐
│                                                                          │
│  ┌─ Applications ─────────┐  ┌─ Configuration Preview ─────────────────┐ │
│  │                        │  │                                         │ │
│  │  > ghostty             │  │  theme = "default"                      │ │
│  │    alacritty           │  │  font-family = "JetBrains Mono"        │ │
│  │    wezterm             │  │  font-size = 14                        │ │
│  │                        │  │  background-opacity = 0.95             │ │
│  │  [3 apps available]    │  │  cursor-style = "bar"                  │ │
│  │                        │  │  ...                                    │ │
│  │                        │  │                                         │ │
│  └────────────────────────┘  └─────────────────────────────────────────┘ │
│                                                                          │
│  Status: ghostty loaded • 12 fields • Press Enter to edit              │
│  Help: ↑↓ navigate • Enter edit • p presets • ? help • q quit          │
└──────────────────────────────────────────────────────────────────────────┘
```

## Responsive Design Strategy

### Small Terminals (< 60 cols)
- Single column layout
- Abbreviated help text
- Essential info only
- Scrollable content

### Medium Terminals (60-100 cols)  
- Compact two-column where beneficial
- Full help text
- Status line with key info

### Large Terminals (> 100 cols)
- Multi-panel layouts
- Rich context information
- Preview panels
- Extended help sidebar

## Implementation Plan

### Phase 1: Fix Current Issues
1. **Responsive Help Text**: Truncate help on narrow screens
2. **Better Space Usage**: Remove excessive padding
3. **Add Status Bar**: Show context info between content and help
4. **Improve Content Spacing**: Better vertical rhythm

### Phase 2: Enhanced Layout Components
1. **Info Panel Component**: Show selected app details
2. **Status Component**: Rich status with icons and colors
3. **Preview Component**: Show config snippets
4. **Breadcrumb Component**: Navigation context

### Phase 3: Advanced Features
1. **Split View Mode**: Side-by-side panels
2. **Tab System**: Multiple views (Apps, Configs, Presets, Help)
3. **Modal Overlays**: Context menus and dialogs
4. **Dynamic Theming**: Responsive color schemes

## Technical Requirements

### Responsive Breakpoints
- **Narrow**: < 60 columns (mobile/minimal)
- **Standard**: 60-100 columns (typical terminal)  
- **Wide**: 100-140 columns (modern terminal)
- **Ultra**: > 140 columns (ultra-wide/split screen)

### Layout Components Needed
- `HeaderComponent`: Title, app context, navigation
- `ContentAreaComponent`: Main interactive content
- `SidebarComponent`: Contextual information
- `StatusBarComponent`: Status, progress, notifications
- `HelpComponent`: Context-aware help text
- `ModalComponent`: Overlays and dialogs

### Performance Considerations
- Lazy render off-screen content
- Efficient re-layout on resize
- Minimal re-renders on state changes
- Smart content truncation
- Cached layout calculations

## User Experience Goals

1. **Immediate Clarity**: User knows exactly what they can do
2. **Efficient Navigation**: Minimal keystrokes to accomplish tasks
3. **Rich Context**: Always show relevant information
4. **Responsive**: Works beautifully on any terminal size
5. **Discoverable**: Features are easy to find and learn
6. **Professional**: Clean, modern, polished appearance