package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	TMDBAPIKey  string
	MediaURL    string
	JWTSecret   string
	CORSOrigins string

	GoogleClientID     string
	TraktClientID      string
	TraktClientSecret  string
	TraktRedirectURI   string
	PublicWebURL       string
	AdminSecret        string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://showtrack:showtrack@postgres:5432/showtrack?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://redis:6379"),
		TMDBAPIKey:  getEnv("TMDB_API_KEY", ""),
		MediaURL:    getEnv("MEDIA_URL", "http://localhost:8090"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		CORSOrigins: getEnv("CORS_ORIGINS", "*"),

		GoogleClientID:    getEnv("GOOGLE_CLIENT_ID", ""),
		TraktClientID:     getEnv("TRAKT_CLIENT_ID", ""),
		TraktClientSecret: getEnv("TRAKT_CLIENT_SECRET", ""),
		TraktRedirectURI:  getEnv("TRAKT_REDIRECT_URI", ""),
		PublicWebURL:      getEnv("PUBLIC_WEB_URL", "http://localhost:3000"),
		AdminSecret:       getEnv("ADMIN_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
