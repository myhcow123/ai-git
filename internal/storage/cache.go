package storage

import (
	"sync"

	"github.com/mychow/ai-git/pkg/types"
)

type MemoryCache struct {
	mu sync.RWMutex

	byID   map[string]*types.Symbol
	byName map[string][]*types.Symbol
	byFile map[string][]*types.Symbol
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		byID:   make(map[string]*types.Symbol),
		byName: make(map[string][]*types.Symbol),
		byFile: make(map[string][]*types.Symbol),
	}
}

func (c *MemoryCache) Add(symbol *types.Symbol) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byID[symbol.ID] = symbol
	c.byName[symbol.Name] = append(c.byName[symbol.Name], symbol)
	c.byFile[symbol.File] = append(c.byFile[symbol.File], symbol)
}

func (c *MemoryCache) AddBatch(symbols []*types.Symbol) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, symbol := range symbols {
		c.byID[symbol.ID] = symbol
		c.byName[symbol.Name] = append(c.byName[symbol.Name], symbol)
		c.byFile[symbol.File] = append(c.byFile[symbol.File], symbol)
	}
}

func (c *MemoryCache) Get(id string) *types.Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.byID[id]
}

func (c *MemoryCache) GetByName(name string) []*types.Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.byName[name]
}

func (c *MemoryCache) GetByFile(file string) []*types.Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.byFile[file]
}

func (c *MemoryCache) GetAll() []*types.Symbol {
	c.mu.RLock()
	defer c.mu.RUnlock()

	symbols := make([]*types.Symbol, 0, len(c.byID))
	for _, s := range c.byID {
		symbols = append(symbols, s)
	}
	return symbols
}

func (c *MemoryCache) Update(symbol *types.Symbol) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if old, exists := c.byID[symbol.ID]; exists {
		c.removeFromSlice(c.byName[old.Name], old.ID)
		c.removeFromSlice(c.byFile[old.File], old.ID)
	}

	c.byID[symbol.ID] = symbol
	c.byName[symbol.Name] = append(c.byName[symbol.Name], symbol)
	c.byFile[symbol.File] = append(c.byFile[symbol.File], symbol)
}

func (c *MemoryCache) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if symbol, exists := c.byID[id]; exists {
		delete(c.byID, id)
		c.removeFromSlice(c.byName[symbol.Name], id)
		c.removeFromSlice(c.byFile[symbol.File], id)
	}
}

func (c *MemoryCache) DeleteByFile(file string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	symbols := c.byFile[file]
	for _, s := range symbols {
		delete(c.byID, s.ID)
		c.removeFromSlice(c.byName[s.Name], s.ID)
	}
	delete(c.byFile, file)
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.byID = make(map[string]*types.Symbol)
	c.byName = make(map[string][]*types.Symbol)
	c.byFile = make(map[string][]*types.Symbol)
}

func (c *MemoryCache) removeFromSlice(slice []*types.Symbol, id string) {
	for i, s := range slice {
		if s.ID == id {
			slice[i] = slice[len(slice)-1]
			slice = slice[:len(slice)-1]
			break
		}
	}
}

func (c *MemoryCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.byID)
}
