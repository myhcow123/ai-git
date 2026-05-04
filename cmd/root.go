package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ai-git",
		Short: "AI-Native Code Management System",
		Long:  `AI-Git is a code management system designed specifically for AI models.`,
	}
	formatFlag     string
	withFlag       []string
	symbolTypeFlag string
	regexFlag      bool
	apiFlag        bool
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(symbolCmd)
	rootCmd.AddCommand(overviewCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(statusCmd)

	rootCmd.PersistentFlags().BoolVar(&apiFlag, "api", false, "Use background API service if available")

	symbolCmd.Flags().StringSliceVar(&withFlag, "with", []string{}, "Additional information (callers, history, dependencies)")
	symbolCmd.Flags().StringVar(&symbolTypeFlag, "type", "", "Filter by symbol type (function, class, method, variable, interface, struct)")
	searchCmd.Flags().StringVar(&formatFlag, "format", "json", "Output format (json, text, compact)")
	searchCmd.Flags().BoolVar(&regexFlag, "regex", false, "Enable regular expression search")
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new AI-Git project",
	Long:  `Initialize a new AI-Git project and build the initial index.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

var symbolCmd = &cobra.Command{
	Use:   "symbol <name>",
	Short: "Query symbol information",
	Long:  `Query detailed information about a symbol (function, class, etc.).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSymbol,
}

var overviewCmd = &cobra.Command{
	Use:   "overview",
	Short: "Show project overview",
	Long:  `Display a comprehensive overview of the project structure.`,
	RunE:  runOverview,
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for symbols",
	Long:  `Search for symbols by name, type, or semantic meaning.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSearch,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project status",
	Long:  `Display the current status of the project.`,
	RunE:  runStatus,
}

func runInit(cmd *cobra.Command, args []string) error {
	if err := ensureAPIRunning(); err != nil {
		return err
	}

	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	absPath, err := GetAbsPath(path)
	if err != nil {
		return err
	}

	client := getAPIClient()
	result, err := client.Init(absPath)
	if err != nil {
		return fmt.Errorf("failed to initialize project: %w", err)
	}

	return utils.OutputSuccess(result)
}

func runSymbol(cmd *cobra.Command, args []string) error {
	if err := ensureAPIRunning(); err != nil {
		return err
	}

	name := args[0]

	client := getAPIClient()
	result, err := client.Read(name)
	if err != nil {
		return fmt.Errorf("failed to read symbol: %w", err)
	}

	return utils.OutputSuccess(result)
}

func runOverview(cmd *cobra.Command, args []string) error {
	if err := ensureAPIRunning(); err != nil {
		return err
	}

	client := getAPIClient()
	result, err := client.Get("/overview")
	if err != nil {
		return fmt.Errorf("failed to get overview: %w", err)
	}

	return utils.OutputSuccess(result)
}

func runSearch(cmd *cobra.Command, args []string) error {
	if err := ensureAPIRunning(); err != nil {
		return err
	}

	query := args[0]

	client := getAPIClient()
	result, err := client.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search: %w", err)
	}

	return utils.OutputSuccess(result)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if err := ensureAPIRunning(); err != nil {
		return err
	}

	client := getAPIClient()
	result, err := client.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	return utils.OutputSuccess(result)
}
