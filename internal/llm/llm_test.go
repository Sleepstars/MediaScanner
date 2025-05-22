package llm

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/sleepstars/mediascanner/internal/config"
	"github.com/sleepstars/mediascanner/internal/worker"
)

// MockOpenAIClient is a mock implementation of the OpenAI client
type MockOpenAIClient struct {
	CreateChatCompletionFunc func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

func (m *MockOpenAIClient) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return m.CreateChatCompletionFunc(ctx, request)
}

// MockSemaphore is a mock implementation of the worker.Semaphore interface
type MockSemaphore struct {
	AcquireFunc func(ctx context.Context) error
	ReleaseFunc func()
}

func (m *MockSemaphore) Acquire(ctx context.Context) error {
	return m.AcquireFunc(ctx)
}

func (m *MockSemaphore) Release() {
	m.ReleaseFunc()
}

func TestNew(t *testing.T) {
	// Test with valid config
	cfg := &config.LLMConfig{
		APIKey: "test-api-key",
		Model:  "gpt-3.5-turbo",
	}
	
	llm, err := New(cfg, nil)
	if err != nil {
		t.Fatalf("New failed with valid config: %v", err)
	}
	
	if llm == nil {
		t.Fatal("Expected non-nil LLM instance")
	}
	
	// Test with empty API key
	cfg = &config.LLMConfig{
		APIKey: "",
		Model:  "gpt-3.5-turbo",
	}
	
	_, err = New(cfg, nil)
	if err == nil {
		t.Fatal("Expected error for empty API key, got nil")
	}
	
	expectedError := "LLM API key is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestRegisterFunction(t *testing.T) {
	cfg := &config.LLMConfig{
		APIKey: "test-api-key",
		Model:  "gpt-3.5-turbo",
	}
	
	llm, _ := New(cfg, nil)
	
	// Register a function
	handlerCalled := false
	handler := func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		handlerCalled = true
		return map[string]string{"result": "success"}, nil
	}
	
	llm.RegisterFunction("testFunction", handler)
	
	// Verify the function was registered
	if _, ok := llm.functionMap["testFunction"]; !ok {
		t.Error("Function was not registered")
	}
}

func TestProcessMediaFile(t *testing.T) {
	// Create a mock OpenAI client
	mockClient := &MockOpenAIClient{
		CreateChatCompletionFunc: func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			// Return a valid response
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `{
								"original_filename": "The.Matrix.1999.1080p.BluRay.x264.mp4",
								"title": "The Matrix",
								"original_title": "The Matrix",
								"year": 1999,
								"media_type": "movie",
								"tmdb_id": 603,
								"category": "movie",
								"subcategory": "sci-fi",
								"destination_path": "/movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4",
								"confidence": 0.95
							}`,
						},
						FinishReason: openai.FinishReasonStop,
					},
				},
			}, nil
		},
	}
	
	// Create a mock semaphore
	acquireCalled := false
	releaseCalled := false
	mockSemaphore := &MockSemaphore{
		AcquireFunc: func(ctx context.Context) error {
			acquireCalled = true
			return nil
		},
		ReleaseFunc: func() {
			releaseCalled = true
		},
	}
	
	// Create LLM with mocks
	cfg := &config.LLMConfig{
		APIKey:       "test-api-key",
		Model:        "gpt-3.5-turbo",
		SystemPrompt: "You are a helpful assistant that analyzes media filenames.",
		MaxRetries:   3,
	}
	
	llm := &LLM{
		client:      mockClient,
		config:      cfg,
		functionMap: make(map[string]FunctionHandler),
		semaphore:   mockSemaphore,
	}
	
	// Process a media file
	filename := "The.Matrix.1999.1080p.BluRay.x264.mp4"
	directoryStructure := map[string][]string{
		"/movies": {"Action", "Sci-Fi", "Drama"},
	}
	
	result, err := llm.ProcessMediaFile(context.Background(), filename, directoryStructure)
	if err != nil {
		t.Fatalf("ProcessMediaFile failed: %v", err)
	}
	
	// Verify semaphore was used
	if !acquireCalled {
		t.Error("Semaphore Acquire was not called")
	}
	
	if !releaseCalled {
		t.Error("Semaphore Release was not called")
	}
	
	// Verify the result
	if result.Title != "The Matrix" {
		t.Errorf("Expected title to be 'The Matrix', got %q", result.Title)
	}
	
	if result.Year != 1999 {
		t.Errorf("Expected year to be 1999, got %d", result.Year)
	}
	
	if result.MediaType != "movie" {
		t.Errorf("Expected media type to be 'movie', got %q", result.MediaType)
	}
}

func TestProcessMediaFile_Error(t *testing.T) {
	// Create a mock OpenAI client that returns an error
	mockClient := &MockOpenAIClient{
		CreateChatCompletionFunc: func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			return openai.ChatCompletionResponse{}, errors.New("API error")
		},
	}
	
	// Create a mock semaphore
	mockSemaphore := &MockSemaphore{
		AcquireFunc: func(ctx context.Context) error {
			return nil
		},
		ReleaseFunc: func() {},
	}
	
	// Create LLM with mocks
	cfg := &config.LLMConfig{
		APIKey:       "test-api-key",
		Model:        "gpt-3.5-turbo",
		SystemPrompt: "You are a helpful assistant that analyzes media filenames.",
		MaxRetries:   0, // No retries for faster test
	}
	
	llm := &LLM{
		client:      mockClient,
		config:      cfg,
		functionMap: make(map[string]FunctionHandler),
		semaphore:   mockSemaphore,
	}
	
	// Process a media file
	filename := "The.Matrix.1999.1080p.BluRay.x264.mp4"
	directoryStructure := map[string][]string{}
	
	_, err := llm.ProcessMediaFile(context.Background(), filename, directoryStructure)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestProcessMediaFile_FunctionCall(t *testing.T) {
	// Create a mock OpenAI client that returns a function call
	firstCall := true
	mockClient := &MockOpenAIClient{
		CreateChatCompletionFunc: func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			if firstCall {
				firstCall = false
				return openai.ChatCompletionResponse{
					Choices: []openai.ChatCompletionChoice{
						{
							Message: openai.ChatCompletionMessage{
								Content: "",
								FunctionCall: &openai.FunctionCall{
									Name:      "searchTMDB",
									Arguments: `{"query":"The Matrix","year":1999,"mediaType":"movie"}`,
								},
							},
							FinishReason: openai.FinishReasonFunctionCall,
						},
					},
				}, nil
			}
			
			// Second call returns the final result
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `{
								"original_filename": "The.Matrix.1999.1080p.BluRay.x264.mp4",
								"title": "The Matrix",
								"original_title": "The Matrix",
								"year": 1999,
								"media_type": "movie",
								"tmdb_id": 603,
								"category": "movie",
								"subcategory": "sci-fi",
								"destination_path": "/movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4",
								"confidence": 0.95
							}`,
						},
						FinishReason: openai.FinishReasonStop,
					},
				},
			}, nil
		},
	}
	
	// Create a mock semaphore
	mockSemaphore := &MockSemaphore{
		AcquireFunc: func(ctx context.Context) error {
			return nil
		},
		ReleaseFunc: func() {},
	}
	
	// Create LLM with mocks
	cfg := &config.LLMConfig{
		APIKey:       "test-api-key",
		Model:        "gpt-3.5-turbo",
		SystemPrompt: "You are a helpful assistant that analyzes media filenames.",
		MaxRetries:   0,
	}
	
	llm := &LLM{
		client:      mockClient,
		config:      cfg,
		functionMap: make(map[string]FunctionHandler),
		semaphore:   mockSemaphore,
	}
	
	// Register the function
	functionCalled := false
	llm.RegisterFunction("searchTMDB", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		functionCalled = true
		
		// Parse the arguments
		var params struct {
			Query     string `json:"query"`
			Year      int    `json:"year"`
			MediaType string `json:"mediaType"`
		}
		
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		
		// Verify the arguments
		if params.Query != "The Matrix" {
			t.Errorf("Expected query to be 'The Matrix', got %q", params.Query)
		}
		
		if params.Year != 1999 {
			t.Errorf("Expected year to be 1999, got %d", params.Year)
		}
		
		if params.MediaType != "movie" {
			t.Errorf("Expected media type to be 'movie', got %q", params.MediaType)
		}
		
		// Return a mock result
		return map[string]interface{}{
			"movies": []map[string]interface{}{
				{
					"id":            603,
					"title":         "The Matrix",
					"original_title": "The Matrix",
					"release_year":  1999,
					"overview":      "Set in the 22nd century, The Matrix tells the story of a computer hacker who joins a group of underground insurgents fighting the vast and powerful computers who now rule the earth.",
				},
			},
		}, nil
	})
	
	// Process a media file
	filename := "The.Matrix.1999.1080p.BluRay.x264.mp4"
	directoryStructure := map[string][]string{}
	
	result, err := llm.ProcessMediaFile(context.Background(), filename, directoryStructure)
	if err != nil {
		t.Fatalf("ProcessMediaFile failed: %v", err)
	}
	
	// Verify the function was called
	if !functionCalled {
		t.Error("Function was not called")
	}
	
	// Verify the result
	if result.Title != "The Matrix" {
		t.Errorf("Expected title to be 'The Matrix', got %q", result.Title)
	}
}

func TestProcessBatchFiles(t *testing.T) {
	// Create a mock OpenAI client
	mockClient := &MockOpenAIClient{
		CreateChatCompletionFunc: func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			// Return a valid response
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `[
								{
									"original_filename": "The.Matrix.1999.1080p.BluRay.x264.mp4",
									"title": "The Matrix",
									"original_title": "The Matrix",
									"year": 1999,
									"media_type": "movie",
									"tmdb_id": 603,
									"category": "movie",
									"subcategory": "sci-fi",
									"destination_path": "/movies/Sci-Fi/The Matrix (1999)/The.Matrix.1999.1080p.BluRay.x264.mp4",
									"confidence": 0.95
								},
								{
									"original_filename": "The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4",
									"title": "The Matrix Reloaded",
									"original_title": "The Matrix Reloaded",
									"year": 2003,
									"media_type": "movie",
									"tmdb_id": 604,
									"category": "movie",
									"subcategory": "sci-fi",
									"destination_path": "/movies/Sci-Fi/The Matrix Reloaded (2003)/The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4",
									"confidence": 0.95
								}
							]`,
						},
						FinishReason: openai.FinishReasonStop,
					},
				},
			}, nil
		},
	}
	
	// Create a mock semaphore
	mockSemaphore := &MockSemaphore{
		AcquireFunc: func(ctx context.Context) error {
			return nil
		},
		ReleaseFunc: func() {},
	}
	
	// Create LLM with mocks
	cfg := &config.LLMConfig{
		APIKey:           "test-api-key",
		Model:            "gpt-3.5-turbo",
		BatchSystemPrompt: "You are a helpful assistant that analyzes batches of media filenames.",
		MaxRetries:       0,
	}
	
	llm := &LLM{
		client:      mockClient,
		config:      cfg,
		functionMap: make(map[string]FunctionHandler),
		semaphore:   mockSemaphore,
	}
	
	// Process batch files
	filenames := []string{
		"The.Matrix.1999.1080p.BluRay.x264.mp4",
		"The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4",
	}
	directoryStructure := map[string][]string{}
	
	results, err := llm.ProcessBatchFiles(context.Background(), filenames, directoryStructure)
	if err != nil {
		t.Fatalf("ProcessBatchFiles failed: %v", err)
	}
	
	// Verify the results
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	
	if results[0].Title != "The Matrix" {
		t.Errorf("Expected first title to be 'The Matrix', got %q", results[0].Title)
	}
	
	if results[1].Title != "The Matrix Reloaded" {
		t.Errorf("Expected second title to be 'The Matrix Reloaded', got %q", results[1].Title)
	}
}

func TestProcessBatchFiles_SemaphoreError(t *testing.T) {
	// Create a mock semaphore that returns an error
	mockSemaphore := &MockSemaphore{
		AcquireFunc: func(ctx context.Context) error {
			return errors.New("semaphore error")
		},
		ReleaseFunc: func() {},
	}
	
	// Create LLM with mocks
	cfg := &config.LLMConfig{
		APIKey:           "test-api-key",
		Model:            "gpt-3.5-turbo",
		BatchSystemPrompt: "You are a helpful assistant that analyzes batches of media filenames.",
		MaxRetries:       0,
	}
	
	llm := &LLM{
		client:      nil, // Not used in this test
		config:      cfg,
		functionMap: make(map[string]FunctionHandler),
		semaphore:   mockSemaphore,
	}
	
	// Process batch files
	filenames := []string{
		"The.Matrix.1999.1080p.BluRay.x264.mp4",
		"The.Matrix.Reloaded.2003.1080p.BluRay.x264.mp4",
	}
	directoryStructure := map[string][]string{}
	
	_, err := llm.ProcessBatchFiles(context.Background(), filenames, directoryStructure)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	
	expectedError := "failed to acquire LLM semaphore: semaphore error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message %q, got %q", expectedError, err.Error())
	}
}

func TestProcessBatchFiles_NoSystemPrompt(t *testing.T) {
	// Create a mock OpenAI client
	mockClient := &MockOpenAIClient{
		CreateChatCompletionFunc: func(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
			// Verify the system prompt was modified correctly
			expectedPrompt := "You are a helpful assistant that analyzes the given filenames."
			if request.Messages[0].Content != expectedPrompt {
				t.Errorf("Expected system prompt %q, got %q", expectedPrompt, request.Messages[0].Content)
			}
			
			// Return a valid response
			return openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: `[{"title":"The Matrix"}]`,
						},
						FinishReason: openai.FinishReasonStop,
					},
				},
			}, nil
		},
	}
	
	// Create a mock semaphore
	mockSemaphore := worker.NewNoOpSemaphore()
	
	// Create LLM with mocks
	cfg := &config.LLMConfig{
		APIKey:       "test-api-key",
		Model:        "gpt-3.5-turbo",
		SystemPrompt: "You are a helpful assistant that analyzes the given filename.",
		MaxRetries:   0,
	}
	
	llm := &LLM{
		client:      mockClient,
		config:      cfg,
		functionMap: make(map[string]FunctionHandler),
		semaphore:   mockSemaphore,
	}
	
	// Process batch files
	filenames := []string{"The.Matrix.1999.1080p.BluRay.x264.mp4"}
	directoryStructure := map[string][]string{}
	
	_, err := llm.ProcessBatchFiles(context.Background(), filenames, directoryStructure)
	if err != nil {
		t.Fatalf("ProcessBatchFiles failed: %v", err)
	}
}
