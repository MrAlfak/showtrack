package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) RequireAdmin(c *fiber.Ctx) error {
	if h.cfg.AdminSecret == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "admin not configured")
	}
	if c.Get("X-Admin-Token") != h.cfg.AdminSecret {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid admin token")
	}
	return c.Next()
}

func (h *Handler) AdminStats(c *fiber.Ctx) error {
	var users, shows, movies, episodes, activities int
	_ = h.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM users`).Scan(&users)
	_ = h.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM shows`).Scan(&shows)
	_ = h.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM movies`).Scan(&movies)
	_ = h.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM user_episodes`).Scan(&episodes)
	_ = h.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM activities`).Scan(&activities)

	return c.JSON(fiber.Map{
		"users":      users,
		"shows":      shows,
		"movies":     movies,
		"watched":    episodes,
		"activities": activities,
	})
}

func (h *Handler) AdminUsers(c *fiber.Ctx) error {
	rows, err := h.db.Query(context.Background(), `
		SELECT id::text, email, COALESCE(username, ''), COALESCE(display_name, ''), created_at
		FROM users ORDER BY created_at DESC LIMIT 100`,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	users := make([]fiber.Map, 0)
	for rows.Next() {
		var id, email, username, displayName string
		var createdAt any
		if err := rows.Scan(&id, &email, &username, &displayName, &createdAt); err != nil {
			continue
		}
		users = append(users, fiber.Map{
			"id":           id,
			"email":        email,
			"username":     username,
			"display_name": displayName,
			"created_at":   createdAt,
		})
	}
	return c.JSON(fiber.Map{"users": users})
}
