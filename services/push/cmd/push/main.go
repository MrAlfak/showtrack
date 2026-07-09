package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type NotificationJob struct {
	UserID  string `json:"user_id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	ShowID  int    `json:"show_id"`
	Episode int    `json:"episode_id"`
}

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	dbURL := env("DATABASE_URL", "postgres://showtrack:showtrack@postgres:5432/showtrack?sslmode=disable")
	redisURL := env("REDIS_URL", "redis://redis:6379")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	opts, _ := redis.ParseURL(redisURL)
	rdb := redis.NewClient(opts)

	log.Println("push service started")

	checkNewEpisodes(ctx, pool, rdb)
	go runEpisodeChecker(ctx, pool, rdb)

	for {
		result, err := rdb.BRPop(ctx, 5*time.Second, "notifications").Result()
		if err != nil {
			continue
		}
		if len(result) < 2 {
			continue
		}

		var job NotificationJob
		if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
			continue
		}
		if err := sendNotification(ctx, pool, job); err != nil {
			log.Printf("send failed: %v", err)
		}
	}
}

func runEpisodeChecker(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client) {
	interval := envDuration("EPISODE_CHECK_INTERVAL", time.Hour)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		checkNewEpisodes(ctx, pool, rdb)
	}
}

func envDuration(key string, fallback time.Duration) time.Duration {
	raw := env(key, "")
	if raw == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func checkNewEpisodes(ctx context.Context, pool *pgxpool.Pool, rdb *redis.Client) {
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT us.user_id::text, sh.title, sh.id, e.id, e.name, sn.season_number, e.episode_number
		FROM episodes e
		JOIN seasons sn ON sn.id = e.season_id
		JOIN shows sh ON sh.id = sn.show_id
		JOIN user_shows us ON us.show_id = sh.id
		WHERE e.air_date = CURRENT_DATE AND e.notified = false`)
	if err != nil {
		log.Printf("episode check: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID, showTitle, epName string
		var showID, epID, seasonNum, epNum int
		if err := rows.Scan(&userID, &showTitle, &showID, &epID, &epName, &seasonNum, &epNum); err != nil {
			continue
		}
		job := NotificationJob{
			UserID:  userID,
			Title:   fmt.Sprintf("New episode: %s", showTitle),
			Body:    fmt.Sprintf("S%dE%d — %s", seasonNum, epNum, epName),
			ShowID:  showID,
			Episode: epID,
		}
		data, _ := json.Marshal(job)
		rdb.LPush(ctx, "notifications", data)
		_, _ = pool.Exec(ctx, `UPDATE episodes SET notified = true WHERE id = $1`, epID)
	}
}

func sendNotification(ctx context.Context, pool *pgxpool.Pool, job NotificationJob) error {
	rows, err := pool.Query(ctx, `
		SELECT token, platform FROM device_tokens
		WHERE user_id = $1::uuid AND is_active = true`, job.UserID)
	if err != nil {
		return err
	}
	defer rows.Close()

	sent := false
	for rows.Next() {
		var token, platform string
		if err := rows.Scan(&token, &platform); err != nil {
			continue
		}
		switch platform {
		case "android", "web":
			if err := sendFCM(token, job); err != nil {
				log.Printf("fcm error: %v", err)
				deactivateToken(ctx, pool, token)
			} else {
				sent = true
			}
		case "ios":
			if err := sendAPNs(token, job); err != nil {
				log.Printf("apns error: %v", err)
				deactivateToken(ctx, pool, token)
			} else {
				sent = true
			}
		}
	}

	status := "failed"
	if sent {
		status = "sent"
	}
	_, _ = pool.Exec(ctx, `
		INSERT INTO notification_log (user_id, episode_id, title, body, status, sent_at)
		VALUES ($1::uuid, $2, $3, $4, $5, NOW())`,
		job.UserID, job.Episode, job.Title, job.Body, status)

	return nil
}

func deactivateToken(ctx context.Context, pool *pgxpool.Pool, token string) {
	_, _ = pool.Exec(ctx, `UPDATE device_tokens SET is_active = false WHERE token = $1`, token)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
