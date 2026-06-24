---
name: reference-temporal-docs
description: Authoritative Temporal docs pages for metrics references and troubleshooting — consult these before reasoning about metric names, error symptoms, or performance bottlenecks
metadata:
  type: reference
---

Official Temporal documentation pages flagged as authoritative sources for this project. Prefer these (via WebFetch) over general knowledge when reasoning about metric names, error messages, or performance bottlenecks.

## Metrics references

- **SDK metrics** — https://docs.temporal.io/references/sdk-metrics
  Names, types, labels, and meaning of metrics emitted by the Temporal SDKs (worker pollers, task slots, schedule-to-start latency, etc.). Use when a dashboard panel queries a `temporal_*` metric and you need to know what it actually measures.
- **Cluster (server) metrics** — https://docs.temporal.io/references/cluster-metrics
  Server-emitted metrics (frontend, history, matching, persistence). Use when a panel queries a `*_request_*`, `*_latency_*`, `service_errors_*`, or other server-side metric.

## Troubleshooting

- **Performance bottlenecks** — https://docs.temporal.io/troubleshooting/performance-bottlenecks
  Decision tree for diagnosing slowness across SDK / frontend / history / matching / persistence. Use when planning a new degradation scenario or interpreting load-test results.
- **Blob size limit error** — https://docs.temporal.io/troubleshooting/blob-size-limit-error
  Workflow / activity payload size limits and how the server signals them.
- **Deadline exceeded error** — https://docs.temporal.io/troubleshooting/deadline-exceeded-error
  Where `DEADLINE_EXCEEDED` originates (client / server / network) and which timeouts to check.
- **Last connection error** — https://docs.temporal.io/troubleshooting/last-connection-error
  SDK client connection-failure surface and what `last_connection_error` actually reports.

How to apply: when the user asks about a metric, error, or symptom, WebFetch the relevant page above before guessing. Cross-reference with [[reference-temporal-repos]] when you need source-code-level confirmation.
