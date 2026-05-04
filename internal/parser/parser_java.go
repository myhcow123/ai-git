package parser

import (
	"fmt"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

func (p *CodeParser) parseJava(path string, content string) ([]*types.Symbol, error) {
	symbols := []*types.Symbol{}
	lines := strings.Split(content, "\n")

	inClass := false
	className := ""
	braceCount := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		if strings.Contains(trimmed, "class ") && !strings.Contains(trimmed, "class.") {
			symbol := p.parseJavaClass(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
				className = symbol.Name
				inClass = true
			}
		}

		if strings.HasPrefix(trimmed, "interface ") {
			symbol := p.parseJavaInterface(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}

		if strings.HasPrefix(trimmed, "enum ") {
			symbol := p.parseJavaEnum(path, trimmed, i+1)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}

		if p.isJavaMethod(trimmed) && inClass {
			symbol := p.parseJavaMethod(path, trimmed, i+1, className)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}

		braceCount += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
		if braceCount <= 0 && inClass {
			inClass = false
			className = ""
		}
	}

	return symbols, nil
}

func (p *CodeParser) parseJavaClass(path, line string, lineNum int) *types.Symbol {
	var name string

	if idx := strings.Index(line, "class "); idx != -1 {
		afterClass := strings.TrimSpace(line[idx+6:])
		parts := strings.Fields(afterClass)
		if len(parts) > 0 {
			name = parts[0]
			if idx := strings.Index(name, "<"); idx != -1 {
				name = name[:idx]
			}
			if idx := strings.Index(name, "{"); idx != -1 {
				name = name[:idx]
			}
		}
	}

	if name == "" {
		return nil
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

func (p *CodeParser) parseJavaInterface(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "interface ")

	var name string
	if idx := strings.Index(line, "<"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			name = parts[0]
		}
	}

	if name == "" {
		return nil
	}

	signature := "interface " + name

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolInterface,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}

func (p *CodeParser) parseJavaEnum(path, line string, lineNum int) *types.Symbol {
	line = strings.TrimPrefix(line, "enum ")

	var name string
	if idx := strings.Index(line, "{"); idx != -1 {
		name = strings.TrimSpace(line[:idx])
	} else {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			name = parts[0]
		}
	}

	if name == "" {
		return nil
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

func (p *CodeParser) isJavaMethod(line string) bool {
	if !strings.Contains(line, "(") || !strings.Contains(line, ")") {
		return false
	}

	if strings.HasPrefix(line, "if ") || strings.HasPrefix(line, "while ") || strings.HasPrefix(line, "for ") || strings.HasPrefix(line, "switch ") || strings.HasPrefix(line, "catch ") {
		return false
	}

	if strings.HasPrefix(line, "new ") || strings.HasPrefix(line, "return ") || strings.HasPrefix(line, "throw ") {
		return false
	}

	words := strings.Fields(line)
	if len(words) < 2 {
		return false
	}

	return true
}

func (p *CodeParser) parseJavaMethod(path, line string, lineNum int, className string) *types.Symbol {
	line = strings.TrimPrefix(line, "public ")
	line = strings.TrimPrefix(line, "private ")
	line = strings.TrimPrefix(line, "protected ")
	line = strings.TrimPrefix(line, "static ")
	line = strings.TrimPrefix(line, "final ")
	line = strings.TrimPrefix(line, "abstract ")
	line = strings.TrimPrefix(line, "synchronized ")
	line = strings.TrimPrefix(line, "native ")
	line = strings.TrimPrefix(line, "@Override ")
	line = strings.TrimSpace(line)

	var name string
	if idx := strings.Index(line, "("); idx != -1 {
		beforeParen := strings.TrimSpace(line[:idx])
		parts := strings.Fields(beforeParen)
		if len(parts) >= 1 {
			name = parts[len(parts)-1]
		}
	}

	if name == "" || name == className {
		return nil
	}

	signature := line

	return &types.Symbol{
		ID:        fmt.Sprintf("%s:%d", path, lineNum),
		Name:      name,
		Type:      types.SymbolMethod,
		File:      path,
		LineStart: lineNum,
		Signature: signature,
	}
}
