# AI-Git 查询库

这个文件夹包含预定义的 AI Query Language (AQL) 查询模板，用于快速执行常见的代码分析任务。

## 查询分类

### 1. 探索类
- **find_all_functions**: 查找所有函数
- **find_classes**: 查找所有类定义

### 2. 分析类
- **find_entry_points**: 查找所有入口点（main 函数、处理器等）

### 3. 质量类
- **find_high_complexity**: 查找高复杂度函数

### 4. 测试类
- **find_test_code**: 查找所有测试函数和测试文件

### 5. API 类
- **find_api_endpoints**: 查找 API 端点定义

### 6. 数据库类
- **find_database_operations**: 查找数据库相关函数

### 7. 安全类
- **find_security_sensitive**: 查找安全敏感代码

### 8. 工具类
- **find_utility_functions**: 查找工具和辅助函数

### 9. 架构类
- **find_interfaces**: 查找所有接口定义

## 使用方法

### 列出所有查询
```bash
ai-git query list
```

### 查看查询详情
```bash
ai-git query show find_all_functions
```

### 搜索查询
```bash
ai-git query search "api"
```

### 按分类列出查询
```bash
ai-git query list --category api
```

### 执行查询
```bash
ai-git query exec find_entry_points
```

### 查看所有分类
```bash
ai-git query categories
```

## 查询模板格式

每个查询模板是一个 JSON 文件，包含以下字段：

```json
{
  "name": "query_name",
  "description": "查询描述",
  "category": "分类",
  "query": "FIND function WHERE condition",
  "examples": [
    "示例查询 1",
    "示例查询 2"
  ],
  "use_cases": [
    "使用场景 1",
    "使用场景 2"
  ]
}
```

## 创建自定义查询

1. 创建新的 JSON 文件
2. 按照模板格式填写查询信息
3. 将文件保存到 `queries/` 目录
4. 使用 `ai-git query list` 验证查询已加载

## 查询示例

### 查找所有函数
```bash
ai-git query exec find_all_functions
```

### 查找高复杂度代码
```bash
ai-git query exec find_high_complexity
```

### 查找 API 端点
```bash
ai-git query exec find_api_endpoints
```

### 查找安全敏感代码
```bash
ai-git query exec find_security_sensitive
```

## 高级用法

### 带参数的查询
某些查询支持参数替换：

```bash
ai-git query exec find_functions_in_file --param file=main.go
```

### 组合查询
可以组合多个查询进行复杂分析：

```bash
# 先查找高复杂度函数，再分析其依赖
ai-git query exec find_high_complexity
ai-git analyze --top 10
```

## 查询最佳实践

1. **从简单开始**: 使用 `find_all_functions` 了解项目结构
2. **识别热点**: 使用 `find_high_complexity` 找到需要重构的代码
3. **安全审计**: 定期使用 `find_security_sensitive` 检查安全相关代码
4. **测试覆盖**: 使用 `find_test_code` 确保测试完整性
5. **API 文档**: 使用 `find_api_endpoints` 生成 API 文档

## 贡献查询

欢迎贡献新的查询模板！请确保：

1. 查询名称清晰描述功能
2. 提供详细的描述和使用场景
3. 包含多个示例
4. 选择合适的分类
5. 测试查询的正确性

## 查询管理器 API

查询管理器提供以下功能：

- `Load()`: 从目录加载所有查询
- `Get(name)`: 获取指定查询
- `List()`: 列出所有查询
- `ListByCategory(category)`: 按分类列出查询
- `Search(keyword)`: 搜索查询
- `Add(query)`: 添加新查询
- `Save(name, path)`: 保存查询到文件
- `Execute(name, params)`: 执行查询

## 集成到工作流

### CI/CD 集成
```bash
# 在 CI 中检查代码质量
ai-git query exec find_high_complexity > complexity_report.txt
```

### 代码审查
```bash
# 生成代码审查报告
ai-git query exec find_entry_points > entry_points.txt
ai-git query exec find_security_sensitive > security_audit.txt
```

### 文档生成
```bash
# 生成 API 文档
ai-git query exec find_api_endpoints > api_docs.md
```
