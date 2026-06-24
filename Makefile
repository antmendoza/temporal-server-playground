.PHONY: up down restart logs reset \
        scenario-shard-contention \
        tidy worker starter

DYNCONFIG = config/dynamicconfig/development.yaml
SCENARIOS = config/dynamicconfig/scenarios

# ── Cluster lifecycle ──────────────────────────────────────────────────────────

up:
	make down
	docker compose up -d

down:
	docker compose down  -v


logs:
	docker compose logs -f temporal

# ── Scenarios ─────────────────────────────────────────────────────────────────
# For some scenarios restart is not needed, we restart the server to start with a clean environment
reset:
	cp $(SCENARIOS)/base.yaml $(DYNCONFIG)
	@echo "→ reset to base (all defaults)"


scenario-frontend-rate-limit-001:
	@echo "→ applying frontend-rate-limit dynamic config and reinitializing cluster"
	make up
	cp $(SCENARIOS)/frontend-rate-limit-001.yaml $(DYNCONFIG)


shards ?= 4
scenario-shard-contention-001:
	cp $(SCENARIOS)/shard-contention-001.yaml $(DYNCONFIG)
	@echo "→ shard contention: reinitializing cluster with NUM_HISTORY_SHARDS=$(shards)"
	make down
	make reset
	NUM_HISTORY_SHARDS=$(shards) make up

# ── Load generator ────────────────────────────────────────────────────────────

tidy:
	cd loadgen && go mod tidy

worker:
	cd loadgen && go run ./worker

starter:
	cd loadgen && go run ./starter -workflows=$(w) -activities=$(a) -children=$(c) -intervalMs=$(intervalMs) -iterations=$(iterations)
