package rag

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strings"
	"time"
)

// UniversalLoader handles all file types by reading them as text
type UniversalLoader struct{}

func NewUniversalLoader() *UniversalLoader {
	return &UniversalLoader{}
}

func (l *UniversalLoader) Load(reader io.Reader, filename string) (string, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(data), nil
}

func (l *UniversalLoader) SupportedExtensions() []string {
	// Support all common file extensions
	return []string{
		// Text files
		".txt", ".log", ".md", ".markdown",
		// Data files
		".csv", ".json", ".xml", ".yaml", ".yml", ".toml",
		// Web files
		".html", ".htm",
		// Code files
		".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h",
		".rs", ".rb", ".php", ".swift", ".kt", ".scala", ".sh",
		".jsx", ".tsx", ".vue", ".css", ".scss", ".sass",
		// Other
		".sql", ".env", ".ini", ".cfg", ".conf",
	}
}

// generateID creates a unique identifier
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// getFileExtension extracts the file extension from a filename
func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.ToLower(ext)
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt32(normA) * sqrt32(normB))
}

// sqrt32 calculates square root of float32
func sqrt32(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(v []float32) []float32 {
	var sum float32
	for _, val := range v {
		sum += val * val
	}

	if sum == 0 {
		return v
	}

	norm := sqrt32(sum)
	result := make([]float32, len(v))
	for i, val := range v {
		result[i] = val / norm
	}

	return result
}

// cleanText performs basic text cleaning
func cleanText(text string) string {
	// Remove excessive whitespace
	text = strings.TrimSpace(text)

	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")

	// Remove null bytes
	text = strings.ReplaceAll(text, "\x00", "")

	return text
}
