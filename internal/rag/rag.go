package rag

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Config struct {
	EmbedderModel     string  `default:"nomic-embed-text-v1.5.Q8_0" split_words:"true"`
	MinScoreThreshold float32 `default:"0.3" split_words:"true"`
	StaticDocsPath    string  `default:"./static" split_words:"true"`
}

// Document represents a processed document with metadata
type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Source   string            `json:"source"`
	FileType string            `json:"file_type"`
}

// Chunk represents a text chunk with embeddings
type Chunk struct {
	ID        string            `json:"id"`
	Content   string            `json:"content"`
	Embedding []float32         `json:"embedding,omitempty"`
	Metadata  map[string]string `json:"metadata"`
	DocID     string            `json:"doc_id"`
}

// QueryResult represents a search result
type QueryResult struct {
	Chunk     Chunk   `json:"chunk"`
	Score     float32 `json:"score"`
	DocID     string  `json:"doc_id"`
	DocSource string  `json:"doc_source"`
}

// Client is the main interface for the RAG system
type Client interface {
	// Document operations
	ProcessDocument(ctx context.Context, reader io.Reader, filename string, metadata map[string]string) (*Document, error)
	StoreDocument(ctx context.Context, doc *Document) error
	GetDocument(ctx context.Context, docID string) (*Document, error)
	DeleteDocument(ctx context.Context, docID string) error

	// Query operations
	Query(ctx context.Context, query string, topK int) ([]QueryResult, error)
	QueryWithFilter(ctx context.Context, query string, topK int, filter map[string]string) ([]QueryResult, error)

	LoadDocuments(ctx context.Context) error
	Unload(ctx context.Context) error
}

// DocumentLoader defines the interface for loading different document types
type DocumentLoader interface {
	Load(reader io.Reader, filename string) (string, error)
	SupportedExtensions() []string
}

// TextChunker splits text into manageable chunks
type TextChunker interface {
	Chunk(text string) ([]string, error)
}

// VectorStore handles embedding storage and retrieval
type VectorStore interface {
	Store(ctx context.Context, chunks []Chunk) error
	Search(ctx context.Context, embedding []float32, topK int) ([]QueryResult, error)
	SearchWithFilter(ctx context.Context, embedding []float32, topK int, filter map[string]string) ([]QueryResult, error)
	Delete(ctx context.Context, docID string) error
}

// Embedder generates embeddings for text
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	Unload(ctx context.Context) error
}

// ragImpl is the concrete implementation of the RAG interface
type ragImpl struct {
	loaders     map[string]DocumentLoader
	chunker     TextChunker
	vectorStore VectorStore
	embedder    Embedder
	docsPath    string
}

// NewRAG creates a new RAG instance
func NewRAG(vectorStore VectorStore, embedder Embedder, cfg *Config) (Client, error) {
	r := &ragImpl{
		loaders:     make(map[string]DocumentLoader),
		vectorStore: vectorStore,
		embedder:    embedder,
		chunker:     newSemanticChunker(),
		docsPath:    cfg.StaticDocsPath,
	}

	// Register universal loader for all file types
	r.RegisterLoader(NewUniversalLoader())

	return r, nil
}

func (r *ragImpl) LoadDocuments(ctx context.Context) error {
	// 2. Load all .txt files from static folder
	files, err := filepath.Glob(filepath.Join(r.docsPath, "*.txt"))
	if err != nil {
		return fmt.Errorf("failed to list files in %s: %w", r.docsPath, err)
	}

	for _, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", filePath, err)
		}
		doc, err := r.ProcessDocument(ctx, file, filepath.Base(filePath), nil)
		if err != nil {
			file.Close()
			return fmt.Errorf("failed to process document %s: %w", filePath, err)
		}
		if err := r.StoreDocument(ctx, doc); err != nil {
			file.Close()
			return fmt.Errorf("failed to store document %s: %w", filePath, err)
		}
		file.Close()
	}
	return nil
}

// RegisterLoader adds a document loader for specific file types
func (r *ragImpl) RegisterLoader(loader DocumentLoader) {
	for _, ext := range loader.SupportedExtensions() {
		r.loaders[ext] = loader
	}
}

// ProcessDocument processes a document from a reader
func (r *ragImpl) ProcessDocument(ctx context.Context, reader io.Reader, filename string, metadata map[string]string) (*Document, error) {
	ext := getFileExtension(filename)
	loader, ok := r.loaders[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	content, err := loader.Load(reader, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load document: %w", err)
	}

	// Clean the content
	content = cleanText(content)

	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["filename"] = filename
	metadata["extension"] = ext

	doc := &Document{
		ID:       generateID(),
		Content:  content,
		Metadata: metadata,
		Source:   filename,
		FileType: ext,
	}

	return doc, nil
}

// StoreDocument processes, chunks, embeds, and stores a document
func (r *ragImpl) StoreDocument(ctx context.Context, doc *Document) error {
	// Chunk the document
	chunks, err := r.chunker.Chunk(doc.Content)
	if err != nil {
		return fmt.Errorf("failed to chunk document: %w", err)
	}

	// Create chunk objects
	var chunkObjs []Chunk
	for i, chunkText := range chunks {
		chunkObjs = append(chunkObjs, Chunk{
			ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, i),
			Content:  chunkText,
			Metadata: doc.Metadata,
			DocID:    doc.ID,
		})
	}

	// Generate embeddings
	texts := make([]string, len(chunkObjs))
	for i, chunk := range chunkObjs {
		texts[i] = chunk.Content
	}

	embeddings, err := r.embedder.EmbedBatch(ctx, texts)
	if err != nil {
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Attach embeddings to chunks
	for i := range chunkObjs {
		chunkObjs[i].Embedding = embeddings[i]
	}

	// Store in vector store
	if err := r.vectorStore.Store(ctx, chunkObjs); err != nil {
		return fmt.Errorf("failed to store chunks: %w", err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (r *ragImpl) GetDocument(ctx context.Context, docID string) (*Document, error) {
	// This would need to be implemented based on your storage backend
	return nil, fmt.Errorf("not implemented")
}

// DeleteDocument removes a document and its chunks
func (r *ragImpl) DeleteDocument(ctx context.Context, docID string) error {
	return r.vectorStore.Delete(ctx, docID)
}

// Query performs a semantic search
func (r *ragImpl) Query(ctx context.Context, query string, topK int) ([]QueryResult, error) {
	// Generate embedding for the query
	embedding, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search vector store
	results, err := r.vectorStore.Search(ctx, embedding, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return results, nil
}

// QueryWithFilter performs a semantic search with metadata filtering
func (r *ragImpl) QueryWithFilter(ctx context.Context, query string, topK int, filter map[string]string) ([]QueryResult, error) {
	embedding, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	results, err := r.vectorStore.SearchWithFilter(ctx, embedding, topK, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search with filter: %w", err)
	}

	return results, nil
}

func (r *ragImpl) Unload(ctx context.Context) error {
	return r.embedder.Unload(ctx)
}
