package storage

import (
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

// TestWriteBatchIoT 测试物联网场景下的批量写入功能
func TestWriteBatchIoT(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "sfsdb_iot_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 LevelDB 实例
	db, err := leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	// 创建 LevelDBStore
	store := &LevelDBStore{
		ldb:        db,
		originalDB: db,
	}

	// 获取 batch
	batch := store.GetBatch()
	batch.Put([]byte("key1"), []byte("value1"))
	batch.Put([]byte("key2"), []byte("value2"))
	batch.Put([]byte("key3"), []byte("value3"))

	// 测试 WriteBatchIoT 函数
	err = store.WriteBatchIoT(batch)
	if err != nil {
		t.Fatalf("WriteBatchIoT 执行失败: %v", err)
	}

	// 验证数据是否正确写入
	testCases := []struct {
		key   string
		value string
	}{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
	}

	for _, tc := range testCases {
		val, err := store.Get([]byte(tc.key))
		if err != nil {
			t.Errorf("获取键 %s 失败: %v", tc.key, err)
			continue
		}
		if string(val) != tc.value {
			t.Errorf("键 %s 的值不匹配，期望: %s, 实际: %s", tc.key, tc.value, string(val))
		}
	}

	t.Log("WriteBatchIoT 测试通过")
}

// TestWriteBatchIoTEmptyBatch 测试空 batch 的处理
func TestWriteBatchIoTEmptyBatch(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "sfsdb_iot_empty_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 LevelDB 实例
	db, err := leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	// 创建 LevelDBStore
	store := &LevelDBStore{
		ldb:        db,
		originalDB: db,
	}

	// 获取空 batch
	batch := store.GetBatch()

	// 测试空 batch 的处理
	err = store.WriteBatchIoT(batch)
	if err != nil {
		t.Fatalf("WriteBatchIoT 处理空 batch 失败: %v", err)
	}

	t.Log("WriteBatchIoT 空 batch 测试通过")
}

// TestWriteBatchIoTNilBatch 测试 nil batch 的处理
func TestWriteBatchIoTNilBatch(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "sfsdb_iot_nil_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 LevelDB 实例
	db, err := leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	// 创建 LevelDBStore
	store := &LevelDBStore{
		ldb:        db,
		originalDB: db,
	}

	// 测试 nil batch 的处理
	err = store.WriteBatchIoT(nil)
	if err == nil {
		t.Error("WriteBatchIoT 应该拒绝 nil batch")
	} else {
		t.Logf("WriteBatchIoT 正确拒绝了 nil batch: %v", err)
	}
}

// TestWriteBatchIoTDataDurability 测试数据持久性（模拟断电场景）
func TestWriteBatchIoTDataDurability(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "sfsdb_iot_durability_test")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 LevelDB 实例
	db, err := leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}

	// 创建 LevelDBStore
	store := &LevelDBStore{
		ldb:        db,
		originalDB: db,
	}

	// 获取 batch
	batch := store.GetBatch()
	batch.Put([]byte("sensor_data_1"), []byte("temperature:25.5"))
	batch.Put([]byte("sensor_data_2"), []byte("humidity:60"))
	batch.Put([]byte("sensor_data_3"), []byte("pressure:1013"))

	// 使用 WriteBatchIoT 写入数据（强制同步到磁盘）
	err = store.WriteBatchIoT(batch)
	if err != nil {
		t.Fatalf("WriteBatchIoT 执行失败: %v", err)
	}

	// 关闭数据库
	db.Close()

	// 重新打开数据库，验证数据是否持久化
	db, err = leveldb.OpenFile(tmpDir, nil)
	if err != nil {
		t.Fatalf("重新打开数据库失败: %v", err)
	}
	defer db.Close()

	// 创建新的 LevelDBStore
	store = &LevelDBStore{
		ldb:        db,
		originalDB: db,
	}

	// 验证数据是否正确持久化
	testCases := []struct {
		key   string
		value string
	}{
		{"sensor_data_1", "temperature:25.5"},
		{"sensor_data_2", "humidity:60"},
		{"sensor_data_3", "pressure:1013"},
	}

	for _, tc := range testCases {
		val, err := store.Get([]byte(tc.key))
		if err != nil {
			t.Errorf("获取键 %s 失败: %v", tc.key, err)
			continue
		}
		if string(val) != tc.value {
			t.Errorf("键 %s 的值不匹配，期望: %s, 实际: %s", tc.key, tc.value, string(val))
		}
	}

	t.Log("WriteBatchIoT 数据持久性测试通过")
}
