package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/models"
)

// TMDBClient represents the TMDB API client
type TMDBClient struct {
	client *tmdb.Client
	config *config.TMDBConfig
	db     *database.Database
}

// NewTMDBClient creates a new TMDB API client
func NewTMDBClient(cfg *config.TMDBConfig, db *database.Database) (*TMDBClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("TMDB API key is required")
	}

	client, err := tmdb.Init(cfg.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize TMDB client: %w", err)
	}

	// Enable auto retry
	client.SetClientAutoRetry()

	return &TMDBClient{
		client: client,
		config: cfg,
		db:     db,
	}, nil
}

// SearchMovie searches for a movie
func (c *TMDBClient) SearchMovie(ctx context.Context, query string, year int) (*MovieSearchResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("movie:%s:%d", query, year)
	cache, err := c.db.GetAPICache("tmdb", cacheKey)
	if err == nil {
		// Cache hit
		var result MovieSearchResult
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	options := map[string]string{
		"language": c.config.Language,
	}
	if year > 0 {
		options["year"] = fmt.Sprintf("%d", year)
	}
	if !c.config.IncludeAdult {
		options["include_adult"] = "false"
	}

	search, err := c.client.GetSearchMovies(query, options)
	if err != nil {
		return nil, fmt.Errorf("TMDB search movie error: %w", err)
	}

	// Process results
	result := &MovieSearchResult{
		Query:  query,
		Year:   year,
		Movies: make([]Movie, 0),
	}

	for _, movie := range search.Results {
		// Extract year from release date
		var releaseYear int
		if movie.ReleaseDate != "" {
			t, err := time.Parse("2006-01-02", movie.ReleaseDate)
			if err == nil {
				releaseYear = t.Year()
			}
		}

		result.Movies = append(result.Movies, Movie{
			ID:            movie.ID,
			Title:         movie.Title,
			OriginalTitle: movie.OriginalTitle,
			ReleaseYear:   releaseYear,
			Overview:      movie.Overview,
			PosterPath:    movie.PosterPath,
			BackdropPath:  movie.BackdropPath,
			Popularity:    movie.Popularity,
			VoteAverage:   movie.VoteAverage,
		})
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tmdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(24 * time.Hour), // Cache for 24 hours
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// SearchTV searches for a TV show
func (c *TMDBClient) SearchTV(ctx context.Context, query string, year int) (*TVSearchResult, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("tv:%s:%d", query, year)
	cache, err := c.db.GetAPICache("tmdb", cacheKey)
	if err == nil {
		// Cache hit
		var result TVSearchResult
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	options := map[string]string{
		"language": c.config.Language,
	}
	if year > 0 {
		options["first_air_date_year"] = fmt.Sprintf("%d", year)
	}
	if !c.config.IncludeAdult {
		options["include_adult"] = "false"
	}

	search, err := c.client.GetSearchTVShow(query, options)
	if err != nil {
		return nil, fmt.Errorf("TMDB search TV error: %w", err)
	}

	// Process results
	result := &TVSearchResult{
		Query: query,
		Year:  year,
		Shows: make([]TVShow, 0),
	}

	for _, show := range search.Results {
		// Extract year from first air date
		var firstAirYear int
		if show.FirstAirDate != "" {
			t, err := time.Parse("2006-01-02", show.FirstAirDate)
			if err == nil {
				firstAirYear = t.Year()
			}
		}

		result.Shows = append(result.Shows, TVShow{
			ID:           show.ID,
			Name:         show.Name,
			OriginalName: show.OriginalName,
			FirstAirYear: firstAirYear,
			Overview:     show.Overview,
			PosterPath:   show.PosterPath,
			BackdropPath: show.BackdropPath,
			Popularity:   show.Popularity,
			VoteAverage:  show.VoteAverage,
		})
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tmdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(24 * time.Hour), // Cache for 24 hours
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetMovieDetails gets details for a movie
func (c *TMDBClient) GetMovieDetails(ctx context.Context, id int) (*MovieDetails, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("movie_details:%d", id)
	cache, err := c.db.GetAPICache("tmdb", cacheKey)
	if err == nil {
		// Cache hit
		var result MovieDetails
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	options := map[string]string{
		"language":           c.config.Language,
		"append_to_response": "credits,images,external_ids",
	}

	movie, err := c.client.GetMovieDetails(id, options)
	if err != nil {
		return nil, fmt.Errorf("TMDB get movie details error: %w", err)
	}

	// Extract year from release date
	var releaseYear int
	if movie.ReleaseDate != "" {
		t, err := time.Parse("2006-01-02", movie.ReleaseDate)
		if err == nil {
			releaseYear = t.Year()
		}
	}

	// Process genres
	genres := make([]string, 0, len(movie.Genres))
	for _, genre := range movie.Genres {
		genres = append(genres, genre.Name)
	}

	// Process production countries
	countries := make([]string, 0, len(movie.ProductionCountries))
	for _, country := range movie.ProductionCountries {
		countries = append(countries, country.Name)
	}

	// Process spoken languages
	languages := make([]string, 0, len(movie.SpokenLanguages))
	for _, language := range movie.SpokenLanguages {
		languages = append(languages, language.Name)
	}

	// Create result
	result := &MovieDetails{
		ID:            movie.ID,
		Title:         movie.Title,
		OriginalTitle: movie.OriginalTitle,
		ReleaseYear:   releaseYear,
		Overview:      movie.Overview,
		PosterPath:    movie.PosterPath,
		BackdropPath:  movie.BackdropPath,
		// Get ImdbID from movie details
		ImdbID:      movie.IMDbID,
		Genres:      genres,
		Countries:   countries,
		Languages:   languages,
		Runtime:     movie.Runtime,
		VoteAverage: movie.VoteAverage,
		VoteCount:   movie.VoteCount,
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tmdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Cache for 7 days
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetTVDetails gets details for a TV show
func (c *TMDBClient) GetTVDetails(ctx context.Context, id int) (*TVDetails, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("tv_details:%d", id)
	cache, err := c.db.GetAPICache("tmdb", cacheKey)
	if err == nil {
		// Cache hit
		var result TVDetails
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	options := map[string]string{
		"language":           c.config.Language,
		"append_to_response": "credits,images,external_ids",
	}

	tv, err := c.client.GetTVDetails(id, options)
	if err != nil {
		return nil, fmt.Errorf("TMDB get TV details error: %w", err)
	}

	// Extract year from first air date
	var firstAirYear int
	if tv.FirstAirDate != "" {
		t, err := time.Parse("2006-01-02", tv.FirstAirDate)
		if err == nil {
			firstAirYear = t.Year()
		}
	}

	// Process genres
	genres := make([]string, 0, len(tv.Genres))
	for _, genre := range tv.Genres {
		genres = append(genres, genre.Name)
	}

	// Process production countries
	countries := make([]string, 0, len(tv.ProductionCountries))
	for _, country := range tv.ProductionCountries {
		countries = append(countries, country.Name)
	}

	// Process spoken languages
	languages := make([]string, 0)
	// Note: SpokenLanguages field is not directly accessible in the current version of the library
	// We would need to use GetTVContentRatings or other methods to get language information

	// Create result
	result := &TVDetails{
		ID:           tv.ID,
		Name:         tv.Name,
		OriginalName: tv.OriginalName,
		FirstAirYear: firstAirYear,
		Overview:     tv.Overview,
		PosterPath:   tv.PosterPath,
		BackdropPath: tv.BackdropPath,
		// Note: ExternalIDs field is not directly accessible in the current version of the library
		// We would need to use GetTVExternalIDs method to get this information
		ImdbID:          "", // Placeholder
		TVDBID:          0,  // Placeholder
		Genres:          genres,
		Countries:       countries,
		Languages:       languages,
		NumberOfSeasons: tv.NumberOfSeasons,
		VoteAverage:     tv.VoteAverage,
		VoteCount:       tv.VoteCount,
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tmdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Cache for 7 days
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetSeasonDetails gets details for a TV show season
func (c *TMDBClient) GetSeasonDetails(ctx context.Context, tvID, seasonNumber int) (*SeasonDetails, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("season_details:%d:%d", tvID, seasonNumber)
	cache, err := c.db.GetAPICache("tmdb", cacheKey)
	if err == nil {
		// Cache hit
		var result SeasonDetails
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	options := map[string]string{
		"language": c.config.Language,
	}

	season, err := c.client.GetTVSeasonDetails(tvID, seasonNumber, options)
	if err != nil {
		return nil, fmt.Errorf("TMDB get season details error: %w", err)
	}

	// Process episodes
	episodes := make([]Episode, 0, len(season.Episodes))
	for _, episode := range season.Episodes {
		episodes = append(episodes, Episode{
			ID:            episode.ID,
			Name:          episode.Name,
			Overview:      episode.Overview,
			EpisodeNumber: episode.EpisodeNumber,
			SeasonNumber:  episode.SeasonNumber,
			StillPath:     episode.StillPath,
			AirDate:       episode.AirDate,
			VoteAverage:   episode.VoteAverage,
		})
	}

	// Create result
	result := &SeasonDetails{
		ID:           season.ID,
		Name:         season.Name,
		Overview:     season.Overview,
		SeasonNumber: season.SeasonNumber,
		PosterPath:   season.PosterPath,
		AirDate:      season.AirDate,
		Episodes:     episodes,
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tmdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Cache for 7 days
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetEpisodeDetails gets details for a TV show episode
func (c *TMDBClient) GetEpisodeDetails(ctx context.Context, tvID, seasonNumber, episodeNumber int) (*EpisodeDetails, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("episode_details:%d:%d:%d", tvID, seasonNumber, episodeNumber)
	cache, err := c.db.GetAPICache("tmdb", cacheKey)
	if err == nil {
		// Cache hit
		var result EpisodeDetails
		if err := json.Unmarshal([]byte(cache.Response), &result); err == nil {
			return &result, nil
		}
	}

	// Cache miss, perform API call
	options := map[string]string{
		"language": c.config.Language,
	}

	episode, err := c.client.GetTVEpisodeDetails(tvID, seasonNumber, episodeNumber, options)
	if err != nil {
		return nil, fmt.Errorf("TMDB get episode details error: %w", err)
	}

	// Create result
	result := &EpisodeDetails{
		ID:            episode.ID,
		Name:          episode.Name,
		Overview:      episode.Overview,
		EpisodeNumber: episode.EpisodeNumber,
		SeasonNumber:  episode.SeasonNumber,
		StillPath:     episode.StillPath,
		AirDate:       episode.AirDate,
		VoteAverage:   episode.VoteAverage,
	}

	// Cache the result
	resultJSON, err := json.Marshal(result)
	if err == nil {
		cache := &models.APICache{
			Provider:  "tmdb",
			Query:     cacheKey,
			Response:  string(resultJSON),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Cache for 7 days
		}
		_ = c.db.CreateAPICache(cache)
	}

	return result, nil
}

// GetImageURL gets the full URL for an image
func (c *TMDBClient) GetImageURL(path string, size string) string {
	return tmdb.GetImageURL(path, size)
}

// Movie represents a movie search result
type Movie struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	ReleaseYear   int     `json:"release_year"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	Popularity    float32 `json:"popularity"`
	VoteAverage   float32 `json:"vote_average"`
}

// TVShow represents a TV show search result
type TVShow struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	OriginalName string  `json:"original_name"`
	FirstAirYear int     `json:"first_air_year"`
	Overview     string  `json:"overview"`
	PosterPath   string  `json:"poster_path"`
	BackdropPath string  `json:"backdrop_path"`
	Popularity   float32 `json:"popularity"`
	VoteAverage  float32 `json:"vote_average"`
}

// MovieSearchResult represents a movie search result
type MovieSearchResult struct {
	Query  string  `json:"query"`
	Year   int     `json:"year"`
	Movies []Movie `json:"movies"`
}

// TVSearchResult represents a TV show search result
type TVSearchResult struct {
	Query string   `json:"query"`
	Year  int      `json:"year"`
	Shows []TVShow `json:"shows"`
}

// MovieDetails represents detailed information about a movie
type MovieDetails struct {
	ID            int64    `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"original_title"`
	ReleaseYear   int      `json:"release_year"`
	Overview      string   `json:"overview"`
	PosterPath    string   `json:"poster_path"`
	BackdropPath  string   `json:"backdrop_path"`
	ImdbID        string   `json:"imdb_id"`
	Genres        []string `json:"genres"`
	Countries     []string `json:"countries"`
	Languages     []string `json:"languages"`
	Runtime       int      `json:"runtime"`
	VoteAverage   float32  `json:"vote_average"`
	VoteCount     int64    `json:"vote_count"`
}

// TVDetails represents detailed information about a TV show
type TVDetails struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	OriginalName    string   `json:"original_name"`
	FirstAirYear    int      `json:"first_air_year"`
	Overview        string   `json:"overview"`
	PosterPath      string   `json:"poster_path"`
	BackdropPath    string   `json:"backdrop_path"`
	ImdbID          string   `json:"imdb_id"`
	TVDBID          int      `json:"tvdb_id"`
	Genres          []string `json:"genres"`
	Countries       []string `json:"countries"`
	Languages       []string `json:"languages"`
	NumberOfSeasons int      `json:"number_of_seasons"`
	VoteAverage     float32  `json:"vote_average"`
	VoteCount       int64    `json:"vote_count"`
}

// SeasonDetails represents detailed information about a TV show season
type SeasonDetails struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Overview     string    `json:"overview"`
	SeasonNumber int       `json:"season_number"`
	PosterPath   string    `json:"poster_path"`
	AirDate      string    `json:"air_date"`
	Episodes     []Episode `json:"episodes"`
}

// EpisodeDetails represents detailed information about a TV show episode
type EpisodeDetails struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	StillPath     string  `json:"still_path"`
	AirDate       string  `json:"air_date"`
	VoteAverage   float32 `json:"vote_average"`
}

// Episode represents a TV show episode
type Episode struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	StillPath     string  `json:"still_path"`
	AirDate       string  `json:"air_date"`
	VoteAverage   float32 `json:"vote_average"`
}
