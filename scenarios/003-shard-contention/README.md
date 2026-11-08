# Shard contention

Every workflow lives on one history shard, picked by
`farm.Fingerprint32(namespaceID + "_" + workflowID) % numShards`. With very few shards, many
workflows land on the same shard, and they queue for that shard's I/O slot
(`history.shardIOConcurrency`, default 1).

The knob here is the shard count, which is fixed when the cluster is created — that is why this
scenario reinstalls the cluster with `shards=4` instead of the default 512.

Dashboard: **003 — Shard Contention** (Grafana, http://localhost:3000).

> Status: built and tested on docker-compose. Not yet re-verified on Kubernetes (server 1.32.2), so
> treat the expected signals below as a guide until it is.

## Setup

Keep these running in their own terminals (see the [main README](../../README.md#getting-started)):

```bash
make forward
make worker
```

Then:

```bash
make scenario-003-shard-contention shards=4
make load iterations=300 concurrent=100
```

For comparison, run the same load on the baseline (`make reset`, 512 shards) first.

## What to look at

The dashboard is organized in tiers, from primary signal to secondary effects:

| Row | Signal | Expected with 4 shards |
|---|---|---|
| Tier 1 — Primary Signals | `semaphore_latency`, `dd_shard_io_semaphore_latency` | Time waiting for the shard I/O slot rises |
| Tier 3 — Persistence Pressure | `persistence_latency{service_name="history"}` | p99 rises as hot shards queue |
| Tier 6 — Shard Topology | per-shard activity | A few shards carry all the load |
| SDK Metrics | worker task slots | Slots **drop** — the opposite of the throttling lab, where slots stay high |

## Questions

- Why can't you fix this one with dynamic config while the cluster is running?
- The worker is busy and schedule-to-start is high. How is this different from Exercise 1 of the
  [throttling lab](../001-throttling/README.md), where schedule-to-start was also high?
