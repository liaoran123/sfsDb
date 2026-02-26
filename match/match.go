package match

import (
	"github.com/liaoran123/sfsDb/util"
)

// 根据TableIter回传的key,value所有能得到的field的值，进行所需匹配
type Match interface {
	//fields *map[string]any接收迭代器传回的值进行匹配
	Match(fields *map[string]any) bool
}

/*
//该结构作用是多个迭代器的主键值进行相同或不相同的匹配
//rule为true时，多个迭代器的主键值必须相同，才匹配成功
//rule为false时，多个迭代器的主键值必须不同，才匹配成功
//rule为false的作用主要是用在跳跃查询中，比如sql语句中 field not in (1,2,3)，则data=map[any]bool{1:true,2:true,3:true}
*/
type AND struct {
	//需要匹配的字段名，与fields *map[string]any中的key对应
	fields []string
	//data 是由(t *TableIter) Map(fields ...string) (data map[any]bool)生成
	//也可以自定义，比如sql语句中 field in (1,2,3)，则data=map[any]bool{1:true,2:true,3:true}
	data map[any]bool
	//rule是判断，true为对应sql语句的 IN 或 AND ，false为NOT IN 或 OR。
	rule bool
}

func NewAND(fields []string, data map[any]bool, rule ...bool) *AND {
	if len(rule) == 0 {
		rule = append(rule, true)
	}
	return &AND{
		fields: fields,
		data:   data,
		rule:   rule[0],
	}
}

// 由于data map[any]bool,any只能是一个值，所以规定，fields 大于1,则需要将fields *map[string]any中的合并值转换为字符串，再进行匹配
func (b *AND) mergeFields(fields *map[string]any) (r any) {
	return util.MergeFields(b.fields, fields)
}
func (b *AND) Match(fields *map[string]any) bool {
	if fields == nil || len(*fields) == 0 {
		return false == b.rule
	}
	value := b.mergeFields(fields)
	_, ok := b.data[value]
	//rule是判断，true为对应sql语句的 IN 或 AND ，false为NOT IN 或 OR。
	return ok == b.rule
}

/*
//复杂字段匹配使用 match.FieldComparison struct 实现的接口
//支持的比较操作符
// Equal 等于 (=)
	Equal ComparisonOperator = iota
	// NotEqual 不等于 (!=)
	NotEqual
	// GreaterThan 大于 (>)
	GreaterThan
	// GreaterThanOrEqual 大于等于 (>=)
	GreaterThanOrEqual
	// LessThan 小于 (<)
	LessThan
	// LessThanOrEqual 小于等于 (<=)
	LessThanOrEqual
	// Like 类似于SQL LIKE操作
	Like
	// Prefix 前缀匹配
	Prefix
	// Suffix 后缀匹配
	Suffix
	// Contains 包含匹配
	Contains
*/
