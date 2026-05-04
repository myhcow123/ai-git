package query

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type QueryTemplate struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Query       string   `json:"query"`
	Examples    []string `json:"examples"`
	UseCases    []string `json:"use_cases"`
}

type Manager struct {
	queries    map[string]*QueryTemplate
	categories map[string][]string
	queryDir   string
}

func NewManager(queryDir string) *Manager {
	return &Manager{
		queries:    make(map[string]*QueryTemplate),
		categories: make(map[string][]string),
		queryDir:   queryDir,
	}
}

func (m *Manager) Load() error {
	if m.queryDir == "" {
		return nil
	}

	info, err := os.Stat(m.queryDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to stat query directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("query path is not a directory: %s", m.queryDir)
	}

	entries, err := os.ReadDir(m.queryDir)
	if err != nil {
		return fmt.Errorf("failed to read query directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(m.queryDir, entry.Name())
		if err := m.loadFile(path); err != nil {
			return fmt.Errorf("failed to load query %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func (m *Manager) loadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var query QueryTemplate
	if err := json.Unmarshal(data, &query); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	if query.Name == "" {
		return fmt.Errorf("query name is required")
	}

	m.queries[query.Name] = &query
	m.categories[query.Category] = append(m.categories[query.Category], query.Name)

	return nil
}

func (m *Manager) Get(name string) (*QueryTemplate, error) {
	query, exists := m.queries[name]
	if !exists {
		return nil, fmt.Errorf("query not found: %s", name)
	}
	return query, nil
}

func (m *Manager) List() []*QueryTemplate {
	var queries []*QueryTemplate
	for _, query := range m.queries {
		queries = append(queries, query)
	}
	sort.Slice(queries, func(i, j int) bool {
		if queries[i].Category != queries[j].Category {
			return queries[i].Category < queries[j].Category
		}
		return queries[i].Name < queries[j].Name
	})
	return queries
}

func (m *Manager) ListByCategory(category string) []*QueryTemplate {
	var queries []*QueryTemplate
	for _, query := range m.queries {
		if query.Category == category {
			queries = append(queries, query)
		}
	}
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Name < queries[j].Name
	})
	return queries
}

func (m *Manager) GetCategories() []string {
	var categories []string
	for cat := range m.categories {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories
}

func (m *Manager) Search(keyword string) []*QueryTemplate {
	var results []*QueryTemplate
	keyword = strings.ToLower(keyword)

	for _, query := range m.queries {
		if strings.Contains(strings.ToLower(query.Name), keyword) ||
			strings.Contains(strings.ToLower(query.Description), keyword) ||
			strings.Contains(strings.ToLower(query.Category), keyword) {
			results = append(results, query)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results
}

func (m *Manager) Add(query *QueryTemplate) error {
	if query.Name == "" {
		return fmt.Errorf("query name is required")
	}

	if _, exists := m.queries[query.Name]; exists {
		return fmt.Errorf("query already exists: %s", query.Name)
	}

	m.queries[query.Name] = query
	m.categories[query.Category] = append(m.categories[query.Category], query.Name)

	return nil
}

func (m *Manager) Save(name string, path string) error {
	query, exists := m.queries[name]
	if !exists {
		return fmt.Errorf("query not found: %s", name)
	}

	data, err := json.MarshalIndent(query, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	if path == "" {
		path = filepath.Join(m.queryDir, query.Category+".json")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (m *Manager) Delete(name string) error {
	query, exists := m.queries[name]
	if !exists {
		return fmt.Errorf("query not found: %s", name)
	}

	category := query.Category
	categoryQueries := m.categories[category]
	for i, qname := range categoryQueries {
		if qname == name {
			m.categories[category] = append(categoryQueries[:i], categoryQueries[i+1:]...)
			break
		}
	}

	delete(m.queries, name)
	return nil
}

func (m *Manager) Execute(name string, params map[string]string) (string, error) {
	query, err := m.Get(name)
	if err != nil {
		return "", err
	}

	result := query.Query
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

func (m *Manager) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_queries": len(m.queries),
		"categories":    len(m.categories),
		"by_category":   m.getCategoryStats(),
	}
}

func (m *Manager) getCategoryStats() map[string]int {
	stats := make(map[string]int)
	for cat, queries := range m.categories {
		stats[cat] = len(queries)
	}
	return stats
}

func FormatQueryList(queries []*QueryTemplate, verbose bool) string {
	var builder strings.Builder

	for i, query := range queries {
		if i > 0 {
			builder.WriteString("\n")
		}

		builder.WriteString(fmt.Sprintf("Name: %s\n", query.Name))
		builder.WriteString(fmt.Sprintf("Category: %s\n", query.Category))
		builder.WriteString(fmt.Sprintf("Description: %s\n", query.Description))

		if verbose {
			builder.WriteString(fmt.Sprintf("Query: %s\n", query.Query))

			if len(query.Examples) > 0 {
				builder.WriteString("Examples:\n")
				for _, example := range query.Examples {
					builder.WriteString(fmt.Sprintf("  - %s\n", example))
				}
			}

			if len(query.UseCases) > 0 {
				builder.WriteString("Use Cases:\n")
				for _, uc := range query.UseCases {
					builder.WriteString(fmt.Sprintf("  - %s\n", uc))
				}
			}
		}
	}

	return builder.String()
}
