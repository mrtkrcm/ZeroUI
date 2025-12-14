# Advanced optimization report (internal)

If you are working on performance changes:

- Start with `make benchmark` and add/adjust a benchmark that covers your change.
- Use `make profile` to inspect CPU hotspots.
- Prefer small, targeted changes and verify they do not regress correctness or UX.

Related code and notes:

- `internal/performance/`
- `internal/performance/optimization_guide.md`
