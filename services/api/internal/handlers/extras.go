package handlers

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) getUserRating(userID, mediaType string, tmdbID int) (int, string, bool) {
	if userID == "" || tmdbID == 0 {
		return 0, "", false
	}
	var score int
	var review string
	err := h.db.QueryRow(context.Background(), `
		SELECT score, review FROM user_ratings
		WHERE user_id = $1::uuid AND media_type = $2 AND tmdb_id = $3`,
		userID, mediaType, tmdbID,
	).Scan(&score, &review)
	if err != nil {
		return 0, "", false
	}
	return score, review, true
}

func (h *Handler) GetMyRating(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	mediaType := c.Params("type")
	tmdbID, _ := strconv.Atoi(c.Params("tmdb_id"))
	if mediaType != "tv" && mediaType != "movie" {
		return fiber.NewError(fiber.StatusBadRequest, "type must be tv or movie")
	}
	if tmdbID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid tmdb_id")
	}
	score, review, ok := h.getUserRating(userID, mediaType, tmdbID)
	if !ok {
		return c.JSON(fiber.Map{"rated": false})
	}
	return c.JSON(fiber.Map{"rated": true, "score": score, "review": review})
}

func (h *Handler) SetMyRating(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	mediaType := c.Params("type")
	tmdbID, _ := strconv.Atoi(c.Params("tmdb_id"))
	if mediaType != "tv" && mediaType != "movie" {
		return fiber.NewError(fiber.StatusBadRequest, "type must be tv or movie")
	}
	if tmdbID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid tmdb_id")
	}

	var body struct {
		Score  int    `json:"score"`
		Review string `json:"review"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if body.Score < 1 || body.Score > 10 {
		return fiber.NewError(fiber.StatusBadRequest, "score must be 1-10")
	}
	review := strings.TrimSpace(body.Review)
	if len(review) > 2000 {
		return fiber.NewError(fiber.StatusBadRequest, "review too long")
	}

	_, err := h.db.Exec(context.Background(), `
		INSERT INTO user_ratings (user_id, media_type, tmdb_id, score, review, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, NOW())
		ON CONFLICT (user_id, media_type, tmdb_id) DO UPDATE SET
			score = EXCLUDED.score,
			review = EXCLUDED.review,
			updated_at = NOW()`,
		userID, mediaType, tmdbID, body.Score, review,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.invalidateUserRecommendations(userID)
	return c.JSON(fiber.Map{"ok": true, "score": body.Score, "review": review})
}

func (h *Handler) DeleteMyRating(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	mediaType := c.Params("type")
	tmdbID, _ := strconv.Atoi(c.Params("tmdb_id"))
	_, err := h.db.Exec(context.Background(), `
		DELETE FROM user_ratings WHERE user_id = $1::uuid AND media_type = $2 AND tmdb_id = $3`,
		userID, mediaType, tmdbID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) GetMyLists(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	rows, err := h.db.Query(context.Background(), `
		SELECT l.id::text, l.name, l.description, l.is_public,
		       (SELECT COUNT(*) FROM custom_list_items i WHERE i.list_id = l.id)
		FROM custom_lists l
		WHERE l.user_id = $1::uuid
		ORDER BY l.updated_at DESC`, userID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	lists := make([]fiber.Map, 0)
	for rows.Next() {
		var id, name, description string
		var isPublic bool
		var itemCount int
		if err := rows.Scan(&id, &name, &description, &isPublic, &itemCount); err != nil {
			continue
		}
		lists = append(lists, fiber.Map{
			"id":          id,
			"name":        name,
			"description": description,
			"is_public":   isPublic,
			"item_count":  itemCount,
		})
	}
	return c.JSON(fiber.Map{"lists": lists})
}

func (h *Handler) CreateList(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}
	if err := c.BodyParser(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name required")
	}
	name := strings.TrimSpace(body.Name)
	if len(name) > 80 {
		return fiber.NewError(fiber.StatusBadRequest, "name too long")
	}
	var listID string
	err := h.db.QueryRow(context.Background(), `
		INSERT INTO custom_lists (user_id, name, description, is_public)
		VALUES ($1::uuid, $2, $3, $4) RETURNING id::text`,
		userID, name, strings.TrimSpace(body.Description), body.IsPublic,
	).Scan(&listID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true, "id": listID})
}

func (h *Handler) GetList(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	listID := c.Params("id")
	list, items, err := h.loadList(listID, userID)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "list not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"list": list, "items": items})
}

func (h *Handler) DeleteList(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	listID := c.Params("id")
	tag, err := h.db.Exec(context.Background(),
		`DELETE FROM custom_lists WHERE id = $1::uuid AND user_id = $2::uuid`, listID, userID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if tag.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "list not found")
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) AddListItem(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	listID := c.Params("id")
	var body struct {
		MediaType string `json:"media_type"`
		TMDBID    int    `json:"tmdb_id"`
		Title     string `json:"title"`
		PosterURL string `json:"poster_url"`
	}
	if err := c.BodyParser(&body); err != nil || body.TMDBID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "tmdb_id required")
	}
	if body.MediaType != "tv" && body.MediaType != "movie" {
		body.MediaType = "tv"
	}

	var owner string
	err := h.db.QueryRow(context.Background(),
		`SELECT user_id::text FROM custom_lists WHERE id = $1::uuid`, listID,
	).Scan(&owner)
	if err == pgx.ErrNoRows || owner != userID {
		return fiber.NewError(fiber.StatusNotFound, "list not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	posterPath := body.PosterURL
	_, err = h.db.Exec(context.Background(), `
		INSERT INTO custom_list_items (list_id, media_type, tmdb_id, title, poster_path)
		VALUES ($1::uuid, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`,
		listID, body.MediaType, body.TMDBID, body.Title, posterPath,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	_, _ = h.db.Exec(context.Background(), `UPDATE custom_lists SET updated_at = NOW() WHERE id = $1::uuid`, listID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) RemoveListItem(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	listID := c.Params("id")
	mediaType := c.Query("media_type", "tv")
	tmdbID, _ := strconv.Atoi(c.Query("tmdb_id"))
	if tmdbID == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "tmdb_id required")
	}

	var owner string
	err := h.db.QueryRow(context.Background(),
		`SELECT user_id::text FROM custom_lists WHERE id = $1::uuid`, listID,
	).Scan(&owner)
	if err == pgx.ErrNoRows || owner != userID {
		return fiber.NewError(fiber.StatusNotFound, "list not found")
	}

	_, err = h.db.Exec(context.Background(), `
		DELETE FROM custom_list_items WHERE list_id = $1::uuid AND media_type = $2 AND tmdb_id = $3`,
		listID, mediaType, tmdbID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) loadList(listID, userID string) (fiber.Map, []fiber.Map, error) {
	var name, description, owner string
	var isPublic bool
	err := h.db.QueryRow(context.Background(), `
		SELECT name, description, is_public, user_id::text
		FROM custom_lists WHERE id = $1::uuid`, listID,
	).Scan(&name, &description, &isPublic, &owner)
	if err != nil {
		return nil, nil, err
	}
	if owner != userID && !isPublic {
		return nil, nil, pgx.ErrNoRows
	}

	rows, err = h.db.Query(context.Background(), `
		SELECT media_type, tmdb_id, title, poster_path
		FROM custom_list_items WHERE list_id = $1::uuid ORDER BY added_at DESC`, listID,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	items := make([]fiber.Map, 0)
	for rows.Next() {
		var mediaType, title, posterPath string
		var tmdbID int
		if err := rows.Scan(&mediaType, &tmdbID, &title, &posterPath); err != nil {
			continue
		}
		poster := posterPath
		if poster != "" && !strings.HasPrefix(poster, "http") {
			poster = h.resolvePoster(posterPath)
		}
		items = append(items, fiber.Map{
			"media_type": mediaType,
			"tmdb_id":    tmdbID,
			"title":      title,
			"poster_url": poster,
		})
	}
	return fiber.Map{
		"id":          listID,
		"name":        name,
		"description": description,
		"is_public":   isPublic,
	}, items, nil
}
