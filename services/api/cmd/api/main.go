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
	api.Post("/auth/google", h.GoogleAuth)
	api.Get("/auth/trakt/callback", h.TraktCallback)

	api.Get("/search", h.Search)
	api.Get("/trending", h.Trending)
	api.Get("/trending/movies", h.TrendingMovies)
	api.Get("/discover", h.Discover)
	api.Get("/genres", h.GetGenres)
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
	protected.Patch("/shows/:id/status", h.UpdateShowStatus)
	protected.Delete("/shows/:id", h.RemoveShow)
	protected.Post("/movies", h.AddMovie)
	protected.Patch("/movies/:id/status", h.UpdateMovieStatus)
	protected.Delete("/movies/:id", h.RemoveMovie)
	protected.Post("/movies/:id/watched", h.MarkMovieWatched)
	protected.Delete("/movies/:id/watched", h.UnmarkMovieWatched)
	protected.Post("/episodes/:id/watched", h.MarkWatched)
	protected.Delete("/episodes/:id/watched", h.UnmarkWatched)
	protected.Get("/me/recommendations", h.GetRecommendations)
	protected.Post("/me/onboarding", h.Onboarding)
	protected.Post("/devices", h.RegisterDevice)
	protected.Get("/me/trakt", h.GetTraktStatus)
	protected.Post("/me/trakt/connect", h.StartTraktConnect)
	protected.Post("/me/trakt/sync", h.SyncTrakt)
	protected.Delete("/me/trakt", h.DisconnectTrakt)
	protected.Get("/me/profile", h.GetMyProfile)
	protected.Patch("/me/profile", h.UpdateMyProfile)
	protected.Get("/me/feed", h.GetFeed)
	protected.Get("/me/following", h.GetFollowing)
	protected.Get("/me/followers", h.GetFollowers)
	protected.Get("/users/search", h.SearchUsers)
	protected.Get("/users/:id", h.GetUserProfile)
	protected.Post("/users/:id/follow", h.FollowUser)
	protected.Delete("/users/:id/follow", h.UnfollowUser)
	protected.Get("/me/ratings/:type/:tmdb_id", h.GetMyRating)
	protected.Put("/me/ratings/:type/:tmdb_id", h.SetMyRating)
	protected.Delete("/me/ratings/:type/:tmdb_id", h.DeleteMyRating)
	protected.Get("/me/lists", h.GetMyLists)
	protected.Post("/me/lists", h.CreateList)
	protected.Get("/me/lists/:id", h.GetList)
	protected.Delete("/me/lists/:id", h.DeleteList)
	protected.Post("/me/lists/:id/items", h.AddListItem)
	protected.Delete("/me/lists/:id/items", h.RemoveListItem)

	admin := api.Group("/admin", h.requireAdmin)
	admin.Get("/stats", h.AdminStats)
	admin.Get("/users", h.AdminUsers)

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
