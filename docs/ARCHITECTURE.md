# Architecture Overview — ZeroUI

ZeroUI follows clean architecture principles with separation of concerns.

Primary layout

```
cmd/                    # CLI commands and entry points
internal/               # Application internals (config, service, tui, etc.)
pkg/                    # Reusable packages
tools/                  # Development tools
.github/workflows/      # CI/CD pipelines
```

Key components

- `internal/config` - Configuration management
- `internal/tui` - Terminal UI components and views
  - `internal/tui/styles` - Design system and themes
  - `internal/tui/animations` - Animation and transition effects
  - `internal/tui/feedback` - Notification and loading systems
  - `internal/tui/help` - Contextual help system
- `internal/toggle` - Core toggle operations
- `internal/service` - Business logic
- `internal/observability` - Metrics and tracing

## 🎉 Delightful UX Architecture

The enhanced UX system is built with a modular, layered architecture:

### Core UX Components

```
internal/tui/
├── styles/theme.go           # 🎨 Design system & themes
├── animations/effects.go     # ⚡ Animation engine
├── feedback/
│   ├── notifications.go      # 🔔 Notification system
│   └── loading.go           # ⏳ Loading states
├── help/contextual.go       # ❓ Contextual help
└── components/forms/
    └── enhanced_config.go   # ⚙️ Enhanced form UX
```

### Architecture Layers

1. **Presentation Layer**: Visual design and styling
2. **Interaction Layer**: Animation and transition effects
3. **Feedback Layer**: Notifications, loading states, and user guidance
4. **Context Layer**: Intelligent state awareness and adaptation
5. **Performance Layer**: Optimized rendering and responsive interactions

### Key Design Principles

- **Modular**: Each UX component is independently testable and maintainable
- **Accessible**: All features support screen readers and keyboard navigation
- **Performant**: Optimized for 60fps interactions and smooth animations
- **Adaptive**: Context-aware behavior that learns from user patterns
- **Beautiful**: Modern design system with consistent visual hierarchy

For more details see docs/README.md and the module code.
