package llm

import (
	"context"
)

// LLMInterface defines the interface for LLM operations
type LLMInterface interface {
	ProcessMediaFile(ctx context.Context, filename string, directoryStructure map[string][]string) (*MediaFileResult, error)
	ProcessBatchFiles(ctx context.Context, filenames []string, directoryStructure map[string][]string) ([]*MediaFileResult, error)
	RegisterFunction(name string, handler FunctionHandler)
}
