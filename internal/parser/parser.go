package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

type Language string

const (
	LanguagePython     Language = "python"
	LanguageJavaScript Language = "javascript"
	LanguageTypeScript Language = "typescript"
	LanguageGo         Language = "go"
	LanguageRust       Language = "rust"
	LanguageJava       Language = "java"
	LanguageC          Language = "c"
	LanguageCPP        Language = "cpp"
	LanguageMarkdown   Language = "markdown"
)

type CodeParser struct {
	languages map[Language]bool
}

func NewCodeParser() *CodeParser {
	return &CodeParser{
		languages: make(map[Language]bool),
	}
}

func (p *CodeParser) DetectLanguage(filename string) Language {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".py":
		return LanguagePython
	case ".js":
		return LanguageJavaScript
	case ".ts", ".tsx":
		return LanguageTypeScript
	case ".go":
		return LanguageGo
	case ".rs":
		return LanguageRust
	case ".java":
		return LanguageJava
	case ".c", ".h":
		return LanguageC
	case ".cpp", ".cc", ".cxx", ".hpp", ".hxx":
		return LanguageCPP
	case ".md", ".markdown":
		return LanguageMarkdown
	default:
		return ""
	}
}

func (p *CodeParser) ParseFile(path string, content string) ([]*types.Symbol, error) {
	lang := p.DetectLanguage(path)
	if lang == "" {
		return nil, fmt.Errorf("unsupported language for file: %s", path)
	}

	switch lang {
	case LanguagePython:
		return p.parsePython(path, content)
	case LanguageGo:
		return p.parseGo(path, content)
	case LanguageJavaScript, LanguageTypeScript:
		return p.parseJavaScript(path, content)
	case LanguageRust:
		return p.parseRust(path, content)
	case LanguageC, LanguageCPP:
		return p.parseC(path, content)
	case LanguageJava:
		return p.parseJava(path, content)
	case LanguageMarkdown:
		return p.parseMarkdown(path, content)
	default:
		return nil, fmt.Errorf("parser not implemented for language: %s", lang)
	}
}
