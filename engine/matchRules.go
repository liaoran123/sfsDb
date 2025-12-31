package engine

// 根据TableIter回传的key,value所有能得到的field的值，进行所需匹配
type MatchKeyValue interface {
	Match(KVs *map[string]any) bool
}

type MatchRule struct {
	// 需要获取匹配的字段名，TableIter会根据这些字段名获取对应的值回传
	// 如果Fields为空，默认是要匹配主键值，单个主键有效。
	//Fields []string
	//valueParse ValueParseFun
	//根据TableIter回传的值，判断与Data是否匹配
	//Data是有其他TableIter生成的索引结构集，作为交集，并集，差集等。
	Data map[any]bool
	//Rule是判断，true为IN / AND，false为NOT IN。

}

func MatchRuleNew(data map[any]bool) *MatchRule {
	return &MatchRule{
		Data: data,
		//Rule: true,
	}
}
func (b *MatchRule) Match(value []byte) bool {
	//v := b.valueParse(table, value)
	return false
}
