package rag

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/leogoesger/lead-funnel/internal/llm"
)

// BasicEmbedder creates simple hash-based embeddings
// Use this as a template to integrate with Kronk or other embedding services
type BasicEmbedder struct {
	dimension int
}

// NewBasicEmbedder creates a new basic embedder
// Set dimension to match your embedding service (e.g., 1536 for most models)
func NewBasicEmbedder(dimension int) *BasicEmbedder {
	if dimension <= 0 {
		dimension = 1536 // Default dimension
	}
	return &BasicEmbedder{
		dimension: dimension,
	}
}

func (e *BasicEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

func (e *BasicEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i := range texts {
		embedding := make([]float32, e.dimension)
		// Create a simple hash-based embedding for consistency
		// Replace this with your Kronk embedding implementation
		hash := hashString(texts[i])
		for j := range embedding {
			// Use hash to generate pseudo-random but consistent values
			embedding[j] = float32((hash+uint32(j))%100) / 100.0
		}
		embeddings[i] = normalizeVector(embedding)
	}

	return embeddings, nil
}

func (e *BasicEmbedder) Unload(ctx context.Context) error {
	return nil
}

// hashString creates a simple hash of a string
func hashString(s string) uint32 {
	var hash uint32 = 0
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return hash
}

type KronkEmbedder struct {
	client *llm.Client
	model  string
}

func NewKronkEmbedder(ctx context.Context, cfg *Config) (*KronkEmbedder, error) {
	llmClient, err := llm.New(ctx, &llm.Config{Model: cfg.EmbedderModel})
	if err != nil {
		return nil, err
	}

	return &KronkEmbedder{
		client: llmClient,
		model:  cfg.EmbedderModel,
	}, nil
}

func (e *KronkEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

func (e *KronkEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	fmt.Printf("  - EmbedBatch called with %d texts\n", len(texts))
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	d := model.D{
		"model": e.model,
		"input": texts,
	}

	fmt.Printf("  - Calling client.Embeddings...\n")
	resp, err := e.client.Embeddings(ctx, d)
	if err != nil {
		return nil, err
	}
	fmt.Printf("  - Got response with %d embeddings\n", len(resp.Data))
	var embeddings [][]float32
	for i := range resp.Data {
		embeddings = append(embeddings, resp.Data[i].Embedding)
	}

	return embeddings, nil
}

func (e *KronkEmbedder) Unload(ctx context.Context) error {
	return e.client.Unload(ctx)
}
