package engine

import (
	"encoding/json"
	"strings"
)

// TableSchema 是Table结构体的映射，用于序列化和反序列化
// 包含Table的元数据，不包含运行时状态（如counter和kvStore）
type TableSchema struct {
	// 表ID
	ID uint8 `json:"id"`
	// 表名
	Name string `json:"name"`

	// 字段定义，string为字段名，any为字段值示例（用于类型推断）
	Fields map[string]any `json:"fields"`

	// 字段ID映射，uint8为字段ID，string为字段名
	FieldsID map[uint8]string `json:"fields_id"`

	// 索引信息，包含所有索引的基本信息
	Indexes []IndexSchema `json:"indexes"`

	// 时间字段映射，标记字段是否为时间类型
	TimeFields map[string]bool `json:"time_fields"`

	// 主键字段列表
	PrimaryFields []string `json:"primary_fields"`
}

// TableSerialization 包含Table的完整序列化信息，包括运行时状态
// 用于完整的Table序列化和反序列化
// kvStore不需要序列化，会在反序列化后重新初始化

type TableSerialization struct {
	// TableSchema 元数据
	Schema *TableSchema `json:"schema"`

	// 自动增值的当前值（counter AutoInt）
	CurrentAutoID int `json:"current_auto_id"`
}

// IndexSchema 是Index的映射，用于序列化和反序列化
// 包含索引的基本信息

type IndexSchema struct {
	// 索引名称
	Name string `json:"name"`

	// 索引类型：primary, normal, fulltext
	Type string `json:"type"`

	// 索引字段，顺序表示索引顺序
	Fields []string `json:"fields"`
}

// ToSchema 将Table转换为TableSchema
// 用于序列化Table结构体的元数据
func (t *Table) ToSchema() *TableSchema {
	// 创建Schema实例
	schema := &TableSchema{
		ID:            t.id,
		Name:          t.GetName(),
		Fields:        t.fields,
		FieldsID:      t.fieldsid,
		Indexes:       make([]IndexSchema, 0, len(t.indexs.GetAllIndexes())),
		TimeFields:    t.timeFields,
		PrimaryFields: t.GetPrimaryFields(),
	}

	// 转换索引信息
	for _, index := range t.indexs.GetAllIndexes() {
		indexSchema := IndexSchema{
			Name:   index.Name(),
			Fields: index.GetFields(),
		}

		// 确定索引类型
		switch index.(type) {
		case PrimaryKey:
			indexSchema.Type = "primary"
		case NormalIndex:
			indexSchema.Type = "normal"
		case FullTextIndex:
			indexSchema.Type = "fulltext"
		default:
			indexSchema.Type = "unknown"
		}

		// 添加到索引列表
		schema.Indexes = append(schema.Indexes, indexSchema)
	}

	return schema
}

// ToSerialization 将Table转换为TableSerialization
// 用于完整序列化Table结构体，包括运行时状态
func (t *Table) ToSerialization() *TableSerialization {
	return &TableSerialization{
		Schema:        t.ToSchema(),
		CurrentAutoID: t.counter.Get(),
	}
}

// FromSchema 从TableSchema创建Table
// 用于从元数据反序列化Table结构体
func FromSchema(schema *TableSchema) (*Table, error) {
	// 创建Table实例
	table, err := TableNew(schema.Name)
	if err != nil {
		return nil, err
	}

	// 设置表ID
	table.id = schema.ID

	// 设置字段
	if err := table.SetFields(schema.Fields); err != nil {
		return nil, err
	}

	// 恢复字段ID映射
	if schema.FieldsID != nil {
		table.fieldsid = schema.FieldsID
	}

	// 恢复时间字段映射
	if schema.TimeFields != nil {
		table.timeFields = schema.TimeFields
	}

	// 恢复主键字段列表
	if schema.PrimaryFields != nil {
		table.primaryFields = schema.PrimaryFields
		table.primaryFieldsLoaded = true
	}

	// 直接操作indexs的indexs切片，避免调用GetPrimaryKey()导致自动创建主键索引
	// 先清空indexs切片
	table.indexs.indexs = make([]Index, 0, len(schema.Indexes))

	// 恢复索引
	for _, indexSchema := range schema.Indexes {
		var index Index
		var err error

		// 根据索引类型创建索引
		switch strings.ToLower(indexSchema.Type) {
		case "primary":
			// 主键索引
			index, err = DefaultPrimaryKeyNew(indexSchema.Name)
		case "normal":
			// 普通索引
			index, err = DefaultNormalIndexNew(indexSchema.Name)
		case "fulltext":
			// 全文索引
			index, err = DefaultFullTextIndexNew(indexSchema.Name)
		default:
			// 未知索引类型，跳过
			continue
		}

		if err != nil {
			return nil, err
		}

		// 添加索引字段
		index.AddFields(indexSchema.Fields...)

		// 直接添加到indexs切片，避免CreateIndex的检查（已经在ToSchema时验证过了）
		table.indexs.indexs = append(table.indexs.indexs, index)
	}

	// 检查是否存在主键索引
	hasPrimaryKey := false
	for _, idx := range table.indexs.indexs {
		if _, ok := idx.(PrimaryKey); ok {
			hasPrimaryKey = true
			break
		}
	}

	// 如果没有主键索引，创建默认主键索引
	if !hasPrimaryKey {
		primaryKey, err := DefaultPrimaryKeyNew("pk")
		if err != nil {
			return nil, err
		}
		primaryKey.AddFields("id")
		table.indexs.indexs = append(table.indexs.indexs, primaryKey)
	}

	return table, nil
}

// FromSerialization 从TableSerialization创建Table
// 用于从完整序列化信息反序列化Table结构体
// kvStore需要在调用此函数后重新初始化
func FromSerialization(serialization *TableSerialization) (*Table, error) {
	// 从Schema创建Table
	table, err := FromSchema(serialization.Schema)
	if err != nil {
		return nil, err
	}

	// 恢复自动增值
	table.counter.Set(serialization.CurrentAutoID)

	// kvStore不需要序列化，会在调用后重新初始化

	return table, nil
}

// MarshalJSON 将TableSchema序列化为JSON
func (ts *TableSchema) MarshalJSON() ([]byte, error) {
	// 使用类型别名避免递归调用
	type Alias TableSchema
	return json.MarshalIndent((*Alias)(ts), "", "  ")
}

// UnmarshalJSON 从JSON反序列化TableSchema
func (ts *TableSchema) UnmarshalJSON(data []byte) error {
	type Alias TableSchema
	return json.Unmarshal(data, (*Alias)(ts))
}

// MarshalJSON 将TableSerialization序列化为JSON
func (ts *TableSerialization) MarshalJSON() ([]byte, error) {
	// 使用类型别名避免递归调用
	type Alias TableSerialization
	return json.MarshalIndent((*Alias)(ts), "", "  ")
}

// UnmarshalJSON 从JSON反序列化TableSerialization
func (ts *TableSerialization) UnmarshalJSON(data []byte) error {
	type Alias TableSerialization
	return json.Unmarshal(data, (*Alias)(ts))
}

// TableToJSON 将Table转换为JSON字符串
func TableToJSON(t *Table) (string, error) {
	serialization := t.ToSerialization()
	bytes, err := json.MarshalIndent(serialization, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// TableFromJSON 从JSON字符串创建Table
func TableFromJSON(jsonStr string) (*Table, error) {
	var serialization TableSerialization
	if err := json.Unmarshal([]byte(jsonStr), &serialization); err != nil {
		return nil, err
	}
	return FromSerialization(&serialization)
}
