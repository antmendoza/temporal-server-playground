---
name: feedback-scenario-design
description: Design principle for every degradation scenario in this playground — one dynamic-config knob, one component, one observable signal
metadata:
  type: feedback
---

Every scenario added to this playground must follow the **"one knob, one component, one signal"** rule:

- **One knob** — the scenario YAML in `config/dynamicconfig/scenarios/<name>.yaml` flips a single dynamic-config setting (or one knob plus a startup-only setting like `NUM_HISTORY_SHARDS`). Avoid bundling multiple unrelated knobs into the same scenario.
- **One component** — the scenario isolates a single Temporal component: frontend, history engine, matching, visibility, persistence, or shard layer. If two components are implicated, split into two scenarios.
- **One signal** — the failure mode has a clear, named primary signal in Grafana (e.g. `service_errors_resource_exhausted`, `approximate_backlog_count`, `cache_miss_rate`). A dedicated dashboard per scenario surfaces that signal first, supporting panels second.

**Why:** keeps each scenario unambiguous as a teaching/demo artifact. If a scenario triggers several symptoms at once, the demo is confusing and the user can't tell which knob caused what.

**How to apply:** when designing or implementing a new scenario, name the knob, name the component, and name the signal up-front in the scenario YAML header comment. If you can't, the scope is too broad — split it. The README "Scenario catalog" table is the canonical inventory of these triples (Scenario / Knob / What you'd see).

Confirmed in conversation on 2026-06-13 when proposing the 6-scenario catalog. See [[reference-temporal-docs]] (performance-bottlenecks page) and [[reference-temporal-repos]] (tsurdilo's troubleshooting patterns) for the runbooks each scenario maps to.
