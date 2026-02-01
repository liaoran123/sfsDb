package extensions

import (
	"github.com/liaoran123/sfsDb/extensions/indexes"
	"github.com/liaoran123/sfsDb/extensions/plugins"
	"github.com/liaoran123/sfsDb/extensions/storage"
	"github.com/liaoran123/sfsDb/extensions/types"
)

// Extensions 扩展系统接口
type Extensions interface {
	// 插件相关
	Plugins() plugins.PluginManager

	// 索引相关
	Indexes() *indexes.IndexRegistry

	// 存储相关
	Storage() *storage.StoreRegistry

	// 类型相关
	Types() *types.TypeRegistry

	// 启动扩展系统
	Start() error

	// 停止扩展系统
	Stop() error
}

// ExtensionConfig 扩展系统配置
type ExtensionConfig struct {
	Plugins plugins.PluginConfig `json:"plugins"`
}

// Extension 扩展系统实现
type Extension struct {
	config        ExtensionConfig
	pluginManager plugins.PluginManager
	indexRegistry *indexes.IndexRegistry
	storeRegistry *storage.StoreRegistry
	typeRegistry  *types.TypeRegistry
}

// NewExtensions 创建扩展系统
func NewExtensions(config ExtensionConfig) Extensions {
	pluginManager := plugins.NewPluginManager(config.Plugins)
	indexRegistry := indexes.NewIndexRegistry()
	storeRegistry := storage.NewStoreRegistry()
	typeRegistry := types.NewTypeRegistry()

	return &Extension{
		config:        config,
		pluginManager: pluginManager,
		indexRegistry: indexRegistry,
		storeRegistry: storeRegistry,
		typeRegistry:  typeRegistry,
	}
}

// Plugins 获取插件管理器
func (e *Extension) Plugins() plugins.PluginManager {
	return e.pluginManager
}

// Indexes 获取索引注册中心
func (e *Extension) Indexes() *indexes.IndexRegistry {
	return e.indexRegistry
}

// Storage 获取存储注册中心
func (e *Extension) Storage() *storage.StoreRegistry {
	return e.storeRegistry
}

// Types 获取类型注册中心
func (e *Extension) Types() *types.TypeRegistry {
	return e.typeRegistry
}

// Start 启动扩展系统
func (e *Extension) Start() error {
	// 启动插件系统
	return e.pluginManager.StartAll()
}

// Stop 停止扩展系统
func (e *Extension) Stop() error {
	// 停止插件系统
	return e.pluginManager.StopAll()
}

// DefaultExtensionConfig 默认扩展系统配置
func DefaultExtensionConfig() ExtensionConfig {
	return ExtensionConfig{
		Plugins: plugins.PluginConfig{
			Enabled: false,
			Dir:     "plugins",
			Plugins: []string{},
		},
	}
}
