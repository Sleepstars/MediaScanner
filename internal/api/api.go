package api

import (
	"fmt"

	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/database"
)

// API represents the API clients
type API struct {
	TMDB    *TMDBClient
	TVDB    *TVDBClient
	Bangumi *BangumiClient
}

// New creates a new API instance
func New(cfg *config.APIConfig, db *database.Database) (*API, error) {
	tmdbClient, err := NewTMDBClient(&cfg.TMDB, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create TMDB client: %w", err)
	}

	tvdbClient, err := NewTVDBClient(&cfg.TVDB, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create TVDB client: %w", err)
	}

	bangumiClient, err := NewBangumiClient(&cfg.Bangumi, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create Bangumi client: %w", err)
	}

	return &API{
		TMDB:    tmdbClient,
		TVDB:    tvdbClient,
		Bangumi: bangumiClient,
	}, nil
}
