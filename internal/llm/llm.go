package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/sleepstars/mediascanner/internal/config"
)

// LLM represents the LLM client
type LLM struct {
	client      *openai.Client
	config      *config.LLMConfig
	functionMap map[string]FunctionHandler
}

// FunctionHandler is a function that handles a function call from the LLM
type FunctionHandler func(ctx context.Context, args json.RawMessage) (interface{}, error)

// New creates a new LLM client
func New(cfg *config.LLMConfig) (*LLM, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("LLM API key is required")
	}

	clientConfig := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		clientConfig.BaseURL = cfg.BaseURL
	}

	client := openai.NewClientWithConfig(clientConfig)

	return &LLM{
		client:      client,
		config:      cfg,
		functionMap: make(map[string]FunctionHandler),
	}, nil
}

// RegisterFunction registers a function handler
func (l *LLM) RegisterFunction(name string, handler FunctionHandler) {
	l.functionMap[name] = handler
}

// ProcessMediaFile processes a media file using the LLM
func (l *LLM) ProcessMediaFile(ctx context.Context, filename string, directoryStructure map[string][]string) (*MediaFileResult, error) {
	// Use the system prompt from configuration
	systemMessage := l.config.SystemPrompt

	// Create the user message with the filename
	userMessage := fmt.Sprintf("Please analyze this filename: %s", filename)

	// Define the functions that can be called by the LLM
	functions := []openai.FunctionDefinition{
		{
			Name:        "searchTMDB",
			Description: "Search for a movie or TV show on TMDB",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
					"year": map[string]interface{}{
						"type":        "integer",
						"description": "The year of release (optional)",
					},
					"mediaType": map[string]interface{}{
						"type":        "string",
						"description": "The type of media to search for (movie, tv)",
						"enum":        []string{"movie", "tv"},
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "searchTVDB",
			Description: "Search for a TV show on TVDB",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "searchBangumi",
			Description: "Search for anime on Bangumi",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	// Create the chat completion request
	request := openai.ChatCompletionRequest{
		Model: l.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
		Functions:    functions,
		FunctionCall: "auto",
	}

	// Process the request with retries
	var response openai.ChatCompletionResponse
	var err error
	for i := 0; i <= l.config.MaxRetries; i++ {
		response, err = l.client.CreateChatCompletion(ctx, request)
		if err == nil {
			break
		}

		// If we've reached the maximum number of retries, return the error
		if i == l.config.MaxRetries {
			return nil, fmt.Errorf("failed to process media file after %d retries: %w", l.config.MaxRetries, err)
		}

		// Wait before retrying
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	// Process function calls
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemMessage,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}

	// Handle function calls
	for {
		// Check if the response contains a function call
		if response.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
			functionCall := response.Choices[0].Message.FunctionCall

			// Add the assistant's message to the conversation
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: response.Choices[0].Message.Content,
				// FunctionCall field has been updated in newer versions of the API
				// We need to handle this differently
			})

			// Execute the function
			handler, ok := l.functionMap[functionCall.Name]
			if !ok {
				return nil, fmt.Errorf("unknown function: %s", functionCall.Name)
			}

			result, err := handler(ctx, json.RawMessage(functionCall.Arguments))
			if err != nil {
				return nil, fmt.Errorf("error executing function %s: %w", functionCall.Name, err)
			}

			// Convert the result to JSON
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("error marshaling function result: %w", err)
			}

			// Add the function response to the conversation
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleFunction,
				Name:    functionCall.Name,
				Content: string(resultJSON),
			})

			// Create a new chat completion request with the updated conversation
			request = openai.ChatCompletionRequest{
				Model:        l.config.Model,
				Messages:     messages,
				Functions:    functions,
				FunctionCall: "auto",
			}

			// Send the request
			response, err = l.client.CreateChatCompletion(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("error creating chat completion: %w", err)
			}

			// Continue the loop to handle additional function calls
			continue
		}

		// If we've reached here, the LLM has provided a final response
		break
	}

	// Parse the final response
	var result MediaFileResult
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing LLM response: %w", err)
	}

	return &result, nil
}

// ProcessBatchFiles processes a batch of media files using the LLM
func (l *LLM) ProcessBatchFiles(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*MediaFileResult, error) {
	// Use the batch system prompt from configuration
	systemMessage := l.config.BatchSystemPrompt

	// If batch system prompt is not set, fall back to modifying the single file system prompt
	if systemMessage == "" {
		systemMessage = strings.Replace(l.config.SystemPrompt, "the given filename", "the given filenames", -1)
		systemMessage = strings.Replace(systemMessage, "Respond with a structured JSON", "Respond with a structured JSON array", -1)
	}

	// Create the user message with the filenames
	userMessage := "Please analyze these filenames:\n"
	for i, filename := range filenames {
		userMessage += fmt.Sprintf("%d. %s\n", i+1, filename)
	}

	// Define the functions that can be called by the LLM
	functions := []openai.FunctionDefinition{
		{
			Name:        "searchTMDB",
			Description: "Search for a movie or TV show on TMDB",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
					"year": map[string]interface{}{
						"type":        "integer",
						"description": "The year of release (optional)",
					},
					"mediaType": map[string]interface{}{
						"type":        "string",
						"description": "The type of media to search for (movie, tv)",
						"enum":        []string{"movie", "tv"},
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "searchTVDB",
			Description: "Search for a TV show on TVDB",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "searchBangumi",
			Description: "Search for anime on Bangumi",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}

	// Create the chat completion request
	request := openai.ChatCompletionRequest{
		Model: l.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
		Functions:    functions,
		FunctionCall: "auto",
	}

	// Process the request with retries
	var response openai.ChatCompletionResponse
	var err error
	for i := 0; i <= l.config.MaxRetries; i++ {
		response, err = l.client.CreateChatCompletion(ctx, request)
		if err == nil {
			break
		}

		// If we've reached the maximum number of retries, return the error
		if i == l.config.MaxRetries {
			return nil, fmt.Errorf("failed to process batch files after %d retries: %w", l.config.MaxRetries, err)
		}

		// Wait before retrying
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	// Process function calls
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemMessage,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}

	// Handle function calls
	for {
		// Check if the response contains a function call
		if response.Choices[0].FinishReason == openai.FinishReasonFunctionCall {
			functionCall := response.Choices[0].Message.FunctionCall

			// Add the assistant's message to the conversation
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: response.Choices[0].Message.Content,
				// FunctionCall field has been updated in newer versions of the API
				// We need to handle this differently
			})

			// Execute the function
			handler, ok := l.functionMap[functionCall.Name]
			if !ok {
				return nil, fmt.Errorf("unknown function: %s", functionCall.Name)
			}

			result, err := handler(ctx, json.RawMessage(functionCall.Arguments))
			if err != nil {
				return nil, fmt.Errorf("error executing function %s: %w", functionCall.Name, err)
			}

			// Convert the result to JSON
			resultJSON, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("error marshaling function result: %w", err)
			}

			// Add the function response to the conversation
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleFunction,
				Name:    functionCall.Name,
				Content: string(resultJSON),
			})

			// Create a new chat completion request with the updated conversation
			request = openai.ChatCompletionRequest{
				Model:        l.config.Model,
				Messages:     messages,
				Functions:    functions,
				FunctionCall: "auto",
			}

			// Send the request
			response, err = l.client.CreateChatCompletion(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("error creating chat completion: %w", err)
			}

			// Continue the loop to handle additional function calls
			continue
		}

		// If we've reached here, the LLM has provided a final response
		break
	}

	// Parse the final response
	var results []*MediaFileResult
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &results)
	if err != nil {
		return nil, fmt.Errorf("error parsing LLM response: %w", err)
	}

	return results, nil
}

// MediaFileResult represents the result of processing a media file
type MediaFileResult struct {
	OriginalFilename string  `json:"original_filename"`
	Title            string  `json:"title"`
	OriginalTitle    string  `json:"original_title,omitempty"`
	Year             int     `json:"year,omitempty"`
	MediaType        string  `json:"media_type"` // movie, tv
	Season           int     `json:"season,omitempty"`
	Episode          int     `json:"episode,omitempty"`
	EpisodeTitle     string  `json:"episode_title,omitempty"`
	TMDBID           int64   `json:"tmdb_id,omitempty"`
	TVDBID           int64   `json:"tvdb_id,omitempty"`
	BangumiID        int64   `json:"bangumi_id,omitempty"`
	ImdbID           string  `json:"imdb_id,omitempty"`
	Category         string  `json:"category"`
	Subcategory      string  `json:"subcategory"`
	DestinationPath  string  `json:"destination_path"`
	Confidence       float64 `json:"confidence,omitempty"`
}
