# temporal-server-playground

Local Temporal environment for learning internals through degradation scenarios. 
Uses dynamic config to induce specific failure modes and observe them in Prometheus/Grafana.


## Stack

| Service | URL |
|---|---|
| Temporal UI | http://localhost:8080 |
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9091 |
| Temporal gRPC | localhost:7233 |
| Elasticsearch | http://localhost:9200 |

## Getting Started

```bash
make up       # start everything
```
```bash
make logs     # follow Temporal server logs
```
```bash
make down     # stop everything
```

## Scenarios

Each scenario modifies `config/dynamicconfig/development.yaml` and destroy and start the environment
(Changing dynamic config does not require server reload, reload is performed to start with a clean environment)


### Frontend rate limit (WIP)
```bash
make scenario-frontend-rate-limit-001
```


### shard contention (WIP)
```bash
make scenario-shard-contention shards=2
```


### reset to defaults
```bash
make reset                 # back to defaults
```



## Load generation


```bash
make tidy #compile go code
```

```bash
make worker #start worker
```

```bash
# w=workflows, a=activities per workflow, 
# c=child workflows per workflow, intervalMs=time between workflow executions (0 = concurrent executions)
# iterations=number of iterations, each iteration start a activities c child workflows
#make starter w=1000 a=3 c=3 interval=0 iterations=100
make starter w=100 a=5 c=3 intervalMs=100 iterations=5
```


### Scenario catalog

Each scenario isolates one component with a single dynamic-config knob, so the
failure signal in Grafana is unambiguous. Status: ✅ implemented, 🚧 WIP, 📋 planned.

| # | Scenario | Status | Knob | What you'd see |
|---|---|---|---|---|
| 1 | **frontend-rate-limit** | 🚧 | `frontend.namespaceRPS` very low | `RESOURCE_EXHAUSTED` on low-priority RPCs first (PollWorkflowTask, PollActivityTask — see [frontend quotas](https://github.com/temporalio/temporal/blob/main/service/frontend/configs/quotas.go)); `schedule_to_start` latency climbs |
| 2 | **history-shard-contention** | ✅ | `history.shardIOConcurrency=1` + `NUM_HISTORY_SHARDS=2` | shard I/O semaphore latency; persistence p99 climbs on hot shards |
| 3 | **matching-backlog** | 📋 | `matching.numTaskqueueReadPartitions=1` + low `matching.rps` | `approximate_backlog_count` grows; `schedule_to_start` p99 on task queues |
| 4 | **visibility-backlog** | 📋 | `history.visibilityProcessorSchedulerWorkerCount=1` + low `history.visibilityProcessorMaxPollRPS` | visibility queue lag; search-attribute updates stale; list queries slow |
| 5 | **history-cache-pressure** | 📋 | tiny `history.hostLevelCacheMaxSize` (e.g. 16) | cache miss rate ↑; `GetWorkflowExecution` reads ↑; history latency ↑ |
| 6 | **persistence-overload** | 📋 | low `history.persistenceMaxQPS` (e.g. 50) | persistence throttling; `RESOURCE_EXHAUSTED_CAUSE_SYSTEM_OVERLOADED`; queue lag |

## Project Structure

```
├── docker-compose.yml
├── Makefile
├── config/
│   └── dynamicconfig/
│       ├── development.yaml          ← server reads this file
│       └── scenarios/                ← copy one to development.yaml via make
├── prometheus/
│   └── prometheus.yml
└── grafana/
    └── provisioning/
        ├── datasources/
        └── dashboards/
```
