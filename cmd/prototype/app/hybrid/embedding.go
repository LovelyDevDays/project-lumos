package hybrid

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/qdrant/go-client/qdrant"
)

// EmbeddingSearcher performs vector-based search using embeddings
type EmbeddingSearcher struct {
	client         *qdrant.Client
	collectionName string
	apiAddress     string
	embeddings     map[string][]float32 // Cache for loaded embeddings
}

// NewEmbeddingSearcher creates a new embedding searcher
func NewEmbeddingSearcher(qdrantAddr, collectionName, apiAddr string) (*EmbeddingSearcher, error) {
	// Parse host and port from address like "localhost:6334"
	host := "localhost"
	port := 6334

	if parts := strings.Split(qdrantAddr, ":"); len(parts) == 2 {
		host = parts[0]
		if p, err := strconv.Atoi(parts[1]); err == nil {
			port = p
		}
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	return &EmbeddingSearcher{
		client:         client,
		collectionName: collectionName,
		apiAddress:     apiAddr,
		embeddings:     make(map[string][]float32),
	}, nil
}

// LoadEmbeddingsFromFile loads pre-computed embeddings from a JSON file
func (e *EmbeddingSearcher) LoadEmbeddingsFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read embeddings file: %w", err)
	}

	var embeddings []struct {
		Payload map[string]interface{} `json:"payload"`
		Vectors []float32              `json:"vectors"`
	}

	if err := json.Unmarshal(data, &embeddings); err != nil {
		return fmt.Errorf("failed to unmarshal embeddings: %w", err)
	}

	// Store embeddings in cache
	for _, emb := range embeddings {
		if key, ok := emb.Payload["key"].(string); ok {
			e.embeddings[key] = emb.Vectors
		}
	}

	return nil
}

// Search performs vector similarity search
func (e *EmbeddingSearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Generate embedding for the query
	queryVector, err := e.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform vector search in Qdrant
	searchResult, err := e.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: e.collectionName,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search in Qdrant: %w", err)
	}

	// Convert Qdrant results to SearchResult
	results := []SearchResult{}
	for _, point := range searchResult {
		result := SearchResult{
			Score:   point.Score,
			Payload: make(map[string]interface{}),
		}

		// Extract key and title from payload
		if keyValue := point.Payload["key"]; keyValue != nil {
			result.Key = keyValue.GetStringValue()
			result.Payload["key"] = result.Key
		}
		if titleValue := point.Payload["title"]; titleValue != nil {
			result.Title = titleValue.GetStringValue()
			result.Payload["title"] = result.Title
		}
		if contentValue := point.Payload["content"]; contentValue != nil {
			result.Content = contentValue.GetStringValue()
			result.Payload["content"] = result.Content
		}

		results = append(results, result)
	}

	return results, nil
}

// generateEmbedding generates embedding vector for a text
func (e *EmbeddingSearcher) generateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Reuse the existing embedding generation logic
	client := openai.NewClient(
		option.WithBaseURL(e.apiAddress),
	)

	resp, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(text),
		},
		EncodingFormat: openai.EmbeddingNewParamsEncodingFormatFloat,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Convert float64 to float32
	embedding := resp.Data[0].Embedding
	result := make([]float32, len(embedding))
	for i, v := range embedding {
		result[i] = float32(v)
	}

	return result, nil
}

// Name returns the name of this searcher
func (e *EmbeddingSearcher) Name() string {
	return "embedding"
}
