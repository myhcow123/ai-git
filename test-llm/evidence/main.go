package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// CLI commands
const (
	cmdAdd      = "add"
	cmdList     = "list"
	cmdShow     = "show"
	cmdUpdate   = "update"
	cmdComplete = "complete"
	cmdDelete   = "delete"
	cmdHelp     = "help"
)

func printUsage() {
	fmt.Println(`Task Manager - A command-line task management system

Usage:
  taskmanager <command> [options]

Commands:
  add <title> [-d description] [-p priority]    Create a new task
  list [-s status]                                List all tasks
  show <id>                                       Show task details
  update <id> [-t title] [-d description] [-s status] [-p priority]  Update a task
  complete <id>                                   Mark task as completed
  delete <id>                                     Delete a task
  help                                           Show this help message

Options:
  -d description    Task description
  -p priority       Task priority (low, medium, high)
  -s status         Filter by status (pending, in_progress, completed)
  -t title          New task title

Examples:
  taskmanager add "Complete project" -d "Finish the task manager" -p high
  taskmanager list -s pending
  taskmanager show abc123
  taskmanager update abc123 -s in_progress
  taskmanager complete abc123
  taskmanager delete abc123`)
}

func main() {
	// Parse command line arguments
	flag.CommandLine.Parse(os.Args[1:])

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Initialize storage and manager
	storage, err := NewTaskStorage("tasks.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	manager := NewTaskManager(storage)

	// Get command
	cmd := os.Args[1]

	switch cmd {
	case cmdAdd:
		handleAdd(manager, os.Args[2:])
	case cmdList:
		handleList(manager, os.Args[2:])
	case cmdShow:
		handleShow(manager, os.Args[2:])
	case cmdUpdate:
		handleUpdate(manager, os.Args[2:])
	case cmdComplete:
		handleComplete(manager, os.Args[2:])
	case cmdDelete:
		handleDelete(manager, os.Args[2:])
	case cmdHelp, "-h", "--help":
		printUsage()
	default:
		if !flag.Parsed() {
			flag.CommandLine.PrintDefaults()
		}
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

// parseFlags parses command flags and returns remaining args
func parseFlags(args []string) (map[string]string, []string) {
	flags := make(map[string]string)
	remaining := []string{}

	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// This is a flag
			flagName := arg[1:]
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags[flagName] = args[i+1]
				i += 2
			} else {
				flags[flagName] = ""
				i++
			}
		} else {
			remaining = append(remaining, arg)
			i++
		}
	}

	return flags, remaining
}

func handleAdd(manager *TaskManager, args []string) {
	flags, remaining := parseFlags(args)

	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Title is required")
		fmt.Fprintln(os.Stderr, "Usage: taskmanager add <title> [-d description] [-p priority]")
		os.Exit(1)
	}

	title := remaining[0]
	description := flags["d"]
	priorityStr := flags["p"]

	// Default priority to medium if not specified
	priority := PriorityMedium
	if priorityStr != "" {
		if !IsValidPriority(priorityStr) {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high\n", priorityStr)
			os.Exit(1)
		}
		priority = Priority(priorityStr)
	}

	task, err := manager.Create(title, description, priority)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task created successfully!\n")
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("Title: %s\n", task.Title)
	fmt.Printf("Status: %s\n", task.Status)
	fmt.Printf("Priority: %s\n", task.Priority)
}

func handleList(manager *TaskManager, args []string) {
	flags, _ := parseFlags(args)

	statusStr := flags["s"]

	var status Status
	if statusStr != "" {
		if !IsValidStatus(statusStr) {
			fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: pending, in_progress, completed\n", statusStr)
			os.Exit(1)
		}
		status = Status(statusStr)
	}

	tasks, err := manager.List(status)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing tasks: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Printf("\n%-36s %-20s %-12s %-8s\n", "ID", "TITLE", "STATUS", "PRIORITY")
	fmt.Println(strings.Repeat("-", 80))

	for _, task := range tasks {
		title := task.Title
		if len(title) > 18 {
			title = title[:15] + "..."
		}
		fmt.Printf("%-36s %-20s %-12s %-8s\n", task.ID, title, task.Status, task.Priority)
	}

	fmt.Printf("\nTotal: %d task(s)\n", len(tasks))
}

func handleShow(manager *TaskManager, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required")
		fmt.Fprintln(os.Stderr, "Usage: taskmanager show <id>")
		os.Exit(1)
	}

	id := args[0]
	task, err := manager.Get(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("═══════════════════════════════════════")
	fmt.Println("              TASK DETAILS              ")
	fmt.Println("═══════════════════════════════════════")
	fmt.Printf("ID:          %s\n", task.ID)
	fmt.Printf("Title:       %s\n", task.Title)
	fmt.Printf("Description: %s\n", task.Description)
	fmt.Printf("Status:      %s\n", task.Status)
	fmt.Printf("Priority:    %s\n", task.Priority)
	fmt.Printf("Created:     %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:     %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println("═══════════════════════════════════════")
}

func handleUpdate(manager *TaskManager, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required")
		fmt.Fprintln(os.Stderr, "Usage: taskmanager update <id> [-t title] [-d description] [-s status] [-p priority]")
		os.Exit(1)
	}

	id := args[0]
	flags, remaining := parseFlags(args[1:])

	updates := make(map[string]interface{})

	// Check for positional title (if not using -t flag)
	if len(remaining) > 0 && flags["t"] == "" {
		updates["title"] = remaining[0]
	}

	if title := flags["t"]; title != "" {
		updates["title"] = title
	}

	if description := flags["d"]; description != "" {
		updates["description"] = description
	}

	if status := flags["s"]; status != "" {
		if !IsValidStatus(status) {
			fmt.Fprintf(os.Stderr, "Error: Invalid status '%s'. Valid values: pending, in_progress, completed\n", status)
			os.Exit(1)
		}
		updates["status"] = status
	}

	if priority := flags["p"]; priority != "" {
		if !IsValidPriority(priority) {
			fmt.Fprintf(os.Stderr, "Error: Invalid priority '%s'. Valid values: low, medium, high\n", priority)
			os.Exit(1)
		}
		updates["priority"] = priority
	}

	if len(updates) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No updates provided")
		fmt.Fprintln(os.Stderr, "Usage: taskmanager update <id> [-t title] [-d description] [-s status] [-p priority]")
		os.Exit(1)
	}

	task, err := manager.Update(id, updates)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error updating task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task updated successfully!\n")
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("Title: %s\n", task.Title)
	fmt.Printf("Status: %s\n", task.Status)
	fmt.Printf("Priority: %s\n", task.Priority)
}

func handleComplete(manager *TaskManager, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required")
		fmt.Fprintln(os.Stderr, "Usage: taskmanager complete <id>")
		os.Exit(1)
	}

	id := args[0]
	task, err := manager.Complete(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error completing task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task marked as completed!\n")
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("Title: %s\n", task.Title)
}

func handleDelete(manager *TaskManager, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: Task ID is required")
		fmt.Fprintln(os.Stderr, "Usage: taskmanager delete <id>")
		os.Exit(1)
	}

	id := args[0]
	if err := manager.Delete(id); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task deleted successfully!\n")
}
