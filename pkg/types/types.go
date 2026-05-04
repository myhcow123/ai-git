package types

import "time"

type SymbolType int

const (
	SymbolFunction SymbolType = iota
	SymbolClass
	SymbolVariable
	SymbolConstant
	SymbolInterface
	SymbolStruct
	SymbolMethod
	SymbolTrait
	SymbolEnum
	SymbolSection
	SymbolTask
	SymbolNote
	SymbolLog
	SymbolCodeBlock
	SymbolChecklist
)

func (st SymbolType) String() string {
	switch st {
	case SymbolFunction:
		return "function"
	case SymbolClass:
		return "class"
	case SymbolVariable:
		return "variable"
	case SymbolConstant:
		return "constant"
	case SymbolInterface:
		return "interface"
	case SymbolStruct:
		return "struct"
	case SymbolMethod:
		return "method"
	case SymbolTrait:
		return "trait"
	case SymbolEnum:
		return "enum"
	case SymbolSection:
		return "section"
	case SymbolTask:
		return "task"
	case SymbolNote:
		return "note"
	case SymbolLog:
		return "log"
	case SymbolCodeBlock:
		return "code_block"
	case SymbolChecklist:
		return "checklist"
	default:
		return "unknown"
	}
}

type Parameter struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default,omitempty"`
	Required bool   `json:"required"`
}

type Symbol struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       SymbolType  `json:"type"`
	File       string      `json:"file"`
	LineStart  int         `json:"line_start"`
	LineEnd    int         `json:"line_end"`
	Signature  string      `json:"signature"`
	Parameters []Parameter `json:"parameters,omitempty"`
	ReturnType string      `json:"return_type,omitempty"`
	Code       string      `json:"code,omitempty"`
	Purpose    string      `json:"purpose,omitempty"`
	Complexity int         `json:"complexity,omitempty"`
	Stability  float64     `json:"stability,omitempty"`
	Embedding  []float64   `json:"embedding,omitempty"`
}

type SnapshotMetadata struct {
	Purpose    string    `json:"purpose"`
	Confidence float64   `json:"confidence"`
	Quality    float64   `json:"quality"`
	CreatedAt  time.Time `json:"created_at"`
}

type Snapshot struct {
	ID        string             `json:"id"`
	Timestamp int64              `json:"timestamp"`
	Message   string             `json:"message,omitempty"`
	Parent    string             `json:"parent,omitempty"`
	Symbols   map[string]*Symbol `json:"symbols"`
	CodeStore map[string]string  `json:"code_store"`
	Metadata  SnapshotMetadata   `json:"metadata"`
}

type ChangeType int

const (
	ChangeAdded ChangeType = iota
	ChangeModified
	ChangeRemoved
	ChangeSignatureChanged
	ChangeLogicChanged
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeAdded:
		return "added"
	case ChangeModified:
		return "modified"
	case ChangeRemoved:
		return "removed"
	case ChangeSignatureChanged:
		return "signature_changed"
	case ChangeLogicChanged:
		return "logic_changed"
	default:
		return "unknown"
	}
}

type SemanticChange struct {
	Symbol string      `json:"symbol"`
	Type   ChangeType  `json:"type"`
	Before interface{} `json:"before,omitempty"`
	After  interface{} `json:"after,omitempty"`
	Reason string      `json:"reason,omitempty"`
}

type SemanticDiff struct {
	FromSnapshot    string           `json:"from_snapshot"`
	ToSnapshot      string           `json:"to_snapshot"`
	SymbolsAdded    []string         `json:"symbols_added"`
	SymbolsModified []string         `json:"symbols_modified"`
	SymbolsRemoved  []string         `json:"symbols_removed"`
	Changes         []SemanticChange `json:"changes"`
}

type EdgeType int

const (
	EdgeCalls EdgeType = iota
	EdgeImports
	EdgeInherits
	EdgeImplements
	EdgeReferences
)

func (et EdgeType) String() string {
	switch et {
	case EdgeCalls:
		return "calls"
	case EdgeImports:
		return "imports"
	case EdgeInherits:
		return "inherits"
	case EdgeImplements:
		return "implements"
	case EdgeReferences:
		return "references"
	default:
		return "unknown"
	}
}

type SymbolEdge struct {
	From    string   `json:"from"`
	To      string   `json:"to"`
	Type    EdgeType `json:"type"`
	Weight  float64  `json:"weight"`
	Context string   `json:"context,omitempty"`
}

type ImpactAnalysis struct {
	DirectImpact   []string `json:"direct_impact"`
	IndirectImpact []string `json:"indirect_impact"`
	TestsAffected  []string `json:"tests_affected"`
	RiskLevel      string   `json:"risk_level"`
}

type QualityAssessment struct {
	Complexity      ComplexityAssessment `json:"complexity"`
	Testability     float64              `json:"testability"`
	Maintainability float64              `json:"maintainability"`
	Security        float64              `json:"security"`
	OverallScore    float64              `json:"overall_score"`
}

type ComplexityAssessment struct {
	Cyclomatic int    `json:"cyclomatic"`
	Cognitive  int    `json:"cognitive"`
	Rating     string `json:"rating"`
}
