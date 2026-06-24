---
name: reference-self-hosted-issues-slackbot
description: Catalog of common self-hosted Temporal issues reported by Slackbot (real-customer pain points) — source of candidate scenarios for the playground
metadata:
  type: reference
---

Slackbot-sourced catalog of common issues from self-hosted Temporal customers. Mine this when picking the next playground scenario to implement — each bullet is a candidate "knob / component / signal" triple under [[feedback-scenario-design]].

Captured 2026-06-21.

## 1. Database / Persistence (highest frequency)
- PostgreSQL connection pool starvation (PgBouncer saturation, cascading failures)
- TOAST cache misses → tail latency spikes on `history_node` queries under load
- `DEALLOCATE ALL` storms on RDS Postgres under high load, bringing down the DB
- DB overwhelm from large, spiky workloads
- Persistence latency causing Temporal service degradation
- Unsupported / edge-case DBs (Cassandra 5, TiDB, AWS Keyspaces, Aurora DSQL)

## 2. Scaling & Shard Configuration
- Wrong shard count at setup (immutable post-deploy) — classic gotcha
- Shard reacquisition thundering herds after DB connectivity loss
- Scaling to 4,096 shards; shard migration
- Multi-cluster / multi-tenant scale; single UI across clusters

## 3. Worker Tuning & Performance
- `schedule_to_start` latency spikes under load
- `context deadline exceeded` errors
- High memory tied to visibility task backlogs
- CPU spikes on DB when the Temporal UI is opened
- Misconfigured worker concurrency, poller counts, rate limits

## 4. Visibility / Elasticsearch
- Visibility latency (customers moving from self-hosted → Cloud find it slower)
- Large `visibility_tasks` backlogs causing memory pressure

## 5. Upgrades & Version Management
- Safe rollback to previous version
- Upgrade sequencing complexity (schema migrations, multi-service rollout)
- Cassandra 5 not yet certified — blocks customers who moved ahead of support

## 6. Authentication & TLS
- TLS config errors ("unknown certificate authority", "failed to build temporal client")
- mTLS setup complexity for demos / dev
- Cross-cluster replication auth (TLS + auth between clusters)
- Token-based / modern auth not natively supported

## 7. Deployment & Infrastructure Sizing
- No official hardware sizing guidance for Cassandra / ES given workload
- Container / cloud issues (Azure Container Apps timeouts, 429s on UI)
- FedRAMP / air-gapped / on-prem config complexity

## 8. Observability & Metrics
- Not collecting metrics from day 1
- No Grafana dashboards → hard to diagnose
- No server-side log access (pain point coming from self-hosted)

## 9. Archival Configuration
- Misconfigured / misunderstood archival — drove a docs fix

---

**How to apply:** Cross-reference against the README "Scenario catalog". Items in #1 (persistence), #3 (worker tuning), and #4 (visibility) align tightly with the already-planned `persistence-overload`, `matching-backlog`, `visibility-backlog`, and `history-cache-pressure` scenarios. Items in #2 (shard reacquisition thundering herd) and #8 (no metrics from day 1) suggest net-new scenarios worth proposing.
