# Review: internal/tui/components/app_grid.go

## Actionable Comments

- PRIORITY [HIGH] Compile error: `min` is used but not defined in this file. Only a test defines `min`. Add a local helper `func min(a, b int) int`.
- PRIORITY [LOW] Minor efficiency/readability: `estimatedSize` includes `leftMargin` which matches the eventual spacer length; it’s fine but ensure it doesn’t drift from actual output if spacing changes.

## Code Suggestions (Unified Diff)

```diff
--- a/internal/tui/components/app_grid.go
+++ b/internal/tui/components/app_grid.go
@@
 func (m *AppGridModel) GetSelectedApp() string {
     if m.selectedIdx >= 0 && m.selectedIdx < len(m.statuses) {
         return m.statuses[m.selectedIdx].Definition.Name
     }
     return ""
 }
+
+// min returns the smaller of two integers
+func min(a, b int) int {
+    if a < b {
+        return a
+    }
+    return b
+}
```

## Priority Order to Fix

1) HIGH: Add `min` helper to restore build.
2) LOW: Keep an eye on spacing calculations with future layout changes.
