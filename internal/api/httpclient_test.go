package api

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultHTTPClientConfig(t *testing.T) {
	config := DefaultHTTPClientConfig()
	
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout to be 30s, got %v", config.Timeout)
	}
	
	if config.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns to be 100, got %d", config.MaxIdleConns)
	}
	
	if config.MaxIdleConnsPerHost != 10 {
		t.Errorf("Expected MaxIdleConnsPerHost to be 10, got %d", config.MaxIdleConnsPerHost)
	}
}

func TestAPISpecificHTTPClientConfig(t *testing.T) {
	tests := []struct {
		apiName             string
		expectedTimeout     time.Duration
		expectedMaxConns    int
		expectedMaxIdle     int
	}{
		{"tmdb", 15 * time.Second, 5, 3},
		{"tvdb", 20 * time.Second, 3, 2},
		{"bangumi", 25 * time.Second, 2, 1},
		{"unknown", 30 * time.Second, 10, 10}, // Should use defaults
	}
	
	for _, tt := range tests {
		t.Run(tt.apiName, func(t *testing.T) {
			config := APISpecificHTTPClientConfig(tt.apiName)
			
			if config.Timeout != tt.expectedTimeout {
				t.Errorf("Expected timeout to be %v, got %v", tt.expectedTimeout, config.Timeout)
			}
			
			if config.MaxConnsPerHost != tt.expectedMaxConns {
				t.Errorf("Expected MaxConnsPerHost to be %d, got %d", tt.expectedMaxConns, config.MaxConnsPerHost)
			}
			
			if config.MaxIdleConnsPerHost != tt.expectedMaxIdle {
				t.Errorf("Expected MaxIdleConnsPerHost to be %d, got %d", tt.expectedMaxIdle, config.MaxIdleConnsPerHost)
			}
		})
	}
}

func TestNewOptimizedHTTPClient(t *testing.T) {
	config := &HTTPClientConfig{
		Timeout:               15 * time.Second,
		KeepAlive:             30 * time.Second,
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   5,
		MaxConnsPerHost:       5,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	
	client := NewOptimizedHTTPClient(config)
	
	if client.Timeout != 15*time.Second {
		t.Errorf("Expected client timeout to be 15s, got %v", client.Timeout)
	}
	
	// Verify that the transport is properly configured
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected client transport to be *http.Transport")
	}
	
	if transport.MaxIdleConns != 50 {
		t.Errorf("Expected MaxIdleConns to be 50, got %d", transport.MaxIdleConns)
	}
	
	if transport.MaxIdleConnsPerHost != 5 {
		t.Errorf("Expected MaxIdleConnsPerHost to be 5, got %d", transport.MaxIdleConnsPerHost)
	}
	
	if transport.MaxConnsPerHost != 5 {
		t.Errorf("Expected MaxConnsPerHost to be 5, got %d", transport.MaxConnsPerHost)
	}
}

func TestNewOptimizedHTTPClientWithNilConfig(t *testing.T) {
	client := NewOptimizedHTTPClient(nil)
	
	// Should use default configuration
	if client.Timeout != 30*time.Second {
		t.Errorf("Expected client timeout to be 30s, got %v", client.Timeout)
	}
	
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected client transport to be *http.Transport")
	}
	
	if transport.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns to be 100, got %d", transport.MaxIdleConns)
	}
}

func BenchmarkNewOptimizedHTTPClient(b *testing.B) {
	config := DefaultHTTPClientConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewOptimizedHTTPClient(config)
	}
}

func BenchmarkAPISpecificHTTPClientConfig(b *testing.B) {
	apis := []string{"tmdb", "tvdb", "bangumi"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		api := apis[i%len(apis)]
		_ = APISpecificHTTPClientConfig(api)
	}
}
