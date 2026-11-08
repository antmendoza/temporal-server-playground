#!/usr/bin/env bash
# Keeps localhost ports pointed at the cluster. Each forward reconnects on its own, so it survives
# `make scenario-*` deleting and recreating the namespace.
set -u
CTX="${CTX:-playground}"
NS="${NS:-playground}"

forward() {  # forward <service> <local:remote>
  while true; do
    kubectl --context "$CTX" -n "$NS" port-forward "svc/$1" "$2" >/dev/null 2>&1
    sleep 2
  done
}

forward temporal-frontend 7233:7233 &
forward temporal-web 8080:8080 &
forward grafana 3000:80 &
forward prometheus-server 9091:80 &
trap 'kill 0' EXIT
wait
