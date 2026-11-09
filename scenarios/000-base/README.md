# Baseline

No degradation: the server's defaults, plus two dev-only settings every scenario keeps. Run this
before any scenario so you know what healthy looks like on the dashboards.

```bash
cd ../..
make reset
make load iterations=300 concurrent=30   # completes in ~5s
```

For a longer look, run steady load instead: `make load duration=10m rps=20 concurrent=200`.

`make reset` is also how you return to a clean cluster after any scenario.

## Dashboards

Two general-purpose dashboards from the
[temporal-server-operations](https://github.com/tsurdilo/temporal-server-operations/tree/main/observability/dashboards)
repo, used unchanged. Each has a full panel reference and changelog in that repo. The version in
the title is the dashboard's own version, not the Temporal server's; both support server v1.20+.

| Dashboard (Grafana) | File | What it covers |
|---|---|---|
| **Temporal Server Dashboard v2.35.0** | `grafana/dashboards/000-temporal-server.json` | The whole server: throughput, service latencies, persistence, throttling, shard movement, History queues, matching, visibility, plus a *Resting Load (Idle Baseline)* row |
| **Temporal Go SDK Dashboard v1.5.0** | `grafana/dashboards/000-temporal-sdk-go.json` | The omes Go worker and load client: requests, pollers, slots, schedule-to-start, workflow and activity latencies |

### Empty panels are expected

On a single local cluster some rows stay empty by design: Cluster Replication, Archival,
Authorization, Namespace Deletion, History Scavenger and the task DLQ. Error panels are empty while
nothing fails. Some SDK panels (local activities, Nexus, failures) only fill when that kind of work
happens.

## What to note down

The scenario READMEs compare against these numbers, so it's worth reading them on your own cluster:

- Service latencies (frontend, history, matching) and persistence latency by operation.
- Persistence QPS per service, and the History queue latencies (transfer, timer, visibility).
- Workflow task schedule-to-start and workflow end-to-end latency on the SDK dashboard.
- The *Resting Load* row with no load running: what the cluster does when nobody is using it.

To update the dashboards, download the newer JSON from that repo over the `000-` files.
