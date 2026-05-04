package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func getWorkspaceIndexPath(workspace string) string {
	return filepath.Join(workspace, ".ai-git", "workspace.json")
}

func getWorkspaceIndex(workspace string) (*WorkspaceIndex, error) {
	indexPath := getWorkspaceIndexPath(workspace)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return &WorkspaceIndex{
			Projects: make(map[string]ProjectMeta),
			Updated:  time.Now().Format(time.RFC3339),
		}, nil
	}

	var index WorkspaceIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse workspace index: %w", err)
	}

	return &index, nil
}

func saveWorkspaceIndex(workspace string, index *WorkspaceIndex) error {
	indexPath := getWorkspaceIndexPath(workspace)
	os.MkdirAll(filepath.Dir(indexPath), 0755)

	index.Updated = time.Now().Format(time.RFC3339)
	data, _ := json.MarshalIndent(index, "", "  ")
	return os.WriteFile(indexPath, data, 0644)
}
