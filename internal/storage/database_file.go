package storage

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

func (s *Storage) SaveFile(path string, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(filesBucket)
		return b.Put([]byte(path), []byte(content))
	})
}

func (s *Storage) GetFile(path string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var content string

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(filesBucket)
		data := b.Get([]byte(path))
		if data == nil {
			return fmt.Errorf("file not found: %s", path)
		}
		content = string(data)
		return nil
	})

	return content, err
}

func (s *Storage) SaveFilesBatch(files map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(filesBucket)
		for path, content := range files {
			if err := b.Put([]byte(path), []byte(content)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Storage) SaveFileMeta(path string, meta *FileMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileMetaBucket)
		return b.Put([]byte(path), data)
	})
}

func (s *Storage) SaveFileMetasBatch(metas map[string]*FileMeta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileMetaBucket)
		for path, meta := range metas {
			data, err := json.Marshal(meta)
			if err != nil {
				continue
			}
			if err := b.Put([]byte(path), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Storage) GetFileMeta(path string) (*FileMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var meta FileMeta
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileMetaBucket)
		data := b.Get([]byte(path))
		if data == nil {
			return fmt.Errorf("file meta not found: %s", path)
		}
		return json.Unmarshal(data, &meta)
	})

	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *Storage) GetAllFileMetas() (map[string]*FileMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metas := make(map[string]*FileMeta)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileMetaBucket)
		return b.ForEach(func(k, v []byte) error {
			var meta FileMeta
			if err := json.Unmarshal(v, &meta); err != nil {
				return err
			}
			metas[string(k)] = &meta
			return nil
		})
	})

	return metas, err
}

func (s *Storage) DeleteFileMeta(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(fileMetaBucket)
		return b.Delete([]byte(path))
	})
}
