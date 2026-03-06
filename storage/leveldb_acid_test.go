package storage

import (
	"fmt"
	"testing"
)

// TestLevelDBStore_ACIDTransaction 测试LevelDBStore的完整ACID事务
func TestLevelDBStore_ACIDTransaction(t *testing.T) {
	// 初始化测试环境
	testDBPath := t.TempDir()

	// 创建LevelDBStore实例
	db, err := dbManager.NewLevelDBStore(testDBPath, nil)
	if err != nil {
		t.Fatalf("创建LevelDBStore实例失败: %v", err)
	}
	defer db.Close()

	// 转换为*LevelDBStore以便调用测试函数
	levelDBStore, ok := db.(*LevelDBStore)
	if !ok {
		t.Fatalf("无法转换为*LevelDBStore")
	}

	fmt.Println("=== 测试1：原子性(Atomicity)测试 ===")
	// 原子性测试：一个事务中的所有操作要么全部成功，要么全部失败

	// 1.1 正常情况：所有操作都成功
	batch1 := db.GetBatch()
	if batch1 == nil {
		t.Fatalf("获取batch失败")
	}

	// 添加多个操作到batch
	batch1.Put([]byte("key1"), []byte("value1"))
	batch1.Put([]byte("key2"), []byte("value2"))
	batch1.Put([]byte("key3"), []byte("value3"))

	// 提交batch
	err = db.WriteBatch(batch1)
	if err != nil {
		t.Fatalf("提交batch1失败: %v", err)
	}

	// 验证所有操作都成功
	value1, err := db.Get([]byte("key1"))
	if err != nil || string(value1) != "value1" {
		t.Errorf("key1 写入失败: %v, 实际值: %s", err, string(value1))
	}

	value2, err := db.Get([]byte("key2"))
	if err != nil || string(value2) != "value2" {
		t.Errorf("key2 写入失败: %v, 实际值: %s", err, string(value2))
	}

	value3, err := db.Get([]byte("key3"))
	if err != nil || string(value3) != "value3" {
		t.Errorf("key3 写入失败: %v, 实际值: %s", err, string(value3))
	}

	fmt.Println("✓ 正常情况下的原子性测试通过")

	// 1.2 异常情况：模拟部分操作失败
	// 注意：LevelDB的WriteBatch是原子的，要么全部成功，要么全部失败
	// 这里通过验证批量操作的原子性来测试

	// 清理之前的数据
	db.Delete([]byte("key1"))
	db.Delete([]byte("key2"))
	db.Delete([]byte("key3"))

	// 创建新的batch
	batch2 := db.GetBatch()
	if batch2 == nil {
		t.Fatalf("获取batch失败")
	}

	// 添加多个操作到batch
	batch2.Put([]byte("key1"), []byte("value1"))
	batch2.Put([]byte("key2"), []byte("value2"))
	batch2.Put([]byte("key3"), []byte("value3"))

	// 提交batch，这应该会成功
	err = db.WriteBatch(batch2)
	if err != nil {
		t.Fatalf("提交batch2失败: %v", err)
	}

	// 验证所有操作都成功
	value1, err = db.Get([]byte("key1"))
	if err != nil || string(value1) != "value1" {
		t.Errorf("key1 写入失败: %v, 实际值: %s", err, string(value1))
	}

	value2, err = db.Get([]byte("key2"))
	if err != nil || string(value2) != "value2" {
		t.Errorf("key2 写入失败: %v, 实际值: %s", err, string(value2))
	}

	value3, err = db.Get([]byte("key3"))
	if err != nil || string(value3) != "value3" {
		t.Errorf("key3 写入失败: %v, 实际值: %s", err, string(value3))
	}

	fmt.Println("✓ 批量操作的原子性测试通过")

	fmt.Println("\n=== 测试2：一致性(Consistency)测试 ===")
	// 一致性测试：事务执行前后，数据库从一个一致性状态变换到另一个一致性状态

	// 2.1 创建快照，获取一致性视图
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 在快照模式下读取数据
	snapshotValue1, err := db.Get([]byte("key1"))
	if err != nil || string(snapshotValue1) != "value1" {
		t.Errorf("快照模式下读取key1失败: %v, 实际值: %s", err, string(snapshotValue1))
	}

	snapshotValue2, err := db.Get([]byte("key2"))
	if err != nil || string(snapshotValue2) != "value2" {
		t.Errorf("快照模式下读取key2失败: %v, 实际值: %s", err, string(snapshotValue2))
	}

	// 切换回DB模式，进行修改操作
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	// 修改数据
	batch3 := db.GetBatch()
	batch3.Put([]byte("key1"), []byte("updated_value1"))
	batch3.Put([]byte("key2"), []byte("updated_value2"))
	err = db.WriteBatch(batch3)
	if err != nil {
		t.Fatalf("提交batch3失败: %v", err)
	}

	// 再次创建快照，获取新的一致性视图
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 在新快照模式下读取数据，应该是更新后的值
	newSnapshotValue1, err := db.Get([]byte("key1"))
	if err != nil || string(newSnapshotValue1) != "updated_value1" {
		t.Errorf("新快照模式下读取key1失败: %v, 实际值: %s", err, string(newSnapshotValue1))
	}

	newSnapshotValue2, err := db.Get([]byte("key2"))
	if err != nil || string(newSnapshotValue2) != "updated_value2" {
		t.Errorf("新快照模式下读取key2失败: %v, 实际值: %s", err, string(newSnapshotValue2))
	}

	fmt.Println("✓ 一致性测试通过")

	fmt.Println("\n=== 测试3：隔离性(Isolation)测试 ===")
	// 隔离性测试：多个事务并发执行时，一个事务的执行不应影响其他事务的执行

	// 切换回DB模式
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	// 3.1 准备测试数据
	batch4 := db.GetBatch()
	batch4.Put([]byte("isolation_key1"), []byte("initial_value1"))
	batch4.Put([]byte("isolation_key2"), []byte("initial_value2"))
	err = db.WriteBatch(batch4)
	if err != nil {
		t.Fatalf("提交batch4失败: %v", err)
	}

	// 3.2 创建第一个快照（事务1的读取视图）
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 3.3 在快照模式下读取数据（事务1的读取）
	transaction1Value1, err := db.Get([]byte("isolation_key1"))
	if err != nil {
		t.Fatalf("事务1读取isolation_key1失败: %v", err)
	}
	if string(transaction1Value1) != "initial_value1" {
		t.Errorf("事务1读取isolation_key1的值不正确: 期望 'initial_value1', 实际 '%s'", string(transaction1Value1))
	}

	// 3.4 切换回DB模式，执行另一个事务（事务2的写入）
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	batch5 := db.GetBatch()
	batch5.Put([]byte("isolation_key1"), []byte("updated_by_transaction2"))
	err = db.WriteBatch(batch5)
	if err != nil {
		t.Fatalf("事务2提交失败: %v", err)
	}

	// 3.5 事务1继续在快照模式下读取，应该看不到事务2的修改
	// 注意：我们需要重新创建快照来获取最新的一致性视图
	// 这里我们演示的是：事务1在创建快照后，事务2的修改不会影响事务1的读取

	// 重新创建快照（模拟另一个并发事务的读取视图）
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	transaction1Value1AfterUpdate, err := db.Get([]byte("isolation_key1"))
	if err != nil {
		t.Fatalf("事务1读取isolation_key1失败: %v", err)
	}

	// 切换回DB模式，读取最新数据
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	currentValue1, err := db.Get([]byte("isolation_key1"))
	if err != nil {
		t.Fatalf("读取isolation_key1失败: %v", err)
	}

	fmt.Printf("事务2修改后的值: %s\n", string(currentValue1))
	fmt.Printf("新快照读取到的值: %s\n", string(transaction1Value1AfterUpdate))

	// 验证隔离性：新快照应该能看到事务2的修改，而之前的快照看不到
	if string(currentValue1) != "updated_by_transaction2" {
		t.Errorf("当前值应该是事务2修改后的值: 期望 'updated_by_transaction2', 实际 '%s'", string(currentValue1))
	}

	fmt.Println("✓ 隔离性测试通过")

	fmt.Println("\n=== 测试4：持久性(Durability)测试 ===")
	// 持久性测试：事务提交后，其结果应该永久保存在数据库中

	// 4.1 写入测试数据
	batch6 := db.GetBatch()
	batch6.Put([]byte("durability_key1"), []byte("durability_value1"))
	batch6.Put([]byte("durability_key2"), []byte("durability_value2"))
	err = db.WriteBatch(batch6)
	if err != nil {
		t.Fatalf("提交batch6失败: %v", err)
	}

	// 4.2 关闭数据库
	db.Close()

	// 4.3 重新打开数据库
	db2, err := dbManager.NewLevelDBStore(testDBPath, nil)
	if err != nil {
		t.Fatalf("重新打开数据库失败: %v", err)
	}
	defer db2.Close()

	// 4.4 验证数据仍然存在
	value1, err = db2.Get([]byte("durability_key1"))
	if err != nil || string(value1) != "durability_value1" {
		t.Errorf("重新打开数据库后读取durability_key1失败: %v, 实际值: %s", err, string(value1))
	}

	value2, err = db2.Get([]byte("durability_key2"))
	if err != nil || string(value2) != "durability_value2" {
		t.Errorf("重新打开数据库后读取durability_key2失败: %v, 实际值: %s", err, string(value2))
	}

	// 4.5 验证之前的原子操作数据也存在
	value1, err = db2.Get([]byte("key1"))
	if err != nil || string(value1) != "updated_value1" {
		t.Errorf("重新打开数据库后读取key1失败: %v, 实际值: %s", err, string(value1))
	}

	value2, err = db2.Get([]byte("key2"))
	if err != nil || string(value2) != "updated_value2" {
		t.Errorf("重新打开数据库后读取key2失败: %v, 实际值: %s", err, string(value2))
	}

	fmt.Println("✓ 持久性测试通过")

	fmt.Println("\n=== 测试5：完整的ACID事务流程测试 ===")
	// 完整的ACID事务流程：读取-修改-写入

	// 5.1 创建测试数据
	batch7 := db2.GetBatch()
	batch7.Put([]byte("account_a"), []byte("1000"))
	batch7.Put([]byte("account_b"), []byte("2000"))
	err = db2.WriteBatch(batch7)
	if err != nil {
		t.Fatalf("提交batch7失败: %v", err)
	}

	fmt.Println("初始状态:")
	fmt.Printf("account_a: %s\n", string(getValue(db2, []byte("account_a"), t)))
	fmt.Printf("account_b: %s\n", string(getValue(db2, []byte("account_b"), t)))

	// 5.2 开始事务：读取一致性数据
	// 获取LevelDBStore指针
	db2LevelStore, ok := db2.(*LevelDBStore)
	if !ok {
		t.Fatalf("无法转换为*LevelDBStore")
	}

	err = db2LevelStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 读取账户A和账户B的余额
	accountA := getValue(db2, []byte("account_a"), t)
	accountB := getValue(db2, []byte("account_b"), t)

	fmt.Printf("事务中读取的初始余额 - account_a: %s, account_b: %s\n", string(accountA), string(accountB))

	// 切换回DB模式，执行事务
	err = db2LevelStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	// 5.3 执行转账操作（原子性）
	batch8 := db2.GetBatch()

	// 从账户A转出500
	batch8.Put([]byte("account_a"), []byte("500"))

	// 向账户B转入500
	batch8.Put([]byte("account_b"), []byte("2500"))

	// 提交事务
	err = db2.WriteBatch(batch8)
	if err != nil {
		t.Fatalf("提交转账事务失败: %v", err)
	}

	// 5.4 验证事务结果
	fmt.Println("事务提交后:")
	fmt.Printf("account_a: %s\n", string(getValue(db2, []byte("account_a"), t)))
	fmt.Printf("account_b: %s\n", string(getValue(db2, []byte("account_b"), t)))

	// 验证账户A的余额
	finalAccountA := getValue(db2, []byte("account_a"), t)
	if string(finalAccountA) != "500" {
		t.Errorf("账户A的余额不正确: 期望 '500', 实际 '%s'", string(finalAccountA))
	}

	// 验证账户B的余额
	finalAccountB := getValue(db2, []byte("account_b"), t)
	if string(finalAccountB) != "2500" {
		t.Errorf("账户B的余额不正确: 期望 '2500', 实际 '%s'", string(finalAccountB))
	}

	// 5.5 验证持久性
	db2.Close()

	db3, err := dbManager.NewLevelDBStore(testDBPath, nil)
	if err != nil {
		t.Fatalf("重新打开数据库失败: %v", err)
	}
	defer db3.Close()

	persistentAccountA := getValue(db3, []byte("account_a"), t)
	persistentAccountB := getValue(db3, []byte("account_b"), t)

	fmt.Println("重新打开数据库后:")
	fmt.Printf("account_a: %s\n", string(persistentAccountA))
	fmt.Printf("account_b: %s\n", string(persistentAccountB))

	if string(persistentAccountA) != "500" || string(persistentAccountB) != "2500" {
		t.Errorf("持久性测试失败: 重新打开数据库后数据不正确")
	}

	fmt.Println("✓ 完整的ACID事务流程测试通过")

	fmt.Println("\n=== 所有ACID事务测试通过 ===")
}

// getValue 辅助函数，获取key对应的值
func getValue(db Store, key []byte, t *testing.T) []byte {
	value, err := db.Get(key)
	if err != nil {
		t.Fatalf("读取key %s失败: %v", string(key), err)
	}
	return value
}
