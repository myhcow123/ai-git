package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo [n]",
	Short: "Undo recent modifications",
	Long: `Undo recent file modifications.

Examples:
  ai-git undo               # Undo last modification
  ai-git undo 3             # Undo last 3 modifications
  ai-git undo --list        # List recent modifications`,
	RunE: runUndo,
}

var usagesCmd = &cobra.Command{
	Use:   "usages <symbol>",
	Short: "Find all usages of a symbol",
	Long: `Find all places where a symbol is used.

Examples:
  ai-git usages functionName
  ai-git usages --type function greet`,
	Args: cobra.ExactArgs(1),
	RunE: runUsages,
}

var (
	undoList   bool
	usagesType string
)

func init() {
	rootCmd.AddCommand(undoCmd)
	rootCmd.AddCommand(usagesCmd)

	undoCmd.Flags().BoolVar(&undoList, "list", false, "List recent modifications")

	usagesCmd.Flags().StringVar(&usagesType, "type", "", "Filter by symbol type")
}

type ModificationRecord struct {
	File     string `json:"file"`
	Original string `json:"original"`
	Modified string `json:"modified"`
	Time     int64  `json:"time"`
}

func runUndo(cmd *cobra.Command, args []string) error {
	if undoList {
		return listModifications()
	}

	n := 1
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &n)
	}

	historyFile := getModificationHistoryFile()
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return fmt.Errorf("no modification history found")
	}

	var history []ModificationRecord
	if err := json.Unmarshal(data, &history); err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	if len(history) == 0 {
		return fmt.Errorf("no modifications to undo")
	}

	if n > len(history) {
		n = len(history)
	}

	undone := 0
	for i := 0; i < n && len(history) > 0; i++ {
		last := history[len(history)-1]
		if err := os.WriteFile(last.File, []byte(last.Original), 0644); err != nil {
			fmt.Printf("Warning: failed to undo %s: %v\n", last.File, err)
			continue
		}
		fmt.Printf("Undone: %s\n", last.File)
		history = history[:len(history)-1]
		undone++
	}

	updatedData, _ := json.Marshal(history)
	os.WriteFile(historyFile, updatedData, 0644)

	return utils.OutputSuccess(map[string]interface{}{
		"status":    "undone",
		"count":     undone,
		"remaining": len(history),
	})
}

func listModifications() error {
	historyFile := getModificationHistoryFile()
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return fmt.Errorf("no modification history found")
	}

	var history []ModificationRecord
	if err := json.Unmarshal(data, &history); err != nil {
		return fmt.Errorf("failed to parse history: %w", err)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"count":  len(history),
		"recent": history,
	})
}

func getModificationHistoryFile() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, ".ai-git", "modifications.json")
}

func recordModification(file, original, modified string) {
	historyFile := getModificationHistoryFile()

	os.MkdirAll(filepath.Dir(historyFile), 0755)

	var history []ModificationRecord
	if data, err := os.ReadFile(historyFile); err == nil {
		json.Unmarshal(data, &history)
	}

	history = append(history, ModificationRecord{
		File:     file,
		Original: original,
		Modified: modified,
		Time:     getCurrentTime(),
	})

	data, _ := json.Marshal(history)
	os.WriteFile(historyFile, data, 0644)
}

func getCurrentTime() int64 {
	return 0
}

func runUsages(cmd *cobra.Command, args []string) error {
	symbolName := args[0]

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	allSymbols, err := engine.GetStorage().GetAllSymbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	usages := []map[string]interface{}{}

	for _, symbol := range allSymbols {
		content, err := engine.GetStorage().GetFile(symbol.File)
		if err != nil {
			continue
		}

		lines := strings.Split(content, "\n")
		for i, line := range lines {
			if strings.Contains(line, symbolName) {
				if symbol.Name == symbolName {
					continue
				}

				usages = append(usages, map[string]interface{}{
					"file":       symbol.File,
					"line":       i + 1,
					"code":       strings.TrimSpace(line),
					"context":    symbol.Name,
					"context_id": symbol.ID,
				})
			}
		}
	}

	return utils.OutputSuccess(map[string]interface{}{
		"symbol":  symbolName,
		"count":   len(usages),
		"usages":  usages,
	})
}
