# Review: internal/config/loader.go

## Actionable Comments

- PRIORITY [MEDIUM] Bug risk: `GetCacheStats` computes `hit_ratio` with potential division by zero when both hits and misses are zero. Guard against zero denominator.
- PRIORITY [LOW] Unused resources: `bufferPool` and `stringBuilderPool` added but not used. Consider utilizing them or removing to avoid confusion.

## Code Suggestions (Unified Diff)

```diff
--- a/internal/config/loader.go
+++ b/internal/config/loader.go
@@
-    return map[string]interface{}{
-        "cache_hits":   l.cacheHits,
-        "cache_misses": l.cacheMisses,
-        "cache_size":   l.appConfigCache.Len(),
-        "hit_ratio":    float64(l.cacheHits) / float64(l.cacheHits + l.cacheMisses),
-    }
+    total := l.cacheHits + l.cacheMisses
+    var ratio float64
+    if total == 0 {
+        ratio = 0.0
+    } else {
+        ratio = float64(l.cacheHits) / float64(total)
+    }
+    return map[string]interface{}{
+        "cache_hits":   l.cacheHits,
+        "cache_misses": l.cacheMisses,
+        "cache_size":   l.appConfigCache.Len(),
+        "hit_ratio":    ratio,
+    }
```

## Priority Order to Fix

1) MEDIUM: Guard for division by zero in `GetCacheStats`.
2) LOW: Review usage of newly added pools; implement or remove.
