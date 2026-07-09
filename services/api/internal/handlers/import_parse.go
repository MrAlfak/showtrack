package handlers

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func parseImportPayload(body []byte) ([]importItem, error) {
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("invalid json payload")
	}

	items := extractImportItems(raw)
	if len(items) == 0 {
		return nil, fmt.Errorf("unsupported import format")
	}
	return items, nil
}

func extractImportItems(raw any) []importItem {
	switch value := raw.(type) {
	case []any:
		return flattenImportArray(value)
	case map[string]any:
		if wrapped, ok := value["items"].([]any); ok && len(wrapped) > 0 {
			return extractImportItems(wrapped)
		}
		for _, key := range []string{
			"shows", "series", "tv", "tv_shows",
			"movies", "films",
			"episodes", "history", "watched_episodes", "watched", "data",
		} {
			if nested, ok := value[key].([]any); ok && len(nested) > 0 {
				if items := flattenImportArray(nested); len(items) > 0 {
					return items
				}
			}
		}
	}
	return nil
}

func flattenImportArray(entries []any) []importItem {
	result := make([]importItem, 0, len(entries))
	for _, entry := range entries {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		switch {
		case hasNestedSeasons(item):
			result = append(result, flattenTVTimeShow(item)...)
		case isTVTimeMovieRecord(item):
			if parsed := movieRecordToImportItem(item); parsed != nil {
				result = append(result, *parsed)
			}
		default:
			result = append(result, normalizeImportItems([]map[string]any{item})...)
		}
	}
	return result
}

func hasNestedSeasons(item map[string]any) bool {
	seasons, ok := item["seasons"].([]any)
	return ok && len(seasons) > 0
}

func isTVTimeMovieRecord(item map[string]any) bool {
	if hasNestedSeasons(item) {
		return false
	}
	mediaType := strings.ToLower(getString(item, "media_type", "type"))
	if mediaType == "movie" || mediaType == "film" {
		return true
	}
	if getInt(item, "season_number", "season", "s") > 0 || getInt(item, "episode_number", "episode", "e") > 0 {
		return false
	}
	return getString(item, "title", "name") != "" || extractExternalID(item, "tmdb") > 0
}

func flattenTVTimeShow(show map[string]any) []importItem {
	title := getString(show, "title", "name", "show_title", "series_name")
	tmdbID := extractExternalID(show, "tmdb")
	if tmdbID == 0 {
		tmdbID = getInt(show, "tmdb_id", "show_tmdb_id")
	}

	seasons, _ := show["seasons"].([]any)
	result := make([]importItem, 0)
	for _, seasonEntry := range seasons {
		season, ok := seasonEntry.(map[string]any)
		if !ok {
			continue
		}
		seasonNumber := getInt(season, "number", "season_number", "season")
		episodes, _ := season["episodes"].([]any)
		for _, episodeEntry := range episodes {
			episode, ok := episodeEntry.(map[string]any)
			if !ok {
				continue
			}
			if !getBool(episode, "is_watched", "watched", "seen") {
				continue
			}
			episodeNumber := getInt(episode, "number", "episode_number", "episode")
			if seasonNumber <= 0 || episodeNumber <= 0 {
				continue
			}
			result = append(result, importItem{
				TMDBID:        tmdbID,
				ShowTitle:     title,
				MediaType:     "tv",
				SeasonNumber:  seasonNumber,
				EpisodeNumber: episodeNumber,
				WatchedAt:     firstTime(episode, "watched_at", "seen_at", "created_at", "updated_at", "date"),
			})
		}
	}
	return result
}

func movieRecordToImportItem(item map[string]any) *importItem {
	if !getBool(item, "is_watched", "watched", "seen") {
		return nil
	}
	title := getString(item, "title", "name", "movie_title")
	tmdbID := extractExternalID(item, "tmdb")
	if tmdbID == 0 {
		tmdbID = getInt(item, "tmdb_id", "movie_tmdb_id")
	}
	if title == "" && tmdbID == 0 {
		return nil
	}
	parsed := importItem{
		TMDBID:    tmdbID,
		ShowTitle: title,
		MediaType: "movie",
		WatchedAt: firstTime(item, "watched_at", "seen_at", "created_at", "updated_at", "date"),
	}
	return &parsed
}

func extractExternalID(item map[string]any, provider string) int {
	ids, ok := item["id"].(map[string]any)
	if !ok {
		return 0
	}
	return getInt(ids, provider)
}

func getBool(item map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := item[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			switch strings.ToLower(strings.TrimSpace(typed)) {
			case "true", "1", "yes":
				return true
			case "false", "0", "no":
				return false
			}
		case float64:
			return typed != 0
		case int:
			return typed != 0
		}
	}
	return false
}

func firstTime(item map[string]any, keys ...string) *time.Time {
	if parsed := getTime(item, keys...); parsed != nil {
		return parsed
	}
	return nil
}
