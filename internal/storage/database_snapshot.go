package storage

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/mychow/ai-git/pkg/types"
	bolt "go.etcd.io/bbolt"
)

func (s *Storage) CreateSnapshot(description string) (*types.Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	symbols := s.cache.GetAll()

	symbolMap := make(map[string]*types.Symbol)
	for _, sym := range symbols {
		symbolMap[sym.ID] = sym
	}

	snapshot := &types.Snapshot{
		ID:        fmt.Sprintf("snap-%d", time.Now().UnixNano()),
		Timestamp: time.Now().Unix(),
		Message:   description,
		Symbols:   symbolMap,
		Metadata: types.SnapshotMetadata{
			Purpose:    description,
			Confidence: 1.0,
			Quality:    1.0,
			CreatedAt:  time.Now(),
		},
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}

	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(snapshotsBucket)
		return b.Put([]byte(snapshot.ID), data)
	}); err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (s *Storage) GetSnapshot(id string) (*types.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var snapshot types.Snapshot
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(snapshotsBucket)
		data := b.Get([]byte(id))
		if data == nil {
			return fmt.Errorf("snapshot not found: %s", id)
		}
		return json.Unmarshal(data, &snapshot)
	})

	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (s *Storage) GetAllSnapshots() ([]*types.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var snapshots []*types.Snapshot
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(snapshotsBucket)
		return b.ForEach(func(k, v []byte) error {
			var snapshot types.Snapshot
			if err := json.Unmarshal(v, &snapshot); err != nil {
				return err
			}
			snapshots = append(snapshots, &snapshot)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp > snapshots[j].Timestamp
	})

	return snapshots, nil
}

func (s *Storage) DeleteSnapshot(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(snapshotsBucket)
		return b.Delete([]byte(id))
	})
}

func (s *Storage) RestoreSnapshot(id string) error {
	snapshot, err := s.GetSnapshot(id)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache.Clear()

	for _, symbol := range snapshot.Symbols {
		s.cache.byID[symbol.ID] = symbol
		s.cache.byName[symbol.Name] = append(s.cache.byName[symbol.Name], symbol)
		s.cache.byFile[symbol.File] = append(s.cache.byFile[symbol.File], symbol)
	}

	return nil
}

func (s *Storage) SaveSnapshot(snapshot *types.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(snapshotsBucket)
		return b.Put([]byte(snapshot.ID), data)
	})
}

func (s *Storage) GetLatestSnapshot() (*types.Snapshot, error) {
	snapshots, err := s.GetAllSnapshots()
	if err != nil {
		return nil, err
	}

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots found")
	}

	return snapshots[0], nil
}

func (s *Storage) SaveSnapshotWithMessage(message string) (*types.Snapshot, error) {
	return s.CreateSnapshot(message)
}

func (s *Storage) RollbackToSnapshot(id string) error {
	return s.RestoreSnapshot(id)
}

func (s *Storage) GetSnapshotHistory() ([]*types.Snapshot, error) {
	return s.GetAllSnapshots()
}

func (s *Storage) DiffSnapshots(id1, id2 string) (map[string]interface{}, error) {
	snap1, err := s.GetSnapshot(id1)
	if err != nil {
		return nil, err
	}

	snap2, err := s.GetSnapshot(id2)
	if err != nil {
		return nil, err
	}

	diff := map[string]interface{}{
		"snapshot1": id1,
		"snapshot2": id2,
		"added":     []string{},
		"removed":   []string{},
		"modified":  []string{},
	}

	for id := range snap2.Symbols {
		if _, exists := snap1.Symbols[id]; !exists {
			diff["added"] = append(diff["added"].([]string), id)
		}
	}

	for id := range snap1.Symbols {
		if _, exists := snap2.Symbols[id]; !exists {
			diff["removed"] = append(diff["removed"].([]string), id)
		}
	}

	return diff, nil
}
