# AI-Git

**为 AI 模型设计的文本读写系统**

**版本: 1.0.0**

[English](README_EN.md) | [中文](README_CN.md)

AI-Git 是一个强大的文本读写系统，为 AI 模型提供统一的文本操作能力。它不仅仅是一个代码索引工具，而是模型的完整工作空间，包含代码管理、任务管理、知识库和文档处理。

## 核心理念

```
一切皆文本：
├── 代码 = 文本
├── Markdown = 文本
├── 配置 = 文本
└── 数据 = 文本

核心操作：
├── 读：解析文本 → 提取结构
└── 写：生成文本 → 修改内容
```

## 特性

### 🔍 代码分析
- **多语言支持**: Go, Python, JavaScript, TypeScript, Rust, Java, C/C++
- **符号提取**: 函数、类、变量、接口、结构体等
- **依赖分析**: 代码依赖关系和影响分析
- **PageRank 分析**: 代码重要性评估

### 📝 文档处理
- **Markdown 解析**: 标题、章节、代码块、待办列表
- **结构化索引**: 快速查询文档内容
- **自动监控**: 文件变化自动更新索引

### 📋 任务管理
- **任务创建**: 创建和拆解任务
- **进度追踪**: 实时更新任务进度
- **执行记录**: 记录每一步操作
- **历史归档**: 完成任务自动归档

### 📚 知识库
- **知识笔记**: 创建和管理知识笔记
- **标签系统**: 灵活的标签分类
- **代码关联**: 笔记与代码双向关联
- **搜索导出**: 快速搜索和导出知识

### 🚀 性能优化
- **内存缓存**: O(1) 快速查询
- **实时索引**: 文件变化自动更新
- **并发处理**: 并行解析提升性能
- **REST API**: 远程访问支持

## 安装

```bash
# 克隆仓库
git clone https://github.com/mychow/ai-git.git
cd ai-git

# 编译
go build -o ai-git .

# 安装到系统
sudo mv ai-git /usr/local/bin/
```

## 快速开始

### 0. 启动后台服务（必须）

**重要：使用 AI-Git 前必须先启动后台服务！**

```bash
# 启动后台服务
ai-git daemon start

# 查看服务状态
ai-git daemon status

# 停止服务
ai-git daemon stop
```

后台服务启动后，索引将常驻内存，所有 CLI 命令将通过 API 快速访问，性能提升约 16 倍。

### 1. 初始化项目

```bash
# 在项目目录初始化
ai-git init

# 查看项目概览
ai-git overview
```

### 2. 查询代码

```bash
# 搜索符号
ai-git search "handler"

# 查询符号详情
ai-git symbol "Process"

# 读取代码
ai-git read main.go:10-20
ai-git read Process
```

### 3. 修改代码

```bash
# 插入代码
ai-git insert main.go:10 --code "// comment"

# 替换代码
ai-git replace Process --with "func Process() { ... }"

# 删除代码
ai-git delete OldFunction
```

### 4. 任务管理

```bash
# 创建任务
ai-git task create "实现用户认证" --priority high

# 更新进度
ai-git task update task-001 --progress 50

# 完成任务
ai-git task complete task-001

# 查看任务列表
ai-git task list
```

### 5. 知识管理

```bash
# 创建笔记
ai-git note create "Go 并发模式" --tags "go,concurrency"

# 搜索知识
ai-git note search "并发"

# 关联代码
ai-git note link note-001 --code "worker.go:Process"

# 导出知识
ai-git note export --format markdown
```

### 6. 启动服务

```bash
# 启动 API 服务
ai-git web --port 8080

# 访问 API
curl http://localhost:8080/api/v1/search?q=handler
```

## 📁 配置文件

AI-Git 使用隐藏目录存储配置和数据：

### 全局配置（用户级别）
```
~/.ai-git/              # 全局配置目录
├── config.json         # 全局配置
├── daemon.json         # 后台服务信息
├── daemon.pid          # 后台服务进程ID
├── tasks/              # 任务数据
├── notes/              # 笔记数据
├── queries/            # 查询模板
└── plugins/            # 插件目录
```

### 项目配置（项目级别）
```
项目目录/.ai-git/       # 项目配置目录
└── ai-git.db           # 项目数据库
```

### 查看配置
```bash
# 查看全局配置
cat ~/.ai-git/config.json

# 查看项目数据库
ls -la .ai-git/
```

**注意：配置文件是可选的，AI-Git 可以直接使用默认配置工作！**

## 命令概览

### 查询命令
| 命令 | 说明 |
|------|------|
| `search` | 搜索符号 |
| `symbol` | 查询符号详情 |
| `read` | 读取代码或文件 |
| `overview` | 项目概览 |
| `status` | 项目状态 |

### 修改命令
| 命令 | 说明 |
|------|------|
| `modify` | 修改代码 |
| `insert` | 插入代码 |
| `replace` | 替换代码 |
| `delete` | 删除代码 |
| `refactor` | 重构代码 |
| `batch` | 批量修改 |

### 分析命令
| 命令 | 说明 |
|------|------|
| `analyze` | PageRank 分析 |
| `deps` | 依赖分析 |
| `impact` | 影响分析 |
| `quality` | 代码质量评估 |
| `pattern` | 设计模式识别 |

### 任务命令
| 命令 | 说明 |
|------|------|
| `task create` | 创建任务 |
| `task list` | 列出任务 |
| `task update` | 更新任务 |
| `task complete` | 完成任务 |
| `task log` | 记录执行日志 |

### 笔记命令
| 命令 | 说明 |
|------|------|
| `note create` | 创建笔记 |
| `note list` | 列出笔记 |
| `note search` | 搜索笔记 |
| `note link` | 关联代码 |
| `note export` | 导出笔记 |

### 项目命令
| 命令 | 说明 |
|------|------|
| `init` | 初始化项目 |
| `workspace` | 工作区管理 |
| `web` | 启动 API 服务 |
| `plugin` | 插件管理 |

## 架构

```
┌─────────────────────────────────────────────┐
│           AI-Git: 文本读写系统               │
│                                             │
│  代码层                                     │
│  ├── 7 种语言解析                           │
│  ├── 符号提取和索引                         │
│  ├── 依赖和影响分析                         │
│  └── 修改和重构                             │
│                                             │
│  任务层                                     │
│  ├── 任务创建和拆解                         │
│  ├── 进度追踪                               │
│  ├── 执行记录                               │
│  └── 历史归档                               │
│                                             │
│  知识层                                     │
│  ├── 知识笔记                               │
│  ├── 上下文管理                             │
│  ├── 代码关联                               │
│  └── 搜索和导出                             │
│                                             │
│  文档层                                     │
│  ├── Markdown 解析                          │
│  ├── 标题和章节索引                         │
│  ├── 待办列表追踪                           │
│  └── 代码块提取                             │
│                                             │
│  服务层                                     │
│  ├── REST API                               │
│  ├── 文件监控                               │
│  ├── 实时索引                               │
│  └── 内存缓存                               │
└─────────────────────────────────────────────┘
```

## 项目结构

```
ai-git/
├── cmd/                    # 命令行工具
│   ├── root.go            # 根命令
│   ├── task.go            # 任务管理命令
│   ├── note.go            # 笔记管理命令
│   └── ...                # 其他命令
│
├── internal/              # 内部包
│   ├── parser/           # 解析器
│   │   ├── parser.go     # 核心解析器
│   │   ├── parser_go.go  # Go 解析
│   │   ├── parser_markdown.go # Markdown 解析
│   │   └── ...           # 其他语言解析
│   │
│   ├── task/             # 任务管理
│   ├── note/             # 知识库管理
│   ├── storage/          # 存储层
│   ├── watcher/          # 文件监控
│   └── api/              # REST API
│
├── pkg/                   # 公共包
│   ├── types/            # 类型定义
│   └── utils/            # 工具函数
│
└── main.go               # 入口文件
```

## API 文档

### 搜索
```
GET /api/v1/search?q=handler
```

### 符号查询
```
GET /api/v1/symbol/:name
```

### 项目概览
```
GET /api/v1/overview
```

### 任务管理
```
POST /api/v1/projects
GET /api/v1/projects
DELETE /api/v1/projects/:id
```

### 监控状态
```
GET /api/v1/status/watcher
```

## 开发

### 构建
```bash
make build
```

### 测试
```bash
make test
```

### 清理
```bash
make clean
```

## 技术栈

- **语言**: Go 1.21+
- **数据库**: BoltDB (持久化) + 内存缓存
- **解析器**: 自定义解析器 (支持 7 种语言)
- **API**: REST API
- **监控**: fsnotify

## 统计

- **代码行数**: ~11,745 行
- **文件数量**: 55 个 Go 文件
- **支持语言**: 7 种 (Go, Python, JavaScript, TypeScript, Rust, Java, C/C++)
- **命令数量**: 40+ 个命令

## 贡献

欢迎贡献代码、报告问题或提出建议！

## 许可证

MIT License

## 作者

GLM-5

---

**AI-Git: 为 AI 模型设计的文本读写系统**
