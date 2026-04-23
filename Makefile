.PHONY: up down restart logs reset \
        scenario-visibility scenario-matching scenario-history scenario-frontend \
        tidy worker starter

DYNCONFIG = config/dynamicconfig/development.yaml
SCENARIOS = config/dynamicconfig/scenarios

# ── Cluster lifecycle ──────────────────────────────────────────────────────────

up:
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

scenario-visibility:
	cp $(SCENARIOS)/visibility-backlog.yaml $(DYNCONFIG)
	@echo "→ visibility backlog: write QPS capped at 5"

scenario-matching:
	cp $(SCENARIOS)/matching-backlog.yaml $(DYNCONFIG)
	@echo "→ matching backlog: 1 partition, low persistence QPS"

scenario-history:
	cp $(SCENARIOS)/history-shard-contention.yaml $(DYNCONFIG)
	@echo "→ history shard contention: persistence QPS capped at 20"

scenario-frontend:
	cp $(SCENARIOS)/frontend-rate-limit.yaml $(DYNCONFIG)
	@echo "→ frontend rate limit: visibility API capped at 1 RPS per namespace"

# ── Load generator ────────────────────────────────────────────────────────────

tidy:
	cd loadgen && go mod tidy

worker:
	cd loadgen && go run ./worker

starter:
	cd loadgen && go run ./starter -workflows=$(w) -activities=$(a) -children=$(c)
