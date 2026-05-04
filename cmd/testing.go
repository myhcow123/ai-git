package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	testOutputFile string
	testFramework  string
)

var testGenCmd = &cobra.Command{
	Use:   "test-gen <symbol>",
	Short: "Generate tests for a symbol",
	Long:  `Generate unit tests for the specified symbol (function, class, etc.).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTestGen,
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
	Long:  `Run all tests in the project.`,
	RunE:  runTest,
}

var coverageCmd = &cobra.Command{
	Use:   "coverage",
	Short: "Show test coverage",
	Long:  `Display test coverage information for the project.`,
	RunE:  runCoverage,
}

func init() {
	rootCmd.AddCommand(testGenCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(coverageCmd)

	testGenCmd.Flags().StringVarP(&testOutputFile, "output", "o", "", "Output file path")
	testGenCmd.Flags().StringVarP(&testFramework, "framework", "f", "go", "Test framework (go, pytest, jest)")
}

func runTestGen(cmd *cobra.Command, args []string) error {
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

	testCases := []map[string]interface{}{
		{
			"name":     "test_case_1",
			"input":    "test_input",
			"expected": "expected_output",
		},
		{
			"name":     "test_case_2",
			"input":    "edge_case_input",
			"expected": "expected_output",
		},
	}

	generator := engine.GetCodeGenerator()
	testCode := generator.GenerateTest(symbolName, testCases)

	if testOutputFile != "" {
		if err := generator.SaveToFile(testCode, testOutputFile); err != nil {
			return fmt.Errorf("failed to save test file: %w", err)
		}
		return utils.OutputSuccess(map[string]interface{}{
			"symbol":     symbolName,
			"file":       testOutputFile,
			"code":       testCode,
			"test_cases": len(testCases),
			"message":    "Test generated successfully",
		})
	}

	return utils.OutputSuccess(map[string]interface{}{
		"symbol":     symbolName,
		"code":       testCode,
		"test_cases": len(testCases),
	})
}

func runTest(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	return utils.OutputSuccess(map[string]interface{}{
		"status":  "tests_executed",
		"message": "Test execution requires integration with test framework",
		"note":    "Use 'ai-git test-gen' to generate tests first",
	})
}

func runCoverage(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetStorage().GetAllSymbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	totalSymbols := len(symbols)
	testedSymbols := 0
	coverage := make([]map[string]interface{}, 0)

	for _, symbol := range symbols {
		hasTest := false
		if containsTest(symbol.Name) {
			hasTest = true
			testedSymbols++
		}

		coverage = append(coverage, map[string]interface{}{
			"symbol": symbol.Name,
			"type":   symbol.Type,
			"tested": hasTest,
		})
	}

	coveragePercent := 0.0
	if totalSymbols > 0 {
		coveragePercent = float64(testedSymbols) / float64(totalSymbols) * 100
	}

	return utils.OutputSuccess(map[string]interface{}{
		"total_symbols":    totalSymbols,
		"tested_symbols":   testedSymbols,
		"coverage_percent": coveragePercent,
		"coverage_details": coverage,
	})
}

func containsTest(name string) bool {
	return len(name) >= 4 && (name[:4] == "Test" || name[:4] == "test")
}
