package record

import (
	"sync"
)

// 对象池，用于复用 Record 和 Records 对象
var (
	recordsPool = &sync.Pool{
		New: func() any {
			// 预分配合理容量，减少后续扩容开销
			return make(Records, 0, 16)
		},
	}
)

/*
func GetRecord() Record {
	return make(Record, 8)
}
*/
// GetRecords 从对象池获取一个 Records 对象
// sync.Pool不会导致内存泄漏，确保返回的数据干净即可。
func GetRecords() Records {
	rs := recordsPool.Get().(Records)
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

	return rs
}

// GetRecordsWithCapacity 从对象池获取一个指定初始容量的 Records 对象
func GetRecordsWithCapacity(capacity int) Records {
	rs := recordsPool.Get().(Records)

	// 清空 Records 中的所有 Record 并将其放回对象池
	// 确保从池中获取的对象不包含任何未释放的 Record

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

	return rs
}

// PutRecords 将 Records 对象放回对象池
func PutRecords(rs Records) {
	if rs == nil {
		return
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
}
