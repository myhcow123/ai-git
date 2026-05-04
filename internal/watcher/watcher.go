package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mychow/ai-git/internal/parser"
	"github.com/mychow/ai-git/internal/storage"
)

type EventType int

const (
	EventCreate EventType = iota
	EventModify
	EventDelete
)

type FileEvent struct {
	Project string
	Path    string
	Type    EventType
}

type ProjectWatcher struct {
	Path      string
	Storage   *storage.Storage
	Parser    *parser.CodeParser
	FileMetas map[string]*storage.FileMeta
	mu        sync.RWMutex
}

type Watcher struct {
	projects  map[string]*ProjectWatcher
	eventChan chan FileEvent
	stopChan  chan struct{}
	mu        sync.RWMutex
	running   bool
}

func NewWatcher() *Watcher {
	return &Watcher{
		projects:  make(map[string]*ProjectWatcher),
		eventChan: make(chan FileEvent, 1000),
		stopChan:  make(chan struct{}),
	}
}

func (w *Watcher) AddProject(projectPath string, stor *storage.Storage) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, exists := w.projects[absPath]; exists {
		return fmt.Errorf("project already being watched: %s", absPath)
	}

	if stor == nil {
		dbPath := filepath.Join(absPath, ".ai-git", "ai-git.db")
		stor, err = storage.NewStorage(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create storage: %w", err)
		}
	}

	pw := &ProjectWatcher{
		Path:      absPath,
		Storage:   stor,
		Parser:    parser.NewCodeParser(),
		FileMetas: make(map[string]*storage.FileMeta),
	}

	metas, err := stor.GetAllFileMetas()
	if err == nil {
		pw.FileMetas = metas
	}

	w.projects[absPath] = pw

	return nil
}

func (w *Watcher) RemoveProject(projectPath string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	delete(w.projects, absPath)
	return nil
}

func (w *Watcher) Start() {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return
	}
	w.running = true
	w.mu.Unlock()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.scanProjects()
		}
	}
}

func (w *Watcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.running {
		close(w.stopChan)
		w.running = false
	}
}

func (w *Watcher) scanProjects() {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, pw := range w.projects {
		pw.Scan()
	}
}

func (pw *ProjectWatcher) Scan() {
	filepath.Walk(pw.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if strings.HasPrefix(filepath.Base(path), ".") || 
			   strings.Contains(path, "node_modules") ||
			   strings.Contains(path, "vendor") {
				return filepath.SkipDir
			}
			return nil
		}

		if !pw.shouldWatch(path) {
			return nil
		}

		pw.checkFile(path, info)
		return nil
	})
}

func (pw *ProjectWatcher) shouldWatch(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	watchedExts := map[string]bool{
		".go":       true,
		".py":       true,
		".js":       true,
		".ts":       true,
		".jsx":      true,
		".tsx":      true,
		".java":     true,
		".c":        true,
		".cpp":      true,
		".h":        true,
		".hpp":      true,
		".rs":       true,
		".md":       true,
		".markdown": true,
	}
	return watchedExts[ext]
}

func (pw *ProjectWatcher) checkFile(path string, info os.FileInfo) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	relPath, _ := filepath.Rel(pw.Path, path)

	currentMeta := &storage.FileMeta{
		ModTime: info.ModTime().Unix(),
		Size:    info.Size(),
	}

	oldMeta, exists := pw.FileMetas[relPath]

	if !exists {
		pw.indexFile(relPath, path)
		pw.FileMetas[relPath] = currentMeta
		pw.Storage.SaveFileMeta(relPath, currentMeta)
		return
	}

	if oldMeta.ModTime != currentMeta.ModTime || oldMeta.Size != currentMeta.Size {
		pw.reindexFile(relPath, path)
		pw.FileMetas[relPath] = currentMeta
		pw.Storage.SaveFileMeta(relPath, currentMeta)
	}
}

func (pw *ProjectWatcher) indexFile(relPath, absPath string) {
	content, err := os.ReadFile(absPath)
	if err != nil {
		return
	}

	symbols, err := pw.Parser.ParseFile(absPath, string(content))
	if err != nil {
		return
	}

	for _, sym := range symbols {
		sym.File = relPath
		pw.Storage.SaveSymbol(sym)
	}

	pw.Storage.SaveFileMeta(relPath, &storage.FileMeta{
		SymbolCount: len(symbols),
	})
}

func (pw *ProjectWatcher) reindexFile(relPath, absPath string) {
	pw.Storage.DeleteSymbolsForFile(relPath)
	pw.indexFile(relPath, absPath)
}

func (pw *ProjectWatcher) GetStats() map[string]interface{} {
	pw.mu.RLock()
	defer pw.mu.RUnlock()

	return map[string]interface{}{
		"path":       pw.Path,
		"file_count": len(pw.FileMetas),
	}
}

func (w *Watcher) GetProjects() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	projects := make([]string, 0, len(w.projects))
	for path := range w.projects {
		projects = append(projects, path)
	}
	return projects
}

func (w *Watcher) GetStats() map[string]interface{} {
	w.mu.RLock()
	defer w.mu.RUnlock()

	projects := make([]map[string]interface{}, 0)
	for _, pw := range w.projects {
		projects = append(projects, pw.GetStats())
	}

	return map[string]interface{}{
		"running":       w.running,
		"project_count": len(w.projects),
		"projects":      projects,
	}
}
