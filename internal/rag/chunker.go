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
	parts := strings.Split(text, "\n\n")

	var chunks []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			chunks = append(chunks, p)
		}
	}

	return chunks, nil
}
