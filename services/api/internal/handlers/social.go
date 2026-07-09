package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

func normalizeUsername(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	var b strings.Builder
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		case r == ' ' || r == '-' || r == '.':
			b.WriteRune('_')
		}
	}
	return b.String()
}

func (h *Handler) GetMyProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	profile, err := h.loadUserProfile(userID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(profile)
}

func (h *Handler) UpdateMyProfile(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	var body struct {
		Username    *string `json:"username"`
		DisplayName *string `json:"display_name"`
		Bio         *string `json:"bio"`
		IsPublic    *bool   `json:"is_public"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}

	if body.Username != nil {
		username := normalizeUsername(*body.Username)
		if !usernamePattern.MatchString(username) {
			return fiber.NewError(fiber.StatusBadRequest, "username must be 3-30 chars: a-z, 0-9, underscore")
		}
		var existing string
		err := h.db.QueryRow(context.Background(),
			`SELECT id::text FROM users WHERE username = $1 AND id <> $2::uuid`, username, userID,
		).Scan(&existing)
		if err == nil {
			return fiber.NewError(fiber.StatusConflict, "username taken")
		}
		if err != pgx.ErrNoRows {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		_, err = h.db.Exec(context.Background(), `UPDATE users SET username = $2 WHERE id = $1::uuid`, userID, username)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	if body.DisplayName != nil {
		_, _ = h.db.Exec(context.Background(), `UPDATE users SET display_name = $2 WHERE id = $1::uuid`, userID, strings.TrimSpace(*body.DisplayName))
	}
	if body.Bio != nil {
		bio := strings.TrimSpace(*body.Bio)
		if len(bio) > 280 {
			return fiber.NewError(fiber.StatusBadRequest, "bio too long")
		}
		_, _ = h.db.Exec(context.Background(), `UPDATE users SET bio = $2 WHERE id = $1::uuid`, userID, bio)
	}
	if body.IsPublic != nil {
		_, _ = h.db.Exec(context.Background(), `UPDATE users SET is_public = $2 WHERE id = $1::uuid`, userID, *body.IsPublic)
	}

	profile, err := h.loadUserProfile(userID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(profile)
}

func (h *Handler) SearchUsers(c *fiber.Ctx) error {
	viewerID := c.Locals("userID").(string)
	query := strings.TrimSpace(c.Query("q"))
	if len(query) < 2 {
		return c.JSON(fiber.Map{"results": []fiber.Map{}})
	}
	pattern := "%" + query + "%"
	rows, err := h.db.Query(context.Background(), `
		SELECT u.id::text, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, ''),
		       EXISTS(SELECT 1 FROM user_follows f WHERE f.follower_id = $1::uuid AND f.following_id = u.id)
		FROM users u
		WHERE u.id <> $1::uuid
		  AND u.is_public = true
		  AND (u.username ILIKE $2 OR u.display_name ILIKE $2 OR u.email ILIKE $2)
		ORDER BY u.display_name NULLS LAST
		LIMIT 20`, viewerID, pattern,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	results := make([]fiber.Map, 0)
	for rows.Next() {
		var id, username, displayName, avatarURL string
		var isFollowing bool
		if err := rows.Scan(&id, &username, &displayName, &avatarURL, &isFollowing); err != nil {
			continue
		}
		results = append(results, fiber.Map{
			"id":           id,
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
			"is_following": isFollowing,
		})
	}
	return c.JSON(fiber.Map{"results": results})
}

func (h *Handler) GetUserProfile(c *fiber.Ctx) error {
	viewerID := c.Locals("userID").(string)
	targetID := c.Params("id")
	profile, err := h.loadUserProfile(targetID, viewerID)
	if err == pgx.ErrNoRows {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if targetID != viewerID {
		isPublic, _ := profile["is_public"].(bool)
		isFollowing, _ := profile["is_following"].(bool)
		if !isPublic && !isFollowing {
			return fiber.NewError(fiber.StatusForbidden, "profile is private")
		}
	}
	return c.JSON(profile)
}

func (h *Handler) FollowUser(c *fiber.Ctx) error {
	followerID := c.Locals("userID").(string)
	followingID := c.Params("id")
	if followerID == followingID {
		return fiber.NewError(fiber.StatusBadRequest, "cannot follow yourself")
	}
	var exists bool
	err := h.db.QueryRow(context.Background(), `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1::uuid)`, followingID).Scan(&exists)
	if err != nil || !exists {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	_, err = h.db.Exec(context.Background(), `
		INSERT INTO user_follows (follower_id, following_id) VALUES ($1::uuid, $2::uuid)
		ON CONFLICT DO NOTHING`, followerID, followingID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.invalidateUserRecommendations(followerID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) UnfollowUser(c *fiber.Ctx) error {
	followerID := c.Locals("userID").(string)
	followingID := c.Params("id")
	_, err := h.db.Exec(context.Background(),
		`DELETE FROM user_follows WHERE follower_id = $1::uuid AND following_id = $2::uuid`,
		followerID, followingID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	h.invalidateUserRecommendations(followerID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) GetFollowing(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	users, err := h.loadFollowList(userID, "following")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"results": users})
}

func (h *Handler) GetFollowers(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	users, err := h.loadFollowList(userID, "followers")
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"results": users})
}

func (h *Handler) GetFeed(c *fiber.Ctx) error {
	viewerID := c.Locals("userID").(string)
	limit := c.QueryInt("limit", 40)
	if limit < 1 || limit > 100 {
		limit = 40
	}
	beforeID := c.QueryInt("before_id", 0)

	query := `
		SELECT a.id, a.user_id::text, a.activity_type, a.payload, a.created_at,
		       COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, '')
		FROM activities a
		JOIN users u ON u.id = a.user_id
		WHERE (
			a.user_id = $1::uuid
			OR a.user_id IN (SELECT following_id FROM user_follows WHERE follower_id = $1::uuid)
		)`
	args := []any{viewerID}
	argPos := 2
	if beforeID > 0 {
		query += fmt.Sprintf(` AND a.id < $%d`, argPos)
		args = append(args, beforeID)
		argPos++
	}
	query += fmt.Sprintf(` ORDER BY a.id DESC LIMIT $%d`, argPos)
	args = append(args, limit)

	rows, err := h.db.Query(context.Background(), query, args...)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	items := make([]fiber.Map, 0)
	for rows.Next() {
		var id int64
		var userID, activityType, username, displayName, avatarURL string
		var payload []byte
		var createdAt any
		if err := rows.Scan(&id, &userID, &activityType, &payload, &createdAt, &username, &displayName, &avatarURL); err != nil {
			continue
		}
		var payloadMap map[string]any
		_ = json.Unmarshal(payload, &payloadMap)
		items = append(items, fiber.Map{
			"id":            id,
			"user_id":       userID,
			"activity_type": activityType,
			"payload":       payloadMap,
			"created_at":    createdAt,
			"username":      username,
			"display_name":  displayName,
			"avatar_url":    avatarURL,
		})
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) loadUserProfile(targetID, viewerID string) (fiber.Map, error) {
	var username, displayName, bio, avatarURL string
	var isPublic bool
	err := h.db.QueryRow(context.Background(), `
		SELECT COALESCE(username, ''), COALESCE(display_name, ''), COALESCE(bio, ''), COALESCE(avatar_url, ''), is_public
		FROM users WHERE id = $1::uuid`, targetID,
	).Scan(&username, &displayName, &bio, &avatarURL, &isPublic)
	if err != nil {
		return nil, err
	}

	stats, _ := h.computeUserStats(targetID)
	var followers, following int
	_ = h.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM user_follows WHERE following_id = $1::uuid`, targetID).Scan(&followers)
	_ = h.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM user_follows WHERE follower_id = $1::uuid`, targetID).Scan(&following)

	isFollowing := false
	if viewerID != "" && viewerID != targetID {
		_ = h.db.QueryRow(context.Background(), `
			SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1::uuid AND following_id = $2::uuid)`,
			viewerID, targetID,
		).Scan(&isFollowing)
	}

	profile := fiber.Map{
		"id":           targetID,
		"username":     username,
		"display_name": displayName,
		"bio":          bio,
		"avatar_url":   avatarURL,
		"is_public":    isPublic,
		"stats":        stats,
		"followers":    followers,
		"following":    following,
		"is_following": isFollowing,
	}

	if targetID == viewerID || isPublic || isFollowing {
		preview, _ := h.loadLibraryPreview(targetID)
		profile["library_preview"] = preview
	}
	return profile, nil
}

func (h *Handler) loadLibraryPreview(userID string) ([]fiber.Map, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT s.tmdb_id, s.title, COALESCE(s.poster_path, ''), 'tv' AS media_type
		FROM user_shows us
		JOIN shows s ON s.id = us.show_id
		WHERE us.user_id = $1::uuid AND us.list_status != 'archived'
		UNION ALL
		SELECT m.tmdb_id, m.title, COALESCE(m.poster_path, ''), 'movie' AS media_type
		FROM user_movies um
		JOIN movies m ON m.id = um.movie_id
		WHERE um.user_id = $1::uuid AND um.list_status != 'archived'
		LIMIT 12`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]fiber.Map, 0)
	for rows.Next() {
		var tmdbID int
		var title, posterPath, mediaType string
		if err := rows.Scan(&tmdbID, &title, &posterPath, &mediaType); err != nil {
			continue
		}
		items = append(items, fiber.Map{
			"tmdb_id":    tmdbID,
			"title":      title,
			"poster_url": h.resolvePoster(posterPath),
			"media_type": mediaType,
		})
	}
	return items, nil
}

func (h *Handler) loadFollowList(userID, mode string) ([]fiber.Map, error) {
	var query string
	if mode == "followers" {
		query = `
			SELECT u.id::text, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, '')
			FROM user_follows f
			JOIN users u ON u.id = f.follower_id
			WHERE f.following_id = $1::uuid
			ORDER BY f.created_at DESC`
	} else {
		query = `
			SELECT u.id::text, COALESCE(u.username, ''), COALESCE(u.display_name, ''), COALESCE(u.avatar_url, '')
			FROM user_follows f
			JOIN users u ON u.id = f.following_id
			WHERE f.follower_id = $1::uuid
			ORDER BY f.created_at DESC`
	}
	rows, err := h.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]fiber.Map, 0)
	for rows.Next() {
		var id, username, displayName, avatarURL string
		if err := rows.Scan(&id, &username, &displayName, &avatarURL); err != nil {
			continue
		}
		results = append(results, fiber.Map{
			"id":           id,
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
	}
	return results, nil
}
