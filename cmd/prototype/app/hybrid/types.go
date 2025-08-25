package hybrid

import (
	"context"
)

// SearchResult represents a single search result
type SearchResult struct {
	Key     string         `json:"key"`
	Title   string         `json:"title"`
	Content string         `json:"content,omitempty"`
	Score   float32        `json:"score"`
	Payload map[string]any `json:"payload,omitempty"`
}

// Searcher is the base interface for all search implementations
type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
	Name() string
}

// Scorer combines and scores results from multiple sources
type Scorer interface {
	Score(results []SearchResult) []SearchResult
	Merge(results [][]SearchResult, weights []float32) []SearchResult
}
