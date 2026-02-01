# Internal API Documentation

本文档描述了sfsDb的内部API，以及如何与外部系统集成。

## 监控系统 API

### 核心功能

#### 1. 指标收集

**API接口**：
```go
// Metrics 指标收集器接口
type Metrics interface {
	// 事务指标
	RecordTransaction(duration time.Duration, success bool)
	GetTransactionCount() int64
	GetTransactionSuccessRate() float64
	GetAverageTransactionDuration() time.Duration

	// 查询指标
	RecordQuery(duration time.Duration, queryType string)
	GetQueryCount() int64
	GetAverageQueryDuration() time.Duration

	// 存储指标
	RecordStorageOperation(opType string, duration time.Duration)
	GetStorageOperationCount() int64

	// 系统指标
	RecordSystemState(concurrency int32, memoryUsage uint64, requests int)
	GetCurrentSystemState() (int32, uint64, int)

	// 重置指标
	Reset()
}
```

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/monitoring/metrics"

// 创建指标收集器
m := metrics.NewDefaultMetrics()

// 记录事务
m.RecordTransaction(time.Since(start), success)

// 记录查询
m.RecordQuery(time.Since(start), "SELECT")

// 记录系统状态
m.RecordSystemState(concurrency, memoryUsage, requests)
```

#### 2. 数据导出接口

**API接口**：
```go
// PrometheusExporter Prometheus指标导出器
type PrometheusExporter struct {
	// ...
}

// NewPrometheusExporter 创建Prometheus指标导出器
func NewPrometheusExporter(metrics metrics.Metrics) *PrometheusExporter

// StartHTTP 启动HTTP服务器，暴露Prometheus指标
func (e *PrometheusExporter) StartHTTP(addr string) error

// Update 更新Prometheus指标
func (e *PrometheusExporter) Update()
```

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/monitoring/exporters"

// 创建Prometheus导出器
exporter := exporters.NewPrometheusExporter(metrics)

// 启动HTTP服务器
go exporter.StartHTTP(":9090")

// 定期更新指标
go func() {
	for {
		exporter.Update()
		time.Sleep(10 * time.Second)
	}
}()
```

#### 3. 内部事件监听

**使用方式**：
在核心代码中嵌入事件监听点，收集内部事件。

```go
// 在事务提交时
func (tx *TableTransaction) Commit() error {
	start := time.Now()
	// 提交逻辑...
	success := err == nil
	// 记录事务指标
	metrics.RecordTransaction(time.Since(start), success)
	return err
}

// 在查询执行时
func (t *Table) Search(indexName string, min, max any) (*TableIter, error) {
	start := time.Now()
	// 查询逻辑...
	// 记录查询指标
	metrics.RecordQuery(time.Since(start), "SEARCH")
	return iter, err
}
```

### 与外部系统集成

1. **Prometheus集成**：
   - 配置Prometheus抓取sfsDb的`/metrics`端点
   - 使用Grafana创建监控仪表盘

2. **告警系统集成**：
   - 配置Prometheus Alertmanager处理告警
   - 集成外部告警通知系统（如邮件、Slack等）

3. **监控数据存储**：
   - 使用InfluxDB等时间序列数据库存储监控数据
   - 配置长期数据保留策略

## 安全系统 API

### 核心功能

#### 1. 认证集成

**API接口**：
```go
// Authenticator 认证器接口
type Authenticator interface {
	// 认证用户
	Authenticate(username, password string) (*User, error)

	// 生成认证令牌
	GenerateToken(user *User) (string, error)

	// 验证令牌
	ValidateToken(token string) (*Claims, error)

	// 根据ID获取用户
	GetUserByID(userID string) (*User, error)
}
```

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/security/auth"

// 创建认证器
config := auth.AuthConfig{
	Enabled:     true,
	JWTSecret:   "secret",
	TokenExpiry: 1 * time.Hour,
	Users:       []auth.User{...},
}
a := auth.NewJWTAuthenticator(config)

// 验证令牌
claims, err := a.ValidateToken(token)
if err != nil {
	// 令牌无效
	return
}

// 获取用户
user, err := a.GetUserByID(claims.UserID)
if err != nil {
	// 用户不存在
	return
}
```

#### 2. 授权检查

**API接口**：
```go
// AccessControl 访问控制接口
type AccessControl interface {
	// 检查用户是否有指定权限
	CheckPermission(userID, permissionID string) (bool, error)

	// 检查用户是否有指定角色
	CheckRole(userID, roleID string) (bool, error)

	// 获取用户的所有角色
	GetUserRoles(userID string) ([]Role, error)

	// 获取角色的所有权限
	GetRolePermissions(roleID string) ([]Permission, error)

	// 为用户分配角色
	AssignRole(userID, roleID string) error

	// 撤销用户的角色
	RevokeRole(userID, roleID string) error
}
```

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/security/access"

// 创建访问控制器
config := access.AccessConfig{
	Enabled: true,
	Roles:   []access.Role{...},
	UserRoles: map[string][]string{...},
}
ac := access.NewRBACAccessControl(config)

// 检查权限
hasPermission, err := ac.CheckPermission(userID, "write")
if err != nil || !hasPermission {
	// 没有权限
	return
}

// 检查角色
hasRole, err := ac.CheckRole(userID, "admin")
if err != nil || !hasRole {
	// 没有角色
	return
}
```

#### 3. 数据加密

**API接口**：
```go
// Encryptor 加密器接口
type Encryptor interface {
	// 加密数据
	Encrypt(plaintext []byte) ([]byte, error)

	// 解密数据
	Decrypt(ciphertext []byte) ([]byte, error)

	// 获取加密算法名称
	Algorithm() string
}
```

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/security/encryption"

// 创建加密器
encryptor, err := encryption.NewAESGCMEncryptor([]byte("secret_key"))
if err != nil {
	return
}

// 加密数据
ciphertext, err := encryptor.Encrypt(plaintext)
if err != nil {
	return
}

// 解密数据
decrypted, err := encryptor.Decrypt(ciphertext)
if err != nil {
	return
}
```

#### 4. 审计事件生成

**API接口**：
```go
// AuditLogger 审计日志记录器接口
type AuditLogger interface {
	// 记录审计事件
	LogEvent(event AuditEvent)

	// 获取审计事件
	GetEvents(filter map[string]string, limit, offset int) ([]AuditEvent, error)

	// 根据ID获取审计事件
	GetEventByID(eventID string) (*AuditEvent, error)

	// 搜索审计事件
	SearchEvents(query string, limit, offset int) ([]AuditEvent, error)
}

// NewAuditEvent 创建审计事件
func NewAuditEvent(eventType, userID, username, action, resource, details string, success bool, ipAddress string) AuditEvent
```

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/security/audit"

// 创建审计日志记录器
config := audit.AuditConfig{
	Enabled:    true,
	MaxEvents:  10000,
	EventTypes: []string{"auth", "access", "data"},
}
logger := audit.NewConsoleAuditLogger(config)

// 记录审计事件
event := audit.NewAuditEvent(
	"access",
	userID,
	username,
	"write",
	"table:users",
	"Updated user record",
	success,
	ipAddress,
)
logger.LogEvent(event)
```

### 与外部系统集成

1. **用户管理集成**：
   - 实现自定义`Authenticator`接口，连接外部用户管理系统
   - 支持LDAP、OAuth2等外部认证系统

2. **权限管理集成**：
   - 实现自定义`AccessControl`接口，连接外部权限管理系统
   - 支持基于外部系统的权限验证

3. **审计日志集成**：
   - 实现自定义`AuditLogger`接口，将审计日志发送到外部系统
   - 支持ELK Stack、Splunk等日志分析系统

4. **密钥管理集成**：
   - 从外部密钥管理服务获取加密密钥
   - 支持AWS KMS、HashiCorp Vault等密钥管理服务

## 扩展系统 API

### 核心功能

#### 1. 插件接口

**API接口**：
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

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/extensions/plugins"

// 创建插件配置
config := plugins.PluginConfig{
	Enabled: true,
	Dir:     "plugins",
	Plugins: []string{"plugin1.so"},
}

// 创建插件管理器
pm := plugins.NewPluginManager(config)

// 加载插件
plugin, err := pm.Load("plugins/plugin1.so")
if err != nil {
	return
}

// 启动插件
plugin.Start()

// 停止插件
plugin.Stop()

// 卸载插件
pm.Unload(plugin.Name())
```

#### 2. 索引接口

**API接口**：
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

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/extensions/indexes"

// 创建索引注册中心
registry := indexes.NewIndexRegistry()

// 注册自定义索引工厂
registry.RegisterFactory("custom", &CustomIndexFactory{})

// 创建自定义索引
index, err := registry.CreateIndex("custom", "my_index", "field1", "field2")
if err != nil {
	return
}

// 使用索引
fieldsBytes := &map[string][]byte{
	"field1": []byte("value1"),
	"field2": []byte("value2"),
}
indexValue := index.JoinValue(fieldsBytes, 1)
```

#### 3. 存储接口

**API接口**：
```go
// Store 存储接口
type Store interface {
	// Put 存储键值对
	Put(key, value []byte) error

	// Get 获取值
	Get(key []byte) ([]byte, error)

	// Delete 删除键值对
	Delete(key []byte) error

	// Batch 获取批处理对象
	Batch() Batch

	// GetBatch 获取批处理对象
	GetBatch() Batch

	// WriteBatch 写入批处理
	WriteBatch(batch Batch) error

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

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/extensions/storage"

// 创建存储注册中心
registry := storage.NewStoreRegistry()

// 注册自定义存储工厂
registry.RegisterFactory("custom", &CustomStoreFactory{})

// 创建自定义存储
store, err := registry.CreateStore("custom", "data")
if err != nil {
	return
}

// 使用存储
store.Put([]byte("key"), []byte("value"))
value, err := store.Get([]byte("key"))
if err != nil {
	return
}

// 关闭存储
store.Close()
```

#### 4. 类型系统

**API接口**：
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

// Decode 解码值
func (r *TypeRegistry) Decode(typeName string, data []byte) (interface{}, error)
```

**使用方式**：
```go
import "github.com/liaoran123/sfsDb/extensions/types"

// 创建类型注册中心
registry := types.NewTypeRegistry()

// 注册自定义类型
registry.RegisterType("custom", &CustomType{})

// 使用自定义类型
value := "custom value"
data, err := registry.Encode("custom", value)
if err != nil {
	return
}

decoded, err := registry.Decode("custom", data)
if err != nil {
	return
}
```

### 与外部系统集成

1. **插件开发**：
   - 按照`Plugin`接口开发外部插件
   - 编译为.so文件，放入插件目录

2. **自定义存储引擎**：
   - 按照`Store`接口开发外部存储引擎
   - 通过存储注册中心注册和使用

3. **自定义数据类型**：
   - 按照`Type`接口开发外部数据类型
   - 通过类型注册中心注册和使用

4. **自定义索引**：
   - 按照`Index`接口开发外部索引类型
   - 通过索引注册中心注册和使用

## 集成最佳实践

1. **模块化集成**：
   - 每个外部系统集成都应该封装为独立的模块
   - 使用依赖注入方式集成到sfsDb

2. **配置管理**：
   - 使用统一的配置管理系统
   - 支持通过配置文件和环境变量配置集成参数

3. **错误处理**：
   - 外部系统集成的错误应该妥善处理
   - 避免外部系统故障影响sfsDb核心功能

4. **性能优化**：
   - 外部系统调用应该异步处理
   - 避免同步调用阻塞sfsDb核心操作

5. **监控和告警**：
   - 监控外部系统集成的状态
   - 为集成故障设置合理的告警规则

6. **安全考虑**：
   - 外部系统集成应该遵循最小权限原则
   - 敏感信息（如密钥、凭证）应该安全存储

7. **文档和示例**：
   - 为每个外部系统集成提供详细的文档
   - 提供完整的集成示例代码

## 版本兼容性

- API接口设计遵循向后兼容原则
- 版本升级时保持API接口的稳定性
- 为不兼容的变更提供迁移指南

## 故障排查

1. **集成故障排查**：
   - 检查外部系统的连接状态
   - 验证集成配置是否正确
   - 查看集成模块的日志

2. **性能问题排查**：
   - 检查外部系统的响应时间
   - 验证集成代码是否存在性能瓶颈
   - 使用性能分析工具定位问题

3. **安全问题排查**：
   - 检查外部系统的安全配置
   - 验证集成代码是否存在安全漏洞
   - 进行安全审计和渗透测试

## 总结

sfsDb的内部API设计为外部系统集成提供了灵活的扩展点，通过实现相应的接口，可以无缝集成各种外部系统，增强sfsDb的功能和性能。

在集成外部系统时，应该遵循最佳实践，确保集成的可靠性、性能和安全性，同时保持sfsDb核心功能的稳定性。