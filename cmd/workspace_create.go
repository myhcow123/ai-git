package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createProjectStructure(path, projectType, name string) error {
	dirs := []string{}

	switch {
	case strings.HasPrefix(projectType, "go"):
		dirs = []string{"cmd", "internal", "pkg", "docs", "configs", "scripts"}
	case strings.HasPrefix(projectType, "python"):
		dirs = []string{"src", "tests", "docs", "configs", "scripts"}
	case strings.HasPrefix(projectType, "rust"):
		dirs = []string{"src", "tests", "docs"}
	case strings.HasPrefix(projectType, "javascript"):
		dirs = []string{"src", "tests", "docs", "public"}
	default:
		dirs = []string{"src", "docs", "tests"}
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(path, dir), 0755); err != nil {
			return err
		}
	}

	aiGitDir := filepath.Join(path, ".ai-git")
	if err := os.MkdirAll(aiGitDir, 0755); err != nil {
		return err
	}

	projectMeta := ProjectMeta{
		Name:        name,
		Type:        projectType,
		Language:    strings.Split(projectType, "-")[0],
		Status:      "active",
		Description: "",
		Created:     time.Now().Format(time.RFC3339),
		Modified:    time.Now().Format(time.RFC3339),
	}
	metaData, _ := json.MarshalIndent(projectMeta, "", "  ")
	os.WriteFile(filepath.Join(aiGitDir, "project.json"), metaData, 0644)

	readmeContent := fmt.Sprintf("# %s\n\nProject created by ai-git.\n\n## Description\n\nTODO: Add project description.\n", name)
	os.WriteFile(filepath.Join(path, "README.md"), []byte(readmeContent), 0644)

	switch {
	case strings.HasPrefix(projectType, "go"):
		goMod := fmt.Sprintf("module %s\n\ngo 1.21\n", name)
		os.WriteFile(filepath.Join(path, "go.mod"), []byte(goMod), 0644)

		mainContent := fmt.Sprintf("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello from %s!\")\n}\n", name)
		os.WriteFile(filepath.Join(path, "cmd", "main.go"), []byte(mainContent), 0644)

	case strings.HasPrefix(projectType, "python"):
		requirements := "# Project dependencies\n"
		os.WriteFile(filepath.Join(path, "requirements.txt"), []byte(requirements), 0644)

		mainContent := fmt.Sprintf("def main():\n    print(\"Hello from %s!\")\n\nif __name__ == \"__main__\":\n    main()\n", name)
		os.WriteFile(filepath.Join(path, "src", "main.py"), []byte(mainContent), 0644)

	case strings.HasPrefix(projectType, "rust"):
		cargoToml := fmt.Sprintf("[package]\nname = \"%s\"\nversion = \"0.1.0\"\nedition = \"2021\"\n\n[dependencies]\n", name)
		os.WriteFile(filepath.Join(path, "Cargo.toml"), []byte(cargoToml), 0644)

		mainContent := fmt.Sprintf("fn main() {\n    println!(\"Hello from %s!\");\n}\n", name)
		os.WriteFile(filepath.Join(path, "src", "main.rs"), []byte(mainContent), 0644)

	case strings.HasPrefix(projectType, "javascript"):
		packageJson := fmt.Sprintf("{\n  \"name\": \"%s\",\n  \"version\": \"1.0.0\",\n  \"main\": \"src/index.js\",\n  \"scripts\": {\n    \"start\": \"node src/index.js\"\n  }\n}\n", name)
		os.WriteFile(filepath.Join(path, "package.json"), []byte(packageJson), 0644)

		mainContent := fmt.Sprintf("console.log(\"Hello from %s!\");\n", name)
		os.WriteFile(filepath.Join(path, "src", "index.js"), []byte(mainContent), 0644)
	}

	return nil
}
