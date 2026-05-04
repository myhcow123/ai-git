package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/pkg/types"
	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	buildOutput   string
	buildPlatform string
	deployEnv     string
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
	Long:  `Build the project for deployment.`,
	RunE:  runBuild,
}

var deployCheckCmd = &cobra.Command{
	Use:   "deploy-check",
	Short: "Check deployment readiness",
	Long:  `Check if the project is ready for deployment.`,
	RunE:  runDeployCheck,
}

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(deployCheckCmd)

	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "Output directory")
	buildCmd.Flags().StringVarP(&buildPlatform, "platform", "p", "", "Target platform (linux, darwin, windows)")
	deployCheckCmd.Flags().StringVarP(&deployEnv, "env", "e", "production", "Deployment environment")
}

func runBuild(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	overview, err := engine.GetOverview()
	if err != nil {
		return fmt.Errorf("failed to get overview: %w", err)
	}

	buildInfo := map[string]interface{}{
		"project":   overview["project"],
		"languages": overview["languages"],
		"platform":  buildPlatform,
		"output":    buildOutput,
		"status":    "build_preview",
		"message":   "Build command requires integration with build tools",
		"commands":  []string{},
	}

	if buildPlatform != "" {
		buildInfo["commands"] = append(buildInfo["commands"].([]string),
			fmt.Sprintf("GOOS=%s GOARCH=amd64 go build", buildPlatform))
	} else {
		buildInfo["commands"] = append(buildInfo["commands"].([]string), "go build")
	}

	return utils.OutputSuccess(buildInfo)
}

func runDeployCheck(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	symbols, err := engine.GetStorage().GetAllSymbols()
	if err != nil {
		return fmt.Errorf("failed to get symbols: %w", err)
	}

	checks := []map[string]interface{}{}

	hasTests := false
	for _, symbol := range symbols {
		if containsTest(symbol.Name) {
			hasTests = true
			break
		}
	}

	checks = append(checks, map[string]interface{}{
		"name":    "tests",
		"status":  boolToStr(hasTests),
		"message": "Project has tests",
	})

	hasMain := false
	for _, symbol := range symbols {
		if symbol.Name == "main" && symbol.Type == types.SymbolFunction {
			hasMain = true
			break
		}
	}

	checks = append(checks, map[string]interface{}{
		"name":    "entry_point",
		"status":  boolToStr(hasMain),
		"message": "Entry point (main function) exists",
	})

	highComplexity := 0
	for _, symbol := range symbols {
		if symbol.Complexity > 20 {
			highComplexity++
		}
	}

	checks = append(checks, map[string]interface{}{
		"name":    "code_quality",
		"status":  boolToStr(highComplexity == 0),
		"message": fmt.Sprintf("%d high complexity functions", highComplexity),
	})

	ready := hasTests && hasMain && highComplexity == 0

	return utils.OutputSuccess(map[string]interface{}{
		"environment":   deployEnv,
		"ready":         ready,
		"checks":        checks,
		"total_checks":  len(checks),
		"passed_checks": countPassed(checks),
		"recommendations": []string{
			"Ensure all tests pass",
			"Review high complexity functions",
			"Update documentation",
			"Check environment variables",
		},
	})
}

func boolToStr(b bool) string {
	if b {
		return "passed"
	}
	return "failed"
}

func countPassed(checks []map[string]interface{}) int {
	count := 0
	for _, check := range checks {
		if check["status"] == "passed" {
			count++
		}
	}
	return count
}
