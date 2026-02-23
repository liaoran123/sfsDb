package engine

import "sync/atomic"

type AutoInt int64

func (a *AutoInt) Add(b int) int {
	return int(atomic.AddInt64((*int64)(a), int64(b)))
}

func (a *AutoInt) Get() int {
	return int(atomic.LoadInt64((*int64)(a)))
}

func (a *AutoInt) Set(b int) {
	atomic.StoreInt64((*int64)(a), int64(b))
}

func (a *AutoInt) Increment() int {
	return int(atomic.AddInt64((*int64)(a), 1))
}

func (a *AutoInt) Decrement() int {
	return int(atomic.AddInt64((*int64)(a), -1))
}

func (a *AutoInt) Reset() {
	atomic.StoreInt64((*int64)(a), 0)
}

// IncrementBy 批量增加指定的值，并返回增加后的值
func (a *AutoInt) IncrementBy(n int) int {
	return int(atomic.AddInt64((*int64)(a), int64(n)))
}

// GetAndIncrementBy 原子地获取当前值并增加指定的数量，返回增加前的值
func (a *AutoInt) GetAndIncrementBy(n int) int {
	return int(atomic.AddInt64((*int64)(a), int64(n)) - int64(n))
}
