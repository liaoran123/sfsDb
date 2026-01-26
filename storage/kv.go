package storage

// Error definitions
var (
	ErrNotFound     = NewError("key not found")
	ErrInvalidKey   = NewError("invalid key")
	ErrInvalidValue = NewError("invalid value")
	ErrStoreClosed  = NewError("store is closed")
)

// Error KV存储错误类型
type Error struct {
	msg string
}

// NewError 创建新的KV存储错误
func NewError(msg string) *Error {
	return &Error{msg: msg}
}

// Error 返回错误消息
func (e *Error) Error() string {
	return e.msg
}

// Store KV存储接口
type Store interface {
	// Get 获取指定key的值
	Get(key []byte) ([]byte, error)

	// Put 设置key-value对
	Put(key []byte, value []byte) error

	// Delete 删除指定key
	Delete(key []byte) error

	// Batch 创建批量操作对象
	GetBatch() Batch

	// WriteBatch 执行批量操作
	WriteBatch(batch Batch, put ...bool) error

	// Iterator 创建迭代器
	//Iterator(slice *util.Range) Iterator
	Iterator(start, limit []byte) Iterator

	// Snapshot 创建快照
	Snapshot() (Snapshot, error)
	// SwitchToSnapshot 切换到数据库模式
	SwitchToSnapshot() error
	// SwitchToDB 切换到数据库模式
	SwitchToDB() error
	// Close 关闭存储
	Close() error
}

// Batch 批量操作接口
type Batch interface {
	// Put 添加put操作
	Put(key []byte, value []byte)

	// Delete 添加delete操作
	Delete(key []byte)

	// Len 获取批量操作数量
	Len() int

	// Reset 重置批量操作
	Reset()
}

// Iterator 迭代器接口
type Iterator interface {

	// First 移动到第一个元素
	First() bool

	// Last 移动到最后一个元素
	Last() bool

	// Seek 移动到大于等于指定key的位置
	Seek(key []byte) bool

	// Next 移动到下一个元素
	Next() bool

	// Prev 移动到前一个元素
	Prev() bool

	// Key 获取当前元素的key
	Key() []byte

	// Value 获取当前元素的value
	Value() []byte

	// Valid 检查迭代器是否有效
	Valid() bool

	// Release 释放迭代器资源
	Release()
}

// Snapshot 快照接口
type Snapshot interface {
	// Get 从快照中获取指定key的值
	Get(key []byte) ([]byte, error)

	// Iterator 从快照中创建迭代器
	Iterator(start, limit []byte) Iterator

	// Release 释放快照资源
	Release() error
}
