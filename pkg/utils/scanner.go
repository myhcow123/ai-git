package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Path     string
	Language string
	Size     int64
}

type ProjectScanner struct {
	rootPath    string
	excludeDirs map[string]bool
	includeExts map[string]bool
	maxFileSize int64
}

func NewProjectScanner(rootPath string) *ProjectScanner {
	return &ProjectScanner{
		rootPath: rootPath,
		excludeDirs: map[string]bool{
			".git":         true,
			"node_modules": true,
			"vendor":       true,
			"dist":         true,
			"build":        true,
			"target":       true,
			"bin":          true,
			"__pycache__":  true,
			".idea":        true,
			".vscode":      true,
		},
		includeExts: map[string]bool{
			".py":       true,
			".go":       true,
			".js":       true,
			".ts":       true,
			".tsx":      true,
			".java":     true,
			".rs":       true,
			".c":        true,
			".cpp":      true,
			".h":        true,
			".hpp":      true,
			".md":       true,
			".markdown": true,
		},
		maxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

func (s *ProjectScanner) Scan() ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(s.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if s.shouldExcludeDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !s.shouldIncludeFile(path, info) {
			return nil
		}

		language := s.detectLanguage(path)
		if language == "" {
			return nil
		}

		files = append(files, FileInfo{
			Path:     path,
			Language: language,
			Size:     info.Size(),
		})

		return nil
	})

	return files, err
}

func (s *ProjectScanner) shouldExcludeDir(path string) bool {
	base := filepath.Base(path)
	return s.excludeDirs[base]
}

func (s *ProjectScanner) shouldIncludeFile(path string, info os.FileInfo) bool {
	if info.Size() > s.maxFileSize {
		return false
	}

	ext := strings.ToLower(filepath.Ext(path))
	return s.includeExts[ext]
}

func (s *ProjectScanner) detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".py":
		return "python"
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".java":
		return "java"
	case ".rs":
		return "rust"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hxx":
		return "cpp"
	case ".md", ".markdown":
		return "markdown"
	default:
		return ""
	}
}

func (s *ProjectScanner) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return string(content), nil
}

func (s *ProjectScanner) GetProjectInfo() (map[string]interface{}, error) {
	absPath, err := filepath.Abs(s.rootPath)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"path":       absPath,
		"name":       filepath.Base(absPath),
		"languages":  make(map[string]int),
		"total_size": int64(0),
	}

	files, err := s.Scan()
	if err != nil {
		return nil, err
	}

	languages := info["languages"].(map[string]int)
	var totalSize int64

	for _, file := range files {
		languages[file.Language]++
		totalSize += file.Size
	}

	info["total_files"] = len(files)
	info["total_size"] = totalSize

	return info, nil
}
