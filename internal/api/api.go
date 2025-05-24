package api

import (
	"fmt"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
	"github.com/sleepstars/mediascanner/internal/ratelimiter"
)

// API represents the API clients
type API struct {
	TMDB    *TMDBClient
	TVDB    *TVDBClient
	Bangumi *BangumiClient

	// Rate limiter for API requests
	RateLimiter *ratelimiter.ProviderRateLimiter
}

// New creates a new API instance
func New(cfg *config.APIConfig, db *database.Database) (*API, error) {
	// Create rate limiter
	rateLimiter := ratelimiter.NewProviderRateLimiter()

	// Configure rate limiters if enabled
	if cfg.RateLimiting.Enabled {
		// Register rate limiters for each provider
		rateLimiter.RegisterLimiter("tmdb", ratelimiter.NewTokenBucketRateLimiter(
			cfg.RateLimiting.TMDB, float64(cfg.RateLimiting.TMDBBurst)))

		rateLimiter.RegisterLimiter("tvdb", ratelimiter.NewTokenBucketRateLimiter(
			cfg.RateLimiting.TVDB, float64(cfg.RateLimiting.TVDBBurst)))

		rateLimiter.RegisterLimiter("bangumi", ratelimiter.NewTokenBucketRateLimiter(
			cfg.RateLimiting.Bangumi, float64(cfg.RateLimiting.BangumiBurst)))
	}

	// Create API clients
	tmdbClient, err := NewTMDBClient(&cfg.TMDB, db, rateLimiter, &cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create TMDB client: %w", err)
	}

	tvdbClient, err := NewTVDBClient(&cfg.TVDB, db, rateLimiter, &cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create TVDB client: %w", err)
	}

	bangumiClient, err := NewBangumiClient(&cfg.Bangumi, db, rateLimiter, &cfg.Cache)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bangumi client: %w", err)
	}

	return &API{
		TMDB:        tmdbClient,
		TVDB:        tvdbClient,
		Bangumi:     bangumiClient,
		RateLimiter: rateLimiter,
	}, nil
}
