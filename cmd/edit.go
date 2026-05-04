package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <symbol|file:line>",
	Short: "Read symbol code or file lines",
	Long: `Read the complete code of a symbol or specific lines from a file.

Examples:
  ai-git read Initialize              # Read symbol code
  ai-git read main.go:10-20           # Read lines 10-20
  ai-git read main.go:10              # Read line 10`,
	Args: cobra.ExactArgs(1),
	RunE:  runRead,
}

var replaceCmd = &cobra.Command{
	Use:   "replace <symbol|file:line-range> --with <code>",
	Short: "Replace code and auto-save",
	Long: `Replace symbol code or file lines with new code, then auto-save.

Examples:
  ai-git replace Initialize --with "func NewFunc() {}"
  ai-git replace main.go:10-20 --with "// replaced"`,
	Args: cobra.ExactArgs(1),
	RunE:  runReplace,
}

var insertCmd = &cobra.Command{
	Use:   "insert <symbol|file:line> --code <code>",
	Short: "Insert code and auto-save",
	Long: `Insert code at specified location and auto-save.

Examples:
  ai-git insert Initialize --before --code "// comment"
  ai-git insert Initialize --after --code "fmt.Println(\"done\")"
  ai-git insert main.go:10 --code "// new line"`,
	Args: cobra.ExactArgs(1),
	RunE:  runInsert,
}

var deleteCmd = &cobra.Command{
	Use:   "delete <symbol|file:line-range>",
	Short: "Delete code and auto-save",
	Long: `Delete symbol code or file lines and auto-save.

Examples:
  ai-git delete Initialize
  ai-git delete main.go:10-20`,
	Args: cobra.ExactArgs(1),
	RunE:  runDelete,
}

var (
	replaceWith string
	insertCode  string
	insertBefore bool
	insertAfter  bool
)

func init() {
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(replaceCmd)
	rootCmd.AddCommand(insertCmd)
	rootCmd.AddCommand(deleteCmd)

	replaceCmd.Flags().StringVar(&replaceWith, "with", "", "New code to replace with")
	_ = replaceCmd.MarkFlagRequired("with")

	insertCmd.Flags().StringVar(&insertCode, "code", "", "Code to insert")
	insertCmd.Flags().BoolVar(&insertBefore, "before", false, "Insert before symbol")
	insertCmd.Flags().BoolVar(&insertAfter, "after", true, "Insert after symbol (default)")
	_ = insertCmd.MarkFlagRequired("code")
}

func runRead(cmd *cobra.Command, args []string) error {
	target := args[0]
	if strings.Contains(target, ":") {
		return readFileLines(target)
	}
	return readSymbol(target)
}

func runReplace(cmd *cobra.Command, args []string) error {
	target := args[0]
	if strings.Contains(target, ":") {
		return replaceFileLines(target, replaceWith)
	}
	return replaceSymbol(target, replaceWith)
}

func runInsert(cmd *cobra.Command, args []string) error {
	target := args[0]
	if strings.Contains(target, ":") {
		return insertAtLine(target, insertCode)
	}
	return insertAtSymbol(target, insertCode, insertBefore)
}

func runDelete(cmd *cobra.Command, args []string) error {
	target := args[0]
	if strings.Contains(target, ":") {
		return deleteFileLines(target)
	}
	return deleteSymbol(target)
}

type LineRange struct {
	Start int
	End   int
}

func parseLineSpec(spec string) LineRange {
	if strings.Contains(spec, "-") {
		parts := strings.Split(spec, "-")
		start, _ := strconv.Atoi(parts[0])
		end, _ := strconv.Atoi(parts[1])
		return LineRange{Start: start, End: end}
	}
	line, _ := strconv.Atoi(spec)
	return LineRange{Start: line, End: line}
}

func resolveFilePath(path string) string {
	if !filepath.IsAbs(path) {
		wd, _ := os.Getwd()
		return filepath.Join(wd, path)
	}
	return path
}

func readFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	return string(content), nil
}

func saveFileContent(path string, content string) error {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	return nil
}

func modifyLines(lines []string, start, end int, newLines []string) []string {
	result := make([]string, 0)
	result = append(result, lines[:start-1]...)
	result = append(result, newLines...)
	result = append(result, lines[end:]...)
	return result
}

func readFileLines(target string) error {
	parts := strings.Split(target, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, use file:line or file:start-end")
	}

	filePath := resolveFilePath(parts[0])
	lineRange := parseLineSpec(parts[1])

	content, err := readFileContent(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(content, "\n")
	start, end := lineRange.Start, lineRange.End

	if start < 1 {
		start = 1
	}
	if end > len(lines) {
		end = len(lines)
	}

	var result strings.Builder
	for i := start; i <= end; i++ {
		result.WriteString(fmt.Sprintf("%d: %s\n", i, lines[i-1]))
	}

	return utils.OutputSuccess(map[string]interface{}{
		"file":  filePath,
		"lines": fmt.Sprintf("%d-%d", start, end),
		"code":  result.String(),
	})
}

func readSymbol(symbolName string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetSymbol(symbolName)
	if err != nil || len(symbols) == 0 {
		return fmt.Errorf("symbol not found: %s", symbolName)
	}

	symbol := symbols[0]
	content, err := engine.GetStorage().GetFile(symbol.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(content, "\n")
	start := symbol.LineStart
	end := symbol.LineEnd
	if end == 0 || end < start {
		end = findSymbolEnd(lines, start-1)
	}

	var code strings.Builder
	for i := start; i <= end && i <= len(lines); i++ {
		code.WriteString(fmt.Sprintf("%d: %s\n", i, lines[i-1]))
	}

	return utils.OutputSuccess(map[string]interface{}{
		"symbol": symbol.Name,
		"type":   symbol.Type.String(),
		"file":   symbol.File,
		"lines":  fmt.Sprintf("%d-%d", start, end),
		"code":   code.String(),
	})
}

func findSymbolEnd(lines []string, startIdx int) int {
	if startIdx >= len(lines) {
		return startIdx + 1
	}

	braceCount := 0
	foundOpen := false

	for i := startIdx; i < len(lines); i++ {
		line := lines[i]
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")

		if strings.Contains(line, "{") {
			foundOpen = true
		}

		if foundOpen && braceCount == 0 {
			return i + 1
		}
	}

	return len(lines)
}

func replaceFileLines(target string, newCode string) error {
	parts := strings.Split(target, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, use file:line or file:start-end")
	}

	filePath := resolveFilePath(parts[0])
	lineRange := parseLineSpec(parts[1])

	originalContent, err := readFileContent(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(originalContent, "\n")
	newLines := modifyLines(lines, lineRange.Start, lineRange.End, strings.Split(newCode, "\n"))
	newContent := strings.Join(newLines, "\n")

	recordModification(filePath, originalContent, newContent)

	if err := saveFileContent(filePath, newContent); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"status":  "replaced",
		"file":    filePath,
		"lines":   fmt.Sprintf("%d-%d", lineRange.Start, lineRange.End),
		"message": "File saved successfully",
	})
}

func replaceSymbol(symbolName string, newCode string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetSymbol(symbolName)
	if err != nil || len(symbols) == 0 {
		return fmt.Errorf("symbol not found: %s", symbolName)
	}

	symbol := symbols[0]
	originalContent, err := engine.GetStorage().GetFile(symbol.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(originalContent, "\n")
	start := symbol.LineStart
	end := symbol.LineEnd
	if end == 0 || end < start {
		end = findSymbolEnd(lines, start-1)
	}

	newLines := modifyLines(lines, start, end, strings.Split(newCode, "\n"))
	newContent := strings.Join(newLines, "\n")

	recordModification(symbol.File, originalContent, newContent)

	if err := saveFileContent(symbol.File, newContent); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"status":  "replaced",
		"symbol":  symbol.Name,
		"file":    symbol.File,
		"lines":   fmt.Sprintf("%d-%d", start, end),
		"message": "File saved successfully",
	})
}

func insertAtLine(target string, code string) error {
	parts := strings.Split(target, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, use file:line")
	}

	filePath := resolveFilePath(parts[0])
	lineNum, _ := strconv.Atoi(parts[1])

	originalContent, err := readFileContent(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(originalContent, "\n")

	if lineNum < 1 {
		lineNum = 1
	}
	if lineNum > len(lines)+1 {
		lineNum = len(lines) + 1
	}

	newLines := modifyLines(lines, lineNum, lineNum-1, strings.Split(code, "\n"))
	newContent := strings.Join(newLines, "\n")

	recordModification(filePath, originalContent, newContent)

	if err := saveFileContent(filePath, newContent); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"status":  "inserted",
		"file":    filePath,
		"at_line": lineNum,
		"message": "File saved successfully",
	})
}

func insertAtSymbol(symbolName string, code string, before bool) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetSymbol(symbolName)
	if err != nil || len(symbols) == 0 {
		return fmt.Errorf("symbol not found: %s", symbolName)
	}

	symbol := symbols[0]
	originalContent, err := engine.GetStorage().GetFile(symbol.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(originalContent, "\n")
	insertLine := symbol.LineStart
	
	if !before {
		end := symbol.LineEnd
		if end == 0 || end < insertLine {
			end = findSymbolEnd(lines, insertLine-1)
		}
		insertLine = end + 1
	}

	newLines := modifyLines(lines, insertLine, insertLine-1, strings.Split(code, "\n"))
	newContent := strings.Join(newLines, "\n")

	recordModification(symbol.File, originalContent, newContent)

	if err := saveFileContent(symbol.File, newContent); err != nil {
		return err
	}

	position := "after"
	if before {
		position = "before"
	}

	return utils.OutputSuccess(map[string]interface{}{
		"status":   "inserted",
		"symbol":   symbol.Name,
		"file":     symbol.File,
		"at_line":  insertLine,
		"position": position,
		"message":  "File saved successfully",
	})
}

func deleteFileLines(target string) error {
	parts := strings.Split(target, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid format, use file:line or file:start-end")
	}

	filePath := resolveFilePath(parts[0])
	lineRange := parseLineSpec(parts[1])

	originalContent, err := readFileContent(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(originalContent, "\n")
	newLines := modifyLines(lines, lineRange.Start, lineRange.End, nil)
	newContent := strings.Join(newLines, "\n")

	recordModification(filePath, originalContent, newContent)

	if err := saveFileContent(filePath, newContent); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"status":  "deleted",
		"file":    filePath,
		"lines":   fmt.Sprintf("%d-%d", lineRange.Start, lineRange.End),
		"message": "File saved successfully",
	})
}

func deleteSymbol(symbolName string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetSymbol(symbolName)
	if err != nil || len(symbols) == 0 {
		return fmt.Errorf("symbol not found: %s", symbolName)
	}

	symbol := symbols[0]
	originalContent, err := engine.GetStorage().GetFile(symbol.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(originalContent, "\n")
	start := symbol.LineStart
	end := symbol.LineEnd
	if end == 0 || end < start {
		end = findSymbolEnd(lines, start-1)
	}

	newLines := modifyLines(lines, start, end, nil)
	newContent := strings.Join(newLines, "\n")

	recordModification(symbol.File, originalContent, newContent)

	if err := saveFileContent(symbol.File, newContent); err != nil {
		return err
	}

	return utils.OutputSuccess(map[string]interface{}{
		"status":  "deleted",
		"symbol":  symbol.Name,
		"file":    symbol.File,
		"lines":   fmt.Sprintf("%d-%d", start, end),
		"message": "File saved successfully",
	})
}
