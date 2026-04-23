

## TODO
- [ ] set a low number of shards to throttle concurrency.



## Slackbot
Common Issues: Self-Hosted Temporal
1. Database / Persistence Problems
   This is the #1 category. Specific flavors include:
- PostgreSQL connection pool starvation (PgBouncer saturation, cascading failures) — seen with Mistral
- TOAST cache misses causing tail latency spikes on history_node queries under load
- DEALLOCATE ALL storms on RDS Postgres under high load, bringing down the DB
- DB overwhelm from large, spiky workloads — seen with Amex
- Persistence latency issues causing Temporal service degradation — seen with Banamex
- Unsupported or edge-case databases (e.g., Cassandra 5, TiDB, AWS Keyspace/Aurora DSQL)

2. Scaling & Shard Configuration
   Wrong shard count at setup time (can't be changed after the fact) — this is a classic gotcha
   Shard reacquisition thundering herds after DB connectivity loss
   Questions around scaling to 4,096 shards and shard migration
   Multi-cluster / multi-tenant scale requirements, including single UI across clusters

3. Worker Tuning & Performance
   schedule_to_start latency spikes under load (JPMC saw >5s under load testing)
   context deadline exceeded errors — seen with Zoom
   High memory usage tied to visibility task backlogs
   CPU spikes on the DB when the Temporal UI is opened
   Misconfigured worker concurrency, poller counts, and rate limits

4. Visibility / Elasticsearch
   Visibility latency issues — customers coming from self-hosted (where visibility was faster) struggle when moving to Cloud [9]
   Large visibility_tasks table backlogs causing memory pressure

5. Upgrades & Version Management
   How to safely roll back to a previous version
   Upgrade sequencing complexity (schema migrations, multi-service rollout)
   Cassandra 5 not yet certified — blocking upgrades for customers who moved ahead of support

6. Authentication & TLS
   TLS configuration errors (unknown certificate authority, failed to build temporal client)
   mTLS setup complexity, especially for demos or dev environments
   Auth for cross-cluster replication (TLS + auth config between clusters) — seen with JPMC
   Token-based / modern auth not natively supported (Atlassian flagged this)

7. Deployment & Infrastructure Sizing
   No clear official guidance on hardware sizing for Cassandra/Elasticsearch given workload size
   Container/cloud deployment issues (e.g., timeout issues on Azure Container Apps, 429 errors on the UI)
   FedRAMP / air-gapped / on-prem requirements adding configuration complexity

8. Observability & Metrics
   Customers not collecting metrics from the start (a common early mistake per SailPoint)
   No Grafana dashboards set up, making it hard to diagnose issues
   No access to server-side logs (a pain point for those used to self-hosted)

9. Archival Configuration
   Misconfigured or misunderstood archival setup — led to a documentation inconsistency being flagged and fixed 
