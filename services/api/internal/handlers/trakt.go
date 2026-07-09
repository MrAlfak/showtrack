package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/showtrack/api/internal/trakt"
)

func (h *Handler) traktClient() *trakt.Client {
	return trakt.New(h.cfg.TraktClientID, h.cfg.TraktClientSecret, h.cfg.TraktRedirectURI)
}

func (h *Handler) GetTraktStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var username string
	err := h.db.QueryRow(context.Background(),
		`SELECT trakt_user FROM trakt_accounts WHERE user_id = $1::uuid`, userID,
	).Scan(&username)
	if err == pgx.ErrNoRows {
		return c.JSON(fiber.Map{"connected": false})
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"connected": true, "username": username})
}

func (h *Handler) StartTraktConnect(c *fiber.Ctx) error {
	client := h.traktClient()
	if !client.Enabled() {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trakt not configured")
	}
	userID := c.Locals("userID").(string)
	state := uuid.NewString()
	_ = h.redis.Set(context.Background(), "trakt:state:"+state, userID, 10*time.Minute).Err()
	return c.JSON(fiber.Map{
		"url": client.AuthorizeURL(state),
	})
}

func (h *Handler) TraktCallback(c *fiber.Ctx) error {
	client := h.traktClient()
	if !client.Enabled() {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trakt not configured")
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing code or state")
	}

	userID, err := h.redis.Get(context.Background(), "trakt:state:"+state).Result()
	if err != nil || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid or expired state")
	}
	_ = h.redis.Del(context.Background(), "trakt:state:"+state).Err()

	token, err := client.ExchangeCode(code)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	settings, err := client.GetSettings(token.AccessToken)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	_, err = h.db.Exec(context.Background(), `
		INSERT INTO trakt_accounts (user_id, trakt_user, access_token, refresh_token, expires_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			trakt_user = EXCLUDED.trakt_user,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()`,
		userID, settings.User.Username, token.AccessToken, token.RefreshToken, expiresAt,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	redirect := fmt.Sprintf("%s/profile?trakt=connected", h.cfg.PublicWebURL)
	return c.Redirect(redirect, fiber.StatusTemporaryRedirect)
}

func (h *Handler) DisconnectTrakt(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	_, err := h.db.Exec(context.Background(), `DELETE FROM trakt_accounts WHERE user_id = $1::uuid`, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) SyncTrakt(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	client := h.traktClient()
	if !client.Enabled() {
		return fiber.NewError(fiber.StatusServiceUnavailable, "trakt not configured")
	}

	accessToken, err := h.ensureTraktToken(userID, client)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	imported := 0
	skipped := 0

	for page := 1; page <= 5; page++ {
		episodes, err := client.GetEpisodeHistory(accessToken, page)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		if len(episodes) == 0 {
			break
		}
		for _, item := range episodes {
			if item.Show.IDs.TMDB == 0 {
				skipped++
				continue
			}
			watchedAt := item.WatchedAt
			ok, err := h.importTraktEpisode(userID, importItem{
				TMDBID:        item.Show.IDs.TMDB,
				ShowTitle:     item.Show.Title,
				MediaType:     "tv",
				SeasonNumber:  item.Episode.Season,
				EpisodeNumber: item.Episode.Number,
				WatchedAt:     &watchedAt,
			})
			if err != nil || !ok {
				skipped++
				continue
			}
			imported++
		}
		if len(episodes) < 100 {
			break
		}
	}

	for page := 1; page <= 3; page++ {
		movies, err := client.GetMovieHistory(accessToken, page)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, err.Error())
		}
		if len(movies) == 0 {
			break
		}
		for _, item := range movies {
			if item.Movie.IDs.TMDB == 0 {
				skipped++
				continue
			}
			watchedAt := item.WatchedAt
			ok, err := h.importMovieItem(userID, importItem{
				TMDBID:    item.Movie.IDs.TMDB,
				ShowTitle: item.Movie.Title,
				MediaType: "movie",
				WatchedAt: &watchedAt,
			})
			if err != nil || !ok {
				skipped++
				continue
			}
			imported++
		}
		if len(movies) < 100 {
			break
		}
	}

	return c.JSON(fiber.Map{"ok": true, "imported": imported, "skipped": skipped})
}

func (h *Handler) ensureTraktToken(userID string, client *trakt.Client) (string, error) {
	var accessToken, refreshToken string
	var expiresAt time.Time
	err := h.db.QueryRow(context.Background(), `
		SELECT access_token, refresh_token, expires_at
		FROM trakt_accounts WHERE user_id = $1::uuid`, userID,
	).Scan(&accessToken, &refreshToken, &expiresAt)
	if err == pgx.ErrNoRows {
		return "", fmt.Errorf("trakt not connected")
	}
	if err != nil {
		return "", err
	}
	if time.Now().Before(expiresAt.Add(-5 * time.Minute)) {
		return accessToken, nil
	}
	token, err := client.RefreshToken(refreshToken)
	if err != nil {
		return "", err
	}
	newExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	_, _ = h.db.Exec(context.Background(), `
		UPDATE trakt_accounts SET access_token = $2, refresh_token = $3, expires_at = $4, updated_at = NOW()
		WHERE user_id = $1::uuid`,
		userID, token.AccessToken, token.RefreshToken, newExpiry,
	)
	return token.AccessToken, nil
}

func (h *Handler) importTraktEpisode(userID string, item importItem) (bool, error) {
	showID, err := h.resolveImportedShow(item)
	if err != nil || showID == 0 {
		return false, err
	}
	_, _ = h.db.Exec(context.Background(),
		`INSERT INTO user_shows (user_id, show_id, list_status) VALUES ($1::uuid, $2, 'watching') ON CONFLICT DO NOTHING`,
		userID, showID,
	)
	var episodeID int
	err = h.db.QueryRow(context.Background(), `
		SELECT e.id FROM episodes e
		JOIN seasons s ON s.id = e.season_id
		WHERE s.show_id = $1 AND s.season_number = $2 AND e.episode_number = $3`,
		showID, item.SeasonNumber, item.EpisodeNumber,
	).Scan(&episodeID)
	if err != nil {
		return false, err
	}
	watchedAt := time.Now().UTC()
	if item.WatchedAt != nil {
		watchedAt = item.WatchedAt.UTC()
	}
	_, err = h.db.Exec(context.Background(), `
		INSERT INTO user_episodes (user_id, episode_id, watched_at)
		VALUES ($1::uuid, $2, $3) ON CONFLICT DO NOTHING`,
		userID, episodeID, watchedAt,
	)
	return err == nil, err
}
