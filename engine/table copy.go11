// 设计原则是当前最简单快捷开发，不考虑通用性和将来扩展要求。
// 除了需要排序的主键和索引需要转换为[]byte外，其他所有字段值，皆转换为字符串存储
package engine

import (
	"bytes"
	"fmt"
	"log"
	"maps"
	"reflect"
	"slices"
	"strings"
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
	fields map[string]any // 字段映射，string为字段名，any为字段值
	/*
		组合主键时，所有索引对应的主键值都是经过转义的。
		这是为了提取主键中的某个字段值作为匹配索引。
		所以在用索引的主键值回表记录时，需要对索引值进行反转义。
	*/
	indexs  *Indexs // 索引集合
	counter AutoInt // 自动增值计数器，使用自定义的AutoInt
	kvStore storage.Store
}

// 新建一个表
// 表名不能包含分隔符SPLIT("-")，否则返回nil
// 每次项目启动都会重新创建，或通过json转换为Table结构体
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
	tb.id = tb.getSysId("sys-tbid")
	tb.setSysId("sys-tbid", tb.id+1)
	return tb, nil
}

func (t *Table) CreateIndex(index Index) error {
	idxid := t.getSysId("sys-idxid")
	err := t.indexs.createIndex(index, idxid)
	if err != nil {
		return err
	}
	t.setSysId("sys-idxid", idxid+1)
	return nil
}
func (t *Table) GetPrimaryKey() PrimaryKey {
	pk := t.indexs.getPrimaryKey()
	if pk == nil {
		//没有主键，需要创建一个默认主键
		pk, _ = DefaultPrimaryKeyNew("id")
		pk.AddFields("id")
		t.CreateIndex(pk)
	}
	return pk
}

// 专为表和索引自动增值id设置
// 最多能创建255个表，因为tbid键值为1字节。
func (t *Table) getSysId(key string) uint8 {
	id, err := t.kvStore.Get([]byte(key))
	if err != nil {
		id = []byte{0}
	}
	r := uint8(id[0])
	return r
}
func (t *Table) setSysId(key string, id uint8) {
	t.kvStore.Put([]byte(key), []byte{id})
}

// 设置表id
func (t *Table) SetId(id uint8) {
	t.id = id
}
func (t *Table) GetId() uint8 {
	return t.id
}

// 获取自动增值的值
func (t *Table) GetAutoInc() int {
	if t.counter.Get() == 0 {
		t.InitAuto()
	}
	return int(t.counter.Increment())
}

// 初始化自动增值的值
func (t *Table) InitAuto() {
	maxValue := t.MaxAutoValue()
	// MaxAutoValue现在直接返回int64类型
	t.counter.Set(int(maxValue))
}

// 获取当前最大自动增值记录的主键值
func (t *Table) MaxAutoValue() int {
	fields := map[string]any{"id": nil} //id为nil时，全表扫描。
	tableIter := t.Search(&fields)
	defer tableIter.Release()
	if tableIter == nil {
		return 0
	}
	if !tableIter.Last() {
		return 0
	}
	key := tableIter.Key() //取最后一个key值
	rkey := key[len(t.GetPrimaryKey().Prefix(t.id))+1:]
	var target any
	// 如果主键字段为空，默认使用"id"
	if len(t.GetPrimaryKey().GetFields()) == 0 || t.GetPrimaryKey().GetFields()[0] == "" {
		target = 0
	} else {
		target = t.fields[t.GetPrimaryKey().GetFields()[0]]
	}
	r := util.Bytes(rkey).ToAny(target)
	// 使用 reflect 包进行类型转换，更灵活地处理各种数值类型
	return util.AnyToInt(r)

}

// 必须先为表预设字段和类型
func (t *Table) SetFields(fields map[string]any) error {
	//检查fields，不能包含分隔符
	for field := range fields {
		if strings.Contains(field, SPLIT) {
			return fmt.Errorf("字段名 '%s' 不能包含分隔符 '%s'", field, SPLIT)
		}
	}
	t.fields = fields
	return nil
}

// 修改字段名
func (t *Table) ReFieldName(oldName, newName string) error {
	//检查新字段名是否包含分隔符
	if strings.Contains(newName, SPLIT) {
		return fmt.Errorf("字段名 '%s' 不能包含分隔符 '%s'", newName, SPLIT)
	}
	//检查旧字段名是否存在
	if _, ok := t.fields[oldName]; !ok {
		return fmt.Errorf("字段 '%s' 不存在于表中", oldName)
	}
	//检查新字段名是否已存在
	if _, ok := t.fields[newName]; ok {
		return fmt.Errorf("字段 '%s' 已存在于表中", newName)
	}
	//修改字段名
	t.fields[newName] = t.fields[oldName]
	delete(t.fields, oldName)
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

// 检查类型是否匹配
func (t *Table) CheckType(fields *map[string]any) error {
	for field, value := range *fields {
		fieldValue, exists := t.GetField(field)
		// 只检查已经存在于表中的字段的类型
		if exists {
			// 检查类型是否匹配
			//主键是nil使用自动增值，其他字段是nil，使用默认值。
			//if field == t.primaryKey.GetFields()[0] && t.primaryKey.GetFields()[0] == "id" && value == nil {
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

// 插入记录
func (t *Table) Insert(fields *map[string]any, batchs ...storage.Batch) (currentID int, err error) {
	if t.fields == nil {
		return 0, fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	//当前自动增值的值
	currentID = -1
	//是否支持默认自动增值主键，单主键并且主键字段名为"id"
	pklen := len(t.GetPrimaryKey().GetFields())
	pkfield := t.GetPrimaryKey().GetFields()[0]
	supportDefault := pklen == 1 && pkfield == "id"
	if supportDefault {
		// 检查是否提供了主键字段
		//使用默认自动增值主键时，不需要提供主键字段，系统自动生成
		_, ok := (*fields)[pkfield]
		if !ok { //未提供主键字段，自动生成主键值
			currentID = t.GetAutoInc()
			(*fields)[pkfield] = currentID
		} else { //提供了主键字段，但是值为nil，自动生成主键值
			if (*fields)[pkfield] == nil {
				currentID = t.GetAutoInc()
				(*fields)[pkfield] = currentID
			} else { //提供了主键字段，且值不为nil，转换为int类型
				currentID = int((*fields)[pkfield].(int))
			}
		}
	}
	// 检查字段类型是否匹配
	if err = t.CheckType(fields); err != nil {
		return 0, err
	}
	// 转换字段为字节数组
	fieldsBytes := t.FieldsToBytes(fields)
	var batch storage.Batch
	//是否用户手动控制事务
	//useBatch := len(batchs) > 0
	if len(batchs) > 0 { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = t.kvStore.GetBatch()
	}
	record := t.FormatRecord(fieldsBytes)
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
	BatchContainer.SetValue(0, record)                               //添加主键value=record
	BatchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) //添加普通索引value=GetPrimaryKey().GetID()
	//添加全文索引key=joinValue,value=nil
	BatchContainer.Operation(fieldsBytes)
	//t.Operation(fieldsBytes, batch, BatchContainer)
	if len(batchs) == 0 {
		// 提交批量操作
		if err = t.kvStore.WriteBatch(batch); err != nil {
			return 0, err
		}
	}
	//fmt.Printf("Insert BatchContainer.Len(): %v\n", BatchContainer.Len())
	return currentID, nil
}

// 格式化记录
// 過濾存在的字段
func (t *Table) FormatRecord(fieldsBytes *map[string][]byte, FilterFields ...string) (r []byte) {
	var buf bytes.Buffer
	var value []byte
	//记录格式：field1:value1-field2:value2-...-fieldN:valueN-
	for field, val := range *fieldsBytes {
		if slices.Contains(FilterFields, field) {
			continue
		}
		value = util.Bytes([]byte(field)).Escape()
		buf.Write(value)
		buf.WriteString(":")
		value = util.Bytes(val).Escape()
		buf.Write(value)
		buf.WriteString(SPLIT)
	}
	//删除最后一个分隔符
	buf.Truncate(buf.Len() - 1)
	return buf.Bytes()
}

// 反序列化记录
func (t *Table) ParseRecord(record []byte) *map[string][]byte {
	if record == nil || t.fields == nil {
		return nil
	}
	bs := util.Bytes(record).Split()
	fields := make(map[string][]byte, len(bs))
	field := ""
	for _, b := range bs {
		//以第一个':'为分隔符，将值分为两部分
		before, _, ok := bytes.Cut(b, []byte{':'})
		if ok {
			field = string(before)
			fields[field] = b[len(field)+1:] //util.Bytes(b[len(field)+1:])
		}
	}
	return &fields
}

// *map[string][]byte ==> *map[string]any
// 与FieldsToBytes相反
func (t *Table) RecordByteToAny(value *map[string][]byte) *map[string]any {
	fields := make(map[string]any, len(*value))
	for field, val := range *value {
		fields[field] = util.Bytes(val).ToAny(t.fields[field])
	}
	return &fields
}

// 反序列化记录，并且将字段值转换为对应的类型
func (t *Table) ParseRecordValue(record []byte) *map[string]any {
	value := t.ParseRecord(record)
	return t.RecordByteToAny(value)
}

// 删除记录
func (t *Table) Delete(fields *map[string]any, batchs ...storage.Batch) error {
	if t.fields == nil {
		return fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	var batch storage.Batch
	//是否用户手动控制事务
	useBatch := len(batchs) > 0
	if useBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = t.kvStore.GetBatch()
	}
	//读取记录
	record, err := t.Read(fields)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}
	//反序列化记录，并且将字段值转换为对应的类型
	//fieldsBytes := t.ParseRecord(record)
	fieldsBytes := t.GetPrimaryKey().Parse(nil, record)
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
	BatchContainer.Operation(fieldsBytes)
	if len(batchs) == 0 { //用户未手动控制事务，自动提交
		t.kvStore.WriteBatch(batch)
	}
	//fmt.Printf("Delete BatchContainer.Len(): %v\n", BatchContainer.Len())
	return nil
}

// 从按主键数据库读取记录
func (t *Table) Read(fields *map[string]any) ([]byte, error) {
	//检查是否提供了所有主键字段
	for _, field := range t.GetPrimaryKey().GetFields() {
		if _, ok := (*fields)[field]; !ok {
			return nil, fmt.Errorf("删除操作必须提供主键字段 '%s'", field)
		}
	}
	fieldsBytes := t.FieldsToBytes(fields)
	key := t.GetPrimaryKey().JoinValue(fieldsBytes, t.id)
	return t.ReadByBytes(key), nil
}

// 更新记录，不支持修改主键字段
func (t *Table) Update(fields *map[string]any, batchs ...storage.Batch) error {
	if t.fields == nil {
		return fmt.Errorf("表 '%s' 未设置字段和类型", t.name)
	}
	// 检查字段类型是否匹配
	if err := t.CheckType(fields); err != nil {
		return err
	}
	var batch storage.Batch
	//是否用户手动控制事务
	useBatch := len(batchs) > 0
	if useBatch { //用户手动控制事务
		batch = batchs[0]
	} else {
		batch = t.kvStore.GetBatch()
	}
	//读取记录
	record, err := t.Read(fields)
	if err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("主键值 '%v' 的记录不存在", fields)
	}
	var updateFields []string
	for field := range *fields {
		//排除主键字段
		if t.GetPrimaryKey().MatchFields(field) {
			continue
		}
		updateFields = append(updateFields, field)
	}
	//检查更新字段个数是否为0
	if len(updateFields) == 0 {
		return nil
	}
	//反序列化记录，并且将字段值转换为对应的类型
	//fieldsBytes := t.ParseRecord(record)
	fieldsBytes := t.GetPrimaryKey().Parse(nil, record)
	BatchContainer := NewBatchContainer(batch, t.indexs, t.id, t.kvStore)
	//删除
	BatchContainer.Operation(fieldsBytes, updateFields...)

	//更新字段值
	for field, val := range *fields {
		//排除主键字段，主键字段不能更新
		if t.GetPrimaryKey().MatchFields(field) {
			continue
		}
		if _, ok := t.fields[field]; ok {
			(*fieldsBytes)[field] = util.AnyToBytes(val)
		}
	}
	//设置新值添加
	record = t.FormatRecord(fieldsBytes)
	BatchContainer.SetValue(0, record)                               //添加主键value=record
	BatchContainer.SetValue(1, t.GetPrimaryKey().GetID(fieldsBytes)) //添加普通索引value=GetPrimaryKey().GetID()
	//添加全文索引key=joinValue,value=nil
	BatchContainer.Operation(fieldsBytes, updateFields...)
	//提交事务
	if len(batchs) == 0 { //用户未手动控制事务，自动提交
		if err := t.kvStore.WriteBatch(batch); err != nil {
			return err
		}
	}
	//fmt.Printf("Update BatchContainer.Len(): %v\n", BatchContainer.Len())
	return nil
}

// 从按主键数据库读取记录
func (t *Table) ReadByBytes(key []byte) []byte {
	v, err := t.kvStore.Get(key)
	if err != nil {
		// ErrNotFound 是正常的未找到错误，不需要打印
		if err != storage.ErrNotFound {
			fmt.Printf("读取记录失败: %v\n", err)
		}
		return nil
	}
	return v
}

// 遍历表所有kv，复制表用
func (t *Table) For() storage.Iterator {
	pfx := []byte{byte(t.id), SPLIT[0]}
	rangeHelper := util.NewRangeHelper(pfx)
	slice := rangeHelper.FromComparison(util.Like, pfx)
	return t.kvStore.Iterator(slice.Start, slice.Limit)
}

// 遍历表所有数据
func (t *Table) ForData() *TableIter {
	pfx := t.GetPrimaryKey().Prefix(t.id)
	rangeHelper := util.NewRangeHelper(pfx)
	slice := rangeHelper.FromComparison(util.Like, []byte(pfx))
	return TableIterNew(t, t.kvStore.Iterator(slice.Start, slice.Limit), t.GetPrimaryKey())
}

// 匹配索引
func (t *Table) MatchIndex(fields ...string) Index {
	return t.indexs.MatchIndex(fields...)
}

// 将数据转换为字节数组，该合适搜索时用。搜索时nil值不能更改,否则导致结果错误
func (t *Table) FieldsToBytesNil(fields *map[string]any) *map[string][]byte {
	result := make(map[string][]byte, len(*fields))
	for k, v := range *fields {
		result[k] = util.AnyToBytes(v)
	}
	return &result
}

func (t *Table) Search(fields *map[string]any, ops ...util.ComparisonOperator) *TableIter {
	var field []string
	for k := range *fields {
		//判断字段是否在表中
		if _, ok := t.fields[k]; !ok {
			//写错误日志
			log.Printf("字段 '%s' 不存在于表 '%s'", k, t.name)
			return nil
		}
		field = append(field, k)
	}
	//匹配索引
	idx := t.MatchIndex(field...)
	var key []byte
	var fieldsBytes *map[string][]byte
	if idx != nil {
		fieldsBytes = t.FieldsToBytesNil(fields)
		key = idx.JoinValue(fieldsBytes, t.id)
	} else {
		/*
			该函数不支持无索引的搜索。
			如果需要支持，可以使用ForData()方法或当前函数设置主键值为nil，则得到遍历全表迭代器，然后配合mach接口自定义匹配规则。
			mach接口自定义匹配规则，理论上可以支持任意查询匹配。
		*/
		return nil
	}
	var op util.ComparisonOperator
	if len(ops) == 0 { //默认是Like操作
		op = util.Like
	} else {
		op = ops[0]
	}
	pfx := idx.Prefix(t.id)
	pfx = append(pfx, SPLIT[0])
	rangeHelper := util.NewRangeHelper(pfx)
	var iter storage.Iterator
	var tbiter *TableIter
	if op != util.NotEqual {
		slice := rangeHelper.FromComparison(op, key)
		iter = t.kvStore.Iterator(slice.Start, slice.Limit)
		tbiter = TableIterNew(t, iter, idx)
	} else { //不等于将会通过主键或索引进行全表扫描，并且设置跳跃区间
		slice := rangeHelper.FromComparison(util.Like, pfx) //遍历前缀，即通过主键或索引全表扫描
		iter = t.kvStore.Iterator(slice.Start, slice.Limit)
		tbiter = TableIterNew(t, iter, idx)
		//设置跳跃区间
		neslice := rangeHelper.FromComparison(util.Like, key) //跳跃区间key=0-1-100==>0-1-101
		tbiter.SetJumpRanges(t.kvStore.Iterator(neslice.Start, neslice.Limit))
	}
	return tbiter
}
