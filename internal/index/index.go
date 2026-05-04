package index

import (
	"sync"

	"github.com/mychow/ai-git/pkg/types"
	"strings"
)

type Index struct {
	mu sync.RWMutex

	NameIndex      map[string][]string
	TypeIndex      map[types.SymbolType][]string
	FileIndex      map[string][]string
	SignatureIndex map[string]string

	Dependencies map[string][]string
	Dependents   map[string][]string
}

func NewIndex() *Index {
	return &Index{
		NameIndex:      make(map[string][]string),
		TypeIndex:      make(map[types.SymbolType][]string),
		FileIndex:      make(map[string][]string),
		SignatureIndex: make(map[string]string),
		Dependencies:   make(map[string][]string),
		Dependents:     make(map[string][]string),
	}
}

func (idx *Index) AddSymbol(symbol *types.Symbol) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	id := symbol.ID

	idx.NameIndex[symbol.Name] = append(idx.NameIndex[symbol.Name], id)
	idx.TypeIndex[symbol.Type] = append(idx.TypeIndex[symbol.Type], id)
	idx.FileIndex[symbol.File] = append(idx.FileIndex[symbol.File], id)

	if symbol.Signature != "" {
		idx.SignatureIndex[symbol.Signature] = id
	}
}

func (idx *Index) RemoveSymbol(symbol *types.Symbol) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	id := symbol.ID

	idx.NameIndex[symbol.Name] = removeFromSlice(idx.NameIndex[symbol.Name], id)
	idx.TypeIndex[symbol.Type] = removeFromSlice(idx.TypeIndex[symbol.Type], id)
	idx.FileIndex[symbol.File] = removeFromSlice(idx.FileIndex[symbol.File], id)

	if symbol.Signature != "" {
		delete(idx.SignatureIndex, symbol.Signature)
	}
}

func (idx *Index) GetByName(name string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	ids := idx.NameIndex[name]
	result := make([]string, len(ids))
	copy(result, ids)
	return result
}

func (idx *Index) GetByType(symbolType types.SymbolType) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	ids := idx.TypeIndex[symbolType]
	result := make([]string, len(ids))
	copy(result, ids)
	return result
}

func (idx *Index) GetByFile(file string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	ids := idx.FileIndex[file]
	result := make([]string, len(ids))
	copy(result, ids)
	return result
}

func (idx *Index) GetBySignature(signature string) (string, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	id, exists := idx.SignatureIndex[signature]
	return id, exists
}

func (idx *Index) AddDependency(from, to string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.Dependencies[from] = append(idx.Dependencies[from], to)
	idx.Dependents[to] = append(idx.Dependents[to], from)
}

func (idx *Index) GetDependencies(id string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	deps := idx.Dependencies[id]
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

func (idx *Index) GetDependents(id string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	deps := idx.Dependents[id]
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

func (idx *Index) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.NameIndex = make(map[string][]string)
	idx.TypeIndex = make(map[types.SymbolType][]string)
	idx.FileIndex = make(map[string][]string)
	idx.SignatureIndex = make(map[string]string)
	idx.Dependencies = make(map[string][]string)
	idx.Dependents = make(map[string][]string)
}

func (idx *Index) Stats() map[string]interface{} {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return map[string]interface{}{
		"name_count":       len(idx.NameIndex),
		"type_count":       len(idx.TypeIndex),
		"file_count":       len(idx.FileIndex),
		"signature_count":  len(idx.SignatureIndex),
		"dependency_count": len(idx.Dependencies),
	}
}

func (idx *Index) GetSymbolsByIDs(ids []string) ([]*types.Symbol, error) {
	return []*types.Symbol{}, nil
}

func (idx *Index) SearchByDescription(description string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	results := []string{}
	for name := range idx.NameIndex {
		if containsMatch(name, description) {
			results = append(results, idx.NameIndex[name]...)
		}
	}
	return results
}

func (idx *Index) SearchByPattern(pattern string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	results := []string{}
	for name := range idx.NameIndex {
		if matchPattern(name, pattern) {
			results = append(results, idx.NameIndex[name]...)
		}
	}
	return results
}

func containsMatch(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func matchPattern(s, pattern string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(pattern))
}

func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
