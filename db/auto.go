package db

import "sync"

//自增原子性操作
type Auto struct {
	i  int
	mu sync.Mutex
}

//初始化
func (a *Auto) Ini(i int) {
	a.i = i
}

//返回当前值并且递增1
func (a *Auto) Increment() int {
	a.mu.Lock()
	i := a.i
	a.i++
	a.mu.Unlock()
	return i
}
