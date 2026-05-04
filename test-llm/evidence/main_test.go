package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
		rest     []string
	}{
		{
			name:     "single flag",
			args:     []string{"-d", "description", "positional"},
			expected: map[string]string{"d": "description"},
			rest:     []string{"positional"},
		},
		{
			name:     "multiple flags",
			args:     []string{"-d", "desc", "-p", "high", "title"},
			expected: map[string]string{"d": "desc", "p": "high"},
			rest:     []string{"title"},
		},
		{
			name:     "flag without value",
			args:     []string{"-d", "-p", "high"},
			expected: map[string]string{"d": "", "p": "high"},
			rest:     []string{},
		},
		{
			name:     "no flags",
			args:     []string{"arg1", "arg2"},
			expected: map[string]string{},
			rest:     []string{"arg1", "arg2"},
		},
		{
			name:     "mixed positional and flags",
			args:     []string{"arg1", "-d", "desc", "arg2"},
			expected: map[string]string{"d": "desc"},
			rest:     []string{"arg1", "arg2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flags, remaining := parseFlags(test.args)

			for key, expectedValue := range test.expected {
				if flags[key] != expectedValue {
					t.Errorf("Expected flag %s='%s', got '%s'", key, expectedValue, flags[key])
				}
			}

			if len(remaining) != len(test.rest) {
				t.Errorf("Expected %d remaining args, got %d", len(test.rest), len(remaining))
			}

			for i, arg := range test.rest {
				if remaining[i] != arg {
					t.Errorf("Expected remaining arg %d='%s', got '%s'", i, arg, remaining[i])
				}
			}
		})
	}
}

func TestPrintUsage(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printUsage()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()

	if !strings.Contains(output, "Task Manager") {
		t.Error("Help output should contain 'Task Manager'")
	}
	if !strings.Contains(output, "add") {
		t.Error("Help output should contain 'add' command")
	}
	if !strings.Contains(output, "list") {
		t.Error("Help output should contain 'list' command")
	}
}

func TestHandleList_Empty(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_list_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handleList(manager, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()

	if !strings.Contains(output, "No tasks found") {
		t.Error("Expected 'No tasks found' for empty list")
	}
}

func TestHandleAdd_ValidInput(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_add_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// This should succeed
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
}

func TestHandleShow_InvalidID(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_show_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// This should fail
	_, err = manager.Get("non-existent-id")
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

func TestCLICommands(t *testing.T) {
	// Test command constants
	if cmdAdd != "add" {
		t.Errorf("Expected cmdAdd='add', got '%s'", cmdAdd)
	}
	if cmdList != "list" {
		t.Errorf("Expected cmdList='list', got '%s'", cmdList)
	}
	if cmdShow != "show" {
		t.Errorf("Expected cmdShow='show', got '%s'", cmdShow)
	}
	if cmdUpdate != "update" {
		t.Errorf("Expected cmdUpdate='update', got '%s'", cmdUpdate)
	}
	if cmdComplete != "complete" {
		t.Errorf("Expected cmdComplete='complete', got '%s'", cmdComplete)
	}
	if cmdDelete != "delete" {
		t.Errorf("Expected cmdDelete='delete', got '%s'", cmdDelete)
	}
	if cmdHelp != "help" {
		t.Errorf("Expected cmdHelp='help', got '%s'", cmdHelp)
	}
}

func TestStatusConstants(t *testing.T) {
	if StatusPending != "pending" {
		t.Errorf("Expected StatusPending='pending', got '%s'", StatusPending)
	}
	if StatusInProgress != "in_progress" {
		t.Errorf("Expected StatusInProgress='in_progress', got '%s'", StatusInProgress)
	}
	if StatusCompleted != "completed" {
		t.Errorf("Expected StatusCompleted='completed', got '%s'", StatusCompleted)
	}
}

func TestPriorityConstants(t *testing.T) {
	if PriorityLow != "low" {
		t.Errorf("Expected PriorityLow='low', got '%s'", PriorityLow)
	}
	if PriorityMedium != "medium" {
		t.Errorf("Expected PriorityMedium='medium', got '%s'", PriorityMedium)
	}
	if PriorityHigh != "high" {
		t.Errorf("Expected PriorityHigh='high', got '%s'", PriorityHigh)
	}
}

func TestHandleList_WithTasks(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_list_with_tasks_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create some tasks
	manager.Create("Task 1", "Description 1", PriorityHigh)
	manager.Create("Task 2", "Description 2", PriorityMedium)
	manager.Create("Task 3", "Description 3", PriorityLow)

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	handleList(manager, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	output := buf.String()

	if !strings.Contains(output, "Total: 3 task(s)") {
		t.Error("Expected 'Total: 3 task(s)' for list with tasks")
	}
}

func TestHandleList_FilterByStatus(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_list_filter_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create tasks
	task1, _ := manager.Create("Task 1", "", PriorityHigh)
	manager.Create("Task 2", "", PriorityMedium)
	manager.Complete(task1.ID)

	// Filter by status
	tasks, _ := manager.List(StatusCompleted)
	if len(tasks) != 1 {
		t.Errorf("Expected 1 completed task, got %d", len(tasks))
	}
}

func TestHandleComplete(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_complete_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("To Complete", "", PriorityHigh)

	// Complete it
	completed, err := manager.Complete(task.ID)
	if err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	if completed.Status != StatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", completed.Status)
	}
}

func TestHandleDelete(t *testing.T) {
	// Create a temp file
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
	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("To Delete", "", PriorityMedium)

	// Delete it
	err = manager.Delete(task.ID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify it's deleted
	_, err = manager.Get(task.ID)
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound after deletion, got %v", err)
	}
}

func TestHandleUpdate(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_update_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("Original Title", "Original Description", PriorityLow)

	// Update it
	updates := map[string]interface{}{
		"title":       "Updated Title",
		"description": "Updated Description",
		"status":      "in_progress",
		"priority":    "high",
	}

	updated, err := manager.Update(task.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", updated.Title)
	}
	if updated.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got '%s'", updated.Description)
	}
	if updated.Status != StatusInProgress {
		t.Errorf("Expected status 'in_progress', got '%s'", updated.Status)
	}
	if updated.Priority != PriorityHigh {
		t.Errorf("Expected priority 'high', got '%s'", updated.Priority)
	}
}

func TestUpdateTimestamp(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_timestamp_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("Test Task", "", PriorityMedium)
	originalUpdated := task.UpdatedAt

	// Update the task
	manager.Update(task.ID, map[string]interface{}{"title": "Updated"})

	// Get the updated task
	updated, _ := manager.Get(task.ID)

	if !updated.UpdatedAt.After(originalUpdated) {
		t.Error("UpdatedAt should be updated after modification")
	}
}

func TestListAllTasks(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_list_all_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create tasks
	manager.Create("Task 1", "", PriorityLow)
	manager.Create("Task 2", "", PriorityMedium)
	manager.Create("Task 3", "", PriorityHigh)

	// List all
	tasks, err := manager.List("")
	if err != nil {
		t.Fatalf("Failed to list tasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}
}

func TestCreateTaskWithEmptyDescription(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_empty_desc_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create a task with empty description
	task, err := manager.Create("Task with Empty Desc", "", PriorityMedium)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	if task.Description != "" {
		t.Errorf("Expected empty description, got '%s'", task.Description)
	}
}

func TestCompleteAlreadyCompletedTask(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_complete_twice_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create and complete a task
	task, _ := manager.Create("Task", "", PriorityHigh)
	manager.Complete(task.ID)

	// Complete again - should still work
	completed, err := manager.Complete(task.ID)
	if err != nil {
		t.Fatalf("Failed to complete task twice: %v", err)
	}

	if completed.Status != StatusCompleted {
		t.Errorf("Expected status 'completed', got '%s'", completed.Status)
	}
}

func TestUpdateTaskWithPartialUpdates(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test_partial_update_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewTaskStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	manager := NewTaskManager(storage)

	// Create a task
	task, _ := manager.Create("Original Title", "Original Description", PriorityMedium)

	// Update only title
	updates := map[string]interface{}{"title": "New Title Only"}
	updated, err := manager.Update(task.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	if updated.Title != "New Title Only" {
		t.Errorf("Expected title 'New Title Only', got '%s'", updated.Title)
	}

	// Description and priority should remain unchanged
	if updated.Description != "Original Description" {
		t.Errorf("Expected description 'Original Description', got '%s'", updated.Description)
	}
	if updated.Priority != PriorityMedium {
		t.Errorf("Expected priority 'medium', got '%s'", updated.Priority)
	}
}

func TestValidateStatusAndPriority(t *testing.T) {
	// Test valid statuses
	if !IsValidStatus("pending") {
		t.Error("pending should be a valid status")
	}
	if !IsValidStatus("in_progress") {
		t.Error("in_progress should be a valid status")
	}
	if !IsValidStatus("completed") {
		t.Error("completed should be a valid status")
	}

	// Test invalid statuses
	if IsValidStatus("invalid") {
		t.Error("invalid should not be a valid status")
	}
	if IsValidStatus("PENDING") {
		t.Error("PENDING (uppercase) should not be a valid status")
	}

	// Test valid priorities
	if !IsValidPriority("low") {
		t.Error("low should be a valid priority")
	}
	if !IsValidPriority("medium") {
		t.Error("medium should be a valid priority")
	}
	if !IsValidPriority("high") {
		t.Error("high should be a valid priority")
	}

	// Test invalid priorities
	if IsValidPriority("invalid") {
		t.Error("invalid should not be a valid priority")
	}
	if IsValidPriority("HIGH") {
		t.Error("HIGH (uppercase) should not be a valid priority")
	}
}
