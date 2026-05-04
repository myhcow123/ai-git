# Task Manager

A command-line task management system implemented in Go, featuring CRUD operations and file-based persistence.

## Features

- **Create Tasks**: Add new tasks with titles, descriptions, and priorities
- **Read Tasks**: View individual tasks or list all tasks
- **Update Tasks**: Modify task properties including status and priority
- **Delete Tasks**: Remove tasks from the system
- **Complete Tasks**: Mark tasks as completed
- **Persistence**: Tasks are stored in a JSON file for data persistence
- **Filtering**: List tasks by status (pending, in_progress, completed)

## Installation

### Prerequisites
- Go 1.16 or higher
- Git

### Steps

1. Clone or download the project
2. Navigate to the project directory
3. Build the application:

```bash
go build -o taskmanager .
```

4. Run the application:

```bash
./taskmanager
```

## Usage

### Commands

#### Add a new task
```bash
./taskmanager add "Task Title" -d "Task description" -p high
```

Options:
- `-d`: Task description (optional)
- `-p`: Task priority (low, medium, high) - defaults to medium

#### List all tasks
```bash
./taskmanager list
```

Options:
- `-s`: Filter by status (pending, in_progress, completed)

#### Show task details
```bash
./taskmanager show <task-id>
```

#### Update a task
```bash
./taskmanager update <task-id> -t "New title" -d "New description" -s in_progress -p high
```

Options:
- `-t`: New title
- `-d`: New description
- `-s`: New status (pending, in_progress, completed)
- `-p`: New priority (low, medium, high)

#### Mark task as completed
```bash
./taskmanager complete <task-id>
```

#### Delete a task
```bash
./taskmanager delete <task-id>
```

#### Show help
```bash
./taskmanager help
```

### Examples

```bash
# Create a high priority task
./taskmanager add "Complete project documentation" -d "Write comprehensive docs" -p high

# List all pending tasks
./taskmanager list -s pending

# Mark a task as in progress
./taskmanager update abc123-def456-ghi789 -s in_progress

# Complete a task
./taskmanager complete abc123-def456-ghi789

# Delete a task
./taskmanager delete abc123-def456-ghi789
```

## Data Storage

Tasks are stored in a JSON file named `tasks.json` in the same directory as the executable. The file is automatically created when the first task is added.

### Data Structure

```json
[
  {
    "id": "uuid-string",
    "title": "Task title",
    "description": "Task description",
    "status": "pending|in_progress|completed",
    "priority": "low|medium|high",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
]
```

## Running Tests

To run the test suite:

```bash
go test -v -cover
```

To view coverage details:

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Project Structure

```
.
├── main.go          # CLI interface and command handlers
├── task.go          # Task data structures and validation
├── storage.go       # File-based storage layer
├── manager.go       # Task manager with CRUD operations
├── manager_test.go  # Manager unit tests
├── storage_test.go  # Storage unit tests
├── main_test.go     # CLI and integration tests
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
├── tasks.json       # Task data file (created automatically)
└── README.md        # This file
```

## API Documentation

### Data Types

#### Status
- `pending`: Task is waiting to be processed
- `in_progress`: Task is currently being worked on
- `completed`: Task has been completed

#### Priority
- `low`: Low priority task
- `medium`: Medium priority task
- `high`: High priority task

### Task Structure

```go
type Task struct {
    ID          string    // Unique identifier (UUID)
    Title       string    // Task title (required)
    Description string    // Task description (optional)
    Status      Status    // Current status
    Priority    Priority  // Task priority
    CreatedAt   time.Time // Creation timestamp
    UpdatedAt   time.Time // Last update timestamp
}
```

### Error Handling

The system returns meaningful error messages for:
- Empty or invalid task titles
- Non-existent task IDs
- Invalid status or priority values
- File I/O errors

### Task Manager Methods

| Method | Description | Parameters |
|--------|-------------|------------|
| Create | Creates a new task | title, description, priority |
| Get | Retrieves a task by ID | id |
| List | Lists all tasks, optionally filtered | status (optional) |
| Update | Updates a task's properties | id, updates map |
| Delete | Removes a task | id |
| Complete | Marks a task as completed | id |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests to ensure everything passes
5. Submit a pull request

## License

This project is open source and available under the MIT License.
