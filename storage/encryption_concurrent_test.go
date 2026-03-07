package storage

import (
	"bytes"
	"sync"
	"testing"
)

// TestConcurrentReads 测试并发读取
func TestConcurrentReads(t *testing.T) {
	mockStore := newThreadSafeMockStore()

	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	config := &EncryptionConfig{
		Enabled:   true,
		MasterKey: masterKey,
	}

	store, err := NewEncryptedStoreWrapper(mockStore, config)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer store.Close()

	testKey := []byte("test_key")
	testValue := []byte("test_value")

	err = store.Put(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to put initial data: %v", err)
	}

	const numGoroutines = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				value, err := store.Get(testKey)
				if err != nil {
					t.Errorf("Failed to get data: %v", err)
					return
				}
				if !bytes.Equal(value, testValue) {
					t.Errorf("Value mismatch: got %s, want %s", value, testValue)
					return
				}
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentWrites 测试并发写入
func TestConcurrentWrites(t *testing.T) {
	mockStore := newThreadSafeMockStore()

	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	config := &EncryptionConfig{
		Enabled:   true,
		MasterKey: masterKey,
	}

	store, err := NewEncryptedStoreWrapper(mockStore, config)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer store.Close()

	const numGoroutines = 10
	const numKeysPerGoroutine = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numKeysPerGoroutine; j++ {
				key := []byte("key_" + string(rune('a'+goroutineID)) + "_" + string(rune('0'+j)))
				value := []byte("value_" + string(rune('a'+goroutineID)) + "_" + string(rune('0'+j)))
				err := store.Put(key, value)
				if err != nil {
					t.Errorf("Failed to put data: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < numKeysPerGoroutine; j++ {
			key := []byte("key_" + string(rune('a'+i)) + "_" + string(rune('0'+j)))
			expectedValue := []byte("value_" + string(rune('a'+i)) + "_" + string(rune('0'+j)))
			value, err := store.Get(key)
			if err != nil {
				t.Errorf("Failed to get data for key %s: %v", key, err)
				continue
			}
			if !bytes.Equal(value, expectedValue) {
				t.Errorf("Value mismatch for key %s: got %s, want %s", key, value, expectedValue)
			}
		}
	}
}

// TestConcurrentReadWrite 测试并发读写
func TestConcurrentReadWrite(t *testing.T) {
	mockStore := newThreadSafeMockStore()

	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	config := &EncryptionConfig{
		Enabled:   true,
		MasterKey: masterKey,
	}

	store, err := NewEncryptedStoreWrapper(mockStore, config)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer store.Close()

	testKey := []byte("concurrent_key")
	initialValue := []byte("initial_value")
	err = store.Put(testKey, initialValue)
	if err != nil {
		t.Fatalf("Failed to put initial data: %v", err)
	}

	const numReaders = 20
	const numWriters = 5
	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	for i := 0; i < numReaders; i++ {
		go func(readerID int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				value, err := store.Get(testKey)
				if err != nil {
					if err != ErrNotFound {
						t.Errorf("Reader %d failed to get data: %v", readerID, err)
					}
					continue
				}
				if value == nil {
					t.Errorf("Reader %d got nil value", readerID)
				}
			}
		}(i)
	}

	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				newValue := []byte("value_" + string(rune('0'+writerID)) + "_" + string(rune('0'+j)))
				err := store.Put(testKey, newValue)
				if err != nil {
					t.Errorf("Writer %d failed to put data: %v", writerID, err)
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

// threadSafeMockStore 线程安全的模拟存储
type threadSafeMockStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newThreadSafeMockStore() *threadSafeMockStore {
	return &threadSafeMockStore{
		data: make(map[string][]byte),
	}
}

func (ms *threadSafeMockStore) Get(key []byte) ([]byte, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	if val, ok := ms.data[string(key)]; ok {
		return val, nil
	}
	return nil, ErrNotFound
}

func (ms *threadSafeMockStore) Put(key, value []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.data[string(key)] = value
	return nil
}

func (ms *threadSafeMockStore) Delete(key []byte) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.data, string(key))
	return nil
}

func (ms *threadSafeMockStore) GetBatch() Batch {
	return &threadSafeMockBatch{
		store: ms,
	}
}

func (ms *threadSafeMockStore) WriteBatch(batch Batch, put ...bool) error {
	mb, ok := batch.(*threadSafeMockBatch)
	if !ok {
		return NewError("invalid batch type")
	}
	ms.mu.Lock()
	defer ms.mu.Unlock()
	for _, op := range mb.ops {
		if op.isDelete {
			delete(ms.data, string(op.key))
		} else {
			ms.data[string(op.key)] = op.value
		}
	}
	return nil
}

func (ms *threadSafeMockStore) Iterator(start, limit []byte) Iterator {
	return &mockIterator{}
}

func (ms *threadSafeMockStore) Snapshot() (Snapshot, error) {
	return &mockSnapshot{}, nil
}

func (ms *threadSafeMockStore) SwitchToSnapshot() error {
	return nil
}

func (ms *threadSafeMockStore) SwitchToDB() error {
	return nil
}

func (ms *threadSafeMockStore) Close() error {
	return nil
}

// threadSafeMockBatch 线程安全的模拟批量操作
type threadSafeMockBatch struct {
	store *threadSafeMockStore
	ops   []struct {
		key      []byte
		value    []byte
		isDelete bool
	}
}

func (mb *threadSafeMockBatch) Put(key, value []byte) {
	mb.ops = append(mb.ops, struct {
		key      []byte
		value    []byte
		isDelete bool
	}{key: key, value: value, isDelete: false})
}

func (mb *threadSafeMockBatch) Delete(key []byte) {
	mb.ops = append(mb.ops, struct {
		key      []byte
		value    []byte
		isDelete bool
	}{key: key, isDelete: true})
}

func (mb *threadSafeMockBatch) Len() int {
	return len(mb.ops)
}

func (mb *threadSafeMockBatch) Reset() {
	mb.ops = nil
}
