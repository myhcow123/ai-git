package aql

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mychow/ai-git/internal/graph"
	"github.com/mychow/ai-git/internal/index"
	"github.com/mychow/ai-git/internal/modify"
	"github.com/mychow/ai-git/internal/parser"
	"github.com/mychow/ai-git/internal/semantic"
	"github.com/mychow/ai-git/internal/storage"
	"github.com/mychow/ai-git/pkg/cache"
	"github.com/mychow/ai-git/pkg/types"
	"github.com/mychow/ai-git/pkg/utils"
)

type Engine struct {
	storage     *storage.Storage
	index       *index.Index
	parser      *parser.CodeParser
	scanner     *utils.ProjectScanner
	graph       *graph.SymbolGraph
	mu          sync.RWMutex
	rootPath    string
	symbolCache *cache.Cache
}

func NewEngine(rootPath string) (*Engine, error) {
	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	dbPath := filepath.Join(absPath, ".ai-git", "ai-git.db")
	storage, err := storage.NewStorage(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	if err := storage.LoadToCache(); err != nil {
		fmt.Printf("Warning: failed to load cache: %v\n", err)
	}

	engine := &Engine{
		storage:     storage,
		index:       index.NewIndex(),
		parser:      parser.NewCodeParser(),
		scanner:     utils.NewProjectScanner(absPath),
		graph:       graph.NewSymbolGraph(),
		rootPath:    absPath,
		symbolCache: cache.NewCache(),
	}

	if err := engine.rebuildIndex(); err != nil {
		fmt.Printf("Warning: failed to rebuild index: %v\n", err)
	}

	return engine, nil
}

func (e *Engine) rebuildIndex() error {
	symbols, err := e.storage.GetAllSymbols()
	if err != nil {
		return err
	}

	for _, symbol := range symbols {
		e.index.AddSymbol(symbol)
		e.graph.AddNode(symbol)
	}

	e.buildGraphEdges(symbols)

	return nil
}

func (e *Engine) buildGraphEdges(symbols []*types.Symbol) {
	symbolMap := make(map[string]*types.Symbol)
	for _, symbol := range symbols {
		symbolMap[symbol.Name] = symbol
	}

	fileContents := make(map[string]string)
	for _, symbol := range symbols {
		if _, exists := fileContents[symbol.File]; !exists {
			content, err := e.storage.GetFile(symbol.File)
			if err == nil {
				fileContents[symbol.File] = content
			}
		}
	}

	for _, symbol := range symbols {
		fileContent, exists := fileContents[symbol.File]
		if !exists {
			continue
		}

		lines := strings.Split(fileContent, "\n")
		startLine := symbol.LineStart - 1
		if startLine < 0 || startLine >= len(lines) {
			continue
		}

		endLine := len(lines)
		if symbol.LineEnd > 0 && symbol.LineEnd < len(lines) {
			endLine = symbol.LineEnd
		} else {
			braceCount := 0
			for i := startLine; i < len(lines); i++ {
				line := lines[i]
				braceCount += strings.Count(line, "{") - strings.Count(line, "}")
				if i > startLine && braceCount == 0 {
					endLine = i + 1
					break
				}
			}
		}

		if endLine > len(lines) {
			endLine = len(lines)
		}

		code := strings.Join(lines[startLine:endLine], "\n")

		for otherName, otherSymbol := range symbolMap {
			if otherName == symbol.Name {
				continue
			}

			if strings.Contains(code, otherName) {
				e.graph.AddEdge(symbol.ID, otherSymbol.ID, types.EdgeCalls, 1.0)
			}
		}
	}
}

func (e *Engine) Initialize() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	fmt.Println("Scanning project files...")
	files, err := e.scanner.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan project: %w", err)
	}

	fmt.Printf("Found %d files to parse\n", len(files))

	symbolCount := 0
	for i, file := range files {
		if (i+1)%10 == 0 || i == len(files)-1 {
			fmt.Printf("Parsing files: %d/%d\n", i+1, len(files))
		}

		content, err := e.scanner.ReadFile(file.Path)
		if err != nil {
			fmt.Printf("Warning: failed to read %s: %v\n", file.Path, err)
			continue
		}

		symbols, err := e.parser.ParseFile(file.Path, content)
		if err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", file.Path, err)
			continue
		}

		for _, symbol := range symbols {
			if err := e.storage.SaveSymbol(symbol); err != nil {
				fmt.Printf("Warning: failed to save symbol %s: %v\n", symbol.ID, err)
				continue
			}
			e.index.AddSymbol(symbol)
			symbolCount++
		}

		if err := e.storage.SaveFile(file.Path, content); err != nil {
			fmt.Printf("Warning: failed to save file %s: %v\n", file.Path, err)
		}
	}

	snapshot := &types.Snapshot{
		ID:        fmt.Sprintf("snap_%d", time.Now().Unix()),
		Timestamp: time.Now().Unix(),
		Symbols:   make(map[string]*types.Symbol),
		Metadata: types.SnapshotMetadata{
			Purpose:    "Initial project scan",
			Confidence: 1.0,
			Quality:    1.0,
			CreatedAt:  time.Now(),
		},
	}

	if err := e.storage.SaveSnapshot(snapshot); err != nil {
		return fmt.Errorf("failed to save initial snapshot: %w", err)
	}

	projectInfo := map[string]interface{}{
		"root_path":     e.rootPath,
		"initialized":   time.Now(),
		"total_files":   len(files),
		"total_symbols": symbolCount,
	}

	if err := e.storage.SetMetadata("project_info", projectInfo); err != nil {
		return fmt.Errorf("failed to save project info: %w", err)
	}

	fmt.Printf("✓ Project initialized successfully\n")
	fmt.Printf("✓ Indexed %d symbols from %d files\n", symbolCount, len(files))

	return nil
}

func (e *Engine) Sync() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	fmt.Println("Scanning for changes...")
	files, err := e.scanner.Scan()
	if err != nil {
		return fmt.Errorf("failed to scan project: %w", err)
	}

	existingMetas, err := e.storage.GetAllFileMetas()
	if err != nil {
		existingMetas = make(map[string]*storage.FileMeta)
	}

	currentFiles := make(map[string]bool)
	for _, f := range files {
		currentFiles[f.Path] = true
	}

	var filesToParse []utils.FileInfo
	var filesToDelete []string

	for path := range existingMetas {
		if !currentFiles[path] {
			filesToDelete = append(filesToDelete, path)
		}
	}

	for _, file := range files {
		existingMeta, hasMeta := existingMetas[file.Path]
		if !hasMeta || existingMeta.Size != file.Size {
			filesToParse = append(filesToParse, file)
		}
	}

	if len(filesToParse) == 0 && len(filesToDelete) == 0 {
		fmt.Println("✓ No changes detected")
		return nil
	}

	fmt.Printf("Found %d files to update, %d files to remove\n", len(filesToParse), len(filesToDelete))

	for _, path := range filesToDelete {
		if err := e.storage.DeleteSymbolsForFile(path); err != nil {
			fmt.Printf("Warning: failed to delete symbols for %s: %v\n", path, err)
		}
		if err := e.storage.DeleteFileMeta(path); err != nil {
			fmt.Printf("Warning: failed to delete file meta for %s: %v\n", path, err)
		}
		fmt.Printf("Removed: %s\n", path)
	}

	if len(filesToParse) == 0 {
		fmt.Println("✓ Sync completed")
		return nil
	}

	for _, file := range filesToParse {
		content, err := e.scanner.ReadFile(file.Path)
		if err != nil {
			continue
		}

		symbols, err := e.parser.ParseFile(file.Path, content)
		if err != nil {
			continue
		}

		if err := e.storage.DeleteSymbolsForFile(file.Path); err != nil {
			fmt.Printf("Warning: failed to delete old symbols for %s: %v\n", file.Path, err)
		}

		if err := e.storage.SaveSymbolsBatch(symbols); err != nil {
			fmt.Printf("Warning: failed to save symbols for %s: %v\n", file.Path, err)
			continue
		}

		for _, symbol := range symbols {
			e.index.AddSymbol(symbol)
		}

		meta := &storage.FileMeta{
			ModTime:     time.Now().Unix(),
			Size:        file.Size,
			SymbolCount: len(symbols),
		}
		if err := e.storage.SaveFileMeta(file.Path, meta); err != nil {
			fmt.Printf("Warning: failed to save file meta for %s: %v\n", file.Path, err)
		}

		fmt.Printf("Updated: %s (%d symbols)\n", file.Path, len(symbols))
	}

	fmt.Println("✓ Sync completed")
	return nil
}

func (e *Engine) GetSymbol(name string) ([]*types.Symbol, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ids := e.index.GetByName(name)
	if len(ids) == 0 {
		allSymbols, err := e.storage.GetAllSymbols()
		if err != nil {
			return nil, fmt.Errorf("symbol not found: %s", name)
		}

		type scoredSymbol struct {
			symbol *types.Symbol
			score  int
		}

		var scoredResults []scoredSymbol
		nameLower := strings.ToLower(name)
		for _, symbol := range allSymbols {
			symbolNameLower := strings.ToLower(symbol.Name)
			if strings.Contains(symbolNameLower, nameLower) {
				score := 0
				if symbol.Name == name {
					score = 100
				} else if symbolNameLower == nameLower {
					score = 90
				} else if strings.HasPrefix(symbol.Name, name) {
					score = 80
				} else if strings.HasPrefix(symbolNameLower, nameLower) {
					score = 70
				} else {
					score = 50
				}
				scoredResults = append(scoredResults, scoredSymbol{symbol: symbol, score: score})
			}
		}

		if len(scoredResults) == 0 {
			return nil, fmt.Errorf("symbol not found: %s", name)
		}

		sort.Slice(scoredResults, func(i, j int) bool {
			return scoredResults[i].score > scoredResults[j].score
		})

		results := make([]*types.Symbol, len(scoredResults))
		for i, sr := range scoredResults {
			results[i] = sr.symbol
		}

		return results, nil
	}

	symbols := make([]*types.Symbol, 0, len(ids))
	for _, id := range ids {
		symbol, err := e.storage.GetSymbol(id)
		if err != nil {
			fmt.Printf("Warning: failed to get symbol %s: %v\n", id, err)
			continue
		}
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

func (e *Engine) SearchSymbols(query string) ([]*types.Symbol, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	allSymbols, err := e.storage.GetAllSymbols()
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	results := make([]*types.Symbol, 0)
	for _, symbol := range allSymbols {
		if containsIgnoreCase(symbol.Name, query) ||
			containsIgnoreCase(symbol.Signature, query) ||
			containsIgnoreCase(symbol.File, query) {
			results = append(results, symbol)
		}
	}

	return results, nil
}

func (e *Engine) SearchSymbolsRegex(pattern string) ([]*types.Symbol, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	allSymbols, err := e.storage.GetAllSymbols()
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	results := make([]*types.Symbol, 0)
	for _, symbol := range allSymbols {
		if re.MatchString(symbol.Name) ||
			re.MatchString(symbol.Signature) ||
			re.MatchString(symbol.File) {
			results = append(results, symbol)
		}
	}

	return results, nil
}

func (e *Engine) GetOverview() (map[string]interface{}, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	allSymbols, err := e.storage.GetAllSymbols()
	if err != nil {
		return nil, fmt.Errorf("failed to get symbols: %w", err)
	}

	stats := map[string]interface{}{
		"total_symbols": len(allSymbols),
		"by_type":       make(map[string]int),
		"by_language":   make(map[string]int),
		"by_file":       make(map[string]int),
	}

	byType := stats["by_type"].(map[string]int)
	byLanguage := stats["by_language"].(map[string]int)
	byFile := stats["by_file"].(map[string]int)

	for _, symbol := range allSymbols {
		byType[symbol.Type.String()]++
		byLanguage[string(e.parser.DetectLanguage(symbol.File))]++
		byFile[symbol.File]++
	}

	stats["index_stats"] = e.index.Stats()

	return stats, nil
}

func (e *Engine) GetStatus() (map[string]interface{}, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var projectInfo map[string]interface{}
	if err := e.storage.GetMetadata("project_info", &projectInfo); err != nil {
		projectInfo = map[string]interface{}{
			"status": "not initialized",
		}
	}

	latestSnapshot, err := e.storage.GetLatestSnapshot()
	if err != nil {
		latestSnapshot = nil
	}

	status := map[string]interface{}{
		"project_info": projectInfo,
		"index_stats":  e.index.Stats(),
	}

	if latestSnapshot != nil {
		status["latest_snapshot"] = map[string]interface{}{
			"id":        latestSnapshot.ID,
			"timestamp": latestSnapshot.Timestamp,
			"symbols":   len(latestSnapshot.Symbols),
		}
	}

	return status, nil
}

func (e *Engine) Close() error {
	return e.storage.Close()
}

func (e *Engine) GetStorage() *storage.Storage {
	return e.storage
}

func (e *Engine) GetIndex() *index.Index {
	return e.index
}

func (e *Engine) GetGraph() *graph.SymbolGraph {
	return e.graph
}

func (e *Engine) GetSemanticEngine() *semantic.IntentEngine {
	return semantic.NewIntentEngine()
}

func (e *Engine) GetCodeGenerator() *modify.CodeGenerator {
	return modify.NewCodeGenerator(e.parser)
}

func (e *Engine) GetParser() *parser.CodeParser {
	return e.parser
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
