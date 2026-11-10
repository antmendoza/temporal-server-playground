.PHONY: cluster up down forward logs reset dynconfig dashboards worker load

CTX ?= playground
NS = playground
KUBECTL = kubectl --context $(CTX) -n $(NS)
HELM = helm --kube-context $(CTX) -n $(NS)
ENV ?= minikube

DYNCONFIG = config/dynamicconfig/development.yaml
SCENARIOS = scenarios
shards ?= 512
OMES_DIR ?= $(CURDIR)/omes
OMES = cd $(OMES_DIR) && go run ./cmd/omes

run ?= playground
scenario ?= workflow_with_single_noop_activity
iterations ?= 1000
duration ?=
concurrent ?= 50
pollers ?=
rps ?= 0
opts ?=

# ── Cluster lifecycle ──────────────────────────────────────────────────────────

cluster:
	minikube start -p $(CTX) --driver=docker --cpus=6 --memory=10g

# The live config is generated (not tracked): start from the baseline when it's missing
$(DYNCONFIG):
	@mkdir -p $(dir $@)
	cp $(SCENARIOS)/000-base/dynamicconfig.yaml $@

# Deletes the whole namespace (Temporal, its DBs, Prometheus, Grafana) and installs it from scratch
up: $(DYNCONFIG)
	make down
	kubectl --context $(CTX) create namespace $(NS)
	$(HELM) install deps k8s/deps --wait --timeout 10m
	@mkdir -p .build
	@{ echo "server:"; echo "  dynamicConfig:"; sed 's/^/    /' $(DYNCONFIG); } > .build/dynamicconfig-values.yaml
	$(HELM) install temporal temporal/temporal --version 1.8.0 \
		-f k8s/values/temporal.yaml -f k8s/values/temporal-$(ENV).yaml -f .build/dynamicconfig-values.yaml \
		--set server.config.persistence.numHistoryShards=$(shards) --wait --timeout 10m
	$(KUBECTL) create configmap grafana-dashboards --from-file=grafana/dashboards/
	$(HELM) install prometheus prometheus-community/prometheus --version 29.45.1 -f k8s/values/prometheus.yaml
	$(HELM) install grafana grafana/grafana --version 10.5.15 -f k8s/values/grafana.yaml

down:
	kubectl --context $(CTX) delete namespace $(NS) --ignore-not-found --wait

# Keep running in its own terminal; reconnects across scenario restarts
forward:
	CTX=$(CTX) NS=$(NS) scripts/port-forward.sh

logs:
	$(KUBECTL) logs -f -l app.kubernetes.io/name=temporal --all-containers --max-log-requests 10 --prefix

# ── Scenarios ─────────────────────────────────────────────────────────────────
# Every scenario starts from a clean namespace with its config in place before boot
reset:
	cp $(SCENARIOS)/000-base/dynamicconfig.yaml $(DYNCONFIG)
	make up

# Push development.yaml to the running server without a reinstall; pods see it in ~1 min
dynconfig: $(DYNCONFIG)
	$(KUBECTL) create configmap temporal-dynamic-config --from-file=dynamic_config.yaml=$(DYNCONFIG) \
		--dry-run=client -o yaml | $(KUBECTL) apply -f -

# Reload grafana/dashboards/*.json into the running Grafana (replace: the ConfigMap is too big for apply)
dashboards:
	$(KUBECTL) create configmap grafana-dashboards --from-file=grafana/dashboards/ --dry-run=client -o yaml \
		| $(KUBECTL) replace -f -
	$(KUBECTL) rollout restart deploy/grafana

scenario-%:
	@mkdir -p $(dir $(DYNCONFIG))
	cp $(SCENARIOS)/$*/dynamicconfig.yaml $(DYNCONFIG)
	@echo "→ $*: reinstalling (NUM_HISTORY_SHARDS=$(shards))"
	make up

# ── Load generator (omes) ─────────────────────────────────────────────────────
# worker and load are paired by run-id (task queue omes-$(run))

# omes is a pinned git submodule; fetch it on first use if the repo was cloned without --recursive
$(OMES_DIR)/go.mod:
	git submodule update --init omes

worker: $(OMES_DIR)/go.mod
	$(OMES) run-worker --language go --run-id $(run) --server-address localhost:7233 \
		--worker-prom-listen-address 0.0.0.0:9095 \
		$(if $(pollers),--worker-max-concurrent-workflow-pollers $(pollers) --worker-max-concurrent-activity-pollers $(pollers))

load: $(OMES_DIR)/go.mod
	$(OMES) run-scenario --scenario $(scenario) --run-id $(run) --server-address localhost:7233 \
		$(if $(duration),--duration $(duration),--iterations $(iterations)) --max-concurrent $(concurrent) --max-iterations-per-second $(rps) \
		--prom-listen-address 0.0.0.0:9096 --continue-on-iteration-failure $(opts)
