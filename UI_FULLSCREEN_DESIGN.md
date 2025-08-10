# ZeroUI Full-Screen Centered Grid Design

## Full-Screen Layout with Large Margins
```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                                                                        │
│                                                                                        │
│                                                                                        │
│                                                                                        │
│                                         ZeroUI                                         │
│                                   9 applications available                            │
│                                                                                        │
│                                                                                        │
│                                                                                        │
│           ╔════════════════════════════╗    ╔════════════════════════════╗           │
│           ║                            ║    ║                            ║           │
│           ║                            ║    ║                            ║           │
│           ║            👻              ║    ║            🖥️              ║           │
│           ║                            ║    ║                            ║           │
│           ║          Ghostty           ║    ║         Alacritty          ║           │
│           ║          Terminal          ║    ║          Terminal          ║           │
│           ║                            ║    ║                            ║           │
│           ║                            ║    ║                            ║           │
│           ║       ✓ Installed          ║    ║       ✓ Installed          ║           │
│           ║       ⚙ Configured         ║    ║       📄 Has Config        ║           │
│           ║                            ║    ║                            ║           │
│           ╚════════════════════════════╝    ╚════════════════════════════╝           │
│                                                                                        │
│           ╔════════════════════════════╗    ╔════════════════════════════╗           │
│           ║                            ║    ║                            ║           │
│           ║                            ║    ║                            ║           │
│           ║            🪟              ║    ║            📝              ║           │
│           ║                            ║    ║                            ║           │
│           ║          WezTerm           ║    ║          VS Code           ║           │
│           ║          Terminal          ║    ║           Editor           ║           │
│           ║                            ║    ║                            ║           │
│           ║                            ║    ║                            ║           │
│           ║       ✗ Not Installed      ║    ║       ✓ Installed          ║           │
│           ║                            ║    ║       ⚙ Configured         ║           │
│           ║                            ║    ║                            ║           │
│           ╚════════════════════════════╝    ╚════════════════════════════╝           │
│                                                                                        │
│           ╔════════════════════════════╗    ╔════════════════════════════╗           │
│           ║                            ║    ║                            ║           │
│           ║                            ║    ║                            ║           │
│           ║            📜              ║    ║            ⚡              ║           │
│           ║                            ║    ║                            ║           │
│           ║          Neovim            ║    ║            Zed             ║           │
│           ║           Editor           ║    ║           Editor           ║           │
│           ║                            ║    ║                            ║           │
│           ║                            ║    ║                            ║           │
│           ║       ✓ Installed          ║    ║       ✗ Not Installed      ║           │
│           ║       📄 Has Config        ║    ║                            ║           │
│           ║                            ║    ║                            ║           │
│           ╚════════════════════════════╝    ╚════════════════════════════╝           │
│                                                                                        │
│                                                                                        │
│                                                                                        │
│                   ↑↓←→ Navigate  •  ⏎ Select  •  a Show All  •  q Quit               │
│                                                                                        │
│                                                                                        │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

## Key Design Features

### 1. Full-Screen Coverage
- Occupies entire terminal window
- No wrapped layouts or borders
- Clean, minimal interface

### 2. Centered Content
- Cards centered both horizontally and vertically
- Large margins on all sides for breathing room
- Elegant spacing between cards

### 3. Card Design (28x12 characters)
```
╔════════════════════════════╗
║                            ║  
║                            ║  
║            👻              ║  <- Large emoji logo
║                            ║  
║          Ghostty           ║  <- App name (bold)
║          Terminal          ║  <- Category (subtle)
║                            ║  
║                            ║  
║       ✓ Installed          ║  <- Status indicators
║       ⚙ Configured         ║  
║                            ║  
╚════════════════════════════╝
```

### 4. Visual States

#### Selected Card (Double Border + Background)
```
╔════════════════════════════╗
║░░░░░░░░░░░░░░░░░░░░░░░░░░░░║
║░░░░░░░░░░░👻░░░░░░░░░░░░░░░║
║░░░░░░░░░Ghostty░░░░░░░░░░░░║
║░░░░░░░░Terminal░░░░░░░░░░░░║
║░░░░░░░✓ Installed░░░░░░░░░░║
║░░░░░░░⚙ Configured░░░░░░░░░║
╚════════════════════════════╝
```

#### Dimmed Card (Not Installed)
```
┌┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┐
┊                            ┊
┊            🪟              ┊  <- Grayed out
┊          WezTerm           ┊
┊          Terminal          ┊
┊       ✗ Not Installed      ┊
┊                            ┊
└┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┘
```

### 5. Responsive Grid Layout
- **Wide terminals (>120 chars)**: 3 columns
- **Medium terminals (80-120 chars)**: 2 columns  
- **Narrow terminals (<80 chars)**: 1 column

### 6. Color Scheme
- **Selected**: Bright cyan border (#212) with subtle background
- **Normal**: Medium gray border (#244)
- **Dimmed**: Dark gray (#238) for uninstalled apps
- **Text**: High contrast white on dark
- **Status**: Green (#70) for installed, gray for not installed

### 7. Navigation
- **Arrow keys**: Move selection between cards
- **Enter**: Select and configure app
- **a**: Toggle between all apps and installed only
- **q**: Quit application
- **ESC**: Back to previous view

## Implementation Details

### Component Structure
```
AppGrid (Full Screen Container)
├── Header (Centered)
│   ├── Title: "ZeroUI"
│   └── Subtitle: "9 applications available"
├── Grid (Centered with large margins)
│   ├── Row 1
│   │   ├── AppCard (Ghostty)
│   │   ├── Spacing (4 chars)
│   │   └── AppCard (Alacritty)
│   ├── Row Spacing (1 line)
│   └── Row 2
│       ├── AppCard (WezTerm)
│       ├── Spacing (4 chars)
│       └── AppCard (VS Code)
└── Footer (Centered)
    └── Navigation hints
```

### Spacing Formula
- **Horizontal margin**: `(terminal_width - (columns * card_width + (columns-1) * spacing)) / 2`
- **Vertical margin**: `(terminal_height - total_content_height) / 2`
- **Card spacing**: 4 characters horizontal, 1 line vertical
- **Minimum margin**: 10 characters on each side

This design provides a modern, spacious interface that feels premium and easy to navigate.