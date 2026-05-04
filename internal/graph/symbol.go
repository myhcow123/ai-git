package graph

import (
	"container/heap"
	"math"
	"sync"

	"github.com/mychow/ai-git/pkg/types"
)

type SymbolGraph struct {
	mu sync.RWMutex

	Nodes map[string]*SymbolNode
	Edges map[string][]*SymbolEdge

	nodeIndex map[string]string
}

type SymbolNode struct {
	ID     string
	Symbol *types.Symbol
	Weight float64
}

type SymbolEdge struct {
	From    string
	To      string
	Type    types.EdgeType
	Weight  float64
	Context string
}

type PriorityQueueItem struct {
	ID       string
	Priority float64
	Index    int
}

type PriorityQueue []*PriorityQueueItem

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PriorityQueueItem)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

func NewSymbolGraph() *SymbolGraph {
	return &SymbolGraph{
		Nodes:     make(map[string]*SymbolNode),
		Edges:     make(map[string][]*SymbolEdge),
		nodeIndex: make(map[string]string),
	}
}

func (g *SymbolGraph) AddNode(symbol *types.Symbol) {
	g.mu.Lock()
	defer g.mu.Unlock()

	node := &SymbolNode{
		ID:     symbol.ID,
		Symbol: symbol,
		Weight: 1.0 / float64(len(g.Nodes)+1),
	}

	g.Nodes[symbol.ID] = node
	g.nodeIndex[symbol.Name] = symbol.ID
}

func (g *SymbolGraph) AddEdge(from, to string, edgeType types.EdgeType, weight float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	edge := &SymbolEdge{
		From:   from,
		To:     to,
		Type:   edgeType,
		Weight: weight,
	}

	g.Edges[from] = append(g.Edges[from], edge)
}

func (g *SymbolGraph) GetNode(id string) (*SymbolNode, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.Nodes[id]
	return node, exists
}

func (g *SymbolGraph) GetEdges(id string) []*SymbolEdge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := g.Edges[id]
	result := make([]*SymbolEdge, len(edges))
	copy(result, edges)
	return result
}

func (g *SymbolGraph) PageRank(iterations int, dampingFactor float64) map[string]float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()

	n := len(g.Nodes)
	if n == 0 {
		return make(map[string]float64)
	}

	scores := make(map[string]float64)
	newScores := make(map[string]float64)

	initialScore := 1.0 / float64(n)
	for id := range g.Nodes {
		scores[id] = initialScore
	}

	for i := 0; i < iterations; i++ {
		for id := range g.Nodes {
			sum := 0.0

			for _, edge := range g.Edges[id] {
				if edge.Type == types.EdgeCalls {
					fromNode := g.Nodes[edge.From]
					if fromNode != nil {
						outDegree := float64(len(g.Edges[edge.From]))
						if outDegree > 0 {
							sum += scores[edge.From] / outDegree
						}
					}
				}
			}

			newScores[id] = (1.0-dampingFactor)/float64(n) + dampingFactor*sum
		}

		for id, score := range newScores {
			scores[id] = score
		}
	}

	for id, node := range g.Nodes {
		node.Weight = scores[id]
	}

	return scores
}

func (g *SymbolGraph) AnalyzeImpact(symbolID string, depth int) *types.ImpactAnalysis {
	g.mu.RLock()
	defer g.mu.RUnlock()

	impact := &types.ImpactAnalysis{
		DirectImpact:   []string{},
		IndirectImpact: []string{},
		TestsAffected:  []string{},
		RiskLevel:      "low",
	}

	visited := make(map[string]bool)
	queue := []string{symbolID}

	for d := 0; d < depth; d++ {
		nextQueue := []string{}

		for _, currentID := range queue {
			if visited[currentID] {
				continue
			}
			visited[currentID] = true

			for _, edge := range g.Edges[currentID] {
				if edge.Type == types.EdgeCalls {
					caller := edge.From

					if d == 0 {
						impact.DirectImpact = append(impact.DirectImpact, caller)
					} else {
						impact.IndirectImpact = append(impact.IndirectImpact, caller)
					}

					if g.isTest(caller) {
						impact.TestsAffected = append(impact.TestsAffected, caller)
					}

					nextQueue = append(nextQueue, caller)
				}
			}
		}

		queue = nextQueue
	}

	impact.RiskLevel = g.calculateRiskLevel(impact)

	return impact
}

func (g *SymbolGraph) FindCallPath(from, to string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	dist := make(map[string]float64)
	prev := make(map[string]string)
	visited := make(map[string]bool)

	for id := range g.Nodes {
		dist[id] = math.Inf(1)
	}
	dist[from] = 0

	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	heap.Push(&pq, &PriorityQueueItem{ID: from, Priority: 0})

	for pq.Len() > 0 {
		current := heap.Pop(&pq).(*PriorityQueueItem)

		if visited[current.ID] {
			continue
		}
		visited[current.ID] = true

		if current.ID == to {
			return g.reconstructPath(prev, to)
		}

		for _, edge := range g.Edges[current.ID] {
			if edge.Type == types.EdgeCalls {
				alt := dist[current.ID] + edge.Weight
				if alt < dist[edge.To] {
					dist[edge.To] = alt
					prev[edge.To] = current.ID
					heap.Push(&pq, &PriorityQueueItem{ID: edge.To, Priority: alt})
				}
			}
		}
	}

	return nil
}

func (g *SymbolGraph) GetCallers(symbolID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var callers []string
	for _, edge := range g.Edges[symbolID] {
		if edge.Type == types.EdgeCalls {
			callers = append(callers, edge.From)
		}
	}
	return callers
}

func (g *SymbolGraph) GetCallees(symbolID string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var callees []string
	for _, edge := range g.Edges[symbolID] {
		if edge.Type == types.EdgeCalls {
			callees = append(callees, edge.To)
		}
	}
	return callees
}

func (g *SymbolGraph) isTest(symbolID string) bool {
	node, exists := g.Nodes[symbolID]
	if !exists {
		return false
	}

	return node.Symbol.Type == types.SymbolFunction &&
		(len(node.Symbol.Name) > 4 && node.Symbol.Name[:4] == "Test")
}

func (g *SymbolGraph) calculateRiskLevel(impact *types.ImpactAnalysis) string {
	total := len(impact.DirectImpact) + len(impact.IndirectImpact)
	testRatio := float64(len(impact.TestsAffected)) / float64(total+1)

	if total > 10 || testRatio < 0.2 {
		return "high"
	} else if total > 5 || testRatio < 0.5 {
		return "medium"
	}
	return "low"
}

func (g *SymbolGraph) reconstructPath(prev map[string]string, to string) []string {
	path := []string{}
	for at := to; at != ""; at = prev[at] {
		path = append([]string{at}, path...)
	}
	return path
}

func (g *SymbolGraph) GetTopSymbols(limit int) []*SymbolNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]*SymbolNode, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		nodes = append(nodes, node)
	}

	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[j].Weight > nodes[i].Weight {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}

	if limit > len(nodes) {
		limit = len(nodes)
	}

	return nodes[:limit]
}

func (g *SymbolGraph) GetStats() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edgeTypes := make(map[types.EdgeType]int)
	for _, edges := range g.Edges {
		for _, edge := range edges {
			edgeTypes[edge.Type]++
		}
	}

	return map[string]interface{}{
		"total_nodes": len(g.Nodes),
		"total_edges": len(g.Edges),
		"edge_types":  edgeTypes,
	}
}
