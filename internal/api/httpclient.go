package api

import (
	"net"
	"net/http"
	"time"
)

// HTTPClientConfig represents HTTP client configuration
type HTTPClientConfig struct {
	Timeout               time.Duration
	KeepAlive             time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	MaxConnsPerHost       int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
}

// DefaultHTTPClientConfig returns default HTTP client configuration
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:               30 * time.Second,
		KeepAlive:             30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewOptimizedHTTPClient creates an optimized HTTP client with connection pooling
func NewOptimizedHTTPClient(config *HTTPClientConfig) *http.Client {
	if config == nil {
		config = DefaultHTTPClientConfig()
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
}

// APISpecificHTTPClientConfig returns HTTP client configurations optimized for specific APIs
func APISpecificHTTPClientConfig(apiName string) *HTTPClientConfig {
	base := DefaultHTTPClientConfig()
	
	switch apiName {
	case "tmdb":
		// TMDB is generally fast and reliable
		base.Timeout = 15 * time.Second
		base.MaxConnsPerHost = 5
		base.MaxIdleConnsPerHost = 3
	case "tvdb":
		// TVDB can be slower
		base.Timeout = 20 * time.Second
		base.MaxConnsPerHost = 3
		base.MaxIdleConnsPerHost = 2
	case "bangumi":
		// Bangumi has stricter rate limits
		base.Timeout = 25 * time.Second
		base.MaxConnsPerHost = 2
		base.MaxIdleConnsPerHost = 1
	}
	
	return base
}
