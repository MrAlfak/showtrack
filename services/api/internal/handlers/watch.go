package handlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/showtrack/api/internal/tmdb"
)

func (h *Handler) GetShowWatchProviders(c *fiber.Ctx) error {
	return h.getWatchProviders(c, "tv")
}

func (h *Handler) GetMovieWatchProviders(c *fiber.Ctx) error {
	return h.getWatchProviders(c, "movie")
}

func (h *Handler) getWatchProviders(c *fiber.Ctx, mediaType string) error {
	tmdbID, err := strconv.Atoi(c.Params("id"))
	if err != nil || tmdbID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}

	country := c.Query("country", "US")
	if country == "" {
		country = "US"
	}

	if h.cfg.TMDBAPIKey == "" {
		return c.JSON(demoWatchProviders(country))
	}

	var resp *tmdb.WatchProvidersResponse
	if mediaType == "movie" {
		resp, err = h.tmdb.GetMovieWatchProviders(tmdbID)
	} else {
		resp, err = h.tmdb.GetTVWatchProviders(tmdbID)
	}
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}

	result, ok := resp.Results[country]
	if !ok {
		for _, fallback := range []string{"US", "GB"} {
			if result, ok = resp.Results[fallback]; ok {
				country = fallback
				break
			}
		}
	}
	if !ok {
		return c.JSON(fiber.Map{
			"country":   country,
			"link":      "",
			"flatrate":  []fiber.Map{},
			"rent":      []fiber.Map{},
			"buy":       []fiber.Map{},
		})
	}

	return c.JSON(fiber.Map{
		"country":  country,
		"link":     result.Link,
		"flatrate": mapWatchProviders(result.Flatrate),
		"rent":     mapWatchProviders(result.Rent),
		"buy":      mapWatchProviders(result.Buy),
	})
}

func mapWatchProviders(providers []tmdb.WatchProvider) []fiber.Map {
	items := make([]fiber.Map, 0, len(providers))
	for _, provider := range providers {
		items = append(items, fiber.Map{
			"id":        provider.ID,
			"name":      provider.Name,
			"logo_url":  tmdb.PosterURL(provider.LogoPath, "w92"),
		})
	}
	return items
}

func demoWatchProviders(country string) fiber.Map {
	return fiber.Map{
		"country": country,
		"link":    fmt.Sprintf("https://www.themoviedb.org/%s/demo/watch", country),
		"flatrate": []fiber.Map{
			{"id": 8, "name": "Netflix", "logo_url": "https://image.tmdb.org/t/p/w92/t2yyOv40HZ89vKI18AW4HjZvjxJ.jpg"},
			{"id": 337, "name": "Disney Plus", "logo_url": "https://image.tmdb.org/t/p/w92/dZ5ztFS9vkhoj4oR66jH2fM61wZ.jpg"},
		},
		"rent": []fiber.Map{
			{"id": 2, "name": "Apple TV", "logo_url": "https://image.tmdb.org/t/p/w92/peURlLlr8Dct60mXiuuMsSZyKEw.jpg"},
		},
		"buy": []fiber.Map{},
	}
}
