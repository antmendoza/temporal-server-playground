# Throttling lab — frontend rate limits and persistence QPS

Temporal has several independent limits that protect the cluster. When any of them trips, the client
sees the same error: `RESOURCE_EXHAUSTED`. In this lab you trip four of them, one at a time, and
learn to tell them apart from three labels on the server metrics:

- `resource_exhausted_cause` — **which kind** of limit: `RpsLimit`, `ConcurrentLimit`, `PersistenceLimit`
- `resource_exhausted_scope` — **how wide** the limit is: `Namespace` (a per-namespace limit, such
  as `frontend.namespaceRPS`) or `System` (a host- or cluster-wide limit, such as `frontend.rps`)
- `service_name` — **which service** rejected the call: frontend, history, matching

Dashboard: **001 — Throttling Lab** (Grafana, http://localhost:3000). The top row,
*Which limit is it?*, is the one to watch in every exercise.

## Setup

Keep these running in their own terminals (see the [main README](../../README.md#getting-started)).
The worker needs enough pollers that it is never the bottleneck, so start it with 50:

```bash
cd ../..
make forward
make worker pollers=50
```

Then start the scenario and a steady load, and leave the load running for the whole lab. Each
workflow runs 10 activities in parallel, each taking 0.2s:

```bash
cd ../..
make scenario-001-throttling
make load scenario=serverless_burst duration=60m rps=25 concurrent=50 \
  opts="--option fast-activity-count=0 --option slow-activity-count=10 --option slow-activity-duration=200ms"
```

`rps=25` is the highest rate this minikube cluster handles with no server contention: at 30/s
schedule-to-start starts to grow on its own, with no limit set.

## Baseline — know what normal looks like

Wait 2 minutes and read the dashboard. Measured on minikube (server 1.32.2):

| Signal | Baseline |
|---|---|
| Workflows started / completed | ~25/s |
| Activities executed | ~250/s |
| `RESOURCE_EXHAUSTED` | none |
| Activity schedule-to-start p50 / p95 / p99 | ~14 / 40 / 85 ms |
| Workflow task schedule-to-start p50 / p95 / p99 | ~10 / 50 / 180 ms |
| Workflow end-to-end p50 / p95 | ~0.38 / 0.49 s (the 10 activities run in parallel) |
| Frontend requests | ~590/s |
| Persistence QPS | history ~1,280, frontend ~76, matching ~2 |

## How each exercise works

While the load generator is running:

1. Open `config/dynamicconfig/development.yaml` and uncomment **one** exercise block.
2. Run:
```bash
cd ../..
make dynconfig
```

3. Wait about **2 minutes**: the ConfigMap takes ~1 min to reach the pods, and the namespace quota
   is recalculated lazily after that.
4. Read the dashboard and compare it with the expected signals.
5. Comment the block out again, `make dynconfig`, and wait until the baseline is back before the next one.

Each limit is set to about half of what the baseline uses, so it bites but doesn't stop the cluster.

---

## Exercise 1 — frontend namespace RPS

```yaml
frontend.namespaceRPS:          # default 2400 per frontend host; baseline uses ~590
  - value: 295
    constraints: {}
```

**Expected**
- *RESOURCE_EXHAUSTED by service · scope · cause · operation*: `frontend · scope:Namespace ·
  RpsLimit · PollWorkflowExecutionHistory`, ~23/s.
- Workflows started and completed drop from ~25/s to ~12.5/s; frontend requests from ~590/s to ~320/s.
- Schedule-to-start and workflow end-to-end latency stay at baseline.

## Exercise 2 — frontend concurrent long-polls

```yaml
frontend.namespaceCount:        # default 1200 per frontend host, per API
  - value: 2
    constraints: {}
```

**Expected**
- *RESOURCE_EXHAUSTED*: `frontend · scope:Namespace · ConcurrentLimit` on `PollWorkflowTaskQueue`
  (~8/s), `PollActivityTaskQueue` (~8/s) and `PollWorkflowExecutionHistory` (~12/s).
- *SDK long-poll failures*: the same polls failing with `RESOURCE_EXHAUSTED` on the client side.
- *Task queue backlog*: workflow tasks pile up (~100 waiting).
- Workflows started drop from ~25/s to ~1/s and completions stop.

## Exercise 3 — history persistence QPS

```yaml
history.persistenceMaxQPS:      # default 9000 per history host; baseline uses ~1,280
  - value: 638
    constraints: {}
```

**Expected**
- *Persistence calls throttled*: `history · scope:System`, across many operations, including
  `GetWorkflowExecution`, `UpdateWorkflowExecution` and background calls like `GetCurrentClusterMetadata`.
- *RESOURCE_EXHAUSTED*: a little `history · scope:System · PersistenceLimit` on `RecordWorkflowTaskStarted`.
  Nothing from the frontend.
- Workflows completed drop from ~25/s to ~12/s; workflow task schedule-to-start p99 rises to ~3s
  and workflow end-to-end p95 to ~2.8s.
- *Persistence latency* stays flat: calls are rejected before they reach the database.

## Exercise 4 — frontend persistence QPS (revert promptly)

```yaml
frontend.persistenceMaxQPS:     # default 2000 per frontend host; baseline uses ~76
  - value: 37
    constraints: {}
```

**Expected**
- *Persistence calls throttled*: `frontend`, mostly `GetCurrentClusterMetadata` (~28/s).
- *RESOURCE_EXHAUSTED*: `frontend · scope:System · PersistenceLimit` on `StartWorkflowExecution` (~26/s).
- *Workflows completed vs. start failures*: completions fall from ~25/s to ~2.5/s; the load client
  sees `StartWorkflowExecution` fail with `DEADLINE_EXCEEDED` (~4/s).

This limit also throttles the frontend's own background calls, so don't leave it on longer than needed.

---

## Reading the signals: from RESOURCE_EXHAUSTED to the limit behind it

When a client gets `RESOURCE_EXHAUSTED`, two panels in the dashboard's top row tell you which limit
rejected it:

- **RESOURCE_EXHAUSTED by service · scope · cause · operation**: every call a service rejected,
  with its cause and scope. This is where you start.
- **Persistence calls throttled by service · scope · operation**: calls a service's own database
  rate limiter rejected before they reached the database. When the first panel says
  `PersistenceLimit`, this one tells you which service's limit it was.

To query them yourself in Prometheus (http://localhost:9091):

```promql
# panel: RESOURCE_EXHAUSTED by service · scope · cause · operation
sum by (service_name, operation, resource_exhausted_cause, resource_exhausted_scope)(rate(service_errors_resource_exhausted[1m]))

# panel: Persistence calls throttled by service · scope · operation
sum by (service_name, operation, resource_exhausted_scope)(rate(persistence_errors_resource_exhausted[1m]))
```

Match what you see to the limit:

| Cause | Scope | Emitted by | Limit to look at | Exercise |
|---|---|---|---|---|
| `RpsLimit` | `Namespace` | frontend | `frontend.namespaceRPS` (or `frontend.globalNamespaceRPS`) | 1 |
| `ConcurrentLimit` | `Namespace` | frontend | `frontend.namespaceCount` | 2 |
| `PersistenceLimit` | `System` | the service in the persistence panel | `<service>.persistenceMaxQPS` | 3 (history), 4 (frontend) |

A host-wide frontend limit such as `frontend.rps` also reports `RpsLimit`; it isn't part of this lab.

Other related knobs to explore after the lab: `frontend.namespaceBurstRatio`,
`frontend.namespaceRPS.visibility`, `matching.persistenceMaxQPS`, `history.rps`, `matching.rps`.
