package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
	bolt "go.etcd.io/bbolt"
)

func (s *Storage) SaveSymbol(symbol *types.Symbol) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(symbol)
	if err != nil {
		return fmt.Errorf("failed to marshal symbol: %w", err)
	}

	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(symbolsBucket)
		return b.Put([]byte(symbol.ID), data)
	}); err != nil {
		return err
	}

	s.cache.byID[symbol.ID] = symbol
	s.cache.byName[symbol.Name] = append(s.cache.byName[symbol.Name], symbol)
	s.cache.byFile[symbol.File] = append(s.cache.byFile[symbol.File], symbol)

	return nil
}

func (s *Storage) GetSymbol(id string) (*types.Symbol, error) {
	symbol := s.cache.Get(id)
	if symbol == nil {
		return nil, fmt.Errorf("symbol not found: %s", id)
	}
	return symbol, nil
}

func (s *Storage) GetAllSymbols() ([]*types.Symbol, error) {
	return s.cache.GetAll(), nil
}

func (s *Storage) GetSymbolsByName(name string) []*types.Symbol {
	return s.cache.GetByName(name)
}

func (s *Storage) GetSymbolsByFile(file string) []*types.Symbol {
	return s.cache.GetByFile(file)
}

func (s *Storage) SymbolCount() int {
	return s.cache.Count()
}

func (s *Storage) DeleteSymbol(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(symbolsBucket)
		return b.Delete([]byte(id))
	}); err != nil {
		return err
	}

	s.cache.Delete(id)
	return nil
}

func (s *Storage) SaveSymbolsBatch(symbols []*types.Symbol) error {
	return s.SaveSymbolsBatchWithProgress(symbols, nil)
}

func (s *Storage) SaveSymbolsBatchWithProgress(symbols []*types.Symbol, progress func(int, int)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batchSize := 500
	total := len(symbols)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := symbols[i:end]
		err := s.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket(symbolsBucket)
			for _, symbol := range batch {
				data, err := json.Marshal(symbol)
				if err != nil {
					continue
				}
				if err := b.Put([]byte(symbol.ID), data); err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		for _, symbol := range batch {
			s.cache.byID[symbol.ID] = symbol
			s.cache.byName[symbol.Name] = append(s.cache.byName[symbol.Name], symbol)
			s.cache.byFile[symbol.File] = append(s.cache.byFile[symbol.File], symbol)
		}

		if progress != nil {
			progress(end, total)
		}
	}

	return nil
}

func (s *Storage) DeleteSymbolsForFile(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(symbolsBucket)
		prefix := filePath + ":"
		
		var keysToDelete [][]byte
		if err := b.ForEach(func(k, v []byte) error {
			if strings.HasPrefix(string(k), prefix) {
				keysToDelete = append(keysToDelete, k)
			}
			return nil
		}); err != nil {
			return err
		}

		for _, k := range keysToDelete {
			if err := b.Delete(k); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.cache.DeleteByFile(filePath)
	return nil
}
