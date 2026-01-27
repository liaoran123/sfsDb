// 设计原则是当前最简单快捷开发，不考虑通用性和将来扩展要求。
// 除了需要排序的主键和索引需要转换为[]byte外，其他所有字段值，皆转换为字符串存储
package engine

import (
	"bytes"
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
}

// 创建或获取一个表
func TableNew(name string) (*Table, error) {
	if storage.KVDb == nil {
		_, err := storage.OpenDefaultDb("./kvdb")
		if err != nil {
			return nil, err
		}
	}
	tb := &Table{
		name:    name,
		fields:  make(map[string]any),
		kvStore: storage.KVDb,
	}
	tb.indexs = NewIndexs(&tb.fields)
	if TableIDManager == nil {
		TableIDManager = NewIDManager(tb.kvStore)
	}
	key := TableIDManager.GenerateTableKey(name)
	id, _, err := TableIDManager.GetOrCreateID(key)
	if err != nil {
		return nil, err
	}
	tb.id = id //tb.getSysNameId(name, "tb") //tb.getSysId("sys-tbid")
	return tb, nil
}

func (t *Table) GetId() uint8 {
	return t.id
}

// 必须先为表预设字段和类型
// 由于进行乐观锁的设计，版本号字段默认是v，占据一个字段，故而只支持254个字段。默认版本号值为0，每次更新时自动增加1
func (t *Table) SetFields(fields map[string]any) error {
	//字段命不能是版本号名称v
	if t.fields["v"] != nil {
		return fmt.Errorf("字段名称 'v' 是默认字段，作为版本号，不能自定义。")
	}

	t.fields = fields
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
	//添加版本号字段，并设置id为255
	t.fields["v"] = 0
	t.fieldsid[uint8(255)] = "v"
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
	//2，修改索引中的字段名
	t.indexs.UpdateFields(oldfield, newfield)
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

// GetPrimary 获取主键字段名
func (t *Table) GetPrimary() []string {
	primaryFields := make([]string, len(t.GetPrimaryKey().GetFields()))
	for _, field := range t.GetPrimaryKey().GetFields() {
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
func (t *Table) GetAllFields() map[string]any {
	// 返回字段的副本，避免直接修改内部状态
	result := make(map[string]any)
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
func (t *Table) CheckType(fields *map[string]any) error {
	for field, value := range *fields {
		fieldValue, exists := t.GetField(field)
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
func (t *Table) FieldsToBytes(fields *map[string]any) *map[string][]byte {
	result := make(map[string][]byte, len(*fields))
	for k, v := range *fields {
		//value为nil时，使用默认值
		//如果是时间类型，则使用当前时间
		if v == nil {
			if reflect.TypeOf(t.fields[k]) == reflect.TypeFor[time.Time]() {
				v = time.Now()
			} else {
				v = t.fields[k]
			}
		}
		result[k] = util.AnyToBytes(v)
	}
	return &result
}

// 格式化记录
// fieldsid  map[uint8]string // id到字段名的映射的关键作用在这里。第一个字节是字段id，后面是字段值。解析方法简单。
// 对应 func (dpk *DefaultPrimaryKey) Parse(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error)
// fieldsBytes必须是与t.fields相同的字段
func (t *Table) FormatRecord(fieldsBytes *map[string][]byte) (r []byte) {
	var buf bytes.Buffer
	//记录格式：field1idvalue1-field2idvalue2-...-fieldNidvalueN
	// 按照t.fields中的字段顺序来格式化记录，确保顺序一致
	for id, field := range t.fieldsid {
		if val, ok := (*fieldsBytes)[field]; ok {
			buf.WriteByte(byte(id))
			buf.Write(util.Bytes(val).Escape())
			buf.WriteString(SPLIT)
		}
	}
	//删除最后一个分隔符
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	return buf.Bytes()
}

// *map[string][]byte ==> *map[string]any
// 与FieldsToBytes相反
func (t *Table) RecordByteToAny(value *map[string][]byte) *map[string]any {
	// 检查参数
	if value == nil || t.fields == nil {
		return nil
	}
	fields := make(map[string]any, len(*value))
	for field, val := range *value {
		fields[field] = util.Bytes(val).ToAny(t.fields[field])
	}
	return &fields
}

// 获取所有索引的名称和id映射
func (t *Table) GetAllIndexNameIdMap() map[string]uint8 {
	return t.indexs.GetAllIndexNameIdMap()
}
