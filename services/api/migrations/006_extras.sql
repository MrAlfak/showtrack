-- Personal ratings, custom lists, binge tracking support
CREATE TABLE user_ratings (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_type TEXT NOT NULL CHECK (media_type IN ('tv', 'movie')),
    tmdb_id INTEGER NOT NULL,
    score SMALLINT NOT NULL CHECK (score BETWEEN 1 AND 10),
    review TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, media_type, tmdb_id)
);

CREATE INDEX IF NOT EXISTS idx_user_ratings_tmdb ON user_ratings (media_type, tmdb_id);

CREATE TABLE custom_lists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_custom_lists_user ON custom_lists (user_id);

CREATE TABLE custom_list_items (
    list_id UUID NOT NULL REFERENCES custom_lists(id) ON DELETE CASCADE,
    media_type TEXT NOT NULL CHECK (media_type IN ('tv', 'movie')),
    tmdb_id INTEGER NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    poster_path TEXT NOT NULL DEFAULT '',
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (list_id, media_type, tmdb_id)
);
