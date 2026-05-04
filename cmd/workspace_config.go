package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type GlobalConfig struct {
	Workspace string             `json:"workspace"`
	Structure WorkspaceStructure `json:"structure"`
	Created   string             `json:"created"`
}

type WorkspaceStructure struct {
	Active      string `json:"active"`
	Experiments string `json:"experiments"`
	Research    string `json:"research"`
	Archive     string `json:"archive"`
}

type ProjectMeta struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`
	Language    string `json:"language"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Created     string `json:"created"`
	Modified    string `json:"modified"`
}

type WorkspaceIndex struct {
	Projects map[string]ProjectMeta `json:"projects"`
	Updated  string                 `json:"updated"`
}

func getGlobalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-git", "config.json")
}

func getGlobalConfig() (*GlobalConfig, error) {
	configPath := getGlobalConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("global config not found, run 'ai-git workspace init' first")
	}

	var config GlobalConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func saveGlobalConfig(config *GlobalConfig) error {
	configPath := getGlobalConfigPath()
	os.MkdirAll(filepath.Dir(configPath), 0755)

	data, _ := json.MarshalIndent(config, "", "  ")
	return os.WriteFile(configPath, data, 0644)
}

func configPath() string {
	return getGlobalConfigPath()
}
