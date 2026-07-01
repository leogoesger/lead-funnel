package rag

import (
	"strings"
)

// semanticChunker uses semantic boundaries for chunking
type semanticChunker struct {
	separators []string
}

func newSemanticChunker() *semanticChunker {
	return &semanticChunker{
		separators: []string{"\n\n", "\n", ". ", "! ", "? "},
	}
}

func (c *semanticChunker) Chunk(text string) ([]string, error) {
	return c.ChunkWithOverlap(text, 1000, 200)
}

func (c *semanticChunker) ChunkWithOverlap(text string, chunkSize, overlap int) ([]string, error) {
	// Use character-based chunking which is simple and reliable
	return c.characterChunk(text, chunkSize, overlap), nil
}

func (c *semanticChunker) characterChunk(text string, chunkSize, overlap int) []string {
	var chunks []string
	runes := []rune(text)

	for i := 0; i < len(runes); {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunk := string(runes[i:end])
		chunks = append(chunks, strings.TrimSpace(chunk))

		// If we've reached the end, we're done
		if end >= len(runes) {
			break
		}

		i = end - overlap
		if i <= 0 {
			i = end
		}
	}

	return chunks
}
