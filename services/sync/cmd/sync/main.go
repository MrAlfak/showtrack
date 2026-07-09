package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/image/webp"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	dbURL := env("DATABASE_URL", "postgres://showtrack:showtrack@postgres:5432/showtrack?sslmode=disable")
	mediaDir := env("MEDIA_DIR", "/data")
	tmdbKey := env("TMDB_API_KEY", "")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	log.Println("sync started")
	syncPosters(ctx, pool, mediaDir, tmdbKey)
	syncMoviePosters(ctx, pool, mediaDir, tmdbKey)
	syncAirDates(ctx, pool, tmdbKey)
	log.Println("sync completed")
}

func syncMoviePosters(ctx context.Context, pool *pgxpool.Pool, mediaDir, tmdbKey string) {
	rows, err := pool.Query(ctx, `
		SELECT id, tmdb_id, poster_path FROM movies
		WHERE poster_path IS NOT NULL AND poster_path != ''
		AND (poster_local IS NULL OR poster_local = '')
		LIMIT 100`)
	if err != nil {
		log.Printf("movie poster query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, tmdbID int
		var posterPath string
		if err := rows.Scan(&id, &tmdbID, &posterPath); err != nil {
			continue
		}
		localPath, err := downloadPoster(mediaDir, tmdbID, posterPath)
		if err != nil {
			log.Printf("movie poster %d: %v", tmdbID, err)
			continue
		}
		_, _ = pool.Exec(ctx, `UPDATE movies SET poster_local = $1 WHERE id = $2`, localPath, id)
		log.Printf("movie poster saved: %s", localPath)
		time.Sleep(300 * time.Millisecond)
	}
}

func syncPosters(ctx context.Context, pool *pgxpool.Pool, mediaDir, tmdbKey string) {
	rows, err := pool.Query(ctx, `
		SELECT id, tmdb_id, poster_path FROM shows
		WHERE poster_path IS NOT NULL AND poster_path != ''
		AND (poster_local IS NULL OR poster_local = '')
		LIMIT 100`)
	if err != nil {
		log.Printf("poster query: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, tmdbID int
		var posterPath string
		if err := rows.Scan(&id, &tmdbID, &posterPath); err != nil {
			continue
		}
		localPath, err := downloadPoster(mediaDir, tmdbID, posterPath)
		if err != nil {
			log.Printf("poster %d: %v", tmdbID, err)
			continue
		}
		_, _ = pool.Exec(ctx, `UPDATE shows SET poster_local = $1 WHERE id = $2`, localPath, id)
		log.Printf("poster saved: %s", localPath)
		time.Sleep(300 * time.Millisecond)
	}
}

func syncAirDates(ctx context.Context, pool *pgxpool.Pool, tmdbKey string) {
	if tmdbKey == "" {
		return
	}
	rows, err := pool.Query(ctx, `
		SELECT e.id, s.season_number, sh.tmdb_id
		FROM episodes e
		JOIN seasons s ON s.id = e.season_id
		JOIN shows sh ON sh.id = s.show_id
		WHERE e.air_date IS NULL
		LIMIT 50`)
	if err != nil {
		return
	}
	defer rows.Close()

	client := &http.Client{Timeout: 15 * time.Second}
	for rows.Next() {
		var epID, seasonNum, showTMDB int
		if err := rows.Scan(&epID, &seasonNum, &showTMDB); err != nil {
			continue
		}
		url := fmt.Sprintf("https://api.themoviedb.org/3/tv/%d/season/%d?api_key=%s", showTMDB, seasonNum, tmdbKey)
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var season struct {
			Episodes []struct {
				EpisodeNumber int    `json:"episode_number"`
				AirDate       string `json:"air_date"`
			} `json:"episodes"`
		}
		if json.Unmarshal(body, &season) != nil {
			continue
		}
		for _, ep := range season.Episodes {
			if ep.AirDate == "" {
				continue
			}
			_, _ = pool.Exec(ctx,
				`UPDATE episodes SET air_date = $1 WHERE id = $2 AND episode_number = $3`,
				ep.AirDate, epID, ep.EpisodeNumber)
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func downloadPoster(mediaDir string, tmdbID int, posterPath string) (string, error) {
	url := fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", posterPath)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	dir := filepath.Join(mediaDir, "posters", fmt.Sprintf("%d", tmdbID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		resp2, err2 := http.Get(url)
		if err2 != nil {
			return "", err
		}
		defer resp2.Body.Close()
		data, _ := io.ReadAll(resp2.Body)
		outPath := filepath.Join(dir, "w500.jpg")
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return "", err
		}
		return fmt.Sprintf("/posters/%d/w500.jpg", tmdbID), nil
	}

	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, nil); err != nil {
		outPath := filepath.Join(dir, "w500.png")
		f, _ := os.Create(outPath)
		defer f.Close()
		_ = png.Encode(f, img)
		return fmt.Sprintf("/posters/%d/w500.png", tmdbID), nil
	}

	outPath := filepath.Join(dir, "w500.webp")
	if err := os.WriteFile(outPath, buf.Bytes(), 0644); err != nil {
		return "", err
	}
	return fmt.Sprintf("/posters/%d/w500.webp", tmdbID), nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
