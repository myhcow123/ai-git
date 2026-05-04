package parser

import (
	"fmt"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

func (p *CodeParser) parseC(path string, content string) ([]*types.Symbol, error) {
	symbols := []*types.Symbol{}
	lines := strings.Split(content, "\n")

	inClass := false
	className := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "class ") {
			parts := strings.Fields(strings.TrimPrefix(trimmed, "class "))
			if len(parts) > 0 {
				className = parts[0]
				if idx := strings.Index(className, ":"); idx != -1 {
					className = className[:idx]
				}
				if idx := strings.Index(className, "{"); idx != -1 {
					className = className[:idx]
				}
				symbols = append(symbols, &types.Symbol{
					ID:        fmt.Sprintf("%s:%d", path, i+1),
					Name:      className,
					Type:      types.SymbolClass,
					File:      path,
					LineStart: i + 1,
					Signature: "class " + className,
				})
				inClass = true
			}
		}

		if strings.Contains(trimmed, "};") && inClass {
			inClass = false
			className = ""
		}

		if strings.HasPrefix(trimmed, "struct ") {
			symbol := p.parseCStruct(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}

		if p.isCFunction(trimmed) {
			symbol := p.parseCFunction(path, trimmed, i+1, className)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols, nil
}

func (p *CodeParser) parseCStruct(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "struct ")

	var name string
	if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		name = strings.Fields(line)[0]
	}

	if name == "" {
		return nil
	}

	signature := "struct " + name

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolStruct,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) isCFunction(line string) bool {
	if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
		return false
	}

	if strings.HasPrefix(line, "#") {
		return false
	}

	if strings.HasPrefix(line, "if ") || strings.HasPrefix(line, "while ") || strings.HasPrefix(line, "for ") || strings.HasPrefix(line, "switch ") {
		return false
	}

	if !strings.Contains(line, "(") || !strings.Contains(line, ")") {
		return false
	}

	if strings.Contains(line, ";") && strings.Index(line, ";") < strings.Index(line, "{") {
		return false
	}

	words := strings.Fields(line)
	if len(words) < 2 {
		return false
	}

	for _, word := range words[:len(words)-1] {
		if word == "return" || word == "delete" || word == "new" || word == "throw" {
			return false
		}
	}

	return true
}

func (p *CodeParser) parseCFunction(path, line string, lineNum int, className string) *types.Symbol {
	var name string
	var signature string

	if idx := strings.Index(line, "("); idx != -1 {
		beforeParen := strings.TrimSpace(line[:idx])
		parts := strings.Fields(beforeParen)
		if len(parts) >= 1 {
			name = parts[len(parts)-1]
		}
	}

	if name == "" || name == "main" && className != "" {
		return nil
	}

	signature = line

	symType := types.SymbolFunction
	if className != "" {
		symType = types.SymbolMethod
	}

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      symType,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}
