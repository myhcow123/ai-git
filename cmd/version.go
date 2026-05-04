package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	saveMessage  string
	rollbackID   string
	historyLimit int
	diffFormat   string
)

var saveCmd = &cobra.Command{
	Use:   "save [message]",
	Short: "Save current project state",
	Long:  `Save the current state of the project with an optional message.`,
	RunE:  runSave,
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback [state-id]",
	Short: "Rollback to a previous state",
	Long:  `Rollback the project to a previous saved state.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRollback,
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show project history",
	Long:  `Display the history of saved project states.`,
	RunE:  runHistory,
}

var diffCmd = &cobra.Command{
	Use:   "diff <state1> <state2>",
	Short: "Compare two project states",
	Long:  `Compare two saved project states and show differences.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func init() {
	rootCmd.AddCommand(saveCmd)
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(diffCmd)

	saveCmd.Flags().StringVarP(&saveMessage, "message", "m", "", "Save message")
	rollbackCmd.Flags().StringVar(&rollbackID, "id", "", "State ID to rollback to")
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 10, "Number of history entries to show")
	diffCmd.Flags().StringVarP(&diffFormat, "format", "f", "text", "Output format (text, json)")
}

func runSave(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		saveMessage = args[0]
	}

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	snapshot, err := engine.GetStorage().SaveSnapshotWithMessage(saveMessage)
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"state_id":  snapshot.ID,
		"message":   saveMessage,
		"timestamp": snapshot.Timestamp,
		"status":    "saved",
	})
}

func runRollback(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		rollbackID = args[0]
	}

	if rollbackID == "" {
		return fmt.Errorf("state ID is required")
	}

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	if err := engine.GetStorage().RollbackToSnapshot(rollbackID); err != nil {
		return fmt.Errorf("failed to rollback: %w", err)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"state_id": rollbackID,
		"status":   "rolled back",
		"message":  "Project rolled back to specified state",
	})
}

func runHistory(cmd *cobra.Command, args []string) error {
	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	snapshots, err := engine.GetStorage().GetSnapshotHistory()
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	history := make([]map[string]interface{}, 0, len(snapshots))
	for _, snap := range snapshots {
		history = append(history, map[string]interface{}{
			"id":        snap.ID,
			"timestamp": snap.Timestamp,
			"message":   snap.Message,
			"symbols":   len(snap.Symbols),
		})
	}

	return utils.OutputSuccess(map[string]interface{}{
		"count":   len(history),
		"history": history,
	})
}

func runDiff(cmd *cobra.Command, args []string) error {
	state1 := args[0]
	state2 := args[1]

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	diff, err := engine.GetStorage().DiffSnapshots(state1, state2)
	if err != nil {
		return fmt.Errorf("failed to diff states: %w", err)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"from":             state1,
		"to":               state2,
		"symbols_added":    diff["added"],
		"symbols_removed":  diff["removed"],
		"symbols_modified": diff["modified"],
	})
}
