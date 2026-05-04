package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspace and projects",
	Long: `Manage workspace and projects.

Examples:
  ai-git workspace init ~/projects    # Initialize workspace
  ai-git workspace list               # List all projects
  ai-git workspace find "api"         # Find projects
  ai-git workspace path <name>        # Get project path`,
}

var workspaceInitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize workspace",
	Long: `Initialize workspace at specified path.

Examples:
  ai-git workspace init ~/projects
  ai-git workspace init /home/user/workspace`,
	Args: cobra.MaximumNArgs(1),
	RunE:  runWorkspaceInit,
}

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `List all projects in workspace.`,
	RunE:  runWorkspaceList,
}

var workspaceFindCmd = &cobra.Command{
	Use:   "find [query]",
	Short: "Find projects",
	Long: `Find projects matching query.

Examples:
  ai-git workspace find "api"
  ai-git workspace find --language go
  ai-git workspace find --status active`,
	Args: cobra.MaximumNArgs(1),
	RunE:  runWorkspaceFind,
}

var workspacePathCmd = &cobra.Command{
	Use:   "path <name>",
	Short: "Get project path",
	Long: `Get full path of a project.

Examples:
  ai-git workspace path ai-git`,
	Args: cobra.ExactArgs(1),
	RunE:  runWorkspacePath,
}

var workspaceOverviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Workspace overview",
	Long:  `Show workspace overview with statistics.`,
	RunE:  runWorkspaceOverview,
}

var newCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new project",
	Long: `Create a new project with standard structure.

Examples:
  ai-git new my-project --type go-cli
  ai-git new my-api --type go-api
  ai-git new my-tool --type python-cli`,
	Args: cobra.ExactArgs(1),
	RunE:  runNewProject,
}

var (
	findLanguage string
	findStatus   string
	findType     string
	projectType  string
)

func init() {
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(newCmd)

	workspaceCmd.AddCommand(workspaceInitCmd)
	workspaceCmd.AddCommand(workspaceListCmd)
	workspaceCmd.AddCommand(workspaceFindCmd)
	workspaceCmd.AddCommand(workspacePathCmd)
	workspaceCmd.AddCommand(workspaceOverviewCmd)

	workspaceFindCmd.Flags().StringVar(&findLanguage, "language", "", "Filter by language")
	workspaceFindCmd.Flags().StringVar(&findStatus, "status", "", "Filter by status")
	workspaceFindCmd.Flags().StringVar(&findType, "type", "", "Filter by project type")

	newCmd.Flags().StringVar(&projectType, "type", "go-cli", "Project type (go-cli, go-api, python-cli, rust-cli, etc.)")
}

func runWorkspaceInit(cmd *cobra.Command, args []string) error {
	var workspacePath string
	if len(args) > 0 {
		workspacePath = args[0]
	} else {
		home, _ := os.UserHomeDir()
		workspacePath = filepath.Join(home, "projects")
	}

	if !filepath.IsAbs(workspacePath) {
		abs, _ := filepath.Abs(workspacePath)
		workspacePath = abs
	}

	config := &GlobalConfig{
		Workspace: workspacePath,
		Structure: WorkspaceStructure{
			Active:      "active",
			Experiments: "experiments",
			Research:    "research",
			Archive:     "archive",
		},
		Created: time.Now().Format(time.RFC3339),
	}

	if err := saveGlobalConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	dirs := []string{
		config.Structure.Active,
		config.Structure.Experiments,
		config.Structure.Research,
		config.Structure.Archive,
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(workspacePath, dir)
		os.MkdirAll(fullPath, 0755)

		langDirs := []string{"go", "python", "rust", "javascript", "java", "other"}
		for _, lang := range langDirs {
			os.MkdirAll(filepath.Join(fullPath, lang), 0755)
		}
	}

	index := &WorkspaceIndex{
		Projects: make(map[string]ProjectMeta),
		Updated:  time.Now().Format(time.RFC3339),
	}
	saveWorkspaceIndex(workspacePath, index)

	return utils.OutputSuccess(map[string]interface{}{
		"status":    "initialized",
		"workspace": workspacePath,
		"config":    configPath(),
		"message":   "Workspace initialized successfully",
	})
}

func runWorkspaceList(cmd *cobra.Command, args []string) error {
	config, err := getGlobalConfig()
	if err != nil {
		return err
	}

	index, err := getWorkspaceIndex(config.Workspace)
	if err != nil {
		return err
	}

	projects := make([]ProjectMeta, 0)
	for _, p := range index.Projects {
		projects = append(projects, p)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"workspace": config.Workspace,
		"count":     len(projects),
		"projects":  projects,
	})
}

func runWorkspaceFind(cmd *cobra.Command, args []string) error {
	config, err := getGlobalConfig()
	if err != nil {
		return err
	}

	index, err := getWorkspaceIndex(config.Workspace)
	if err != nil {
		return err
	}

	query := ""
	if len(args) > 0 {
		query = strings.ToLower(args[0])
	}

	results := make([]ProjectMeta, 0)
	for _, p := range index.Projects {
		match := true

		if query != "" {
			match = match && (strings.Contains(strings.ToLower(p.Name), query) ||
				strings.Contains(strings.ToLower(p.Description), query))
		}

		if findLanguage != "" {
			match = match && strings.EqualFold(p.Language, findLanguage)
		}

		if findStatus != "" {
			match = match && strings.EqualFold(p.Status, findStatus)
		}

		if findType != "" {
			match = match && strings.EqualFold(p.Type, findType)
		}

		if match {
			results = append(results, p)
		}
	}

	return utils.OutputSuccess(map[string]interface{}{
		"query":    query,
		"count":    len(results),
		"projects": results,
	})
}

func runWorkspacePath(cmd *cobra.Command, args []string) error {
	config, err := getGlobalConfig()
	if err != nil {
		return err
	}

	index, err := getWorkspaceIndex(config.Workspace)
	if err != nil {
		return err
	}

	name := args[0]
	project, exists := index.Projects[name]
	if !exists {
		return fmt.Errorf("project not found: %s", name)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"name": name,
		"path": project.Path,
		"full": filepath.Join(config.Workspace, project.Path),
	})
}

func runWorkspaceOverview(cmd *cobra.Command, args []string) error {
	config, err := getGlobalConfig()
	if err != nil {
		return err
	}

	index, err := getWorkspaceIndex(config.Workspace)
	if err != nil {
		return err
	}

	stats := map[string]int{
		"total":    len(index.Projects),
		"active":   0,
		"inactive": 0,
	}

	byLanguage := make(map[string]int)
	byType := make(map[string]int)

	for _, p := range index.Projects {
		if p.Status == "active" {
			stats["active"]++
		} else {
			stats["inactive"]++
		}
		byLanguage[p.Language]++
		byType[p.Type]++
	}

	return utils.OutputSuccess(map[string]interface{}{
		"workspace":   config.Workspace,
		"stats":       stats,
		"by_language": byLanguage,
		"by_type":     byType,
		"updated":     index.Updated,
	})
}

func runNewProject(cmd *cobra.Command, args []string) error {
	config, err := getGlobalConfig()
	if err != nil {
		return err
	}

	name := args[0]

	language := "other"
	status := "active"

	parts := strings.Split(projectType, "-")
	if len(parts) > 0 {
		language = parts[0]
	}

	switch language {
	case "go":
		language = "go"
	case "python":
		language = "python"
	case "rust":
		language = "rust"
	case "js", "javascript", "typescript":
		language = "javascript"
	case "java":
		language = "java"
	default:
		language = "other"
	}

	var categoryDir string
	switch status {
	case "active":
		categoryDir = config.Structure.Active
	case "experiment":
		categoryDir = config.Structure.Experiments
	case "research":
		categoryDir = config.Structure.Research
	default:
		categoryDir = config.Structure.Active
	}

	projectPath := filepath.Join(categoryDir, language, name)
	fullPath := filepath.Join(config.Workspace, projectPath)

	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("project already exists: %s", fullPath)
	}

	if err := createProjectStructure(fullPath, projectType, name); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	index, err := getWorkspaceIndex(config.Workspace)
	if err != nil {
		return err
	}

	index.Projects[name] = ProjectMeta{
		Name:        name,
		Path:        projectPath,
		Type:        projectType,
		Language:    language,
		Status:      status,
		Description: "",
		Created:     time.Now().Format(time.RFC3339),
		Modified:    time.Now().Format(time.RFC3339),
	}

	if err := saveWorkspaceIndex(config.Workspace, index); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"status":   "created",
		"name":     name,
		"path":     fullPath,
		"type":     projectType,
		"language": language,
		"message":  "Project created successfully",
	})
}
