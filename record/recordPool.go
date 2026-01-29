package record

import (
	"runtime"
	"sync"
)

// 对象池，用于复用 Record 和 Records 对象
var (
	recordPool = &sync.Pool{
		New: func() any {
			// 预分配一定容量，减少后续扩容开销
			return make(Record, 8)
		},
	}

	recordsPool = &sync.Pool{
		New: func() any {
			// 预分配合理容量，减少后续扩容开销
			return make(Records, 0, 16)
		},
	}
)

// GetRecord 从对象池获取一个 Record 对象
// sync.Pool不会导致内存泄漏，确保返回的数据干净即可。
func GetRecord() Record {
	r := recordPool.Get().(Record)
	// 清空 Record 中的所有字段，确保返回的数据干净
	for k := range r {
		delete(r, k)
	}
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
	// 注意：由于 Go 的函数参数是按值传递的，我们无法在这里清除原始对象的 finalizer
	// finalizer 会在对象不再被引用时触发，但由于我们已经将对象放回池，
	// 当对象再次被获取时，会被重新设置 finalizer，这不会导致问题
}

// GetRecords 从对象池获取一个 Records 对象
// sync.Pool不会导致内存泄漏，确保返回的数据干净即可。
func GetRecords() Records {
	rs := recordsPool.Get().(Records)
	// 清空 Records 中的所有 Record 并将其放回对象池
	// 确保从池中获取的对象不包含任何未释放的 Record
	//这里主要是以防在gc回收时，记录中的Record没有被正确释放，多重保证。主要释放是下面的gc finalizer。
	for i, r := range rs {
		if r != nil {
			putRecordInternal(r)
			rs[i] = nil
		}
	}
	// 重置长度
	rs = rs[:0]

	// 检查容量是否过大，避免内存膨胀
	// 如果容量超过阈值，创建一个新的、容量适中的 Records 对象
	const maxCapacity = 1000 // 可根据实际情况调整阈值
	if cap(rs) > maxCapacity {
		// 创建一个新的、容量适中的 Records 对象
		newRs := make(Records, 0, 100) // 初始容量设为 100，可根据实际情况调整
		// 将旧的对象丢弃，让垃圾回收器回收
		rs = newRs
	}

	// 创建一个临时变量，用于设置finalizer
	temp := rs
	//这里只要是为了在gc回收时清除其中的Record，
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
			// 检查容量是否过大，避免内存膨胀
			if cap(obj) > maxCapacity {
				// 如果容量过大，创建一个新的、容量适中的对象放回池
				newObj := make(Records, 0, 100)
				recordsPool.Put(newObj)
			} else {
				// 重置 slice 长度并放回池
				recordsPool.Put(obj[:0])
			}
			// 清空指针，避免重复处理
			*ptr = nil
		}
	})
	return rs
}

// GetRecordsWithCapacity 从对象池获取一个指定初始容量的 Records 对象
func GetRecordsWithCapacity(capacity int) Records {
	rs := recordsPool.Get().(Records)

	// 清空 Records 中的所有 Record 并将其放回对象池
	// 确保从池中获取的对象不包含任何未释放的 Record
	for i, r := range rs {
		if r != nil {
			putRecordInternal(r)
			rs[i] = nil
		}
	}

	// 检查容量是否过大，避免内存膨胀
	const maxCapacity = 1000 // 可根据实际情况调整阈值
	if cap(rs) > maxCapacity {
		// 如果容量过大，创建一个新的、容量适中的 Records 对象
		newRs := make(Records, 0, capacity)
		rs = newRs
	} else if cap(rs) < capacity {
		// 如果当前容量不足，创建一个新的 slice 并将旧的放回池
		newRs := make(Records, 0, capacity)
		recordsPool.Put(rs[:0])
		rs = newRs
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
			// 检查容量是否过大，避免内存膨胀
			if cap(obj) > maxCapacity {
				// 如果容量过大，创建一个新的、容量适中的对象放回池
				newObj := make(Records, 0, 100)
				recordsPool.Put(newObj)
			} else {
				// 重置 slice 长度并放回池
				recordsPool.Put(obj[:0])
			}
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
	// 检查容量是否过大，避免内存膨胀
	const maxCapacity = 1000 // 可根据实际情况调整阈值
	if cap(rs) > maxCapacity {
		// 如果容量过大，创建一个新的、容量适中的对象放回池
		newRs := make(Records, 0, 100)
		recordsPool.Put(newRs)
	} else {
		// 重置 slice 长度并放回池
		recordsPool.Put(rs[:0])
	}
	// 注意：由于 Go 的函数参数是按值传递的，我们无法在这里清除原始对象的 finalizer
	// finalizer 会在对象不再被引用时触发，但由于我们已经将对象放回池，
	// 当对象再次被获取时，会被重新设置 finalizer，这不会导致问题
}
