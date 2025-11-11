package db

/* Go 语言的接口提供了一种实现多态的方式。不同的结构体可以实现相同的接口，从而实现类似继承的行为。
多态用接口，继承用包含父类
关键词索引或全文索引
*/
type Index interface {
	Setkey()
}
