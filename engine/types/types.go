package types

// DataType 数据类型枚举
type DataType int

const (
	TypeInvalid DataType = iota
	TypeBool
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeFloat32
	TypeFloat64
	TypeString
	TypeBytes
	TypeDate
	TypeDateTime
	TypeTime
	TypeUUID
	TypeJSON
)

// TypeName 返回数据类型的名称
func (t DataType) TypeName() string {
	switch t {
	case TypeInvalid:
		return "invalid"
	case TypeBool:
		return "bool"
	case TypeInt8:
		return "int8"
	case TypeInt16:
		return "int16"
	case TypeInt32:
		return "int32"
	case TypeInt64:
		return "int64"
	case TypeUint8:
		return "uint8"
	case TypeUint16:
		return "uint16"
	case TypeUint32:
		return "uint32"
	case TypeUint64:
		return "uint64"
	case TypeFloat32:
		return "float32"
	case TypeFloat64:
		return "float64"
	case TypeString:
		return "string"
	case TypeBytes:
		return "bytes"
	case TypeDate:
		return "date"
	case TypeDateTime:
		return "datetime"
	case TypeTime:
		return "time"
	case TypeUUID:
		return "uuid"
	case TypeJSON:
		return "json"
	default:
		return "unknown"
	}
}

// IsNumeric 检查是否为数值类型
func (t DataType) IsNumeric() bool {
	return t >= TypeInt8 && t <= TypeFloat64
}

// IsInteger 检查是否为整数类型
func (t DataType) IsInteger() bool {
	return t >= TypeInt8 && t <= TypeUint64
}

// IsFloat 检查是否为浮点数类型
func (t DataType) IsFloat() bool {
	return t == TypeFloat32 || t == TypeFloat64
}

// IsString 检查是否为字符串类型
func (t DataType) IsString() bool {
	return t == TypeString
}

// IsBytes 检查是否为字节类型
func (t DataType) IsBytes() bool {
	return t == TypeBytes
}

// IsDate 检查是否为日期类型
func (t DataType) IsDate() bool {
	return t == TypeDate || t == TypeDateTime || t == TypeTime
}

// Field 字段定义
type Field struct {
	Name     string   // 字段名
	Type     DataType // 数据类型
	Nullable bool     // 是否允许为空
	Default  any      // 默认值
	Comment  string   // 字段注释
}

// IndexType 索引类型枚举
type IndexType int

const (
	IndexTypeInvalid  IndexType = iota
	IndexTypeBTree              // B树索引
	IndexTypeHash               // 哈希索引
	IndexTypeFullText           // 全文索引
	IndexTypeBitmap             // 位图索引
)

// IndexName 返回索引类型的名称
func (t IndexType) IndexName() string {
	switch t {
	case IndexTypeInvalid:
		return "invalid"
	case IndexTypeBTree:
		return "btree"
	case IndexTypeHash:
		return "hash"
	case IndexTypeFullText:
		return "fulltext"
	case IndexTypeBitmap:
		return "bitmap"
	default:
		return "unknown"
	}
}

// IndexDef 索引定义
type IndexDef struct {
	Name    string            // 索引名
	Type    IndexType         // 索引类型
	Fields  []string          // 索引字段
	Options map[string]string // 索引选项
}

// TableSchema 表结构定义
type TableSchema struct {
	Name       string            // 表名
	Fields     []*Field          // 字段列表
	PrimaryKey []string          // 主键字段
	Indexes    []*IndexDef       // 索引定义
	Options    map[string]string // 表选项
}
