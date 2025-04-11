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

// TVDBClient represents the TVDB API client
type TVDBClient struct {
	apiKey     string
	baseURL    string
	language   string
	httpClient *http.Client
	db         *database.Database
}

// NewTVDBClient creates a new TVDB API client
func NewTVDBClient(cfg *config.TVDBConfig, db *database.Database) (*TVDBClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("TVDB API key is required")
	}

	return &TVDBClient{
		apiKey:     cfg.APIKey,
		baseURL:    "https://api.thetvdb.com/v4",
		language:   cfg.Language,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		db:         db,
	}, nil
}

// SearchSeries searches for a TV series
func (c *TVDBClient) SearchSeries(ctx context.Context, query string) (*TVDBSearchResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("search:%s", query)
	cache, err := c.db.GetAPICache("tvdb", cacheKey)
	if err == nil {
		// Cache hit
		var result TVDBSearchResult
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	endpoint := fmt.Sprintf("%s/search?query=%s", c.baseURL, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "application/json")
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
		return nil, fmt.Errorf("TVDB API error: %s - %s", resp.Status, string(body))
	}

	var apiResp struct {
		Data []struct {
			ID           int    `json:"id"`
			Name         string `json:"name"`
			Type         string `json:"type"`
			FirstAired   string `json:"first_aired,omitempty"`
			Overview     string `json:"overview,omitempty"`
			PosterURL    string `json:"poster,omitempty"`
			BackdropURL  string `json:"backdrop,omitempty"`
			Status       string `json:"status,omitempty"`
			Network      string `json:"network,omitempty"`
			TVDBScore    int    `json:"tvdb_score,omitempty"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Process results
	result := &TVDBSearchResult{
		Query:  query,
		Series: make([]TVDBSeries, 0),
	}

	for _, item := range apiResp.Data {
		if item.Type != "series" {
			continue
		}

		// Extract year from first aired date
		var firstAiredYear int
		if item.FirstAired != "" {
			t, err := time.Parse("2006-01-02", item.FirstAired)
			if err == nil {
				firstAiredYear = t.Year()
			}
		}

		result.Series = append(result.Series, TVDBSeries{
			ID:            item.ID,
			Name:          item.Name,
			FirstAiredYear: firstAiredYear,
			Overview:      item.Overview,
			PosterURL:     item.PosterURL,
			BackdropURL:   item.BackdropURL,
			Status:        item.Status,
			Network:       item.Network,
			TVDBScore:     item.TVDBScore,
		})
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tvdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(24 * time.Hour), // Cache for 24 hours
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetSeriesDetails gets details for a TV series
func (c *TVDBClient) GetSeriesDetails(ctx context.Context, id int) (*TVDBSeriesDetails, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("series:%d", id)
	cache, err := c.db.GetAPICache("tvdb", cacheKey)
	if err == nil {
		// Cache hit
		var result TVDBSeriesDetails
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	endpoint := fmt.Sprintf("%s/series/%d/extended", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "application/json")
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
		return nil, fmt.Errorf("TVDB API error: %s - %s", resp.Status, string(body))
	}

	var apiResp struct {
		Data struct {
			ID           int    `json:"id"`
			Name         string `json:"name"`
			Overview     string `json:"overview"`
			FirstAired   string `json:"first_aired"`
			Status       string `json:"status"`
			Network      string `json:"network"`
			ImdbID       string `json:"imdb_id"`
			PosterURL    string `json:"poster"`
			BackdropURL  string `json:"backdrop"`
			Seasons      []struct {
				ID           int    `json:"id"`
				Name         string `json:"name"`
				Number       int    `json:"number"`
				EpisodeCount int    `json:"episode_count"`
				Overview     string `json:"overview"`
				PosterURL    string `json:"poster"`
			} `json:"seasons"`
			Genres []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"genres"`
			Countries []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"countries"`
			Languages []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"languages"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Extract year from first aired date
	var firstAiredYear int
	if apiResp.Data.FirstAired != "" {
		t, err := time.Parse("2006-01-02", apiResp.Data.FirstAired)
		if err == nil {
			firstAiredYear = t.Year()
		}
	}

	// Process genres
	genres := make([]string, 0, len(apiResp.Data.Genres))
	for _, genre := range apiResp.Data.Genres {
		genres = append(genres, genre.Name)
	}

	// Process countries
	countries := make([]string, 0, len(apiResp.Data.Countries))
	for _, country := range apiResp.Data.Countries {
		countries = append(countries, country.Name)
	}

	// Process languages
	languages := make([]string, 0, len(apiResp.Data.Languages))
	for _, language := range apiResp.Data.Languages {
		languages = append(languages, language.Name)
	}

	// Process seasons
	seasons := make([]TVDBSeason, 0, len(apiResp.Data.Seasons))
	for _, season := range apiResp.Data.Seasons {
		seasons = append(seasons, TVDBSeason{
			ID:           season.ID,
			Name:         season.Name,
			Number:       season.Number,
			EpisodeCount: season.EpisodeCount,
			Overview:     season.Overview,
			PosterURL:    season.PosterURL,
		})
	}

	// Create result
	result := &TVDBSeriesDetails{
		ID:            apiResp.Data.ID,
		Name:          apiResp.Data.Name,
		Overview:      apiResp.Data.Overview,
		FirstAiredYear: firstAiredYear,
		Status:        apiResp.Data.Status,
		Network:       apiResp.Data.Network,
		ImdbID:        apiResp.Data.ImdbID,
		PosterURL:     apiResp.Data.PosterURL,
		BackdropURL:   apiResp.Data.BackdropURL,
		Seasons:       seasons,
		Genres:        genres,
		Countries:     countries,
		Languages:     languages,
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tvdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Cache for 7 days
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetSeasonEpisodes gets episodes for a TV series season
func (c *TVDBClient) GetSeasonEpisodes(ctx context.Context, seasonID int) (*TVDBSeasonEpisodes, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("season_episodes:%d", seasonID)
	cache, err := c.db.GetAPICache("tvdb", cacheKey)
	if err == nil {
		// Cache hit
		var result TVDBSeasonEpisodes
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	endpoint := fmt.Sprintf("%s/seasons/%d/extended", c.baseURL, seasonID)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Accept", "application/json")
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
		return nil, fmt.Errorf("TVDB API error: %s - %s", resp.Status, string(body))
	}

	var apiResp struct {
		Data struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			Number   int    `json:"number"`
			SeriesID int    `json:"series_id"`
			Episodes []struct {
				ID            int    `json:"id"`
				Name          string `json:"name"`
				Overview      string `json:"overview"`
				EpisodeNumber int    `json:"episode_number"`
				SeasonNumber  int    `json:"season_number"`
				AirDate       string `json:"air_date"`
				ImageURL      string `json:"image"`
			} `json:"episodes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Process episodes
	episodes := make([]TVDBEpisode, 0, len(apiResp.Data.Episodes))
	for _, episode := range apiResp.Data.Episodes {
		episodes = append(episodes, TVDBEpisode{
			ID:            episode.ID,
			Name:          episode.Name,
			Overview:      episode.Overview,
			EpisodeNumber: episode.EpisodeNumber,
			SeasonNumber:  episode.SeasonNumber,
			AirDate:       episode.AirDate,
			ImageURL:      episode.ImageURL,
		})
	}

	// Create result
	result := &TVDBSeasonEpisodes{
		ID:       apiResp.Data.ID,
		Name:     apiResp.Data.Name,
		Number:   apiResp.Data.Number,
		SeriesID: apiResp.Data.SeriesID,
		Episodes: episodes,
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tvdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Cache for 7 days
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// TVDBSeries represents a TV series search result
type TVDBSeries struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	FirstAiredYear int    `json:"first_aired_year"`
	Overview       string `json:"overview"`
	PosterURL      string `json:"poster_url"`
	BackdropURL    string `json:"backdrop_url"`
	Status         string `json:"status"`
	Network        string `json:"network"`
	TVDBScore      int    `json:"tvdb_score"`
}

// TVDBSearchResult represents a TV series search result
type TVDBSearchResult struct {
	Query  string       `json:"query"`
	Series []TVDBSeries `json:"series"`
}

// TVDBSeriesDetails represents detailed information about a TV series
type TVDBSeriesDetails struct {
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Overview       string      `json:"overview"`
	FirstAiredYear int         `json:"first_aired_year"`
	Status         string      `json:"status"`
	Network        string      `json:"network"`
	ImdbID         string      `json:"imdb_id"`
	PosterURL      string      `json:"poster_url"`
	BackdropURL    string      `json:"backdrop_url"`
	Seasons        []TVDBSeason `json:"seasons"`
	Genres         []string    `json:"genres"`
	Countries      []string    `json:"countries"`
	Languages      []string    `json:"languages"`
}

// TVDBSeason represents a TV series season
type TVDBSeason struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Number       int    `json:"number"`
	EpisodeCount int    `json:"episode_count"`
	Overview     string `json:"overview"`
	PosterURL    string `json:"poster_url"`
}

// TVDBSeasonEpisodes represents episodes for a TV series season
type TVDBSeasonEpisodes struct {
	ID       int          `json:"id"`
	Name     string       `json:"name"`
	Number   int          `json:"number"`
	SeriesID int          `json:"series_id"`
	Episodes []TVDBEpisode `json:"episodes"`
}

// TVDBEpisode represents a TV series episode
type TVDBEpisode struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	EpisodeNumber int    `json:"episode_number"`
	SeasonNumber  int    `json:"season_number"`
	AirDate       string `json:"air_date"`
	ImageURL      string `json:"image_url"`
}
