package modify

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mychow/ai-git/internal/graph"
	"github.com/mychow/ai-git/internal/index"
	"github.com/mychow/ai-git/internal/parser"
	"github.com/mychow/ai-git/pkg/types"
)

type LocationType string

const (
	LocationSymbol   LocationType = "symbol"
	LocationRelative LocationType = "relative"
	LocationSemantic LocationType = "semantic"
	LocationPattern  LocationType = "pattern"
)

type Location struct {
	Type     LocationType
	Target   string
	Position string
	File     string
	Line     int
}

type CodeLocation struct {
	File      string
	LineStart int
	LineEnd   int
	Code      string
	Context   string
}

type LocationEngine struct {
	index   *index.Index
	graph   *graph.SymbolGraph
	parser  *parser.CodeParser
	storage Storage
}

type Storage interface {
	GetSymbol(id string) (*types.Symbol, error)
}

func NewLocationEngine(idx *index.Index, g *graph.SymbolGraph, p *parser.CodeParser, s Storage) *LocationEngine {
	return &LocationEngine{
		index:   idx,
		graph:   g,
		parser:  p,
		storage: s,
	}
}

func (e *LocationEngine) Locate(loc Location) (*CodeLocation, error) {
	switch loc.Type {
	case LocationSymbol:
		return e.locateSymbol(loc.Target)
	case LocationRelative:
		return e.locateRelative(loc.Target, loc.Position)
	case LocationSemantic:
		return e.locateSemantic(loc.Target)
	case LocationPattern:
		return e.locatePattern(loc.Target)
	default:
		return nil, fmt.Errorf("unknown location type: %s", loc.Type)
	}
}

func (e *LocationEngine) locateSymbol(symbolName string) (*CodeLocation, error) {
	ids := e.index.GetByName(symbolName)
	if len(ids) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	id := ids[0]
	symbol, err := e.getSymbolByID(id)
	if err != nil {
		return nil, err
	}

	return &CodeLocation{
		File:      symbol.File,
		LineStart: symbol.LineStart,
		LineEnd:   symbol.LineEnd,
		Code:      symbol.Code,
		Context:   symbol.Signature,
	}, nil
}

func (e *LocationEngine) locateRelative(symbolName, position string) (*CodeLocation, error) {
	baseLoc, err := e.locateSymbol(symbolName)
	if err != nil {
		return nil, err
	}

	switch position {
	case "before_first_line":
		return &CodeLocation{
			File:      baseLoc.File,
			LineStart: baseLoc.LineStart,
			LineEnd:   baseLoc.LineStart,
			Code:      "",
			Context:   "before function",
		}, nil

	case "after_last_line":
		return &CodeLocation{
			File:      baseLoc.File,
			LineStart: baseLoc.LineEnd,
			LineEnd:   baseLoc.LineEnd,
			Code:      "",
			Context:   "after function",
		}, nil

	default:
		return baseLoc, nil
	}
}

func (e *LocationEngine) locateSemantic(description string) (*CodeLocation, error) {
	results := e.index.SearchByDescription(description)
	if len(results) == 0 {
		return nil, fmt.Errorf("no symbol found matching description: %s", description)
	}

	topResult := results[0]
	symbol, err := e.getSymbolByID(topResult)
	if err != nil {
		return nil, err
	}

	return &CodeLocation{
		File:      symbol.File,
		LineStart: symbol.LineStart,
		LineEnd:   symbol.LineEnd,
		Code:      symbol.Code,
		Context:   symbol.Signature,
	}, nil
}

func (e *LocationEngine) locatePattern(pattern string) (*CodeLocation, error) {
	results := e.index.SearchByPattern(pattern)
	if len(results) == 0 {
		return nil, fmt.Errorf("no symbol found matching pattern: %s", pattern)
	}

	topResult := results[0]
	symbol, err := e.getSymbolByID(topResult)
	if err != nil {
		return nil, err
	}

	return &CodeLocation{
		File:      symbol.File,
		LineStart: symbol.LineStart,
		LineEnd:   symbol.LineEnd,
		Code:      symbol.Code,
		Context:   symbol.Signature,
	}, nil
}

func (e *LocationEngine) getSymbolByID(id string) (*types.Symbol, error) {
	if e.storage == nil {
		return nil, fmt.Errorf("storage not available")
	}

	symbol, err := e.storage.GetSymbol(id)
	if err != nil {
		return nil, fmt.Errorf("symbol not found: %s", id)
	}

	return symbol, nil
}

type OperationType string

const (
	OperationAddParameter    OperationType = "add_parameter"
	OperationRemoveParameter OperationType = "remove_parameter"
	OperationModifySignature OperationType = "modify_signature"
	OperationInsertCode      OperationType = "insert_code"
	OperationReplaceCode     OperationType = "replace_code"
	OperationDeleteCode      OperationType = "delete_code"
)

type Operation struct {
	Type           OperationType
	Location       *Location
	Modifications  []Modification
	AutoUpdateRefs bool
}

type Modification struct {
	Type      string
	Parameter *types.Parameter
	Code      string
	Position  string
	OldValue  string
	NewValue  string
}

type ModificationResult struct {
	Success         bool
	FilesChanged    []string
	SymbolsModified []string
	Errors          []string
}

type ModifyEngine struct {
	locationEngine *LocationEngine
	graph          *graph.SymbolGraph
	index          *index.Index
}

func NewModifyEngine(locEngine *LocationEngine, g *graph.SymbolGraph, idx *index.Index) *ModifyEngine {
	return &ModifyEngine{
		locationEngine: locEngine,
		graph:          g,
		index:          idx,
	}
}

func (e *ModifyEngine) Apply(op Operation) (*ModificationResult, error) {
	result := &ModificationResult{
		Success:         false,
		FilesChanged:    []string{},
		SymbolsModified: []string{},
		Errors:          []string{},
	}

	loc, err := e.locationEngine.Locate(*op.Location)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, err
	}

	content, err := os.ReadFile(loc.File)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read file: %v", err))
		return result, err
	}

	lines := strings.Split(string(content), "\n")

	switch op.Type {
	case OperationAddParameter:
		newLines, err := e.addParameter(lines, loc, op.Modifications[0])
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}
		lines = newLines

	case OperationInsertCode:
		newLines, err := e.insertCode(lines, loc, op.Modifications[0])
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}
		lines = newLines

	case OperationReplaceCode:
		newLines, err := e.replaceCode(lines, loc, op.Modifications[0])
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}
		lines = newLines

	case OperationDeleteCode:
		newLines, err := e.deleteCode(lines, loc)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			return result, err
		}
		lines = newLines

	default:
		return result, fmt.Errorf("unsupported operation type: %s", op.Type)
	}

	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(loc.File, []byte(newContent), 0644); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to write file: %v", err))
		return result, err
	}

	result.Success = true
	result.FilesChanged = append(result.FilesChanged, loc.File)

	return result, nil
}

func (e *ModifyEngine) addParameter(lines []string, loc *CodeLocation, mod Modification) ([]string, error) {
	if loc.LineStart >= len(lines) {
		return nil, fmt.Errorf("line number out of range")
	}

	signatureLine := lines[loc.LineStart-1]

	closingParen := strings.LastIndex(signatureLine, ")")
	if closingParen == -1 {
		return nil, fmt.Errorf("invalid function signature")
	}

	param := mod.Parameter
	paramStr := param.Name
	if param.Type != "" {
		paramStr += " " + param.Type
	}
	if param.Default != "" {
		paramStr += "=" + param.Default
	}

	if strings.Contains(signatureLine[:closingParen], "(") {
		if signatureLine[closingParen-1] != '(' {
			paramStr = ", " + paramStr
		}
	}

	newSignature := signatureLine[:closingParen] + paramStr + signatureLine[closingParen:]
	lines[loc.LineStart-1] = newSignature

	return lines, nil
}

func (e *ModifyEngine) insertCode(lines []string, loc *CodeLocation, mod Modification) ([]string, error) {
	code := mod.Code

	insertLine := loc.LineEnd
	if insertLine == 0 || insertLine < loc.LineStart {
		insertLine = e.findSymbolEnd(lines, loc.LineStart)
	}

	position := mod.Position
	if position == "" {
		position = "after"
	}

	var insertAt int
	switch position {
	case "before":
		insertAt = loc.LineStart - 1
	case "inside":
		insertAt = loc.LineStart
	case "after":
		insertAt = insertLine
	default:
		insertAt = insertLine
	}

	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(lines) {
		insertAt = len(lines)
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertAt]...)
	newLines = append(newLines, code)
	newLines = append(newLines, lines[insertAt:]...)

	return newLines, nil
}

func (e *ModifyEngine) replaceCode(lines []string, loc *CodeLocation, mod Modification) ([]string, error) {
	if mod.Type == "rename" && mod.OldValue != "" && mod.NewValue != "" {
		for i, line := range lines {
			lines[i] = strings.ReplaceAll(line, mod.OldValue, mod.NewValue)
		}
		return lines, nil
	}

	if mod.Code != "" {
		if loc.LineEnd == 0 || loc.LineEnd < loc.LineStart {
			loc.LineEnd = e.findSymbolEnd(lines, loc.LineStart)
		}

		newLines := make([]string, 0, len(lines)-(loc.LineEnd-loc.LineStart)+1)
		newLines = append(newLines, lines[:loc.LineStart-1]...)
		newLines = append(newLines, mod.Code)
		if loc.LineEnd < len(lines) {
			newLines = append(newLines, lines[loc.LineEnd:]...)
		}
		return newLines, nil
	}

	return lines, nil
}

func (e *ModifyEngine) findSymbolEnd(lines []string, startLine int) int {
	if startLine <= 0 || startLine > len(lines) {
		return startLine
	}

	braceCount := 0
	foundOpenBrace := false

	for i := startLine - 1; i < len(lines); i++ {
		line := lines[i]
		for _, ch := range line {
			if ch == '{' {
				braceCount++
				foundOpenBrace = true
			} else if ch == '}' {
				braceCount--
				if foundOpenBrace && braceCount == 0 {
					return i + 1
				}
			}
		}
	}

	return startLine
}

func (e *ModifyEngine) deleteCode(lines []string, loc *CodeLocation) ([]string, error) {
	if loc.LineEnd == 0 || loc.LineEnd < loc.LineStart {
		loc.LineEnd = e.findSymbolEnd(lines, loc.LineStart)
	}

	if loc.LineStart > len(lines) {
		return nil, fmt.Errorf("line number out of range")
	}

	newLines := make([]string, 0, len(lines)-(loc.LineEnd-loc.LineStart+1))
	newLines = append(newLines, lines[:loc.LineStart-1]...)
	if loc.LineEnd < len(lines) {
		newLines = append(newLines, lines[loc.LineEnd:]...)
	}

	return newLines, nil
}

func (e *ModifyEngine) AnalyzeImpact(symbolName string) (*types.ImpactAnalysis, error) {
	ids := e.index.GetByName(symbolName)
	if len(ids) == 0 {
		return nil, fmt.Errorf("symbol not found: %s", symbolName)
	}

	return e.graph.AnalyzeImpact(ids[0], 3), nil
}

type CodeGenerator struct {
	parser *parser.CodeParser
}

func NewCodeGenerator(p *parser.CodeParser) *CodeGenerator {
	return &CodeGenerator{
		parser: p,
	}
}

func (g *CodeGenerator) GenerateFunction(name, returnType string, params []types.Parameter, body string) string {
	paramStrs := make([]string, len(params))
	for i, param := range params {
		paramStr := param.Name
		if param.Type != "" {
			paramStr += " " + param.Type
		}
		paramStrs[i] = paramStr
	}

	signature := fmt.Sprintf("func %s(%s)", name, strings.Join(paramStrs, ", "))
	if returnType != "" {
		signature += " " + returnType
	}

	return fmt.Sprintf("%s {\n%s\n}", signature, body)
}

func (g *CodeGenerator) GenerateClass(name string, fields []types.Parameter, methods []string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("type %s struct {\n", name))
	for _, field := range fields {
		builder.WriteString(fmt.Sprintf("    %s %s\n", field.Name, field.Type))
	}
	builder.WriteString("}\n\n")

	for _, method := range methods {
		builder.WriteString(method)
		builder.WriteString("\n")
	}

	return builder.String()
}

func (g *CodeGenerator) GenerateTest(functionName string, testCases []map[string]interface{}) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("func Test%s(t *testing.T) {\n", strings.Title(functionName)))

	for i, tc := range testCases {
		builder.WriteString(fmt.Sprintf("    t.Run(\"test_case_%d\", func(t *testing.T) {\n", i+1))

		if input, ok := tc["input"]; ok {
			builder.WriteString(fmt.Sprintf("        input := %v\n", input))
		}

		if expected, ok := tc["expected"]; ok {
			builder.WriteString(fmt.Sprintf("        expected := %v\n", expected))
		}

		builder.WriteString(fmt.Sprintf("        result := %s(input)\n", functionName))
		builder.WriteString("        if result != expected {\n")
		builder.WriteString("            t.Errorf(\"expected %v, got %v\", expected, result)\n")
		builder.WriteString("        }\n")
		builder.WriteString("    })\n")
	}

	builder.WriteString("}\n")

	return builder.String()
}

func (g *CodeGenerator) FormatCode(code string, language string) string {
	lines := strings.Split(code, "\n")
	formatted := make([]string, len(lines))

	for i, line := range lines {
		formatted[i] = strings.TrimRight(line, " \t")
	}

	return strings.Join(formatted, "\n")
}

func (g *CodeGenerator) SaveToFile(code, filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
