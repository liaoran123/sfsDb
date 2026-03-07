package storage

import (
	"bytes"
	"os"
	"testing"
)

// TestEncryptionBasic 测试基本的加密解密功能
func TestEncryptionBasic(t *testing.T) {
	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	// 创建加密配置
	encryptConfig := &EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}

	// 创建加密存储
	encryptedStore, err := dbManager.NewLevelDBStore("./test_encrypted_db", nil, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer encryptedStore.Close()
	defer RemoveDir("./test_encrypted_db")

	// 测试基本的Put和Get操作
	testKey := []byte("test_key")
	testValue := []byte("test_value")

	// 存储加密数据
	err = encryptedStore.Put(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to put encrypted data: %v", err)
	}

	// 获取解密数据
	retrievedValue, err := encryptedStore.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get encrypted data: %v", err)
	}

	// 验证数据一致性
	if !bytes.Equal(retrievedValue, testValue) {
		t.Fatalf("Retrieved value does not match original: got %s, want %s", retrievedValue, testValue)
	}

	// 测试删除操作
	err = encryptedStore.Delete(testKey)
	if err != nil {
		t.Fatalf("Failed to delete encrypted data: %v", err)
	}

	// 验证数据已删除
	_, err = encryptedStore.Get(testKey)
	if err == nil {
		t.Fatal("Expected ErrNotFound after deletion, got nil")
	}
	if err != ErrNotFound {
		t.Fatalf("Expected ErrNotFound after deletion, got %v", err)
	}
}

// TestEncryptionBatch 测试加密批量操作
func TestEncryptionBatch(t *testing.T) {
	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	// 创建加密配置
	encryptConfig := &EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}

	// 创建加密存储
	encryptedStore, err := dbManager.NewLevelDBStore("./test_encrypted_batch_db", nil, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer encryptedStore.Close()
	defer RemoveDir("./test_encrypted_batch_db")

	// 创建批量操作
	batch := encryptedStore.GetBatch()

	// 添加多个操作到批量
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range testData {
		batch.Put([]byte(k), []byte(v))
	}

	// 执行批量操作
	err = encryptedStore.WriteBatch(batch)
	if err != nil {
		t.Fatalf("Failed to write encrypted batch: %v", err)
	}

	// 验证所有数据都已正确存储和加密
	for k, v := range testData {
		retrievedValue, err := encryptedStore.Get([]byte(k))
		if err != nil {
			t.Fatalf("Failed to get encrypted batch data for key %s: %v", k, err)
		}

		if !bytes.Equal(retrievedValue, []byte(v)) {
			t.Fatalf("Retrieved value for key %s does not match original: got %s, want %s", k, retrievedValue, v)
		}
	}
}

// TestEncryptionIterator 测试加密迭代器
func TestEncryptionIterator(t *testing.T) {
	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	// 创建加密配置
	encryptConfig := &EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}

	// 创建加密存储
	encryptedStore, err := dbManager.NewLevelDBStore("./test_encrypted_iterator_db", nil, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer encryptedStore.Close()
	defer RemoveDir("./test_encrypted_iterator_db")

	// 存储测试数据
	testData := map[string]string{
		"a_key": "a_value",
		"b_key": "b_value",
		"c_key": "c_value",
	}

	for k, v := range testData {
		err = encryptedStore.Put([]byte(k), []byte(v))
		if err != nil {
			t.Fatalf("Failed to put encrypted data: %v", err)
		}
	}

	// 使用迭代器遍历数据
	iter := encryptedStore.Iterator(nil, nil)
	defer iter.Release()

	// 收集迭代结果
	iteratedData := make(map[string]string)
	for iter.First(); iter.Valid(); iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		iteratedData[key] = value
	}

	// 验证迭代结果与原始数据一致
	if len(iteratedData) != len(testData) {
		t.Fatalf("Iterated data count mismatch: got %d, want %d", len(iteratedData), len(testData))
	}

	for k, v := range testData {
		if iteratedData[k] != v {
			t.Fatalf("Iterated value for key %s does not match original: got %s, want %s", k, iteratedData[k], v)
		}
	}
}

// TestEncryptionSnapshot 测试加密快照
func TestEncryptionSnapshot(t *testing.T) {
	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	// 创建加密配置
	encryptConfig := &EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}

	// 创建加密存储
	encryptedStore, err := dbManager.NewLevelDBStore("./test_encrypted_snapshot_db", nil, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer encryptedStore.Close()
	defer RemoveDir("./test_encrypted_snapshot_db")

	// 存储初始数据
	testKey := []byte("test_key")
	testValue1 := []byte("test_value1")
	testValue2 := []byte("test_value2")

	err = encryptedStore.Put(testKey, testValue1)
	if err != nil {
		t.Fatalf("Failed to put initial data: %v", err)
	}

	// 创建快照
	snapshot, err := encryptedStore.Snapshot()
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}
	defer snapshot.Release()

	// 更新数据
	err = encryptedStore.Put(testKey, testValue2)
	if err != nil {
		t.Fatalf("Failed to update data: %v", err)
	}

	// 从快照读取数据，应该得到旧值
	snapshotValue, err := snapshot.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get data from snapshot: %v", err)
	}

	if !bytes.Equal(snapshotValue, testValue1) {
		t.Fatalf("Snapshot value does not match initial value: got %s, want %s", snapshotValue, testValue1)
	}

	// 从主存储读取数据，应该得到新值
	storeValue, err := encryptedStore.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get data from store: %v", err)
	}

	if !bytes.Equal(storeValue, testValue2) {
		t.Fatalf("Store value does not match updated value: got %s, want %s", storeValue, testValue2)
	}
}

// TestEncryptionReEncrypt 测试密钥轮换功能
func TestEncryptionReEncrypt(t *testing.T) {
	// 生成初始密钥
	oldKey := make([]byte, 32)
	for i := range oldKey {
		oldKey[i] = byte(i)
	}

	// 生成新密钥
	newKey := make([]byte, 32)
	for i := range newKey {
		newKey[i] = byte(255 - i)
	}

	// 创建初始加密配置
	oldEncryptConfig := &EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: oldKey,
	}

	// 创建加密存储
	encryptedStore, err := dbManager.NewLevelDBStore("./test_encrypted_reencrypt_db", nil, oldEncryptConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted store: %v", err)
	}
	defer encryptedStore.Close()
	defer RemoveDir("./test_encrypted_reencrypt_db")

	// 存储测试数据
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	for k, v := range testData {
		err = encryptedStore.Put([]byte(k), []byte(v))
		if err != nil {
			t.Fatalf("Failed to put data: %v", err)
		}
	}

	// 执行密钥轮换
	encryptedWrapper, ok := encryptedStore.(*EncryptedStoreWrapper)
	if !ok {
		t.Fatal("Failed to convert to EncryptedStoreWrapper")
	}

	err = encryptedWrapper.ReEncrypt(newKey)
	if err != nil {
		t.Fatalf("Failed to re-encrypt data: %v", err)
	}

	// 验证数据仍然可以正确读取
	for k, v := range testData {
		retrievedValue, err := encryptedStore.Get([]byte(k))
		if err != nil {
			t.Fatalf("Failed to get data after re-encryption: %v", err)
		}

		if !bytes.Equal(retrievedValue, []byte(v)) {
			t.Fatalf("Retrieved value after re-encryption does not match original: got %s, want %s", retrievedValue, v)
		}
	}
}

// TestEncryptionWithPassword 测试使用密码派生密钥
func TestEncryptionWithPassword(t *testing.T) {
	// 创建加密配置，使用密码派生密钥
	encryptConfig := &EncryptionConfig{
		Enabled:    true,
		Algorithm:  "AES-256-GCM",
		Password:   "test_password",
		Salt:       []byte("test_salt"),
		Iterations: 100000,
	}

	// 创建加密存储
	encryptedStore, err := dbManager.NewLevelDBStore("./test_encrypted_password_db", nil, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to create encrypted store with password: %v", err)
	}
	defer encryptedStore.Close()
	defer RemoveDir("./test_encrypted_password_db")

	// 测试基本的Put和Get操作
	testKey := []byte("test_key")
	testValue := []byte("test_value")

	err = encryptedStore.Put(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to put encrypted data with password: %v", err)
	}

	retrievedValue, err := encryptedStore.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get encrypted data with password: %v", err)
	}

	if !bytes.Equal(retrievedValue, testValue) {
		t.Fatalf("Retrieved value does not match original: got %s, want %s", retrievedValue, testValue)
	}
}

// TestEncryptionDisabled 测试未启用加密的情况
func TestEncryptionDisabled(t *testing.T) {
	// 创建加密配置，但禁用加密
	encryptConfig := &EncryptionConfig{
		Enabled: false,
	}

	// 创建存储，应该返回普通的LevelDBStore
	store, err := dbManager.NewLevelDBStore("./test_encrypted_disabled_db", nil, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to create store with encryption disabled: %v", err)
	}
	defer store.Close()
	defer RemoveDir("./test_encrypted_disabled_db")

	// 验证返回的是LevelDBStore而不是EncryptedStoreWrapper
	_, isEncrypted := store.(*EncryptedStoreWrapper)
	if isEncrypted {
		t.Fatal("Expected LevelDBStore, got EncryptedStoreWrapper when encryption is disabled")
	}

	// 测试基本操作
	testKey := []byte("test_key")
	testValue := []byte("test_value")

	err = store.Put(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	retrievedValue, err := store.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if !bytes.Equal(retrievedValue, testValue) {
		t.Fatalf("Retrieved value does not match original: got %s, want %s", retrievedValue, testValue)
	}
}

// TestScenarioConfig 测试场景配置
func TestScenarioConfig(t *testing.T) {
	/*
			const (
			ScenarioEmbedded = "embedded" // 嵌入式场景，低延迟、高并发
			ScenarioIoT      = "iot"      // 物联网场景，低功耗、高并发
			ScenarioEdge     = "edge"     // 边缘场景，低延迟、高并发
			ScenarioGame     = "game"     // 游戏场景，低延迟、高并发
			ScenarioDefault  = "default"  // 默认场景，平衡配置
		)
	*/
	// 测试各个场景配置
	testScenarios := []struct {
		name     string
		scenario string
	}{
		{"Embedded", ScenarioEmbedded},
		{"IoT", ScenarioIoT},
		{"Edge", ScenarioEdge},
		{"Game", ScenarioGame},
	}

	for _, ts := range testScenarios {
		t.Run(ts.name, func(t *testing.T) {
			// 使用场景配置创建存储
			opts := GetScenarioOptions(ts.scenario)
			store, err := dbManager.NewLevelDBStore("./test_scenario_"+ts.name, opts)
			if err != nil {
				t.Fatalf("Failed to create store with scenario %s: %v", ts.name, err)
			}
			defer store.Close()
			defer RemoveDir("./test_scenario_" + ts.name)

			// 测试基本的Put和Get操作
			testKey := []byte("test_key")
			testValue := []byte("test_value")

			err = store.Put(testKey, testValue)
			if err != nil {
				t.Fatalf("Failed to put data with scenario %s: %v", ts.name, err)
			}

			retrievedValue, err := store.Get(testKey)
			if err != nil {
				t.Fatalf("Failed to get data with scenario %s: %v", ts.name, err)
			}

			if !bytes.Equal(retrievedValue, testValue) {
				t.Fatalf("Retrieved value with scenario %s does not match: got %s, want %s", ts.name, retrievedValue, testValue)
			}
		})
	}
}

// TestScenarioConfigWithEncryption 测试场景配置与加密结合
func TestScenarioConfigWithEncryption(t *testing.T) {
	// 生成测试密钥
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	// 创建加密配置
	encryptConfig := &EncryptionConfig{
		Enabled:   true,
		Algorithm: "AES-256-GCM",
		MasterKey: masterKey,
	}

	// 测试各个场景配置与加密结合
	testScenarios := []struct {
		name     string
		scenario string
	}{
		{"Embedded", ScenarioEmbedded},
		{"IoT", ScenarioIoT},
		{"Edge", ScenarioEdge},
		{"Game", ScenarioGame},
	}

	for _, ts := range testScenarios {
		t.Run(ts.name, func(t *testing.T) {
			// 使用场景配置和加密创建存储
			opts := GetScenarioOptions(ts.scenario)
			store, err := dbManager.NewLevelDBStore("./test_scenario_enc_"+ts.name, opts, encryptConfig)
			if err != nil {
				t.Fatalf("Failed to create store with scenario %s and encryption: %v", ts.name, err)
			}
			defer store.Close()
			defer RemoveDir("./test_scenario_enc_" + ts.name)

			// 测试基本的Put和Get操作
			testKey := []byte("test_key")
			testValue := []byte("test_value")

			err = store.Put(testKey, testValue)
			if err != nil {
				t.Fatalf("Failed to put encrypted data with scenario %s: %v", ts.name, err)
			}

			retrievedValue, err := store.Get(testKey)
			if err != nil {
				t.Fatalf("Failed to get encrypted data with scenario %s: %v", ts.name, err)
			}

			if !bytes.Equal(retrievedValue, testValue) {
				t.Fatalf("Retrieved encrypted value with scenario %s does not match: got %s, want %s", ts.name, retrievedValue, testValue)
			}

			// 测试删除操作
			err = store.Delete(testKey)
			if err != nil {
				t.Fatalf("Failed to delete encrypted data with scenario %s: %v", ts.name, err)
			}

			// 验证数据已删除
			_, err = store.Get(testKey)
			if err == nil {
				t.Fatal("Expected ErrNotFound after deletion, got nil")
			}
			if err != ErrNotFound {
				t.Fatalf("Expected ErrNotFound after deletion, got %v", err)
			}
		})
	}
}

// TestScenarioConfigWithEncryptionDisabled 测试场景配置与禁用加密
func TestScenarioConfigWithEncryptionDisabled(t *testing.T) {
	// 创建加密配置，但禁用加密
	encryptConfig := &EncryptionConfig{
		Enabled: false,
	}

	// 使用场景配置和禁用加密创建存储
	opts := GetScenarioOptions(ScenarioEdge)
	store, err := dbManager.NewLevelDBStore("./test_scenario_enc_disabled", opts, encryptConfig)
	if err != nil {
		t.Fatalf("Failed to create store with scenario and encryption disabled: %v", err)
	}
	defer store.Close()
	defer RemoveDir("./test_scenario_enc_disabled")

	// 验证返回的是LevelDBStore而不是EncryptedStoreWrapper
	_, isEncrypted := store.(*EncryptedStoreWrapper)
	if isEncrypted {
		t.Fatal("Expected LevelDBStore, got EncryptedStoreWrapper when encryption is disabled")
	}

	// 测试基本操作
	testKey := []byte("test_key")
	testValue := []byte("test_value")

	err = store.Put(testKey, testValue)
	if err != nil {
		t.Fatalf("Failed to put data: %v", err)
	}

	retrievedValue, err := store.Get(testKey)
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if !bytes.Equal(retrievedValue, testValue) {
		t.Fatalf("Retrieved value does not match: got %s, want %s", retrievedValue, testValue)
	}
}

// TestNewEncryptedStoreWrapperEdgeCases 测试 NewEncryptedStoreWrapper 的各种边缘情况
func TestNewEncryptedStoreWrapperEdgeCases(t *testing.T) {
	testCases := []struct {
		name            string
		config          *EncryptionConfig
		expectError     bool
		expectEncrypted bool
	}{
		{
			name: "Encryption not enabled",
			config: &EncryptionConfig{
				Enabled: false,
			},
			expectError:     false,
			expectEncrypted: false,
		},
		{
			name: "No master key and no password",
			config: &EncryptionConfig{
				Enabled: true,
			},
			expectError:     true,
			expectEncrypted: false,
		},
		{
			name: "Invalid algorithm",
			config: &EncryptionConfig{
				Enabled:   true,
				Algorithm: "INVALID-ALGORITHM",
				MasterKey: make([]byte, 32),
			},
			expectError:     true,
			expectEncrypted: false,
		},
		{
			name: "With master key",
			config: &EncryptionConfig{
				Enabled:   true,
				Algorithm: "AES-256-GCM",
				MasterKey: make([]byte, 32),
			},
			expectError:     false,
			expectEncrypted: true,
		},
		{
			name: "With password and salt and iterations",
			config: &EncryptionConfig{
				Enabled:    true,
				Password:   "test_password",
				Salt:       []byte("test_salt"),
				Iterations: 50000,
			},
			expectError:     false,
			expectEncrypted: true,
		},
		{
			name: "With password only (no salt, no iterations)",
			config: &EncryptionConfig{
				Enabled:  true,
				Password: "test_password",
			},
			expectError:     false,
			expectEncrypted: true,
		},
		{
			name: "With password and salt only (no iterations)",
			config: &EncryptionConfig{
				Enabled:  true,
				Password: "test_password",
				Salt:     []byte("test_salt"),
			},
			expectError:     false,
			expectEncrypted: true,
		},
		{
			name: "With password and iterations only (no salt)",
			config: &EncryptionConfig{
				Enabled:    true,
				Password:   "test_password",
				Iterations: 50000,
			},
			expectError:     false,
			expectEncrypted: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 直接使用 dbManager.NewLevelDBStore 创建存储
			store, err := dbManager.NewLevelDBStore("./test_enc_wrapper_"+tc.name, nil, tc.config)

			if tc.expectError && err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			defer func() {
				if store != nil {
					store.Close()
				}
				RemoveDir("./test_enc_wrapper_" + tc.name)
			}()

			if !tc.expectError && store == nil {
				t.Fatalf("Expected store, got nil")
			}

			if !tc.expectError && store != nil {
				// 验证是否是加密存储
				wrapper, isEncrypted := store.(*EncryptedStoreWrapper)
				if tc.expectEncrypted && !isEncrypted {
					t.Fatalf("Expected EncryptedStoreWrapper, got %T", store)
				}
				if !tc.expectEncrypted && isEncrypted {
					t.Fatalf("Expected LevelDBStore, got EncryptedStoreWrapper")
				}

				// 如果是加密存储，进行额外验证
				if tc.expectEncrypted && wrapper != nil {
					// 验证 config 是否正确保存
					savedConfig := wrapper.GetEncryptionConfig()
					if savedConfig == nil {
						t.Fatalf("Expected saved config, got nil")
					}

					// 如果是密码派生，验证盐值和迭代次数已保存
					if len(tc.config.MasterKey) == 0 {
						if savedConfig.Salt == nil {
							t.Fatalf("Expected salt to be saved")
						}
						if len(savedConfig.Salt) == 0 {
							t.Fatalf("Expected salt to have length > 0")
						}
						if savedConfig.Iterations <= 0 {
							t.Fatalf("Expected iterations to be > 0, got %d", savedConfig.Iterations)
						}
					}
				}

				// 测试基本的功能
				testKey := []byte("test_key")
				testValue := []byte("test_value")

				err = store.Put(testKey, testValue)
				if err != nil {
					t.Fatalf("Failed to put data: %v", err)
				}

				retrievedValue, err := store.Get(testKey)
				if err != nil {
					t.Fatalf("Failed to get data: %v", err)
				}

				if !bytes.Equal(retrievedValue, testValue) {
					t.Fatalf("Retrieved value does not match: got %s, want %s", retrievedValue, testValue)
				}
			}
		})
	}
}

// mockStore 用于测试的模拟 Store
type mockStore struct {
	data map[string][]byte
}

func newMockStore() *mockStore {
	return &mockStore{
		data: make(map[string][]byte),
	}
}

func (ms *mockStore) Get(key []byte) ([]byte, error) {
	if val, ok := ms.data[string(key)]; ok {
		return val, nil
	}
	return nil, ErrNotFound
}

func (ms *mockStore) Put(key, value []byte) error {
	ms.data[string(key)] = value
	return nil
}

func (ms *mockStore) Delete(key []byte) error {
	delete(ms.data, string(key))
	return nil
}

func (ms *mockStore) GetBatch() Batch {
	return &mockBatch{
		store: ms,
	}
}

func (ms *mockStore) WriteBatch(batch Batch, put ...bool) error {
	mb, ok := batch.(*mockBatch)
	if !ok {
		return NewError("invalid batch type")
	}
	for _, op := range mb.ops {
		if op.isDelete {
			delete(ms.data, string(op.key))
		} else {
			ms.data[string(op.key)] = op.value
		}
	}
	return nil
}

func (ms *mockStore) Iterator(start, limit []byte) Iterator {
	return &mockIterator{}
}

func (ms *mockStore) Snapshot() (Snapshot, error) {
	return &mockSnapshot{}, nil
}

func (ms *mockStore) SwitchToSnapshot() error {
	return nil
}

func (ms *mockStore) SwitchToDB() error {
	return nil
}

func (ms *mockStore) Close() error {
	return nil
}

// mockBatch 模拟的批量操作
type mockBatch struct {
	store *mockStore
	ops   []struct {
		key      []byte
		value    []byte
		isDelete bool
	}
}

func (mb *mockBatch) Put(key, value []byte) {
	mb.ops = append(mb.ops, struct {
		key      []byte
		value    []byte
		isDelete bool
	}{key: key, value: value, isDelete: false})
}

func (mb *mockBatch) Delete(key []byte) {
	mb.ops = append(mb.ops, struct {
		key      []byte
		value    []byte
		isDelete bool
	}{key: key, isDelete: true})
}

func (mb *mockBatch) Len() int {
	return len(mb.ops)
}

func (mb *mockBatch) Reset() {
	mb.ops = nil
}

// mockIterator 模拟的迭代器
type mockIterator struct{}

func (mi *mockIterator) First() bool          { return false }
func (mi *mockIterator) Last() bool           { return false }
func (mi *mockIterator) Seek(key []byte) bool { return false }
func (mi *mockIterator) Next() bool           { return false }
func (mi *mockIterator) Prev() bool           { return false }
func (mi *mockIterator) Key() []byte          { return nil }
func (mi *mockIterator) Value() []byte        { return nil }
func (mi *mockIterator) Valid() bool          { return false }
func (mi *mockIterator) Release()             {}

// mockSnapshot 模拟的快照
type mockSnapshot struct{}

func (ms *mockSnapshot) Get(key []byte) ([]byte, error)        { return nil, nil }
func (ms *mockSnapshot) Iterator(start, limit []byte) Iterator { return &mockIterator{} }
func (ms *mockSnapshot) Release() error                        { return nil }

// TestNewEncryptedStoreWrapperDirect 直接测试 NewEncryptedStoreWrapper
func TestNewEncryptedStoreWrapperDirect(t *testing.T) {
	mockStore := newMockStore()
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	tests := []struct {
		name        string
		config      *EncryptionConfig
		expectError bool
	}{
		{
			name: "Encryption not enabled",
			config: &EncryptionConfig{
				Enabled: false,
			},
			expectError: true,
		},
		{
			name: "No master key and no password",
			config: &EncryptionConfig{
				Enabled: true,
			},
			expectError: true,
		},
		{
			name: "Invalid algorithm",
			config: &EncryptionConfig{
				Enabled:   true,
				Algorithm: "INVALID-ALGORITHM",
				MasterKey: masterKey,
			},
			expectError: true,
		},
		{
			name: "Invalid key length (too short)",
			config: &EncryptionConfig{
				Enabled:   true,
				MasterKey: make([]byte, 16),
			},
			expectError: true,
		},
		{
			name: "Invalid key length (too long)",
			config: &EncryptionConfig{
				Enabled:   true,
				MasterKey: make([]byte, 64),
			},
			expectError: true,
		},
		{
			name: "Valid master key",
			config: &EncryptionConfig{
				Enabled:   true,
				MasterKey: masterKey,
			},
			expectError: false,
		},
		{
			name: "Valid with password",
			config: &EncryptionConfig{
				Enabled:  true,
				Password: "test_password",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper, err := NewEncryptedStoreWrapper(mockStore, tt.config)
			if tt.expectError && err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tt.expectError && wrapper == nil {
				t.Fatalf("Expected wrapper, got nil")
			}

			if !tt.expectError && wrapper != nil {
				savedConfig := wrapper.GetEncryptionConfig()
				if savedConfig == nil {
					t.Fatalf("Expected config to be saved")
				}
			}
		})
	}
}

// TestNewEncryptedStoreWrapperNoSideEffects 测试 NewEncryptedStoreWrapper 不修改传入的 config
func TestNewEncryptedStoreWrapperNoSideEffects(t *testing.T) {
	mockStore := newMockStore()
	originalConfig := &EncryptionConfig{
		Enabled:    true,
		Password:   "test_password",
		Salt:       nil,
		Iterations: 0,
	}

	configCopy := *originalConfig

	wrapper, err := NewEncryptedStoreWrapper(mockStore, originalConfig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer wrapper.Close()

	if originalConfig.Salt != nil {
		t.Fatalf("Original config Salt was modified, expected nil")
	}
	if originalConfig.Iterations != 0 {
		t.Fatalf("Original config Iterations was modified, expected 0")
	}

	savedConfig := wrapper.GetEncryptionConfig()
	if savedConfig.Salt == nil || len(savedConfig.Salt) == 0 {
		t.Fatalf("Expected Salt to be saved in wrapper config")
	}
	if savedConfig.Iterations <= 0 {
		t.Fatalf("Expected Iterations to be saved in wrapper config")
	}

	if !bytes.Equal(originalConfig.MasterKey, configCopy.MasterKey) ||
		originalConfig.Password != configCopy.Password ||
		originalConfig.Enabled != configCopy.Enabled ||
		originalConfig.Algorithm != configCopy.Algorithm {
		t.Fatalf("Original config was modified unexpectedly")
	}
}

// TestAESGCMEncryptor 测试 AESGCMEncryptor
func TestAESGCMEncryptor(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		expectError bool
	}{
		{
			name:        "Invalid key length - 16 bytes",
			key:         make([]byte, 16),
			expectError: true,
		},
		{
			name:        "Invalid key length - 64 bytes",
			key:         make([]byte, 64),
			expectError: true,
		},
		{
			name:        "Valid key - 32 bytes",
			key:         make([]byte, 32),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor, err := NewAESGCMEncryptor(tt.key)
			if tt.expectError && err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if encryptor == nil {
					t.Fatalf("Expected encryptor, got nil")
				}
				if encryptor.Algorithm() != "AES-256-GCM" {
					t.Fatalf("Expected algorithm AES-256-GCM, got %s", encryptor.Algorithm())
				}
			}
		})
	}
}

// TestAESGCMEncryptorEncryptDecrypt 测试加密和解密
func TestAESGCMEncryptorEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	encryptor, err := NewAESGCMEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	testData := []struct {
		name string
		data []byte
	}{
		{"Empty data", []byte{}},
		{"Small data", []byte("hello world")},
		{"Large data", bytes.Repeat([]byte("a"), 10000)},
		{"Binary data", []byte{0x00, 0x01, 0xFF, 0xFE}},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := encryptor.Encrypt(tt.data)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			if bytes.Equal(ciphertext, tt.data) {
				t.Fatalf("Ciphertext should not equal plaintext")
			}

			plaintext, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if !bytes.Equal(plaintext, tt.data) {
				t.Fatalf("Decrypted data doesn't match original: got %x, want %x", plaintext, tt.data)
			}
		})
	}
}

// TestAESGCMEncryptorDecryptInvalid 测试解密无效数据
func TestAESGCMEncryptorDecryptInvalid(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	encryptor, err := NewAESGCMEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	tests := []struct {
		name        string
		ciphertext  []byte
		expectError bool
	}{
		{"Too short", []byte{1}, true},
		{"Corrupted data", bytes.Repeat([]byte{0}, 100), true},
		{"Wrong nonce", func() []byte {
			ct, _ := encryptor.Encrypt([]byte("test"))
			ct[0] ^= 0xFF
			return ct
		}(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			if tt.expectError && err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

// TestDeriveKey 测试 DeriveKey 函数
func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name        string
		password    []byte
		salt        []byte
		iterations  int
		expectError bool
	}{
		{
			name:        "All parameters provided",
			password:    []byte("test_password"),
			salt:        []byte("test_salt_123456"),
			iterations:  50000,
			expectError: false,
		},
		{
			name:        "No salt, should generate",
			password:    []byte("test_password"),
			salt:        nil,
			iterations:  100000,
			expectError: false,
		},
		{
			name:        "No iterations, should use default",
			password:    []byte("test_password"),
			salt:        []byte("test_salt_123456"),
			iterations:  0,
			expectError: false,
		},
		{
			name:        "Negative iterations, should use default",
			password:    []byte("test_password"),
			salt:        []byte("test_salt_123456"),
			iterations:  -100,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := DeriveKey(tt.password, tt.salt, tt.iterations)
			if tt.expectError && err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if len(key) != 32 {
					t.Fatalf("Expected key length 32, got %d", len(key))
				}
			}
		})
	}
}

// TestDeriveKeyConsistency 测试 DeriveKey 的一致性
func TestDeriveKeyConsistency(t *testing.T) {
	password := []byte("test_password")
	salt := []byte("test_salt_123456")
	iterations := 100000

	key1, err := DeriveKey(password, salt, iterations)
	if err != nil {
		t.Fatalf("First derive failed: %v", err)
	}

	key2, err := DeriveKey(password, salt, iterations)
	if err != nil {
		t.Fatalf("Second derive failed: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Fatalf("Keys not consistent: %x != %x", key1, key2)
	}

	key3, err := DeriveKey(password, []byte("different_salt"), iterations)
	if err != nil {
		t.Fatalf("Third derive failed: %v", err)
	}

	if bytes.Equal(key1, key3) {
		t.Fatalf("Keys should differ with different salts")
	}
}

// RemoveDir 删除目录（用于测试清理）
func RemoveDir(path string) {
	os.RemoveAll(path)
}
