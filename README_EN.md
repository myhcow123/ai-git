# AI-Git

**Text Read/Write System Designed for AI Models**

**Version: 1.0.0**

[English](README_EN.md) | [中文](README_CN.md)

AI-Git is a powerful text read/write system that provides unified text operation capabilities for AI models. It's not just a code indexing tool, but a complete workspace for models, including code management, task management, knowledge base, and document processing.

## Core Philosophy

```
Everything is text:
├── Code = Text
├── Markdown = Text
├── Config = Text
└── Data = Text

Core operations:
├── Read: Parse text → Extract structure
└── Write: Generate text → Modify content
```

## Features

### 🔍 Code Analysis
- **Multi-language Support**: Go, Python, JavaScript, TypeScript, Rust, Java, C/C++
- **Symbol Extraction**: Functions, classes, variables, interfaces, structs, etc.
- **Dependency Analysis**: Code dependency and impact analysis
- **PageRank Analysis**: Code importance evaluation

### 📝 Document Processing
- **Markdown Parsing**: Headers, sections, code blocks, checklists
- **Structured Indexing**: Fast document content queries
- **Auto Monitoring**: Automatic index updates on file changes

### 📋 Task Management
- **Task Creation**: Create and break down tasks
- **Progress Tracking**: Real-time task progress updates
- **Execution Logs**: Record every operation step
- **History Archive**: Automatic archiving of completed tasks

### 📚 Knowledge Base
- **Knowledge Notes**: Create and manage knowledge notes
- **Tag System**: Flexible tag categorization
- **Code Association**: Bidirectional linking between notes and code
- **Search & Export**: Fast search and export of knowledge

### 🚀 Performance Optimization
- **Memory Cache**: O(1) fast queries
- **Real-time Indexing**: Automatic updates on file changes
- **Concurrent Processing**: Parallel parsing for performance
- **REST API**: Remote access support

## Installation

```bash
# Clone repository
git clone https://github.com/mychow/ai-git.git
cd ai-git

# Build
go build -o ai-git .

# Install to system
sudo mv ai-git /usr/local/bin/
```

## Quick Start

### 0. Start Background Service (Required)

**Important: You must start the background service before using AI-Git!**

```bash
# Start background service
ai-git daemon start

# Check service status
ai-git daemon status

# Stop service
ai-git daemon stop
```

Once the background service is started, the index will be kept in memory. All CLI commands will access it via API for approximately 16x performance improvement.

### 1. Initialize Project

```bash
# Initialize in project directory
ai-git init

# View project overview
ai-git overview
```

### 2. Query Code

```bash
# Search symbols
ai-git search "handler"

# Query symbol details
ai-git symbol "Process"

# Read code
ai-git read main.go:10-20
ai-git read Process
```

### 3. Modify Code

```bash
# Insert code
ai-git insert main.go:10 --code "// comment"

# Replace code
ai-git replace Process --with "func Process() { ... }"

# Delete code
ai-git delete OldFunction
```

### 4. Task Management

```bash
# Create task
ai-git task create "Implement user authentication" --priority high

# Update progress
ai-git task update task-001 --progress 50

# Complete task
ai-git task complete task-001

# View task list
ai-git task list
```

### 5. Knowledge Management

```bash
# Create note
ai-git note create "Go Concurrency Patterns" --tags "go,concurrency"

# Search knowledge
ai-git note search "concurrency"

# Link code
ai-git note link note-001 --code "worker.go:Process"

# Export knowledge
ai-git note export --format markdown
```

### 6. Start Service

```bash
# Start API service
ai-git web --port 8080

# Access API
curl http://localhost:8080/api/v1/search?q=handler
```

## 📁 Configuration Files

AI-Git uses hidden directories to store configuration and data:

### Global Configuration (User Level)
```
~/.ai-git/              # Global config directory
├── config.json         # Global configuration
├── daemon.json         # Background service info
├── daemon.pid          # Background service process ID
├── tasks/              # Task data
├── notes/              # Note data
├── queries/            # Query templates
└── plugins/            # Plugin directory
```

### Project Configuration (Project Level)
```
project/.ai-git/        # Project config directory
└── ai-git.db           # Project database
```

### View Configuration
```bash
# View global configuration
cat ~/.ai-git/config.json

# View project database
ls -la .ai-git/
```

**Note: Configuration files are optional. AI-Git can work with default settings!**

## Command Overview

### Query Commands
| Command | Description |
|---------|-------------|
| `search` | Search symbols |
| `symbol` | Query symbol details |
| `read` | Read code or files |
| `overview` | Project overview |
| `status` | Project status |

### Modify Commands
| Command | Description |
|---------|-------------|
| `modify` | Modify code |
| `insert` | Insert code |
| `replace` | Replace code |
| `delete` | Delete code |
| `refactor` | Refactor code |
| `batch` | Batch modifications |

### Analysis Commands
| Command | Description |
|---------|-------------|
| `analyze` | PageRank analysis |
| `deps` | Dependency analysis |
| `impact` | Impact analysis |
| `quality` | Code quality assessment |
| `pattern` | Design pattern recognition |

### Task Commands
| Command | Description |
|---------|-------------|
| `task create` | Create task |
| `task list` | List tasks |
| `task update` | Update task |
| `task complete` | Complete task |
| `task log` | Record execution log |

### Note Commands
| Command | Description |
|---------|-------------|
| `note create` | Create note |
| `note list` | List notes |
| `note search` | Search notes |
| `note link` | Link code |
| `note export` | Export notes |

### Project Commands
| Command | Description |
|---------|-------------|
| `init` | Initialize project |
| `workspace` | Workspace management |
| `web` | Start API service |
| `plugin` | Plugin management |

## Architecture

```
┌─────────────────────────────────────────────┐
│         AI-Git: Text Read/Write System       │
│                                             │
│  Code Layer                                 │
│  ├── 7 language parsers                     │
│  ├── Symbol extraction & indexing           │
│  ├── Dependency & impact analysis           │
│  └── Modification & refactoring             │
│                                             │
│  Task Layer                                 │
│  ├── Task creation & breakdown              │
│  ├── Progress tracking                      │
│  ├── Execution logs                         │
│  └── History archive                        │
│                                             │
│  Knowledge Layer                            │
│  ├── Knowledge notes                        │
│  ├── Context management                     │
│  ├── Code association                       │
│  └── Search & export                        │
│                                             │
│  Document Layer                             │
│  ├── Markdown parsing                       │
│  ├── Header & section indexing              │
│  ├── Checklist tracking                     │
│  └── Code block extraction                  │
│                                             │
│  Service Layer                              │
│  ├── REST API                               │
│  ├── File watching                          │
│  ├── Real-time indexing                     │
│  └── Memory cache                           │
└─────────────────────────────────────────────┘
```

## Project Structure

```
ai-git/
├── cmd/                    # CLI tools
│   ├── root.go            # Root command
│   ├── task.go            # Task management commands
│   ├── note.go            # Note management commands
│   └── ...                # Other commands
│
├── internal/              # Internal packages
│   ├── parser/           # Parsers
│   │   ├── parser.go     # Core parser
│   │   ├── parser_go.go  # Go parser
│   │   ├── parser_markdown.go # Markdown parser
│   │   └── ...           # Other language parsers
│   │
│   ├── task/             # Task management
│   ├── note/             # Knowledge base management
│   ├── storage/          # Storage layer
│   ├── watcher/          # File watching
│   └── api/              # REST API
│
├── pkg/                   # Public packages
│   ├── types/            # Type definitions
│   └── utils/            # Utility functions
│
└── main.go               # Entry point
```

## API Documentation

### Search
```
GET /api/v1/search?q=handler
```

### Symbol Query
```
GET /api/v1/symbol/:name
```

### Project Overview
```
GET /api/v1/overview
```

### Task Management
```
POST /api/v1/projects
GET /api/v1/projects
DELETE /api/v1/projects/:id
```

### Watcher Status
```
GET /api/v1/status/watcher
```

## Development

### Build
```bash
make build
```

### Test
```bash
make test
```

### Clean
```bash
make clean
```

## Tech Stack

- **Language**: Go 1.21+
- **Database**: BoltDB (persistence) + Memory cache
- **Parser**: Custom parsers (7 languages)
- **API**: REST API
- **Monitoring**: fsnotify

## Statistics

- **Lines of Code**: ~11,745
- **Files**: 55 Go files
- **Languages**: 7 (Go, Python, JavaScript, TypeScript, Rust, Java, C/C++)
- **Commands**: 40+ commands

## Contributing

Contributions, bug reports, and suggestions are welcome!

## License

MIT License

## Author

GLM-5

---

**AI-Git: Text Read/Write System Designed for AI Models**
