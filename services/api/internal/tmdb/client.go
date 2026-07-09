package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.themoviedb.org/3"

type Client struct {
	apiKey string
	http   *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) get(path string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("api_key", c.apiKey)
	u := fmt.Sprintf("%s%s?%s", baseURL, path, params.Encode())

	resp, err := c.http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("tmdb %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

type SearchResult struct {
	Page    int `json:"page"`
	Results []struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		Title        string  `json:"title"`
		Overview     string  `json:"overview"`
		PosterPath   string  `json:"poster_path"`
		ProfilePath  string  `json:"profile_path"`
		MediaType    string  `json:"media_type"`
		VoteAverage  float64 `json:"vote_average"`
		FirstAirDate string  `json:"first_air_date"`
	} `json:"results"`
}

func (c *Client) SearchMulti(query string) (*SearchResult, error) {
	params := url.Values{"query": {query}, "language": {"en-US"}}
	body, err := c.get("/search/multi", params)
	if err != nil {
		return nil, err
	}
	var result SearchResult
	return &result, json.Unmarshal(body, &result)
}

type TVShow struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Overview         string  `json:"overview"`
	PosterPath       string  `json:"poster_path"`
	BackdropPath     string  `json:"backdrop_path"`
	Status           string  `json:"status"`
	FirstAirDate     string  `json:"first_air_date"`
	VoteAverage      float64 `json:"vote_average"`
	Genres           []Genre `json:"genres"`
	NumberOfSeasons  int     `json:"number_of_seasons"`
	Seasons          []Season `json:"seasons"`
	Credits          Credits `json:"credits"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Season struct {
	ID            int    `json:"id"`
	SeasonNumber  int    `json:"season_number"`
	Name          string `json:"name"`
	EpisodeCount  int    `json:"episode_count"`
	PosterPath    string `json:"poster_path"`
	Episodes      []Episode `json:"episodes"`
}

type Episode struct {
	ID            int    `json:"id"`
	EpisodeNumber int    `json:"episode_number"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	AirDate       string `json:"air_date"`
	StillPath     string `json:"still_path"`
	Runtime       int    `json:"runtime"`
}

type Credits struct {
	Cast []CastMember `json:"cast"`
}

type CastMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
}

type Person struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Biography      string `json:"biography"`
	Birthday       string `json:"birthday"`
	PlaceOfBirth   string `json:"place_of_birth"`
	ProfilePath    string `json:"profile_path"`
	CombinedCredits struct {
		Cast []CreditItem `json:"cast"`
	} `json:"combined_credits"`
}

type CreditItem struct {
	ID           int     `json:"id"`
	MediaType    string  `json:"media_type"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	Character    string  `json:"character"`
	PosterPath   string  `json:"poster_path"`
	ReleaseDate  string  `json:"release_date"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
}

type Movie struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	Runtime      int     `json:"runtime"`
	ReleaseDate  string  `json:"release_date"`
	VoteAverage  float64 `json:"vote_average"`
	Genres       []Genre `json:"genres"`
	Credits      Credits `json:"credits"`
}

type TrendingResponse struct {
	Results []struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Title       string  `json:"title"`
		PosterPath  string  `json:"poster_path"`
		VoteAverage float64 `json:"vote_average"`
	} `json:"results"`
}

func (c *Client) GetTVShow(id int) (*TVShow, error) {
	params := url.Values{"append_to_response": {"credits"}}
	body, err := c.get(fmt.Sprintf("/tv/%d", id), params)
	if err != nil {
		return nil, err
	}
	var show TVShow
	return &show, json.Unmarshal(body, &show)
}

func (c *Client) GetSeason(showID, season int) (*Season, error) {
	body, err := c.get(fmt.Sprintf("/tv/%d/season/%d", showID, season), nil)
	if err != nil {
		return nil, err
	}
	var s Season
	return &s, json.Unmarshal(body, &s)
}

func (c *Client) GetPerson(id int) (*Person, error) {
	params := url.Values{"append_to_response": {"combined_credits"}}
	body, err := c.get(fmt.Sprintf("/person/%d", id), params)
	if err != nil {
		return nil, err
	}
	var p Person
	return &p, json.Unmarshal(body, &p)
}

func (c *Client) GetMovie(id int) (*Movie, error) {
	params := url.Values{"append_to_response": {"credits"}}
	body, err := c.get(fmt.Sprintf("/movie/%d", id), params)
	if err != nil {
		return nil, err
	}
	var movie Movie
	return &movie, json.Unmarshal(body, &movie)
}

func (c *Client) TrendingTV() (*TrendingResponse, error) {
	body, err := c.get("/trending/tv/week", nil)
	if err != nil {
		return nil, err
	}
	var t TrendingResponse
	return &t, json.Unmarshal(body, &t)
}

func (c *Client) TrendingMovies() (*TrendingResponse, error) {
	body, err := c.get("/trending/movie/week", nil)
	if err != nil {
		return nil, err
	}
	var t TrendingResponse
	return &t, json.Unmarshal(body, &t)
}

func PosterURL(path string, size string) string {
	if path == "" {
		return ""
	}
	if size == "" {
		size = "w500"
	}
	return fmt.Sprintf("https://image.tmdb.org/t/p/%s%s", size, path)
}

type WatchProvider struct {
	ID      int    `json:"provider_id"`
	Name    string `json:"provider_name"`
	LogoPath string `json:"logo_path"`
}

type WatchProvidersResult struct {
	Link     string          `json:"link"`
	Flatrate []WatchProvider `json:"flatrate"`
	Rent     []WatchProvider `json:"rent"`
	Buy      []WatchProvider `json:"buy"`
}

type WatchProvidersResponse struct {
	ID      int                              `json:"id"`
	Results map[string]WatchProvidersResult  `json:"results"`
}

func (c *Client) GetTVWatchProviders(id int) (*WatchProvidersResponse, error) {
	body, err := c.get(fmt.Sprintf("/tv/%d/watch/providers", id), nil)
	if err != nil {
		return nil, err
	}
	var result WatchProvidersResponse
	return &result, json.Unmarshal(body, &result)
}

func (c *Client) GetMovieWatchProviders(id int) (*WatchProvidersResponse, error) {
	body, err := c.get(fmt.Sprintf("/movie/%d/watch/providers", id), nil)
	if err != nil {
		return nil, err
	}
	var result WatchProvidersResponse
	return &result, json.Unmarshal(body, &result)
}
