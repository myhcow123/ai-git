package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mychow/ai-git/internal/aql"
	"github.com/mychow/ai-git/internal/modify"
	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	batchFile   string
	batchDryRun bool
)

var batchCmd = &cobra.Command{
	Use:   "batch <script>",
	Short: "Batch modify code",
	Long:  `Apply multiple modifications from a script file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runBatch,
}

func init() {
	rootCmd.AddCommand(batchCmd)
	batchCmd.Flags().StringVarP(&batchFile, "file", "f", "", "Script file path")
	batchCmd.Flags().BoolVar(&batchDryRun, "dry-run", false, "Preview changes without applying")
}

func runBatch(cmd *cobra.Command, args []string) error {
	var scriptPath string
	if len(args) > 0 {
		scriptPath = args[0]
	} else if batchFile != "" {
		scriptPath = batchFile
	} else {
		return fmt.Errorf("script file path is required")
	}

	content, err := ioutil.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read script file: %w", err)
	}

	var script BatchScript
	if err := json.Unmarshal(content, &script); err != nil {
		return fmt.Errorf("failed to parse script: %w", err)
	}

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	fileBackups := make(map[string]string)
	transactionMode := !batchDryRun

	if transactionMode {
		for _, op := range script.Operations {
			symbols, err := engine.GetSymbol(op.Target)
			if err != nil || len(symbols) == 0 {
				continue
			}
			symbol := symbols[0]
			if _, exists := fileBackups[symbol.File]; !exists {
				fileContent, err := engine.GetStorage().GetFile(symbol.File)
				if err == nil {
					fileBackups[symbol.File] = fileContent
				}
			}
		}
	}

	results := make([]map[string]interface{}, 0)
	hasError := false

	for i, op := range script.Operations {
		result := map[string]interface{}{
			"operation": i + 1,
			"type":      op.Type,
			"target":    op.Target,
		}

		symbols, _ := engine.GetSymbol(op.Target)
		if len(symbols) == 0 {
			result["status"] = "skipped"
			result["message"] = "Symbol not found - operation skipped"
			results = append(results, result)
			continue
		}

		if batchDryRun {
			result["status"] = "dry-run"
			result["message"] = "Preview only - no changes applied"
		} else {
			err := executeBatchOperation(engine, op)
			if err != nil {
				result["status"] = "failed"
				result["error"] = err.Error()
				hasError = true

				if transactionMode {
					for file, backup := range fileBackups {
						_ = engine.GetStorage().SaveFile(file, backup)
					}
					result["rollback"] = "Transaction rolled back due to error"
					break
				}
			} else {
				result["status"] = "success"
				result["message"] = "Operation completed"
				
				fileContent, err := engine.GetStorage().GetFile(symbols[0].File)
				if err == nil {
					engine.GetParser().ParseFile(symbols[0].File, fileContent)
				}
			}
		}

		results = append(results, result)

		if hasError && transactionMode {
			break
		}
	}

	return utils.OutputSuccess(map[string]interface{}{
		"script":      scriptPath,
		"operations":  len(script.Operations),
		"dry_run":     batchDryRun,
		"transaction": transactionMode,
		"results":     results,
	})
}

type BatchScript struct {
	Name       string           `json:"name"`
	Version    string           `json:"version"`
	Operations []BatchOperation `json:"operations"`
}

type BatchOperation struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
}

func executeBatchOperation(engine *aql.Engine, op BatchOperation) error {
	symbols, err := engine.GetSymbol(op.Target)
	if err != nil || len(symbols) == 0 {
		return nil
	}

	symbol := symbols[0]

	locEngine := modify.NewLocationEngine(
		engine.GetIndex(),
		nil,
		engine.GetParser(),
		engine.GetStorage(),
	)

	modEngine := modify.NewModifyEngine(locEngine, nil, engine.GetIndex())

	loc := modify.Location{
		Type:   modify.LocationSymbol,
		Target: symbol.Name,
	}

	var opType modify.OperationType
	switch op.Type {
	case "insert":
		opType = modify.OperationInsertCode
	case "add-param":
		opType = modify.OperationAddParameter
	case "replace":
		opType = modify.OperationReplaceCode
	case "delete":
		opType = modify.OperationDeleteCode
	default:
		return fmt.Errorf("unknown operation type: %s", op.Type)
	}

	operation := modify.Operation{
		Type:          opType,
		Location:      &loc,
		Modifications: []modify.Modification{},
	}

	if code, ok := op.Parameters["code"].(string); ok {
		modType := "insert"
		if opType == modify.OperationReplaceCode {
			modType = "replace"
		}
		
		position := "after"
		if pos, ok := op.Parameters["position"].(string); ok {
			position = pos
		}
		
		operation.Modifications = append(operation.Modifications, modify.Modification{
			Type:     modType,
			Code:     code,
			Position: position,
		})
	}

	_, err = modEngine.Apply(operation)
	return err
}
