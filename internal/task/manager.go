package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusBlocked    TaskStatus = "blocked"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

type SubTask struct {
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

type TaskLog struct {
	Timestamp time.Time `json:"timestamp"`
	Step      string    `json:"step"`
	Tool      string    `json:"tool"`
	Input     string    `json:"input"`
	Output    string    `json:"output"`
	Status    string    `json:"status"`
}

type Task struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Progress    int          `json:"progress"`
	Tags        []string     `json:"tags"`
	SubTasks    []SubTask    `json:"subtasks"`
	Logs        []TaskLog    `json:"logs"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CompletedAt *time.Time   `json:"completed_at"`
}

type Manager struct {
	tasksDir string
	tasks    map[string]*Task
}

func NewManager(tasksDir string) *Manager {
	return &Manager{
		tasksDir: tasksDir,
		tasks:    make(map[string]*Task),
	}
}

func (m *Manager) Init() error {
	dirs := []string{
		filepath.Join(m.tasksDir, "active"),
		filepath.Join(m.tasksDir, "backlog"),
		filepath.Join(m.tasksDir, "archive"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return nil
}

func (m *Manager) Create(name string, opts ...TaskOption) (*Task, error) {
	task := &Task{
		ID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Name:      name,
		Status:    TaskStatusPending,
		Priority:  TaskPriorityMedium,
		Progress:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(task)
	}

	if err := m.saveTask(task); err != nil {
		return nil, err
	}

	m.tasks[task.ID] = task
	return task, nil
}

func (m *Manager) Get(id string) (*Task, error) {
	if task, exists := m.tasks[id]; exists {
		return task, nil
	}

	task, err := m.loadTask(id)
	if err != nil {
		return nil, err
	}

	m.tasks[id] = task
	return task, nil
}

func (m *Manager) Update(id string, opts ...TaskOption) (*Task, error) {
	task, err := m.Get(id)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(task)
	}

	task.UpdatedAt = time.Now()

	if err := m.saveTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (m *Manager) Complete(id string) (*Task, error) {
	now := time.Now()
	return m.Update(id,
		WithStatus(TaskStatusCompleted),
		WithProgress(100),
		WithCompletedAt(&now),
	)
}

func (m *Manager) List(status TaskStatus) ([]*Task, error) {
	var tasks []*Task

	activeDir := filepath.Join(m.tasksDir, "active")
	files, err := os.ReadDir(activeDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			id := strings.TrimSuffix(file.Name(), ".json")
			task, err := m.Get(id)
			if err != nil {
				continue
			}

			if status == "" || task.Status == status {
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

func (m *Manager) AddLog(id string, log TaskLog) error {
	task, err := m.Get(id)
	if err != nil {
		return err
	}

	log.Timestamp = time.Now()
	task.Logs = append(task.Logs, log)
	task.UpdatedAt = time.Now()

	return m.saveTask(task)
}

func (m *Manager) Archive(id string) error {
	task, err := m.Get(id)
	if err != nil {
		return err
	}

	oldPath := m.getTaskPath(task, "active")
	newPath := m.getTaskPath(task, "archive")

	archiveDir := filepath.Dir(newPath)
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}

	return os.Rename(oldPath, newPath)
}

func (m *Manager) saveTask(task *Task) error {
	path := m.getTaskPath(task, "active")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(task, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (m *Manager) loadTask(id string) (*Task, error) {
	dirs := []string{"active", "backlog", "archive"}

	for _, dir := range dirs {
		path := filepath.Join(m.tasksDir, dir, id+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return nil, err
		}

		return &task, nil
	}

	return nil, fmt.Errorf("task not found: %s", id)
}

func (m *Manager) getTaskPath(task *Task, category string) string {
	return filepath.Join(m.tasksDir, category, task.ID+".json")
}

type TaskOption func(*Task)

func WithDescription(desc string) TaskOption {
	return func(t *Task) { t.Description = desc }
}

func WithStatus(status TaskStatus) TaskOption {
	return func(t *Task) { t.Status = status }
}

func WithPriority(priority TaskPriority) TaskOption {
	return func(t *Task) { t.Priority = priority }
}

func WithProgress(progress int) TaskOption {
	return func(t *Task) {
		if progress < 0 {
			progress = 0
		}
		if progress > 100 {
			progress = 100
		}
		t.Progress = progress
	}
}

func WithTags(tags []string) TaskOption {
	return func(t *Task) { t.Tags = tags }
}

func WithSubTasks(subtasks []SubTask) TaskOption {
	return func(t *Task) { t.SubTasks = subtasks }
}

func WithCompletedAt(t *time.Time) TaskOption {
	return func(task *Task) { task.CompletedAt = t }
}
