# Review Index

This index summarizes the actionable comments and priority order across modified files.

## High Priority

1) pkg/reference/fast_extractor.go
   - Fix buffer pool type mismatch to prevent panic.
   - See: .rovodev/reviews/review_pkg_reference_fast_extractor.md

2) internal/validation/validator.go
   - Remove stray backslashes in operators (compile error).
   - Persist rule updates in `v.optimizeSchema` so precomputed maps/regex are effective.
   - See: .rovodev/reviews/review_internal_validation_validator.md

3) internal/tui/components/app_grid.go
   - Add missing `min` helper to restore build.
   - See: .rovodev/reviews/review_internal_tui_app_grid.md

## Medium Priority

4) internal/tui/app.go
   - Only schedule animation ticks in animating view(s).
   - See: .rovodev/reviews/review_internal_tui_app.md

5) internal/config/loader.go
   - Guard division by zero in `GetCacheStats`.
   - See: .rovodev/reviews/review_internal_config_loader.md

## Low Priority

6) internal/config/custom_parser.go
   - Handle Stat error and set a fallback scanner buffer.
   - See: .rovodev/reviews/review_internal_config_custom_parser.md

7) pkg/configextractor/extractor.go
   - Remove dead code, clarify SupportedApps, consider re-enabling fallback strategies.
   - See: .rovodev/reviews/review_pkg_configextractor_extractor.md

Notes:
- Other changes (snapshots, docs deletions, tests) appear consistent with layout/UX refresh and benchmark simplifications. Revisit docs (README) to ensure removed CLI commands are no longer referenced.
