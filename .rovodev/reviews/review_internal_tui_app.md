# Review: internal/tui/app.go

## Actionable Comments

- PRIORITY [MEDIUM] Performance: `AnimationTickMsg` tick is scheduled even when not in `AppGridView`. If Huh views don’t consume the message, this is wasted work. Only schedule ticks in views that animate (e.g., `AppGridView`).
- PRIORITY [LOW] UX/Help: `renderHelp` computes bindings but calls `m.help.View(&keys.AppKeyMap{})`. If you intend dynamic help content, consider passing computed bindings (depending on your help usage pattern).

## Code Suggestions (Unified Diff)

Limit animation tick scheduling to when animating view is active:

```diff
--- a/internal/tui/app.go
+++ b/internal/tui/app.go
@@
-        // Reduce animation frequency for better performance
-        cmds = append(cmds, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
-            return components.AnimationTickMsg{}
-        }))
+        // Reduce animation frequency for better performance, only when in an animating view
+        if m.state == AppGridView {
+            cmds = append(cmds, tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
+                return components.AnimationTickMsg{}
+            }))
+        }
```

## Priority Order to Fix

1) MEDIUM: Reduce unnecessary tick scheduling.
2) LOW: Consider wiring help bindings for richer contextual help.
