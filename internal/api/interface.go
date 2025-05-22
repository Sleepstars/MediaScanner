package api

import (
	"context"
	tmdb "github.com/cyruzin/golang-tmdb"
)

// TMDBClientInterface defines the interface for TMDB client operations
type TMDBClientInterface interface {
	GetSearchMovies(query string, options map[string]string) (*tmdb.SearchMovies, error)
	GetSearchTVShow(query string, options map[string]string) (*tmdb.SearchTVShows, error)
	GetMovieDetails(id int, options map[string]string) (*tmdb.MovieDetails, error)
	GetTVDetails(id int, options map[string]string) (*tmdb.TVDetails, error)
	GetTVSeasonDetails(id, seasonNumber int, options map[string]string) (*tmdb.TVSeasonDetails, error)
	GetTVEpisodeDetails(id, seasonNumber, episodeNumber int, options map[string]string) (*tmdb.TVEpisodeDetails, error)
	SetClientAutoRetry()
}

// APIInterface defines the interface for API operations
type APIInterface interface {
	SearchMovie(ctx context.Context, query string, year int) (*MovieSearchResult, error)
	SearchTV(ctx context.Context, query string, year int) (*TVSearchResult, error)
	GetMovieDetails(ctx context.Context, id int) (*MovieDetails, error)
	GetTVDetails(ctx context.Context, id int) (*TVDetails, error)
	GetSeasonDetails(ctx context.Context, tvID, seasonNumber int) (*SeasonDetails, error)
	GetEpisodeDetails(ctx context.Context, tvID, seasonNumber, episodeNumber int) (*EpisodeDetails, error)
}
