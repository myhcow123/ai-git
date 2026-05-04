package main

import (
	"os"
	"testing"
	"time"
)

func setupTestStorage(t *testing.T) (*TaskStorage, func()) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_tasks_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to create storage: %v", err)
	}

	cleanup := func() {
		storage = nil
		os.Remove(tmpFile.Name())
	}

	return storage, cleanup
}

func TestNewTask(t *testing.T) {
	task := NewTask("Test Task", "Test Description", PriorityHigh)

	if task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", task.Title)
	}
	if task.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", task.Description)
	}
	if task.Status != StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", task.Status)
	}
	if task.Priority != PriorityHigh {
		t.Errorf("Expected priority 'high', got '%s'", task.Priority)
	}
	if task.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if task.CreatedAt.IsZero() {
		t.Error("Expected non-zero CreatedAt")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("Expected non-zero UpdatedAt")
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status  string
		isValid bool
	}{
		{"pending", true},
		{"in_progress", true},
		{"completed", true},
		{"invalid", false},
		{"", false},
		{"PENDING", false},
	}

	for _, test := range tests {
		result := IsValidStatus(test.status)
		if result != test.isValid {
			t.Errorf("IsValidStatus('%s') = %v, expected %v", test.status, result, test.isValid)
		}
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		isValid  bool
	}{
		{"low", true},
		{"medium", true},
		{"high", true},
		{"invalid", false},
		{"", false},
		{"HIGH", false},
	}

	for _, test := range tests {
		result := IsValidPriority(test.priority)
		if result != test.isValid {
			t.Errorf("IsValidPriority('%s') = %v, expected %v", test.priority, result, test.isValid)
		}
	}
}

func TestTaskManager_Create(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	manager := NewTaskManager(storage)

	// Test valid creation
	task, err := manager.Create("Test Task", "Test Description", PriorityHigh)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	if task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", task.Title)
	}
	if task.Priority != PriorityHigh {
		t.Errorf("Expected priority 'high', got '%s'", task.Priority)
	}

	// Test empty title
	_, err = manager.Create("", "Description", PriorityMedium)
	if err != ErrEmptyTitle {
		t.Errorf("Expected ErrEmptyTitle, got %v", err)
	}

	// Test invalid priority
	_, err = manager.Create("Title", "Description", "invalid")
	if err != ErrInvalidPriority {
		t.Errorf("Expected ErrInvalidPriority, got %v", err)
	}

	// Test whitespace-only title
	_, err = manager.Create("   ", "Description", PriorityMedium)
	if err != ErrEmptyTitle {
		t.Errorf("Expected ErrEmptyTitle for whitespace title, got %v", err)
	}
}

func TestTaskManager_Get(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("Test Task", "Description", PriorityMedium)

	// Test valid get
	retrieved, err := manager.Get(task.ID)
	if err != nil {
		t.Fatalf("Failed to get task: %v", err)
	}
	if retrieved.ID != task.ID {
		t.Errorf("Expected ID '%s', got '%s'", task.ID, retrieved.ID)
	}

	// Test non-existent task
	_, err = manager.Get("non-existent-id")
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}

	// Test empty ID
	_, err = manager.Get("")
	if err != ErrInvalidID {
		t.Errorf("Expected ErrInvalidID, got %v", err)
	}
}

func TestTaskManager_List(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	manager := NewTaskManager(storage)

	// Create tasks with different statuses
	task1, _ := manager.Create("Task 1", "", PriorityLow)
	_, _ = manager.Create("Task 2", "", PriorityMedium)
	_, _ = manager.Create("Task 3", "", PriorityHigh)

	// Complete one task
	manager.Complete(task1.ID)

	// Test list all
	tasks, err := manager.List("")
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Test filter by status
	pendingTasks, err := manager.List(StatusPending)
	if err != nil {
		t.Fatalf("Failed to list pending tasks: %v", err)
	}
	if len(pendingTasks) != 2 {
		t.Errorf("Expected 2 pending tasks, got %d", len(pendingTasks))
	}

	// Test filter for completed
	completedTasks, err := manager.List(StatusCompleted)
	if err != nil {
		t.Fatalf("Failed to list completed tasks: %v", err)
	}
	if len(completedTasks) != 1 {
		t.Errorf("Expected 1 completed task, got %d", len(completedTasks))
	}
}

func TestTaskManager_Update(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("Original Title", "Original Description", PriorityLow)

	// Test update title
	updates := map[string]interface{}{"title": "Updated Title"}
	updated, err := manager.Update(task.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
	}

	// Test update status
	updates = map[string]interface{}{"status": "in_progress"}
	updated, err = manager.Update(task.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update task status: %v", err)
	}
	if updated.Status != StatusInProgress {
		t.Errorf("Expected status 'in_progress', got '%s'", updated.Status)
	}

	// Test update priority
	updates = map[string]interface{}{"priority": "high"}
	updated, err = manager.Update(task.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update task priority: %v", err)
	}
	if updated.Priority != PriorityHigh {
		t.Errorf("Expected priority 'high', got '%s'", updated.Priority)
	}

	// Test update description
	updates = map[string]interface{}{"description": "New description"}
	updated, err = manager.Update(task.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update task description: %v", err)
	}
	if updated.Description != "New description" {
		t.Errorf("Expected description 'New description', got '%s'", updated.Description)
	}

	// Test multiple updates
	updates = map[string]interface{}{
		"title":       "Multi Update",
		"status":      "completed",
		"priority":    "low",
		"description": "Updated in one call",
	}
	updated, err = manager.Update(task.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update task with multiple updates: %v", err)
	}
	if updated.Title != "Multi Update" || updated.Status != StatusCompleted {
		t.Error("Failed to apply multiple updates correctly")
	}

	// Test invalid status update
	updates = map[string]interface{}{"status": "invalid_status"}
	_, err = manager.Update(task.ID, updates)
	if err != ErrInvalidStatus {
		t.Errorf("Expected ErrInvalidStatus, got %v", err)
	}

	// Test invalid priority update
	updates = map[string]interface{}{"priority": "invalid_priority"}
	_, err = manager.Update(task.ID, updates)
	if err != ErrInvalidPriority {
		t.Errorf("Expected ErrInvalidPriority, got %v", err)
	}

	// Test empty title update
	updates = map[string]interface{}{"title": ""}
	_, err = manager.Update(task.ID, updates)
	if err != ErrEmptyTitle {
		t.Errorf("Expected ErrEmptyTitle, got %v", err)
	}

	// Test non-existent task
	_, err = manager.Update("non-existent-id", updates)
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}

	// Test empty ID
	_, err = manager.Update("", updates)
	if err != ErrInvalidID {
		t.Errorf("Expected ErrInvalidID, got %v", err)
	}
}

func TestTaskManager_Delete(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("To Delete", "Will be deleted", PriorityMedium)

	// Test successful delete
	err := manager.Delete(task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify task is deleted
	_, err = manager.Get(task.ID)
	if err != ErrTaskNotFound {
		t.Error("Task still exists after deletion")
	}

	// Test delete non-existent task
	err = manager.Delete("non-existent-id")
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}

	// Test delete with empty ID
	err = manager.Delete("")
	if err != ErrInvalidID {
		t.Errorf("Expected ErrInvalidID, got %v", err)
	}
}

func TestTaskManager_Complete(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("To Complete", "Will be completed", PriorityHigh)

	// Test successful complete
	completed, err := manager.Complete(task.ID)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}
	if completed.Status != StatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", completed.Status)
	}

	// Test complete non-existent task
	_, err = manager.Complete("non-existent-id")
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}

	// Test complete with empty ID
	_, err = manager.Complete("")
	if err != ErrInvalidID {
		t.Errorf("Expected ErrInvalidID, got %v", err)
	}
}

func TestTaskStorage_Persistence(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_persistence_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Create storage and add tasks
	storage1, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager1 := NewTaskManager(storage1)

	task1, _ := manager1.Create("Task 1", "Description 1", PriorityHigh)
	_, _ = manager1.Create("Task 2", "Description 2", PriorityMedium)

	// Create new storage from same file
	storage2, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create second storage: %v", err)
	}
	manager2 := NewTaskManager(storage2)

	// Verify tasks are persisted
	retrieved1, err := manager2.Get(task1.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve task 1: %v", err)
	}
	if retrieved1.Title != "Task 1" {
		t.Errorf("Expected title 'Task 1', got '%s'", retrieved1.Title)
	}

	// Verify updates persist
	manager1.Update(task1.ID, map[string]interface{}{"status": "completed"})
	storage2.Reload()

	retrieved1, _ = manager2.Get(task1.ID)
	if retrieved1.Status != StatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", retrieved1.Status)
	}
}

func TestTaskStorage_Delete(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_delete_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	task := NewTask("To Delete", "Will be deleted", PriorityMedium)
	storage.Save(task)

	// Delete existing task
	if !storage.Delete(task.ID) {
		t.Error("Delete returned false for existing task")
	}

	// Verify deletion
	if _, exists := storage.Get(task.ID); exists {
		t.Error("Task still exists after deletion")
	}

	// Delete non-existing task
	if storage.Delete("non-existent") {
		t.Error("Delete returned true for non-existing task")
	}
}

func TestTaskStorage_Filter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_filter_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Create tasks with different statuses
	tasks := []*Task{
		NewTask("Task 1", "", PriorityLow),
		NewTask("Task 2", "", PriorityMedium),
		NewTask("Task 3", "", PriorityHigh),
	}
	tasks[0].Status = StatusPending
	tasks[1].Status = StatusInProgress
	tasks[2].Status = StatusCompleted

	for _, task := range tasks {
		storage.Save(task)
	}

	// Test filter pending
	pending := storage.Filter(StatusPending)
	if len(pending) != 1 {
		t.Errorf("Expected 1 pending task, got %d", len(pending))
	}

	// Test filter in_progress
	inProgress := storage.Filter(StatusInProgress)
	if len(inProgress) != 1 {
		t.Errorf("Expected 1 in_progress task, got %d", len(inProgress))
	}

	// Test filter completed
	completed := storage.Filter(StatusCompleted)
	if len(completed) != 1 {
		t.Errorf("Expected 1 completed task, got %d", len(completed))
	}

	// Test filter with empty status (all tasks)
	all := storage.Filter("")
	if len(all) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(all))
	}
}

func TestTaskUpdatedAt(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	manager := NewTaskManager(storage)
	task, _ := manager.Create("Test Task", "", PriorityMedium)

	originalUpdatedAt := task.UpdatedAt

	// Wait a bit to ensure timestamp changes
	time.Sleep(10 * time.Millisecond)

	// Update the task
	manager.Update(task.ID, map[string]interface{}{"title": "Updated Title"})

	// Get the task again
	updated, _ := manager.Get(task.ID)

	if !updated.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt was not updated after modification")
	}
}
