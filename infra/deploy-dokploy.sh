#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PROJECT="${COMPOSE_PROJECT_NAME:-showtrack-platform-axnc3w}"
FILE="${COMPOSE_FILE:-infra/docker-compose.dokploy.yml}"
REG="${SHOWTRACK_LOCAL_REGISTRY:-localhost:5000}"
STATE_DIR="${SHOWTRACK_STATE_DIR:-/tmp/showtrack-deploy-state}"
LOG="${SHOWTRACK_PULL_LOG:-/tmp/showtrack-pull.log}"

API_GHCR="${SHOWTRACK_API_GHCR:-ghcr.io/mralfak/showtrack-api:latest}"
WEB_GHCR="${SHOWTRACK_WEB_GHCR:-ghcr.io/mralfak/showtrack-web:latest}"
API_LOCAL="${SHOWTRACK_API_IMAGE:-$REG/showtrack-api:latest}"
WEB_LOCAL="${SHOWTRACK_WEB_IMAGE:-$REG/showtrack-web:latest}"

mkdir -p "$STATE_DIR"

log() {
  echo "==> $(date -Is) $*" | tee -a "$LOG"
}

have_image() {
  docker image inspect "$1" >/dev/null 2>&1
}

start_mirror_job() {
  local name=$1 ghcr=$2 local=$3 marker=$4 lock=$5
  if have_image "$local" || [[ -f "$marker" ]]; then
    return 0
  fi
  if [[ -f "$lock" ]]; then
    log "$name mirror already running — redeploy in a few minutes"
    exit 0
  fi
  touch "$lock"
  log "starting background mirror: $ghcr -> $local"
  (
    set -e
    docker pull "$ghcr"
    docker tag "$ghcr" "$local"
    docker push "$local"
    touch "$marker"
    rm -f "$lock"
    log "$name mirror complete"
  ) >>"$LOG" 2>&1 &
  disown || true
  log "$name mirror queued — redeploy after it finishes (~5–15 min on slow links)"
  exit 0
}

export SHOWTRACK_API_IMAGE="$API_LOCAL"
export SHOWTRACK_WEB_IMAGE="$WEB_LOCAL"

log "deploy from $ROOT"

# Always ensure infra is up (small public images, fast)
docker compose -p "$PROJECT" -f "$FILE" up -d postgres redis media --remove-orphans --no-build --pull missing

start_mirror_job "api" "$API_GHCR" "$API_LOCAL" "$STATE_DIR/api.done" "$STATE_DIR/api.lock"
start_mirror_job "web" "$WEB_GHCR" "$WEB_LOCAL" "$STATE_DIR/web.done" "$STATE_DIR/web.lock"

if ! have_image "$API_LOCAL" || ! have_image "$WEB_LOCAL"; then
  log "waiting for mirrored images — redeploy once api/web are ready"
  exit 0
fi

log "starting api + web"
docker compose -p "$PROJECT" -f "$FILE" up -d api web --remove-orphans --no-build --pull never
log "deploy complete"
