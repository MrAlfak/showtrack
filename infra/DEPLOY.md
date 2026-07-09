# Dokploy Deployment Guide

Deploy ShowTrack as a Docker Compose stack on [Dokploy](https://dokploy.com).

## Prerequisites

- Dokploy server with Traefik enabled
- Domain names for API, web, and media (or use Dokploy-generated domains)
- [TMDB API key](https://www.themoviedb.org/settings/api)
- (Optional) Firebase service account JSON for Android/web push
- (Optional) Apple APNs `.p8` key for iOS push

## Step 1 — Create Compose Project

1. In Dokploy, create a new **Compose** application
2. Connect this repository
3. Set compose file path: `infra/docker-compose.yml`
4. Set working directory / context to repo root

## Step 2 — Environment Variables

Copy from `.env.example` and set in Dokploy:

| Variable | Required | Example |
|----------|----------|---------|
| `TMDB_API_KEY` | Yes | `your-tmdb-key` |
| `JWT_SECRET` | Yes | long random string |
| `POSTGRES_PASSWORD` | Yes | strong password |
| `API_DOMAIN` | Yes | `api.example.com` |
| `WEB_DOMAIN` | Yes | `app.example.com` |
| `MEDIA_DOMAIN` | Yes | `media.example.com` |
| `NEXT_PUBLIC_API_URL` | Yes | `https://api.example.com/api/v1` |
| `MEDIA_URL` | Yes | `https://media.example.com` |
| `CORS_ORIGINS` | Yes | `https://app.example.com` |
| `APNS_KEY_ID` | No | Apple key ID |
| `APNS_TEAM_ID` | No | Apple team ID |
| `APNS_BUNDLE_ID` | No | `com.example.showtrack` |
| `APNS_PRODUCTION` | No | `true` (use `false` for sandbox) |
| `NEXT_PUBLIC_FIREBASE_*` | No | Web push (see `.env.example`) |

## Step 3 — Push Secrets (Optional)

Create `infra/secrets/` on the server (or mount via Dokploy volumes):

```
infra/secrets/
├── fcm.json      # Firebase service account
└── apns.p8       # Apple push key
```

Without these files, the push service runs in **dry-run** mode (logs only).

## Step 4 — Deploy

Deploy the stack from Dokploy. Services started:

| Service | Port | Purpose |
|---------|------|---------|
| `postgres` | 5432 | Database (migrations auto-run on first boot) |
| `redis` | 6379 | Cache + notification queue |
| `api` | 8080 | REST API |
| `web` | 3000 | Next.js UI |
| `media` | 80 | Self-hosted posters |
| `push` | — | FCM/APNs notification worker |
| `sync` | — | One-shot poster sync (see schedules) |

## Step 5 — Domains (Traefik)

Compose labels route traffic by host:

- `API_DOMAIN` → `api` service
- `WEB_DOMAIN` → `web` service
- `MEDIA_DOMAIN` → `media` service

Point DNS A records to your Dokploy server IP.

## Step 6 — Scheduled Jobs

Add Dokploy schedules:

| Name | Cron | Command |
|------|------|---------|
| `sync-nightly` | `0 3 * * *` | `docker compose -f infra/docker-compose.yml run --rm sync` |

The `push` service checks for new episodes every hour by default (`EPISODE_CHECK_INTERVAL=1h`).

## Step 7 — Verify

```bash
curl https://api.example.com/api/v1/health
# {"status":"ok"}

curl https://api.example.com/api/v1/trending
# trending shows JSON
```

Open `https://app.example.com`, register an account, search a show, add to library.

## Import from TV Time

Supported formats in **Profile → Import**:

1. ShowTrack export (`showtrack-watch-history-v1`)
2. Flat JSON with `tmdb_id`, `season_number`, `episode_number`
3. TV Time nested JSON (shows with `seasons[].episodes[]` and `is_watched`)
4. TV Time movie list with `is_watched: true`

For GDPR zip exports, extract the JSON files first, then import the watched-history file.

## Troubleshooting

| Issue | Fix |
|-------|-----|
| API 502 | Check `api` container logs; verify `DATABASE_URL` |
| Empty trending | Set `TMDB_API_KEY` and restart `api` |
| Posters missing | Run `docker compose run --rm sync` |
| CORS errors | Set `CORS_ORIGINS` to exact web domain |
| Push dry-run only | Mount `fcm.json` / `apns.p8` in `infra/secrets/` |
