package semantic

import (
	"regexp"
	"strings"

	"github.com/mychow/ai-git/pkg/types"
)

type IntentEngine struct {
	patterns map[string]*regexp.Regexp
}

func NewIntentEngine() *IntentEngine {
	return &IntentEngine{
		patterns: map[string]*regexp.Regexp{
			"create":    regexp.MustCompile(`(?i)(create|add|new|insert|make|build)`),
			"read":      regexp.MustCompile(`(?i)(get|read|fetch|find|retrieve|load|query)`),
			"update":    regexp.MustCompile(`(?i)(update|modify|change|edit|set|patch)`),
			"delete":    regexp.MustCompile(`(?i)(delete|remove|drop|clear|destroy)`),
			"validate":  regexp.MustCompile(`(?i)(validate|check|verify|ensure|confirm)`),
			"process":   regexp.MustCompile(`(?i)(process|handle|execute|run|perform)`),
			"calculate": regexp.MustCompile(`(?i)(calculate|compute|count|sum|average)`),
			"convert":   regexp.MustCompile(`(?i)(convert|transform|parse|format|encode)`),
			"search":    regexp.MustCompile(`(?i)(search|find|lookup|filter|match)`),
			"send":      regexp.MustCompile(`(?i)(send|post|submit|emit|publish)`),
		},
	}
}

func (e *IntentEngine) InferIntent(symbol *types.Symbol) *Intent {
	intent := &Intent{
		Primary:    "unknown",
		Secondary:  []string{},
		Confidence: 0.0,
	}

	name := strings.ToLower(symbol.Name)
	signature := strings.ToLower(symbol.Signature)

	intentScores := make(map[string]float64)

	for intentType, pattern := range e.patterns {
		nameMatches := pattern.FindAllString(name, -1)
		sigMatches := pattern.FindAllString(signature, -1)

		score := float64(len(nameMatches))*2.0 + float64(len(sigMatches))
		if score > 0 {
			intentScores[intentType] = score
		}
	}

	if len(intentScores) == 0 {
		intent.Primary = "utility"
		intent.Confidence = 0.3
		return intent
	}

	maxScore := 0.0
	for intentType, score := range intentScores {
		if score > maxScore {
			maxScore = score
			intent.Primary = intentType
		}
	}

	threshold := maxScore * 0.5
	for intentType, score := range intentScores {
		if intentType != intent.Primary && score >= threshold {
			intent.Secondary = append(intent.Secondary, intentType)
		}
	}

	totalScore := 0.0
	for _, score := range intentScores {
		totalScore += score
	}
	intent.Confidence = maxScore / (totalScore + 1.0)

	return intent
}

type Intent struct {
	Primary    string   `json:"primary"`
	Secondary  []string `json:"secondary"`
	Confidence float64  `json:"confidence"`
}

type PatternRecognizer struct {
	designPatterns map[string]*DesignPattern
}

type DesignPattern struct {
	Name        string
	Description string
	CheckFunc   func(*types.Symbol) bool
}

func NewPatternRecognizer() *PatternRecognizer {
	return &PatternRecognizer{
		designPatterns: map[string]*DesignPattern{
			"singleton": {
				Name:        "Singleton",
				Description: "Ensure a class only has one instance",
				CheckFunc:   isSingleton,
			},
			"factory": {
				Name:        "Factory",
				Description: "Create objects without specifying their concrete class",
				CheckFunc:   isFactory,
			},
			"repository": {
				Name:        "Repository",
				Description: "Abstract data access layer",
				CheckFunc:   isRepository,
			},
			"service": {
				Name:        "Service",
				Description: "Business logic layer",
				CheckFunc:   isService,
			},
			"controller": {
				Name:        "Controller",
				Description: "Handle HTTP requests",
				CheckFunc:   isController,
			},
		},
	}
}

func (r *PatternRecognizer) Recognize(symbol *types.Symbol) []PatternMatch {
	matches := []PatternMatch{}

	for patternName, pattern := range r.designPatterns {
		if pattern.CheckFunc(symbol) {
			matches = append(matches, PatternMatch{
				Pattern:     patternName,
				Description: pattern.Description,
				Confidence:  0.8,
			})
		}
	}

	return matches
}

type PatternMatch struct {
	Pattern     string  `json:"pattern"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

func isSingleton(symbol *types.Symbol) bool {
	if symbol.Type != types.SymbolClass {
		return false
	}

	name := strings.ToLower(symbol.Name)
	return strings.Contains(name, "singleton") ||
		strings.Contains(name, "instance") ||
		strings.Contains(name, "getinstance")
}

func isFactory(symbol *types.Symbol) bool {
	if symbol.Type != types.SymbolClass {
		return false
	}

	name := strings.ToLower(symbol.Name)
	return strings.Contains(name, "factory") ||
		strings.Contains(name, "builder") ||
		strings.Contains(name, "creator")
}

func isRepository(symbol *types.Symbol) bool {
	if symbol.Type != types.SymbolClass {
		return false
	}

	name := strings.ToLower(symbol.Name)
	if !strings.Contains(name, "repository") && !strings.Contains(name, "repo") {
		return false
	}

	repoMethods := []string{"find", "save", "delete", "update", "get", "create"}
	methodCount := 0

	for _, method := range repoMethods {
		if strings.Contains(strings.ToLower(symbol.Signature), method) {
			methodCount++
		}
	}

	return methodCount >= 2
}

func isService(symbol *types.Symbol) bool {
	if symbol.Type != types.SymbolClass {
		return false
	}

	name := strings.ToLower(symbol.Name)
	return strings.Contains(name, "service") ||
		strings.Contains(name, "manager") ||
		strings.Contains(name, "handler")
}

func isController(symbol *types.Symbol) bool {
	if symbol.Type != types.SymbolClass {
		return false
	}

	name := strings.ToLower(symbol.Name)
	return strings.Contains(name, "controller") ||
		strings.Contains(name, "handler") ||
		strings.Contains(name, "api")
}

type QualityAssessor struct {
	complexityThreshold int
}

func NewQualityAssessor() *QualityAssessor {
	return &QualityAssessor{
		complexityThreshold: 10,
	}
}

func (a *QualityAssessor) Assess(symbol *types.Symbol) *types.QualityAssessment {
	assessment := &types.QualityAssessment{
		Complexity:      a.assessComplexity(symbol),
		Testability:     a.assessTestability(symbol),
		Maintainability: a.assessMaintainability(symbol),
		Security:        a.assessSecurity(symbol),
	}

	assessment.OverallScore = a.calculateOverallScore(assessment)

	return assessment
}

func (a *QualityAssessor) assessComplexity(symbol *types.Symbol) types.ComplexityAssessment {
	cyclomatic := a.calculateCyclomaticComplexity(symbol)
	cognitive := a.calculateCognitiveComplexity(symbol)

	rating := "good"
	if cyclomatic > 10 || cognitive > 15 {
		rating = "poor"
	} else if cyclomatic > 5 || cognitive > 10 {
		rating = "moderate"
	}

	return types.ComplexityAssessment{
		Cyclomatic: cyclomatic,
		Cognitive:  cognitive,
		Rating:     rating,
	}
}

func (a *QualityAssessor) assessTestability(symbol *types.Symbol) float64 {
	score := 1.0

	if symbol.Type == types.SymbolFunction || symbol.Type == types.SymbolMethod {
		paramCount := len(symbol.Parameters)
		if paramCount > 5 {
			score -= 0.2
		}

		if strings.Contains(symbol.Signature, "global") ||
			strings.Contains(symbol.Signature, "singleton") {
			score -= 0.3
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (a *QualityAssessor) assessMaintainability(symbol *types.Symbol) float64 {
	score := 1.0

	nameLength := len(symbol.Name)
	if nameLength < 3 || nameLength > 30 {
		score -= 0.2
	}

	if symbol.Signature != "" {
		sigLength := len(symbol.Signature)
		if sigLength > 100 {
			score -= 0.3
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (a *QualityAssessor) assessSecurity(symbol *types.Symbol) float64 {
	score := 1.0

	signature := strings.ToLower(symbol.Signature)

	dangerousPatterns := []string{
		"exec(", "eval(", "system(", "shell(",
		"sql", "query", "execute",
		"password", "secret", "key", "token",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(signature, pattern) {
			score -= 0.2
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

func (a *QualityAssessor) calculateOverallScore(assessment *types.QualityAssessment) float64 {
	complexityScore := 1.0
	if assessment.Complexity.Rating == "poor" {
		complexityScore = 0.3
	} else if assessment.Complexity.Rating == "moderate" {
		complexityScore = 0.7
	}

	overallScore := (complexityScore +
		assessment.Testability +
		assessment.Maintainability +
		assessment.Security) / 4.0

	return overallScore
}

func (a *QualityAssessor) calculateCyclomaticComplexity(symbol *types.Symbol) int {
	if symbol.Code == "" {
		return 1
	}

	complexity := 1

	keywords := []string{"if", "else", "for", "while", "case", "catch", "&&", "||", "?"}

	code := symbol.Code
	for _, keyword := range keywords {
		complexity += strings.Count(code, keyword)
	}

	return complexity
}

func (a *QualityAssessor) calculateCognitiveComplexity(symbol *types.Symbol) int {
	if symbol.Code == "" {
		return 0
	}

	complexity := 0

	nestingLevel := 0
	lines := strings.Split(symbol.Code, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "if") ||
			strings.HasPrefix(trimmed, "for") ||
			strings.HasPrefix(trimmed, "while") {
			nestingLevel++
			complexity += nestingLevel
		}

		if strings.Contains(trimmed, "}") && nestingLevel > 0 {
			nestingLevel--
		}
	}

	return complexity
}
