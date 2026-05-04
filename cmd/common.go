package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mychow/ai-git/internal/aql"
	"github.com/mychow/ai-git/internal/config"
)

var useAPI bool

func GetEngine(projectPath string) (*aql.Engine, error) {
	if shouldUseAPI() && useAPI {
		return nil, fmt.Errorf("API mode: use APIClient instead")
	}

	path := projectPath
	if path == "" {
		path = "."
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	engine, err := aql.NewEngine(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize engine: %w", err)
	}

	return engine, nil
}

func GetEngineOrAPI(projectPath string) (interface{}, error) {
	if shouldUseAPI() {
		return getAPIClient(), nil
	}
	return GetEngine(projectPath)
}

func GetEngineFromArgs(args []string) (*aql.Engine, error) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	return GetEngine(path)
}

func GetAbsPath(path string) (string, error) {
	if path == "" {
		path = "."
	}
	return filepath.Abs(path)
}

func LoadConfig(configPath string) (*config.Config, error) {
	if configPath == "" {
		defaultCfg := config.DefaultConfig
		return &defaultCfg, nil
	}
	return config.Load(configPath)
}

func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
