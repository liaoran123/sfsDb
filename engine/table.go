// 设计原则是当前最简单快捷开发，不考虑通用性和将来扩展要求。
// 除了需要排序的主键和索引需要转换为[]byte外，其他所有字段值，皆转换为字符串存储
package engine

import (
	"fmt"
	"maps"
	"reflect"
	"time"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

const SPLIT = util.SPLIT //分隔符
/*
表结构，半结构泛型化支持
提供灵活的组合索引优化查询
*/

type Table struct {
	id   uint8  // 表id
	name string // 表名
	/*
		半结构，可以随意增加字段
		泛型化，字段可以存储任意类型，不限制某种类型。至于有什么作用，看自己发挥。
		底层支持泛型，业务上则由自己定义规则。
		默认固定一个id字段为自动增值，当值为nil时，使用counter自动增值。
	*/
	fields         map[string]any   // 字段映射，string为字段名，any为字段值
	fieldsid       map[uint8]string // id到字段名的映射
	indexs         *Indexs          // 索引集合
	counter        AutoInt          // 自动增值计数器，使用自定义的AutoInt
	kvStore        storage.Store
	fieldIDManager *IDManager // 字段ID管理器
	indexIDManager *IDManager // 索引ID管理器
	TableCache
}

// 缓存结构体
type TableCache struct {
	timeFields          map[string]bool // 标记字段是否为时间类型
	primaryFields       []string        // 缓存的主键字段列表
	primaryFieldsLoaded bool            // 主键字段列表是否已加载
}

// 创建或获取一个表
func TableNew(name string) (*Table, error) {
	// 获取 DBManager 实例
	dbMgr := storage.GetDBManager()
	// 检查数据库是否已初始化
	if dbMgr.GetDB() == nil {
		_, err := dbMgr.OpenDB("./kvdb")
		if err != nil {
			return nil, err
		}
	}
	tb := &Table{
		name:     name,
		fields:   make(map[string]any),
		fieldsid: make(map[uint8]string),
		kvStore:  dbMgr.GetDB(),
	}
	tb.timeFields = make(map[string]bool)
	tb.indexs = NewIndexs(&tb.fields)
	if TableIDManager == nil {
		TableIDManager = NewIDManager(tb.kvStore)
	}
	key := TableIDManager.GenerateTableKey(name)
	id, _, err := TableIDManager.GetOrCreateID(key)
	if err != nil {
		return nil, err
	}
	tb.id = id    //tb.getSysNameId(name, "tb") //tb.getSysId("sys-tbid")
	tb.InitAuto() //初始化自动增值计数器。支持单一ID生成模式，要么系统自动增长，要么用户自定义ID。
	return tb, nil
}

func (t *Table) GetId() uint8 {
	return t.id
}

// GetPrimaryFields 获取主键字段列表（带缓存）
func (t *Table) GetPrimaryFields() []string {
	if !t.primaryFieldsLoaded {
		t.primaryFields = t.GetPrimaryKey().GetFields()
		t.primaryFieldsLoaded = true
	}
	return t.primaryFields
}

// ResetPrimaryFields 重置主键字段列表缓存
func (t *Table) ResetPrimaryFields() {
	t.primaryFieldsLoaded = false
	t.primaryFields = nil
}

// 必须先为表预设字段和类型
func (t *Table) SetFields(fields map[string]any) error {
	// 创建字段的副本，避免修改原始映射
	fieldsCopy := make(map[string]any, len(fields))
	for k, v := range fields {
		fieldsCopy[k] = v
	}

	// 更新表字段
	t.fields = fieldsCopy
	t.fieldsid = make(map[uint8]string, len(t.fields))
	if t.fieldIDManager == nil {
		t.fieldIDManager = NewIDManager(t.kvStore)
	}
	var fkey string
	for field := range t.fields {
		fkey = t.fieldIDManager.GenerateFieldKey(t.id, field)
		id, _, err := t.fieldIDManager.GetOrCreateID(fkey)
		if err != nil {
			return err
		}
		t.fieldsid[id] = field
	}

	// 更新时间字段映射
	t.initTimeFields()
	t.ResetPrimaryFields()
	return nil
}

// 修改字段名称
func (t *Table) UpdateFieldName(oldfield string, newfield string) error {
	if _, ok := t.fields[oldfield]; !ok {
		return fmt.Errorf("字段 '%s' 不存在于表中", oldfield)
	}
	if t.fieldIDManager == nil {
		t.fieldIDManager = NewIDManager(t.kvStore)
	}
	//1，修改IDManager中的字段名
	fkey := t.fieldIDManager.GenerateFieldKey(t.id, oldfield)
	t.fieldIDManager.UpdateKey(fkey, newfield)
	//2，修改fields映射中的字段名
	if value, exists := t.fields[oldfield]; exists {
		delete(t.fields, oldfield)
		t.fields[newfield] = value
	}
	//3，修改fieldsid映射中的字段名
	for id, name := range t.fieldsid {
		if name == oldfield {
			t.fieldsid[id] = newfield
			break
		}
	}
	//4，修改索引中的字段名
	t.indexs.UpdateFields(oldfield, newfield)
	t.ResetPrimaryFields()
	return nil
}

// 获取自动增值的值
func (t *Table) AutoValue() int {
	if t.counter.Get() == 0 {
		t.counter.Set(t.MaxAutoValue())
	}
	return t.counter.Increment()
}

// GetName 获取表名
func (t *Table) GetName() string {
	return t.name
}

// GetStore 获取存储实例
func (t *Table) GetStore() storage.Store {
	return t.kvStore
}

// GetPrimary 获取主键字段名
func (t *Table) GetPrimary() []string {
	primaryFields := make([]string, len(t.GetPrimaryFields()))
	for _, field := range t.GetPrimaryFields() {
		primaryFields = append(primaryFields, field)
	}
	return primaryFields
}

// 获取单个字段值
func (t *Table) GetField(field string) (any, bool) {
	if t.fields == nil {
		return nil, false
	}
	value, exists := t.fields[field]
	return value, exists
}

// 获取所有字段值，用于添加记录时，直接复制，无需自行创建。
//
//go:inline
func (t *Table) GetAllFields() map[string]any {
	// 返回字段的副本，避免直接修改内部状态
	result := make(map[string]any, len(t.fields))
	maps.Copy(result, t.fields)
	return result
}

// 获取所有字段名称和id映射
func (t *Table) GetAllFieldNameIdMap() map[string]uint8 {
	fieldNameIdMap := make(map[string]uint8, len(t.fieldsid))
	for id, field := range t.fieldsid {
		fieldNameIdMap[field] = uint8(id)
	}
	return fieldNameIdMap
}

// 获取所有字段名
func (t *Table) GetFieldsName() []string {
	result := make([]string, 0, len(t.fields))
	for field := range t.fields {
		result = append(result, field)
	}
	return result
}

// 检查类型是否匹配
//
//go:inline
func (t *Table) CheckType(fields *map[string]any) error {
	for field, value := range *fields {
		fieldValue, exists := t.fields[field]
		// 只检查已经存在于表中的字段的类型
		if exists {
			// 检查类型是否匹配
			//主键是nil使用自动增值，其他字段是nil，使用默认值。
			if value == nil {
				continue
			}
			if reflect.TypeOf(fieldValue) != reflect.TypeOf(value) {
				return fmt.Errorf("字段 '%s' 的类型 '%T' 与提供的值 '%T' 类型不匹配", field, fieldValue, value)
			}
		} else {
			return fmt.Errorf("字段 '%s' 不存在于表中", field)
		}
	}
	return nil
}

// 将数据转换为字节数组，该合适添加时用。搜索时nil值不能更改
// *map[string]any ==> *map[string][]byte
// 与RecordByteToAny相反
//
//go:inline
func (t *Table) FieldsToBytes(fields *map[string]any) *map[string][]byte {
	result := GlobalFieldsBytesPool.Get()
	for k, v := range *fields {
		//value为nil时，使用默认值
		//如果是时间类型，则使用当前时间
		if v == nil {
			if t.isTimeField(k) {
				v = time.Now()
			} else {
				v = t.fields[k]
			}
		}
		result[k] = util.AnyToBytes(v)
	}
	return &result
}

// isTimeField 检查字段是否为时间类型
func (t *Table) isTimeField(field string) bool {
	if t.timeFields == nil {
		t.initTimeFields()
	}
	return t.timeFields[field]
}

// initTimeFields 初始化时间字段映射
func (t *Table) initTimeFields() {
	t.timeFields = make(map[string]bool)
	for field, value := range t.fields {
		t.timeFields[field] = reflect.TypeOf(value) == reflect.TypeFor[time.Time]()
	}
}

// BatchFieldsToBytes 批量转换多个记录
//
//go:inline
func (t *Table) BatchFieldsToBytes(records []*map[string]any) []*map[string][]byte {
	results := make([]*map[string][]byte, len(records))
	for i, fields := range records {
		results[i] = t.FieldsToBytes(fields)
	}
	return results
}

// 格式化记录
// fieldsid  map[uint8]string // id到字段名的映射的关键作用在这里。第一个字节是字段id，后面是字段值。解析方法简单。
// 对应 func (dpk *DefaultPrimaryKey) Parse(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error)
// fieldsBytes必须是与t.fields相同的字段
// 编译器优化 ： //go:inline 提示编译器进行内联优化，减少函数调用开销

//go:inline
func (t *Table) FormatRecord(fieldsBytes *map[string][]byte) []byte {
	/*
		对于 FormatRecord 这种高频调用的方法，减少遍历次数和代码复杂度的收益，通常大于精确计算缓冲区大小所带来的内存节省。因此，在这个特定场景中，保守估计是更好的选择。
		当然，在某些特殊场景下（例如处理非常大的记录，或对内存使用有严格要求的环境），精确计算可能更合适。但对于大多数常规使用场景，保守估计的方案更加平衡和实用。
	*/
	// 估算缓冲区大小，减少扩容次数
	// 每个字段至少需要 1 字节的 ID + 1 字节的分隔符
	estimatedSize := len(t.fieldsid) * 16 // 保守估计每个字段平均大小
	data := make([]byte, 0, estimatedSize)
	//记录格式：field1idvalue1-field2idvalue2-...-fieldNidvalueN
	// 按照t.fields中的字段顺序来格式化记录，确保顺序一致
	for id, field := range t.fieldsid {
		if val, ok := (*fieldsBytes)[field]; ok {
			// 转义值但不修改原始数据
			escapedVal := util.Bytes(val).Escape()
			// 直接使用 append 操作字节切片
			data = append(data, byte(id))
			data = append(data, escapedVal...)
			data = append(data, SPLIT...)
		}
	}
	//删除最后一个分隔符
	if len(data) > 0 {
		data = data[:len(data)-len(SPLIT)]
	}
	return data
}

// BatchFormatRecords 批量格式化多个记录
// 对于批量处理场景，此方法比多次调用 FormatRecord 更高效
//
//go:inline
func (t *Table) BatchFormatRecords(records []*map[string][]byte) [][]byte {
	// 预分配结果切片，减少扩容次数 , 比多次调用高效
	results := make([][]byte, len(records))

	// 批量处理所有记录
	for i, fieldsBytes := range records {
		results[i] = t.FormatRecord(fieldsBytes)
	}

	return results
}

// *map[string][]byte ==> *map[string]any
// 与FieldsToBytes相反
//
//go:inline
func (t *Table) RecordByteToAny(value *map[string][]byte) *map[string]any {
	// 检查参数
	if value == nil || t.fields == nil {
		return nil
	}
	fields := GetAnyMap()
	// 直接使用从对象池获取的 map，Go 会自动处理 map 的扩容
	for field, val := range *value {
		fields[field] = util.Bytes(val).ToAny(t.fields[field])
	}
	return &fields
}

// BatchRecordByteToAny 批量转换多个记录
//
//go:inline
func (t *Table) BatchRecordByteToAny(records []*map[string][]byte) []*map[string]any {
	results := make([]*map[string]any, len(records))
	for i, value := range records {
		results[i] = t.RecordByteToAny(value)
	}
	return results
}

// 获取所有索引的名称和id映射
func (t *Table) GetAllIndexNameIdMap() map[string]uint8 {
	return t.indexs.GetAllIndexNameIdMap()
}

// parseParams 解析操作的参数
// 解析可变参数，返回batch和timeout。这2个参考都可能不需要提供。

/*
	func (t *Table) parseParams(params ...any) (storage.Batch, time.Duration) {
		var batch storage.Batch
		var timeout time.Duration

		// 处理可变参数
		for _, param := range params {
			switch v := param.(type) {
			case storage.Batch:
				batch = v
			case time.Duration:
				timeout = v
			}
		}

		return batch, timeout
	}

// prepareBatch 准备批量操作的batch
// 如果不是手动事务从外部传入batch，则使用创建一个batch，如果是手动事务，则使用外部传入的batch。

	func (t *Table) prepareBatch1(batch storage.Batch) (storage.Batch, bool, error) {
		userProvidedBatch := batch != nil

		// 如果没有提供batch，使用默认batch
		if batch == nil {
			batch = t.kvStore.GetBatch()
			if batch == nil {
				return nil, false, fmt.Errorf("failed to get batch")
			}
		}

		return batch, userProvidedBatch, nil
	}

// prepareBatch 准备批量操作的 batch
// prepareBatch 准备批量操作的batch
// 如果不是手动事务从外部传入batch，则使用创建一个batch，如果是手动事务，则使用外部传入的batch。

	func (t *Table) prepareBatch1(batchs ...storage.Batch) (storage.Batch, bool, error) {
		var batch storage.Batch
		userProvidedBatch := len(batchs) > 0

		if userProvidedBatch {
			batch = batchs[0]
			if batch == nil {
				return nil, false, fmt.Errorf("batch cannot be nil")
			}
		} else {
			batch = t.kvStore.GetBatch()
			if batch == nil {
				return nil, false, fmt.Errorf("failed to get batch")
			}
		}

		return batch, userProvidedBatch, nil
	}
*/
func (t *Table) prepareBatch(batchs ...storage.Batch) (storage.Batch, bool) {
	var batch storage.Batch
	//是否用户手动控制事务
	userProvidedBatch := len(batchs) > 0
	if userProvidedBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = t.kvStore.GetBatch()
	}
	return batch, userProvidedBatch
}

/*
// commitTransaction 提交事务

	func (t *Table) commitTransaction(batch storage.Batch, userProvidedBatch bool) error {
		//提交事务
		if !userProvidedBatch { //如果不是手动事务，则提交默认batch。否则由外部提交。
			if err := t.kvStore.WriteBatch(batch); err != nil {
				return err
			}
		}
		return nil
	}

// OpenTable 根据表名打开已存在的表并加载其结构信息
// 参数:
//   name: 表名
// 返回:
//   error: 错误信息
func (t *Table) OpenTable(name string) error {
	// 获取 DBManager 实例
	dbMgr := storage.GetDBManager()
	// 检查数据库是否已初始化
	if dbMgr.GetDB() == nil {
		_, err := dbMgr.OpenDB("./kvdb")
		if err != nil {
			return err
		}
	}

	// 设置表名
	t.name = name

	// 初始化字段映射
	t.fields = make(map[string]any)
	t.fieldsid = make(map[uint8]string)

	// 设置存储实例
	t.kvStore = dbMgr.GetDB()

	// 初始化索引集合
	t.indexs = NewIndexs(&t.fields)

	// 初始化表ID管理器
	if TableIDManager == nil {
		TableIDManager = NewIDManager(t.kvStore)
	}

	// 获取表ID
	key := TableIDManager.GenerateTableKey(name)
	id, _, err := TableIDManager.GetOrCreateID(key)
	if err != nil {
		return err
	}
	t.id = id
	t.InitAuto() //初始化自动增值计数器。支持单一ID生成模式，要么系统自动增长，要么用户自定义ID。
	// 尝试加载字段ID映射
	// 这里可以通过系统管理器获取字段信息
	// 或者从存储中读取字段ID映射

	// 加载时间字段映射
	t.timeFields = make(map[string]bool)

	// 重置主键字段缓存
	t.primaryFields = []string{}
	t.primaryFieldsLoaded = false

	return nil
}
*/
// OpenTable 创建并返回一个新的表实例
// 参数:
//
//	name: 表名
//
// 返回:
//
//	*Table: 创建的表实例
//	error: 错误信息
