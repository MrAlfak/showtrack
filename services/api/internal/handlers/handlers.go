package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/showtrack/api/internal/config"
	"github.com/showtrack/api/internal/tmdb"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db    *pgxpool.Pool
	redis *redis.Client
	cfg   config.Config
	tmdb  *tmdb.Client
}

type episodePayload struct {
	ID             int    `json:"id"`
	EpisodeNumber  int    `json:"episode_number"`
	Name           string `json:"name"`
	Overview       string `json:"overview"`
	AirDate        string `json:"air_date,omitempty"`
	Runtime        int    `json:"runtime"`
	StillURL       string `json:"still_url"`
	Watched        bool   `json:"watched"`
}

type seasonPayload struct {
	ID           int              `json:"id"`
	SeasonNumber int              `json:"season_number"`
	Name         string           `json:"name"`
	EpisodeCount int              `json:"episode_count"`
	Episodes     []episodePayload `json:"episodes"`
}

type dashboardStatsPayload struct {
	Shows    int `json:"shows"`
	Movies   int `json:"movies"`
	Episodes int `json:"episodes"`
	Total    int `json:"total"`
	Hours    int `json:"hours"`
	Streak   int `json:"streak"`
	BingeToday int `json:"binge_today"`
}

type upcomingPayload struct {
	ShowTitle     string `json:"show_title"`
	ShowTMDBID    int    `json:"show_tmdb_id"`
	EpisodeID     int    `json:"episode_id"`
	SeasonNumber  int    `json:"season_number"`
	EpisodeNumber int    `json:"episode_number"`
	EpisodeName   string `json:"episode_name"`
	AirDate       string `json:"air_date,omitempty"`
	PosterURL     string `json:"poster_url"`
}

type importItem struct {
	TMDBID        int
	ShowTitle     string
	MediaType     string
	SeasonNumber  int
	EpisodeNumber int
	WatchedAt     *time.Time
}

func New(db *pgxpool.Pool, redis *redis.Client, cfg config.Config) *Handler {
	return &Handler{
		db:    db,
		redis: redis,
		cfg:   cfg,
		tmdb:  tmdb.New(cfg.TMDBAPIKey),
	}
}

func (h *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "service": "showtrack-api"})
}

type authRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req authRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Email == "" || req.Password == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email and password required")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "hash failed")
	}
	id := uuid.New()
	_, err = h.db.Exec(context.Background(),
		`INSERT INTO users (id, email, password_hash, display_name) VALUES ($1, $2, $3, $4)`,
		id, req.Email, string(hash), req.DisplayName,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, "email already exists")
	}
	token, err := h.signToken(id.String())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "token failed")
	}
	return c.JSON(fiber.Map{"token": token, "user_id": id})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req authRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	var id uuid.UUID
	var hash *string
	err := h.db.QueryRow(context.Background(),
		`SELECT id, password_hash FROM users WHERE email = $1`, req.Email,
	).Scan(&id, &hash)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}
	if hash == nil || *hash == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "use google sign-in for this account")
	}
	if bcrypt.CompareHashAndPassword([]byte(*hash), []byte(req.Password)) != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}
	token, err := h.signToken(id.String())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "token failed")
	}
	return c.JSON(fiber.Map{"token": token, "user_id": id})
}

func (h *Handler) signToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func (h *Handler) Search(c *fiber.Ctx) error {
	q := c.Query("q")
	if q == "" {
		return c.JSON(fiber.Map{"results": []any{}})
	}

	cacheKey := fmt.Sprintf("search:%s", q)
	if cached, err := h.redis.Get(context.Background(), cacheKey).Result(); err == nil {
		return c.Type("json").SendString(cached)
	}

	if h.cfg.TMDBAPIKey == "" {
		return c.JSON(fiber.Map{"results": demoSearch(q)})
	}

	result, err := h.tmdb.SearchMulti(q)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	items := make([]fiber.Map, 0, len(result.Results))
	for _, r := range result.Results {
		title := r.Name
		if title == "" {
			title = r.Title
		}
		poster := r.PosterPath
		if r.MediaType == "person" {
			poster = r.ProfilePath
		}
		items = append(items, fiber.Map{
			"id":           r.ID,
			"title":        title,
			"overview":     r.Overview,
			"media_type":   r.MediaType,
			"poster_url":   h.posterURL(poster),
			"vote_average": r.VoteAverage,
		})
	}
	resp := fiber.Map{"results": items}
	data, _ := json.Marshal(resp)
	_ = h.redis.Set(context.Background(), cacheKey, data, time.Hour).Err()
	return c.JSON(resp)
}

func (h *Handler) Trending(c *fiber.Ctx) error {
	if cached, err := h.redis.Get(context.Background(), "trending:tv").Result(); err == nil {
		return c.Type("json").SendString(cached)
	}

	if h.cfg.TMDBAPIKey == "" {
		return c.JSON(fiber.Map{"results": demoTrending()})
	}

	trending, err := h.tmdb.TrendingTV()
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	items := make([]fiber.Map, 0, len(trending.Results))
	for _, r := range trending.Results {
		items = append(items, fiber.Map{
			"id":           r.ID,
			"tmdb_id":      r.ID,
			"title":        r.Name,
			"media_type":   "tv",
			"poster_url":   h.posterURL(r.PosterPath),
			"vote_average": r.VoteAverage,
		})
	}
	resp := fiber.Map{"results": items}
	data, _ := json.Marshal(resp)
	_ = h.redis.Set(context.Background(), "trending:tv", data, 24*time.Hour).Err()
	return c.JSON(resp)
}

func (h *Handler) GetShow(c *fiber.Ctx) error {
	tmdbID := c.Params("id")
	userID := h.optionalUserID(c)

	var showID int
	var posterPath, posterLocal, title, overview, status string
	var voteAverage float64
	err := h.db.QueryRow(context.Background(),
		`SELECT id, title, overview, status, vote_average, COALESCE(poster_path, ''), COALESCE(poster_local, '') FROM shows WHERE tmdb_id = $1`,
		tmdbID,
	).Scan(&showID, &title, &overview, &status, &voteAverage, &posterPath, &posterLocal)

	if err == pgx.ErrNoRows {
		if h.cfg.TMDBAPIKey == "" {
			show := demoShow(tmdbID)
			return c.JSON(show)
		}
		return h.fetchAndReturnShow(c, tmdbID)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	cast, _ := h.getShowCast(showID)
	seasons, _ := h.getShowSeasons(showID, userID)
	inLibrary := false
	progress := 0.0
	watchedEpisodes := 0
	totalEpisodes := 0
	if userID != "" {
		inLibrary, _ = h.userHasShow(userID, showID)
		watchedEpisodes, totalEpisodes, _ = h.getUserProgress(userID, showID)
		if totalEpisodes > 0 {
			progress = float64(watchedEpisodes) / float64(totalEpisodes) * 100
		}
	}

	resp := fiber.Map{
		"id":           showID,
		"tmdb_id":      tmdbID,
		"title":        title,
		"overview":     overview,
		"status":       status,
		"vote_average": voteAverage,
		"poster_url":   h.resolvePoster(firstNonEmpty(posterLocal, posterPath)),
		"cast":         cast,
		"seasons":      seasons,
		"in_library":   inLibrary,
		"progress":     progress,
		"watched":      watchedEpisodes,
		"total":        totalEpisodes,
	}
	if userID != "" {
		var tmdbInt int
		fmt.Sscanf(tmdbID, "%d", &tmdbInt)
		if score, review, ok := h.getUserRating(userID, "tv", tmdbInt); ok {
			resp["user_rating"] = fiber.Map{"score": score, "review": review}
		}
	}
	return c.JSON(resp)
}

func (h *Handler) fetchAndReturnShow(c *fiber.Ctx, tmdbIDStr string) error {
	var tmdbID int
	fmt.Sscanf(tmdbIDStr, "%d", &tmdbID)

	show, err := h.tmdb.GetTVShow(tmdbID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "show not found")
	}

	genres, _ := json.Marshal(show.Genres)
	var showID int
	err = h.db.QueryRow(context.Background(), `
		INSERT INTO shows (tmdb_id, title, overview, poster_path, backdrop_path, status, first_air_date, vote_average, genres)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (tmdb_id) DO UPDATE SET updated_at = NOW()
		RETURNING id`,
		show.ID, show.Name, show.Overview, show.PosterPath, show.BackdropPath,
		show.Status, nullDate(show.FirstAirDate), show.VoteAverage, genres,
	).Scan(&showID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if syncErr := h.syncShowDetails(showID, show); syncErr != nil {
		return fiber.NewError(fiber.StatusInternalServerError, syncErr.Error())
	}

	for _, member := range show.Credits.Cast {
		personID, _ := h.upsertPerson(member.ID, member.Name, member.ProfilePath)
		if personID > 0 {
			_, _ = h.db.Exec(context.Background(), `
				INSERT INTO show_cast (show_id, person_id, character_name, cast_order)
				VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`,
				showID, personID, member.Character, member.Order,
			)
		}
	}

	cast, _ := h.getShowCast(showID)
	seasons, _ := h.getShowSeasons(showID, "")
	return c.JSON(fiber.Map{
		"id":           showID,
		"tmdb_id":      show.ID,
		"title":        show.Name,
		"overview":     show.Overview,
		"status":       show.Status,
		"vote_average": show.VoteAverage,
		"poster_url":   h.posterURL(show.PosterPath),
		"cast":         cast,
		"seasons":      seasons,
		"in_library":   false,
		"progress":     0,
		"watched":      0,
		"total":        0,
	})
}

func (h *Handler) GetPerson(c *fiber.Ctx) error {
	tmdbID := c.Params("id")

	var personID int
	var name, bio, profilePath, profileLocal string
	err := h.db.QueryRow(context.Background(),
		`SELECT id, name, COALESCE(biography,''), COALESCE(profile_path,''), COALESCE(profile_local,'') FROM persons WHERE tmdb_id = $1`,
		tmdbID,
	).Scan(&personID, &name, &bio, &profilePath, &profileLocal)

	if err == pgx.ErrNoRows {
		if h.cfg.TMDBAPIKey == "" {
			return c.JSON(demoPerson(tmdbID))
		}
		return h.fetchAndReturnPerson(c, tmdbID)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	credits, _ := h.getPersonCredits(personID)

	return c.JSON(fiber.Map{
		"id":          personID,
		"tmdb_id":     tmdbID,
		"name":        name,
		"biography":   bio,
		"profile_url": h.resolvePoster(firstNonEmpty(profileLocal, profilePath)),
		"credits":     credits,
	})
}

func (h *Handler) fetchAndReturnPerson(c *fiber.Ctx, tmdbIDStr string) error {
	var tmdbID int
	fmt.Sscanf(tmdbIDStr, "%d", &tmdbID)

	person, err := h.tmdb.GetPerson(tmdbID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "person not found")
	}

	personID, _ := h.upsertPerson(person.ID, person.Name, person.ProfilePath)
	_, _ = h.db.Exec(context.Background(),
		`UPDATE persons SET biography = $1, birthday = $2, place_of_birth = $3 WHERE id = $4`,
		person.Biography, nullDate(person.Birthday), person.PlaceOfBirth, personID,
	)

	for _, credit := range person.CombinedCredits.Cast {
		title := credit.Name
		if title == "" {
			title = credit.Title
		}
		date := credit.FirstAirDate
		if date == "" {
			date = credit.ReleaseDate
		}
		_, _ = h.db.Exec(context.Background(), `
			INSERT INTO person_credits (person_id, media_type, media_tmdb_id, title, character_name, poster_path, release_date, vote_average)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT DO NOTHING`,
			personID, credit.MediaType, credit.ID, title, credit.Character,
			credit.PosterPath, nullDate(date), credit.VoteAverage,
		)
	}

	credits, _ := h.getPersonCredits(personID)
	return c.JSON(fiber.Map{
		"id":          personID,
		"tmdb_id":     person.ID,
		"name":        person.Name,
		"biography":   person.Biography,
		"profile_url": h.posterURL(person.ProfilePath),
		"credits":     credits,
	})
}

func (h *Handler) GetLibrary(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	listStatus := c.Query("list_status")
	shows, err := h.getLibraryItems(userID, listStatus)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	movies, err := h.getMovieLibraryItems(userID, listStatus)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"shows": shows, "movies": movies})
}

func (h *Handler) GetDashboard(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	library, err := h.getActiveLibraryItems(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	movieLibrary, err := h.getActiveMovieLibraryItems(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	upcoming, err := h.getUpcomingEpisodes(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	stats, err := h.computeUserStats(userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	combined := append(library, movieLibrary...)

	return c.JSON(fiber.Map{
		"stats":    stats,
		"library":  combined,
		"upcoming": upcoming,
	})
}

func (h *Handler) ExportWatchHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	rows, err := h.db.Query(context.Background(), `
		SELECT s.tmdb_id, s.title, sn.season_number, e.episode_number, e.name, ue.watched_at
		FROM user_episodes ue
		JOIN episodes e ON e.id = ue.episode_id
		JOIN seasons sn ON sn.id = e.season_id
		JOIN shows s ON s.id = sn.show_id
		WHERE ue.user_id = $1::uuid
		ORDER BY ue.watched_at DESC`, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	items := []fiber.Map{}
	for rows.Next() {
		var tmdbID, seasonNumber, episodeNumber int
		var title, episodeName string
		var watchedAt time.Time
		if err := rows.Scan(&tmdbID, &title, &seasonNumber, &episodeNumber, &episodeName, &watchedAt); err != nil {
			continue
		}
		items = append(items, fiber.Map{
			"tmdb_id":        tmdbID,
			"show_title":     title,
			"season_number":  seasonNumber,
			"episode_number": episodeNumber,
			"episode_name":   episodeName,
			"watched_at":     watchedAt.UTC().Format(time.RFC3339),
			"media_type":     "tv",
		})
	}

	movieRows, err := h.db.Query(context.Background(), `
		SELECT m.tmdb_id, m.title, um.watched_at
		FROM user_movies um
		JOIN movies m ON m.id = um.movie_id
		WHERE um.user_id = $1::uuid AND um.watched = true
		ORDER BY um.watched_at DESC`, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer movieRows.Close()

	for movieRows.Next() {
		var tmdbID int
		var title string
		var watchedAt *time.Time
		if err := movieRows.Scan(&tmdbID, &title, &watchedAt); err != nil || watchedAt == nil {
			continue
		}
		items = append(items, fiber.Map{
			"tmdb_id":    tmdbID,
			"show_title": title,
			"watched_at": watchedAt.UTC().Format(time.RFC3339),
			"media_type": "movie",
		})
	}
	return c.JSON(fiber.Map{
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"format":      "showtrack-watch-history-v1",
		"items":       items,
	})
}

func (h *Handler) ImportWatchHistory(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)

	items, err := parseImportPayload(c.Body())
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if len(items) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "no importable items found")
	}

	imported := 0
	skipped := 0
	errors := []string{}

	for _, item := range items {
		if item.MediaType == "movie" {
			if ok, err := h.importMovieItem(userID, item); err != nil {
				skipped++
				if len(errors) < 10 {
					errors = append(errors, fmt.Sprintf("movie import failed: %s", firstNonEmpty(item.ShowTitle, strconv.Itoa(item.TMDBID))))
				}
				continue
			} else if ok {
				imported++
			} else {
				skipped++
				if len(errors) < 10 {
					errors = append(errors, fmt.Sprintf("movie not found: %s", firstNonEmpty(item.ShowTitle, strconv.Itoa(item.TMDBID))))
				}
			}
			continue
		}

		showID, err := h.resolveImportedShow(item)
		if err != nil || showID == 0 {
			skipped++
			if len(errors) < 10 {
				errors = append(errors, fmt.Sprintf("show not found: %s", firstNonEmpty(item.ShowTitle, strconv.Itoa(item.TMDBID))))
			}
			continue
		}

		_, _ = h.db.Exec(context.Background(),
			`INSERT INTO user_shows (user_id, show_id) VALUES ($1::uuid, $2) ON CONFLICT DO NOTHING`,
			userID, showID,
		)

		var episodeID int
		err = h.db.QueryRow(context.Background(), `
			SELECT e.id
			FROM episodes e
			JOIN seasons s ON s.id = e.season_id
			WHERE s.show_id = $1 AND s.season_number = $2 AND e.episode_number = $3`,
			showID, item.SeasonNumber, item.EpisodeNumber,
		).Scan(&episodeID)
		if err != nil {
			skipped++
			if len(errors) < 10 {
				errors = append(errors, fmt.Sprintf("episode missing: %s S%dE%d", firstNonEmpty(item.ShowTitle, strconv.Itoa(item.TMDBID)), item.SeasonNumber, item.EpisodeNumber))
			}
			continue
		}

		watchedAt := time.Now().UTC()
		if item.WatchedAt != nil {
			watchedAt = item.WatchedAt.UTC()
		}
		_, err = h.db.Exec(context.Background(), `
			INSERT INTO user_episodes (user_id, episode_id, watched_at)
			VALUES ($1::uuid, $2, $3)
			ON CONFLICT (user_id, episode_id) DO NOTHING`,
			userID, episodeID, watchedAt,
		)
		if err != nil {
			skipped++
			if len(errors) < 10 {
				errors = append(errors, fmt.Sprintf("failed import: %s S%dE%d", firstNonEmpty(item.ShowTitle, strconv.Itoa(item.TMDBID)), item.SeasonNumber, item.EpisodeNumber))
			}
			continue
		}
		imported++
	}

	return c.JSON(fiber.Map{
		"ok":       true,
		"imported": imported,
		"skipped":  skipped,
		"errors":   errors,
	})
}

func (h *Handler) AddShow(c *fiber.Ctx) error {
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

	var showID int
	err := h.db.QueryRow(context.Background(),
		`SELECT id FROM shows WHERE tmdb_id = $1`, body.TMDBID,
	).Scan(&showID)
	if err == pgx.ErrNoRows {
		if h.cfg.TMDBAPIKey != "" {
			show, tmdbErr := h.tmdb.GetTVShow(body.TMDBID)
			if tmdbErr != nil {
				return fiber.NewError(fiber.StatusNotFound, "show not found on tmdb")
			}
			genres, _ := json.Marshal(show.Genres)
			err = h.db.QueryRow(context.Background(), `
				INSERT INTO shows (tmdb_id, title, overview, poster_path, status, vote_average, genres)
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
				show.ID, show.Name, show.Overview, show.PosterPath, show.Status, show.VoteAverage, genres,
			).Scan(&showID)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
			if syncErr := h.syncShowDetails(showID, show); syncErr != nil {
				return fiber.NewError(fiber.StatusInternalServerError, syncErr.Error())
			}
		} else {
			return fiber.NewError(fiber.StatusNotFound, "show not found")
		}
	}

	_, err = h.db.Exec(context.Background(), `
		INSERT INTO user_shows (user_id, show_id, list_status) VALUES ($1::uuid, $2, $3)
		ON CONFLICT (user_id, show_id) DO UPDATE SET list_status = EXCLUDED.list_status`,
		userID, showID, listStatus,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.logShowAdded(userID, showID, listStatus)
	h.invalidateUserRecommendations(userID)
	return c.JSON(fiber.Map{"ok": true, "show_id": showID, "list_status": listStatus})
}

func (h *Handler) RemoveShow(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	showID := c.Params("id")
	_, err := h.db.Exec(context.Background(),
		`DELETE FROM user_shows WHERE user_id = $1::uuid AND show_id = $2`, userID, showID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) MarkWatched(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	epID := c.Params("id")
	_, err := h.db.Exec(context.Background(),
		`INSERT INTO user_episodes (user_id, episode_id) VALUES ($1::uuid, $2) ON CONFLICT DO NOTHING`,
		userID, epID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.logEpisodeWatched(userID, epID)
	h.invalidateUserRecommendations(userID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) UnmarkWatched(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	epID := c.Params("id")
	_, err := h.db.Exec(context.Background(),
		`DELETE FROM user_episodes WHERE user_id = $1::uuid AND episode_id = $2`, userID, epID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) RegisterDevice(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var body struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
		Version  string `json:"app_version"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	_, err := h.db.Exec(context.Background(), `
		INSERT INTO device_tokens (user_id, token, platform, app_version, last_used_at)
		VALUES ($1::uuid, $2, $3, $4, NOW())
		ON CONFLICT (user_id, token) DO UPDATE SET is_active = true, last_used_at = NOW()`,
		userID, body.Token, body.Platform, body.Version,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) upsertPerson(tmdbID int, name, profilePath string) (int, error) {
	var id int
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO persons (tmdb_id, name, profile_path)
		VALUES ($1, $2, $3)
		ON CONFLICT (tmdb_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, tmdbID, name, profilePath,
	).Scan(&id)
	return id, err
}

func (h *Handler) getShowCast(showID int) ([]fiber.Map, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT p.tmdb_id, p.name, sc.character_name,
		       COALESCE(p.profile_local, p.profile_path, '')
		FROM show_cast sc
		JOIN persons p ON p.id = sc.person_id
		WHERE sc.show_id = $1
		ORDER BY sc.cast_order LIMIT 20`, showID,
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

func (h *Handler) getShowSeasons(showID int, userID string) ([]seasonPayload, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT sn.id, sn.season_number, sn.name, sn.episode_count
		FROM seasons sn WHERE sn.show_id = $1 ORDER BY sn.season_number`, showID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	seasons := []seasonPayload{}
	for rows.Next() {
		var id, num, epCount int
		var name string
		if err := rows.Scan(&id, &num, &name, &epCount); err != nil {
			continue
		}
		episodes, _ := h.getSeasonEpisodes(id, userID)
		seasons = append(seasons, seasonPayload{
			ID:           id,
			SeasonNumber: num,
			Name:         name,
			EpisodeCount: epCount,
			Episodes:     episodes,
		})
	}
	return seasons, nil
}

func (h *Handler) getSeasonEpisodes(seasonID int, userID string) ([]episodePayload, error) {
	query := `
		SELECT e.id, e.episode_number, COALESCE(e.name, ''), COALESCE(e.overview, ''),
		       COALESCE(e.air_date::text, ''), COALESCE(e.runtime, 0), COALESCE(e.still_path, ''),
		       CASE WHEN ue.id IS NULL THEN false ELSE true END AS watched
		FROM episodes e
		LEFT JOIN user_episodes ue ON ue.episode_id = e.id AND ue.user_id = NULLIF($2, '')::uuid
		WHERE e.season_id = $1
		ORDER BY e.episode_number`
	rows, err := h.db.Query(context.Background(), query, seasonID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	episodes := []episodePayload{}
	for rows.Next() {
		var item episodePayload
		var stillPath string
		if err := rows.Scan(&item.ID, &item.EpisodeNumber, &item.Name, &item.Overview, &item.AirDate, &item.Runtime, &stillPath, &item.Watched); err != nil {
			continue
		}
		item.StillURL = h.resolvePoster(stillPath)
		episodes = append(episodes, item)
	}
	return episodes, nil
}

func (h *Handler) getPersonCredits(personID int) ([]fiber.Map, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT media_type, media_tmdb_id, title, character_name,
		       COALESCE(poster_local, poster_path, ''), vote_average
		FROM person_credits WHERE person_id = $1
		ORDER BY release_date DESC NULLS LAST`, personID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	credits := []fiber.Map{}
	for rows.Next() {
		var mediaType, title, character, poster string
		var mediaID int
		var vote float64
		if err := rows.Scan(&mediaType, &mediaID, &title, &character, &poster, &vote); err != nil {
			continue
		}
		credits = append(credits, fiber.Map{
			"media_type":  mediaType,
			"tmdb_id":     mediaID,
			"title":       title,
			"character":   character,
			"poster_url":  h.resolvePoster(poster),
			"vote_average": vote,
		})
	}
	return credits, nil
}

func (h *Handler) getLibraryItems(userID string, listStatus string) ([]fiber.Map, error) {
	statusFilter := ""
	args := []any{userID}
	if listStatus != "" && validListStatuses[listStatus] {
		statusFilter = " AND us.list_status = $2"
		args = append(args, listStatus)
	}
	rows, err := h.db.Query(context.Background(), `
		SELECT s.id, s.tmdb_id, s.title, s.status, s.vote_average,
		       COALESCE(s.poster_local, s.poster_path, ''),
		       COUNT(e.id) AS total_eps,
		       COUNT(ue.id) AS watched_eps,
		       us.list_status
		FROM user_shows us
		JOIN shows s ON s.id = us.show_id
		LEFT JOIN seasons sn ON sn.show_id = s.id
		LEFT JOIN episodes e ON e.season_id = sn.id
		LEFT JOIN user_episodes ue ON ue.episode_id = e.id AND ue.user_id = $1::uuid
		WHERE us.user_id = $1::uuid`+statusFilter+`
		GROUP BY s.id, us.list_status, us.added_at
		ORDER BY us.added_at DESC`, args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shows := []fiber.Map{}
	for rows.Next() {
		var id, tmdbID, totalEps, watchedEps int
		var title, status, poster, itemListStatus string
		var voteAverage float64
		if err := rows.Scan(&id, &tmdbID, &title, &status, &voteAverage, &poster, &totalEps, &watchedEps, &itemListStatus); err != nil {
			continue
		}
		progress := 0.0
		if totalEps > 0 {
			progress = float64(watchedEps) / float64(totalEps) * 100
		}
		shows = append(shows, fiber.Map{
			"id":           id,
			"tmdb_id":      tmdbID,
			"title":        title,
			"media_type":   "tv",
			"status":       status,
			"list_status":  itemListStatus,
			"vote_average": voteAverage,
			"poster_url":   h.resolvePoster(poster),
			"progress":     progress,
			"watched":      watchedEps,
			"total":        totalEps,
		})
	}
	return shows, nil
}

func (h *Handler) getActiveLibraryItems(userID string) ([]fiber.Map, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT s.id, s.tmdb_id, s.title, s.status, s.vote_average,
		       COALESCE(s.poster_local, s.poster_path, ''),
		       COUNT(e.id) AS total_eps,
		       COUNT(ue.id) AS watched_eps,
		       us.list_status
		FROM user_shows us
		JOIN shows s ON s.id = us.show_id
		LEFT JOIN seasons sn ON sn.show_id = s.id
		LEFT JOIN episodes e ON e.season_id = sn.id
		LEFT JOIN user_episodes ue ON ue.episode_id = e.id AND ue.user_id = $1::uuid
		WHERE us.user_id = $1::uuid
		  AND us.list_status IN ('watching', 'plan_to_watch')
		GROUP BY s.id, us.list_status, us.added_at
		ORDER BY us.added_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shows := []fiber.Map{}
	for rows.Next() {
		var id, tmdbID, totalEps, watchedEps int
		var title, status, poster, itemListStatus string
		var voteAverage float64
		if err := rows.Scan(&id, &tmdbID, &title, &status, &voteAverage, &poster, &totalEps, &watchedEps, &itemListStatus); err != nil {
			continue
		}
		progress := 0.0
		if totalEps > 0 {
			progress = float64(watchedEps) / float64(totalEps) * 100
		}
		shows = append(shows, fiber.Map{
			"id":           id,
			"tmdb_id":      tmdbID,
			"title":        title,
			"media_type":   "tv",
			"status":       status,
			"list_status":  itemListStatus,
			"vote_average": voteAverage,
			"poster_url":   h.resolvePoster(poster),
			"progress":     progress,
			"watched":      watchedEps,
			"total":        totalEps,
		})
	}
	return shows, nil
}

func (h *Handler) getUpcomingEpisodes(userID string) ([]upcomingPayload, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT s.title, s.tmdb_id, e.id, sn.season_number, e.episode_number,
		       COALESCE(e.name, ''), COALESCE(e.air_date::text, ''), COALESCE(s.poster_local, s.poster_path, '')
		FROM user_shows us
		JOIN shows s ON s.id = us.show_id
		JOIN seasons sn ON sn.show_id = s.id
		JOIN episodes e ON e.season_id = sn.id
		LEFT JOIN user_episodes ue ON ue.episode_id = e.id AND ue.user_id = $1::uuid
		WHERE us.user_id = $1::uuid
		  AND us.list_status IN ('watching', 'plan_to_watch')
		  AND ue.id IS NULL
		ORDER BY
		  CASE WHEN e.air_date IS NULL THEN 1 ELSE 0 END,
		  e.air_date ASC,
		  s.title ASC,
		  sn.season_number ASC,
		  e.episode_number ASC
		LIMIT 8`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []upcomingPayload{}
	for rows.Next() {
		var item upcomingPayload
		var posterPath string
		if err := rows.Scan(&item.ShowTitle, &item.ShowTMDBID, &item.EpisodeID, &item.SeasonNumber, &item.EpisodeNumber, &item.EpisodeName, &item.AirDate, &posterPath); err != nil {
			continue
		}
		item.PosterURL = h.resolvePoster(posterPath)
		items = append(items, item)
	}
	return items, nil
}

func (h *Handler) resolveImportedShow(item importItem) (int, error) {
	if item.TMDBID > 0 {
		var showID int
		err := h.db.QueryRow(context.Background(), `SELECT id FROM shows WHERE tmdb_id = $1`, item.TMDBID).Scan(&showID)
		if err == nil {
			return showID, nil
		}
		if err != pgx.ErrNoRows {
			return 0, err
		}
		if h.cfg.TMDBAPIKey == "" {
			return 0, pgx.ErrNoRows
		}
		show, tmdbErr := h.tmdb.GetTVShow(item.TMDBID)
		if tmdbErr != nil {
			return 0, tmdbErr
		}
		genres, _ := json.Marshal(show.Genres)
		err = h.db.QueryRow(context.Background(), `
			INSERT INTO shows (tmdb_id, title, overview, poster_path, backdrop_path, status, first_air_date, vote_average, genres)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (tmdb_id) DO UPDATE SET title = EXCLUDED.title
			RETURNING id`,
			show.ID, show.Name, show.Overview, show.PosterPath, show.BackdropPath,
			show.Status, nullDate(show.FirstAirDate), show.VoteAverage, genres,
		).Scan(&showID)
		if err != nil {
			return 0, err
		}
		if syncErr := h.syncShowDetails(showID, show); syncErr != nil {
			return 0, syncErr
		}
		return showID, nil
	}

	if item.ShowTitle != "" {
		var showID int
		err := h.db.QueryRow(context.Background(), `SELECT id FROM shows WHERE LOWER(title) = LOWER($1) LIMIT 1`, item.ShowTitle).Scan(&showID)
		if err == nil {
			return showID, nil
		}
	}

	return 0, pgx.ErrNoRows
}

func (h *Handler) posterURL(path string) string {
	if path == "" {
		return "/placeholder-poster.svg"
	}
	if path[0] == '/' {
		return h.cfg.MediaURL + path
	}
	return tmdb.PosterURL(path, "w500")
}

func (h *Handler) localPoster(path string) string {
	if path == "" {
		return ""
	}
	return h.cfg.MediaURL + path
}

func (h *Handler) resolvePoster(path string) string {
	if path == "" {
		return "/placeholder-poster.svg"
	}
	if path[0] == '/' {
		return h.cfg.MediaURL + path
	}
	return tmdb.PosterURL(path, "w500")
}

func nullDate(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func normalizeImportItems(items []map[string]any) []importItem {
	result := make([]importItem, 0, len(items))
	for _, item := range items {
		mediaType := getString(item, "media_type", "type")
		if mediaType == "" {
			if getInt(item, "season_number", "season", "s") == 0 && getInt(item, "episode_number", "episode", "e") == 0 {
				mediaType = "movie"
			} else {
				mediaType = "tv"
			}
		}
		parsed := importItem{
			TMDBID:        getInt(item, "tmdb_id", "show_tmdb_id", "series_tmdb_id", "movie_tmdb_id"),
			ShowTitle:     getString(item, "show_title", "series_name", "seriesName", "movie_title", "title", "name"),
			MediaType:     mediaType,
			SeasonNumber:  getInt(item, "season_number", "season", "s"),
			EpisodeNumber: getInt(item, "episode_number", "episode", "e"),
			WatchedAt:     getTime(item, "watched_at", "seen_at", "updated_at", "date"),
		}
		if parsed.MediaType == "movie" && (parsed.TMDBID > 0 || parsed.ShowTitle != "") {
			result = append(result, parsed)
			continue
		}
		if parsed.SeasonNumber > 0 && parsed.EpisodeNumber > 0 && (parsed.TMDBID > 0 || parsed.ShowTitle != "") {
			parsed.MediaType = "tv"
			result = append(result, parsed)
		}
	}
	return result
}

func getString(item map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return strings.TrimSpace(typed)
			}
		}
	}
	return ""
}

func getInt(item map[string]any, keys ...string) int {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed)
		case int:
			return typed
		case string:
			n, err := strconv.Atoi(strings.TrimSpace(typed))
			if err == nil {
				return n
			}
		}
	}
	return 0
}

func getTime(item map[string]any, keys ...string) *time.Time {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			continue
		}
		for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
			if parsed, parseErr := time.Parse(layout, text); parseErr == nil {
				return &parsed
			}
		}
	}
	return nil
}

func (h *Handler) optionalUserID(c *fiber.Ctx) string {
	header := c.Get("Authorization")
	if header == "" {
		return ""
	}
	tokenStr := header
	if len(header) > 7 && header[:7] == "Bearer " {
		tokenStr = header[7:]
	}
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	userID, _ := claims["sub"].(string)
	return userID
}

func (h *Handler) userHasShow(userID string, showID int) (bool, error) {
	var exists bool
	err := h.db.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM user_shows WHERE user_id = $1::uuid AND show_id = $2)`,
		userID, showID,
	).Scan(&exists)
	return exists, err
}

func (h *Handler) getUserProgress(userID string, showID int) (int, int, error) {
	var watched, total int
	err := h.db.QueryRow(context.Background(), `
		SELECT COUNT(ue.id), COUNT(e.id)
		FROM seasons sn
		JOIN episodes e ON e.season_id = sn.id
		LEFT JOIN user_episodes ue ON ue.episode_id = e.id AND ue.user_id = $1::uuid
		WHERE sn.show_id = $2`, userID, showID,
	).Scan(&watched, &total)
	return watched, total, err
}

func (h *Handler) syncShowDetails(showID int, show *tmdb.TVShow) error {
	ctx := context.Background()
	var err error
	for _, season := range show.Seasons {
		var seasonID int
		err = h.db.QueryRow(ctx, `
			INSERT INTO seasons (show_id, season_number, name, episode_count, poster_path)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (show_id, season_number) DO UPDATE
			SET name = EXCLUDED.name, episode_count = EXCLUDED.episode_count, poster_path = EXCLUDED.poster_path
			RETURNING id`,
			showID, season.SeasonNumber, season.Name, season.EpisodeCount, season.PosterPath,
		).Scan(&seasonID)
		if err != nil {
			return err
		}

		seasonDetails, seasonErr := h.tmdb.GetSeason(show.ID, season.SeasonNumber)
		if seasonErr != nil {
			continue
		}
		for _, episode := range seasonDetails.Episodes {
			_, err = h.db.Exec(ctx, `
				INSERT INTO episodes (season_id, episode_number, name, overview, air_date, still_path, runtime)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (season_id, episode_number) DO UPDATE
				SET name = EXCLUDED.name,
				    overview = EXCLUDED.overview,
				    air_date = EXCLUDED.air_date,
				    still_path = EXCLUDED.still_path,
				    runtime = EXCLUDED.runtime`,
				seasonID, episode.EpisodeNumber, episode.Name, episode.Overview, nullDate(episode.AirDate), episode.StillPath, episode.Runtime,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
