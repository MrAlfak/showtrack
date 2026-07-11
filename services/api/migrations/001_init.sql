CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE shows (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER UNIQUE NOT NULL,
    title TEXT NOT NULL,
    overview TEXT,
    poster_path TEXT,
    poster_local TEXT,
    backdrop_path TEXT,
    status TEXT,
    first_air_date DATE,
    vote_average REAL,
    genres JSONB DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shows_tmdb ON shows(tmdb_id);
CREATE INDEX idx_shows_title_trgm ON shows USING gin(title gin_trgm_ops);

CREATE TABLE seasons (
    id SERIAL PRIMARY KEY,
    show_id INTEGER NOT NULL REFERENCES shows(id) ON DELETE CASCADE,
    season_number INTEGER NOT NULL,
    name TEXT,
    episode_count INTEGER DEFAULT 0,
    poster_path TEXT,
    UNIQUE(show_id, season_number)
);

CREATE TABLE episodes (
    id SERIAL PRIMARY KEY,
    season_id INTEGER NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    episode_number INTEGER NOT NULL,
    name TEXT,
    overview TEXT,
    air_date DATE,
    still_path TEXT,
    runtime INTEGER,
    notified BOOLEAN DEFAULT FALSE,
    UNIQUE(season_id, episode_number)
);

CREATE INDEX idx_episodes_air_date ON episodes(air_date);

CREATE TABLE persons (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER UNIQUE NOT NULL,
    name TEXT NOT NULL,
    profile_path TEXT,
    profile_local TEXT,
    biography TEXT,
    birthday DATE,
    place_of_birth TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_persons_name_trgm ON persons USING gin(name gin_trgm_ops);

CREATE TABLE person_credits (
    id SERIAL PRIMARY KEY,
    person_id INTEGER NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    media_type TEXT NOT NULL CHECK (media_type IN ('tv', 'movie')),
    media_tmdb_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    character_name TEXT,
    poster_path TEXT,
    poster_local TEXT,
    release_date DATE,
    vote_average REAL,
    UNIQUE(person_id, media_type, media_tmdb_id)
);

CREATE TABLE show_cast (
    id SERIAL PRIMARY KEY,
    show_id INTEGER NOT NULL REFERENCES shows(id) ON DELETE CASCADE,
    person_id INTEGER NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    character_name TEXT,
    cast_order INTEGER DEFAULT 0,
    UNIQUE(show_id, person_id)
);

CREATE TABLE user_shows (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    show_id INTEGER NOT NULL REFERENCES shows(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, show_id)
);

CREATE TABLE user_episodes (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id INTEGER NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    watched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, episode_id)
);

CREATE TABLE device_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('android', 'ios', 'web')),
    app_version TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, token)
);

CREATE TABLE notification_log (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    episode_id INTEGER REFERENCES episodes(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    body TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE trending_cache (
    id SERIAL PRIMARY KEY,
    media_type TEXT NOT NULL DEFAULT 'tv',
    tmdb_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    poster_path TEXT,
    poster_local TEXT,
    vote_average REAL,
    rank_position INTEGER NOT NULL,
    cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
