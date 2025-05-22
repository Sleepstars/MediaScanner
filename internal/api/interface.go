package api

import (
	"context"
)

// APIInterface defines the interface for API operations
type APIInterface interface {
	SearchMovie(ctx context.Context, query string, year int) (*MovieSearchResult, error)
	SearchTV(ctx context.Context, query string, year int) (*TVSearchResult, error)
	GetMovieDetails(ctx context.Context, id int) (*MovieDetails, error)
	GetTVDetails(ctx context.Context, id int) (*TVDetails, error)
	GetSeasonDetails(ctx context.Context, tvID, seasonNumber int) (*SeasonDetails, error)
	GetEpisodeDetails(ctx context.Context, tvID, seasonNumber, episodeNumber int) (*EpisodeDetails, error)
}
