package engine

// 根据TableIter回传的key,value所有能得到的field的值，进行所需匹配
type Match interface {
	//fields *map[string]any接收迭代器传回的值进行匹配
	Match(fields *map[string]any) bool
}

type AND struct {
	//data是TableIter生成的键值map，作为交集，并集，差集等。
	data map[any]bool
	//Rule是判断，true为对应sql语句的 IN 或 AND ，false为NOT IN 或 OR。
	rule bool
}

func ANDNew(data map[any]bool, rule bool) *AND {
	return &AND{
		data: data,
		rule: rule,
	}
}
func (b *AND) Match(fields *map[string]any) bool {
	//v := b.valueParse(table, value)
	return false
}
