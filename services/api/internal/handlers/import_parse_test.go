package handlers

import (
	"encoding/json"
	"testing"
)

func TestParseTVTimeNestedShows(t *testing.T) {
	payload := []byte(`[
		{
			"title": "Station Eleven",
			"id": { "tvdb": 366529 },
			"seasons": [
				{
					"number": 1,
					"episodes": [
						{ "number": 1, "is_watched": true, "watched_at": "2024-01-15 10:00:00" },
						{ "number": 2, "is_watched": false }
					]
				}
			]
		}
	]`)

	items, err := parseImportPayload(payload)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ShowTitle != "Station Eleven" || items[0].SeasonNumber != 1 || items[0].EpisodeNumber != 1 {
		t.Fatalf("unexpected item: %+v", items[0])
	}
}

func TestParseTVTimeMovies(t *testing.T) {
	payload := []byte(`{
		"movies": [
			{ "title": "The Matrix", "is_watched": true, "watched_at": "2024-02-01T12:00:00Z" },
			{ "title": "Inception", "is_watched": false }
		]
	}`)

	items, err := parseImportPayload(payload)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 movie, got %d", len(items))
	}
	if items[0].MediaType != "movie" || items[0].ShowTitle != "The Matrix" {
		t.Fatalf("unexpected movie: %+v", items[0])
	}
}

func TestParseShowTrackExport(t *testing.T) {
	payload := []byte(`{
		"format": "showtrack-watch-history-v1",
		"items": [
			{ "tmdb_id": 1396, "show_title": "Breaking Bad", "season_number": 1, "episode_number": 1, "media_type": "tv" },
			{ "tmdb_id": 603, "show_title": "The Matrix", "media_type": "movie", "watched_at": "2024-03-01T00:00:00Z" }
		]
	}`)

	items, err := parseImportPayload(payload)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d (%s)", len(items), mustJSON(items))
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
