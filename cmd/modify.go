package cmd

import (
	"fmt"

	"github.com/mychow/ai-git/internal/modify"
	"github.com/mychow/ai-git/pkg/types"
	"github.com/mychow/ai-git/pkg/utils"
	"github.com/spf13/cobra"
)

var (
	modifyType         string
	modifyPosition     string
	modifyCode         string
	modifyParamName    string
	modifyParamType    string
	modifyParamDefault string
	modifyAutoUpdate   bool
)

var modifyCmd = &cobra.Command{
	Use:   "modify <symbol>",
	Short: "Modify code",
	Long:  `Modify code at the specified symbol location.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runModify,
}

var generateCmd = &cobra.Command{
	Use:   "generate <type>",
	Short: "Generate code",
	Long:  `Generate code (function, class, test) based on specifications.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(modifyCmd)
	rootCmd.AddCommand(generateCmd)

	modifyCmd.Flags().StringVarP(&modifyType, "type", "t", "insert", "Modification type (insert, add-param, replace, delete)")
	modifyCmd.Flags().StringVarP(&modifyPosition, "position", "p", "after", "Position for insert (before, after, before-first-line, after-last-line)")
	modifyCmd.Flags().StringVarP(&modifyCode, "code", "c", "", "Code to insert")
	modifyCmd.Flags().StringVar(&modifyParamName, "param-name", "", "Parameter name to add")
	modifyCmd.Flags().StringVar(&modifyParamType, "param-type", "", "Parameter type")
	modifyCmd.Flags().StringVar(&modifyParamDefault, "param-default", "", "Parameter default value")
	modifyCmd.Flags().BoolVar(&modifyAutoUpdate, "auto-update", false, "Automatically update references")

	generateCmd.Flags().StringVar(&modifyParamName, "name", "", "Function/class name")
	generateCmd.Flags().StringVar(&modifyParamType, "return", "", "Return type for function")
	generateCmd.Flags().StringVarP(&modifyCode, "body", "b", "", "Function body")
	generateCmd.Flags().StringVar(&modifyPosition, "file", "", "Output file path")
}

func runModify(cmd *cobra.Command, args []string) error {
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

	if modifyPosition != "" {
		loc.Type = modify.LocationRelative
		loc.Position = modifyPosition
	}

	var opType modify.OperationType
	switch modifyType {
	case "insert":
		opType = modify.OperationInsertCode
	case "add-param":
		opType = modify.OperationAddParameter
	case "replace":
		opType = modify.OperationReplaceCode
	case "delete":
		opType = modify.OperationDeleteCode
	default:
		return fmt.Errorf("unknown modification type: %s", modifyType)
	}

	op := modify.Operation{
		Type:           opType,
		Location:       &loc,
		AutoUpdateRefs: modifyAutoUpdate,
		Modifications:  []modify.Modification{},
	}

	if modifyType == "insert" && modifyCode != "" {
		op.Modifications = append(op.Modifications, modify.Modification{
			Type: "insert",
			Code: modifyCode,
		})
	}

	if modifyType == "add-param" && modifyParamName != "" {
		param := &types.Parameter{
			Name:    modifyParamName,
			Type:    modifyParamType,
			Default: modifyParamDefault,
		}
		op.Modifications = append(op.Modifications, modify.Modification{
			Type:      "add_parameter",
			Parameter: param,
		})
	}

	result, err := modEngine.Apply(op)
	if err != nil {
		return fmt.Errorf("modification failed: %w", err)
	}

	return utils.OutputSuccess(map[string]interface{}{
		"success":          result.Success,
		"files_changed":    result.FilesChanged,
		"symbols_modified": result.SymbolsModified,
		"errors":           result.Errors,
	})
}

func runGenerate(cmd *cobra.Command, args []string) error {
	genType := args[0]

	engine, err := GetEngine("")
	if err != nil {
		return err
	}
	defer engine.Close()

	generator := modify.NewCodeGenerator(engine.GetParser())

	var generatedCode string

	switch genType {
	case "function":
		if modifyParamName == "" {
			return fmt.Errorf("function name required (--name)")
		}
		params := []types.Parameter{}
		generatedCode = generator.GenerateFunction(modifyParamName, modifyParamType, params, modifyCode)

	case "class":
		if modifyParamName == "" {
			return fmt.Errorf("class name required (--name)")
		}
		fields := []types.Parameter{}
		methods := []string{}
		generatedCode = generator.GenerateClass(modifyParamName, fields, methods)

	case "test":
		if modifyParamName == "" {
			return fmt.Errorf("function name required (--name)")
		}
		testCases := []map[string]interface{}{
			{
				"input":    "test_input",
				"expected": "expected_output",
			},
		}
		generatedCode = generator.GenerateTest(modifyParamName, testCases)

	default:
		return fmt.Errorf("unknown generation type: %s", genType)
	}

	generatedCode = generator.FormatCode(generatedCode, "go")

	if modifyPosition != "" {
		if err := generator.SaveToFile(generatedCode, modifyPosition); err != nil {
			return fmt.Errorf("failed to save file: %w", err)
		}
		return utils.OutputSuccess(map[string]interface{}{
			"type":    genType,
			"file":    modifyPosition,
			"code":    generatedCode,
			"message": "Code generated and saved successfully",
		})
	}

	return utils.OutputSuccess(map[string]interface{}{
		"type": genType,
		"code": generatedCode,
	})
}
