package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrTaskNotFound is returned when a task is not found
	ErrTaskNotFound = errors.New("task not found")
	// ErrInvalidID is returned when an invalid task ID is provided
	ErrInvalidID = errors.New("invalid task ID")
	// ErrEmptyTitle is returned when the task title is empty
	ErrEmptyTitle = errors.New("task title cannot be empty")
	// ErrInvalidStatus is returned when an invalid status is provided
	ErrInvalidStatus = errors.New("invalid status")
	// ErrInvalidPriority is returned when an invalid priority is provided
	ErrInvalidPriority = errors.New("invalid priority")
)

// TaskManager handles task operations
type TaskManager struct {
	storage *TaskStorage
}

// NewTaskManager creates a new TaskManager instance
func NewTaskManager(storage *TaskStorage) *TaskManager {
	return &TaskManager{
		storage: storage,
	}
}

// Create creates a new task with the given parameters
func (m *TaskManager) Create(title, description string, priority Priority) (*Task, error) {
	// Validate title
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrEmptyTitle
	}

	// Validate priority
	if !IsValidPriority(string(priority)) {
		return nil, ErrInvalidPriority
	}

	// Create new task
	task := NewTask(title, description, priority)

	// Save to storage
	if err := m.storage.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// Get retrieves a task by ID
func (m *TaskManager) Get(id string) (*Task, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	task, exists := m.storage.Get(id)
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// List returns all tasks, optionally filtered by status
func (m *TaskManager) List(status Status) ([]*Task, error) {
	tasks := m.storage.Filter(status)
	return tasks, nil
}

// Update updates a task with the given ID using the provided updates
func (m *TaskManager) Update(id string, updates map[string]interface{}) (*Task, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	task, exists := m.storage.Get(id)
	if !exists {
		return nil, ErrTaskNotFound
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		title = strings.TrimSpace(title)
		if title == "" {
			return nil, ErrEmptyTitle
		}
		task.Title = title
	}

	if description, ok := updates["description"].(string); ok {
		task.Description = description
	}

	if status, ok := updates["status"].(string); ok {
		if !IsValidStatus(status) {
			return nil, ErrInvalidStatus
		}
		task.Status = Status(status)
	}

	if priority, ok := updates["priority"].(string); ok {
		if !IsValidPriority(priority) {
			return nil, ErrInvalidPriority
		}
		task.Priority = Priority(priority)
	}

	// Update timestamp
	task.UpdatedAt = time.Now()

	// Save changes
	if err := m.storage.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// Delete removes a task by ID
func (m *TaskManager) Delete(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	if !m.storage.Delete(id) {
		return ErrTaskNotFound
	}

	return nil
}

// Complete marks a task as completed
func (m *TaskManager) Complete(id string) (*Task, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	task, exists := m.storage.Get(id)
	if !exists {
		return nil, ErrTaskNotFound
	}

	task.Status = StatusCompleted
	task.UpdatedAt = time.Now()

	if err := m.storage.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}
