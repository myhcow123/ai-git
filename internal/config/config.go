package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Project     ProjectConfig     `json:"project"`
	Storage     StorageConfig     `json:"storage"`
	Parser      ParserConfig      `json:"parser"`
	Index       IndexConfig       `json:"index"`
	Graph       GraphConfig       `json:"graph"`
	Semantic    SemanticConfig    `json:"semantic"`
	Performance PerformanceConfig `json:"performance"`
	Output      OutputConfig      `json:"output"`
}

type ProjectConfig struct {
	Name        string   `json:"name"`
	Root        string   `json:"root"`
	IgnoreDirs  []string `json:"ignore_dirs"`
	IgnoreFiles []string `json:"ignore_files"`
	Languages   []string `json:"languages"`
}

type StorageConfig struct {
	Type     string `json:"type"`
	Database string `json:"database"`
	ReadOnly bool   `json:"read_only"`
}

type ParserConfig struct {
	MaxFileSize   int64    `json:"max_file_size"`
	ExtractDocs   bool     `json:"extract_docs"`
	ExtractParams bool     `json:"extract_params"`
	IgnoreDirs    []string `json:"ignore_dirs"`
}

type IndexConfig struct {
	EnableName bool `json:"enable_name"`
	EnableType bool `json:"enable_type"`
	EnableFile bool `json:"enable_file"`
	EnableSig  bool `json:"enable_sig"`
	EnableDeps bool `json:"enable_deps"`
}

type GraphConfig struct {
	EnablePageRank     bool    `json:"enable_pagerank"`
	PageRankDamping    float64 `json:"pagerank_damping"`
	PageRankIterations int     `json:"pagerank_iterations"`
	EnableImpact       bool    `json:"enable_impact"`
}

type SemanticConfig struct {
	EnableIntent  bool `json:"enable_intent"`
	EnablePattern bool `json:"enable_pattern"`
	EnableQuality bool `json:"enable_quality"`
	MaxComplexity int  `json:"max_complexity"`
}

type PerformanceConfig struct {
	ParallelWorkers int  `json:"parallel_workers"`
	EnableCache     bool `json:"enable_cache"`
	CacheSize       int  `json:"cache_size"`
	EnableVector    bool `json:"enable_vector"`
	VectorDimension int  `json:"vector_dimension"`
}

type OutputConfig struct {
	Format    string `json:"format"`
	Color     bool   `json:"color"`
	Indent    bool   `json:"indent"`
	ShowStats bool   `json:"show_stats"`
	ShowTime  bool   `json:"show_time"`
	Verbose   bool   `json:"verbose"`
}

var DefaultConfig = Config{
	Project: ProjectConfig{
		IgnoreDirs:  []string{".git", "node_modules", "vendor", "dist", "build"},
		IgnoreFiles: []string{"*.min.js", "*.min.css", "*.lock"},
		Languages:   []string{"go", "python", "javascript", "typescript"},
	},
	Storage: StorageConfig{
		Type:     "boltdb",
		Database: ".ai-git.db",
		ReadOnly: false,
	},
	Parser: ParserConfig{
		MaxFileSize:   10 * 1024 * 1024,
		ExtractDocs:   true,
		ExtractParams: true,
		IgnoreDirs:    []string{".git", "node_modules", "vendor"},
	},
	Index: IndexConfig{
		EnableName: true,
		EnableType: true,
		EnableFile: true,
		EnableSig:  true,
		EnableDeps: true,
	},
	Graph: GraphConfig{
		EnablePageRank:     true,
		PageRankDamping:    0.85,
		PageRankIterations: 20,
		EnableImpact:       true,
	},
	Semantic: SemanticConfig{
		EnableIntent:  true,
		EnablePattern: true,
		EnableQuality: true,
		MaxComplexity: 15,
	},
	Performance: PerformanceConfig{
		ParallelWorkers: 4,
		EnableCache:     true,
		CacheSize:       1000,
		EnableVector:    true,
		VectorDimension: 128,
	},
	Output: OutputConfig{
		Format:    "json",
		Color:     true,
		Indent:    true,
		ShowStats: true,
		ShowTime:  true,
		Verbose:   false,
	},
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = findConfigFile()
	}

	if path == "" {
		cfg := DefaultConfig
		return &cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	applyDefaults(&cfg)

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func findConfigFile() string {
	candidates := []string{
		".ai-git.json",
		"ai-git.json",
		".ai-gitrc",
		"ai-git.config.json",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func applyDefaults(cfg *Config) {
	if cfg.Project.IgnoreDirs == nil {
		cfg.Project.IgnoreDirs = DefaultConfig.Project.IgnoreDirs
	}
	if cfg.Project.IgnoreFiles == nil {
		cfg.Project.IgnoreFiles = DefaultConfig.Project.IgnoreFiles
	}
	if cfg.Project.Languages == nil {
		cfg.Project.Languages = DefaultConfig.Project.Languages
	}
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = DefaultConfig.Storage.Type
	}
	if cfg.Storage.Database == "" {
		cfg.Storage.Database = DefaultConfig.Storage.Database
	}
	if cfg.Parser.MaxFileSize == 0 {
		cfg.Parser.MaxFileSize = DefaultConfig.Parser.MaxFileSize
	}
	if cfg.Graph.PageRankDamping == 0 {
		cfg.Graph.PageRankDamping = DefaultConfig.Graph.PageRankDamping
	}
	if cfg.Graph.PageRankIterations == 0 {
		cfg.Graph.PageRankIterations = DefaultConfig.Graph.PageRankIterations
	}
	if cfg.Semantic.MaxComplexity == 0 {
		cfg.Semantic.MaxComplexity = DefaultConfig.Semantic.MaxComplexity
	}
	if cfg.Performance.ParallelWorkers == 0 {
		cfg.Performance.ParallelWorkers = DefaultConfig.Performance.ParallelWorkers
	}
	if cfg.Performance.CacheSize == 0 {
		cfg.Performance.CacheSize = DefaultConfig.Performance.CacheSize
	}
	if cfg.Performance.VectorDimension == 0 {
		cfg.Performance.VectorDimension = DefaultConfig.Performance.VectorDimension
	}
	if cfg.Output.Format == "" {
		cfg.Output.Format = DefaultConfig.Output.Format
	}
}

func (c *Config) ShouldIgnoreDir(dir string) bool {
	for _, ignore := range c.Project.IgnoreDirs {
		if dir == ignore || strings.HasPrefix(dir, ignore+"/") {
			return true
		}
	}
	for _, ignore := range c.Parser.IgnoreDirs {
		if dir == ignore || strings.HasPrefix(dir, ignore+"/") {
			return true
		}
	}
	return false
}

func (c *Config) ShouldIgnoreFile(file string) bool {
	for _, pattern := range c.Project.IgnoreFiles {
		matched, err := filepath.Match(pattern, filepath.Base(file))
		if err == nil && matched {
			return true
		}
	}
	return false
}

func (c *Config) IsLanguageEnabled(lang string) bool {
	if len(c.Project.Languages) == 0 {
		return true
	}
	for _, l := range c.Project.Languages {
		if strings.EqualFold(l, lang) {
			return true
		}
	}
	return false
}

func (c *Config) GetDatabasePath() string {
	if filepath.IsAbs(c.Storage.Database) {
		return c.Storage.Database
	}
	if c.Project.Root != "" {
		return filepath.Join(c.Project.Root, c.Storage.Database)
	}
	return c.Storage.Database
}

func GenerateSampleConfig(path string) error {
	cfg := DefaultConfig
	cfg.Project.Name = "my-project"
	cfg.Project.Root = "."
	return cfg.Save(path)
}
