package engine

import (
	"bytes"
	"errors"
	"slices"
	"strings"

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
	// 获取索引前缀
	Name() string
	SetName(name string) error
	//拼接前缀
	Prefix(tbname string) []byte
	// 拼接索引前缀+索引值
	// 系统对kv数据库优化设计，将已经存在key值索引数据，从value中略去。使用时再重新分解拼接。
	// 所以，所有拼接数据组合主键或索引都需要进行转义。否则分解可能存在转义问题。
	JoinValue(fieldsBytes *map[string][]byte, tbname string, existFields ...string) []byte
	// 匹配索引字段
	MatchFields(fields ...string) bool
}

/* 都是简单函数，没有提供多态的必要性。
type AddFields func(field ...string)
type GetFields func() []string
type JoinValue func(fieldsBytes *map[string][]byte, pfx string, existFields ...string) []byte
type MatchFields func(fields ...string) bool
*/

// ------------------------------------------
// 基础索引结构体，包含所有索引类型共有的字段和方法
type BaseIndex struct {
	fields []string
	name   string
}

// 基础索引的通用方法
func (bi *BaseIndex) SetName(name string) error {
	//不能包含SPLIT
	if strings.Contains(name, SPLIT) {
		return errors.New("name can not contain SPLIT")
	}
	bi.name = name
	return nil
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

func (bi *BaseIndex) Prefix(tbname string) []byte {
	return JoinAndToBytes(tbname, bi.name)
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
func Join(fieldsBytes *map[string][]byte, fields []string, existFields ...string) []byte {
	if !exist(fields, existFields...) {
		return nil
	}
	var Value bytes.Buffer
	for _, fit := range fields {
		if v, ok := (*fieldsBytes)[fit]; ok {
			Value.Write(v)
			Value.Write([]byte(SPLIT))
		}
	}
	//删除最后一个分隔符
	Value.Truncate(Value.Len() - 1)
	return Value.Bytes()
}

func (bi *BaseIndex) JoinValue(fieldsBytes *map[string][]byte, tbname string, existFields ...string) []byte {
	if !exist(bi.fields, existFields...) {
		return nil
	}
	var Value bytes.Buffer
	bpxf := JoinAndToBytes(tbname, bi.name) //通常就是表明和索引名拼接
	if bpxf != nil {
		Value.Write(util.Bytes(bpxf))
		Value.Write([]byte(SPLIT))
	}
	Value.Write(Join(fieldsBytes, bi.fields))
	return Value.Bytes()
}

func Match(fields []string, existFields ...string) bool {
	count := 0
	for _, fit := range fields {
		//判断是否存在切片中
		if slices.Contains(existFields, fit) {
			count++
		}
	}
	return count == len(existFields) //兼容匹配，主键字段可以少于索引字段
}
func (bi *BaseIndex) MatchFields(fields ...string) bool {
	return Match(bi.fields, fields...)
}

// ------------------------------------------
// 主键接口，嵌入基础索引接口
type PrimaryKey interface {
	Index
	// 设置主键ID
	GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte
}

// ------------------------------------------
// 默认主键索引
type DefaultPrimaryKey struct {
	BaseIndex // 嵌入基础索引
}

func DefaultPrimaryKeyNew(name string) *DefaultPrimaryKey {
	return &DefaultPrimaryKey{
		BaseIndex: BaseIndex{
			name: name,
		},
	}
}

// 過濾存在的字段
// 系統設計爲過濾key存在的字段不在value中儲存。
func (dpk *DefaultPrimaryKey) GetID(fieldsBytes *map[string][]byte, existFields ...string) []byte {
	var Value bytes.Buffer
	for i, fit := range dpk.fields {
		if slices.Contains(existFields, fit) {
			continue
		}
		fieldlen := len(dpk.fields)
		if v, ok := (*fieldsBytes)[fit]; ok {
			Value.Write(v)
			if i < fieldlen-1 {
				Value.Write([]byte(SPLIT))
			}
		}
	}
	return Value.Bytes()
}

// ------------------------------------------
// 普通索引接口，嵌入基础索引接口
type NormalIndex interface {
	Index
}

// ------------------------------------------
// 默认普通索引
type DefaultNormalIndex struct {
	BaseIndex // 嵌入基础索引
}

func DefaultNormalIndexNew(name string) *DefaultNormalIndex {
	return &DefaultNormalIndex{
		BaseIndex: BaseIndex{
			name: name,
		},
	}
}

// ------------------------------------------
// 全文索引接口，嵌入基础索引接口
type FullTextIndex interface {
	Index
	SetFullField(field string, len int) error
	GetFtlen(fields ...string) int
	// 拼接全文索引值
	JoinFullValues(fieldsBytes *map[string][]byte, tbname string, existFields ...string) [][]byte
	// 分词方法
	Tokenize(nr string, ftlen int) (tokens []string)
}

// ------------------------------------------
// 全文索引字段类型
type FullTextIndexField struct {
	Fields string // 全文索引字段名
	Len    int    // 全文索引字段长度
}

// ------------------------------------------
// 默认全文索引
type DefaultFullTextIndex struct {
	BaseIndex                      // 嵌入基础索引
	ftfields  []FullTextIndexField // 全文索引字段列表
}

func DefaultFullTextIndexNew() *DefaultFullTextIndex {
	return &DefaultFullTextIndex{
		BaseIndex: BaseIndex{
			name: "ft",
		},
	}
}
func (dfi *DefaultFullTextIndex) GetFtlen(fields ...string) int {
	for _, ftfit := range dfi.ftfields {
		if slices.Contains(fields, ftfit.Fields) {
			return ftfit.Len
		}
	}
	return 5
}

// 指定那个字段是全文索引字段，以及索引长度
func (dfi *DefaultFullTextIndex) SetFullField(field string, len int) error {
	//len限制为3-11个字符
	if len < 3 || len > 11 {
		return errors.New("SetFullField len must be between 3 and 11")
	}
	dfi.ftfields = append(dfi.ftfields, FullTextIndexField{
		Fields: field,
		Len:    len,
	})
	return nil
}

// 调用基类JoinValue
func (dfi *DefaultFullTextIndex) JoinValue(fieldsBytes *map[string][]byte, tbname string, existFields ...string) []byte {
	if !exist(dfi.fields, existFields...) {
		return nil
	}
	for _, ftfit := range dfi.ftfields {
		if v, ok := (*fieldsBytes)[ftfit.Fields]; ok {
			//将转换为string中文，如果长度大于ftlen个字符，则截取前ftlen个字符
			if len([]rune(string(v))) > int(ftfit.Len) {
				(*fieldsBytes)[ftfit.Fields] = []byte(string([]rune(string(v))[:ftfit.Len]))
			}
		}
	}
	var Value bytes.Buffer
	bpxf := JoinAndToBytes(tbname, dfi.name) //通常就是表明和索引名拼接
	if bpxf != nil {
		Value.Write(util.Bytes(bpxf))
		Value.Write([]byte(SPLIT))
	}
	Value.Write(Join(fieldsBytes, dfi.fields))
	//rvalue := Value.Bytes()
	//rvalue = rvalue[:len(rvalue)-1] //全文索引是like不能与其他索引一样，后面的SPLIT要去掉
	return Value.Bytes()
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
func (dfi *DefaultFullTextIndex) JoinFullValues(fieldsBytes *map[string][]byte, tbname string, existFields ...string) [][]byte {
	if !exist(dfi.fields, existFields...) {
		return nil
	}
	//全文索引数据量大，所以使用对象池，避免频繁分配内存
	r := util.GetBytesArray() //[][]byte
	var fieldValue []byte
	var exists bool
	curpfx := util.GetBytesArray()
	ftpxf := JoinAndToBytes(tbname, dfi.name)
	var isnil bool
	for _, fit := range dfi.fields {
		// 获取字段的实际值
		fieldValue, exists = (*fieldsBytes)[fit]
		if !exists {
			continue // 跳过不存在的字段
		}
		var isFullTextField bool
		var ftlen int

		//组合全文索引时，全文索引字段和普通索引字段组合，需要判断那个字段是全文索引字段
		for _, ftfit := range dfi.ftfields {
			if fit == ftfit.Fields {
				isFullTextField = true
				ftlen = ftfit.Len
				if ftlen == 0 {
					ftlen = 5
				}
				break
			}
		}
		switch isFullTextField {
		case true: // 全文索引字段
			// 将字段值转换为字符串，全文索引只支持字符串
			fieldStr := string(fieldValue)
			// 对字段值进行分词
			tokens := dfi.Tokenize(fieldStr, ftlen)
			defer util.PutStringSlice(tokens)
			var btoken []byte
			for _, token := range tokens {
				if token == "" {
					continue
				}
				btoken = []byte(token)
				btoken = util.Bytes(btoken).Escape()
				curpfx = append(curpfx, bytes.Join([][]byte{ftpxf, append([]byte{}, btoken...)}, []byte(SPLIT)))
			}
		case false: // 普通索引字段
			fieldValue = util.Bytes(fieldValue).Escape()
			curpfx = append(curpfx, bytes.Join([][]byte{ftpxf, append([]byte{}, fieldValue...)}, []byte(SPLIT)))
		}
		//返回结果是否为空
		isnil = len(r) == 0
		switch isnil {
		case true:
			r = append(r, append([][]byte{}, curpfx...)...)
			curpfx = curpfx[:0]
		case false: // 拼接当前索引前缀到结果中,一对多和多对一关系，多对多等。多对多可能并无意义，但是系统只是支持。
			for i, idx := range r {
				for _, pfx := range curpfx {
					var tidx []byte
					tidx = append(tidx, append([]byte{}, idx...)...)
					var tpfx []byte
					tpfx = append(tpfx, append([]byte{}, pfx...)...)
					tidx = bytes.Join([][]byte{tidx, append([]byte{}, tpfx...)}, []byte{})
					r[i] = append([]byte{}, tidx...)
				}
			}
		}
		util.PutBytesArray(curpfx)
		ftpxf = ftpxf[:0]
	}
	return r
}
