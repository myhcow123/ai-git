package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"reflect"
	"sync"
)

type Plugin interface {
	Name() string
	Version() string
	Description() string
	Initialize(ctx *Context) error
	Execute(ctx *Context, args map[string]interface{}) (interface{}, error)
	Shutdown() error
}

type Context struct {
	ProjectRoot string
	Config      map[string]interface{}
	Storage     interface{}
	Index       interface{}
	Graph       interface{}
	Logger      Logger
}

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type Manager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
	ctx     *Context
}

func NewManager(ctx *Context) *Manager {
	return &Manager{
		plugins: make(map[string]Plugin),
		ctx:     ctx,
	}
}

func (m *Manager) Register(p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.plugins[p.Name()]; exists {
		return fmt.Errorf("plugin %s already registered", p.Name())
	}

	if err := p.Initialize(m.ctx); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", p.Name(), err)
	}

	m.plugins[p.Name()] = p
	return nil
}

func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	if err := p.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugin %s: %w", name, err)
	}

	delete(m.plugins, name)
	return nil
}

func (m *Manager) Get(name string) (Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, exists := m.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return p, nil
}

func (m *Manager) List() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []PluginInfo
	for name, p := range m.plugins {
		list = append(list, PluginInfo{
			Name:        name,
			Version:     p.Version(),
			Description: p.Description(),
		})
	}
	return list
}

func (m *Manager) Execute(name string, args map[string]interface{}) (interface{}, error) {
	m.mu.RLock()
	p, exists := m.plugins[name]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return p.Execute(m.ctx, args)
}

func (m *Manager) LoadFromPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat plugin path: %w", err)
	}

	if info.IsDir() {
		return m.loadFromDir(path)
	}
	return m.loadFromFile(path)
}

func (m *Manager) loadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) == ".so" {
			pluginPath := filepath.Join(dir, name)
			if err := m.loadFromFile(pluginPath); err != nil {
				m.ctx.Logger.Error("Failed to load plugin %s: %v", name, err)
			}
		}
	}

	return nil
}

func (m *Manager) loadFromFile(path string) error {
	plug, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	sym, err := plug.Lookup("Plugin")
	if err != nil {
		return fmt.Errorf("failed to lookup Plugin symbol: %w", err)
	}

	p, ok := sym.(Plugin)
	if !ok {
		return fmt.Errorf("plugin does not implement Plugin interface: %T", sym)
	}

	return m.Register(p)
}

func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, p := range m.plugins {
		if err := p.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown %s: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

type HookType string

const (
	BeforeParse HookType = "before_parse"
	AfterParse  HookType = "after_parse"
	BeforeIndex HookType = "before_index"
	AfterIndex  HookType = "after_index"
	BeforeQuery HookType = "before_query"
	AfterQuery  HookType = "after_query"
)

type Hook func(ctx *Context, data interface{}) (interface{}, error)

type HookRegistry struct {
	hooks map[HookType][]Hook
	mu    sync.RWMutex
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make(map[HookType][]Hook),
	}
}

func (r *HookRegistry) Register(hookType HookType, hook Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.hooks[hookType] = append(r.hooks[hookType], hook)
}

func (r *HookRegistry) Execute(hookType HookType, ctx *Context, data interface{}) (interface{}, error) {
	r.mu.RLock()
	hooks := r.hooks[hookType]
	r.mu.RUnlock()

	var result interface{} = data
	var err error

	for _, hook := range hooks {
		result, err = hook(ctx, result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

type BasePlugin struct {
	name        string
	version     string
	description string
}

func (p *BasePlugin) Name() string        { return p.name }
func (p *BasePlugin) Version() string     { return p.version }
func (p *BasePlugin) Description() string { return p.description }
func (p *BasePlugin) Initialize(ctx *Context) error {
	return nil
}
func (p *BasePlugin) Shutdown() error {
	return nil
}

func NewBasePlugin(name, version, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
	}
}

type PluginBuilder struct {
	plugin      *BasePlugin
	executeFunc func(ctx *Context, args map[string]interface{}) (interface{}, error)
}

func NewPluginBuilder(name string) *PluginBuilder {
	return &PluginBuilder{
		plugin: NewBasePlugin(name, "1.0.0", ""),
	}
}

func (b *PluginBuilder) Version(version string) *PluginBuilder {
	b.plugin.version = version
	return b
}

func (b *PluginBuilder) Description(desc string) *PluginBuilder {
	b.plugin.description = desc
	return b
}

func (b *PluginBuilder) Execute(fn func(ctx *Context, args map[string]interface{}) (interface{}, error)) *PluginBuilder {
	b.executeFunc = fn
	return b
}

func (b *PluginBuilder) Build() Plugin {
	return &builtPlugin{
		BasePlugin:  b.plugin,
		executeFunc: b.executeFunc,
	}
}

type builtPlugin struct {
	*BasePlugin
	executeFunc func(ctx *Context, args map[string]interface{}) (interface{}, error)
}

func (p *builtPlugin) Execute(ctx *Context, args map[string]interface{}) (interface{}, error) {
	if p.executeFunc == nil {
		return nil, fmt.Errorf("no execute function defined")
	}
	return p.executeFunc(ctx, args)
}

func ValidatePlugin(p Plugin) error {
	if p.Name() == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	if p.Version() == "" {
		return fmt.Errorf("plugin version cannot be empty")
	}
	return nil
}

func GetPluginInfo(p Plugin) map[string]interface{} {
	info := map[string]interface{}{
		"name":        p.Name(),
		"version":     p.Version(),
		"description": p.Description(),
		"type":        reflect.TypeOf(p).String(),
	}
	return info
}
