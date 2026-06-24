---
name: feedback-dashboard-organization
description: Group Grafana panels by perspective (SDK worker vs server-side), not by symptom — metric prefix tells you which section a panel belongs in
metadata:
  type: feedback
---

In every scenario dashboard under `grafana/provisioning/dashboards/`, panels must be grouped by **perspective**, not by symptom:

- **"SDK Client & Worker Perspective"** — panels whose PromQL queries metrics emitted by the worker process via the SDK. These have the `temporal_*` prefix and come from the SDK metrics handler (see [[reference-temporal-docs]] → `references/sdk-metrics`). Examples:
  - `temporal_worker_task_slots_available`
  - `temporal_workflow_task_schedule_to_start_latency_seconds`
  - `temporal_activity_schedule_to_start_latency_seconds`
  - `temporal_long_request_*` (poll-operation failure / rate)
  - `temporal_request_failure` (by status code)
- **Server-side sections** (e.g. "Rate Limiter — Primary Signals", "Frontend", "History", "Persistence") — panels whose queries hit metrics emitted by the Temporal server (see [[reference-temporal-docs]] → `references/cluster-metrics`). These typically use prefixes like `service_*`, `persistence_*`, `frontend_*`, `history_*`, `matching_*`, `service_errors_*`.

**Why:** the SDK-vs-server distinction is easy to miss because some metric names (`schedule_to_start_latency`) sound server-ish but are actually emitted by the worker. Mis-grouping makes the dashboard misleading — e.g. it implies the server is reporting a problem when really the worker is. Confirmed on 2026-06-13 when `Schedule-to-Start Latency — Workflow Tasks`, `Schedule-to-Start Latency — Activities`, and `Worker Task Slots Available` were moved out of "Task Queue Delivery Impact" into "SDK Client & Worker Perspective" because all three query `temporal_*` SDK metrics.

**How to apply:** before placing a new panel, check the metric prefix in its PromQL query. `temporal_*` → SDK section. Anything else → the matching server-side section. When adding a new dashboard from scratch, start with the SDK section (worker-visible symptoms) plus one server-side section (root-cause signals from the affected component).
