# Baseline

No degradation: the server's defaults, plus two dev-only settings every scenario keeps. Run this
before any scenario so you know what healthy looks like on the dashboards.

```bash
make reset
make load iterations=300 concurrent=30   # completes in ~5s, no RESOURCE_EXHAUSTED
```

`make reset` is also how you return to a clean cluster after any scenario.
