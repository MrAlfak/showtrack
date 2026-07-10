#!/usr/bin/env bash
set -euo pipefail

PROJECT="${COMPOSE_PROJECT_NAME:-showtrack-platform-axnc3w}"
FILE="${COMPOSE_FILE:-infra/docker-compose.dokploy.yml}"
LOG="${SHOWTRACK_PULL_LOG:-/tmp/showtrack-pull.log}"
PIDFILE="${SHOWTRACK_PULL_PID:-/tmp/showtrack-pull.pid}"

have_images() {
  docker image inspect ghcr.io/mralfak/showtrack-api:latest >/dev/null 2>&1 \
    && docker image inspect ghcr.io/mralfak/showtrack-web:latest >/dev/null 2>&1
}

pull_images() {
  echo "==> $(date -Is) pulling showtrack-api"
  docker pull ghcr.io/mralfak/showtrack-api:latest
  echo "==> $(date -Is) pulling showtrack-web"
  docker pull ghcr.io/mralfak/showtrack-web:latest
  echo "==> $(date -Is) pull complete"
}

start_background_pull() {
  if [[ -f "$PIDFILE" ]] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
    echo "==> Pull already running (pid $(cat "$PIDFILE")). Redeploy when finished."
    exit 0
  fi
  echo "==> Starting background GHCR pull (Dokploy 5min timeout workaround)"
  nohup bash -c "$(declare -f pull_images); pull_images" >>"$LOG" 2>&1 &
  echo $! >"$PIDFILE"
  echo "==> Pull pid $(cat "$PIDFILE"). Log: $LOG"
  echo "==> Redeploy in ~5-10 minutes after pull completes."
  exit 0
}

if ! have_images; then
  start_background_pull
fi

if [[ -f "$PIDFILE" ]] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null; then
  echo "==> Waiting for background pull..."
  wait "$(cat "$PIDFILE")" || true
  rm -f "$PIDFILE"
fi

if ! have_images; then
  echo "==> Images still missing; starting pull again"
  pull_images
fi

echo "==> Starting stack"
docker compose -p "$PROJECT" -f "$FILE" up -d --remove-orphans --no-build --pull never
echo "==> Deploy complete"
