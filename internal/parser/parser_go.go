package parser

import (
	"fmt"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

func (p *CodeParser) parseGo(path string, content string) ([]*types.Symbol, error) {
	symbols := []*types.Symbol{}
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "func ") {
			symbol := p.parseGoFunction(path, line, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(line, "type ") && strings.Contains(line, "struct") {
			symbol := p.parseGoStruct(path, line, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(line, "type ") && strings.Contains(line, "interface") {
			symbol := p.parseGoInterface(path, line, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols, nil
}

func (p *CodeParser) parseGoFunction(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "func ")

	var name string
	var signature string

	if strings.HasPrefix(line, "(") {
		receiverEnd := strings.Index(line, ")")
		if receiverEnd != -1 {
			namePart := strings.TrimSpace(line[receiverEnd+1:])
			if idx := strings.Index(namePart, "("); idx != -1 {
				name = strings.TrimSpace(namePart[:idx])
			} else {
				name = strings.Fields(namePart)[0]
			}
			signature = "func " + line
			return &types.Symbol{
				ID:        fmt.Sprintf("%s:%d", path, lineNum),
				Name:      name,
				Type:      types.SymbolMethod,
				File:      path,
				LineStart: lineNum,
				Signature: signature,
			}
		}
	}

	if idx := strings.Index(line, "("); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		name = strings.Fields(line)[0]
	}

	signature = "func " + line

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolFunction,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseGoStruct(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "type ")
	line = strings.TrimSuffix(line, "struct")
	line = strings.TrimSpace(line)

	name := strings.Fields(line)[0]
	signature := "type " + name + " struct"

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolStruct,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseGoInterface(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "type ")
	line = strings.TrimSuffix(line, "interface")
	line = strings.TrimSpace(line)

	name := strings.Fields(line)[0]
	signature := "type " + name + " interface"

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolInterface,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}
