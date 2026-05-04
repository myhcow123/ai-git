package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var explainDetail bool

var explainCmd = &cobra.Command{
	Use:   "explain <symbol>",
	Short: "Explain code functionality",
	Long:  `Explain what a symbol (function, class, etc.) does in natural language.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExplain,
}

func init() {
	rootCmd.AddCommand(explainCmd)
	explainCmd.Flags().BoolVarP(&explainDetail, "detail", "d", false, "Show detailed explanation")
}

func runExplain(cmd *cobra.Command, args []string) error {
	symbolName := args[0]

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetSymbol(symbolName)
	if err != nil {
		return fmt.Errorf("failed to get symbol: %w", err)
	}

	if len(symbols) == 0 {
		return fmt.Errorf("symbol not found: %s", symbolName)
	}

	symbol := symbols[0]

	intent := engine.GetSemanticEngine().InferIntent(symbol)

	pattern := "unknown"

	explanation := map[string]interface{}{
		"symbol":     symbolName,
		"type":       symbol.Type,
		"file":       symbol.File,
		"lines":      fmt.Sprintf("%d-%d", symbol.LineStart, symbol.LineEnd),
		"signature":  symbol.Signature,
		"intent":     intent,
		"pattern":    pattern,
		"complexity": symbol.Complexity,
	}

	if explainDetail {
		explanation["code"] = symbol.Code
		explanation["parameters"] = symbol.Parameters
	}

	return utils.OutputSuccess(explanation)
}
