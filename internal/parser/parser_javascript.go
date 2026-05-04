package parser

import (
	"fmt"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

func (p *CodeParser) parseJavaScript(path string, content string) ([]*types.Symbol, error) {
	symbols := []*types.Symbol{}
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "function ") {
			symbol := p.parseJSFunction(path, line, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(line, "const ") || strings.HasPrefix(line, "let ") || strings.HasPrefix(line, "var ") {
			if strings.Contains(line, "=>") || strings.Contains(line, "function") {
				symbol := p.parseJSArrowFunction(path, line, i+1)
				if symbol != nil {
					symbols = append(symbols, symbol)
				}
			}
		} else if strings.HasPrefix(line, "class ") {
			symbol := p.parseJSClass(path, line, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols, nil
}

func (p *CodeParser) parseJSFunction(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "function ")
	parts := strings.SplitN(line, "(", 2)
	if len(parts) < 1 {
		return nil
	}

	name := strings.TrimSpace(parts[0])
	signature := "function " + line

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolFunction,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseJSArrowFunction(path, line string, lineNum int) *types.Symbol {
	var name string

	if strings.HasPrefix(line, "const ") {
		line = strings.TrimPrefix(line, "const ")
	} else if strings.HasPrefix(line, "let ") {
		line = strings.TrimPrefix(line, "let ")
	} else if strings.HasPrefix(line, "var ") {
		line = strings.TrimPrefix(line, "var ")
	}

	if idx := strings.Index(line, "="); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		return nil
	}

	signature := line

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolFunction,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseJSClass(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "class ")
	parts := strings.Fields(line)
	if len(parts) < 1 {
		return nil
	}

	name := parts[0]
	signature := "class " + name

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolClass,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}
