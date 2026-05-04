package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mychow/ai-git/internal/query"
	"github.com/spf13/cobra"
)

var (
	queryDir      string
	queryCategory string
	queryVerbose  bool
	querySearch   string
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Manage and execute AI-Git queries",
	Long:  `Manage and execute predefined AI-Git queries (AQL templates).`,
}

var queryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available queries",
	Long:  `List all available predefined queries.`,
	RunE:  runQueryList,
}

var queryShowCmd = &cobra.Command{
	Use:   "show <query-name>",
	Short: "Show query details",
	Long:  `Show detailed information about a specific query.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runQueryShow,
}

var querySearchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search queries",
	Long:  `Search queries by name, description, or category.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runQuerySearch,
}

var queryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List query categories",
	Long:  `List all available query categories.`,
	RunE:  runQueryCategories,
}

var queryExecuteCmd = &cobra.Command{
	Use:   "exec <query-name>",
	Short: "Execute a query",
	Long:  `Execute a predefined query and return results.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runQueryExecute,
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.AddCommand(queryListCmd)
	queryCmd.AddCommand(queryShowCmd)
	queryCmd.AddCommand(querySearchCmd)
	queryCmd.AddCommand(queryCategoriesCmd)
	queryCmd.AddCommand(queryExecuteCmd)

	queryCmd.PersistentFlags().StringVarP(&queryDir, "dir", "d", "", "Query directory path")
	queryListCmd.Flags().StringVarP(&queryCategory, "category", "c", "", "Filter by category")
	queryListCmd.Flags().BoolVarP(&queryVerbose, "verbose", "v", false, "Show detailed information")
	querySearchCmd.Flags().BoolVarP(&queryVerbose, "verbose", "v", false, "Show detailed information")
}

func getQueryManager() (*query.Manager, error) {
	dir := queryDir
	if dir == "" {
		execPath, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("failed to get executable path: %w", err)
		}
		dir = filepath.Join(filepath.Dir(execPath), "queries")

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			homeDir, _ := os.UserHomeDir()
			dir = filepath.Join(homeDir, ".ai-git", "queries")
		}
	}

	mgr := query.NewManager(dir)
	if err := mgr.Load(); err != nil {
		return nil, fmt.Errorf("failed to load queries: %w", err)
	}

	return mgr, nil
}

func runQueryList(cmd *cobra.Command, args []string) error {
	mgr, err := getQueryManager()
	if err != nil {
		return err
	}

	var queries []*query.QueryTemplate
	if queryCategory != "" {
		queries = mgr.ListByCategory(queryCategory)
	} else {
		queries = mgr.List()
	}

	if len(queries) == 0 {
		fmt.Println("No queries found")
		return nil
	}

	output := query.FormatQueryList(queries, queryVerbose)
	fmt.Println(output)

	if !queryVerbose {
		fmt.Printf("\nTotal: %d queries\n", len(queries))
		fmt.Println("Use 'ai-git query show <name>' for details")
		fmt.Println("Use 'ai-git query list -v' for verbose output")
	}

	return nil
}

func runQueryShow(cmd *cobra.Command, args []string) error {
	mgr, err := getQueryManager()
	if err != nil {
		return err
	}

	name := args[0]
	queryTemplate, err := mgr.Get(name)
	if err != nil {
		return err
	}

	fmt.Printf("Name: %s\n", queryTemplate.Name)
	fmt.Printf("Category: %s\n", queryTemplate.Category)
	fmt.Printf("Description: %s\n", queryTemplate.Description)
	fmt.Printf("\nQuery:\n  %s\n", queryTemplate.Query)

	if len(queryTemplate.Examples) > 0 {
		fmt.Println("\nExamples:")
		for i, example := range queryTemplate.Examples {
			fmt.Printf("  %d. %s\n", i+1, example)
		}
	}

	if len(queryTemplate.UseCases) > 0 {
		fmt.Println("\nUse Cases:")
		for i, uc := range queryTemplate.UseCases {
			fmt.Printf("  %d. %s\n", i+1, uc)
		}
	}

	return nil
}

func runQuerySearch(cmd *cobra.Command, args []string) error {
	mgr, err := getQueryManager()
	if err != nil {
		return err
	}

	keyword := args[0]
	queries := mgr.Search(keyword)

	if len(queries) == 0 {
		fmt.Printf("No queries found matching '%s'\n", keyword)
		return nil
	}

	output := query.FormatQueryList(queries, queryVerbose)
	fmt.Println(output)

	fmt.Printf("\nFound %d queries matching '%s'\n", len(queries), keyword)

	return nil
}

func runQueryCategories(cmd *cobra.Command, args []string) error {
	mgr, err := getQueryManager()
	if err != nil {
		return err
	}

	categories := mgr.GetCategories()
	stats := mgr.GetStats()

	fmt.Println("Query Categories:")
	fmt.Println("=================")

	for _, cat := range categories {
		count := stats["by_category"].(map[string]int)[cat]
		fmt.Printf("  %s (%d queries)\n", cat, count)
	}

	fmt.Printf("\nTotal: %d categories, %d queries\n", len(categories), stats["total_queries"])

	return nil
}

func runQueryExecute(cmd *cobra.Command, args []string) error {
	mgr, err := getQueryManager()
	if err != nil {
		return err
	}

	name := args[0]
	queryStr, err := mgr.Execute(name, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Executing query: %s\n", name)
	fmt.Printf("Query: %s\n\n", queryStr)

	fmt.Println("Note: Query execution requires integration with the AQL parser.")
	fmt.Println("This feature will execute the query and return results.")

	return nil
}
