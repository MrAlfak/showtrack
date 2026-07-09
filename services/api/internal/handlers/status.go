package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

var validListStatuses = map[string]bool{
	"watching":       true,
	"plan_to_watch":  true,
	"watched":        true,
	"dropped":        true,
	"archived":       true,
}

func (h *Handler) UpdateShowStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	showID := c.Params("id")

	var body struct {
		ListStatus string `json:"list_status"`
	}
	if err := c.BodyParser(&body); err != nil || !validListStatuses[body.ListStatus] {
		return fiber.NewError(fiber.StatusBadRequest, "list_status must be watching, plan_to_watch, watched, dropped, or archived")
	}

	tag, err := h.db.Exec(context.Background(), `
		UPDATE user_shows SET list_status = $3
		WHERE user_id = $1::uuid AND show_id = $2`,
		userID, showID, body.ListStatus,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if tag.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "show not in library")
	}
	showIDInt, _ := strconv.Atoi(showID)
	h.logStatusChanged(userID, "tv", showIDInt, body.ListStatus)
	return c.JSON(fiber.Map{"ok": true, "list_status": body.ListStatus})
}

func (h *Handler) UpdateMovieStatus(c *fiber.Ctx) error {
	userID := c.Locals("userID").(string)
	movieID := c.Params("id")

	var body struct {
		ListStatus string `json:"list_status"`
	}
	if err := c.BodyParser(&body); err != nil || !validListStatuses[body.ListStatus] {
		return fiber.NewError(fiber.StatusBadRequest, "list_status must be watching, plan_to_watch, watched, dropped, or archived")
	}

	tag, err := h.db.Exec(context.Background(), `
		UPDATE user_movies SET list_status = $3
		WHERE user_id = $1::uuid AND movie_id = $2`,
		userID, movieID, body.ListStatus,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if tag.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "movie not in library")
	}
	movieIDInt, _ := strconv.Atoi(movieID)
	h.logStatusChanged(userID, "movie", movieIDInt, body.ListStatus)
	return c.JSON(fiber.Map{"ok": true, "list_status": body.ListStatus})
}

func (h *Handler) computeUserStats(userID string) (dashboardStatsPayload, error) {
	stats := dashboardStatsPayload{}

	err := h.db.QueryRow(context.Background(), `
		SELECT
			(SELECT COUNT(*) FROM user_shows WHERE user_id = $1::uuid AND list_status != 'archived'),
			(SELECT COUNT(*) FROM user_movies WHERE user_id = $1::uuid AND list_status != 'archived'),
			(SELECT COUNT(*) FROM user_episodes WHERE user_id = $1::uuid),
			(SELECT COUNT(*) FROM user_movies WHERE user_id = $1::uuid AND watched = true)`,
		userID,
	).Scan(&stats.Shows, &stats.Movies, &stats.Episodes, &stats.Total)
	if err != nil {
		return stats, err
	}

	var watchMinutes int
	err = h.db.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(COALESCE(e.runtime, 45)), 0)
		FROM user_episodes ue
		JOIN episodes e ON e.id = ue.episode_id
		WHERE ue.user_id = $1::uuid`, userID,
	).Scan(&watchMinutes)
	if err != nil {
		return stats, err
	}

	var movieMinutes int
	_ = h.db.QueryRow(context.Background(), `
		SELECT COALESCE(SUM(COALESCE(m.runtime, 120)), 0)
		FROM user_movies um
		JOIN movies m ON m.id = um.movie_id
		WHERE um.user_id = $1::uuid AND um.watched = true`, userID,
	).Scan(&movieMinutes)

	stats.Hours = (watchMinutes + movieMinutes) / 60
	stats.Streak, _ = h.computeWatchStreak(userID)
	_ = h.db.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM user_episodes
		WHERE user_id = $1::uuid AND watched_at >= CURRENT_DATE`, userID,
	).Scan(&stats.BingeToday)
	return stats, nil
}

func (h *Handler) computeWatchStreak(userID string) (int, error) {
	rows, err := h.db.Query(context.Background(), `
		SELECT DISTINCT DATE(watched_at AT TIME ZONE 'UTC') AS d
		FROM user_episodes
		WHERE user_id = $1::uuid
		ORDER BY d DESC
		LIMIT 400`, userID,
	)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			continue
		}
		dates = append(dates, d)
	}
	if len(dates) == 0 {
		return 0, nil
	}

	streak := 1
	for i := 1; i < len(dates); i++ {
		prev, err1 := parseDate(dates[i-1])
		curr, err2 := parseDate(dates[i])
		if err1 != nil || err2 != nil {
			break
		}
		if prev.Sub(curr).Hours() == 24 {
			streak++
			continue
		}
		break
	}
	return streak, nil
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
