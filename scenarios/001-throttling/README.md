# Throttling lab — frontend rate limits and persistence QPS

Temporal has several independent limits that protect the cluster. When any of them trips, the client
sees the same error: `RESOURCE_EXHAUSTED`. In this lab you trip four of them, one at a time, and
learn to tell them apart from two labels on the server metrics:

- `resource_exhausted_cause` — **which kind** of limit: `RpsLimit`, `ConcurrentLimit`, `PersistenceLimit`
- `service_name` — **which service** rejected the call: frontend, history, matching

Dashboard: **001 — Throttling Lab** (Grafana, http://localhost:3000). The top row,
*Which limit is it?*, is the one to watch in every exercise.

## Setup

Keep these running in their own terminals (see the [main README](../../README.md#getting-started)):

```bash
make forward
make worker
```

Then start the scenario and a steady load, and leave the load running for the whole lab:

```bash
make scenario-001-throttling
make load duration=30m rps=20 concurrent=200
```

## Baseline — know what normal looks like

Wait 2 minutes and read the top row. At `rps=20` on minikube you should see roughly:

| Signal | Baseline |
|---|---|
| Workflows completed | ~20/s |
| `RESOURCE_EXHAUSTED` | none |
| Persistence QPS | history ~615, frontend ~60, matching ~5 |
| Workflow task schedule-to-start p95 | ~0.02s |

## How each exercise works

1. Open `config/dynamicconfig/development.yaml` and uncomment **one** exercise block.
2. Run `make dynconfig`.
3. Wait about **2 minutes**: the ConfigMap takes ~1 min to reach the pods, and the namespace quota
   is recalculated lazily after that.
4. Read the dashboard and answer the questions.
5. Comment the block out again, `make dynconfig`, and wait until the baseline is back before the next one.

---

## Exercise 1 — frontend namespace RPS

```yaml
frontend.namespaceRPS:          # default 2400 per frontend host
  - value: 50
    constraints: {}
```

**Expected**
- *RESOURCE_EXHAUSTED by service · cause · operation*: `frontend · RpsLimit`, mostly on
  `PollWorkflowExecutionHistory`, a little on `PollWorkflowTaskQueue`.
- *Workflows completed* drops from ~20/s to ~10/s.
- *Schedule-to-Start Latency — Workflow Tasks* rises to ~10s.

**Questions**
- The load client calls `PollWorkflowExecutionHistory` to wait for each workflow's result. Why is it
  throttled far more than the worker's `PollWorkflowTaskQueue`? (Hint: on 1.32, poll APIs wait for
  a rate-limit token before failing.)
- Workflow task schedule-to-start went up, but the worker is idle. Where is the time going?

## Exercise 2 — frontend concurrent long-polls

```yaml
frontend.namespaceCount:        # default 1200 per frontend host, per API
  - value: 20
    constraints: {}
```

**Expected**
- `frontend · ConcurrentLimit` on `PollWorkflowTaskQueue` and `PollActivityTaskQueue`.
- *Workflows completed* stays at ~20/s and schedule-to-start stays low: **no impact**.

**Questions**
- The dashboard shows errors, yet nothing slowed down. What does that say about the worker's poller
  count compared to what this load needs?
- Would you page someone for this error rate? What would you check first?

## Exercise 3 — history persistence QPS

```yaml
history.persistenceMaxQPS:      # default 9000 per history host
  - value: 200
    constraints: {}
```

**Expected**
- *Persistence calls throttled*: `history`, across many operations (including background ones like
  `GetCurrentClusterMetadata` and task-queue range completes).
- *RESOURCE_EXHAUSTED*: `PersistenceLimit` from `history`, and also from `matching`
  (`AddWorkflowTask`, `PollWorkflowTaskQueue`). No `RpsLimit` on the frontend.
- *Workflows completed* drops from ~20/s to ~3/s; schedule-to-start ~10s.
- *Persistence latency* stays flat: calls are rejected before they reach the database.

**Questions**
- Matching reports `PersistenceLimit`, but matching's own persistence QPS is tiny. Which service is
  the real bottleneck, and which panel proves it?
- The database is not slow. What would you tell a customer who says "Postgres is overloaded"?

## Exercise 4 — frontend persistence QPS (revert promptly)

```yaml
frontend.persistenceMaxQPS:     # default 2000 per frontend host
  - value: 10
    constraints: {}
```

**Expected**
- *Persistence calls throttled*: `frontend`, mostly `GetNamespace` and `GetCurrentClusterMetadata`.
- *RESOURCE_EXHAUSTED*: `frontend · PersistenceLimit` on `StartWorkflowExecution`.
- *Workflows completed vs. start failures*: completions fall to 0, start failures ~20/s.

This limit also throttles the frontend's membership heartbeat (`UpsertClusterMembership`), so don't
leave it on longer than needed.

**Questions**
- Exercises 1 and 4 both hit the frontend. How do you tell them apart from the client side alone?
  And from the server side?

---

## Which limit is it? — the two queries

```promql
sum by (service_name, operation, resource_exhausted_cause)(rate(service_errors_resource_exhausted[1m]))
sum by (service_name, operation)(rate(persistence_errors_resource_exhausted[1m]))
```

| Cause | Emitted by | Limit to look at |
|---|---|---|
| `RpsLimit` | frontend | `frontend.namespaceRPS`, `frontend.globalNamespaceRPS`, `frontend.rps` |
| `ConcurrentLimit` | frontend | `frontend.namespaceCount` |
| `PersistenceLimit` | the service in `persistence_errors_resource_exhausted` | `<service>.persistenceMaxQPS` |

Other related knobs to explore after the lab: `frontend.namespaceBurstRatio`,
`frontend.namespaceRPS.visibility`, `matching.persistenceMaxQPS`, `history.rps`, `matching.rps`.
