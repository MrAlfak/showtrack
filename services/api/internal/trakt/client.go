package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiBase = "https://api.trakt.tv"

type Client struct {
	clientID     string
	clientSecret string
	redirectURI  string
	http         *http.Client
}

func New(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		http:         &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c.clientID != "" && c.clientSecret != "" && c.redirectURI != ""
}

func (c *Client) AuthorizeURL(state string) string {
	params := url.Values{
		"response_type": {"code"},
		"client_id":     {c.clientID},
		"redirect_uri":  {c.redirectURI},
		"state":         {state},
	}
	return "https://trakt.tv/oauth/authorize?" + params.Encode()
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func (c *Client) ExchangeCode(code string) (*TokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"code":          code,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"redirect_uri":  c.redirectURI,
		"grant_type":    "authorization_code",
	})
	return c.postToken("/oauth/token", body)
}

func (c *Client) RefreshToken(refreshToken string) (*TokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"refresh_token": refreshToken,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"redirect_uri":  c.redirectURI,
		"grant_type":    "refresh_token",
	})
	return c.postToken("/oauth/token", body)
}

func (c *Client) postToken(path string, body []byte) (*TokenResponse, error) {
	req, err := http.NewRequest(http.MethodPost, apiBase+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("trakt token %d: %s", res.StatusCode, string(raw))
	}
	var token TokenResponse
	return &token, json.Unmarshal(raw, &token)
}

type UserSettings struct {
	User struct {
		Username string `json:"username"`
	} `json:"user"`
}

func (c *Client) GetSettings(accessToken string) (*UserSettings, error) {
	req, err := http.NewRequest(http.MethodGet, apiBase+"/users/settings", nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, accessToken)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("trakt settings %d", res.StatusCode)
	}
	var settings UserSettings
	return &settings, json.Unmarshal(raw, &settings)
}

type HistoryEpisode struct {
	WatchedAt time.Time `json:"watched_at"`
	Episode   struct {
		Season int `json:"season"`
		Number int `json:"number"`
	} `json:"episode"`
	Show struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		IDs   struct {
			TMDB int `json:"tmdb"`
			Trakt int `json:"trakt"`
		} `json:"ids"`
	} `json:"show"`
}

type HistoryMovie struct {
	WatchedAt time.Time `json:"watched_at"`
	Movie     struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		IDs   struct {
			TMDB int `json:"tmdb"`
		} `json:"ids"`
	} `json:"movie"`
}

func (c *Client) GetEpisodeHistory(accessToken string, page int) ([]HistoryEpisode, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/sync/history/episodes?page=%d&limit=100", apiBase, page), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, accessToken)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("trakt history %d", res.StatusCode)
	}
	var items []HistoryEpisode
	return items, json.Unmarshal(raw, &items)
}

func (c *Client) GetMovieHistory(accessToken string, page int) ([]HistoryMovie, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/sync/history/movies?page=%d&limit=100", apiBase, page), nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, accessToken)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("trakt movie history %d", res.StatusCode)
	}
	var items []HistoryMovie
	return items, json.Unmarshal(raw, &items)
}

func (c *Client) setHeaders(req *http.Request, accessToken string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", c.clientID)
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
}
