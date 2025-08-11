# Review: internal/validation/validator.go

## Actionable Comments

- PRIORITY [HIGH] Syntax error: Backslashes in operators in `optimizeSchema` free function (`rule.Pattern \!= ""` and `err \!= nil`). Remove backslashes; these prevent compilation.
- PRIORITY [HIGH] Logic bug: `v.optimizeSchema` iterates with `for _, rule := range ...` and doesn’t write back the updated `rule` to `schema.Fields`. The precomputed `enumMap` and `compiledRegex` aren’t persisted. Iterate with `for fieldName, rule := range ...` and assign `schema.Fields[fieldName] = rule`.
- PRIORITY [MEDIUM] Performance note: Good preallocation in `validateBasic`. Consider profiling other hot paths for similar preallocation if needed.

## Code Suggestions (Unified Diff)

Persist modifications in `v.optimizeSchema` and remove stray backslashes in the free function:

```diff
--- a/internal/validation/validator.go
+++ b/internal/validation/validator.go
@@
-func (v *Validator) optimizeSchema(schema *Schema) {
-    for _, rule := range schema.Fields {
+func (v *Validator) optimizeSchema(schema *Schema) {
+    for fieldName, rule := range schema.Fields {
         // Pre-build enum map for O(1) lookups
         if len(rule.Enum) > 0 {
             rule.enumMap = make(map[string]struct{}, len(rule.Enum))
             for _, enum := range rule.Enum {
                 rule.enumMap[enum] = struct{}{}
             }
         }
         
         // Pre-compile regex patterns
         if rule.Pattern != "" {
             if compiled, err := regexp.Compile(rule.Pattern); err == nil {
                 rule.compiledRegex = compiled
             }
         }
+        // Persist modifications
+        schema.Fields[fieldName] = rule
     }
 }
@@
-func optimizeSchema(schema *Schema) error {
+func optimizeSchema(schema *Schema) error {
     for fieldName, rule := range schema.Fields {
         // Pre-build enum map for O(1) lookups
         if len(rule.Enum) > 0 {
             rule.enumMap = make(map[string]struct{}, len(rule.Enum))
             for _, val := range rule.Enum {
                 rule.enumMap[val] = struct{}{}
             }
         }
         
         // Pre-compile regex patterns
-        if rule.Pattern \!= "" {
-            compiled, err := regexp.Compile(rule.Pattern)
-            if err \!= nil {
+        if rule.Pattern != "" {
+            compiled, err := regexp.Compile(rule.Pattern)
+            if err != nil {
                 return fmt.Errorf("invalid regex pattern for field %s: %w", fieldName, err)
             }
             rule.compiledRegex = compiled
         }
         
         // Update the rule in the schema
         schema.Fields[fieldName] = rule
     }
     
     return nil
 }
```

## Priority Order to Fix

1) HIGH: Remove backslashes in operators to restore compilation.
2) HIGH: Persist rule updates in `v.optimizeSchema` for actual performance benefit.
3) MEDIUM: Optional profiling-guided preallocation improvements in other paths.
