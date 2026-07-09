package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/showtrack/api/internal/recommend"
	"github.com/showtrack/api/internal/tmdb"
)

type genreItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func demoGenres() []genreItem {
	return []genreItem{
		{ID: 18, Name: "Drama"},
		{ID: 10765, Name: "Sci-Fi & Fantasy"},
		{ID: 80, Name: "Crime"},
		{ID: 35, Name: "Comedy"},
		{ID: 53, Name: "Thriller"},
		{ID: 16, Name: "Animation"},
	}
}

func (h *Handler) GetGenres(c *fiber.Ctx) error {
	mediaType := c.Query("type", "tv")
	cacheKey := fmt.Sprintf("genres:%s", mediaType)

	if cached, err := h.redis.Get(context.Background(), cacheKey).Result(); err == nil {
		return c.Type("json").SendString(cached)
	}

	var genres []genreItem
	if h.cfg.TMDBAPIKey == "" {
		genres = demoGenres()
	} else if mediaType == "movie" {
		list, err := h.tmdb.MovieGenres()
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		for _, g := range list.Genres {
			genres = append(genres, genreItem{ID: g.ID, Name: g.Name})
		}
	} else {
		list, err := h.tmdb.TVGenres()
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		for _, g := range list.Genres {
			genres = append(genres, genreItem{ID: g.ID, Name: g.Name})
		}
	}

	resp := fiber.Map{"genres": genres}
	data, _ := json.Marshal(resp)
	_ = h.redis.Set(context.Background(), cacheKey, data, 7*24*time.Hour).Err()
	return c.JSON(resp)
}

func (h *Handler) Discover(c *fiber.Ctx) error {
	mediaType := c.Query("type", "tv")
	genreID, _ := strconv.Atoi(c.Query("genre"))
	sort := c.Query("sort", "popularity.desc")
	if sort == "" {
		sort = "popularity.desc"
	}

	cacheKey := fmt.Sprintf("discover:%s:%d:%s", mediaType, genreID, sort)
	if cached, err := h.redis.Get(context.Background(), cacheKey).Result(); err == nil {
		return c.Type("json").SendString(cached)
	}

	if h.cfg.TMDBAPIKey == "" {
		results := demoTrending()
		if mediaType == "movie" {
			for i := range results {
				results[i]["media_type"] = "movie"
			}
		}
		return c.JSON(fiber.Map{"results": results})
	}

	var result *tmdb.TrendingResponse

	if mediaType == "movie" {
		result, err = h.tmdb.DiscoverMovie(genreID, sort)
	} else {
		result, err = h.tmdb.DiscoverTV(genreID, sort)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	if result == nil {
		return c.JSON(fiber.Map{"results": []any{}})
	}

	items := make([]fiber.Map, 0, len(result.Results))
	for _, r := range result.Results {
		title := r.Name
		if title == "" {
			title = r.Title
		}
		items = append(items, fiber.Map{
			"id":           r.ID,
			"tmdb_id":      r.ID,
			"title":        title,
			"media_type":   mediaType,
			"poster_url":   h.posterURL(r.PosterPath),
			"vote_average": r.VoteAverage,
		})
	}
	resp := fiber.Map{"results": items}
	data, _ := json.Marshal(resp)
	_ = h.redis.Set(context.Background(), cacheKey, data, 6*time.Hour).Err()
	return c.JSON(resp)
}

func (h *Handler) invalidateUserRecommendations(userID string) {
	_ = h.redis.Del(context.Background(), fmt.Sprintf("recommendations:%s", userID)).Err()
}

func (h *Handler) GetRecommendations(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	cacheKey := fmt.Sprintf("recommendations:%s", userID)
	if cached, err := h.redis.Get(context.Background(), cacheKey).Result(); err == nil {
		return c.Type("json").SendString(cached)
	}

	engine := recommend.Engine{
		DB:        h.db,
		TMDB:      h.tmdb,
		HasTMDB:   h.cfg.TMDBAPIKey != "",
		PosterURL: h.posterURL,
	}
	out, err := engine.Generate(context.Background(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if len(out.Results) == 0 && !engine.HasTMDB {
		out.Results = demoTrendingScored()
		out.Engine = "demo"
	}

	resp := fiber.Map{
		"results":       out.Results,
		"seed_title":    out.SeedTitle,
		"engine":        out.Engine,
		"explanation":   out.Explanation,
		"signal_counts": out.SignalCounts,
	}
	data, _ := json.Marshal(resp)
	_ = h.redis.Set(context.Background(), cacheKey, data, 6*time.Hour).Err()
	return c.JSON(resp)
}

func demoTrendingScored() []recommend.ScoredItem {
	items := make([]recommend.ScoredItem, 0, len(demoTrending()))
	for _, row := range demoTrending() {
		tmdbID, _ := row["tmdb_id"].(int)
		if tmdbID == 0 {
			if f, ok := row["tmdb_id"].(float64); ok {
				tmdbID = int(f)
			}
		}
		title, _ := row["title"].(string)
		poster, _ := row["poster_url"].(string)
		mediaType, _ := row["media_type"].(string)
		if mediaType == "" {
			mediaType = "tv"
		}
		vote, _ := row["vote_average"].(float64)
		items = append(items, recommend.ScoredItem{
			TMDBID:      tmdbID,
			MediaType:   mediaType,
			Title:       title,
			PosterURL:   poster,
			VoteAverage: vote,
			Reason:      "trending",
		})
	}
	return items
}

func (h *Handler) userLibraryTMDBSet(userID string) map[int]bool {
	set := map[int]bool{}
	rows, err := h.db.Query(context.Background(), `
		SELECT s.tmdb_id FROM user_shows us JOIN shows s ON s.id = us.show_id WHERE us.user_id = $1::uuid
		UNION
		SELECT m.tmdb_id FROM user_movies um JOIN movies m ON m.id = um.movie_id WHERE um.user_id = $1::uuid`,
		userID,
	)
	if err != nil {
		return set
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if rows.Scan(&id) == nil {
			set[id] = true
		}
	}
	return set
}

func (h *Handler) Onboarding(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var body struct {
		Items []struct {
			TMDBID    int    `json:"tmdb_id"`
			MediaType string `json:"media_type"`
		} `json:"items"`
	}
	if err := c.BodyParser(&body); err != nil || len(body.Items) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "items required")
	}

	added := 0
	for _, item := range body.Items {
		if item.TMDBID == 0 {
			continue
		}
		mediaType := item.MediaType
		if mediaType == "" {
			mediaType = "tv"
		}
		if mediaType == "movie" {
			var movieID int
			err := h.db.QueryRow(context.Background(), `SELECT id FROM movies WHERE tmdb_id = $1`, item.TMDBID).Scan(&movieID)
			if err == pgx.ErrNoRows && h.cfg.TMDBAPIKey != "" {
				movie, tmdbErr := h.tmdb.GetMovie(item.TMDBID)
				if tmdbErr != nil {
					continue
				}
				genres, _ := json.Marshal(movie.Genres)
				err = h.db.QueryRow(context.Background(), `
					INSERT INTO movies (tmdb_id, title, overview, poster_path, runtime, release_date, vote_average, genres)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
					movie.ID, movie.Title, movie.Overview, movie.PosterPath, movie.Runtime,
					nullDate(movie.ReleaseDate), movie.VoteAverage, genres,
				).Scan(&movieID)
				if err != nil {
					continue
				}
			}
			if movieID > 0 {
				_, _ = h.db.Exec(context.Background(), `
					INSERT INTO user_movies (user_id, movie_id, list_status)
					VALUES ($1::uuid, $2, 'plan_to_watch') ON CONFLICT DO NOTHING`,
					userID, movieID,
				)
				added++
			}
			continue
		}

		var showID int
		err := h.db.QueryRow(context.Background(), `SELECT id FROM shows WHERE tmdb_id = $1`, item.TMDBID).Scan(&showID)
		if err == pgx.ErrNoRows && h.cfg.TMDBAPIKey != "" {
			show, tmdbErr := h.tmdb.GetTVShow(item.TMDBID)
			if tmdbErr != nil {
				continue
			}
			genres, _ := json.Marshal(show.Genres)
			err = h.db.QueryRow(context.Background(), `
				INSERT INTO shows (tmdb_id, title, overview, poster_path, status, vote_average, genres)
				VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
				show.ID, show.Name, show.Overview, show.PosterPath, show.Status, show.VoteAverage, genres,
			).Scan(&showID)
			if err == nil {
				_ = h.syncShowDetails(showID, show)
			}
		}
		if showID > 0 {
			_, _ = h.db.Exec(context.Background(), `
				INSERT INTO user_shows (user_id, show_id, list_status)
				VALUES ($1::uuid, $2, 'plan_to_watch') ON CONFLICT DO NOTHING`,
				userID, showID,
			)
			added++
		}
	}

	h.invalidateUserRecommendations(userID)
	return c.JSON(fiber.Map{"ok": true, "added": added})
}
