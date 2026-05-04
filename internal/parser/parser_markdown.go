package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

type MarkdownBlock struct {
	Type     string
	Level    int
	Content  string
	Line     int
	Metadata map[string]string
	Items    []ChecklistItem
}

type ChecklistItem struct {
	Text     string
	Checked  bool
	Line     int
}

func (p *CodeParser) parseMarkdown(path string, content string) ([]*types.Symbol, error) {
	symbols := []*types.Symbol{}
	lines := strings.Split(content, "\n")

	inCodeBlock := false
	codeBlockStart := 0
	codeBlockLang := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockStart = i + 1
				codeBlockLang = strings.TrimPrefix(trimmed, "```")
			} else {
				inCodeBlock = false
				symbol := &types.Symbol{
					ID:        fmt.Sprintf("%s:codeblock:%d", path, codeBlockStart),
					Name:      fmt.Sprintf("CodeBlock:%s", codeBlockLang),
					Type:      types.SymbolCodeBlock,
					File:      path,
					LineStart: codeBlockStart,
					LineEnd:   i + 1,
					Signature: fmt.Sprintf("```%s", codeBlockLang),
				}
				symbols = append(symbols, symbol)
			}
			continue
		}

		if inCodeBlock {
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}

			title := strings.TrimSpace(trimmed[level:])
			if title == "" {
				continue
			}

			symbol := &types.Symbol{
				ID:        fmt.Sprintf("%s:section:%d", path, i+1),
				Name:      title,
				Type:      types.SymbolSection,
				File:      path,
				LineStart: i + 1,
				LineEnd:   i + 1,
				Signature: strings.Repeat("#", level) + " " + title,
			}
			symbols = append(symbols, symbol)
		}

		if strings.HasPrefix(trimmed, "- [") || strings.HasPrefix(trimmed, "* [") {
			checked := strings.Contains(trimmed, "[x]") || strings.Contains(trimmed, "[X]")
			
			var text string
			if idx := strings.Index(trimmed, "] "); idx != -1 {
				text = strings.TrimSpace(trimmed[idx+2:])
			}

			symbol := &types.Symbol{
				ID:        fmt.Sprintf("%s:checklist:%d", path, i+1),
				Name:      text,
				Type:      types.SymbolChecklist,
				File:      path,
				LineStart: i + 1,
				LineEnd:   i + 1,
				Signature: trimmed,
			}
			if checked {
				symbol.Code = "checked"
			}
			symbols = append(symbols, symbol)
		}
	}

	return symbols, nil
}

func ParseMarkdownMetadata(content string) map[string]string {
	metadata := make(map[string]string)
	lines := strings.Split(content, "\n")

	inMetadata := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			inMetadata = !inMetadata
			continue
		}

		if inMetadata {
			if strings.Contains(trimmed, ":") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					metadata[key] = value
				}
			}
		}

		if !inMetadata && trimmed != "" {
			break
		}
	}

	return metadata
}

func ParseMarkdownLinks(content string) []string {
	re := regexp.MustCompile(`\[\[([^\]]+)\]\]|\[([^\]]+)\]\(([^)]+)\)`)
	matches := re.FindAllStringSubmatch(content, -1)

	links := []string{}
	for _, match := range matches {
		if match[1] != "" {
			links = append(links, match[1])
		} else if match[3] != "" {
			links = append(links, match[3])
		}
	}

	return links
}

func ParseMarkdownTags(content string) []string {
	re := regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	tags := []string{}
	seen := make(map[string]bool)
	for _, match := range matches {
		tag := match[1]
		if !seen[tag] {
			tags = append(tags, tag)
			seen[tag] = true
		}
	}

	return tags
}

func ExtractSection(content string, sectionTitle string) string {
	lines := strings.Split(content, "\n")
	var result []string
	found := false
	targetLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}

			title := strings.TrimSpace(trimmed[level:])

			if !found && strings.EqualFold(title, sectionTitle) {
				found = true
				targetLevel = level
				continue
			}

			if found && level <= targetLevel {
				break
			}
		}

		if found {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func GetMarkdownTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
		}
	}
	return ""
}
