package index

import (
	"math"
	"sort"
	"sync"
)

type Vector struct {
	ID   string
	Data []float64
	Norm float64
}

type VectorIndex struct {
	vectors   map[string]*Vector
	dimension int
	mu        sync.RWMutex
}

func NewVectorIndex(dimension int) *VectorIndex {
	return &VectorIndex{
		vectors:   make(map[string]*Vector),
		dimension: dimension,
	}
}

func (idx *VectorIndex) Add(id string, data []float64) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if len(data) != idx.dimension {
		return
	}

	norm := 0.0
	for _, v := range data {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	idx.vectors[id] = &Vector{
		ID:   id,
		Data: data,
		Norm: norm,
	}
}

func (idx *VectorIndex) Remove(id string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.vectors, id)
}

func (idx *VectorIndex) Search(query []float64, k int) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(query) != idx.dimension {
		return []SearchResult{}
	}

	queryNorm := 0.0
	for _, v := range query {
		queryNorm += v * v
	}
	queryNorm = math.Sqrt(queryNorm)

	results := make([]SearchResult, 0, len(idx.vectors))

	for id, vec := range idx.vectors {
		similarity := idx.cosineSimilarity(query, queryNorm, vec)
		results = append(results, SearchResult{
			ID:         id,
			Similarity: similarity,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if k > len(results) {
		k = len(results)
	}

	return results[:k]
}

func (idx *VectorIndex) cosineSimilarity(query []float64, queryNorm float64, vec *Vector) float64 {
	if queryNorm == 0 || vec.Norm == 0 {
		return 0
	}

	dotProduct := 0.0
	for i := range query {
		dotProduct += query[i] * vec.Data[i]
	}

	return dotProduct / (queryNorm * vec.Norm)
}

func (idx *VectorIndex) EuclideanDistance(query []float64, k int) []SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if len(query) != idx.dimension {
		return []SearchResult{}
	}

	results := make([]SearchResult, 0, len(idx.vectors))

	for id, vec := range idx.vectors {
		distance := idx.euclideanDist(query, vec)
		results = append(results, SearchResult{
			ID:         id,
			Similarity: 1.0 / (1.0 + distance),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if k > len(results) {
		k = len(results)
	}

	return results[:k]
}

func (idx *VectorIndex) euclideanDist(query []float64, vec *Vector) float64 {
	sum := 0.0
	for i := range query {
		diff := query[i] - vec.Data[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

func (idx *VectorIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.vectors)
}

type SearchResult struct {
	ID         string
	Similarity float64
}

type EmbeddingModel interface {
	Encode(text string) []float64
	EncodeBatch(texts []string) [][]float64
}

type SimpleEmbeddingModel struct {
	dimension int
}

func NewSimpleEmbeddingModel(dimension int) *SimpleEmbeddingModel {
	return &SimpleEmbeddingModel{
		dimension: dimension,
	}
}

func (m *SimpleEmbeddingModel) Encode(text string) []float64 {
	vector := make([]float64, m.dimension)

	for i, ch := range text {
		if i >= m.dimension {
			break
		}
		vector[i] = float64(ch) / 255.0
	}

	norm := 0.0
	for _, v := range vector {
		norm += v * v
	}
	norm = math.Sqrt(norm)

	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector
}

func (m *SimpleEmbeddingModel) EncodeBatch(texts []string) [][]float64 {
	vectors := make([][]float64, len(texts))
	for i, text := range texts {
		vectors[i] = m.Encode(text)
	}
	return vectors
}
