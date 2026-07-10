#!/usr/bin/env bash
set -euo pipefail

PROJECT="${COMPOSE_PROJECT_NAME:-showtrack-platform-axnc3w}"
FILE="${COMPOSE_FILE:-infra/docker-compose.dokploy.yml}"

echo "==> Pull GHCR images (parallel)"
docker pull ghcr.io/mralfak/showtrack-api:latest &
API_PID=$!
docker pull ghcr.io/mralfak/showtrack-web:latest &
WEB_PID=$!
wait "$API_PID" "$WEB_PID"

echo "==> Start stack (no build)"
docker compose -p "$PROJECT" -f "$FILE" up -d --remove-orphans --no-build --pull never

echo "==> Deploy complete"
