package plugins

import (
	"errors"
	"fmt"
	"plugin"
	"sync"
)

// Plugin 插件接口
type Plugin interface {
	// Name 获取插件名称
	Name() string

	// Version 获取插件版本
	Version() string

	// Description 获取插件描述
	Description() string

	// Init 初始化插件
	Init() error

	// Start 启动插件
	Start() error

	// Stop 停止插件
	Stop() error

	// IsRunning 检查插件是否运行中
	IsRunning() bool
}

// PluginManager 插件管理器接口
type PluginManager interface {
	// Load 加载插件
	Load(path string) (Plugin, error)

	// Unload 卸载插件
	Unload(name string) error

	// GetPlugin 获取插件
	GetPlugin(name string) (Plugin, error)

	// ListPlugins 列出所有插件
	ListPlugins() []Plugin

	// StartAll 启动所有插件
	StartAll() error

	// StopAll 停止所有插件
	StopAll() error
}

// PluginConfig 插件配置
type PluginConfig struct {
	Enabled bool     `json:"enabled"`
	Dir     string   `json:"dir"`
	Plugins []string `json:"plugins"`
}

// DefaultPluginManager 默认插件管理器
type DefaultPluginManager struct {
	config     PluginConfig
	plugins    map[string]Plugin
	mutex      sync.RWMutex
}

// NewPluginManager 创建插件管理器
func NewPluginManager(config PluginConfig) PluginManager {
	return &DefaultPluginManager{
		config:  config,
		plugins: make(map[string]Plugin),
	}
}

// Load 加载插件
func (pm *DefaultPluginManager) Load(path string) (Plugin, error) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 加载插件
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	// 查找插件导出符号
	symbol, err := p.Lookup("Plugin")
	if err != nil {
		return nil, err
	}

	// 类型断言
	pluginInstance, ok := symbol.(Plugin)
	if !ok {
		return nil, errors.New("invalid plugin type")
	}

	// 初始化插件
	if err := pluginInstance.Init(); err != nil {
		return nil, err
	}

	// 存储插件
	pm.plugins[pluginInstance.Name()] = pluginInstance

	return pluginInstance, nil
}

// Unload 卸载插件
func (pm *DefaultPluginManager) Unload(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 获取插件
	pluginInstance, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// 停止插件
	if err := pluginInstance.Stop(); err != nil {
		return err
	}

	// 删除插件
	delete(pm.plugins, name)

	return nil
}

// GetPlugin 获取插件
func (pm *DefaultPluginManager) GetPlugin(name string) (Plugin, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	pluginInstance, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return pluginInstance, nil
}

// ListPlugins 列出所有插件
func (pm *DefaultPluginManager) ListPlugins() []Plugin {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, pluginInstance := range pm.plugins {
		plugins = append(plugins, pluginInstance)
	}

	return plugins
}

// StartAll 启动所有插件
func (pm *DefaultPluginManager) StartAll() error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, pluginInstance := range pm.plugins {
		if err := pluginInstance.Start(); err != nil {
			return err
		}
	}

	return nil
}

// StopAll 停止所有插件
func (pm *DefaultPluginManager) StopAll() error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	for _, pluginInstance := range pm.plugins {
		if err := pluginInstance.Stop(); err != nil {
			return err
		}
	}

	return nil
}

// BasePlugin 基础插件实现
type BasePlugin struct {
	name        string
	version     string
	description string
	running     bool
	mutex       sync.RWMutex
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(name, version, description string) *BasePlugin {
	return &BasePlugin{
		name:        name,
		version:     version,
		description: description,
		running:     false,
	}
}

// Name 获取插件名称
func (p *BasePlugin) Name() string {
	return p.name
}

// Version 获取插件版本
func (p *BasePlugin) Version() string {
	return p.version
}

// Description 获取插件描述
func (p *BasePlugin) Description() string {
	return p.description
}

// Init 初始化插件
func (p *BasePlugin) Init() error {
	return nil
}

// Start 启动插件
func (p *BasePlugin) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.running = true
	return nil
}

// Stop 停止插件
func (p *BasePlugin) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.running = false
	return nil
}

// IsRunning 检查插件是否运行中
func (p *BasePlugin) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.running
}
