package parser

import (
	"fmt"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

func (p *CodeParser) parsePython(path string, content string) ([]*types.Symbol, error) {
	symbols := []*types.Symbol{}
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "def ") {
			symbol := p.parsePythonFunction(path, line, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(line, "class ") {
			symbol := p.parsePythonClass(path, line, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols, nil
}

func (p *CodeParser) parsePythonFunction(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "def ")
	parts := strings.SplitN(line, "(", 2)
	if len(parts) < 2 {
		return nil
	}

	name := strings.TrimSpace(parts[0])
	signature := "def " + line

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolFunction,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parsePythonClass(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "class ")
	parts := strings.SplitN(line, ":", 2)
	if len(parts) < 1 {
		return nil
	}

	name := strings.TrimSpace(parts[0])
	if idx := strings.Index(name, "("); idx != -1 {
		name = name[:idx]
	}

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
