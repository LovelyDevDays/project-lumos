package hybrid

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// BM42Searcher implements BM42 search using Python scripts
type BM42Searcher struct {
	collection string
	qdrantHost string
	qdrantPort int
	scriptPath string
}

// NewBM42Searcher creates a new BM42 searcher
func NewBM42Searcher(host string, port int, collection string) (*BM42Searcher, error) {
	// Find script path
	scriptPath := filepath.Join("scripts", "bm42_search.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// Try from project root
		scriptPath = filepath.Join("..", "..", "scripts", "bm42_search.py")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("BM42 search script not found")
		}
	}

	return &BM42Searcher{
		collection: collection,
		qdrantHost: host,
		qdrantPort: port,
		scriptPath: scriptPath,
	}, nil
}

// Search performs BM42 search using Python script
func (s *BM42Searcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Create temporary output file
	tmpFile, err := os.CreateTemp("", "bm42_search_*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_ = tmpFile.Close()

	// Run Python script
	cmd := exec.CommandContext(ctx, "python3", s.scriptPath,
		"--query", query,
		"--collection", s.collection,
		"--host", s.qdrantHost,
		"--port", fmt.Sprintf("%d", s.qdrantPort),
		"--limit", fmt.Sprintf("%d", limit),
		"--output", tmpFile.Name(),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run BM42 search: %w\nOutput: %s", err, output)
	}

	// Read results from temp file
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to read results: %w", err)
	}

	// Parse results
	var searchOutput struct {
		Query   string `json:"query"`
		Count   int    `json:"count"`
		Results []struct {
			ID      int     `json:"id"`
			Score   float32 `json:"score"`
			Key     string  `json:"key"`
			Title   string  `json:"title"`
			Content string  `json:"content"`
		} `json:"results"`
	}

	if err := json.Unmarshal(data, &searchOutput); err != nil {
		return nil, fmt.Errorf("failed to parse results: %w", err)
	}

	// Convert to SearchResult
	results := make([]SearchResult, 0, len(searchOutput.Results))
	for _, r := range searchOutput.Results {
		results = append(results, SearchResult{
			Key:     r.Key,
			Title:   r.Title,
			Content: r.Content,
			Score:   r.Score,
			Payload: map[string]interface{}{
				"search_method": "bm42",
			},
		})
	}

	fmt.Printf("[BM42] Found %d results\n", len(results))
	return results, nil
}

// Name returns the name of this searcher
func (s *BM42Searcher) Name() string {
	return "bm42"
}

// IndexDocuments indexes documents using BM42 Python script
func IndexDocuments(ctx context.Context, inputFile string, collection string, host string, port int) error {
	// Find script path
	scriptPath := filepath.Join("scripts", "bm42_indexer.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		scriptPath = filepath.Join("..", "..", "scripts", "bm42_indexer.py")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return fmt.Errorf("BM42 indexer script not found")
		}
	}

	// Run Python script
	cmd := exec.CommandContext(ctx, "python3", scriptPath,
		"--input", inputFile,
		"--collection", collection,
		"--host", host,
		"--port", fmt.Sprintf("%d", port),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run BM42 indexer: %w\nOutput: %s", err, output)
	}

	fmt.Printf("[BM42] Indexing output:\n%s\n", output)
	return nil
}
