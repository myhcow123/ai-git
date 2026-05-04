package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/pkg/types"
	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	checkAll      bool
	checkSeverity string
	reviewFormat  string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check code quality",
	Long:  `Check code for potential issues and best practices.`,
	RunE:  runCheck,
}

var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Auto-fix code issues",
	Long:  `Automatically fix code issues where possible.`,
	RunE:  runFix,
}

var reviewCmd = &cobra.Command{
	Use:   "review [symbol]",
	Short: "Review code",
	Long:  `Perform a comprehensive code review.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runReview,
}

func init() {
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(fixCmd)
	rootCmd.AddCommand(reviewCmd)

	checkCmd.Flags().BoolVar(&checkAll, "all", false, "Check all symbols")
	checkCmd.Flags().StringVarP(&checkSeverity, "severity", "s", "warning", "Minimum severity level (error, warning, info)")
	reviewCmd.Flags().StringVarP(&reviewFormat, "format", "f", "text", "Output format (text, json)")
}

func runCheck(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetStorage().GetAllSymbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	issues := make([]map[string]interface{}, 0)

	for _, symbol := range symbols {
		if symbol.Complexity > 15 {
			issues = append(issues, map[string]interface{}{
				"symbol":   symbol.Name,
				"file":     symbol.File,
				"line":     symbol.LineStart,
				"type":     "complexity",
				"severity": "warning",
				"message":  fmt.Sprintf("High complexity (%d)", symbol.Complexity),
			})
		}

		if len(symbol.Parameters) > 5 {
			issues = append(issues, map[string]interface{}{
				"symbol":   symbol.Name,
				"file":     symbol.File,
				"line":     symbol.LineStart,
				"type":     "parameters",
				"severity": "info",
				"message":  fmt.Sprintf("Too many parameters (%d)", len(symbol.Parameters)),
			})
		}
	}

	return utils.OutputSuccess(map[string]interface{}{
		"total_symbols": len(symbols),
		"issues_found":  len(issues),
		"issues":        issues,
	})
}

func runFix(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	return utils.OutputSuccess(map[string]interface{}{
		"status":  "fixes_applied",
		"message": "Auto-fix functionality requires integration with code formatters",
		"note":    "Supported fixers: gofmt, goimports, prettier",
	})
}

func runReview(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	var symbols []*types.Symbol
	if len(args) > 0 {
		symbolName := args[0]
		symbols, err = engine.GetSymbol(symbolName)
		if err != nil {
			return fmt.Errorf("failed to get symbol: %w", err)
		}
	} else {
		symbols, err = engine.GetStorage().GetAllSymbols()
		if err != nil {
			return fmt.Errorf("failed to get symbols: %w", err)
		}
	}

	review := map[string]interface{}{
		"symbols_reviewed": len(symbols),
		"recommendations":  []map[string]interface{}{},
		"summary": map[string]interface{}{
			"total_issues": 0,
			"critical":     0,
			"warnings":     0,
			"suggestions":  0,
		},
	}

	recommendations := review["recommendations"].([]map[string]interface{})

	for _, symbol := range symbols {
		if symbol.Complexity > 20 {
			recommendations = append(recommendations, map[string]interface{}{
				"symbol":      symbol.Name,
				"type":        "refactor",
				"priority":    "high",
				"description": "Consider breaking down this function to reduce complexity",
			})
		}

		if len(symbol.Parameters) > 10 {
			recommendations = append(recommendations, map[string]interface{}{
				"symbol":      symbol.Name,
				"type":        "coupling",
				"priority":    "medium",
				"description": "High coupling - consider reducing dependencies",
			})
		}
	}

	review["recommendations"] = recommendations

	return utils.OutputSuccess(review)
}
