package record

import (
	"runtime"
	"sync"
	"sync/atomic"
)

// 对象池，用于复用 Record 和 Records 对象
var (
	recordPool = &sync.Pool{
		New: func() any {
			atomic.AddInt64(&recordCreated, 1)
			atomic.AddInt64(&globalRecordCreated, 1)
			// 预分配一定容量，减少后续扩容开销
			return make(Record, 8)
		},
	}

	recordsPool = &sync.Pool{
		New: func() any {
			atomic.AddInt64(&recordsCreated, 1)
			atomic.AddInt64(&globalRecordsCreated, 1)
			// 预分配合理容量，减少后续扩容开销
			return make(Records, 0, 16)
		},
	}

	// 对象池跟踪计数器
	recordCreated    int64 // 已创建的 Record 对象数
	recordGet        int64 // 从池获取的 Record 对象数
	recordPutManual  int64 // 手动放回池的 Record 对象数
	recordsCreated   int64 // 已创建的 Records 对象数
	recordsGet       int64 // 从池获取的 Records 对象数
	recordsPutManual int64 // 手动放回池的 Records 对象数

	// 全局跟踪计数器
	globalRecordGet      int64 // 全局从池获取的 Record 对象数
	globalRecordPut      int64 // 全局放回池的 Record 对象数
	globalRecordsGet     int64 // 全局从池获取的 Records 对象数
	globalRecordsPut     int64 // 全局放回池的 Records 对象数
	globalRecordCreated  int64 // 全局创建的 Record 对象数
	globalRecordsCreated int64 // 全局创建的 Records 对象数

)

// GetRecord 从对象池获取一个 Record 对象
func GetRecord() Record {
	r := recordPool.Get().(Record)
	atomic.AddInt64(&recordGet, 1)
	atomic.AddInt64(&globalRecordGet, 1)
	// 创建一个临时变量，用于设置finalizer
	// 由于Go函数参数是按值传递的，我们需要确保finalizer设置在正确的对象上
	temp := r
	runtime.SetFinalizer(&temp, func(ptr *Record) {
		if *ptr != nil {
			obj := *ptr
			// 清空 Record 中的所有字段
			for k := range obj {
				delete(obj, k)
			}
			// 将对象放回池
			recordPool.Put(obj)
			// 注意：不再增加全局放回计数，因为会导致重复计数
			// 手动调用 PutRecord 时已经增加了计数
			// 清空指针，避免重复处理
			*ptr = nil
		}
	})
	return r
}

// putRecordInternal 将 Record 对象放回对象池（内部使用，不增加统计计数）
func putRecordInternal(r Record) {
	if r == nil {
		return
	}
	// 直接清空 Record 中的所有字段
	for k := range r {
		delete(r, k)
	}
	// 将对象放回池
	recordPool.Put(r)
	// 注意：内部使用，不增加统计计数
}

// PutRecord 将 Record 对象放回对象池
func PutRecord(r Record) {
	if r == nil {
		return
	}
	// 直接清空 Record 中的所有字段
	for k := range r {
		delete(r, k)
	}
	// 将对象放回池
	recordPool.Put(r)
	atomic.AddInt64(&recordPutManual, 1)
	atomic.AddInt64(&globalRecordPut, 1)
	// 注意：由于 Go 的函数参数是按值传递的，我们无法在这里清除原始对象的 finalizer
	// finalizer 会在对象不再被引用时触发，但由于我们已经将对象放回池，
	// 当对象再次被获取时，会被重新设置 finalizer，这不会导致问题
}

// GetRecords 从对象池获取一个 Records 对象
func GetRecords() Records {
	rs := recordsPool.Get().(Records)
	atomic.AddInt64(&recordsGet, 1)
	atomic.AddInt64(&globalRecordsGet, 1)

	// 清空 Records 中的所有 Record 并将其放回对象池
	// 确保从池中获取的对象不包含任何未释放的 Record
	for i, r := range rs {
		if r != nil {
			putRecordInternal(r)
			rs[i] = nil
		}
	}
	// 重置长度
	rs = rs[:0]

	// 创建一个临时变量，用于设置finalizer
	temp := rs
	runtime.SetFinalizer(&temp, func(ptr *Records) {
		if *ptr != nil {
			obj := *ptr
			// 清空 Records 中的所有 Record 并将其放回对象池
			for i, r := range obj {
				if r != nil {
					putRecordInternal(r)
					obj[i] = nil
				}
			}
			// 重置 slice 长度并放回池
			recordsPool.Put(obj[:0])
			// 注意：不再增加全局放回计数，因为会导致重复计数
			// 手动调用 PutRecords 时已经增加了计数
			// 清空指针，避免重复处理
			*ptr = nil
		}
	})
	return rs
}

// GetRecordsWithCapacity 从对象池获取一个指定初始容量的 Records 对象
func GetRecordsWithCapacity(capacity int) Records {
	rs := recordsPool.Get().(Records)
	atomic.AddInt64(&recordsGet, 1)
	atomic.AddInt64(&globalRecordsGet, 1)

	// 清空 Records 中的所有 Record 并将其放回对象池
	// 确保从池中获取的对象不包含任何未释放的 Record
	for i, r := range rs {
		if r != nil {
			putRecordInternal(r)
			rs[i] = nil
		}
	}

	// 调整容量以匹配指定大小
	if cap(rs) < capacity {
		// 如果当前容量不足，创建一个新的 slice 并将旧的放回池
		newRs := make(Records, 0, capacity)
		recordsPool.Put(rs[:0])
		atomic.AddInt64(&recordsPutManual, 1)
		atomic.AddInt64(&globalRecordsPut, 1)
		rs = newRs
		atomic.AddInt64(&recordsCreated, 1)
		atomic.AddInt64(&globalRecordsCreated, 1)
	} else {
		// 重置长度
		rs = rs[:0]
	}

	// 创建一个临时变量，用于设置finalizer
	temp := rs
	runtime.SetFinalizer(&temp, func(ptr *Records) {
		if *ptr != nil {
			obj := *ptr
			// 清空 Records 中的所有 Record 并将其放回对象池
			for i, r := range obj {
				if r != nil {
					putRecordInternal(r)
					obj[i] = nil
				}
			}
			// 重置 slice 长度并放回池
			recordsPool.Put(obj[:0])
			// 注意：不再增加全局放回计数，因为会导致重复计数
			// 手动调用 PutRecords 时已经增加了计数
			// 清空指针，避免重复处理
			*ptr = nil
		}
	})
	return rs
}

// PutRecords 将 Records 对象放回对象池
func PutRecords(rs Records) {
	if rs == nil {
		return
	}
	// 清空 Records 中的所有 Record 并将其放回对象池
	for i, r := range rs {
		if r != nil {
			putRecordInternal(r)
			rs[i] = nil
		}
	}
	// 重置 slice 长度并放回池
	recordsPool.Put(rs[:0])
	atomic.AddInt64(&recordsPutManual, 1)
	atomic.AddInt64(&globalRecordsPut, 1)
	// 注意：由于 Go 的函数参数是按值传递的，我们无法在这里清除原始对象的 finalizer
	// finalizer 会在对象不再被引用时触发，但由于我们已经将对象放回池，
	// 当对象再次被获取时，会被重新设置 finalizer，这不会导致问题
}

// PoolStats 对象池使用统计信息
func PoolStats() map[string]int64 {
	recordPutManual := atomic.LoadInt64(&recordPutManual)
	recordsPutManual := atomic.LoadInt64(&recordsPutManual)

	return map[string]int64{
		"recordCreated":    atomic.LoadInt64(&recordCreated),
		"recordGet":        atomic.LoadInt64(&recordGet),
		"recordPutManual":  recordPutManual,
		"recordsCreated":   atomic.LoadInt64(&recordsCreated),
		"recordsGet":       atomic.LoadInt64(&recordsGet),
		"recordsPutManual": recordsPutManual,
	}
}

// GlobalPoolStats 全局对象池使用统计信息
func GlobalPoolStats() map[string]int64 {
	return map[string]int64{
		"globalRecordCreated":  atomic.LoadInt64(&globalRecordCreated),
		"globalRecordGet":      atomic.LoadInt64(&globalRecordGet),
		"globalRecordPut":      atomic.LoadInt64(&globalRecordPut),
		"globalRecordsCreated": atomic.LoadInt64(&globalRecordsCreated),
		"globalRecordsGet":     atomic.LoadInt64(&globalRecordsGet),
		"globalRecordsPut":     atomic.LoadInt64(&globalRecordsPut),
	}
}

// ResetPoolStats 重置对象池使用统计信息
func ResetPoolStats() {
	atomic.StoreInt64(&recordCreated, 0)
	atomic.StoreInt64(&recordGet, 0)
	atomic.StoreInt64(&recordPutManual, 0)
	atomic.StoreInt64(&recordsCreated, 0)
	atomic.StoreInt64(&recordsGet, 0)
	atomic.StoreInt64(&recordsPutManual, 0)
}

// ResetGlobalPoolStats 重置全局对象池使用统计信息
func ResetGlobalPoolStats() {
	atomic.StoreInt64(&globalRecordCreated, 0)
	atomic.StoreInt64(&globalRecordGet, 0)
	atomic.StoreInt64(&globalRecordPut, 0)
	atomic.StoreInt64(&globalRecordsCreated, 0)
	atomic.StoreInt64(&globalRecordsGet, 0)
	atomic.StoreInt64(&globalRecordsPut, 0)
}
