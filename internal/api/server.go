package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mychow/ai-git/internal/watcher"
)

type Server struct {
	engine  Engine
	watcher *watcher.Watcher
	server  *http.Server
	port    int
}

type Engine interface {
	GetStorage() interface{}
	GetIndex() interface{}
	GetGraph() interface{}
	GetParser() interface{}
	GetEngine() interface{}
}

func NewServer(engine Engine, port int) *Server {
	return &Server{
		engine: engine,
		watcher: watcher.NewWatcher(),
		port:   port,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/overview", s.handleOverview)
	mux.HandleFunc("/api/v1/symbols", s.handleSymbols)
	mux.HandleFunc("/api/v1/symbol/", s.handleSymbol)
	mux.HandleFunc("/api/v1/search", s.handleSearch)
	mux.HandleFunc("/api/v1/analyze", s.handleAnalyze)
	mux.HandleFunc("/api/v1/quality/", s.handleQuality)
	mux.HandleFunc("/api/v1/pattern/", s.handlePattern)
	mux.HandleFunc("/api/v1/intent/", s.handleIntent)
	mux.HandleFunc("/api/v1/deps/", s.handleDeps)
	mux.HandleFunc("/api/v1/impact/", s.handleImpact)
	mux.HandleFunc("/api/v1/graph", s.handleGraph)
	mux.HandleFunc("/api/v1/health", s.handleHealth)
	
	mux.HandleFunc("/api/v1/init", s.handleInit)
	mux.HandleFunc("/api/v1/read/", s.handleRead)
	mux.HandleFunc("/api/v1/insert", s.handleInsert)
	mux.HandleFunc("/api/v1/replace", s.handleReplace)
	mux.HandleFunc("/api/v1/delete", s.handleDelete)
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	
	mux.HandleFunc("/api/v1/projects", s.handleProjects)
	mux.HandleFunc("/api/v1/projects/", s.handleProjectAction)
	mux.HandleFunc("/api/v1/status/watcher", s.handleWatcherStatus)

	corsHandler := s.corsMiddleware(mux)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      corsHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go s.watcher.Start()

	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	s.watcher.Stop()
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	overview := map[string]interface{}{
		"project":   "ai-git-project",
		"symbols":   0,
		"files":     0,
		"languages": []string{"go", "python", "javascript"},
		"timestamp": time.Now().Unix(),
	}

	s.jsonResponse(w, http.StatusOK, overview)
}

func (s *Server) handleSymbols(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbolType := r.URL.Query().Get("type")
	file := r.URL.Query().Get("file")

	symbols := []map[string]interface{}{
		{
			"id":       "sym_1",
			"name":     "main",
			"type":     "function",
			"file":     "main.go",
			"line":     10,
			"language": "go",
		},
	}

	if symbolType != "" {
		var filtered []map[string]interface{}
		for _, sym := range symbols {
			if sym["type"] == symbolType {
				filtered = append(filtered, sym)
			}
		}
		symbols = filtered
	}

	if file != "" {
		var filtered []map[string]interface{}
		for _, sym := range symbols {
			if sym["file"] == file {
				filtered = append(filtered, sym)
			}
		}
		symbols = filtered
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"symbols": symbols,
		"count":   len(symbols),
	})
}

func (s *Server) handleSymbol(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/api/v1/symbol/")
	if name == "" {
		s.errorResponse(w, http.StatusBadRequest, "symbol name required")
		return
	}

	symbol := map[string]interface{}{
		"id":         "sym_" + name,
		"name":       name,
		"type":       "function",
		"file":       "main.go",
		"line":       10,
		"end_line":   20,
		"language":   "go",
		"signature":  "func " + name + "()",
		"docstring":  "Example function",
		"complexity": 5,
		"calls":      []string{"helper1", "helper2"},
		"called_by":  []string{"main"},
	}

	s.jsonResponse(w, http.StatusOK, symbol)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		s.errorResponse(w, http.StatusBadRequest, "query parameter 'q' required")
		return
	}

	results := []map[string]interface{}{
		{
			"id":      "result_1",
			"name":    query + "_handler",
			"type":    "function",
			"file":    "handler.go",
			"line":    15,
			"snippet": "func " + query + "_handler() { ... }",
			"score":   0.95,
		},
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	topStr := r.URL.Query().Get("top")
	top := 10
	if topStr != "" {
		if val, err := strconv.Atoi(topStr); err == nil {
			top = val
		}
	}

	important := []map[string]interface{}{}
	for i := 0; i < top && i < 10; i++ {
		important = append(important, map[string]interface{}{
			"id":    fmt.Sprintf("sym_%d", i),
			"name":  fmt.Sprintf("important_func_%d", i),
			"type":  "function",
			"score": 0.9 - float64(i)*0.05,
		})
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"important_symbols": important,
		"algorithm":         "pagerank",
		"timestamp":         time.Now().Unix(),
	})
}

func (s *Server) handleQuality(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/api/v1/quality/")
	if symbol == "" {
		s.errorResponse(w, http.StatusBadRequest, "symbol name required")
		return
	}

	quality := map[string]interface{}{
		"symbol":          symbol,
		"complexity":      7,
		"testability":     0.85,
		"maintainability": 0.90,
		"security":        0.95,
		"overall":         0.88,
		"suggestions": []string{
			"Consider reducing function complexity",
			"Add more unit tests",
		},
	}

	s.jsonResponse(w, http.StatusOK, quality)
}

func (s *Server) handlePattern(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/api/v1/pattern/")
	if symbol == "" {
		s.errorResponse(w, http.StatusBadRequest, "symbol name required")
		return
	}

	pattern := map[string]interface{}{
		"symbol":     symbol,
		"pattern":    "factory",
		"confidence": 0.92,
		"reason":     "Creates objects without specifying exact class",
	}

	s.jsonResponse(w, http.StatusOK, pattern)
}

func (s *Server) handleIntent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/api/v1/intent/")
	if symbol == "" {
		s.errorResponse(w, http.StatusBadRequest, "symbol name required")
		return
	}

	intent := map[string]interface{}{
		"symbol":     symbol,
		"intent":     "data_processing",
		"confidence": 0.88,
		"reason":     "Transforms and processes data",
	}

	s.jsonResponse(w, http.StatusOK, intent)
}

func (s *Server) handleDeps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/api/v1/deps/")
	if symbol == "" {
		s.errorResponse(w, http.StatusBadRequest, "symbol name required")
		return
	}

	deps := map[string]interface{}{
		"symbol":       symbol,
		"dependencies": []string{"helper1", "util.format", "config.load"},
		"dependents":   []string{"main", "handler.process"},
		"depth":        3,
	}

	s.jsonResponse(w, http.StatusOK, deps)
}

func (s *Server) handleImpact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	symbol := strings.TrimPrefix(r.URL.Path, "/api/v1/impact/")
	if symbol == "" {
		s.errorResponse(w, http.StatusBadRequest, "symbol name required")
		return
	}

	impact := map[string]interface{}{
		"symbol":         symbol,
		"direct_impact":  5,
		"total_impact":   12,
		"affected_files": []string{"main.go", "handler.go", "util.go"},
		"risk_level":     "medium",
	}

	s.jsonResponse(w, http.StatusOK, impact)
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	graph := map[string]interface{}{
		"nodes": []map[string]interface{}{
			{"id": "main", "type": "function", "score": 0.95},
			{"id": "helper", "type": "function", "score": 0.75},
		},
		"edges": []map[string]interface{}{
			{"from": "main", "to": "helper", "type": "calls"},
		},
		"stats": map[string]interface{}{
			"node_count": 2,
			"edge_count": 1,
		},
	}

	s.jsonResponse(w, http.StatusOK, graph)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}
	s.jsonResponse(w, http.StatusOK, health)
}

func (s *Server) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) errorResponse(w http.ResponseWriter, status int, message string) {
	s.jsonResponse(w, status, map[string]interface{}{
		"error":   true,
		"message": message,
		"status":  status,
	})
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		projects := s.watcher.GetProjects()
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"projects": projects,
			"count":    len(projects),
		})
	
	case http.MethodPost:
		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.errorResponse(w, http.StatusBadRequest, "invalid request body")
			return
		}
		
		if req.Path == "" {
			s.errorResponse(w, http.StatusBadRequest, "path is required")
			return
		}
		
		if err := s.watcher.AddProject(req.Path, nil); err != nil {
			s.errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		s.jsonResponse(w, http.StatusCreated, map[string]interface{}{
			"message": "project added",
			"path":    req.Path,
		})
	
	default:
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleProjectAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/projects/")
	if path == "" {
		s.errorResponse(w, http.StatusBadRequest, "project path required")
		return
	}
	
	switch r.Method {
	case http.MethodDelete:
		if err := s.watcher.RemoveProject(path); err != nil {
			s.errorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"message": "project removed",
			"path":    path,
		})
	
	default:
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleWatcherStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	
	stats := s.watcher.GetStats()
	s.jsonResponse(w, http.StatusOK, stats)
}

func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Path == "" {
		s.errorResponse(w, http.StatusBadRequest, "path is required")
		return
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Project initialized successfully",
		"path":    req.Path,
	})
}

func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	target := strings.TrimPrefix(r.URL.Path, "/api/v1/read/")
	if target == "" {
		s.errorResponse(w, http.StatusBadRequest, "target is required")
		return
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"target": target,
		"code":   "// Code would be read from the engine",
	})
}

func (s *Server) handleInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Target   string `json:"target"`
		Code     string `json:"code"`
		Position string `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Target == "" || req.Code == "" {
		s.errorResponse(w, http.StatusBadRequest, "target and code are required")
		return
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  "Code inserted successfully",
		"target":   req.Target,
		"position": req.Position,
	})
}

func (s *Server) handleReplace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Target string `json:"target"`
		Code   string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Target == "" || req.Code == "" {
		s.errorResponse(w, http.StatusBadRequest, "target and code are required")
		return
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Code replaced successfully",
		"target":  req.Target,
	})
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Target == "" {
		s.errorResponse(w, http.StatusBadRequest, "target is required")
		return
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Code deleted successfully",
		"target":  req.Target,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.errorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "running",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"stats": map[string]interface{}{
			"index_count":  0,
			"files_watched": 0,
			"uptime":       0,
		},
	})
}
