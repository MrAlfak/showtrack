# ShowTrack

Self-hosted TV show tracker — inspired by TV Time, built with **shadcn/ui** design, **Go** backend, and **Dokploy** deployment.

## Features

- Track TV shows and movies, mark episodes and films watched
- Search shows, movies, and people (TMDB)
- Person pages with full filmography
- Poster images self-hosted (WebP via sync worker)
- Import/export watch history (TV + movies)
- Push notifications (FCM + APNs) via dedicated container
- Dark-first shadcn-inspired mobile UI (PWA-ready)

## Stack

| Layer | Tech |
|---|---|
| Web UI | Next.js 16 + shadcn/ui + Tailwind v4 |
| API | Go + Fiber |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Media | Nginx (poster CDN) |
| Sync | Go worker (TMDB → DB + posters) |
| Push | Go service (FCM v1 + APNs HTTP/2) |
| Deploy | Docker Compose on Dokploy |

## Project Structure

```
showtrack/
├── apps/web/           # Next.js UI (shadcn design)
├── services/
│   ├── api/            # REST API
│   ├── sync/           # TMDB sync + poster download
│   └── push/           # Notification sender
└── infra/
    ├── docker-compose.yml
    ├── migrations/
    └── nginx/
```

## Quick Start (Development)

### Web UI only (demo data)

```bash
cd apps/web
npm install
npm run dev
```

Open http://localhost:3000 — works with demo data, no backend required.

### Full stack (Docker)

```bash
cp .env.example .env
# Add your TMDB_API_KEY to .env

cd infra
docker compose up -d --build
```

Services:
- Web: http://localhost:3000 (or WEB_DOMAIN)
- API: http://localhost:8080/api/v1/health
- Media: http://localhost (poster files)

### Nightly sync (Dokploy Schedule)

```bash
# Cron: 0 3 * * *
docker compose run --rm sync
```

## Dokploy Deployment

See [infra/DEPLOY.md](infra/DEPLOY.md) for step-by-step Dokploy setup.

1. Create a new **Compose** project in Dokploy
2. Point to this repo, set compose file: `infra/docker-compose.yml`
3. Set environment variables from `.env.example`
4. Add domains for `api`, `web`, `media` services (Traefik labels included)
5. Create schedules:
   - `sync-nightly`: `0 3 * * *` → `docker compose run --rm sync`
6. Mount push secrets at `infra/secrets/`:
   - `fcm.json` (Firebase service account)
   - `apns.p8` (Apple push key)

## API Endpoints

```
GET  /api/v1/health
GET  /api/v1/search?q=
GET  /api/v1/trending
GET  /api/v1/trending/movies
GET  /api/v1/shows/:tmdb_id
GET  /api/v1/movies/:tmdb_id
GET  /api/v1/persons/:tmdb_id
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/me/library          (auth)
GET  /api/v1/me/dashboard        (auth)
GET  /api/v1/me/export           (auth)
POST /api/v1/me/import           (auth)
POST /api/v1/shows               (auth)
DELETE /api/v1/shows/:id         (auth)
POST /api/v1/movies              (auth)
DELETE /api/v1/movies/:id        (auth)
POST /api/v1/movies/:id/watched  (auth)
DELETE /api/v1/movies/:id/watched (auth)
POST /api/v1/episodes/:id/watched (auth)
DELETE /api/v1/episodes/:id/watched (auth)
POST /api/v1/devices             (auth)
```

## TMDB Attribution

This product uses the TMDB API but is not endorsed or certified by TMDB.
Display TMDB logo and attribution in production builds.

## License

MIT
