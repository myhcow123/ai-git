package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mychow/ai-git/pkg/types"
	bolt "go.etcd.io/bbolt"
)

type Storage struct {
	db    *bolt.DB
	path  string
	mu    sync.RWMutex
	cache *MemoryCache
}

var (
	symbolsBucket   = []byte("symbols")
	snapshotsBucket = []byte("snapshots")
	filesBucket     = []byte("files")
	metadataBucket  = []byte("metadata")
	fileMetaBucket  = []byte("file_meta")
)

type FileMeta struct {
	ModTime     int64  `json:"mod_time"`
	Size        int64  `json:"size"`
	Hash        string `json:"hash"`
	SymbolCount int    `json:"symbol_count"`
}

func NewStorage(path string) (*Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{symbolsBucket, snapshotsBucket, filesBucket, metadataBucket, fileMetaBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("failed to create bucket: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{
		db:    db,
		path:  path,
		cache: NewMemoryCache(),
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) LoadToCache() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache.Clear()

	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(symbolsBucket)
		return b.ForEach(func(k, v []byte) error {
			var symbol types.Symbol
			if err := json.Unmarshal(v, &symbol); err != nil {
				return err
			}
			s.cache.byID[symbol.ID] = &symbol
			s.cache.byName[symbol.Name] = append(s.cache.byName[symbol.Name], &symbol)
			s.cache.byFile[symbol.File] = append(s.cache.byFile[symbol.File], &symbol)
			return nil
		})
	})
}

func (s *Storage) SetMetadata(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(metadataBucket)
		return b.Put([]byte(key), data)
	})
}

func (s *Storage) GetMetadata(key string, value interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(metadataBucket)
		data := b.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("metadata not found: %s", key)
		}
		return json.Unmarshal(data, value)
	})
}
