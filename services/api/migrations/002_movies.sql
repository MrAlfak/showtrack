CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER UNIQUE NOT NULL,
    title TEXT NOT NULL,
    overview TEXT,
    poster_path TEXT,
    poster_local TEXT,
    backdrop_path TEXT,
    runtime INTEGER,
    release_date DATE,
    vote_average REAL,
    genres JSONB DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_movies_tmdb ON movies(tmdb_id);
CREATE INDEX idx_movies_title_trgm ON movies USING gin(title gin_trgm_ops);

CREATE TABLE movie_cast (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    person_id INTEGER NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    character_name TEXT,
    cast_order INTEGER DEFAULT 0,
    UNIQUE(movie_id, person_id)
);

CREATE TABLE user_movies (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    watched BOOLEAN DEFAULT FALSE,
    watched_at TIMESTAMPTZ,
    UNIQUE(user_id, movie_id)
);
