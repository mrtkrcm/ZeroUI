# ZeroUI Design System

A comprehensive design system showcase documenting the visual language and component library used in the ZeroUI project.

## 📁 Files Created

- **`design-system-showcase.html`** - Interactive showcase with live examples
- **`DESIGN_SYSTEM.md`** - This documentation file

## 🎨 Design System Features

### **Design Tokens Extracted**
Based on analysis of the ZeroUI codebase (Lipgloss styles, TUI components):

- **Colors**: Primary purple (#7D56F4), secondary pink, success green, error red
- **Typography**: Monospace for terminal/code, Sans-serif for UI
- **Spacing**: Terminal-friendly rem-based scale (0.25rem - 3rem)
- **Layout**: Card-based design with terminal aesthetics

### **Component Library**

#### **Core Components**
1. **Terminal Title** - Based on `titleStyle` from app.go
2. **Application Lists** - Reflecting TUI app selection interface
3. **Status Messages** - Success/Error/Warning/Info states
4. **Configuration Display** - Key-value pairs with modification indicators
5. **Interactive Buttons** - Terminal-style CTAs with hover states
6. **Loading States** - Progress bars and spinner components

#### **Layout Components**
- Cards with terminal-style borders
- Grid system for component display
- Responsive sections with proper spacing
- Code block styling with syntax highlighting

### **Interactive Features**

#### **Live Demonstrations**
- ✅ **Clickable List Items** - Show selection states
- ✅ **Animated Progress Bars** - Loading state demos
- ✅ **Button Interactions** - Click feedback and hover states
- ✅ **Color Swatches** - Live color previews with usage info

#### **Code Examples**
- Go/Lipgloss source code from actual codebase
- CSS equivalents for web implementation
- Usage guidelines for each component
- Implementation notes and best practices

## 🛠️ Technical Implementation

### **CSS Architecture**
- **CSS Custom Properties** - Complete design token system
- **Responsive Design** - Mobile and desktop compatibility
- **Accessibility** - WCAG AA contrast ratios, keyboard navigation
- **Performance** - Efficient animations, optimized CSS

### **Component Mapping**
Direct translation from ZeroUI's actual components:

```go
// ZeroUI TUI (Go/Lipgloss)
titleStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#7D56F4")).
    Bold(true)

selectedStyle := lipgloss.NewStyle().
    Foreground(lipgloss.Color("#F25D94")).
    Bold(true)
```

```css
/* Design System (CSS)
.terminal-title {
    color: var(--primary-purple);
    font-weight: bold;
}

.selected {
    color: var(--secondary-pink);
    font-weight: bold;
}
```

## 📊 Component Coverage

### **From ZeroUI Codebase Analysis**
- ✅ **TUI App Selection** → Interactive list components
- ✅ **Config Edit View** → Configuration field display
- ✅ **Status Messages** → Semantic status components
- ✅ **Loading States** → Progress and spinner components
- ✅ **Color Scheme** → Complete color palette extraction
- ✅ **Typography** → Terminal and UI font stacks

### **Web-Compatible Extensions**
- ✅ **Button Components** - Various states and types
- ✅ **Form Elements** - Based on Bubbles textinput style
- ✅ **Card Layout** - For component organization
- ✅ **Grid System** - Responsive layout framework

## 🚀 Usage

### **View the Showcase**
Open `design-system-showcase.html` in a web browser to explore:

1. **Design Tokens** - Colors, typography, spacing
2. **UI Components** - All reusable components with examples
3. **Interactive Demos** - Live component interactions
4. **Usage Guidelines** - When and how to use each component
5. **Code Examples** - Implementation details and best practices

### **For Development**
Use this showcase as:
- **Reference** - Component specifications and usage
- **Testing** - Visual regression testing baseline  
- **Documentation** - Living style guide for team use
- **Implementation Guide** - Code examples and patterns

## 🎯 Benefits

### **Design Consistency**
- Unified visual language across CLI and TUI interfaces
- Consistent color usage and typography scales
- Standardized component behaviors and interactions

### **Development Efficiency**
- Pre-built component library with usage examples
- CSS custom properties for easy theming
- Responsive patterns ready for implementation
- Accessibility compliance built-in

### **Maintainability**
- Single source of truth for design decisions
- Easy updates through CSS custom properties
- Component documentation with usage guidelines
- Visual testing reference for quality assurance

---

## 📈 Next Steps

1. **Integration** - Incorporate design system into web documentation
2. **Extension** - Add more components as ZeroUI grows
3. **Theming** - Create light/dark mode variations
4. **Testing** - Use as baseline for visual regression tests

The design system provides a solid foundation for consistent, accessible, and maintainable user interface development across the ZeroUI ecosystem.