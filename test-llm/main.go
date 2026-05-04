package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

var logFile *os.File

func main() {
	fmt.Println("[Main] AI-Git LLM 测试启动")
	
	var err error
	logFile, err = os.Create(fmt.Sprintf("log_%s.txt", time.Now().Format("20060102_150405")))
	if err != nil {
		fmt.Printf("创建日志文件失败: %v\n", err)
		return
	}
	defer logFile.Close()
	
	log("========================================")
	log("AI-Git LLM 测试启动")
	log("时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	log("========================================")
	
	cfg := loadConfig()
	
	if cfg.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: MINIMAX_API_KEY not set")
		log("错误: MINIMAX_API_KEY 未设置")
		return
	}
	
	log("配置信息:")
	log("  模型: %s", cfg.Model)
	log("  Base URL: %s", cfg.BaseURL)
	
	client := openai.NewClient(option.WithAPIKey(cfg.APIKey), option.WithBaseURL(cfg.BaseURL))
	
	systemPrompt, _ := readFile("prompt.txt")
	userTask, _ := readFile("user.txt")
	
	log("\n系统提示词长度: %d 字符", len(systemPrompt))
	log("用户任务长度: %d 字符", len(userTask))
	log("\n用户任务:\n%s", userTask)
	log("\n========================================")
	
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(userTask),
	}
	
	tools := loadTools("tools.json")
	
	fmt.Printf("\n[Main] 任务: %s\n", truncateString(userTask, 50))
	
	for iteration := 1; ; iteration++ {
		log("\n\n========== 迭代 %d ==========", iteration)
		fmt.Printf("\n[迭代 %d] ═══════════════════════\n", iteration)
		
		params := openai.ChatCompletionNewParams{
			Model:    cfg.Model,
			Messages: messages,
			Tools:    tools,
		}
		
		resp, err := client.Chat.Completions.New(context.Background(), params)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
			log("错误: %v", err)
			return
		}
		
		msg := resp.Choices[0].Message
		
		log("\n响应信息:")
		log("  Tool Calls: %d", len(msg.ToolCalls))
		log("  Total Tokens: %d", resp.Usage.TotalTokens)
		log("  Prompt Tokens: %d", resp.Usage.PromptTokens)
		log("  Completion Tokens: %d", resp.Usage.CompletionTokens)
		
		if msg.Content != "" {
			log("\n模型思考:\n%s", msg.Content)
		}
		
		fmt.Printf("响应: tool_calls=%d, total_tokens=%d\n", len(msg.ToolCalls), resp.Usage.TotalTokens)
		
		if len(msg.ToolCalls) == 0 {
			log("\n========================================")
			log("任务完成")
			log("========================================")
			log("\n最终输出:\n%s", msg.Content)
			
			fmt.Println("\n=== 任务完成 ===")
			fmt.Println(msg.Content)
			saveOutput(msg.Content, "output_done.txt")
			return
		}
		
		var assistantMsg openai.ChatCompletionAssistantMessageParam
		if msg.Content != "" {
			assistantMsg.Content.OfString = openai.String(msg.Content)
		}
		
		for _, tc := range msg.ToolCalls {
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, openai.ChatCompletionMessageToolCallUnionParam{
				OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
					ID: tc.ID,
					Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				},
			})
		}
		
		messages = append(messages, openai.ChatCompletionMessageParamUnion{OfAssistant: &assistantMsg})
		
		for _, tc := range msg.ToolCalls {
			log("\n----------------------------------------")
			log("工具调用: %s", tc.Function.Name)
			log("参数: %s", tc.Function.Arguments)
			
			fmt.Printf("  [%s] ", tc.Function.Name)
			result := executeTool(tc.Function.Name, tc.Function.Arguments)
			
			log("\n执行结果:")
			log("%s", result)
			log("----------------------------------------")
			
			messages = append(messages, openai.ToolMessage(result, tc.ID))
		}
	}
}

func log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(logFile, msg)
}

func loadConfig() *Config {
	cfg := &Config{}
	
	if data, err := os.ReadFile("config.json"); err == nil {
		json.Unmarshal(data, cfg)
	}
	
	cfg.APIKey = expandEnv(cfg.APIKey)
	cfg.BaseURL = expandEnv(cfg.BaseURL)
	cfg.Model = expandEnv(cfg.Model)
	
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.minimaxi.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "MiniMax-M2.7"
	}
	
	return cfg
}

func expandEnv(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		return os.Getenv(s[2 : len(s)-1])
	}
	return s
}

func loadTools(path string) []openai.ChatCompletionToolUnionParam {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	
	var configs []struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Parameters  map[string]interface{} `json:"parameters"`
	}
	
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil
	}
	
	var tools []openai.ChatCompletionToolUnionParam
	for _, cfg := range configs {
		tools = append(tools, openai.ChatCompletionToolUnionParam{
			OfFunction: &openai.ChatCompletionFunctionToolParam{
				Function: openai.FunctionDefinitionParam{
					Name:        cfg.Name,
					Description: openai.String(cfg.Description),
					Parameters:  cfg.Parameters,
				},
			},
		})
	}
	return tools
}

func executeTool(name, arguments string) string {
	params := parseArguments(arguments)
	
	var result string
	switch name {
	case "terminal":
		cmd, _ := params["cmd"].(string)
		result = toolCommandRun(cmd)
	default:
		result = fmt.Sprintf("error: unknown tool %s", name)
	}
	
	if len(result) > 100 {
		fmt.Printf("%s...\n", result[:100])
	} else {
		fmt.Println(result)
	}
	
	return result
}

func toolCommandRun(cmd string) string {
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	result := string(out)
	if len(result) > 10000 {
		result = result[:10000] + "\n... (truncated)"
	}
	if err != nil {
		return fmt.Sprintf("error: %v\n%s", err, result)
	}
	if result == "" {
		return "(no output)"
	}
	return result
}

func parseArguments(s string) map[string]interface{} {
	var r map[string]interface{}
	json.Unmarshal([]byte(s), &r)
	return r
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func saveOutput(content, filename string) {
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		fmt.Printf("保存输出失败: %v\n", err)
		log("保存输出失败: %v", err)
	} else {
		fmt.Printf("输出已保存到: %s\n", filename)
		log("输出已保存到: %s", filename)
	}
}
