package handlers

import "github.com/gofiber/fiber/v2"

func demoSearch(q string) []fiber.Map {
	return []fiber.Map{
		{"id": 1396, "title": "Breaking Bad", "media_type": "tv", "overview": "A chemistry teacher turned meth maker.", "poster_url": "https://image.tmdb.org/t/p/w500/ggFHVNu6YYI5L9pCfOacjizRGt.jpg", "vote_average": 8.9},
		{"id": 66732, "title": "Stranger Things", "media_type": "tv", "overview": "Kids face supernatural forces.", "poster_url": "https://image.tmdb.org/t/p/w500/49WJfeN0moxb9IPfGn8AIqMGskD.jpg", "vote_average": 8.6},
		{"id": 17419, "title": "Bryan Cranston", "media_type": "person", "overview": "", "poster_url": "https://image.tmdb.org/t/p/w185/fnglrDfkcsEuHQgvhoL8n0kev1.jpg", "vote_average": 0},
	}
}

func demoTrending() []fiber.Map {
	return []fiber.Map{
		{"id": 1399, "tmdb_id": 1399, "title": "Game of Thrones", "media_type": "tv", "poster_url": "https://image.tmdb.org/t/p/w500/1XS1oqL89opfnbLl8WnZY1O1uJx.jpg", "vote_average": 8.4},
		{"id": 1396, "tmdb_id": 1396, "title": "Breaking Bad", "media_type": "tv", "poster_url": "https://image.tmdb.org/t/p/w500/ggFHVNu6YYI5L9pCfOacjizRGt.jpg", "vote_average": 8.9},
		{"id": 66732, "tmdb_id": 66732, "title": "Stranger Things", "media_type": "tv", "poster_url": "https://image.tmdb.org/t/p/w500/49WJfeN0moxb9IPfGn8AIqMGskD.jpg", "vote_average": 8.6},
		{"id": 94997, "tmdb_id": 94997, "title": "House of the Dragon", "media_type": "tv", "poster_url": "https://image.tmdb.org/t/p/w500/z2yahlsthc9HtPJk2Y7ggxfW6zH.jpg", "vote_average": 8.4},
	}
}

func demoShow(tmdbID string) fiber.Map {
	return fiber.Map{
		"id": 1, "tmdb_id": tmdbID, "title": "Breaking Bad",
		"overview": "When an unassuming high school chemistry teacher discovers he has a rare form of lung cancer, he decides to team up with a former student and create a top of the line crystal meth in a used RV, to provide for his family once he is gone.",
		"status": "Ended", "vote_average": 8.9,
		"poster_url": "https://image.tmdb.org/t/p/w500/ggFHVNu6YYI5L9pCfOacjizRGt.jpg",
		"cast": []fiber.Map{
			{"tmdb_id": 17419, "name": "Bryan Cranston", "character": "Walter White", "profile_url": "https://image.tmdb.org/t/p/w185/fnglrDfkcsEuHQgvhoL8n0kev1.jpg"},
			{"tmdb_id": 84497, "name": "Aaron Paul", "character": "Jesse Pinkman", "profile_url": "https://image.tmdb.org/t/p/w185/ycyc9NJOvZnsgFsqXJzLpYizwTr.jpg"},
		},
		"seasons": []fiber.Map{
			{"id": 1, "season_number": 1, "name": "Season 1", "episode_count": 7},
			{"id": 2, "season_number": 2, "name": "Season 2", "episode_count": 13},
		},
	}
}

func demoPerson(tmdbID string) fiber.Map {
	return fiber.Map{
		"id": 1, "tmdb_id": tmdbID, "name": "Bryan Cranston",
		"biography": "Bryan Lee Cranston is an American actor, director, producer and screenwriter.",
		"profile_url": "https://image.tmdb.org/t/p/w185/fnglrDfkcsEuHQgvhoL8n0kev1.jpg",
		"credits": []fiber.Map{
			{"media_type": "tv", "tmdb_id": 1396, "title": "Breaking Bad", "character": "Walter White", "poster_url": "https://image.tmdb.org/t/p/w500/ggFHVNu6YYI5L9pCfOacjizRGt.jpg", "vote_average": 8.9},
			{"media_type": "tv", "tmdb_id": 1695, "title": "Malcolm in the Middle", "character": "Hal", "poster_url": "https://image.tmdb.org/t/p/w500/6XqfrMnC8bk7t8xYy8j9mJQF9wF.jpg", "vote_average": 8.0},
		},
	}
}
