package rag

import (
	"context"
	"sync"
)

// InMemoryVectorStore is a simple in-memory implementation for testing/development
type InMemoryVectorStore struct {
	chunks            map[string]Chunk
	mu                sync.RWMutex
	minScoreThreshold float32
}

func NewInMemoryVectorStore(cfg *Config) *InMemoryVectorStore {
	return &InMemoryVectorStore{
		chunks:            make(map[string]Chunk),
		minScoreThreshold: cfg.MinScoreThreshold,
	}
}

func (s *InMemoryVectorStore) Store(ctx context.Context, chunks []Chunk) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, chunk := range chunks {
		s.chunks[chunk.ID] = chunk
	}

	return nil
}

func (s *InMemoryVectorStore) Search(ctx context.Context, embedding []float32, topK int) ([]QueryResult, error) {
	return s.SearchWithFilter(ctx, embedding, topK, nil)
}

func (s *InMemoryVectorStore) SearchWithFilter(ctx context.Context, embedding []float32, topK int, filter map[string]string) ([]QueryResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []QueryResult

	for _, chunk := range s.chunks {
		// Apply filters if provided
		if filter != nil && !matchesFilter(chunk.Metadata, filter) {
			continue
		}

		// Calculate similarity
		score := cosineSimilarity(embedding, chunk.Embedding)

		// Skip results below threshold
		if score < s.minScoreThreshold {
			continue
		}

		results = append(results, QueryResult{
			Chunk:     chunk,
			Score:     score,
			DocID:     chunk.DocID,
			DocSource: chunk.Metadata["filename"],
		})
	}

	// Sort by score (descending)
	sortResultsByScore(results)

	// Return top K
	if len(results) > topK {
		results = results[:topK]
	}

	return results, nil
}

func (s *InMemoryVectorStore) Delete(ctx context.Context, docID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, chunk := range s.chunks {
		if chunk.DocID == docID {
			delete(s.chunks, id)
		}
	}

	return nil
}

// matchesFilter checks if chunk metadata matches all filter criteria
func matchesFilter(metadata, filter map[string]string) bool {
	for key, value := range filter {
		if metadata[key] != value {
			return false
		}
	}
	return true
}

// sortResultsByScore sorts query results by score in descending order
func sortResultsByScore(results []QueryResult) {
	// Simple bubble sort for small result sets
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
