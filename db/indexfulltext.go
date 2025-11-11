package db

/* 全文索引
继承 Index interface
*/
type IndexFullText struct {
	//len uint8 //默认为5。
}

func (f *IndexFullText) Setkey() {

}
