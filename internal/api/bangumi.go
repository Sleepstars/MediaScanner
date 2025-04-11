package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/models"
)

// BangumiClient represents the Bangumi API client
type BangumiClient struct {
	apiKey     string
	baseURL    string
	language   string
	userAgent  string
	httpClient *http.Client
	db         *database.Database
}

// NewBangumiClient creates a new Bangumi API client
func NewBangumiClient(cfg *config.BangumiConfig, db *database.Database) (*BangumiClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("Bangumi API key is required")
	}

	// Use the configured User-Agent or set a default one
	// Bangumi API requires a proper User-Agent
	// https://github.com/bangumi/api/blob/master/docs-raw/user%20agent.md
	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = "sleepstars/MediaScanner (https://github.com/sleepstars/MediaScanner)"
	}

	return &BangumiClient{
		apiKey:     cfg.APIKey,
		baseURL:    "https://api.bgm.tv/v0",
		language:   cfg.Language,
		userAgent:  userAgent,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		db:         db,
	}, nil
}

// SearchAnime searches for anime
func (c *BangumiClient) SearchAnime(ctx context.Context, query string) (*BangumiSearchResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("search:%s", query)
	cache, err := c.db.GetAPICache("bangumi", cacheKey)
	if err == nil {
		// Cache hit
		var result BangumiSearchResult
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	endpoint := fmt.Sprintf("%s/search/subjects?keyword=%s&type=2", c.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.language != "" {
		req.Header.Set("Accept-Language", c.language)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bangumi API error: %s - %s", resp.Status, string(body))
	}

	var apiResp struct {
		Data []struct {
			ID      int    `json:"id"`
			Name    string `json:"name"`
			NameCN  string `json:"name_cn"`
			Type    int    `json:"type"`
			Summary string `json:"summary"`
			Date    string `json:"date"`
			Images  struct {
				Small  string `json:"small"`
				Medium string `json:"medium"`
				Large  string `json:"large"`
			} `json:"images"`
			Rating struct {
				Score float64 `json:"score"`
				Count int     `json:"total"`
			} `json:"rating"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Process results
	result := &BangumiSearchResult{
		Query: query,
		Anime: make([]BangumiAnime, 0),
	}

	for _, item := range apiResp.Data {
		// Extract year from date
		var year int
		if len(item.Date) >= 4 {
			fmt.Sscanf(item.Date[:4], "%d", &year)
		}

		result.Anime = append(result.Anime, BangumiAnime{
			ID:       item.ID,
			Name:     item.Name,
			NameCN:   item.NameCN,
			Summary:  item.Summary,
			Year:     year,
			ImageURL: item.Images.Large,
			Rating:   item.Rating.Score,
		})
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "bangumi",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(24 * time.Hour), // Cache for 24 hours
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetAnimeDetails gets details for an anime
func (c *BangumiClient) GetAnimeDetails(ctx context.Context, id int) (*BangumiAnimeDetails, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("anime:%d", id)
	cache, err := c.db.GetAPICache("bangumi", cacheKey)
	if err == nil {
		// Cache hit
		var result BangumiAnimeDetails
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	endpoint := fmt.Sprintf("%s/subjects/%d", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.language != "" {
		req.Header.Set("Accept-Language", c.language)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Bangumi API error: %s - %s", resp.Status, string(body))
	}

	var apiResp struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		NameCN   string `json:"name_cn"`
		Type     int    `json:"type"`
		Summary  string `json:"summary"`
		Date     string `json:"date"`
		Platform int    `json:"platform"`
		Images   struct {
			Small  string `json:"small"`
			Medium string `json:"medium"`
			Large  string `json:"large"`
		} `json:"images"`
		Rating struct {
			Score float64 `json:"score"`
			Count int     `json:"total"`
		} `json:"rating"`
		Tags []struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		} `json:"tags"`
		Infobox []struct {
			Key   string      `json:"key"`
			Value interface{} `json:"value"`
		} `json:"infobox"`
		Episodes []struct {
			ID      int    `json:"id"`
			Type    int    `json:"type"`
			Name    string `json:"name"`
			NameCN  string `json:"name_cn"`
			Sort    int    `json:"sort"`
			AirDate string `json:"airdate"`
		} `json:"episodes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Extract year from date
	var year int
	if len(apiResp.Date) >= 4 {
		fmt.Sscanf(apiResp.Date[:4], "%d", &year)
	}

	// Process tags
	tags := make([]string, 0, len(apiResp.Tags))
	for _, tag := range apiResp.Tags {
		tags = append(tags, tag.Name)
	}

	// Process episodes
	episodes := make([]BangumiEpisode, 0, len(apiResp.Episodes))
	for _, episode := range apiResp.Episodes {
		episodes = append(episodes, BangumiEpisode{
			ID:      episode.ID,
			Type:    episode.Type,
			Name:    episode.Name,
			NameCN:  episode.NameCN,
			Sort:    episode.Sort,
			AirDate: episode.AirDate,
		})
	}

	// Create result
	result := &BangumiAnimeDetails{
		ID:       apiResp.ID,
		Name:     apiResp.Name,
		NameCN:   apiResp.NameCN,
		Summary:  apiResp.Summary,
		Year:     year,
		Platform: apiResp.Platform,
		ImageURL: apiResp.Images.Large,
		Rating:   apiResp.Rating.Score,
		Tags:     tags,
		Episodes: episodes,
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "bangumi",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Cache for 7 days
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// BangumiAnime represents an anime search result
type BangumiAnime struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	NameCN   string  `json:"name_cn"`
	Summary  string  `json:"summary"`
	Year     int     `json:"year"`
	ImageURL string  `json:"image_url"`
	Rating   float64 `json:"rating"`
}

// BangumiSearchResult represents an anime search result
type BangumiSearchResult struct {
	Query string         `json:"query"`
	Anime []BangumiAnime `json:"anime"`
}

// BangumiAnimeDetails represents detailed information about an anime
type BangumiAnimeDetails struct {
	ID       int              `json:"id"`
	Name     string           `json:"name"`
	NameCN   string           `json:"name_cn"`
	Summary  string           `json:"summary"`
	Year     int              `json:"year"`
	Platform int              `json:"platform"`
	ImageURL string           `json:"image_url"`
	Rating   float64          `json:"rating"`
	Tags     []string         `json:"tags"`
	Episodes []BangumiEpisode `json:"episodes"`
}

// BangumiEpisode represents an anime episode
type BangumiEpisode struct {
	ID      int    `json:"id"`
	Type    int    `json:"type"`
	Name    string `json:"name"`
	NameCN  string `json:"name_cn"`
	Sort    int    `json:"sort"`
	AirDate string `json:"air_date"`
}
