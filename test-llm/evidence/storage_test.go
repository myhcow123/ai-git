package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTaskStorage_CreatingNewFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_new_storage_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "tasks.json")

	storage, err := NewTaskStorage(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create storage for new file: %v", err)
	}

	if storage == nil {
		t.Error("Expected non-nil storage")
	}

	// Should be empty
	tasks := storage.GetAll()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestNewTaskStorage_LoadingExistingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_load_storage_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "tasks.json")

	// Create a storage and add tasks
	storage1, err := NewTaskStorage(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	task := NewTask("Existing Task", "Description", PriorityHigh)
	storage1.Save(task)

	// Create a new storage from the same file
	storage2, err := NewTaskStorage(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create second storage: %v", err)
	}

	// Verify task was loaded
	tasks := storage2.GetAll()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Title != "Existing Task" {
		t.Errorf("Expected title 'Existing Task', got '%s'", tasks[0].Title)
	}
}

func TestTaskStorage_Save(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_save_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	task := NewTask("Save Test", "Testing save", PriorityMedium)
	err = storage.Save(task)
	if err != nil {
		t.Fatalf("Failed to save task: %v", err)
	}

	// Verify task was saved
	retrieved, exists := storage.Get(task.ID)
	if !exists {
		t.Error("Task was not saved")
	}

	if retrieved.Title != "Save Test" {
		t.Errorf("Expected title 'Save Test', got '%s'", retrieved.Title)
	}
}

func TestTaskStorage_Get(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_get_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	task := NewTask("Get Test", "Testing get", PriorityLow)
	storage.Save(task)

	// Test existing task
	retrieved, exists := storage.Get(task.ID)
	if !exists {
		t.Error("Expected task to exist")
	}
	if retrieved.ID != task.ID {
		t.Errorf("Expected ID '%s', got '%s'", task.ID, retrieved.ID)
	}

	// Test non-existing task
	_, exists = storage.Get("non-existent-id")
	if exists {
		t.Error("Expected task to not exist")
	}
}

func TestTaskStorage_GetAll(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_getall_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test empty storage
	tasks := storage.GetAll()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}

	// Add tasks
	task1 := NewTask("Task 1", "", PriorityLow)
	task2 := NewTask("Task 2", "", PriorityMedium)
	task3 := NewTask("Task 3", "", PriorityHigh)
	storage.Save(task1)
	storage.Save(task2)
	storage.Save(task3)

	// Test all tasks
	tasks = storage.GetAll()
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}
}

func TestTaskStorage_Reload(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_reload_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Create storage and add task
	storage1, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	storage1.Save(NewTask("Task 1", "", PriorityHigh))

	// Create second storage and reload
	storage2, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create second storage: %v", err)
	}

	// Initial load should have 1 task
	tasks := storage2.GetAll()
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(tasks))
	}

	// Add another task via storage1
	storage1.Save(NewTask("Task 2", "", PriorityMedium))

	// Reload storage2
	err = storage2.Reload()
	if err != nil {
		t.Fatalf("Failed to reload storage: %v", err)
	}

	// Should now have 2 tasks
	tasks = storage2.GetAll()
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks after reload, got %d", len(tasks))
	}
}

func TestTaskStorage_ConcurrentAccess(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_concurrent_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Test concurrent reads and writes
	done := make(chan bool, 2)

	// Write goroutine
	go func() {
		for i := 0; i < 100; i++ {
			task := NewTask("Task", "", PriorityMedium)
			storage.Save(task)
		}
		done <- true
	}()

	// Read goroutine
	go func() {
		for i := 0; i < 100; i++ {
			storage.GetAll()
		}
		done <- true
	}()

	// Wait for goroutines
	<-done
	<-done
}

func TestTaskStorage_LoadEmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_empty_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Write empty content
	err = os.WriteFile(tmpFile.Name(), []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	// Should not fail
	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage with empty file: %v", err)
	}

	tasks := storage.GetAll()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}
