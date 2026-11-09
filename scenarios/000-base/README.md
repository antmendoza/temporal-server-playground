# Baseline

No degradation: the server's defaults, plus two dev-only settings every scenario keeps. Run this
before any scenario so you know what healthy looks like on the dashboards.

Start the worker with 50 pollers, so it is never the bottleneck (the default, 2, is too few for
this load):

```bash
cd ../..
make worker pollers=50
```

Then reset the cluster and run the baseline load: each workflow runs 10 activities in parallel,
each taking 0.2s. The scenarios use this same worker and load, so this is what they compare against.

```bash
cd ../..
make reset
make load scenario=serverless_burst duration=10m rps=25 concurrent=50 \
  opts="--option fast-activity-count=0 --option slow-activity-count=10 --option slow-activity-duration=200ms"
```

`make reset` is also how you return to a clean cluster after any scenario.

## Baseline numbers

Wait 2 minutes, then read the dashboards. Measured on minikube (6 CPUs, 10 GiB, server 1.32.2):

| Signal | Baseline at `rps=25` |
|---|---|
| Workflows started / completed | ~25/s |
| Activities executed | ~250/s |
| `RESOURCE_EXHAUSTED` | none |
| Activity schedule-to-start p50 / p95 / p99 | ~14 / 40 / 85 ms |
| Workflow task schedule-to-start p50 / p95 / p99 | ~10 / 50 / 180 ms |
| Workflow end-to-end p50 / p95 | ~0.38 / 0.49 s (the 10 activities run in parallel) |
| Frontend requests | ~590/s |
| Persistence QPS | history ~1,280, frontend ~76, matching ~2 |

**`rps=25` is the highest rate this cluster handles without contention.** At 30 workflows/s,
schedule-to-start starts to grow with no limit set: the server itself is saturated (history at
~1,500 persistence calls/s on a single node). Above that, slowdowns you see come from the cluster,
not from a scenario.

**Your numbers may differ.** They depend on the laptop's CPU and on the resources given to Docker
Desktop and minikube. If your baseline shows schedule-to-start growing or throttling at `rps=25`,
lower `rps` until it is clean, and expect the scenario values to shift accordingly.

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

The scenarios compare against these numbers, so it's worth reading them on your own cluster:

- Service latencies (frontend, history, matching) and persistence latency by operation.
- Persistence QPS per service, and the History queue latencies (transfer, timer, visibility).
- Workflow task schedule-to-start and workflow end-to-end latency on the SDK dashboard.
- The *Resting Load* row with no load running: what the cluster does when nobody is using it.

To update the dashboards, download the newer JSON from that repo over the `000-` files.
