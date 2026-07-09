package recommend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/showtrack/api/internal/tmdb"
)

const (
	weightCollaborative = 0.30
	weightFollowing     = 0.20
	weightGenre         = 0.30
	weightSimilar       = 0.20
	maxResults          = 16
)

type ScoredItem struct {
	TMDBID      int     `json:"tmdb_id"`
	MediaType   string  `json:"media_type"`
	Title       string  `json:"title"`
	PosterURL   string  `json:"poster_url"`
	VoteAverage float64 `json:"vote_average,omitempty"`
	Score       float64 `json:"score,omitempty"`
	Reason      string  `json:"rec_reason,omitempty"`
}

type Output struct {
	Results       []ScoredItem   `json:"results"`
	SeedTitle     string         `json:"seed_title,omitempty"`
	Engine        string         `json:"engine"`
	Explanation   string         `json:"explanation,omitempty"`
	SignalCounts  map[string]int `json:"signal_counts,omitempty"`
}

type Engine struct {
	DB        *pgxpool.Pool
	TMDB      *tmdb.Client
	HasTMDB   bool
	PosterURL func(string) string
}

type candidateKey struct {
	mediaType string
	tmdbID    int
}

type accumulator struct {
	item   ScoredItem
	score  float64
	reason string
}

func (e *Engine) Generate(ctx context.Context, userID string) (*Output, error) {
	excluded := e.librarySet(ctx, userID)
	if len(excluded) == 0 && !e.hasAnyLibrary(ctx, userID) {
		return &Output{Engine: "cold_start", Results: []ScoredItem{}, Explanation: "add shows to your library for personalized picks"}, nil
	}

	genreWeights, topGenreName := e.genreWeights(ctx, userID)
	seeds, seedTitle := e.topSeeds(ctx, userID)
	merged := map[candidateKey]*accumulator{}

	add := func(item ScoredItem, rawScore float64, weight float64, reason string) {
		if item.TMDBID == 0 || item.MediaType == "" {
			return
		}
		key := candidateKey{mediaType: item.MediaType, tmdbID: item.TMDBID}
		if excluded[key] {
			return
		}
		points := rawScore * weight
		if points <= 0 {
			return
		}
		entry, ok := merged[key]
		if !ok {
			item.Reason = reason
			merged[key] = &accumulator{item: item, score: points, reason: reason}
			return
		}
		entry.score += points
		if entry.score > points {
			entry.reason = reason
			entry.item.Reason = reason
		}
	}

	signalCounts := map[string]int{}

	for _, item := range e.collaborative(ctx, userID, excluded) {
		add(item, item.Score, weightCollaborative, "collaborative")
		signalCounts["collaborative"]++
	}
	for _, item := range e.fromFollowing(ctx, userID, excluded) {
		add(item, item.Score, weightFollowing, "following")
		signalCounts["following"]++
	}
	for _, item := range e.fromGenres(ctx, genreWeights, excluded) {
		add(item, item.Score, weightGenre, "genre")
		signalCounts["genre"]++
	}
	for _, item := range e.fromSimilar(seeds, excluded) {
		add(item, item.Score, weightSimilar, "similar")
		signalCounts["similar"]++
	}

	if len(merged) == 0 {
		return &Output{
			Engine:      "hybrid",
			SeedTitle:   seedTitle,
			Results:     []ScoredItem{},
			Explanation: "not enough data yet — keep watching and rating titles",
		}, nil
	}

	ranked := make([]*accumulator, 0, len(merged))
	for _, entry := range merged {
		entry.item.Score = math.Round(entry.score*100) / 100
		ranked = append(ranked, entry)
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].item.VoteAverage > ranked[j].item.VoteAverage
		}
		return ranked[i].score > ranked[j].score
	})

	limit := maxResults
	if len(ranked) < limit {
		limit = len(ranked)
	}
	results := make([]ScoredItem, 0, limit)
	for i := 0; i < limit; i++ {
		results = append(results, ranked[i].item)
	}

	explanation := buildExplanation(signalCounts, topGenreName, seedTitle)
	return &Output{
		Results:      results,
		SeedTitle:    seedTitle,
		Engine:       "hybrid",
		Explanation:  explanation,
		SignalCounts: signalCounts,
	}, nil
}

func buildExplanation(signals map[string]int, genre string, seed string) string {
	parts := []string{}
	if signals["collaborative"] > 0 {
		parts = append(parts, "viewers with similar taste")
	}
	if signals["following"] > 0 {
		parts = append(parts, "people you follow")
	}
	if signals["genre"] > 0 && genre != "" {
		parts = append(parts, fmt.Sprintf("your interest in %s", genre))
	}
	if signals["similar"] > 0 && seed != "" {
		parts = append(parts, fmt.Sprintf("titles like %s", seed))
	}
	if len(parts) == 0 {
		return "personalized from your library"
	}
	return "Based on " + strings.Join(parts, ", ")
}

func (e *Engine) librarySet(ctx context.Context, userID string) map[candidateKey]bool {
	set := map[candidateKey]bool{}
	rows, err := e.DB.Query(ctx, `
		SELECT s.tmdb_id, 'tv' FROM user_shows us JOIN shows s ON s.id = us.show_id WHERE us.user_id = $1::uuid
		UNION ALL
		SELECT m.tmdb_id, 'movie' FROM user_movies um JOIN movies m ON m.id = um.movie_id WHERE um.user_id = $1::uuid`,
		userID,
	)
	if err != nil {
		return set
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var mediaType string
		if rows.Scan(&id, &mediaType) == nil {
			set[candidateKey{mediaType: mediaType, tmdbID: id}] = true
		}
	}
	return set
}

func (e *Engine) hasAnyLibrary(ctx context.Context, userID string) bool {
	var count int
	_ = e.DB.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM user_shows WHERE user_id = $1::uuid) +
			(SELECT COUNT(*) FROM user_movies WHERE user_id = $1::uuid)`,
		userID,
	).Scan(&count)
	return count > 0
}

type seedItem struct {
	TMDBID    int
	MediaType string
	Title     string
	Activity  float64
}

func (e *Engine) topSeeds(ctx context.Context, userID string) ([]seedItem, string) {
	seeds := make([]seedItem, 0)

	rows, err := e.DB.Query(ctx, `
		SELECT s.tmdb_id, s.title,
		       COALESCE((SELECT COUNT(*)::float FROM user_episodes ue
		                 JOIN episodes ep ON ep.id = ue.episode_id
		                 JOIN seasons sn ON sn.id = ep.season_id
		                 WHERE sn.show_id = s.id AND ue.user_id = us.user_id), 0) AS watched,
		       COALESCE((SELECT score::float FROM user_ratings ur
		                 WHERE ur.user_id = us.user_id AND ur.media_type = 'tv' AND ur.tmdb_id = s.tmdb_id), 0) AS rating
		FROM user_shows us
		JOIN shows s ON s.id = us.show_id
		WHERE us.user_id = $1::uuid
		ORDER BY us.added_at DESC
		LIMIT 12`, userID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var tmdbID int
			var title string
			var watched, rating float64
			if rows.Scan(&tmdbID, &title, &watched, &rating) == nil {
				activity := watched*1.5 + rating*2
				if us, ok := listStatusWeight(ctx, e.DB, userID, tmdbID, "tv"); ok {
					activity += us
				}
				seeds = append(seeds, seedItem{TMDBID: tmdbID, MediaType: "tv", Title: title, Activity: activity})
			}
		}
	}

	movieRows, err := e.DB.Query(ctx, `
		SELECT m.tmdb_id, m.title,
		       CASE WHEN um.watched THEN 3.0 ELSE 0 END +
		       COALESCE((SELECT score::float FROM user_ratings ur
		                 WHERE ur.user_id = um.user_id AND ur.media_type = 'movie' AND ur.tmdb_id = m.tmdb_id), 0) * 2 AS activity
		FROM user_movies um
		JOIN movies m ON m.id = um.movie_id
		WHERE um.user_id = $1::uuid
		ORDER BY um.added_at DESC
		LIMIT 8`, userID,
	)
	if err == nil {
		defer movieRows.Close()
		for movieRows.Next() {
			var tmdbID int
			var title string
			var activity float64
			if movieRows.Scan(&tmdbID, &title, &activity) == nil {
				seeds = append(seeds, seedItem{TMDBID: tmdbID, MediaType: "movie", Title: title, Activity: activity})
			}
		}
	}

	sort.Slice(seeds, func(i, j int) bool { return seeds[i].Activity > seeds[j].Activity })
	if len(seeds) > 5 {
		seeds = seeds[:5]
	}
	seedTitle := ""
	if len(seeds) > 0 {
		seedTitle = seeds[0].Title
	}
	return seeds, seedTitle
}

func listStatusWeight(ctx context.Context, db *pgxpool.Pool, userID string, tmdbID int, mediaType string) (float64, bool) {
	if mediaType != "tv" {
		return 0, false
	}
	var status string
	err := db.QueryRow(ctx, `
		SELECT us.list_status FROM user_shows us
		JOIN shows s ON s.id = us.show_id
		WHERE us.user_id = $1::uuid AND s.tmdb_id = $2`, userID, tmdbID,
	).Scan(&status)
	if err != nil {
		return 0, false
	}
	switch status {
	case "watching":
		return 2, true
	case "plan_to_watch":
		return 1, true
	case "watched":
		return 1.5, true
	default:
		return 0.5, true
	}
}

func (e *Engine) genreWeights(ctx context.Context, userID string) (map[int]float64, string) {
	weights := map[int]float64{}
	names := map[int]string{}

	rows, _ := e.DB.Query(ctx, `
		SELECT s.genres, 1.5 FROM user_shows us JOIN shows s ON s.id = us.show_id WHERE us.user_id = $1::uuid
		UNION ALL
		SELECT m.genres, 1.0 FROM user_movies um JOIN movies m ON m.id = um.movie_id WHERE um.user_id = $1::uuid`,
		userID,
	)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var raw []byte
			var w float64
			if rows.Scan(&raw, &w) != nil {
				continue
			}
			applyGenreBytes(weights, names, raw, w)
		}
	}

	ratingRows, _ := e.DB.Query(ctx, `
		SELECT s.genres, ur.score::float * 0.4
		FROM user_ratings ur
		JOIN shows s ON s.tmdb_id = ur.tmdb_id
		WHERE ur.user_id = $1::uuid AND ur.media_type = 'tv'
		UNION ALL
		SELECT m.genres, ur.score::float * 0.4
		FROM user_ratings ur
		JOIN movies m ON m.tmdb_id = ur.tmdb_id
		WHERE ur.user_id = $1::uuid AND ur.media_type = 'movie'`,
		userID,
	)
	if ratingRows != nil {
		defer ratingRows.Close()
		for ratingRows.Next() {
			var raw []byte
			var w float64
			if ratingRows.Scan(&raw, &w) != nil {
				continue
			}
			applyGenreBytes(weights, names, raw, w)
		}
	}

	topGenreName := ""
	topWeight := 0.0
	for id, w := range weights {
		if w > topWeight {
			topWeight = w
			topGenreName = names[id]
		}
	}
	return weights, topGenreName
}

func applyGenreBytes(weights map[int]float64, names map[int]string, raw []byte, w float64) {
	if len(raw) == 0 {
		return
	}
	var genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if json.Unmarshal(raw, &genres) != nil {
		return
	}
	for _, g := range genres {
		if g.ID == 0 {
			continue
		}
		weights[g.ID] += w
		if g.Name != "" {
			names[g.ID] = g.Name
		}
	}
}

func (e *Engine) collaborative(ctx context.Context, userID string, excluded map[candidateKey]bool) []ScoredItem {
	rows, err := e.DB.Query(ctx, `
		SELECT s2.tmdb_id, 'tv', s2.title, COALESCE(s2.poster_path, ''), COALESCE(s2.vote_average, 0),
		       COUNT(*)::float AS co_count
		FROM user_shows us_me
		JOIN user_shows us_them ON us_me.show_id = us_them.show_id AND us_them.user_id <> us_me.user_id
		JOIN user_shows us_other ON us_other.user_id = us_them.user_id AND us_other.show_id <> us_me.show_id
		JOIN shows s2 ON s2.id = us_other.show_id
		WHERE us_me.user_id = $1::uuid
		GROUP BY s2.tmdb_id, s2.title, s2.poster_path, s2.vote_average
		ORDER BY co_count DESC
		LIMIT 24`, userID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	maxCo := 1.0
	items := make([]ScoredItem, 0)
	for rows.Next() {
		var item ScoredItem
		var co float64
		var posterPath string
		if err := rows.Scan(&item.TMDBID, &item.MediaType, &item.Title, &posterPath, &item.VoteAverage, &co); err != nil {
			continue
		}
		if excluded[candidateKey{item.MediaType, item.TMDBID}] {
			continue
		}
		item.PosterURL = e.PosterURL(posterPath)
		item.Score = co
		if co > maxCo {
			maxCo = co
		}
		items = append(items, item)
	}
	for i := range items {
		items[i].Score = items[i].Score / maxCo
	}
	return items
}

func (e *Engine) fromFollowing(ctx context.Context, userID string, excluded map[candidateKey]bool) []ScoredItem {
	rows, err := e.DB.Query(ctx, `
		SELECT s.tmdb_id, 'tv', s.title, COALESCE(s.poster_path, ''), COALESCE(s.vote_average, 0),
		       COUNT(*)::float AS freq
		FROM user_follows f
		JOIN user_shows us ON us.user_id = f.following_id
		JOIN shows s ON s.id = us.show_id
		WHERE f.follower_id = $1::uuid
		GROUP BY s.tmdb_id, s.title, s.poster_path, s.vote_average
		UNION ALL
		SELECT m.tmdb_id, 'movie', m.title, COALESCE(m.poster_path, ''), COALESCE(m.vote_average, 0),
		       COUNT(*)::float AS freq
		FROM user_follows f
		JOIN user_movies um ON um.user_id = f.following_id
		JOIN movies m ON m.id = um.movie_id
		WHERE f.follower_id = $1::uuid
		GROUP BY m.tmdb_id, m.title, m.poster_path, m.vote_average
		ORDER BY freq DESC
		LIMIT 20`, userID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	maxFreq := 1.0
	items := make([]ScoredItem, 0)
	for rows.Next() {
		var item ScoredItem
		var posterPath string
		var freq float64
		if err := rows.Scan(&item.TMDBID, &item.MediaType, &item.Title, &posterPath, &item.VoteAverage, &freq); err != nil {
			continue
		}
		if excluded[candidateKey{item.MediaType, item.TMDBID}] {
			continue
		}
		item.PosterURL = e.PosterURL(posterPath)
		item.Score = freq
		if freq > maxFreq {
			maxFreq = freq
		}
		items = append(items, item)
	}
	for i := range items {
		items[i].Score = items[i].Score / maxFreq
	}
	return items
}

func (e *Engine) fromGenres(ctx context.Context, weights map[int]float64, excluded map[candidateKey]bool) []ScoredItem {
	if len(weights) == 0 {
		return nil
	}

	type genreScore struct {
		id    int
		score float64
	}
	ranked := make([]genreScore, 0, len(weights))
	for id, w := range weights {
		ranked = append(ranked, genreScore{id, w})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	if len(ranked) > 3 {
		ranked = ranked[:3]
	}

	items := make([]ScoredItem, 0)
	seen := map[candidateKey]bool{}

	for _, genre := range ranked {
		if e.HasTMDB && e.TMDB != nil {
			for _, mediaType := range []string{"tv", "movie"} {
				var resp *tmdb.TrendingResponse
				var err error
				if mediaType == "movie" {
					resp, err = e.TMDB.DiscoverMovie(genre.id, "vote_average.desc")
				} else {
					resp, err = e.TMDB.DiscoverTV(genre.id, "vote_average.desc")
				}
				if err != nil || resp == nil {
					continue
				}
				for i, r := range resp.Results {
					if i >= 8 {
						break
					}
					title := r.Name
					if title == "" {
						title = r.Title
					}
					item := ScoredItem{
						TMDBID:      r.ID,
						MediaType:   mediaType,
						Title:       title,
						PosterURL:   e.PosterURL(r.PosterPath),
						VoteAverage: r.VoteAverage,
						Score:       genre.score / float64(i+2),
					}
					key := candidateKey{mediaType, item.TMDBID}
					if excluded[key] || seen[key] {
						continue
					}
					seen[key] = true
					items = append(items, item)
				}
			}
			continue
		}

		rows, err := e.DB.Query(ctx, `
			SELECT tmdb_id, title, COALESCE(poster_path, ''), COALESCE(vote_average, 0), 'tv'
			FROM shows
			WHERE EXISTS (
				SELECT 1 FROM jsonb_array_elements(genres) g
				WHERE (g->>'id')::int = $1
			)
			ORDER BY vote_average DESC NULLS LAST
			LIMIT 8`, genre.id,
		)
		if err == nil {
			for rows.Next() {
				var item ScoredItem
				var posterPath string
				if rows.Scan(&item.TMDBID, &item.Title, &posterPath, &item.VoteAverage, &item.MediaType) == nil {
					key := candidateKey{item.MediaType, item.TMDBID}
					if excluded[key] || seen[key] {
						continue
					}
					seen[key] = true
					item.PosterURL = e.PosterURL(posterPath)
					item.Score = genre.score
					items = append(items, item)
				}
			}
			rows.Close()
		}
	}

	maxScore := 1.0
	for _, item := range items {
		if item.Score > maxScore {
			maxScore = item.Score
		}
	}
	for i := range items {
		items[i].Score = items[i].Score / maxScore
	}
	return items
}

func (e *Engine) fromSimilar(seeds []seedItem, excluded map[candidateKey]bool) []ScoredItem {
	if !e.HasTMDB || e.TMDB == nil || len(seeds) == 0 {
		return nil
	}

	items := make([]ScoredItem, 0)
	seen := map[candidateKey]bool{}

	for _, seed := range seeds {
		var resp *tmdb.TrendingResponse
		var err error
		if seed.MediaType == "movie" {
			resp, err = e.TMDB.RecommendationsMovie(seed.TMDBID)
			if err != nil || resp == nil || len(resp.Results) == 0 {
				resp, err = e.TMDB.SimilarMovie(seed.TMDBID)
			}
		} else {
			resp, err = e.TMDB.RecommendationsTV(seed.TMDBID)
			if err != nil || resp == nil || len(resp.Results) == 0 {
				resp, err = e.TMDB.SimilarTV(seed.TMDBID)
			}
		}
		if err != nil || resp == nil {
			continue
		}
		seedBoost := seed.Activity
		if seedBoost < 1 {
			seedBoost = 1
		}
		for i, r := range resp.Results {
			if i >= 6 {
				break
			}
			title := r.Name
			if title == "" {
				title = r.Title
			}
			mediaType := seed.MediaType
			item := ScoredItem{
				TMDBID:      r.ID,
				MediaType:   mediaType,
				Title:       title,
				PosterURL:   e.PosterURL(r.PosterPath),
				VoteAverage: r.VoteAverage,
				Score:       seedBoost / float64(i+1),
			}
			key := candidateKey{mediaType, item.TMDBID}
			if excluded[key] || seen[key] {
				continue
			}
			seen[key] = true
			items = append(items, item)
		}
	}

	maxScore := 1.0
	for _, item := range items {
		if item.Score > maxScore {
			maxScore = item.Score
		}
	}
	for i := range items {
		items[i].Score = items[i].Score / maxScore
	}
	return items
}
