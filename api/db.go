package api

import (
	"github.com/liaoran123/sfsDb/engine/table"
	"github.com/liaoran123/sfsDb/engine/types"
	"github.com/liaoran123/sfsDb/storage"
)

// Config 数据库配置
type Config struct {
	Path string // 数据库存储路径
}

// DB 数据库接口
type DB interface {
	// Open 打开数据库
	Open(config Config) error

	// Close 关闭数据库
	Close() error

	// CreateTable 创建表
	CreateTable(name string, schema *types.TableSchema) (Table, error)

	// DropTable 删除表
	DropTable(name string) error

	// GetTable 获取表
	GetTable(name string) (Table, error)

	// ListTables 列出所有表
	ListTables() ([]string, error)

	// Begin 开始事务
	Begin() (Tx, error)

	// Exec 执行SQL语句
	Exec(sql string, args ...any) (Result, error)

	// Query 查询SQL语句
	Query(sql string, args ...any) (Result, error)
}

// Table 数据库表接口
type Table interface {
	// GetName 获取表名
	GetName() string

	// GetSchema 获取表结构
	GetSchema() *types.TableSchema

	// Insert 插入记录
	Insert(values map[string]any) (Result, error)

	// Update 更新记录
	Update(where Expr, values map[string]any) (Result, error)

	// Delete 删除记录
	Delete(where Expr) (Result, error)

	// Select 查询记录
	Select(fields []string, where Expr) (Result, error)

	// Get 获取单条记录
	Get(primaryKey any) (map[string]any, error)
}

// Tx 事务接口
type Tx interface {
	// CreateTable 创建表
	CreateTable(name string, schema *types.TableSchema) (Table, error)

	// DropTable 删除表
	DropTable(name string) error

	// GetTable 获取表
	GetTable(name string) (Table, error)

	// Exec 执行SQL语句
	Exec(sql string, args ...any) (Result, error)

	// Query 查询SQL语句
	Query(sql string, args ...any) (Result, error)

	// Commit 提交事务
	Commit() error

	// Rollback 回滚事务
	Rollback() error
}

// Result 查询结果接口
type Result interface {
	// Next 移动到下一条记录
	Next() bool

	// Scan 将当前记录扫描到指定的变量中
	Scan(dest ...any) error

	// Map 将当前记录转换为map
	Map() (map[string]any, error)

	// Count 获取结果集大小
	Count() int64

	// Close 关闭结果集
	Close() error

	// LastInsertId 返回最后插入的ID
	LastInsertId() (int64, error)

	// RowsAffected 返回受影响的行数
	RowsAffected() (int64, error)
}

// Expr 查询表达式接口
type Expr interface {
	// SQL 返回SQL表达式
	SQL() (string, []any)
}

// DBImpl 数据库实现
type DBImpl struct {
	name    string
	config  Config
	kvStore storage.Store
	tables  map[string]Table
	isOpen  bool
}

// TableImpl 表实现
type TableImpl struct {
	name   string
	db     *DBImpl
	table  table.Table
	schema *types.TableSchema
}

// NewDB 创建新的数据库实例
func NewDB() DB {
	return &DBImpl{
		tables: make(map[string]Table),
		isOpen: false,
	}
}

// Open 打开数据库
func (db *DBImpl) Open(config Config) error {
	// 创建KV存储
	kvStore, err := storage.NewStore(storage.StoreConfig{
		Path: config.Path,
	})
	if err != nil {
		return err
	}

	db.config = config
	db.kvStore = kvStore
	db.isOpen = true

	// 加载表元数据
	err = db.loadTables()
	if err != nil {
		kvStore.Close()
		db.isOpen = false
		return err
	}

	return nil
}

// Close 关闭数据库
func (db *DBImpl) Close() error {
	if !db.isOpen {
		return nil
	}

	// 关闭所有表
	for _, table := range db.tables {
		table.(*TableImpl).table.Close()
	}

	// 关闭KV存储
	err := db.kvStore.Close()
	if err != nil {
		return err
	}

	db.isOpen = false
	return nil
}

// loadTables 加载表元数据
func (db *DBImpl) loadTables() error {
	// TODO: 实现表元数据加载
	return nil
}

// CreateTable 创建表
func (db *DBImpl) CreateTable(name string, schema *types.TableSchema) (Table, error) {
	if !db.isOpen {
		return nil, NewDBError("database is not open")
	}

	// 检查表名是否已存在
	if _, exists := db.tables[name]; exists {
		return nil, NewDBError("table already exists")
	}

	// 创建底层表
	underlyingTable, err := table.NewTable(name, schema, db.kvStore)
	if err != nil {
		return nil, err
	}

	// 创建API表
	apiTable := &TableImpl{
		name:   name,
		db:     db,
		table:  underlyingTable,
		schema: schema,
	}

	// 添加到表映射
	db.tables[name] = apiTable

	// 保存表元数据
	err = db.saveTableMetadata(name, schema)
	if err != nil {
		underlyingTable.Close()
		delete(db.tables, name)
		return nil, err
	}

	return apiTable, nil
}

// saveTableMetadata 保存表元数据
func (db *DBImpl) saveTableMetadata(name string, schema *types.TableSchema) error {
	// TODO: 实现表元数据保存
	return nil
}

// GetTable 获取表
func (db *DBImpl) GetTable(name string) (Table, error) {
	if !db.isOpen {
		return nil, NewDBError("database is not open")
	}

	table, exists := db.tables[name]
	if !exists {
		return nil, NewDBError("table not found")
	}

	return table, nil
}

// DropTable 删除表
func (db *DBImpl) DropTable(name string) error {
	if !db.isOpen {
		return NewDBError("database is not open")
	}

	table, exists := db.tables[name]
	if !exists {
		return NewDBError("table not found")
	}

	// 关闭表
	if err := table.(*TableImpl).table.Close(); err != nil {
		return err
	}

	// 删除表元数据
	if err := db.deleteTableMetadata(name); err != nil {
		return err
	}

	// 从表映射中删除
	delete(db.tables, name)

	return nil
}

// deleteTableMetadata 删除表元数据
func (db *DBImpl) deleteTableMetadata(name string) error {
	// TODO: 实现表元数据删除
	return nil
}

// ListTables 列出所有表
func (db *DBImpl) ListTables() ([]string, error) {
	if !db.isOpen {
		return nil, NewDBError("database is not open")
	}

	tables := make([]string, 0, len(db.tables))
	for name := range db.tables {
		tables = append(tables, name)
	}

	return tables, nil
}

// Begin 开始事务
func (db *DBImpl) Begin() (Tx, error) {
	if !db.isOpen {
		return nil, NewDBError("database is not open")
	}

	// TODO: 实现事务
	return nil, NewDBError("transaction not implemented yet")
}

// Exec 执行SQL语句
func (db *DBImpl) Exec(sql string, args ...interface{}) (Result, error) {
	if !db.isOpen {
		return nil, NewDBError("database is not open")
	}

	// TODO: 实现SQL执行
	return nil, NewDBError("SQL exec not implemented yet")
}

// Query 查询SQL语句
func (db *DBImpl) Query(sql string, args ...interface{}) (Result, error) {
	if !db.isOpen {
		return nil, NewDBError("database is not open")
	}

	// TODO: 实现SQL查询
	return nil, NewDBError("SQL query not implemented yet")
}

// GetName 获取表名
func (t *TableImpl) GetName() string {
	return t.name
}

// GetSchema 获取表结构
func (t *TableImpl) GetSchema() *types.TableSchema {
	return t.schema
}

// Insert 插入记录
func (t *TableImpl) Insert(values map[string]any) (Result, error) {
	// TODO: 实现插入操作
	return nil, NewDBError("insert not implemented yet")
}

// Update 更新记录
func (t *TableImpl) Update(where Expr, values map[string]any) (Result, error) {
	// TODO: 实现更新操作
	return nil, NewDBError("update not implemented yet")
}

// Delete 删除记录
func (t *TableImpl) Delete(where Expr) (Result, error) {
	// TODO: 实现删除操作
	return nil, NewDBError("delete not implemented yet")
}

// Select 查询记录
func (t *TableImpl) Select(fields []string, where Expr) (Result, error) {
	// TODO: 实现查询操作
	return nil, NewDBError("select not implemented yet")
}

// Get 获取单条记录
func (t *TableImpl) Get(primaryKey any) (map[string]any, error) {
	return t.table.Get(primaryKey)
}

// DBError 数据库错误类型
type DBError struct {
	msg string
}

// NewDBError 创建数据库错误
func NewDBError(msg string) *DBError {
	return &DBError{msg: msg}
}

// Error 返回错误消息
func (e *DBError) Error() string {
	return e.msg
}

// Open 打开数据库
func Open(config Config) (DB, error) {
	db := NewDB()
	err := db.Open(config)
	if err != nil {
		return nil, err
	}
	return db, nil
}
