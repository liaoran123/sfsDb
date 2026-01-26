package record

// RecordsOperation defines a function type for records operations
// that take multiple Records arguments and return a single Records result
// 记录操作函数类型，接受多个 Records 参数并返回单个 Records 结果
type RecordsOperation func(rs Records, other ...Records) Records

// Apply applies a dynamic operation to the records with other records
// 动态应用操作函数到记录集合
func (rs Records) Apply(op RecordsOperation, other ...Records) Records {
	return op(rs, other...)
}
