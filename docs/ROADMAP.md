# Roadmap

This roadmap tracks user-facing direction at a high level. Engineering-focused work lives in `improvements-roadmap.md`.

## Near-term

- Stabilize core workflows: `toggle`, `cycle`, `preset`, `backup`, and `ui`.
- Improve app coverage by expanding curated reference metadata (`configs/`) and plugins.
- Tighten CLI/TUI UX (help, search/filter, previews where applicable).

### Raycast Integration Improvements

#### Phase 1: Critical Fixes ✅ COMPLETED (Dec 2025)

- **Fix CLI Command Usage**: Added `list values <app>` and `list changed <app>` commands with proper output parsing for key-value extraction.
- **Enable Full Feature Set**: Implemented complete `keymap` command suite (`list`, `add`, `remove`, `edit`, `validate`, `presets`, `conflicts`) instead of disabling features.
- **Fix Version Check**: Added `zeroui version` command and updated all references (Dockerfile, scripts, extension) to replace non-existent `--version` flag.
- **Improve Binary Discovery**: Enhanced path resolution with 20+ fallback locations, cross-platform support (macOS/Linux/Windows), and detailed error messages with platform/architecture info.
- **Error Handling**: Implemented retry logic, graceful degradation, comprehensive logging, and fallback parsing for different ZeroUI output formats.

#### Phase 2: React Hooks & Performance (Medium Priority, 1-2 weeks)

- **Migrate to Raycast Utils**: Replace custom async logic with `useCachedPromise`, `useExec`, and `usePromise` hooks for better performance and reliability.
- **Smart Caching**: Implement stale-while-revalidate strategy with `useCachedPromise` for menu bar and background refresh.
- **Background Refresh**: Add menu bar command with configurable intervals for real-time status updates.
- **Cross-Command Communication**: Enable deeplinks and `launchCommand()` for seamless navigation between list/detail views.
- **Loading States**: Add proper loading indicators, skeleton screens, and optimistic updates.
- **Error Handling**: Implement proper toast notifications with retry actions and graceful degradation.
- **Form Integration**: Use `useForm` hook for configuration editing with validation and error states.

#### Phase 3: Advanced Features & AI Integration (Low Priority, 2-4 weeks)

- **AI-Powered Features**: Integrate Raycast AI tools for configuration suggestions, validation, and natural language queries.
- **Enhanced Integrations**: Support ZeroUI's preset system, backup management UI, theme switching, and bulk operations.
- **Developer Tools**: Add performance monitoring dashboard, extension health checks, detailed logging, and debugging tools.
- **Form Improvements**: Implement `useForm` hook for configuration editing with validation and error handling.
- **Pagination Support**: Add pagination for large configuration lists using Raycast's pagination API.
- **Local Storage**: Use `useLocalStorage` for persisting user preferences and recent actions.
- **Accessibility**: Ensure all components follow Raycast accessibility guidelines with proper ARIA labels and keyboard navigation.

#### Phase 4: Testing & Documentation (Ongoing)

- **Comprehensive Testing**: Unit tests for utilities, integration tests with mock ZeroUI, E2E testing with real ZeroUI, and visual regression tests.
- **Documentation Updates**: Update README with correct command usage, add troubleshooting guides, maintain API documentation, and create extension-specific guides.
- **Store Publishing**: Prepare for Raycast Store submission with proper manifest, screenshots, and compliance checks.
- **Debugging Tools**: Implement React DevTools integration, VSCode debugger support, and extension health monitoring.
- **Performance Monitoring**: Add cache hit/miss tracking, command execution timing, and memory usage monitoring.

## Mid-term

- Distribution: binary releases for major platforms and a simple install story.
- Shell integration: completions and better ergonomics.
- Safer config editing: clearer previews, better restore UX, and guardrails.
- Raycast Store: Complete extension publishing with user onboarding, feedback collection, and update management.

## Raycast Extension Best Practices Implementation

### Phase 5: Production Readiness (Medium Priority, 2-3 weeks)

- **Manifest Optimization**: Configure proper command modes (view/no-view/menu-bar), intervals, and argument handling.
- **Preference Management**: Implement user preferences for API keys, cache settings, and custom paths.
- **Cross-Platform Compatibility**: Ensure extension works across different macOS versions and hardware.
- **Security**: Implement secure storage for sensitive data and proper input validation.
- **Internationalization**: Add support for multiple languages if needed.

### Phase 6: Advanced UX Patterns (Low Priority, 3-4 weeks)

- **Custom Components**: Build reusable components for configuration editing, status displays, and action panels.
- **Keyboard Shortcuts**: Implement comprehensive keyboard navigation and shortcuts.
- **Contextual Actions**: Add smart action suggestions based on user behavior and configuration state.
- **Progressive Enhancement**: Ensure core functionality works without advanced features, with enhancements layered on top.
- **Offline Support**: Handle network failures gracefully with cached data and offline indicators.

### Phase 7: Ecosystem Integration (Ongoing)

- **Plugin System**: Support for ZeroUI plugins through the Raycast extension.
- **Third-Party Integrations**: Connect with popular tools (VSCode, JetBrains, etc.) for enhanced workflows.
- **Community Features**: Add sharing, templates, and community-driven presets.
- **Analytics**: Implement usage analytics for improving user experience (privacy-compliant).
- **Feedback Loop**: Build mechanisms for user feedback and feature requests directly in the extension.

## Long-term

- Broader plugin ecosystem (more apps, clearer authoring and packaging story).
- Optional integrations (e.g., launchers) where they improve the daily workflow.
