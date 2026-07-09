package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) GetMovie(c *fiber.Ctx) error {
	tmdbID := c.Params("id")
	userID := h.optionalUserID(c)

	var movieID int
	var posterPath, posterLocal, title, overview string
	var runtime int
	var releaseDate *string
	var voteAverage float64
	err := h.db.QueryRow(context.Background(), `
		SELECT id, title, overview, runtime, release_date::text, vote_average,
		       COALESCE(poster_path, ''), COALESCE(poster_local, '')
		FROM movies WHERE tmdb_id = $1`, tmdbID,
	).Scan(&movieID, &title, &overview, &runtime, &releaseDate, &voteAverage, &posterPath, &posterLocal)

	if err == pgx.ErrNoRows {
		if h.cfg.TMDBAPIKey == "" {
			return c.JSON(demoMovie(tmdbID))
		}
		return h.fetchAndReturnMovie(c, tmdbID)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	cast, _ := h.getMovieCast(movieID)
	inLibrary := false
	watched := false
	if userID != "" {
		inLibrary, watched, _ = h.userMovieState(userID, movieID)
	}

	release := ""
	if releaseDate != nil {
		release = *releaseDate
	}

	resp := fiber.Map{
		"id":           movieID,
		"tmdb_id":      tmdbID,
		"title":        title,
		"overview":     overview,
		"runtime":      runtime,
		"release_date": release,
		"vote_average": voteAverage,
		"poster_url":   h.resolvePoster(firstNonEmpty(posterLocal, posterPath)),
		"cast":         cast,
		"in_library":   inLibrary,
		"watched":      watched,
	}
	if userID != "" {
		var tmdbInt int
		fmt.Sscanf(tmdbID, "%d", &tmdbInt)
		if score, review, ok := h.getUserRating(userID, "movie", tmdbInt); ok {
			resp["user_rating"] = fiber.Map{"score": score, "review": review}
		}
	}
	return c.JSON(resp)
}

func (h *Handler) fetchAndReturnMovie(c *fiber.Ctx, tmdbIDStr string) error {
	var tmdbID int
	fmt.Sscanf(tmdbIDStr, "%d", &tmdbID)

	movie, err := h.tmdb.GetMovie(tmdbID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "movie not found")
	}

	genres, _ := json.Marshal(movie.Genres)
	var movieID int
	err = h.db.QueryRow(context.Background(), `
		INSERT INTO movies (tmdb_id, title, overview, poster_path, backdrop_path, runtime, release_date, vote_average, genres)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tmdb_id) DO UPDATE SET updated_at = NOW()
		RETURNING id`,
		movie.ID, movie.Title, movie.Overview, movie.PosterPath, movie.BackdropPath,
		movie.Runtime, nullDate(movie.ReleaseDate), movie.VoteAverage, genres,
	).Scan(&movieID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	for _, member := range movie.Credits.Cast {
		personID, _ := h.upsertPerson(member.ID, member.Name, member.ProfilePath)
		if personID > 0 {
			_, _ = h.db.Exec(context.Background(), `
				INSERT INTO movie_cast (movie_id, person_id, character_name, cast_order)
				VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
				movieID, personID, member.Character, member.Order,
			)
		}
	}

	cast, _ := h.getMovieCast(movieID)
	return c.JSON(fiber.Map{
		"id":           movieID,
		"tmdb_id":      movie.ID,
		"title":        movie.Title,
		"overview":     movie.Overview,
		"runtime":      movie.Runtime,
		"release_date": movie.ReleaseDate,
		"vote_average": movie.VoteAverage,
		"poster_url":   h.posterURL(movie.PosterPath),
		"cast":         cast,
		"in_library":   false,
		"watched":      false,
	})
}

func (h *Handler) AddMovie(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var body struct {
		TMDBID     int    `json:"tmdb_id"`
		ListStatus string `json:"list_status"`
	}
	if err := c.BodyParser(&body); err != nil || body.TMDBID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "tmdb_id required")
	}
	listStatus := body.ListStatus
	if listStatus == "" {
		listStatus = "watching"
	}
	if !validListStatuses[listStatus] {
		return fiber.NewError(fiber.StatusBadRequest, "invalid list_status")
	}

	var movieID int
	err := h.db.QueryRow(context.Background(),
		`SELECT id FROM movies WHERE tmdb_id = $1`, body.TMDBID,
	).Scan(&movieID)
	if err == pgx.ErrNoRows {
		if h.cfg.TMDBAPIKey == "" {
			return fiber.NewError(fiber.StatusNotFound, "movie not found")
		}
		movie, tmdbErr := h.tmdb.GetMovie(body.TMDBID)
		if tmdbErr != nil {
			return fiber.NewError(fiber.StatusNotFound, "movie not found on tmdb")
		}
		genres, _ := json.Marshal(movie.Genres)
		err = h.db.QueryRow(context.Background(), `
			INSERT INTO movies (tmdb_id, title, overview, poster_path, backdrop_path, runtime, release_date, vote_average, genres)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
			movie.ID, movie.Title, movie.Overview, movie.PosterPath, movie.BackdropPath,
			movie.Runtime, nullDate(movie.ReleaseDate), movie.VoteAverage, genres,
		).Scan(&movieID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		for _, member := range movie.Credits.Cast {
			personID, _ := h.upsertPerson(member.ID, member.Name, member.ProfilePath)
			if personID > 0 {
				_, _ = h.db.Exec(context.Background(), `
					INSERT INTO movie_cast (movie_id, person_id, character_name, cast_order)
					VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
					movieID, personID, member.Character, member.Order,
				)
			}
		}
	}

	_, err = h.db.Exec(context.Background(), `
		INSERT INTO user_movies (user_id, movie_id, list_status) VALUES ($1::uuid, $2, $3)
		ON CONFLICT (user_id, movie_id) DO UPDATE SET list_status = EXCLUDED.list_status`,
		userID, movieID, listStatus,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.logMovieAdded(userID, movieID, listStatus)
	h.invalidateUserRecommendations(userID)
	return c.JSON(fiber.Map{"ok": true, "movie_id": movieID, "list_status": listStatus})
}

func (h *Handler) RemoveMovie(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	movieID := c.Params("id")
	_, err := h.db.Exec(context.Background(),
		`DELETE FROM user_movies WHERE user_id = $1::uuid AND movie_id = $2`, userID, movieID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) MarkMovieWatched(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	movieID := c.Params("id")
	_, err := h.db.Exec(context.Background(), `
		UPDATE user_movies SET watched = true, watched_at = NOW()
		WHERE user_id = $1::uuid AND movie_id = $2`,
		userID, movieID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.logMovieWatched(userID, movieID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) UnmarkMovieWatched(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	movieID := c.Params("id")
	_, err := h.db.Exec(context.Background(), `
		UPDATE user_movies SET watched = false, watched_at = NULL
		WHERE user_id = $1::uuid AND movie_id = $2`,
		userID, movieID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) TrendingMovies(c *fiber.Ctx) error {
	if cached, err := h.redis.Get(context.Background(), "trending:movie").Result(); err == nil {
		return c.Type("json").SendString(cached)
	}

	if h.cfg.TMDBAPIKey == "" {
		return c.JSON(fiber.Map{"results": demoTrendingMovies()})
	}

	trending, err := h.tmdb.TrendingMovies()
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	items := make([]fiber.Map, 0, len(trending.Results))
	for _, r := range trending.Results {
		title := r.Title
		if title == "" {
			title = r.Name
		}
		items = append(items, fiber.Map{
			"id":           r.ID,
			"tmdb_id":      r.ID,
			"title":        title,
			"media_type":   "movie",
			"poster_url":   h.posterURL(r.PosterPath),
			"vote_average": r.VoteAverage,
		})
	}
	resp := fiber.Map{"results": items}
	data, _ := json.Marshal(resp)
	_ = h.redis.Set(context.Background(), "trending:movie", data, 24*time.Hour).Err()
	return c.JSON(resp)
}

func (h *Handler) getMovieCast(movieID int) ([]fiber.Map, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT p.tmdb_id, p.name, mc.character_name,
		       COALESCE(p.profile_local, p.profile_path, '')
		FROM movie_cast mc
		JOIN persons p ON p.id = mc.person_id
		WHERE mc.movie_id = $1
		ORDER BY mc.cast_order LIMIT 20`, movieID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cast := []fiber.Map{}
	for rows.Next() {
		var tmdbID int
		var name, character, profile string
		if err := rows.Scan(&tmdbID, &name, &character, &profile); err != nil {
			continue
		}
		cast = append(cast, fiber.Map{
			"tmdb_id":     tmdbID,
			"name":        name,
			"character":   character,
			"profile_url": h.resolvePoster(profile),
		})
	}
	return cast, nil
}

func (h *Handler) getMovieLibraryItems(userID string, listStatus string) ([]fiber.Map, error) {
	statusFilter := ""
	args := []any{userID}
	if listStatus != "" && validListStatuses[listStatus] {
		statusFilter = " AND um.list_status = $2"
		args = append(args, listStatus)
	}
	rows, err := h.db.Query(context.Background(), `
		SELECT m.id, m.tmdb_id, m.title, m.vote_average,
		       COALESCE(m.poster_local, m.poster_path, ''),
		       um.watched, um.list_status
		FROM user_movies um
		JOIN movies m ON m.id = um.movie_id
		WHERE um.user_id = $1::uuid`+statusFilter+`
		ORDER BY um.added_at DESC`, args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []fiber.Map{}
	for rows.Next() {
		var id, tmdbID int
		var title, poster, itemListStatus string
		var voteAverage float64
		var watched bool
		if err := rows.Scan(&id, &tmdbID, &title, &voteAverage, &poster, &watched, &itemListStatus); err != nil {
			continue
		}
		progress := 0.0
		if watched {
			progress = 100
		}
		movies = append(movies, fiber.Map{
			"id":           id,
			"tmdb_id":      tmdbID,
			"title":        title,
			"media_type":   "movie",
			"list_status":  itemListStatus,
			"vote_average": voteAverage,
			"poster_url":   h.resolvePoster(poster),
			"progress":     progress,
			"watched":      boolToInt(watched),
			"total":        1,
		})
	}
	return movies, nil
}

func (h *Handler) getActiveMovieLibraryItems(userID string) ([]fiber.Map, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT m.id, m.tmdb_id, m.title, m.vote_average,
		       COALESCE(m.poster_local, m.poster_path, ''),
		       um.watched, um.list_status
		FROM user_movies um
		JOIN movies m ON m.id = um.movie_id
		WHERE um.user_id = $1::uuid
		  AND um.list_status IN ('watching', 'plan_to_watch')
		ORDER BY um.added_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []fiber.Map{}
	for rows.Next() {
		var id, tmdbID int
		var title, poster, itemListStatus string
		var voteAverage float64
		var watched bool
		if err := rows.Scan(&id, &tmdbID, &title, &voteAverage, &poster, &watched, &itemListStatus); err != nil {
			continue
		}
		progress := 0.0
		if watched {
			progress = 100
		}
		movies = append(movies, fiber.Map{
			"id":           id,
			"tmdb_id":      tmdbID,
			"title":        title,
			"media_type":   "movie",
			"list_status":  itemListStatus,
			"vote_average": voteAverage,
			"poster_url":   h.resolvePoster(poster),
			"progress":     progress,
			"watched":      boolToInt(watched),
			"total":        1,
		})
	}
	return movies, nil
}

func (h *Handler) userMovieState(userID string, movieID int) (bool, bool, error) {
	var inLibrary, watched bool
	err := h.db.QueryRow(context.Background(), `
		SELECT true, COALESCE(watched, false) FROM user_movies
		WHERE user_id = $1::uuid AND movie_id = $2`,
		userID, movieID,
	).Scan(&inLibrary, &watched)
	if err == pgx.ErrNoRows {
		return false, false, nil
	}
	return inLibrary, watched, err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func demoMovie(tmdbID string) fiber.Map {
	return fiber.Map{
		"id": 1, "tmdb_id": tmdbID, "title": "The Shawshank Redemption",
		"overview": "Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.",
		"runtime": 142, "release_date": "1994-09-23", "vote_average": 8.7,
		"poster_url": "https://image.tmdb.org/t/p/w500/9cqNxx0GxF0bflZmeSMuL5tnGzr.jpg",
		"cast": []fiber.Map{
			{"tmdb_id": 504, "name": "Tim Robbins", "character": "Andy Dufresne", "profile_url": "https://image.tmdb.org/t/p/w185/9cbE6oK5N2QqHuW6AAhRpmBEKXH.jpg"},
			{"tmdb_id": 192, "name": "Morgan Freeman", "character": "Ellis Boyd 'Red' Redding", "profile_url": "https://image.tmdb.org/t/p/w185/nIMl0LQmV2dpn7nzR2JXQ8f6kW4.jpg"},
		},
		"in_library": false,
		"watched":    false,
	}
}

func (h *Handler) importMovieItem(userID string, item importItem) (bool, error) {
	movieID, err := h.resolveImportedMovie(item)
	if err != nil || movieID == 0 {
		return false, err
	}

	watchedAt := time.Now().UTC()
	if item.WatchedAt != nil {
		watchedAt = item.WatchedAt.UTC()
	}

	_, err = h.db.Exec(context.Background(), `
		INSERT INTO user_movies (user_id, movie_id, watched, watched_at)
		VALUES ($1::uuid, $2, true, $3)
		ON CONFLICT (user_id, movie_id) DO UPDATE SET watched = true, watched_at = EXCLUDED.watched_at`,
		userID, movieID, watchedAt,
	)
	return err == nil, err
}

func (h *Handler) resolveImportedMovie(item importItem) (int, error) {
	if item.TMDBID > 0 {
		var movieID int
		err := h.db.QueryRow(context.Background(), `SELECT id FROM movies WHERE tmdb_id = $1`, item.TMDBID).Scan(&movieID)
		if err == nil {
			return movieID, nil
		}
		if err != pgx.ErrNoRows {
			return 0, err
		}
		if h.cfg.TMDBAPIKey == "" {
			return 0, pgx.ErrNoRows
		}
		movie, tmdbErr := h.tmdb.GetMovie(item.TMDBID)
		if tmdbErr != nil {
			return 0, tmdbErr
		}
		genres, _ := json.Marshal(movie.Genres)
		err = h.db.QueryRow(context.Background(), `
			INSERT INTO movies (tmdb_id, title, overview, poster_path, backdrop_path, runtime, release_date, vote_average, genres)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (tmdb_id) DO UPDATE SET title = EXCLUDED.title
			RETURNING id`,
			movie.ID, movie.Title, movie.Overview, movie.PosterPath, movie.BackdropPath,
			movie.Runtime, nullDate(movie.ReleaseDate), movie.VoteAverage, genres,
		).Scan(&movieID)
		return movieID, err
	}

	if item.ShowTitle != "" {
		var movieID int
		err := h.db.QueryRow(context.Background(), `SELECT id FROM movies WHERE LOWER(title) = LOWER($1) LIMIT 1`, item.ShowTitle).Scan(&movieID)
		if err == nil {
			return movieID, nil
		}
	}

	return 0, pgx.ErrNoRows
}

func demoTrendingMovies() []fiber.Map {
	return []fiber.Map{
		{"id": 278, "tmdb_id": 278, "title": "The Shawshank Redemption", "media_type": "movie", "poster_url": "https://image.tmdb.org/t/p/w500/9cqNxx0GxF0bflZmeSMuL5tnGzr.jpg", "vote_average": 8.7},
		{"id": 238, "tmdb_id": 238, "title": "The Godfather", "media_type": "movie", "poster_url": "https://image.tmdb.org/t/p/w500/3bhkrj58Vtu7enYsRolD1fZdja1.jpg", "vote_average": 8.7},
		{"id": 240, "tmdb_id": 240, "title": "The Godfather Part II", "media_type": "movie", "poster_url": "https://image.tmdb.org/t/p/w500/hek3koQnPfS6B6wz0L0IaS8ocQr.jpg", "vote_average": 8.6},
		{"id": 424, "tmdb_id": 424, "title": "Schindler's List", "media_type": "movie", "poster_url": "https://image.tmdb.org/t/p/w500/sF1U4EUQS8YHUYjNl3pMGNIQyr0.jpg", "vote_average": 8.6},
	}
}