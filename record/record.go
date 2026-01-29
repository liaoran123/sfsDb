package record

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
)

type Record map[string]any

// 对象池，用于复用 Record 和 Records 对象
var (
	recordPool = &sync.Pool{
		New: func() any {
			return make(Record)
		},
	}

	recordsPool = &sync.Pool{
		New: func() any {
			return make(Records, 0, 10)
		},
	}
)

// GetRecord 从对象池获取一个 Record 对象
func GetRecord() Record {
	r := recordPool.Get().(Record)
	// 设置 finalizer，当对象被垃圾回收时自动放回池
	runtime.SetFinalizer(&r, func(ptr *Record) {
		if *ptr != nil {
			PutRecord(*ptr)
		}
	})
	return r
}

// PutRecord 将 Record 对象放回对象池
func PutRecord(r Record) {
	if r == nil {
		return
	}
	// 清除 finalizer，避免重复释放
	runtime.SetFinalizer(&r, nil)
	// 清空 Record 中的所有字段
	for k := range r {
		delete(r, k)
	}
	recordPool.Put(r)
}

// GetRecords 从对象池获取一个 Records 对象
func GetRecords() Records {
	rs := recordsPool.Get().(Records)
	// 设置 finalizer，当对象被垃圾回收时自动放回池
	runtime.SetFinalizer(&rs, func(ptr *Records) {
		if *ptr != nil {
			PutRecords(*ptr)
		}
	})
	return rs
}

// PutRecords 将 Records 对象放回对象池
func PutRecords(rs Records) {
	if rs == nil {
		return
	}
	// 清除 finalizer，避免重复释放
	runtime.SetFinalizer(&rs, nil)
	// 清空 Records 中的所有 Record 并将其放回对象池
	for i, r := range rs {
		if r != nil {
			PutRecord(r)
			rs[i] = nil
		}
	}
	// 重置 slice 长度
	recordsPool.Put(rs[:0])
}

// BatchSelect 批量选择多个记录的字段
// 减少多次调用 Select 方法的开销
func BatchSelect(records Records, fields ...string) Records {
	if len(records) == 0 || len(fields) == 0 {
		return records
	}

	result := GetRecords()
	result = result[:0]

	for _, r := range records {
		if r == nil {
			result = append(result, nil)
			continue
		}

		selected := GetRecord()
		for _, field := range fields {
			selected[field] = r[field]
		}
		result = append(result, selected)
	}

	return result
}

// BatchOperation 批量对多个记录执行操作
// 减少多次调用 Operation 方法的开销
func BatchOperation(records Records, op ...Operation) Records {
	if len(records) == 0 || len(op) == 0 {
		return records
	}

	result := GetRecords()
	result = result[:0]

	for _, r := range records {
		if r == nil {
			result = append(result, nil)
			continue
		}

		newRecord := GetRecord()
		for k, v := range r {
			newRecord[k] = v
		}

		newField := op[0].NewField()
		if _, exists := newRecord[newField]; exists {
			PutRecord(newRecord)
			panic(fmt.Sprintf("field '%s' already exists in record, cannot add duplicate field", newField))
		}

		newRecord[newField] = op[0].Evaluate(r)
		result = append(result, newRecord)
	}

	return result
}

// BatchOperationVertical 批量对多个记录执行垂直操作
// 减少多次调用 OperationVertical 方法的开销
func BatchOperationVertical(records Records, op ...VerticalOperation) Records {
	if len(records) == 0 || len(op) == 0 {
		return records
	}

	// 创建记录副本
	result := GetRecords()
	result = result[:0]
	result = append(result, make(Records, len(records))...)
	copy(result, records)

	// 应用每个垂直操作
	for _, verticalOp := range op {
		opResult := verticalOp.Evaluate(records)
		newField := verticalOp.NewField()

		for i := range result {
			if result[i] == nil {
				result[i] = GetRecord()
			} else {
				// 检查字段是否已存在
				if _, exists := result[i][newField]; exists {
					// 创建新记录以避免覆盖
					newRecord := GetRecord()
					for k, v := range result[i] {
						newRecord[k] = v
					}
					newRecord[newField] = opResult

					// 释放旧记录
					oldRecord := result[i]
					result[i] = newRecord
					PutRecord(oldRecord)
				} else {
					// 直接添加字段到现有记录
					result[i][newField] = opResult
				}
			}
		}
	}

	return result
}

// Select selects fields from a record
// Returns a new record with the selected fields
func (r Record) Select(keys ...string) Record {
	if r == nil {
		return nil
	}
	if len(keys) == 0 {
		return r
	}
	result := GetRecord()
	// 预分配容量
	for k := range result {
		delete(result, k)
	}

	for _, key := range keys {
		result[key] = r[key]
	}
	return result
}

type Records []Record

// Select selects fields from records
// Returns a new records with the selected fields
func (rs Records) Select(fields ...string) Records {
	if len(fields) == 0 {
		return rs
	}
	result := GetRecords()
	result = result[:0]
	result = append(result, make(Records, len(rs))...)

	for i, r := range rs {
		result[i] = r.Select(fields...)
	}
	return result
}

// Operation performs operations on records
// Returns a new records with the results of the operations
func (rs Records) Operation(op ...Operation) Records {
	if len(rs) == 0 {
		return nil
	}
	if len(op) == 0 {
		return rs
	}

	result := GetRecords()
	result = result[:0]

	for _, r := range rs {
		newRecord := GetRecord()
		// 复制所有字段
		for k, v := range r {
			newRecord[k] = v
		}

		newField := op[0].NewField()
		if _, exists := newRecord[newField]; exists {
			PutRecord(newRecord)
			panic(fmt.Sprintf("field '%s' already exists in record, cannot add duplicate field", newField))
		}

		newRecord[newField] = op[0].Evaluate(r)
		result = append(result, newRecord)
	}

	return result
}

func (rs Records) Contains(r Record) bool {
	for _, record := range rs {
		if equalRecords(record, r) {
			return true
		}
	}
	return false
}

// equalRecords 比较两个 Record 是否相等
// 比 reflect.DeepEqual 更高效，因为它只比较 map 的内容
func equalRecords(a, b Record) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if val, exists := b[k]; !exists || !equalValues(v, val) {
			return false
		}
	}
	return true
}

// equalValues 比较两个值是否相等
func equalValues(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 对于基本类型，直接比较
	switch a := a.(type) {
	case string:
		b, ok := b.(string)
		return ok && a == b
	case int:
		b, ok := b.(int)
		return ok && a == b
	case int64:
		b, ok := b.(int64)
		return ok && a == b
	case float64:
		b, ok := b.(float64)
		return ok && a == b
	case bool:
		b, ok := b.(bool)
		return ok && a == b
	case []byte:
		b, ok := b.([]byte)
		if !ok || len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	case map[string]any:
		b, ok := b.(map[string]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for k, v := range a {
			if val, exists := b[k]; !exists || !equalValues(v, val) {
				return false
			}
		}
		return true
	case []any:
		b, ok := b.([]any)
		if !ok || len(a) != len(b) {
			return false
		}
		for i := range a {
			if !equalValues(a[i], b[i]) {
				return false
			}
		}
		return true
	default:
		// 对于其他类型，使用 reflect.DeepEqual
		return reflect.DeepEqual(a, b)
	}
}

// OperationVertical performs vertical operations on records
func (rs Records) OperationVertical(op ...VerticalOperation) Records {
	if len(rs) == 0 {
		return nil
	}
	if len(op) == 0 {
		return rs
	}

	// Create a copy of records to avoid modifying the original
	result := GetRecords()
	result = result[:0]
	result = append(result, make(Records, len(rs))...)
	copy(result, rs)

	// Apply each vertical operation
	for _, verticalOp := range op {
		// Calculate the result of the vertical operation
		opResult := verticalOp.Evaluate(rs)
		newField := verticalOp.NewField()

		// Add the result to each record
		for i := range result {
			if result[i] == nil {
				result[i] = GetRecord()
			} else {
				// Check if the field already exists
				if _, exists := result[i][newField]; exists {
					// Create a new record to avoid overwriting
					newRecord := GetRecord()
					for k, v := range result[i] {
						newRecord[k] = v
					}
					newRecord[newField] = opResult
					// 释放旧的 Record
					oldRecord := result[i]
					result[i] = newRecord
					PutRecord(oldRecord)
				} else {
					// Directly add the field to the existing record
					result[i][newField] = opResult
				}
			}
		}
	}

	return result
}

func (rs Records) Intersect(other ...Records) Records {
	if len(rs) == 0 || len(other) == 0 {
		return nil
	}

	result := GetRecords()
	result = result[:0]

	for _, r := range rs {
		found := false
		for _, o := range other {
			if o.Contains(r) {
				found = true
				break
			}
		}
		if found {
			result = append(result, r)
		}
	}

	return result
}

func (rs Records) Union(other ...Records) Records {
	if len(rs) == 0 {
		if len(other) == 0 {
			return nil
		}

		result := GetRecords()
		result = result[:0]
		for _, o := range other {
			for _, r := range o {
				if !result.Contains(r) {
					result = append(result, r)
				}
			}
		}
		return result
	}

	if len(other) == 0 {
		return rs
	}

	result := GetRecords()
	result = result[:0]
	for _, r := range rs {
		result = append(result, r)
	}

	for _, o := range other {
		for _, r := range o {
			if !result.Contains(r) {
				result = append(result, r)
			}
		}
	}

	return result
}

func (rs Records) Difference(other ...Records) Records {
	if len(rs) == 0 {
		return nil
	}

	if len(other) == 0 {
		return rs
	}

	result := GetRecords()
	result = result[:0]
	for _, r := range rs {
		found := false
		for _, o := range other {
			if o.Contains(r) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, r)
		}
	}

	return result
}
