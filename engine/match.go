package engine

import "fmt"

// 根据TableIter回传的key,value所有能得到的field的值，进行所需匹配
type Match interface {
	//fields *map[string]any接收迭代器传回的值进行匹配
	Match(fields *map[string]any) bool
}

type AND struct {
	//需要匹配的字段名，与fields *map[string]any中的key对应
	fields []string
	//data是TableIter生成的键值map，作为存在或不存在的判断。
	//data 是由(t *TableIter) Map(fields ...string) (data map[any]bool)生成
	//也可以自定义，比如sql语句中 field in (1,2,3)，则data=map[any]bool{1:true,2:true,3:true}
	data map[any]bool
	//rule是判断，true为对应sql语句的 IN 或 AND ，false为NOT IN 或 OR。
	rule bool
}

func ANDNew(data map[any]bool, rule bool) *AND {
	return &AND{
		data: data,
		rule: rule,
	}
}

//由于data map[any]bool,any只能是一个值，所以规定，fields 大于1,则需要将fields *map[string]any中的合并值转换为字符串，再进行匹配
func (b *AND) mergeFields(fields *map[string]any) (r any) {
	if len(b.fields) > 1 {
		r = ""
		for _, f := range b.fields {
			//用分隔符SPLIT将fields中的值合并为一个字符串
			r = r.(string) + fmt.Sprintf("%v", (*fields)[f]) + SPLIT
		}
		// 去掉最后一个分隔符
		r = r.(string)[:len(r.(string))-1]
	} else {
		r = (*fields)[b.fields[0]]
	}
	return r
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
