# temporal-server-playground

A local Temporal cluster for learning server internals through degradation scenarios. Each scenario
changes the server's dynamic config to induce a specific failure mode, a load generator drives
traffic, and you observe the effect in Grafana.

## Stack

Runs on a local minikube cluster (profile and kube context `playground`, namespace `playground`),
installed with Helm: the official `temporalio/temporal` chart (frontend, history, matching and worker
as separate pods), PostgreSQL + Elasticsearch from the local `k8s/deps` chart, Prometheus and Grafana
from their community charts. Every `make` command passes `--kube-context playground` explicitly.

| Service | URL (via `make forward`) |
|---|---|
| Temporal UI | http://localhost:8080 |
| Grafana | http://localhost:3000 |
| Prometheus | http://localhost:9091 |
| Temporal gRPC | localhost:7233 |

## Getting started

Needs Docker Desktop, `minikube`, `kubectl`, `helm` and Go. Clone with the omes submodule:

```bash
git clone --recursive git@github.com:antmendoza/temporal-server-playground.git
```

```bash

minikube delete -p playground
make cluster  # once: create the minikube cluster
make up       # install everything into a fresh namespace (first run pulls images, ~5 min)
```

Then keep these two running, each in its own terminal, for as long as you use the playground:

```bash
make forward  # port-forwards above; reconnect on their own across reinstalls
```

```bash
make worker   # omes Go worker; survives reinstalls and reconnects
```

Other commands:

```bash
make logs        # follow Temporal server logs
make down        # delete the namespace (Temporal, DBs, Prometheus, Grafana)
make dashboards  # reload grafana/dashboards/*.json into the running Grafana. Use this to reload the dashboards after changing them manually.
```

## Scenarios

Each scenario lives in `scenarios/<name>/`: a `README.md` with the walkthrough and a
`dynamicconfig.yaml`. Start one with:

```bash
make scenario-<name>     # e.g. make scenario-001-throttling
```

This copies the scenario's config to `config/dynamicconfig/development.yaml`, deletes the
`playground` namespace and reinstalls everything, so every run starts clean: no workflows, no
metrics, config in place before the server boots. `make reset` does the same with the baseline.

Some scenarios have exercises you apply on a running cluster: edit
`config/dynamicconfig/development.yaml`, then `make dynconfig` (no reinstall; takes ~1–2 min to apply).

Each exercise changes one knob and names the signal you should see.
Status: ✅ verified on Kubernetes, 🚧 WIP, 📋 planned.

| Scenario | Status | What it teaches |
|---|---|---|
| [000-base](scenarios/000-base/README.md) | ✅ | What healthy looks like — run it first |
| [001-throttling](scenarios/001-throttling/README.md) | ✅ | Four limits behind one `RESOURCE_EXHAUSTED`: `RpsLimit`, `ConcurrentLimit`, `PersistenceLimit` on frontend and history, told apart by cause and service |
| [002-history-queues](scenarios/002-history-queues/README.md) | ✅ | Which History queue is behind — transfer, timer or visibility — from queue latency vs. schedule latency |
| [003-shard-contention](scenarios/003-shard-contention/README.md) | 🚧 | Too few history shards: workflows queue for the same shard's I/O slot (not yet re-verified on Kubernetes) |
| matching-backlog | 📋 | `matching.numTaskqueueReadPartitions=1` + low `matching.rps` → `approximate_backlog_count` grows |
| shard-handoff | 📋 | Scale History to 2–3 pods and delete one under load → `ShardOwnershipLost`, `sharditem_acquisition_latency`; then `history.shardLingerTimeLimit` |
| size-limits | 📋 | Low `limit.historyCount.error` → workflows terminated with nobody asking; Continue-as-New |
| hot-workflow | 📋 | Signal fan-in to one execution → `BusyWorkflow`; then `history.taskSchedulerEnableExecutionQueueScheduler` |
| history-cache-pressure | 📋 | tiny `history.hostLevelCacheMaxSize` → cache misses, more `GetWorkflowExecution` reads, history latency up |

## Load generation

Load comes from [omes](https://github.com/temporalio/omes), included as a git submodule in `omes/`
and pinned to a tested commit (`make worker` / `make load` fetch it if it's missing). To move to a
newer omes: `git -C omes pull`, test, then commit the new pin. The worker and the load are separate
so the worker survives a reinstall; they are paired by `run` (task queue `omes-<run>`).

```bash
# iterations=total, or duration=e.g. 30m for steady load; concurrent=max in flight;
# rps=max iterations started per second (0 = unlimited); scenario=omes scenario; opts=extra omes flags
make load iterations=300 concurrent=30
make load duration=30m rps=20 concurrent=200
make load scenario=throughput_stress concurrent=100 rps=50 opts="--option nexus-enabled=false"
```

`make load` keeps going when iterations fail, since degrading the server is the point.
`cd omes && go run ./cmd/omes list-scenarios` lists the omes scenarios.
Prometheus scrapes the worker (:9095) and the load client (:9096) as job `temporal-sdk`.

## Other environments

Environment-specific settings live in `k8s/values/temporal-<ENV>.yaml` (`ENV=minikube` by default),
so another cluster means a new values file plus `CTX=<kube context>`.

## Project structure

```
├── Makefile
├── scenarios/<name>/
│   ├── README.md                 ← walkthrough: setup, expected signals, questions
│   └── dynamicconfig.yaml        ← copied to config/dynamicconfig/development.yaml by make
├── config/dynamicconfig/
│   └── development.yaml          ← the live config; edit it for exercises, then make dynconfig
├── k8s/
│   ├── deps/                     ← local chart: PostgreSQL + Elasticsearch (minikube only)
│   └── values/                   ← Helm values: temporal(-<env>), prometheus, grafana
├── grafana/dashboards/           ← loaded into Grafana as a ConfigMap
├── omes/                         ← load generator (git submodule, pinned)
└── scripts/port-forward.sh
```
