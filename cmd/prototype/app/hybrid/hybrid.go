package hybrid

import (
	"context"
	"fmt"
	"sort"
)

// HybridSearcher combines multiple search strategies
type HybridSearcher struct {
	searchers             map[string]Searcher
	weights               map[string]float32
	config                *SearchConfig
	overlapBonus          float32 // Bonus multiplier for documents appearing in multiple searches
	enableOverlapPriority bool    // Enable overlap prioritization
}

// NewHybridSearcher creates a new hybrid searcher
func NewHybridSearcher(config *SearchConfig) *HybridSearcher {
	return &HybridSearcher{
		searchers:             make(map[string]Searcher),
		weights:               make(map[string]float32),
		config:                config,
		overlapBonus:          2.0,  // Default: 2x bonus for overlapping documents
		enableOverlapPriority: true, // Default: enable overlap priority
	}
}

// AddSearcher adds a searcher with a specific weight
func (h *HybridSearcher) AddSearcher(name string, searcher Searcher, weight float32) {
	h.searchers[name] = searcher
	h.weights[name] = weight
}

// SetOverlapBonus sets the bonus multiplier for overlapping documents
func (h *HybridSearcher) SetOverlapBonus(bonus float32) {
	h.overlapBonus = bonus
}

// SetOverlapPriority enables or disables overlap prioritization
func (h *HybridSearcher) SetOverlapPriority(enable bool) {
	h.enableOverlapPriority = enable
}

// Search performs hybrid search combining all registered searchers
func (h *HybridSearcher) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if len(h.searchers) == 0 {
		return nil, fmt.Errorf("no searchers configured")
	}

	// Collect results from all searchers
	allResults := make(map[string][]SearchResult)

	for name, searcher := range h.searchers {
		// Search with original query
		results, err := searcher.Search(ctx, query, limit)
		if err != nil {
			continue
		}
		allResults[name] = results
	}

	// Merge and score results
	mergedResults := h.mergeResults(allResults)

	// Apply minimum score filter if configured
	if h.config != nil && h.config.MinScore > 0 {
		filtered := []SearchResult{}
		for _, result := range mergedResults {
			if result.Score >= h.config.MinScore {
				filtered = append(filtered, result)
			}
		}
		mergedResults = filtered
	}

	// Limit final results
	if limit > 0 && len(mergedResults) > limit {
		mergedResults = mergedResults[:limit]
	}

	return mergedResults, nil
}

// mergeResults merges results from multiple searchers with weighted scoring and overlap bonus
func (h *HybridSearcher) mergeResults(allResults map[string][]SearchResult) []SearchResult {
	// Map to store combined scores for each unique result
	scoreMap := make(map[string]*SearchResult)
	contributionMap := make(map[string]map[string]float32) // Track contributions from each searcher
	sourceCountMap := make(map[string]int)                 // Track how many searchers returned each document

	for searcherName, results := range allResults {
		weight := h.weights[searcherName]

		// Normalize scores within each searcher's results
		normalizedResults := h.normalizeScores(results)

		for _, result := range normalizedResults {
			// Track source count
			sourceCountMap[result.Key]++

			if existing, exists := scoreMap[result.Key]; exists {
				// Update weighted score
				existing.Score += result.Score * weight

				// Track contribution
				if contributionMap[result.Key] == nil {
					contributionMap[result.Key] = make(map[string]float32)
				}
				contributionMap[result.Key][searcherName] = result.Score * weight
			} else {
				// Create new result entry
				newResult := result
				newResult.Score = result.Score * weight
				scoreMap[result.Key] = &newResult

				// Initialize contribution tracking
				contributionMap[result.Key] = map[string]float32{
					searcherName: result.Score * weight,
				}
			}
		}
	}

	// Convert map to slice with overlap bonus
	finalResults := []SearchResult{}
	for key, result := range scoreMap {
		// Apply overlap bonus if document appears in multiple searchers
		if h.enableOverlapPriority && sourceCountMap[key] > 1 {
			result.Score *= h.overlapBonus

			// Add overlap metadata
			if result.Payload == nil {
				result.Payload = make(map[string]interface{})
			}
			result.Payload["overlap_count"] = sourceCountMap[key]
			result.Payload["overlap_bonus_applied"] = true
		}

		// Add contribution metadata
		if result.Payload == nil {
			result.Payload = make(map[string]interface{})
		}
		result.Payload["score_contributions"] = contributionMap[key]
		result.Payload["source_count"] = sourceCountMap[key]

		finalResults = append(finalResults, *result)
	}

	// Sort with overlap priority
	sort.Slice(finalResults, func(i, j int) bool {
		if h.enableOverlapPriority {
			// First priority: documents with overlap
			iOverlap := finalResults[i].Payload["source_count"].(int) > 1
			jOverlap := finalResults[j].Payload["source_count"].(int) > 1

			if iOverlap != jOverlap {
				return iOverlap // Overlap documents come first
			}
		}
		// Second priority: score
		return finalResults[i].Score > finalResults[j].Score
	})

	return finalResults
}

// normalizeScores normalizes scores to [0, 1] range
func (h *HybridSearcher) normalizeScores(results []SearchResult) []SearchResult {
	if len(results) == 0 {
		return results
	}

	// Find max score
	var maxScore float32
	for _, result := range results {
		if result.Score > maxScore {
			maxScore = result.Score
		}
	}

	// Avoid division by zero
	if maxScore == 0 {
		return results
	}

	// Normalize scores
	normalized := make([]SearchResult, len(results))
	for i, result := range results {
		normalized[i] = result
		normalized[i].Score = result.Score / maxScore
	}

	return normalized
}

// Name returns the name of this searcher
func (h *HybridSearcher) Name() string {
	return "hybrid"
}
