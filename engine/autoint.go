package engine

import "sync"

type AutoInt int

var autoIntLock sync.Mutex

func (a *AutoInt) Add(b int) int {
	//添加读写锁
	autoIntLock.Lock()
	defer autoIntLock.Unlock()
	*a += AutoInt(b)
	return int(*a)
}
func (a *AutoInt) Get() int {
	autoIntLock.Lock()
	defer autoIntLock.Unlock()
	return int(*a)
}
func (a *AutoInt) Set(b int) {
	autoIntLock.Lock()
	defer autoIntLock.Unlock()
	*a = AutoInt(b)
}
func (a *AutoInt) Increment() int {
	autoIntLock.Lock()
	defer autoIntLock.Unlock()
	*a++
	return int(*a)
}
func (a *AutoInt) Decrement() int {
	autoIntLock.Lock()
	defer autoIntLock.Unlock()
	*a--
	return int(*a)
}
func (a *AutoInt) Reset() {
	autoIntLock.Lock()
	defer autoIntLock.Unlock()
	*a = 0
}
