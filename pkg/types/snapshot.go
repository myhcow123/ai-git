package types

type SnapshotInfo struct {
	ID        string
	Timestamp int64
	Message   string
	Symbols   map[string]*Symbol
}

type SnapshotDiff struct {
	FromSnapshot    string
	ToSnapshot      string
	SymbolsAdded    []string
	SymbolsRemoved  []string
	SymbolsModified []string
	Changes         []Change
}

type Change struct {
	Symbol string
	Type   string
	Before interface{}
	After  interface{}
}
