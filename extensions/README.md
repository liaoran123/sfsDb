# Extensions Package

扩展包提供了sfsDb的扩展系统，包括插件管理、索引扩展、存储引擎扩展和数据类型扩展。

## 功能特性

- **插件系统**：支持动态加载和管理插件
- **索引扩展**：支持自定义索引类型和实现
- **存储引擎扩展**：支持多种存储后端
- **数据类型扩展**：支持自定义数据类型
- **灵活的注册机制**：基于注册中心的扩展管理

## 核心模块

### 1. Plugins

插件模块，负责插件的加载、管理和生命周期控制。

#### 主要接口

```go
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
```

### 2. Indexes

索引模块，负责索引的注册和管理。

#### 主要接口

```go
// Index 索引接口
type Index interface {
	// Name 获取索引名称
	Name() string

	// AddFields 添加索引字段
	AddFields(fields ...string)

	// GetFields 获取索引字段
	GetFields() []string

	// JoinValue 连接索引值
	JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte

	// MatchFields 匹配索引字段
	MatchFields(fields ...string) bool

	// Parse 解析索引值
	Parse(fields []string, value []byte) *map[string][]byte
}

// IndexRegistry 索引注册中心
type IndexRegistry struct {
	factories map[string]IndexFactory
}

// RegisterFactory 注册索引工厂
func (r *IndexRegistry) RegisterFactory(name string, factory IndexFactory) error

// CreateIndex 创建索引
func (r *IndexRegistry) CreateIndex(factoryName, name string, fields ...string) (Index, error)
```

### 3. Storage

存储模块，负责存储引擎的注册和管理。

#### 主要接口

```go
// Store 存储接口
type Store interface {
	// Put 存储键值对
	Put(key, value []byte) error

	// Get 获取值
	Get(key []byte) ([]byte, error)

	// Delete 删除键值对
	Delete(key []byte) error

	// Iterator 创建迭代器
	Iterator(start, limit []byte) Iterator

	// Snapshot 创建快照
	Snapshot() (Snapshot, error)

	// Close 关闭存储
	Close() error
}

// StoreRegistry 存储注册中心
type StoreRegistry struct {
	factories map[string]StoreFactory
}

// RegisterFactory 注册存储工厂
func (r *StoreRegistry) RegisterFactory(name string, factory StoreFactory) error

// CreateStore 创建存储
func (r *StoreRegistry) CreateStore(factoryName, path string) (Store, error)
```

### 4. Types

类型模块，负责数据类型的注册和管理。

#### 主要接口

```go
// Type 数据类型接口
type Type interface {
	// Name 获取类型名称
	Name() string

	// Size 获取类型大小
	Size() int

	// Encode 编码值
	Encode(value interface{}) ([]byte, error)

	// Decode 解码值
	Decode(data []byte) (interface{}, error)

	// Validate 验证值
	Validate(value interface{}) error
}

// TypeRegistry 类型注册中心
type TypeRegistry struct {
	types map[string]Type
}

// RegisterType 注册类型
func (r *TypeRegistry) RegisterType(name string, typ Type) error

// Encode 编码值
func (r *TypeRegistry) Encode(typeName string, value interface{}) ([]byte, error)
```

## 使用示例

### 1. 初始化扩展系统

```go
import (
	"github.com/liaoran123/sfsDb/extensions"
)

// 创建扩展配置
config := extensions.ExtensionConfig{
	Plugins: plugins.PluginConfig{
		Enabled: true,
		Dir:     "plugins",
		Plugins: []string{
			"plugin1.so",
			"plugin2.so",
		},
	},
}

// 创建扩展系统
ext := extensions.NewExtensions(config)

// 启动扩展系统
ext.Start()
```

### 2. 管理插件

```go
// 加载插件
plugin, err := ext.Plugins().Load("plugins/plugin1.so")
if err != nil {
	// 加载失败
	return
}

// 启动插件
plugin.Start()

// 列出所有插件
plugins := ext.Plugins().ListPlugins()
for _, p := range plugins {
	fmt.Printf("Plugin: %s (v%s) - %s\n", p.Name(), p.Version(), p.Description())
}

// 停止插件
plugin.Stop()

// 卸载插件
ext.Plugins().Unload(plugin.Name())
```

### 3. 扩展索引

```go
// 注册自定义索引工厂
indexRegistry := ext.Indexes()
indexRegistry.RegisterFactory("custom", &CustomIndexFactory{})

// 创建自定义索引
index, err := indexRegistry.CreateIndex("custom", "my_index", "field1", "field2")
if err != nil {
	// 创建失败
	return
}

// 使用索引
fieldsBytes := &map[string][]byte{
	"field1": []byte("value1"),
	"field2": []byte("value2"),
}
indexValue := index.JoinValue(fieldsBytes, 1)
```

### 4. 扩展存储引擎

```go
// 注册自定义存储工厂
storeRegistry := ext.Storage()
storeRegistry.RegisterFactory("custom", &CustomStoreFactory{})

// 创建自定义存储
store, err := storeRegistry.CreateStore("custom", "data")
if err != nil {
	// 创建失败
	return
}

// 使用存储
store.Put([]byte("key"), []byte("value"))
value, err := store.Get([]byte("key"))
if err != nil {
	// 获取失败
	return
}
fmt.Println(string(value))

// 关闭存储
store.Close()
```

### 5. 扩展数据类型

```go
// 注册自定义类型
typeRegistry := ext.Types()
typeRegistry.RegisterType("custom", &CustomType{})

// 使用自定义类型
value := "custom value"
data, err := typeRegistry.Encode("custom", value)
if err != nil {
	// 编码失败
	return
}

decoded, err := typeRegistry.Decode("custom", data)
if err != nil {
	// 解码失败
	return
}
fmt.Println(decoded)
```

## 配置选项

| 配置项 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| plugins.enabled | bool | false | 是否启用插件系统 |
| plugins.dir | string | "plugins" | 插件目录 |
| plugins.plugins | []string | [] | 要加载的插件列表 |

## 最佳实践

1. **插件设计**：插件应保持轻量，专注于单一功能
2. **错误处理**：插件应妥善处理错误，避免影响主系统
3. **资源管理**：插件应正确管理资源，避免资源泄漏
4. **版本控制**：插件应实现版本管理，支持升级和回滚
5. **文档完善**：为插件提供详细的文档和使用示例

## 故障排查

### 常见问题

1. **插件加载失败**
   - 检查插件文件是否存在
   - 检查插件是否实现了正确的接口
   - 检查插件依赖是否满足

2. **索引创建失败**
   - 检查索引工厂是否注册
   - 检查索引字段是否有效
   - 检查索引配置是否正确

3. **存储引擎初始化失败**
   - 检查存储路径是否可写
   - 检查存储配置是否正确
   - 检查存储引擎依赖是否满足

4. **数据类型编码/解码失败**
   - 检查类型是否注册
   - 检查值是否符合类型要求
   - 检查数据是否完整

## 版本历史

- **v1.0.0**：初始版本，支持插件系统和索引扩展
- **v1.1.0**：添加存储引擎扩展和数据类型扩展
- **v1.2.0**：优化扩展系统，提高稳定性和性能
