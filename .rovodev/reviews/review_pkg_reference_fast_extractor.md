# Review: pkg/reference/fast_extractor.go

## Actionable Comments

- PRIORITY [HIGH] Bug: bufferPool stores the wrong type. `bufferPool.New` currently returns `new(bytes.Buffer)`, but `parseStreamingCLI` does `e.bufferPool.Get().([]byte)`. This will panic due to type assertion mismatch. Fix by pooling `[]byte` slices instead of `*bytes.Buffer`.
- PRIORITY [MEDIUM] Enhancement: `scanner.Buffer` max is hard-coded to 2MB. Consider making this configurable if larger configs are expected.

## Code Suggestions (Unified Diff)

```diff
--- a/pkg/reference/fast_extractor.go
+++ b/pkg/reference/fast_extractor.go
@@
-        workerPool: make(chan struct{}, 8), // Max 8 concurrent operations
-        bufferPool: sync.Pool{
-            New: func() interface{} {
-                return new(bytes.Buffer)
-            },
-        },
+        workerPool: make(chan struct{}, 8), // Max 8 concurrent operations
+        bufferPool: sync.Pool{
+            New: func() interface{} {
+                // Pre-size 64KB buffer for scanner
+                return make([]byte, 0, 64*1024)
+            },
+        },
@@
-    scanner := bufio.NewScanner(r)
-    // Use buffer pool to reduce allocations
-    buf := e.bufferPool.Get().([]byte)
+    scanner := bufio.NewScanner(r)
+    // Use buffer pool to reduce allocations
+    buf := e.bufferPool.Get().([]byte)
     defer func() {
         // Reset buffer before returning to pool
         buf = buf[:0]
         e.bufferPool.Put(buf)
     }()
-    
-    scanner.Buffer(buf[:cap(buf)], 2*1024*1024) // Use pooled buffer, 2MB max for large configs
+    
+    scanner.Buffer(buf[:cap(buf)], 2*1024*1024) // Use pooled buffer, 2MB max for large configs
```

## Priority Order to Fix

1) HIGH: Fix buffer pool type mismatch to prevent runtime panic.
2) MEDIUM: Consider making scanner max buffer configurable.
