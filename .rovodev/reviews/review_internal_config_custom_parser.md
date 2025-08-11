# Review: internal/config/custom_parser.go

## Actionable Comments

- PRIORITY [LOW] Robustness: `file.Stat()` error is ignored. If it fails, scanner stays at default and large lines may cause `bufio.ErrTooLong`. Consider handling errors and setting a safe buffer size manually.

## Code Suggestions (Unified Diff)

```diff
--- a/internal/config/custom_parser.go
+++ b/internal/config/custom_parser.go
@@
-    fileInfo, _ := file.Stat()
-    scanner := bufio.NewScanner(file)
-    
-    if fileInfo != nil {
+    fileInfo, statErr := file.Stat()
+    scanner := bufio.NewScanner(file)
+    
+    if statErr == nil && fileInfo != nil {
         // Use adaptive buffer sizing: quarter of file size, max 64KB, min 4KB
         bufSize := int(fileInfo.Size() / 4)
         if bufSize > 64*1024 {
             bufSize = 64 * 1024
         } else if bufSize < 4*1024 {
             bufSize = 4 * 1024
         }
         scanner.Buffer(make([]byte, 0, bufSize), bufSize)
+    } else {
+        // Fallback to a reasonable maximum line size to avoid ErrTooLong
+        const maxLine = 64 * 1024
+        scanner.Buffer(make([]byte, 0, maxLine), maxLine)
     }
```

## Priority Order to Fix

1) LOW: Handle `Stat` error and ensure scanner buffer is adequate for long lines.
