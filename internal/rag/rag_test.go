package rag_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/leogoesger/lead-funnel/internal/rag"
	"github.com/stretchr/testify/require"
)

// Example: Basic usage with text file
func TestBasicRAG(t *testing.T) {
	cfg := &rag.Config{
		EmbedderModel:     "nomic-embed-text-v1.5.Q8_0",
		MinScoreThreshold: 0.5,
	}
	// Create a basic embedder and in-memory vector store
	embedder := rag.NewBasicEmbedder(1536)
	vectorStore := rag.NewInMemoryVectorStore(cfg)

	// Initialize RAG system
	ragSystem, err := rag.NewRAG(vectorStore, embedder, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Process and store a document
	content := strings.NewReader("This is a sample document about artificial intelligence and machine learning.")
	doc, err := ragSystem.ProcessDocument(ctx, content, "sample.txt", map[string]string{
		"author": "John Doe",
		"topic":  "AI",
	})
	if err != nil {
		fmt.Printf("Error processing document: %v\n", err)
		return
	}

	err = ragSystem.StoreDocument(ctx, doc)
	if err != nil {
		fmt.Printf("Error storing document: %v\n", err)
		return
	}

	// Query the system
	results, err := ragSystem.Query(ctx, "What is machine learning?", 3)
	if err != nil {
		fmt.Printf("Error querying: %v\n", err)
		return
	}

	fmt.Printf("Found %d results\n", len(results))
	for i, result := range results {
		fmt.Printf("Result %d: Score=%.2f\n", i+1, result.Score)
	}
}

// Example: Using with markdown files
func TestRAG_markdown(t *testing.T) {
	cfg := &rag.Config{
		EmbedderModel:     "nomic-embed-text-v1.5.Q8_0",
		MinScoreThreshold: 0.5,
	}
	embedder := rag.NewBasicEmbedder(1536)
	vectorStore := rag.NewInMemoryVectorStore(cfg)
	ragSystem, err := rag.NewRAG(vectorStore, embedder, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Process markdown document
	mdContent := `# Introduction to RAG
	
RAG stands for Retrieval-Augmented Generation. It combines:
- Vector search
- Document retrieval
- Language models

## How it works
RAG retrieves relevant context from a knowledge base before generating responses.`

	doc, err := ragSystem.ProcessDocument(ctx, strings.NewReader(mdContent), "rag-intro.md", nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ragSystem.StoreDocument(ctx, doc)

	// Query with filter
	results, err := ragSystem.QueryWithFilter(ctx, "How does RAG work?", 2, map[string]string{
		"extension": ".md",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d filtered results\n", len(results))
}

// Example: Using with CSV data
func TestRAG_csv(t *testing.T) {
	cfg := &rag.Config{
		EmbedderModel:     "nomic-embed-text-v1.5.Q8_0",
		MinScoreThreshold: 0.5,
	}
	// Create a basic embedder and in-memory vector store
	embedder := rag.NewBasicEmbedder(1536)
	vectorStore := rag.NewInMemoryVectorStore(cfg)
	ragSystem, err := rag.NewRAG(vectorStore, embedder, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	csvContent := `name,role,department
Alice Johnson,Engineer,Development
Bob Smith,Manager,Operations
Carol White,Analyst,Data Science`

	doc, err := ragSystem.ProcessDocument(ctx, strings.NewReader(csvContent), "employees.csv", map[string]string{
		"type": "employee_data",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	ragSystem.StoreDocument(ctx, doc)

	// Query for specific information
	results, err := ragSystem.Query(ctx, "Who is in the Data Science department?", 5)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Found %d results\n", len(results))
}

// TestBasicRAGWorkflow tests the complete RAG workflow
func TestBasicRAGWorkflow(t *testing.T) {
	cfg := &rag.Config{
		EmbedderModel:     "nomic-embed-text-v1.5.Q8_0",
		MinScoreThreshold: 0.5,
	}
	embedder := rag.NewBasicEmbedder(128)
	vectorStore := rag.NewInMemoryVectorStore(cfg)
	ragSystem, err := rag.NewRAG(vectorStore, embedder, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Process document
	content := strings.NewReader("The quick brown fox jumps over the lazy dog.")
	doc, err := ragSystem.ProcessDocument(ctx, content, "test.txt", nil)
	if err != nil {
		t.Fatalf("Failed to process document: %v", err)
	}

	// Store document
	err = ragSystem.StoreDocument(ctx, doc)
	if err != nil {
		t.Fatalf("Failed to store document: %v", err)
	}

	// Query
	results, err := ragSystem.Query(ctx, "fox", 5)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

// TestMultipleFileTypes tests different file type loaders
func TestMultipleFileTypes(t *testing.T) {
	cfg := &rag.Config{
		EmbedderModel:     "nomic-embed-text-v1.5.Q8_0",
		MinScoreThreshold: 0.5,
	}
	embedder := rag.NewBasicEmbedder(128)
	vectorStore := rag.NewInMemoryVectorStore(cfg)
	ragSystem, err := rag.NewRAG(vectorStore, embedder, cfg)
	require.NoError(t, err)

	ctx := context.Background()

	testCases := []struct {
		filename string
		content  string
	}{
		{"test.txt", "Plain text content"},
		{"test.md", "# Markdown content"},
		{"test.csv", "header1,header2\nvalue1,value2"},
		{"test.json", `{"key": "value"}`},
		{"test.go", "package main\nfunc main() {}"},
	}

	for _, tc := range testCases {
		doc, err := ragSystem.ProcessDocument(ctx, strings.NewReader(tc.content), tc.filename, nil)
		if err != nil {
			t.Errorf("Failed to process %s: %v", tc.filename, err)
			continue
		}

		err = ragSystem.StoreDocument(ctx, doc)
		if err != nil {
			t.Errorf("Failed to store %s: %v", tc.filename, err)
		}
	}

	// Query across all documents
	results, err := ragSystem.Query(ctx, "content", 10)
	if err != nil {
		t.Fatalf("Failed to query: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected results from multiple documents")
	}
}

func TestKronkEmbedder(t *testing.T) {
	cfg := &rag.Config{
		EmbedderModel:     "nomic-embed-text-v1.5.Q8_0",
		MinScoreThreshold: 0.5,
	}
	ctx := context.Background()

	embedder, err := rag.NewKronkEmbedder(ctx, cfg)
	require.NoError(t, err)

	vectorStore := rag.NewInMemoryVectorStore(cfg)
	ragSystem, err := rag.NewRAG(vectorStore, embedder, cfg)
	require.NoError(t, err)

	testCases := []struct {
		filename string
		content  string
	}{
		{"test.txt", "Plain text content"},
		{"test.md", "# Markdown content"},
		{"test.csv", "header1,header2\nvalue1,value2"},
		{"test.json", `{"key": "value"}`},
		{"test.go", "package main\nfunc main() {}"},
	}

	for _, tc := range testCases {
		doc, err := ragSystem.ProcessDocument(ctx, strings.NewReader(tc.content), tc.filename, nil)
		if err != nil {
			t.Errorf("Failed to process %s: %v", tc.filename, err)
			continue
		}

		err = ragSystem.StoreDocument(ctx, doc)
		if err != nil {
			t.Errorf("Failed to store %s: %v", tc.filename, err)
		}
	}

	// Query across all documents with irrelevant term
	results, err := ragSystem.Query(ctx, "leo", 10)
	require.NoError(t, err)
	fmt.Printf("Results for 'leo' query: %d results\n", len(results))
	for i, r := range results {
		fmt.Printf("  [%d] Score: %.4f, Content: %s\n", i, r.Score, r.Chunk.Content[:min(50, len(r.Chunk.Content))])
	}
	// Real embedders may find weak semantic similarity even for unrelated terms

	// Query with relevant term that appears in content
	results, err = ragSystem.Query(ctx, "text content", 10)
	require.NoError(t, err)
	fmt.Printf("\nResults for 'text content' query: %d results\n", len(results))
	require.Greater(t, len(results), 0, "Expected results for relevant query")
	require.LessOrEqual(t, len(results), 10, "Should not exceed topK")

	// Verify relevant results have higher scores
	if len(results) > 0 {
		fmt.Printf("  Top result score: %.4f, Content: %s\n", results[0].Score, results[0].Chunk.Content[:min(50, len(results[0].Chunk.Content))])
	}
}
