package util

import "fmt"

const (
	SPLIT = "-"
)

// 系统保留键名
var ReservedKeys = []string{
	"sys",        //系统表
	"table",      //表目录
	"file",       //文件表
	"systableid", //系统表的最大当前id
}

func MergeFields(fns []string, fields *map[string]any) (r any) {
	if len(fns) > 1 {
		r = ""
		for _, f := range fns {
			//用分隔符SPLIT将fields中的值合并为一个字符串
			r = r.(string) + fmt.Sprintf("%v", (*fields)[f]) + SPLIT
		}
		// 去掉最后一个分隔符
		r = r.(string)[:len(r.(string))-1]
	} else {
		r = (*fields)[fns[0]]
	}
	return r
}
