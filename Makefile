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
# The server watches development.yaml and reloads on change — no restart needed.

reset:
	cp $(SCENARIOS)/base.yaml $(DYNCONFIG)
	@echo "→ reset to base (all defaults)"

# ── Scenarios ─────────────────────────────────────────────────────────────────
# restart needed.
shards ?= 4

scenario-shard-contention:
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
	cd loadgen && go run ./starter -workflows=$(w) -activities=$(a) -children=$(c) -interval=$(interval) -iterations=$(iterations)
