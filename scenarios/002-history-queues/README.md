# WIP: History queues lab — transfer, timer and visibility

Every change to a Workflow is committed together with the background **tasks** it causes: dispatch
a Workflow Task to Matching, fire a timer, update the List/Count record. Those tasks wait in
per-shard **queues** in the database until History runs them. Each queue has its own **reader**
(loads tasks from the database) and a host-wide **scheduler** with a pool of workers (runs them).

| Queue | Its tasks | When it falls behind |
|---|---|---|
| Transfer | Send Workflow and Activity Tasks to Matching, start children, deliver Signals | Workers sit idle while schedule-to-start climbs |
| Timer | Fire user timers, timeouts, retry backoffs, retention deletes | Timers and timeouts fire late |
| Visibility | Write the records behind List and Count | Search results lag real state |

In this lab you starve one queue at a time and learn to tell which one is behind from two metrics,
both labelled by `task_type` (`Transfer…`, `Timer…`, `Visibility…`):

- `task_latency_queue` — from task creation to completion: **is the queue keeping up?**
- `task_latency_schedule` — how long a loaded task waits for a worker: **are the workers the problem?**

Dashboard: **002 — History Queues** (Grafana, http://localhost:3000). The top row,
*Which queue is behind?*, is the one to watch in every exercise.

Background reading: *How the Temporal History Service works* → "Queues turn writes into side
effects" (in the Temporal server learning plan notes).

## Setup

Keep these running in their own terminals (see the [main README](../../README.md#getting-started)):

```bash
make forward
make worker
```

Then start the scenario and a steady load of timer workflows (each holds 10 timers of 2s), and
leave the load running for the whole lab:

```bash
cd ../..
make scenario-002-history-queues
make load scenario=workflow_with_many_timers duration=60m rps=60 concurrent=1000 \
  opts="--option concurrent-timers=10 --option timer-duration=2s"
```

## Baseline — know what normal looks like

Wait 2 minutes. At this load on minikube you should see roughly:

| Signal | Baseline |
|---|---|
| Workflows completed | ~60/s |
| Tasks/s | transfer ~180, timer ~240, visibility ~180 |
| `task_latency_queue` p95 | transfer ~0.03s, timer ~0.03s, visibility ~1s |
| `task_latency_schedule` p95 | ~0.001s for all three |
| Workflow end-to-end p95 | ~4s |

Visibility is slower than the others even when healthy: it writes to Elasticsearch.

## How each exercise works

1. Open `config/dynamicconfig/development.yaml` and uncomment **one** exercise block.
2. Run `make dynconfig` and wait about **2 minutes**.
3. Read the dashboard and answer the questions.
4. Comment the block out again and `make dynconfig`. The backlog takes **~5 minutes to drain**:
   wait until *Queue latency* is back to baseline before the next exercise. While it drains, the
   latency shown is the age of the old tasks finishing, so it looks worse before it looks better.

Exercises 1–3 throttle a queue's reader with `history.<queue>ProcessorMaxPollHostRPS`: how many
times per second all of that queue's readers on the host may load tasks from the database. The
default is unlimited; `1` starves it.

---

## Exercise 1 — the transfer queue falls behind

```yaml
history.transferProcessorMaxPollHostRPS:
  - value: 1
    constraints: {}
```

**Expected**
- *Queue latency*: **Transfer** jumps from ~0.03s to ~100s. Timer and Visibility stay normal.
- *Schedule latency*: Transfer stays at ~0.001s — the workers are idle, waiting for tasks.
- *Workflow task schedule-to-start* (SDK row) climbs to ~10s; completions drop from ~60/s to ~1.5/s.

**Questions**
- Your workers are idle but schedule-to-start is 10s. Where is the Workflow Task waiting, and why
  would adding workers not help?
- Why do timer and visibility tasks/s also drop, even though their queues are healthy?

## Exercise 2 — the timer queue falls behind

```yaml
history.timerProcessorMaxPollHostRPS:
  - value: 1
    constraints: {}
```

**Expected**
- *Queue latency*: **Timer** jumps to ~70s. Transfer stays ~0.02s.
- *Workflow task schedule-to-start* stays low (~0.2s): the worker looks healthy.
- *Workflow end-to-end* rises and completions drop from ~60/s to ~2/s — 2-second timers are
  firing over a minute late.

**Questions**
- Compare with Exercise 1: both make workflows slow. Which SDK metric tells them apart?
- Which other things ride on the timer queue and would also be late here? (Hint: timeouts, retries.)

## Exercise 3 — the visibility queue falls behind

```yaml
history.visibilityProcessorMaxPollHostRPS:
  - value: 1
    constraints: {}
```

**Expected**
- *Queue latency*: **Visibility** jumps from ~1s to ~50s; tasks/s from ~180 to ~13.
- Workflows are unaffected: still ~60/s completed.
- List and Count lie. Run this twice, 30s apart:
  ```bash
  temporal workflow count --address localhost:7233
  ```
  The count grows by a few hundred while ~1,800 workflows actually ran. New workflows show up
  late, so even the number of `Running` workflows goes *down*.

> If `temporal` dials the wrong server, prefix it with `env -u TEMPORAL_CONFIG_FILE`.

**Questions**
- A customer says "workflows are missing from the UI" but their workers see no errors. What would
  you check first?
- Why does a visibility backlog never slow the workflows themselves down?

## Exercise 4 — a knob that doesn't matter here

```yaml
history.transferProcessorSchedulerWorkerCount:   # default 512 per host
  - value: 1
    constraints: {}
```

**Expected**
- *Schedule latency*: Transfer rises slightly (~0.001s → ~0.01s). Nothing else changes.
- One worker keeps up with ~180 transfer tasks/s on this cluster.

**Questions**
- `task_latency_schedule` is the metric that would show a saturated worker pool. Why is it the
  right one to check before raising `…SchedulerWorkerCount`?

---

## Which queue is behind? — the queries

```promql
# Is each queue keeping up?
histogram_quantile(0.95, sum by (le, queue)(label_replace(rate(task_latency_queue_bucket{namespace="default"}[1m]), "queue", "$1", "task_type", "(Transfer|Timer|Visibility).*")))
# Are the workers the bottleneck?
histogram_quantile(0.95, sum by (le, queue)(label_replace(rate(task_latency_schedule_bucket{namespace="default"}[1m]), "queue", "$1", "task_type", "(Transfer|Timer|Visibility).*")))
```

| Queue latency | Schedule latency | Meaning |
|---|---|---|
| high | low | Tasks aren't being loaded fast enough: reader throttled, or the database is slow |
| high | high | Workers are saturated: check `…SchedulerWorkerCount` and host CPU |
| low | low | That queue is fine — look elsewhere |

Other knobs from the same family: `history.<queue>ProcessorMaxPollRPS` (per shard),
`history.<queue>TaskBatchSize`, `history.queuePendingTaskCriticalCount` (backlog isolation).
