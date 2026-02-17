package engine

import (
	"bytes"
	"errors"
	"slices"

	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/util"
)

// ------------------------------------------

// 基础索引接口，定义所有索引类型共有的方法
type Index interface {
	// 添加索引字段
	AddFields(field ...string)
	// 获取索引字段列表
	GetFields() []string
	Len() int
	setId(id uint8)
	GetId() uint8
	Name() string
	SetName(name string) error
	//修改索引字段名称
	UpdateFields(oldfields string, newfields string)
	//删除索引字段
	DeleteFields(field ...string)
	//拼接前缀，表id+索引id
	Prefix(tbid uint8) []byte
	//拼接值，字段值拼接
	Join(fieldsBytes *map[string][]byte) []byte
	//拼接前缀+值。调用Prefix方法
	JoinPrefix(tbid uint8, val []byte) []byte
	// 拼接索引前缀+索引值，调用JoinPrefix，Join方法
	JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte
	//是否唯一key，key=JoinValue方法返回的值
	IsUnique(key []byte) bool
	// 匹配索引字段
	MatchFields(fields ...string) bool
	// 解析kv的value值，返回字段值map
	// 主键索引解析出记录。其他索引解析出主键值。因为支持组合主键，需要解析。
	//Parse(fields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)

}

// ------------------------------------------
// 基础索引结构体，包含所有索引类型共有的字段和方法
type BaseIndex struct {
	fields []string
	id     uint8
	name   string
}

// 基础索引的通用方法
func (bi *BaseIndex) SetName(name string) error {
	if name == "" {
		return errors.New("name can not be empty")
	}
	bi.name = name
	return nil
}

func (bi *BaseIndex) setId(id uint8) {
	bi.id = id
}

func (bi *BaseIndex) GetId() uint8 {
	return bi.id
}

func (bi *BaseIndex) Name() string {
	return bi.name
}

func (bi *BaseIndex) Len() int {
	return len(bi.fields)
}

func (bi *BaseIndex) GetFields() []string {
	return bi.fields
}

func (bi *BaseIndex) AddFields(field ...string) {
	bi.fields = append(bi.fields, field...)
}

// 修改索引字段名称
func (bi *BaseIndex) UpdateFields(oldfields string, newfields string) {
	// 替换字段名
	for i := range bi.fields {
		if bi.fields[i] == oldfields {
			bi.fields[i] = newfields
		}
	}
}

// 删除索引字段
func (bi *BaseIndex) DeleteFields(field ...string) {
	for i := range bi.fields {
		if bi.fields[i] == field[0] {
			bi.fields = append(bi.fields[:i], bi.fields[i+1:]...)
			break
		}
	}
}

func JoinAndToBytes(v ...string) []byte {
	var Value bytes.Buffer
	for i, fit := range v {
		if fit == "" {
			continue
		}
		Value.Write([]byte(fit))
		if i < len(v)-1 {
			Value.Write([]byte(SPLIT))
		}
	}
	return Value.Bytes()
}

// 拼接前缀，表id+索引id
func (bi *BaseIndex) Prefix(tbid uint8) []byte {
	//fmt.Printf(" %d %s%d\n", tbid, SPLIT, bi.id)
	return []byte{byte(tbid), SPLIT[0], byte(bi.id)}
}

// 存在某个字段
func exist(fields []string, existFields ...string) bool {
	if len(existFields) == 0 {
		return true
	}
	for _, fit := range existFields {
		//判断是否存在切片中
		if slices.Contains(fields, fit) {
			return true
		}
	}
	return false
}

// 系统对kv数据库优化设计，将已经存在key值索引数据，从value中略去。使用时再重新分解拼接。
// 所有拼接数据都要进行转义
// 添加匹配字段功能
func Join(fieldsBytes *map[string][]byte, fields []string) []byte {
	var Value bytes.Buffer
	for _, fit := range fields {
		if v, ok := (*fieldsBytes)[fit]; ok {
			if v == nil {
				continue
			}
			Value.Write(v)
			Value.Write([]byte(SPLIT))
		}
	}
	//删除最后一个分隔符
	if Value.Len() > 0 {
		Value.Truncate(Value.Len() - 1)
	}
	return Value.Bytes()
}

func (bi *BaseIndex) JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte {
	if !exist(bi.fields, existFields...) {
		return nil
	}
	val := bi.Join(fieldsBytes)
	return bi.JoinPrefix(tbid, val)
}

// IsUnique方法返回false，表示不是唯一索引。
// 传入参考key=JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte
func (bi *BaseIndex) IsUnique(key []byte) bool {
	_, err := storage.GetDBManager().GetDB().Get(key)
	if err != nil {
		return true
	}
	return false
}

func (bi *BaseIndex) JoinPrefix(tbid uint8, val []byte) []byte {
	var Value bytes.Buffer
	//表id和索引id拼接
	bpxf := bi.Prefix(tbid) //[]byte{byte(tbid), SPLIT[0], byte(bi.id)}
	Value.Write(util.Bytes(bpxf))
	Value.Write([]byte(SPLIT))
	Value.Write(val)
	return Value.Bytes()
}
func (bi *BaseIndex) Join(fieldsBytes *map[string][]byte) []byte {
	return Join(fieldsBytes, bi.fields)
}

// 前缀规则匹配索引字段
func MatchFields(fields []string, existFields ...string) bool {
	count := 0
	for _, fit := range fields {
		//判断是否存在索引切片中
		if slices.Contains(existFields, fit) {
			count++
		} else { //依照索引前缀匹配规则，有一个不匹配，则后面的都不会匹配。
			break
		}
	}
	//如索引是3个字段组合，需要匹配的是2个字段，如何2个字段都按索引字段先后顺序匹配（即前缀匹配），则成功。
	return count == len(existFields)
}

func (bi *BaseIndex) MatchFields(fields ...string) bool {
	return MatchFields(bi.fields, fields...)
}

// 将索引的value值转换为主键map值
// value=fval1+SPLIT+fval2+SPLIT+...+fieldn
// fields []string, value []byte, 顺序必须相同
// pkfieldTypeLen，按照定长截取字段值。
func (bi *BaseIndex) Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error) {
	pflen := 0
	pklen := len(primaryFields)
	for _, fit := range primaryFields {
		if len, ok := (*pkfieldTypeLen)[fit]; ok {
			if len == 0 && pklen > 1 { // 组合主键时，主键类型的长度未指定，无法解析主键值。
				//返回错误
				return nil, errors.New("使用可变长度类型作为组合主键，需要注册指定长度。")
			}
			pflen += int(len) + 1 //每个字段之间用分隔符隔开
		}
	}
	pflen -= 1 //最后一个分隔符不需要计算
	val := value[len(value)-pflen:]
	fieldsBytes := GlobalFieldsBytesPool.Get()
	if pklen == 1 { // 单主键时，直接返回值。 这段代码是必须的，在单主键情况下，不需要进行复杂的解析。而且单主键支持非固定长度类型，也可以正常解析。
		fieldsBytes[primaryFields[0]] = val
		return &fieldsBytes, nil
	}

	var fieldTypeLen uint8
	pos := 0
	// value=fval1+SPLIT+fval2+SPLIT+...+fieldn
	//获取每个字段的值，fval1,fval2,...,fieldn
	for _, fit := range primaryFields {
		fieldTypeLen = (*pkfieldTypeLen)[fit]
		fieldsBytes[fit] = val[pos : pos+int(fieldTypeLen)]
		pos += int(fieldTypeLen) + 1 // 跳过分隔符，下一个字段的起始位置
	}

	return &fieldsBytes, nil
}

// ------------------------------------------
// 主键接口，嵌入基础索引接口
type PrimaryKey interface {
	Index
	// 设置主键ID
	GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte
	GetfieldTypeLen(tablefields *map[string]any) *map[string]uint8
	Parse(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error)
}

// ------------------------------------------
// 默认主键索引
// 组合主键时只支持固定长度的类型的组合。
// 字符串类型，必须指定长度。否则无法解析或存在转义问题导致bug。
type DefaultPrimaryKey struct {
	BaseIndex // 嵌入基础索引
}

func DefaultPrimaryKeyNew(name string) (*DefaultPrimaryKey, error) {
	dpk := &DefaultPrimaryKey{
		BaseIndex: BaseIndex{
			name: name,
		},
	}
	if err := dpk.SetName(name); err != nil {
		return nil, err
	}
	return dpk, nil
}

// 获取主键字段的总长度
// 组合主键时只支持固定长度的类型的组合。
// 字符串类型，必须指定长度。否则无法解析或存在转义问题导致bug。
func (dpk *DefaultPrimaryKey) GetfieldTypeLen(tablefields *map[string]any) *map[string]uint8 {
	r := make(map[string]uint8)
	flen := 0
	for _, fit := range dpk.fields {
		if val, ok := (*tablefields)[fit]; ok {
			//固定长度类型，系统字段提取，字符串类型需要注册指定长度util.RegisterTypeSize。
			flen = util.TypeSize(val)
			r[fit] = uint8(flen)
		}
	}
	return &r
}

// 過濾存在的字段
// 系統設計爲過濾key存在的字段不在value中儲存。
func (dpk *DefaultPrimaryKey) GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte {
	var Value bytes.Buffer
	fieldlen := len(dpk.fields)
	for i, fit := range dpk.fields {
		if slices.Contains(existFields, fit) {
			continue
		}
		if v, ok := (*fieldsBytes)[fit]; ok {
			Value.Write(v)
			if i < fieldlen-1 {
				Value.Write([]byte(SPLIT))
			}
		}
	}
	return Value.Bytes()
}

/*
//
// 与func (t *Table) FormatRecord(fieldsBytes *map[string][]byte) (r []byte) 相对应
// 主键索引值记录格式：field1idvalue1-field2idvalue2-...-fieldNidvalueN
// Parse(fields []string, value []byte)
// fieldsid  map[uint8]string // id到字段名的映射的关键作用在这里。第一个字节是字段id，后面是字段值。解析方法简单。
*/
// 反格式化，解析函数，将索引的value转换为记录对应的map[string][]byte
func (dpk *DefaultPrimaryKey) Parse(fieldsid map[uint8]string, value []byte) (*map[string][]byte, error) {
	if fieldsid == nil {
		return nil, errors.New("DefaultPrimaryKey Parse error: fieldsid is nil")
	}
	vals := util.Bytes(value).Split()
	fieldsBytes := GlobalFieldsBytesPool.Get()
	var key string
	for _, val := range vals {
		key = fieldsid[val[0]]
		fieldsBytes[key] = val[1:]
	}
	return &fieldsBytes, nil
}

// ------------------------------------------
// 普通索引接口，嵌入基础索引接口
type NormalIndex interface {
	Index
	// 将索引的value转换为主键map值
	// 由于NormalIndex完全匹配index接口，所以需要一个Tag方法来区别是否是二级索引。
	Tag() bool
	Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}

// ------------------------------------------
// 默认普通索引，二级索引
type DefaultNormalIndex struct {
	BaseIndex // 嵌入基础索引
}

func DefaultNormalIndexNew(name string) (*DefaultNormalIndex, error) {
	dni := &DefaultNormalIndex{
		BaseIndex: BaseIndex{
			name: name,
		},
	}
	if err := dni.SetName(name); err != nil {
		return nil, err
	}
	return dni, nil
}

// Tag方法返回true，表示是二级索引。
func (dni *DefaultNormalIndex) Tag() bool {
	return true
}

// ------------------------------------------
// 全文索引接口，嵌入基础索引接口
type FullTextIndex interface {
	Index
	SetFullField(field string, len int) error
	//GetFtlen() int
	// 拼接全文索引值
	JoinFullValues(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) [][]byte
	// 分词方法
	Tokenize(nr string, ftlen int) (tokens []string)
	Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error)
}

// ------------------------------------------
// 默认全文索引
type DefaultFullTextIndex struct {
	BaseIndex // 嵌入基础索引
	//全文索引切分字段
	ftsplit string
	//切分长度
	ftlen int
	//ftfields  []FullTextIndexField // 全文索引字段列表
}

func DefaultFullTextIndexNew(name string) (*DefaultFullTextIndex, error) {
	dfi := &DefaultFullTextIndex{
		BaseIndex: BaseIndex{
			name: name,
		},
	}
	if err := dfi.SetName(name); err != nil {
		return nil, err
	}
	return dfi, nil
}
func (dfi *DefaultFullTextIndex) Parse(primaryFields []string, pkfieldTypeLen *map[string]uint8, value []byte) (*map[string][]byte, error) {
	//检测primaryFields在全文索引的前面还是后面
	idx := slices.Index(dfi.fields, primaryFields[0])
	//全文索引必须在前面或后面全量加上主键字段（单或多主键），否则全文索引不成立。
	if idx == -1 {
		//返回错误
		return nil, errors.New("DefaultFullTextIndex Parse error: primaryFields not in fields\n全文索引必须在前面或后面全量加上主键字段（单或多主键），否则全文索引不成立。")
	}
	pflen := 0
	pklen := len(primaryFields)
	for _, fit := range primaryFields {
		if len, ok := (*pkfieldTypeLen)[fit]; ok {
			if len == 0 && pklen > 1 { // 组合主键时，主键类型的长度未指定，无法解析主键值。
				//中断程序
				return nil, errors.New("使用可变长度类型作为组合主键，需要注册指定长度。")
			}
			pflen += int(len) + 1 //每个字段之间用分隔符隔开
		}
	}
	pflen -= 1 //最后一个分隔符不需要计算
	var val []byte
	//如果primaryFields在全文索引的前面。
	if idx == 0 { //需要判断前面或后面，取值不同，其他与基类逻辑一样。
		val = value[len(dfi.Prefix(0)):pflen]
	} else { //如果primaryFields在全文索引的后面，与基类方法一样。
		val = value[len(value)-pflen:]
	}
	fieldsBytes := GlobalFieldsBytesPool.Get()
	if pklen == 1 { // 单主键时，直接返回值。 这段代码是必须的，在单主键情况下，不需要进行复杂的解析。而且单主键支持非固定长度类型，也可以正常解析。
		fieldsBytes[primaryFields[0]] = val
		return &fieldsBytes, nil
	}

	var fieldTypeLen uint8
	pos := 0
	// value=fval1+SPLIT+fval2+SPLIT+...+fieldn
	//获取每个字段的值，fval1,fval2,...,fieldn
	for _, fit := range primaryFields {
		fieldTypeLen = (*pkfieldTypeLen)[fit]
		fieldsBytes[fit] = val[pos : pos+int(fieldTypeLen)]
		pos += int(fieldTypeLen) + 1 // 跳过分隔符，下一个字段的起始位置
	}

	return &fieldsBytes, nil

}

// 指定那个字段是全文索引字段，以及索引长度
func (dfi *DefaultFullTextIndex) SetFullField(field string, len int) error {
	//len限制为3-11个字符
	if len < 3 || len > 11 {
		return errors.New("SetFullField len must be between 3 and 11")
	}
	//判断field是否在fields中
	if !exist(dfi.fields, field) {
		return errors.New("SetFullField field not in fields")
	}
	dfi.ftlen = len
	dfi.ftsplit = field
	return nil
}

// 调用基类JoinValue
// 对应全文索引，该函数仅作搜索前缀用。
// 拼接全文索引前缀,所有全文索引值都会进行转义。
func (dfi *DefaultFullTextIndex) JoinValue(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) []byte {
	if v, ok := (*fieldsBytes)[dfi.ftsplit]; ok {
		//将转换为string中文，如果长度大于ftlen个字符，则截取前ftlen个字符
		if len([]rune(string(v))) > int(dfi.ftlen) {
			tval := []byte(string([]rune(string(v))[:dfi.ftlen]))
			(*fieldsBytes)[dfi.ftsplit] = tval
		}
	}
	return dfi.BaseIndex.JoinValue(fieldsBytes, tbid, existFields...)
}
func (dfi *DefaultFullTextIndex) Tokenize(nr string, ftlen int) (tokens []string) {
	tokens = util.GetStringSlice()
	var knr string //, fid
	var ml, cl int
	var r, idxstr []rune
	r = []rune(nr)
	cl = len([]rune(nr))
	for cl > 0 {
		ml = min(cl, ftlen)
		idxstr = r[:ml]
		knr = string(idxstr)
		tokens = append(tokens, knr)
		r = r[1:]
		cl = len(r)
	}
	return
}

// put时拼接key的值
// 全文索引考据级别的切词算法。
// 只有全文索引才需要转义
func (dfi *DefaultFullTextIndex) JoinFullValues(fieldsBytes *map[string][]byte, tbid uint8, existFields ...string) [][]byte {
	if !exist(dfi.fields, existFields...) {
		return nil
	}
	fieldlen := len(dfi.fields)
	indexfieldsBytes := make([][][]byte, fieldlen)
	loop := 0
	/*
		####拼接策略####
		//将所有需要拼接的值，全部储存进入indexfieldsBytes([][][]byte)，再循环拼接。
		// 普通索引的结果值是[]byte，全文索引的结果值是[][]byte，兼容载装则需要[][][]byte
	*/
	var val []byte
	var ok bool
	//将所有需要拼接的值，全部储存进入indexfieldsBytes([][][]byte)
	for _, f := range dfi.fields {
		if val, ok = (*fieldsBytes)[f]; !ok || val == nil {
			continue // 跳过不存在的字段
		}
		if f == dfi.ftsplit { //全文索引切分字段
			tokens := dfi.Tokenize(string(val), dfi.ftlen)
			defer util.PutStringSlice(tokens)
			for _, token := range tokens {
				if token == "" {
					continue
				}
				indexfieldsBytes[loop] = append(indexfieldsBytes[loop], []byte(token))
			}
		} else {
			indexfieldsBytes[loop] = [][]byte{val}
		}
		loop++
	}
	ftpxf := dfi.Prefix(tbid) //[]byte{byte(tbid), SPLIT[0], byte(dfi.id), SPLIT[0]}
	ftpxf = append(ftpxf, []byte(SPLIT)...)
	//拼接indexfieldsBytes成为全文索引值
	//return CombineBytes(indexfieldsBytes, []byte(SPLIT), ftpxf)
	if len(indexfieldsBytes) == 0 {
		return [][]byte{}
	}
	// 转义分隔符
	escapedSep := []byte(SPLIT)
	// 从第一个数组开始，复制元素以避免修改原始数据
	result := util.GetBytesArray() //全文索引数据量大，所以使用对象池，避免频繁分配内存

	// 初始化第一个数组，转义所有元素
	for _, b := range indexfieldsBytes[0] {
		//b = util.Bytes(b).Escape()
		b = append(ftpxf, b...)
		result = append(result, append([]byte{}, b...)) // 复制数组，避免共享底层数组
	}
	// 全文索引算法，多个字节数组，使用指定分隔符返回第一个数组与其他数组的所有可能拼接组合
	// 处理剩余数组，生成所有可能的拼接组合
	for _, array := range indexfieldsBytes[1:] {
		newResult := util.GetBytesArray() // 使用对象池获取新的切片
		for _, existing := range result {
			// existing 已经是转义过的结果，不需要再次转义
			for _, nextBytes := range array {
				//nextBytes = util.Bytes(nextBytes).Escape() // 转义新的数组元素
				// 创建新的组合：existing + escapedSep + nextBytes
				combined := make([]byte, 0, len(existing)+len(escapedSep)+len(nextBytes))
				combined = append(combined, existing...)
				combined = append(combined, escapedSep...)
				combined = append(combined, nextBytes...)
				newResult = append(newResult, combined)
			}
		}
		// 旧的result将被覆盖，外部调用者不会使用，所以可以安全释放
		util.PutBytesArray(result) // 归还旧的result到对象池
		result = newResult
	}
	// 直接返回result，外部调用者使用后需要归还到对象池
	return result
}

// 全文索引算法
// CombineBytes 接收多个字节数组，使用指定分隔符返回第一个数组与其他数组的所有可能拼接组合
// 返回的结果是从对象池获取的，外部调用者使用后需要通过 PutBytesArray 归还到对象池
func CombineBytes(arrays [][][]byte, sep []byte, prefix []byte) [][]byte {
	if len(arrays) == 0 {
		return [][]byte{}
	}
	// 转义分隔符
	escapedSep := sep
	// 从第一个数组开始，复制元素以避免修改原始数据
	result := util.GetBytesArray() //全文索引数据量大，所以使用对象池，避免频繁分配内存

	// 初始化第一个数组，转义所有元素
	for _, b := range arrays[0] {
		//b = util.Bytes(b).Escape()
		b = append(prefix, b...)
		result = append(result, append([]byte{}, b...)) // 复制数组，避免共享底层数组
	}

	// 处理剩余数组，生成所有可能的拼接组合
	for _, array := range arrays[1:] {
		newResult := util.GetBytesArray() // 使用对象池获取新的切片
		for _, existing := range result {
			// existing 已经是转义过的结果，不需要再次转义
			for _, nextBytes := range array {
				//nextBytes = util.Bytes(nextBytes).Escape() // 转义新的数组元素
				// 创建新的组合：existing + escapedSep + nextBytes
				combined := make([]byte, 0, len(existing)+len(escapedSep)+len(nextBytes))
				combined = append(combined, existing...)
				combined = append(combined, escapedSep...)
				combined = append(combined, nextBytes...)
				newResult = append(newResult, combined)
			}
		}
		// 旧的result将被覆盖，外部调用者不会使用，所以可以安全释放
		util.PutBytesArray(result) // 归还旧的result到对象池
		result = newResult
	}
	// 直接返回result，外部调用者使用后需要归还到对象池
	return result
}
