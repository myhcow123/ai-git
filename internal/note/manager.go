package note

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mychow/ai-git/internal/parser"
)

type NoteType string

const (
	NoteTypeKnowledge NoteType = "knowledge"
	NoteTypeContext   NoteType = "context"
	NoteTypeLog       NoteType = "log"
)

type Note struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Type        NoteType  `json:"type"`
	Content     string    `json:"content"`
	Tags        []string  `json:"tags"`
	Links       []string  `json:"links"`
	CodeRefs    []string  `json:"code_refs"`
	TaskRefs    []string  `json:"task_refs"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Manager struct {
	notesDir string
	notes    map[string]*Note
}

func NewManager(notesDir string) *Manager {
	return &Manager{
		notesDir: notesDir,
		notes:    make(map[string]*Note),
	}
}

func (m *Manager) Init() error {
	dirs := []string{
		filepath.Join(m.notesDir, "knowledge"),
		filepath.Join(m.notesDir, "context"),
		filepath.Join(m.notesDir, "log"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	return nil
}

func (m *Manager) Create(title string, noteType NoteType, opts ...NoteOption) (*Note, error) {
	note := &Note{
		ID:        fmt.Sprintf("note-%d", time.Now().UnixNano()),
		Title:     title,
		Type:      noteType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(note)
	}

	note.Tags = parser.ParseMarkdownTags(note.Content)
	note.Links = parser.ParseMarkdownLinks(note.Content)

	if err := m.saveNote(note); err != nil {
		return nil, err
	}

	m.notes[note.ID] = note
	return note, nil
}

func (m *Manager) Get(id string) (*Note, error) {
	if note, exists := m.notes[id]; exists {
		return note, nil
	}

	note, err := m.loadNote(id)
	if err != nil {
		return nil, err
	}

	m.notes[id] = note
	return note, nil
}

func (m *Manager) Update(id string, opts ...NoteOption) (*Note, error) {
	note, err := m.Get(id)
	if err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(note)
	}

	note.Tags = parser.ParseMarkdownTags(note.Content)
	note.Links = parser.ParseMarkdownLinks(note.Content)
	note.UpdatedAt = time.Now()

	if err := m.saveNote(note); err != nil {
		return nil, err
	}

	return note, nil
}

func (m *Manager) Delete(id string) error {
	note, err := m.Get(id)
	if err != nil {
		return err
	}

	path := m.getNotePath(note)
	return os.Remove(path)
}

func (m *Manager) List(noteType NoteType) ([]*Note, error) {
	var notes []*Note

	var searchDirs []string
	if noteType != "" {
		searchDirs = []string{string(noteType)}
	} else {
		searchDirs = []string{"knowledge", "context", "log"}
	}

	for _, dir := range searchDirs {
		fullDir := filepath.Join(m.notesDir, dir)
		files, err := os.ReadDir(fullDir)
		if err != nil {
			continue
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") {
				id := strings.TrimSuffix(file.Name(), ".json")
				note, err := m.Get(id)
				if err != nil {
					continue
				}
				notes = append(notes, note)
			}
		}
	}

	return notes, nil
}

func (m *Manager) Search(query string) ([]*Note, error) {
	notes, err := m.List("")
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []*Note

	for _, note := range notes {
		if strings.Contains(strings.ToLower(note.Title), query) ||
			strings.Contains(strings.ToLower(note.Content), query) {
			results = append(results, note)
			continue
		}

		for _, tag := range note.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, note)
				break
			}
		}
	}

	return results, nil
}

func (m *Manager) LinkCode(noteID string, codeRef string) error {
	note, err := m.Get(noteID)
	if err != nil {
		return err
	}

	for _, ref := range note.CodeRefs {
		if ref == codeRef {
			return nil
		}
	}

	note.CodeRefs = append(note.CodeRefs, codeRef)
	note.UpdatedAt = time.Now()

	return m.saveNote(note)
}

func (m *Manager) LinkTask(noteID string, taskID string) error {
	note, err := m.Get(noteID)
	if err != nil {
		return err
	}

	for _, ref := range note.TaskRefs {
		if ref == taskID {
			return nil
		}
	}

	note.TaskRefs = append(note.TaskRefs, taskID)
	note.UpdatedAt = time.Now()

	return m.saveNote(note)
}

func (m *Manager) Export(format string) (string, error) {
	notes, err := m.List("")
	if err != nil {
		return "", err
	}

	var result strings.Builder

	for _, note := range notes {
		switch format {
		case "markdown", "md":
			result.WriteString(fmt.Sprintf("# %s\n\n", note.Title))
			result.WriteString(fmt.Sprintf("Type: %s\n", note.Type))
			result.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(note.Tags, ", ")))
			result.WriteString(fmt.Sprintf("Created: %s\n\n", note.CreatedAt.Format(time.RFC3339)))
			result.WriteString(note.Content)
			result.WriteString("\n\n---\n\n")
		case "json":
			data, _ := json.MarshalIndent(note, "", "  ")
			result.Write(data)
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

func (m *Manager) saveNote(note *Note) error {
	path := m.getNotePath(note)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(note, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (m *Manager) loadNote(id string) (*Note, error) {
	dirs := []string{"knowledge", "context", "log"}

	for _, dir := range dirs {
		path := filepath.Join(m.notesDir, dir, id+".json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var note Note
		if err := json.Unmarshal(data, &note); err != nil {
			return nil, err
		}

		return &note, nil
	}

	return nil, fmt.Errorf("note not found: %s", id)
}

func (m *Manager) getNotePath(note *Note) string {
	return filepath.Join(m.notesDir, string(note.Type), note.ID+".json")
}

type NoteOption func(*Note)

func WithContent(content string) NoteOption {
	return func(n *Note) { n.Content = content }
}

func WithTags(tags []string) NoteOption {
	return func(n *Note) { n.Tags = tags }
}

func WithCodeRefs(refs []string) NoteOption {
	return func(n *Note) { n.CodeRefs = refs }
}

func WithTaskRefs(refs []string) NoteOption {
	return func(n *Note) { n.TaskRefs = refs }
}
