package handlers

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) logActivity(userID, activityType string, payload fiber.Map) {
	if userID == "" || activityType == "" {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = h.db.Exec(context.Background(),
		`INSERT INTO activities (user_id, activity_type, payload) VALUES ($1::uuid, $2, $3)`,
		userID, activityType, data,
	)
}

func (h *Handler) logEpisodeWatched(userID, episodeID string) {
	var showTitle, episodeName, posterPath string
	var tmdbID, seasonNumber, episodeNumber int
	err := h.db.QueryRow(context.Background(), `
		SELECT s.title, s.tmdb_id, COALESCE(s.poster_path, ''), s2.season_number, e.episode_number, COALESCE(e.name, '')
		FROM episodes e
		JOIN seasons s2 ON s2.id = e.season_id
		JOIN shows s ON s.id = s2.show_id
		WHERE e.id = $1`, episodeID,
	).Scan(&showTitle, &tmdbID, &posterPath, &seasonNumber, &episodeNumber, &episodeName)
	if err != nil {
		return
	}
	h.logActivity(userID, "episode_watched", fiber.Map{
		"title":           showTitle,
		"tmdb_id":         tmdbID,
		"media_type":      "tv",
		"poster_url":      h.resolvePoster(posterPath),
		"season_number":   seasonNumber,
		"episode_number":  episodeNumber,
		"episode_name":    episodeName,
	})
}

func (h *Handler) logMovieWatched(userID, movieID string) {
	var title, posterPath string
	var tmdbID int
	err := h.db.QueryRow(context.Background(),
		`SELECT title, tmdb_id, COALESCE(poster_path, '') FROM movies WHERE id = $1`, movieID,
	).Scan(&title, &tmdbID, &posterPath)
	if err != nil {
		return
	}
	h.logActivity(userID, "movie_watched", fiber.Map{
		"title":      title,
		"tmdb_id":    tmdbID,
		"media_type": "movie",
		"poster_url": h.resolvePoster(posterPath),
	})
}

func (h *Handler) logShowAdded(userID string, showID int, listStatus string) {
	var title, posterPath string
	var tmdbID int
	err := h.db.QueryRow(context.Background(),
		`SELECT title, tmdb_id, COALESCE(poster_path, '') FROM shows WHERE id = $1`, showID,
	).Scan(&title, &tmdbID, &posterPath)
	if err != nil {
		return
	}
	h.logActivity(userID, "show_added", fiber.Map{
		"title":       title,
		"tmdb_id":     tmdbID,
		"media_type":  "tv",
		"poster_url":  h.resolvePoster(posterPath),
		"list_status": listStatus,
	})
}

func (h *Handler) logMovieAdded(userID string, movieID int, listStatus string) {
	var title, posterPath string
	var tmdbID int
	err := h.db.QueryRow(context.Background(),
		`SELECT title, tmdb_id, COALESCE(poster_path, '') FROM movies WHERE id = $1`, movieID,
	).Scan(&title, &tmdbID, &posterPath)
	if err != nil {
		return
	}
	h.logActivity(userID, "movie_added", fiber.Map{
		"title":       title,
		"tmdb_id":     tmdbID,
		"media_type":  "movie",
		"poster_url":  h.resolvePoster(posterPath),
		"list_status": listStatus,
	})
}

func (h *Handler) logStatusChanged(userID, mediaType string, mediaID int, listStatus string) {
	var title, posterPath string
	var tmdbID int
	var query string
	if mediaType == "movie" {
		query = `SELECT title, tmdb_id, COALESCE(poster_path, '') FROM movies WHERE id = $1`
	} else {
		query = `SELECT title, tmdb_id, COALESCE(poster_path, '') FROM shows WHERE id = $1`
		mediaType = "tv"
	}
	err := h.db.QueryRow(context.Background(), query, mediaID).Scan(&title, &tmdbID, &posterPath)
	if err != nil {
		return
	}
	h.logActivity(userID, "status_changed", fiber.Map{
		"title":       title,
		"tmdb_id":     tmdbID,
		"media_type":  mediaType,
		"poster_url":  h.resolvePoster(posterPath),
		"list_status": listStatus,
	})
}
