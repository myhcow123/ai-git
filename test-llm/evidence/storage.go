package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// TaskStorage handles persistence of tasks to JSON file
type TaskStorage struct {
	filePath string
	mu       sync.RWMutex
	tasks    map[string]*Task
}

// NewTaskStorage creates a new TaskStorage instance
func NewTaskStorage(filePath string) (*TaskStorage, error) {
	storage := &TaskStorage{
		filePath: filePath,
		tasks:    make(map[string]*Task),
	}

	// Load existing tasks from file
	if err := storage.load(); err != nil {
		// If file doesn't exist, it's not an error - just start fresh
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load tasks: %w", err)
		}
	}

	return storage, nil
}

// load reads tasks from the JSON file
func (s *TaskStorage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	s.tasks = make(map[string]*Task)
	for _, task := range tasks {
		s.tasks[task.ID] = task
	}

	return nil
}

// save writes tasks to the JSON file
func (s *TaskStorage) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// Save stores a task and persists to disk
func (s *TaskStorage) Save(task *Task) error {
	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	return s.save()
}

// Get retrieves a task by ID
func (s *TaskStorage) Get(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	return task, exists
}

// GetAll returns all tasks
func (s *TaskStorage) GetAll() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// Delete removes a task by ID
func (s *TaskStorage) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return false
	}

	delete(s.tasks, id)
	return true
}

// Filter returns tasks matching the given status filter
func (s *TaskStorage) Filter(status Status) []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*Task
	for _, task := range s.tasks {
		if status == "" || task.Status == status {
			filtered = append(filtered, task)
		}
	}

	return filtered
}

// Reload re-reads the tasks from disk
func (s *TaskStorage) Reload() error {
	return s.load()
}
