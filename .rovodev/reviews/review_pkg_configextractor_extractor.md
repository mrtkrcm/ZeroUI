# Review: pkg/configextractor/extractor.go

## Actionable Comments

- PRIORITY [LOW] Dead code: `bestConfig` variable is never set before being checked. Remove for clarity.
- PRIORITY [LOW] Ineffective loop: `SupportedApps` iterates strategies without using them; then returns fallback list. Either implement capability discovery or remove the loop.
- PRIORITY [LOW] Strategy coverage: `NewLocal()` and `NewBuiltin()` are commented out. Consider re-enabling them (with lower priority) to improve resiliency when network/CLI are unavailable.

## Code Suggestions (Unified Diff)

Remove dead `bestConfig` and simplify:

```diff
--- a/pkg/configextractor/extractor.go
+++ b/pkg/configextractor/extractor.go
@@
-    // Collect results, return best one
-    var bestConfig *Config
-    var lastErr error
+    // Collect results
+    var lastErr error
@@
-    if bestConfig != nil {
-        e.cache.Set(cacheKey, bestConfig)
-        return bestConfig, nil
-    }
-    
     return nil, fmt.Errorf("all extraction strategies failed for %s: %w", app, lastErr)
```

Optionally implement strategy-based SupportedApps or remove loop placeholder:

```diff
--- a/pkg/configextractor/extractor.go
+++ b/pkg/configextractor/extractor.go
@@
-    // Fallback: collect from strategies
-    appSet := make(map[string]bool)
-    for _ = range e.strategies {
-        // This would need to be enhanced based on strategy implementation
-        // For now, return common apps
-    }
-    
-    apps := []string{"ghostty", "zed", "alacritty", "wezterm", "tmux", "git", "neovim"}
-    result := make([]string, 0, len(apps))
-    for app := range appSet {
-        result = append(result, app)
-    }
-    if len(result) == 0 {
-        return apps // fallback
-    }
-    
-    sort.Strings(result)
-    return result
+    // Fallback list (until strategies expose discovery)
+    return []string{"ghostty", "zed", "alacritty", "wezterm", "tmux", "git", "neovim"}
```

## Priority Order to Fix

1) LOW: Remove dead code and clarify SupportedApps behavior.
2) LOW: Consider re-enabling fallback strategies for resilience.
