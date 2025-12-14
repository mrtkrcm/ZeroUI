# ZeroUI Design System Showcase

This showcases the native design system components, styles, and patterns used in the ZeroUI terminal application.

## How to Access

```bash
# Interactive mode (default)
zeroui design-system
zeroui showcase
zeroui ds
zeroui demo

# Non-interactive mode (static output)
zeroui design-system --interactive=false
```

## What's Included

The design system showcase demonstrates:

### 🎨 **Color Palette & Themes**
- Primary colors (Purple #7D56F4)
- Secondary colors (Pink #FF6B9D, Gold #C9A96E)
- Status colors (Success, Warning, Error, Info)
- Neutral colors and backgrounds

### 📝 **Typography & Text Styles**
- Title hierarchy (H1, H2, H3)
- Body text and captions
- Code styling (inline and blocks)
- Text emphasis (bold, italic, underline)

### 🧩 **UI Components**
- **Lists**: Interactive selection lists with navigation
- **Buttons**: Primary, secondary, and disabled states
- **Input Fields**: Text input with focus states
- **Progress Bars**: Animated progress indicators
- **Spinners**: Loading animations
- **Viewports**: Scrollable content areas

### 📐 **Layout Patterns**
- Spacing system (xs, sm, md, lg, xl)
- Alignment options (left, center, right)
- Container styles with padding and borders
- Multi-column layouts

### ⚡ **Interactive Elements**
- Live text input (try typing!)
- Key binding demonstrations
- Mouse support indicators
- Navigation patterns

### 🎬 **Animations & Loading States**
- Spinning loaders with various styles
- Animated progress bars
- Blinking cursor effects
- Smooth transitions

### ❌ **Error States & Feedback**
- Error messages with proper styling
- Warning notifications
- Success confirmations
- Validation states (valid/invalid/pending)

### 📦 **Box Drawing & Borders**
- Normal borders
- Rounded corners
- Thick borders
- Double-line borders
- Complex nested layouts

### 🚀 **Real ZeroUI Examples**
- Actual application selection interface
- Configuration editing screens
- Real styling from the live application

## Technical Implementation

Built using:
- **Bubble Tea**: TUI framework for Go
- **Bubbles**: Pre-built UI components
- **Lipgloss**: Styling and layout engine
- **Native terminal rendering**: Real terminal output, not web simulation

All components shown are the actual implementations used throughout the ZeroUI application, providing an authentic representation of the design system as it appears in the terminal.

## Navigation

When running in interactive mode:
- **Tab/Shift+Tab**: Navigate between sections  
- **1-9**: Jump directly to specific sections
- **Enter**: Select highlighted section
- **Arrow keys**: Navigate within sections
- **Q**: Quit the showcase
- **Esc**: Go back to previous view

The showcase provides a comprehensive view of the ZeroUI design language, demonstrating consistency across all UI components and interactions in the terminal application.