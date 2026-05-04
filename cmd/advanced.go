package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/internal/graph"
	"github.com/mychow/ai-git/internal/semantic"
	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	topFlag   int
	depthFlag int
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(qualityCmd)
	rootCmd.AddCommand(patternCmd)
	rootCmd.AddCommand(intentCmd)
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(impactCmd)

	analyzeCmd.Flags().IntVarP(&topFlag, "top", "n", 10, "Number of top symbols to show")
	depsCmd.Flags().IntVarP(&depthFlag, "depth", "d", 1, "Dependency depth")
	impactCmd.Flags().IntVarP(&depthFlag, "depth", "d", 3, "Impact analysis depth")
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze code importance using PageRank",
	Long:  `Analyze code importance using PageRank algorithm and show top symbols.`,
	RunE:  runAnalyze,
}

var qualityCmd = &cobra.Command{
	Use:   "quality <symbol>",
	Short: "Assess code quality",
	Long:  `Assess code quality including complexity, testability, maintainability, and security.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runQuality,
}

var patternCmd = &cobra.Command{
	Use:   "pattern <symbol>",
	Short: "Recognize design patterns",
	Long:  `Recognize design patterns in the specified symbol.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runPattern,
}

var intentCmd = &cobra.Command{
	Use:   "intent <symbol>",
	Short: "Infer code intent",
	Long:  `Infer the intent/purpose of the specified symbol.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runIntent,
}

var depsCmd = &cobra.Command{
	Use:   "deps <symbol>",
	Short: "Show dependencies",
	Long:  `Show dependencies and dependents of the specified symbol.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDeps,
}

var impactCmd = &cobra.Command{
	Use:   "impact <symbol>",
	Short: "Analyze modification impact",
	Long:  `Analyze the impact of modifying the specified symbol.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runImpact,
}

func runAnalyze(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetStorage().GetAllSymbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	g := graph.NewSymbolGraph()
	for _, symbol := range symbols {
		g.AddNode(symbol)
	}

	_ = g.PageRank(20, 0.85)

	topSymbols := g.GetTopSymbols(topFlag)

	results := make([]map[string]interface{}, 0, len(topSymbols))
	for _, node := range topSymbols {
		results = append(results, map[string]interface{}{
			"name":       node.Symbol.Name,
			"type":       node.Symbol.Type.String(),
			"file":       node.Symbol.File,
			"line":       node.Symbol.LineStart,
			"importance": node.Weight,
		})
	}

	return utils.OutputSuccess(map[string]interface{}{
		"analysis_type": "pagerank",
		"total_symbols": len(symbols),
		"top_symbols":   results,
	})
}

func runQuality(cmd *cobra.Command, args []string) error {
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
	assessor := semantic.NewQualityAssessor()
	assessment := assessor.Assess(symbol)

	return utils.OutputSuccess(map[string]interface{}{
		"symbol":  symbolName,
		"quality": assessment,
	})
}

func runPattern(cmd *cobra.Command, args []string) error {
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
	recognizer := semantic.NewPatternRecognizer()
	patterns := recognizer.Recognize(symbol)

	return utils.OutputSuccess(map[string]interface{}{
		"symbol":   symbolName,
		"patterns": patterns,
	})
}

func runIntent(cmd *cobra.Command, args []string) error {
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
	intentEngine := semantic.NewIntentEngine()
	intent := intentEngine.InferIntent(symbol)

	return utils.OutputSuccess(map[string]interface{}{
		"symbol": symbolName,
		"intent": intent,
	})
}

func runDeps(cmd *cobra.Command, args []string) error {
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
	graph := engine.GetGraph()

	dependencies := graph.GetCallees(symbol.ID)
	dependents := graph.GetCallers(symbol.ID)

	depNames := make([]string, 0, len(dependencies))
	for _, depID := range dependencies {
		if sym, err := engine.GetStorage().GetSymbol(depID); err == nil {
			depNames = append(depNames, sym.Name)
		}
	}

	dependentNames := make([]string, 0, len(dependents))
	for _, depID := range dependents {
		if sym, err := engine.GetStorage().GetSymbol(depID); err == nil {
			dependentNames = append(dependentNames, sym.Name)
		}
	}

	return utils.OutputSuccess(map[string]interface{}{
		"symbol":       symbolName,
		"dependencies": depNames,
		"dependents":   dependentNames,
		"depth":        depthFlag,
	})
}

func runImpact(cmd *cobra.Command, args []string) error {
	symbolName := args[0]

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetStorage().GetAllSymbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	g := graph.NewSymbolGraph()
	for _, symbol := range symbols {
		g.AddNode(symbol)
	}

	ids := engine.GetIndex().GetByName(symbolName)
	if len(ids) == 0 {
		return fmt.Errorf("symbol not found: %s", symbolName)
	}

	id := ids[0]
	impact := g.AnalyzeImpact(id, depthFlag)

	return utils.OutputSuccess(map[string]interface{}{
		"symbol":          symbolName,
		"impact_analysis": impact,
	})
}
