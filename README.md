# AI-Git

**Text Read/Write System Designed for AI Models**

**Version: 1.0.0**

[English](README_EN.md) | [中文](README_CN.md)

AI-Git is a powerful text read/write system that provides unified text operation capabilities for AI models. It's not just a code indexing tool, but a complete workspace for models, including code management, task management, knowledge base, and document processing.

## 💡 Core Advantage: Massive Token Savings

**The Problem**: Traditional LLM-based code assistants read entire files on every interaction, causing token usage to accumulate rapidly across multiple iterations.

**The Solution**: AI-Git's symbol-based architecture fundamentally changes this:

- **Symbol-Level Access**: Query specific functions, classes, or symbols instead of entire files
- **Structured Index**: Pre-parsed code structure enables precise, targeted queries
- **Minimal Context**: Only relevant code snippets are included in LLM context

**Real-World Evidence**: In our [LLM testing report](TEST_REPORT.md), a complex multi-file project implementation required 53 iterations but only consumed ~38K tokens total. Traditional approaches would require reading entire codebase on each iteration, leading to exponentially higher token usage.

**Why It Matters**:
- ✅ **Cost Efficiency**: Dramatically reduce API costs for LLM interactions
- ✅ **Speed**: Smaller context means faster response times
- ✅ **Scalability**: Handle large codebases without token limits
- ✅ **Precision**: Work with exact symbols, not entire files

See [TEST_REPORT.md](TEST_REPORT.md) for detailed analysis and [SUMMARY.md](SUMMARY.md) for test summary.

## 🎯 Key Highlights

### ⚡ Rapid Codebase Parsing
- **Multi-language Support**: Parse 7 languages (Go, Python, JavaScript, TypeScript, Rust, Java, C/C++)
- **Fast Indexing**: Concurrent parsing with memory cache for O(1) queries
- **Real-time Updates**: Automatic re-indexing on file changes via file watcher
- **Symbol Extraction**: Extract functions, classes, variables, interfaces, structs, etc.
- **Dependency Graph**: Build and analyze code dependency relationships

### 🔍 Advanced Symbol System
- **Symbol Search**: Fast fuzzy search across all symbols
- **Symbol Details**: Get complete information about any symbol (type, location, signature, documentation)
- **Usage Tracking**: Find all usages of a symbol across the codebase
- **Dependency Analysis**: Show what a symbol depends on and what depends on it
- **Impact Analysis**: Analyze the impact of modifying a symbol
- **PageRank Analysis**: Identify the most important symbols in your codebase

### 🛠️ Comprehensive Tool Suite
- **40+ CLI Commands**: Query, modify, analyze, and manage your codebase
- **REST API**: Full-featured API for remote access and integration
- **Batch Operations**: Execute modifications on multiple files simultaneously
- **Undo System**: Safely undo recent modifications with full history tracking
- **Workspace Management**: Manage multiple projects with isolated databases
- **Background Service**: Daemon mode for persistent indexing and fast queries

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
- **Quality Assessment**: Complexity, testability, maintainability, security analysis
- **Pattern Recognition**: Identify design patterns in your code
- **Intent Inference**: Understand the purpose of code symbols

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
git clone https://github.com/myhcow123/ai-git.git
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

# Find all usages of a symbol
ai-git usages functionName

# Show symbol dependencies
ai-git deps Process

# Analyze modification impact
ai-git impact Process
```

### 3. Modify Code

```bash
# Insert code
ai-git insert main.go:10 --code "// comment"

# Replace code
ai-git replace Process --with "func Process() { ... }"

# Delete code
ai-git delete OldFunction

# Batch modifications
ai-git batch modifications.json

# Undo recent modifications
ai-git undo          # Undo last modification
ai-git undo 3        # Undo last 3 modifications
ai-git undo --list   # List recent modifications
```

### 4. Analyze Code

```bash
# PageRank analysis - find most important symbols
ai-git analyze --top 10

# Code quality assessment
ai-git quality Process

# Design pattern recognition
ai-git pattern Handler

# Intent inference
ai-git intent main

# Dependency analysis with depth
ai-git deps Process --depth 2
```

### 5. Task Management

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

### 6. Knowledge Management

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

### 7. Query System (AQL)

```bash
# List all predefined queries
ai-git query list

# Show query details
ai-git query show find_entry_points

# Search queries
ai-git query search "api"

# Execute a query
ai-git query exec find_high_complexity

# List query categories
ai-git query categories
```

### 8. Workspace Management

```bash
# Create a new workspace
ai-git workspace create myproject

# List all workspaces
ai-git workspace list

# Switch workspace
ai-git workspace use myproject

# Configure workspace
ai-git workspace config set key value
```

### 9. Start Service

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

## 📚 Complete Command Reference

### Query Commands
| Command | Description | Example |
|---------|-------------|---------|
| `search` | Search symbols with fuzzy matching | `ai-git search "handler"` |
| `symbol` | Get detailed symbol information | `ai-git symbol Process` |
| `read` | Read code or files | `ai-git read main.go:10-20` |
| `overview` | Get project overview | `ai-git overview` |
| `status` | Check project status | `ai-git status` |
| `usages` | Find all usages of a symbol | `ai-git usages functionName` |

### Modify Commands
| Command | Description | Example |
|---------|-------------|---------|
| `modify` | Modify code with AI assistance | `ai-git modify Process` |
| `insert` | Insert code at specific location | `ai-git insert main.go:10 --code "..."` |
| `replace` | Replace code | `ai-git replace Process --with "..."` |
| `delete` | Delete code | `ai-git delete OldFunction` |
| `refactor` | Refactor code | `ai-git refactor Process` |
| `batch` | Execute batch modifications | `ai-git batch mods.json` |
| `undo` | Undo recent modifications | `ai-git undo 3` |

### Analysis Commands
| Command | Description | Example |
|---------|-------------|---------|
| `analyze` | PageRank analysis for code importance | `ai-git analyze --top 10` |
| `deps` | Show symbol dependencies | `ai-git deps Process --depth 2` |
| `impact` | Analyze modification impact | `ai-git impact Process` |
| `quality` | Assess code quality | `ai-git quality Process` |
| `pattern` | Recognize design patterns | `ai-git pattern Handler` |
| `intent` | Infer code intent | `ai-git intent main` |

### Query System (AQL)
| Command | Description | Example |
|---------|-------------|---------|
| `query list` | List all queries | `ai-git query list` |
| `query show` | Show query details | `ai-git query show find_entry_points` |
| `query search` | Search queries | `ai-git query search "api"` |
| `query exec` | Execute a query | `ai-git query exec find_high_complexity` |
| `query categories` | List categories | `ai-git query categories` |

### Task Commands
| Command | Description | Example |
|---------|-------------|---------|
| `task create` | Create a new task | `ai-git task create "..." --priority high` |
| `task list` | List all tasks | `ai-git task list` |
| `task update` | Update task progress | `ai-git task update task-001 --progress 50` |
| `task complete` | Mark task as complete | `ai-git task complete task-001` |
| `task log` | Record execution log | `ai-git task log task-001 "..."` |

### Note Commands
| Command | Description | Example |
|---------|-------------|---------|
| `note create` | Create a knowledge note | `ai-git note create "..." --tags "go"` |
| `note list` | List all notes | `ai-git note list` |
| `note search` | Search notes | `ai-git note search "concurrency"` |
| `note link` | Link note to code | `ai-git note link note-001 --code "..."` |
| `note export` | Export notes | `ai-git note export --format markdown` |

### Workspace Commands
| Command | Description | Example |
|---------|-------------|---------|
| `workspace create` | Create a workspace | `ai-git workspace create myproject` |
| `workspace list` | List workspaces | `ai-git workspace list` |
| `workspace use` | Switch workspace | `ai-git workspace use myproject` |
| `workspace config` | Configure workspace | `ai-git workspace config set key value` |
| `workspace index` | Index workspace | `ai-git workspace index` |

### Service Commands
| Command | Description | Example |
|---------|-------------|---------|
| `daemon start` | Start background service | `ai-git daemon start` |
| `daemon stop` | Stop background service | `ai-git daemon stop` |
| `daemon status` | Check service status | `ai-git daemon status` |
| `web` | Start REST API server | `ai-git web --port 8080` |

### Other Commands
| Command | Description | Example |
|---------|-------------|---------|
| `init` | Initialize project | `ai-git init` |
| `version` | Show version | `ai-git version` |
| `edit` | Edit code with editor | `ai-git edit main.go` |
| `explain` | Explain code | `ai-git explain Process` |
| `test` | Run tests | `ai-git test` |
| `deploy` | Deploy code | `ai-git deploy` |

## 🎯 AQL Query Library

AI-Git includes a powerful query system with predefined templates:

### Available Query Categories

1. **Exploration**: Find functions, classes, and entry points
2. **Analysis**: Analyze code structure and patterns
3. **Quality**: Find high complexity code and technical debt
4. **Testing**: Locate test code and coverage gaps
5. **API**: Find API endpoints and handlers
6. **Database**: Locate database operations
7. **Security**: Identify security-sensitive code
8. **Utilities**: Find utility functions and helpers
9. **Architecture**: Analyze interfaces and abstractions

### Example Queries

```bash
# Find all entry points (main functions, handlers)
ai-git query exec find_entry_points

# Find high complexity functions
ai-git query exec find_high_complexity

# Find API endpoints
ai-git query exec find_api_endpoints

# Find security-sensitive code
ai-git query exec find_security_sensitive

# Find test code
ai-git query exec find_test_code
```

## Architecture

```
┌─────────────────────────────────────────────┐
│         AI-Git: Text Read/Write System       │
│                                             │
│  Code Layer                                 │
│  ├── 7 language parsers                     │
│  ├── Symbol extraction & indexing           │
│  ├── Dependency & impact analysis           │
│  ├── Modification & refactoring             │
│  ├── Quality assessment & patterns          │
│  └── Usage tracking & undo system           │
│                                             │
│  Query Layer (AQL)                          │
│  ├── Predefined query templates             │
│  ├── Query execution engine                 │
│  ├── Category management                    │
│  └── Custom query support                   │
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
│  ├── Memory cache                           │
│  └── Background daemon                      │
└─────────────────────────────────────────────┘
```

## Project Structure

```
ai-git/
├── cmd/                    # CLI tools
│   ├── root.go            # Root command
│   ├── query.go           # Query commands
│   ├── task.go            # Task management commands
│   ├── note.go            # Note management commands
│   ├── workspace*.go      # Workspace management
│   ├── advanced.go        # Analysis commands
│   ├── tools.go           # Utility commands
│   └── ...                # Other commands
│
├── internal/              # Internal packages
│   ├── parser/           # Parsers
│   │   ├── parser.go     # Core parser
│   │   ├── parser_go.go  # Go parser
│   │   ├── parser_c.go   # C/C++ parser
│   │   └── ...           # Other language parsers
│   │
│   ├── query/            # Query management
│   │   ├── manager.go    # Query manager
│   │   └── engine.go     # Query execution engine
│   │
│   ├── aql/              # AI Query Language
│   │   ├── parser.go     # AQL parser
│   │   └── engine.go     # AQL engine
│   │
│   ├── graph/            # Graph analysis
│   │   └── symbol.go     # Symbol graph & PageRank
│   │
│   ├── semantic/         # Semantic analysis
│   │   └── semantic.go   # Intent inference
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
├── queries/               # Predefined AQL queries
│   ├── analysis.json     # Analysis queries
│   ├── api.json          # API queries
│   ├── quality.json      # Quality queries
│   ├── security.json     # Security queries
│   └── ...               # Other query categories
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
- **Analysis**: PageRank, dependency graphs, semantic analysis

## Statistics

- **Lines of Code**: ~11,745
- **Files**: 55 Go files
- **Languages**: 7 (Go, Python, JavaScript, TypeScript, Rust, Java, C/C++)
- **Commands**: 40+ commands
- **Query Templates**: 10+ predefined queries
- **Test Coverage**: 51.9%

## Use Cases

### For AI Models
- **Code Understanding**: Rapidly parse and understand large codebases
- **Code Generation**: Generate code with full context awareness
- **Refactoring**: Safe refactoring with impact analysis
- **Documentation**: Auto-generate documentation from code

### For Developers
- **Code Navigation**: Fast symbol search and navigation
- **Code Review**: Identify quality issues and patterns
- **Dependency Management**: Understand code dependencies
- **Technical Debt**: Find high complexity and problematic code

### For Teams
- **Knowledge Sharing**: Document and share code knowledge
- **Onboarding**: Help new developers understand the codebase
- **Code Quality**: Monitor and improve code quality
- **Security Audits**: Identify security-sensitive code

## Contributing

Contributions, bug reports, and suggestions are welcome!

## License

MIT License

## Author

GLM-5

---

**AI-Git: Text Read/Write System Designed for AI Models**
