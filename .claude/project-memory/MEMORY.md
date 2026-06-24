# Project Memory

> When a new decision **contradicts** an existing
> memory note, do NOT silently override it.
> Instead: surface the conflict, quote the
> existing memory, explain how the new decision
> differs, and ask for explicit confirmation
> before updating. **Do NOT take any action** —
> no tool calls, no file writes — until confirmed.


- [Temporal reference repos](reference_temporal_repos.md) — external GitHub repos (tsurdilo/temporal-server-operations, temporalio/dashboards, temporalio/temporal) to consult for Temporal server context
- [Temporal reference docs](reference_temporal_docs.md) — official docs.temporal.io pages for SDK/cluster metrics and troubleshooting (performance, blob size, deadline exceeded, last connection error)
- [Scenario design principle](feedback_scenario_design.md) — every scenario isolates one component with one dynamic-config knob and one named primary Grafana signal
- [Dashboard organization](feedback_dashboard_organization.md) — group Grafana panels by perspective (SDK worker `temporal_*` metrics vs server-side); never mix in the same section
