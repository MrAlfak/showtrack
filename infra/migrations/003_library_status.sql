-- TV Time-style library list types
ALTER TABLE user_shows
    ADD COLUMN IF NOT EXISTS list_status TEXT NOT NULL DEFAULT 'watching'
        CHECK (list_status IN ('watching', 'plan_to_watch', 'watched', 'dropped', 'archived'));

ALTER TABLE user_movies
    ADD COLUMN IF NOT EXISTS list_status TEXT NOT NULL DEFAULT 'watching'
        CHECK (list_status IN ('watching', 'plan_to_watch', 'watched', 'dropped', 'archived'));

CREATE INDEX IF NOT EXISTS idx_user_shows_list_status ON user_shows (user_id, list_status);
CREATE INDEX IF NOT EXISTS idx_user_movies_list_status ON user_movies (user_id, list_status);
