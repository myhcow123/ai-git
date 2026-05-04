package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mychow/ai-git/internal/api"
	"github.com/mychow/ai-git/internal/aql"
	"github.com/mychow/ai-git/internal/config"
	"github.com/spf13/cobra"
)

var (
	webPort    int
	configFile string
	daemonMode bool
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start AI-Git web server",
	Long:  `Start AI-Git web server with REST API and optional web interface.`,
	RunE:  runWeb,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage AI-Git configuration",
	Long:  `Manage AI-Git configuration file and settings.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate sample configuration file",
	Long:  `Generate a sample configuration file with default settings.`,
	RunE:  runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration settings.`,
	RunE:  runConfigShow,
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage AI-Git plugins",
	Long:  `Manage AI-Git plugins including listing, loading, and executing plugins.`,
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List loaded plugins",
	Long:  `List all currently loaded plugins.`,
	RunE:  runPluginList,
}

func init() {
	rootCmd.AddCommand(webCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pluginCmd)

	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Web server port")
	webCmd.Flags().StringVarP(&configFile, "config", "c", "", "Configuration file path")
	webCmd.Flags().BoolVar(&daemonMode, "daemon", false, "Run as daemon (background service)")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configInitCmd.Flags().StringVarP(&configFile, "output", "o", ".ai-git.json", "Output configuration file path")

	pluginCmd.AddCommand(pluginListCmd)
}

func runWeb(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		defaultCfg := config.DefaultConfig
		cfg = &defaultCfg
	}

	engine, err := initEngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize engine: %w", err)
	}

	server := api.NewServer(engine, webPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		if !daemonMode {
			fmt.Println("\nShutting down server...")
		}
		server.Stop()
		os.Exit(0)
	}()

	if !daemonMode {
		fmt.Printf("Starting AI-Git API server on port %d...\n", webPort)
		fmt.Printf("API endpoint: http://localhost:%d/api/v1/\n", webPort)
		fmt.Println("Press Ctrl+C to stop")
	}

	return server.Start()
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	outputPath, _ := cmd.Flags().GetString("output")

	if err := config.GenerateSampleConfig(outputPath); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	fmt.Printf("Configuration file generated: %s\n", outputPath)
	fmt.Println("Edit this file to customize AI-Git settings")
	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("Current AI-Git Configuration:")
	fmt.Println("============================")
	fmt.Printf("Project Root: %s\n", cfg.Project.Root)
	fmt.Printf("Database: %s\n", cfg.Storage.Database)
	fmt.Printf("Languages: %v\n", cfg.Project.Languages)
	fmt.Printf("Ignore Dirs: %v\n", cfg.Project.IgnoreDirs)
	fmt.Printf("Parallel Workers: %d\n", cfg.Performance.ParallelWorkers)
	fmt.Printf("Cache Enabled: %v\n", cfg.Performance.EnableCache)
	fmt.Printf("Vector Index: %v\n", cfg.Performance.EnableVector)
	fmt.Printf("Output Format: %s\n", cfg.Output.Format)
	return nil
}

func runPluginList(cmd *cobra.Command, args []string) error {
	fmt.Println("Loaded Plugins:")
	fmt.Println("===============")
	fmt.Println("No plugins loaded")
	fmt.Println("\nTo load plugins:")
	fmt.Println("  1. Place .so files in ~/.ai-git/plugins/")
	fmt.Println("  2. Or use: ai-git plugin load <path>")
	return nil
}

func initEngine(cfg *config.Config) (api.Engine, error) {
	engine, err := GetEngine(cfg.Project.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize engine: %w", err)
	}
	return &engineWrapper{engine: engine}, nil
}

type engineWrapper struct {
	engine *aql.Engine
}

func (e *engineWrapper) GetStorage() interface{}  { return e.engine.GetStorage() }
func (e *engineWrapper) GetIndex() interface{}    { return e.engine.GetIndex() }
func (e *engineWrapper) GetGraph() interface{}    { return e.engine.GetGraph() }
func (e *engineWrapper) GetParser() interface{}   { return e.engine.GetParser() }
func (e *engineWrapper) GetEngine() interface{}   { return e.engine }
