package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"github.com/showtrack/api/internal/config"
	"github.com/showtrack/api/internal/db"
	"github.com/showtrack/api/internal/handlers"
	"github.com/showtrack/api/internal/middleware"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	redis := db.ConnectRedis(cfg.RedisURL)
	defer redis.Close()

	app := fiber.New(fiber.Config{
		AppName:      "ShowTrack API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	h := handlers.New(pool, redis, cfg)

	api := app.Group("/api/v1")
	api.Get("/health", h.Health)

	api.Post("/auth/register", h.Register)
	api.Post("/auth/login", h.Login)

	api.Get("/search", h.Search)
	api.Get("/trending", h.Trending)
	api.Get("/trending/movies", h.TrendingMovies)
	api.Get("/shows/:id", h.GetShow)
	api.Get("/shows/:id/watch", h.GetShowWatchProviders)
	api.Get("/movies/:id", h.GetMovie)
	api.Get("/movies/:id/watch", h.GetMovieWatchProviders)
	api.Get("/persons/:id", h.GetPerson)

	protected := api.Group("", middleware.Auth(cfg.JWTSecret))
	protected.Get("/me/library", h.GetLibrary)
	protected.Get("/me/dashboard", h.GetDashboard)
	protected.Get("/me/export", h.ExportWatchHistory)
	protected.Post("/me/import", h.ImportWatchHistory)
	protected.Post("/shows", h.AddShow)
	protected.Delete("/shows/:id", h.RemoveShow)
	protected.Post("/movies", h.AddMovie)
	protected.Delete("/movies/:id", h.RemoveMovie)
	protected.Post("/movies/:id/watched", h.MarkMovieWatched)
	protected.Delete("/movies/:id/watched", h.UnmarkMovieWatched)
	protected.Post("/episodes/:id/watched", h.MarkWatched)
	protected.Delete("/episodes/:id/watched", h.UnmarkWatched)
	protected.Post("/devices", h.RegisterDevice)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("API listening on :%s", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	_ = app.Shutdown()
}
