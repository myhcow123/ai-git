package main

import (
	"time"

	"github.com/google/uuid"
)

// Status represents the status of a task
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

// Priority represents the priority of a task
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Task represents a task in the management system
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewTask creates a new task with a generated ID and current timestamps
func NewTask(title, description string, priority Priority) *Task {
	now := time.Now()
	return &Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      StatusPending,
		Priority:    priority,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsValidStatus checks if a status string is valid
func IsValidStatus(s string) bool {
	switch Status(s) {
	case StatusPending, StatusInProgress, StatusCompleted:
		return true
	}
	return false
}

// IsValidPriority checks if a priority string is valid
func IsValidPriority(p string) bool {
	switch Priority(p) {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}
