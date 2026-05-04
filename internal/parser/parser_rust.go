package parser

import (
	"fmt"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

func (p *CodeParser) parseRust(path string, content string) ([]*types.Symbol, error) {
	symbols := []*types.Symbol{}
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "fn ") {
			symbol := p.parseRustFunction(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "pub fn ") {
			symbol := p.parseRustFunction(path, strings.TrimPrefix(trimmed, "pub "), i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "struct ") {
			symbol := p.parseRustStruct(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "pub struct ") {
			symbol := p.parseRustStruct(path, strings.TrimPrefix(trimmed, "pub "), i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "trait ") {
			symbol := p.parseRustTrait(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "pub trait ") {
			symbol := p.parseRustTrait(path, strings.TrimPrefix(trimmed, "pub "), i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "enum ") {
			symbol := p.parseRustEnum(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "pub enum ") {
			symbol := p.parseRustEnum(path, strings.TrimPrefix(trimmed, "pub "), i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		} else if strings.HasPrefix(trimmed, "impl ") {
			symbol := p.parseRustImpl(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}
	}

	return symbols, nil
}

func (p *CodeParser) parseRustFunction(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "fn ")

	var name string
	if idx := strings.Index(line, "<"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else if idx := strings.Index(line, "("); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		name = strings.Fields(line)[0]
	}

	signature := "fn " + line

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolFunction,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseRustStruct(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "struct ")

	var name string
	if idx := strings.Index(line, "<"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		name = strings.Fields(line)[0]
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

func (p *CodeParser) parseRustTrait(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "trait ")

	var name string
	if idx := strings.Index(line, "<"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		name = strings.Fields(line)[0]
	}

	signature := "trait " + name

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolTrait,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseRustEnum(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "enum ")

	var name string
	if idx := strings.Index(line, "<"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		name = strings.Fields(line)[0]
	}

	signature := "enum " + name

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolEnum,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseRustImpl(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "impl ")

	var name string
	if idx := strings.Index(line, "<"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		name = strings.Fields(line)[0]
	}

	if strings.Contains(name, " for ") {
		parts := strings.SplitN(name, " for ", 2)
		name = strings.TrimSpace(parts[0])
	}

	signature := "impl " + name

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolStruct,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}
