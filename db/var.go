package db

//数据格式：字段id(数据)。如：1(abcde)2(12345)
const LEFT_SPLIT = "("  //左分隔符
const RIGHT_SPLIT = ")" //右分隔符
const INDEX_SPLIT = "_" //索引分隔符

var IsLitEndian = IsLittleEndian()
