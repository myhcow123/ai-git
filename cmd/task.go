package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/mychow/ai-git/internal/task"
	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long: `Manage tasks for model-driven development.

Examples:
  ai-git task create "实现用户认证" --priority high
  ai-git task list --status active
  ai-git task update task-001 --progress 50
  ai-git task complete task-001`,
}

var taskCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskCreate,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	RunE:  runTaskList,
}

var taskGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get task details",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskGet,
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskUpdate,
}

var taskCompleteCmd = &cobra.Command{
	Use:   "complete <id>",
	Short: "Mark task as completed",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskComplete,
}

var taskLogCmd = &cobra.Command{
	Use:   "log <id>",
	Short: "Add log entry to task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskLog,
}

var taskArchiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive completed task",
	Args:  cobra.ExactArgs(1),
	RunE:  runTaskArchive,
}

var (
	taskPriority  string
	taskStatus    string
	taskProgress  int
	taskTags      string
	taskSubtasks  string
	taskLogStep   string
	taskLogTool   string
	taskLogResult string
)

func init() {
	rootCmd.AddCommand(taskCmd)

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskCompleteCmd)
	taskCmd.AddCommand(taskLogCmd)
	taskCmd.AddCommand(taskArchiveCmd)

	taskCreateCmd.Flags().StringVar(&taskPriority, "priority", "medium", "Task priority (low, medium, high)")
	taskCreateCmd.Flags().StringVar(&taskTags, "tags", "", "Comma-separated tags")
	taskCreateCmd.Flags().StringVar(&taskSubtasks, "subtasks", "", "JSON array of subtasks")

	taskListCmd.Flags().StringVar(&taskStatus, "status", "", "Filter by status")

	taskUpdateCmd.Flags().StringVar(&taskStatus, "status", "", "Update status")
	taskUpdateCmd.Flags().IntVar(&taskProgress, "progress", -1, "Update progress (0-100)")
	taskUpdateCmd.Flags().StringVar(&taskPriority, "priority", "", "Update priority")

	taskLogCmd.Flags().StringVar(&taskLogStep, "step", "", "Step description")
	taskLogCmd.Flags().StringVar(&taskLogTool, "tool", "", "Tool used")
	taskLogCmd.Flags().StringVar(&taskLogResult, "result", "", "Result")
}

func getTaskManager() (*task.Manager, error) {
	home, _ := os.UserHomeDir()
	tasksDir := filepath.Join(home, ".ai-git", "tasks")

	mgr := task.NewManager(tasksDir)
	if err := mgr.Init(); err != nil {
		return nil, err
	}

	return mgr, nil
}

func runTaskCreate(cmd *cobra.Command, args []string) error {
	mgr, err := getTaskManager()
	if err != nil {
		return err
	}

	name := args[0]

	var opts []task.TaskOption

	switch taskPriority {
	case "high":
		opts = append(opts, task.WithPriority(task.TaskPriorityHigh))
	case "low":
		opts = append(opts, task.WithPriority(task.TaskPriorityLow))
	default:
		opts = append(opts, task.WithPriority(task.TaskPriorityMedium))
	}

	if taskTags != "" {
		tags := strings.Split(taskTags, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
		opts = append(opts, task.WithTags(tags))
	}

	if taskSubtasks != "" {
		var subtasks []task.SubTask
		if err := json.Unmarshal([]byte(taskSubtasks), &subtasks); err == nil {
			opts = append(opts, task.WithSubTasks(subtasks))
		}
	}

	t, err := mgr.Create(name, opts...)
	if err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":        t.ID,
		"name":      t.Name,
		"status":    t.Status,
		"priority":  t.Priority,
		"message":   "Task created successfully",
	})
}

func runTaskList(cmd *cobra.Command, args []string) error {
	mgr, err := getTaskManager()
	if err != nil {
		return err
	}

	var status task.TaskStatus
	if taskStatus != "" {
		status = task.TaskStatus(taskStatus)
	}

	tasks, err := mgr.List(status)
	if err != nil {
		return err
	}

	result := make([]map[string]interface{}, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, map[string]interface{}{
			"id":       t.ID,
			"name":     t.Name,
			"status":   t.Status,
			"priority": t.Priority,
			"progress": t.Progress,
		})
	}

	return utils.OutputSuccess(map[string]interface{}{
		"count": len(result),
		"tasks": result,
	})
}

func runTaskGet(cmd *cobra.Command, args []string) error {
	mgr, err := getTaskManager()
	if err != nil {
		return err
	}

	id := args[0]
	t, err := mgr.Get(id)
	if err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":          t.ID,
		"name":        t.Name,
		"description": t.Description,
		"status":      t.Status,
		"priority":    t.Priority,
		"progress":    t.Progress,
		"tags":        t.Tags,
		"subtasks":    t.SubTasks,
		"logs":        t.Logs,
		"created_at":  t.CreatedAt,
		"updated_at":  t.UpdatedAt,
	})
}

func runTaskUpdate(cmd *cobra.Command, args []string) error {
	mgr, err := getTaskManager()
	if err != nil {
		return err
	}

	id := args[0]

	var opts []task.TaskOption

	if taskStatus != "" {
		opts = append(opts, task.WithStatus(task.TaskStatus(taskStatus)))
	}

	if taskProgress >= 0 {
		opts = append(opts, task.WithProgress(taskProgress))
	}

	if taskPriority != "" {
		switch taskPriority {
		case "high":
			opts = append(opts, task.WithPriority(task.TaskPriorityHigh))
		case "low":
			opts = append(opts, task.WithPriority(task.TaskPriorityLow))
		default:
			opts = append(opts, task.WithPriority(task.TaskPriorityMedium))
		}
	}

	t, err := mgr.Update(id, opts...)
	if err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":       t.ID,
		"status":   t.Status,
		"progress": t.Progress,
		"message":  "Task updated successfully",
	})
}

func runTaskComplete(cmd *cobra.Command, args []string) error {
	mgr, err := getTaskManager()
	if err != nil {
		return err
	}

	id := args[0]
	t, err := mgr.Complete(id)
	if err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":           t.ID,
		"status":       t.Status,
		"progress":     t.Progress,
		"completed_at": t.CompletedAt,
		"message":      "Task completed successfully",
	})
}

func runTaskLog(cmd *cobra.Command, args []string) error {
	mgr, err := getTaskManager()
	if err != nil {
		return err
	}

	id := args[0]

	log := task.TaskLog{
		Step:   taskLogStep,
		Tool:   taskLogTool,
		Output: taskLogResult,
		Status: "success",
	}

	if err := mgr.AddLog(id, log); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":      id,
		"message": "Log added successfully",
	})
}

func runTaskArchive(cmd *cobra.Command, args []string) error {
	mgr, err := getTaskManager()
	if err != nil {
		return err
	}

	id := args[0]

	if err := mgr.Archive(id); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"id":      id,
		"message": "Task archived successfully",
	})
}
