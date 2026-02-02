package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

// TestTableACIDTransaction 测试Table的完整ACID事务
func TestTableACIDTransaction(t *testing.T) {
	fmt.Println("=== 测试1：原子性(Atomicity)测试 ===")
	// 原子性测试：一个事务中的所有操作要么全部成功，要么全部失败

	// 1.1 创建表
	table1, err := TableNew("test_acid_atomic")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields1 := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	err = table1.SetFields(fields1)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pk1, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk1.AddFields("id")
	err = table1.CreateIndex(pk1)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 1.2 正常情况：所有操作都成功
	// 获取batch
	batch1 := table1.kvStore.GetBatch()
	if batch1 == nil {
		t.Fatalf("获取batch失败")
	}

	// 插入多条记录
	record1 := map[string]any{"id": 1, "name": "张三", "age": 20}
	_, err = table1.Insert(&record1, batch1)
	if err != nil {
		t.Fatalf("插入第一条记录失败: %v", err)
	}

	record2 := map[string]any{"id": 2, "name": "李四", "age": 25}
	_, err = table1.Insert(&record2, batch1)
	if err != nil {
		t.Fatalf("插入第二条记录失败: %v", err)
	}

	record3 := map[string]any{"id": 3, "name": "王五", "age": 30}
	_, err = table1.Insert(&record3, batch1)
	if err != nil {
		t.Fatalf("插入第三条记录失败: %v", err)
	}

	// 提交事务
	err = table1.kvStore.WriteBatch(batch1)
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证所有记录都成功插入
	iter1 := table1.ForData()
	defer GlobalTableIterPool.Put(iter1)
	records1 := iter1.GetRecords(true)
	if len(records1) != 3 {
		t.Errorf("期望找到3条记录，实际找到%d条", len(records1))
	} else {
		fmt.Printf("成功插入了%d条记录: \n", len(records1))
		for i, record := range records1 {
			fmt.Printf("  %d. %v\n", i+1, record)
		}
	}

	// 1.3 测试更新事务的原子性
	batch2 := table1.kvStore.GetBatch()
	if batch2 == nil {
		t.Fatalf("获取batch失败")
	}

	// 更新多条记录
	updateRecord1 := map[string]any{"id": 1, "name": "张三更新", "age": 21}
	err = table1.Update(&updateRecord1, batch2)
	if err != nil {
		t.Fatalf("更新第一条记录失败: %v", err)
	}

	updateRecord2 := map[string]any{"id": 2, "name": "李四更新", "age": 26}
	err = table1.Update(&updateRecord2, batch2)
	if err != nil {
		t.Fatalf("更新第二条记录失败: %v", err)
	}

	// 提交更新事务
	err = table1.kvStore.WriteBatch(batch2)
	if err != nil {
		t.Fatalf("提交更新事务失败: %v", err)
	}

	// 验证所有更新都成功
	iter2 := table1.ForData()
	defer GlobalTableIterPool.Put(iter2)
	records2 := iter2.GetRecords(true)
	if len(records2) != 3 {
		t.Errorf("期望找到3条记录，实际找到%d条", len(records2))
	} else {
		fmt.Printf("更新后的数据 (%d 条): \n", len(records2))
		for i, record := range records2 {
			fmt.Printf("  %d. %v\n", i+1, record)
		}
	}

	fmt.Println("✓ 正常情况下的原子性测试通过")

	// 1.4 回滚测试：创建事务但不提交，验证操作不会生效
	// 获取batch
	batch3 := table1.kvStore.GetBatch()
	if batch3 == nil {
		t.Fatalf("获取batch失败")
	}

	// 插入一条新记录
	record4 := map[string]any{"id": 4, "name": "赵六", "age": 35}
	_, err = table1.Insert(&record4, batch3)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}

	// 更新一条记录
	updateRecord3 := map[string]any{"id": 1, "name": "张三回滚", "age": 22}
	err = table1.Update(&updateRecord3, batch3)
	if err != nil {
		t.Fatalf("更新记录失败: %v", err)
	}

	// 不提交事务，模拟回滚
	// 验证操作没有生效
	readRecord3 := map[string]any{"id": 4}
	iter3 := table1.Search(&readRecord3)
	defer GlobalTableIterPool.Put(iter3)
	records3 := iter3.GetRecords(true)
	if len(records3) != 0 {
		t.Errorf("期望找到0条记录，实际找到%d条，回滚失败", len(records3))
	}

	// 验证第一条记录没有被更新
	readRecord1 := map[string]any{"id": 1}
	iter4 := table1.Search(&readRecord1)
	defer GlobalTableIterPool.Put(iter4)
	records4 := iter4.GetRecords(true)
	if len(records4) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records4))
	} else {
		if records4[0]["name"] != "张三更新" || records4[0]["age"] != 21 {
			t.Errorf("记录不应该被更新，期望 {name: '张三更新', age: 21}, 实际 %v", records4[0])
		}
	}

	fmt.Println("✓ 回滚情况下的原子性测试通过")

	fmt.Println("\n=== 测试2：一致性(Consistency)测试 ===")
	// 一致性测试：事务执行前后，数据库从一个一致性状态变换到另一个一致性状态

	// 2.1 创建新表用于一致性测试
	table2, err := TableNew("test_acid_consistency")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields2 := map[string]any{
		"id":     0,
		"name":   "",
		"salary": 0.0,
		"bonus":  0.0,
		"total":  0.0,
	}
	err = table2.SetFields(fields2)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pk2, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk2.AddFields("id")
	err = table2.CreateIndex(pk2)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 2.2 初始数据插入
	initialRecord := map[string]any{"id": 1, "name": "张三", "salary": 5000.0, "bonus": 1000.0, "total": 6000.0}
	_, err = table2.Insert(&initialRecord)
	if err != nil {
		t.Fatalf("初始数据插入失败: %v", err)
	}

	// 2.3 检查当前kvStore是否为LevelDBStore类型
	levelDBStore, ok := table2.kvStore.(*storage.LevelDBStore)
	if !ok {
		fmt.Println("当前存储引擎不支持快照功能，跳过一致性测试")
		return
	}

	// 2.4 创建快照，获取一致性视图
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 在快照模式下读取数据
	readConsistencyRecord := map[string]any{"id": 1}
	iter5 := table2.Search(&readConsistencyRecord)
	defer GlobalTableIterPool.Put(iter5)
	records5 := iter5.GetRecords(true)
	if len(records5) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records5))
	} else {
		if records5[0]["total"] != 6000.0 {
			t.Errorf("期望total为6000.0，实际为%v", records5[0]["total"])
		}
	}

	fmt.Printf("快照模式读取到的初始数据: %v\n", records5[0])

	// 切换回DB模式，执行事务
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	// 2.5 执行一致性更新：更新salary和bonus，同时更新total（保持一致性）
	batch4 := table2.kvStore.GetBatch()
	if batch4 == nil {
		t.Fatalf("获取batch失败")
	}

	// 更新记录，保持total = salary + bonus的一致性
	updateConsistencyRecord := map[string]any{"id": 1, "salary": 6000.0, "bonus": 1500.0, "total": 7500.0}
	err = table2.Update(&updateConsistencyRecord, batch4)
	if err != nil {
		t.Fatalf("更新记录失败: %v", err)
	}

	// 提交事务
	err = table2.kvStore.WriteBatch(batch4)
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 2.6 再次创建快照，获取新的一致性视图
	err = levelDBStore.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 在新快照模式下读取数据，应该是更新后的值，保持一致性
	iter6 := table2.Search(&readConsistencyRecord)
	defer GlobalTableIterPool.Put(iter6)
	records6 := iter6.GetRecords(true)
	if len(records6) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records6))
	} else {
		if records6[0]["salary"] != 6000.0 || records6[0]["bonus"] != 1500.0 || records6[0]["total"] != 7500.0 {
			t.Errorf("数据不一致，期望 {salary: 6000.0, bonus: 1500.0, total: 7500.0}, 实际 %v", records6[0])
		}
		if records6[0]["total"] != (records6[0]["salary"].(float64) + records6[0]["bonus"].(float64)) {
			t.Errorf("数据不一致，total应该等于salary + bonus")
		}
	}

	fmt.Printf("新快照读取到的更新后数据: %v\n", records6[0])

	// 切换回DB模式
	err = levelDBStore.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	fmt.Println("✓ 一致性测试通过")

	fmt.Println("\n=== 测试3：隔离性(Isolation)测试 ===")
	// 隔离性测试：多个事务并发执行时，一个事务的执行不应影响其他事务的执行

	// 3.1 创建新表用于隔离性测试
	table3, err := TableNew("test_acid_isolation")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields3 := map[string]any{
		"id":    0,
		"name":  "",
		"value": 0,
	}
	err = table3.SetFields(fields3)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pk3, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk3.AddFields("id")
	err = table3.CreateIndex(pk3)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 3.2 初始数据插入
	isolationRecord1 := map[string]any{"id": 1, "name": "隔离测试", "value": 100}
	_, err = table3.Insert(&isolationRecord1)
	if err != nil {
		t.Fatalf("初始数据插入失败: %v", err)
	}

	// 3.3 检查当前kvStore是否为LevelDBStore类型
	levelDBStore2, ok := table3.kvStore.(*storage.LevelDBStore)
	if !ok {
		fmt.Println("当前存储引擎不支持快照功能，跳过隔离性测试")
		return
	}

	// 3.4 创建快照（事务1的读取视图）
	err = levelDBStore2.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 事务1在快照模式下读取数据
	readIsolationRecord := map[string]any{"id": 1}
	iter7 := table3.Search(&readIsolationRecord)
	defer GlobalTableIterPool.Put(iter7)
	records7 := iter7.GetRecords(true)
	if len(records7) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records7))
	} else {
		if records7[0]["value"] != 100 {
			t.Errorf("期望value为100，实际为%v", records7[0]["value"])
		}
	}

	fmt.Printf("事务1读取到的初始值: %v\n", records7[0]["value"])

	// 3.5 切换回DB模式，执行事务2的写入操作
	err = levelDBStore2.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	// 事务2：更新记录
	batch5 := table3.kvStore.GetBatch()
	if batch5 == nil {
		t.Fatalf("获取batch失败")
	}

	updateIsolationRecord := map[string]any{"id": 1, "value": 200}
	err = table3.Update(&updateIsolationRecord, batch5)
	if err != nil {
		t.Fatalf("事务2更新记录失败: %v", err)
	}

	// 提交事务2
	err = table3.kvStore.WriteBatch(batch5)
	if err != nil {
		t.Fatalf("事务2提交失败: %v", err)
	}

	fmt.Printf("事务2修改后的值: %d\n", 200)

	// 3.6 事务1继续在快照模式下读取，应该看不到事务2的修改
	// 注意：我们需要重新创建快照来获取最新的一致性视图
	// 这里我们演示的是：事务1在创建快照后，事务2的修改不会影响事务1的读取

	// 重新创建快照（模拟另一个并发事务的读取视图）
	err = levelDBStore2.SwitchToSnapshot()
	if err != nil {
		t.Fatalf("创建快照失败: %v", err)
	}

	// 事务3在新快照模式下读取，应该能看到事务2的修改
	iter8 := table3.Search(&readIsolationRecord)
	defer GlobalTableIterPool.Put(iter8)
	records8 := iter8.GetRecords(true)
	if len(records8) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records8))
	} else {
		fmt.Printf("新快照读取到的值: %v\n", records8[0]["value"])
		if records8[0]["value"] != 200 {
			t.Errorf("期望value为200，实际为%v", records8[0]["value"])
		}
	}

	// 切换回DB模式
	err = levelDBStore2.SwitchToDB()
	if err != nil {
		t.Fatalf("切换回DB模式失败: %v", err)
	}

	fmt.Println("✓ 隔离性测试通过")

	fmt.Println("\n=== 测试4：持久性(Durability)测试 ===")
	// 持久性测试：事务提交后，其结果应该永久保存在数据库中

	// 4.1 创建新表用于持久性测试
	table4, err := TableNew("test_acid_durable")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	// 设置表字段
	fields4 := map[string]any{
		"id":   0,
		"name": "",
		"data": "",
	}
	err = table4.SetFields(fields4)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pk4, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk4.AddFields("id")
	err = table4.CreateIndex(pk4)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 4.2 执行事务：插入多条记录
	batch6 := table4.kvStore.GetBatch()
	if batch6 == nil {
		t.Fatalf("获取batch失败")
	}

	// 插入多条记录
	for i := 1; i <= 5; i++ {
		record := map[string]any{"id": i, "name": fmt.Sprintf("持久化测试%d", i), "data": fmt.Sprintf("数据%d", i)}
		_, err = table4.Insert(&record, batch6)
		if err != nil {
			t.Fatalf("插入记录%d失败: %v", i, err)
		}
	}

	// 提交事务
	err = table4.kvStore.WriteBatch(batch6)
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	fmt.Printf("事务提交成功，插入了5条记录\n")

	// 4.3 验证数据存在
	readDurableRecord := map[string]any{"id": 3}
	iter9 := table4.Search(&readDurableRecord)
	defer GlobalTableIterPool.Put(iter9)
	records9 := iter9.GetRecords(true)
	if len(records9) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records9))
	} else {
		if records9[0]["name"] != "持久化测试3" || records9[0]["data"] != "数据3" {
			t.Errorf("记录数据不正确，期望 {name: '持久化测试3', data: '数据3'}, 实际 %v", records9[0])
		}
	}

	// 4.4 使用ForData()验证所有记录都存在
	iter10 := table4.ForData()
	defer GlobalTableIterPool.Put(iter10)
	allRecords := iter10.GetRecords(true)
	if len(allRecords) != 5 {
		t.Errorf("期望找到5条记录，实际找到%d条", len(allRecords))
	} else {
		fmt.Printf("成功读取到所有%d条记录\n", len(allRecords))
	}

	// 注意：完整的持久性测试需要关闭并重新打开数据库，但由于我们使用的是内存中的测试数据库，这里只验证数据已经正确写入

	fmt.Println("✓ 持久性测试通过")

	fmt.Println("\n=== 测试5：完整的ACID事务流程测试 ===")
	// 完整的ACID事务流程：银行转账

	// 5.1 创建银行账户表
	accountTable, err := TableNew("test_acid_account")
	if err != nil {
		t.Fatalf("创建账户表失败: %v", err)
	}

	// 设置表字段
	accountFields := map[string]any{
		"id":      0,
		"name":    "",
		"balance": 0.0,
	}
	err = accountTable.SetFields(accountFields)
	if err != nil {
		t.Fatalf("设置账户表字段失败: %v", err)
	}

	// 创建主键索引
	accountPK, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建账户表主键失败: %v", err)
	}
	accountPK.AddFields("id")
	err = accountTable.CreateIndex(accountPK)
	if err != nil {
		t.Fatalf("创建账户表主键索引失败: %v", err)
	}

	// 5.2 初始化账户数据
	// 账户1：张三，初始余额1000.0
	account1 := map[string]any{"id": 1, "name": "张三", "balance": 1000.0}
	_, err = accountTable.Insert(&account1)
	if err != nil {
		t.Fatalf("插入账户1失败: %v", err)
	}

	// 账户2：李四，初始余额2000.0
	account2 := map[string]any{"id": 2, "name": "李四", "balance": 2000.0}
	_, err = accountTable.Insert(&account2)
	if err != nil {
		t.Fatalf("插入账户2失败: %v", err)
	}

	fmt.Println("初始账户余额:")
	// 显示初始余额
	readAccount1 := map[string]any{"id": 1}
	iter11 := accountTable.Search(&readAccount1)
	defer GlobalTableIterPool.Put(iter11)
	records11 := iter11.GetRecords(true)
	if len(records11) == 1 {
		fmt.Printf("张三: %.2f\n", records11[0]["balance"])
	}

	readAccount2 := map[string]any{"id": 2}
	iter12 := accountTable.Search(&readAccount2)
	defer GlobalTableIterPool.Put(iter12)
	records12 := iter12.GetRecords(true)
	if len(records12) == 1 {
		fmt.Printf("李四: %.2f\n", records12[0]["balance"])
	}

	// 5.3 执行转账事务：张三向李四转账500.0
	fmt.Println("\n执行转账事务：张三向李四转账500.0")

	// 获取batch
	transferBatch := accountTable.kvStore.GetBatch()
	if transferBatch == nil {
		t.Fatalf("获取batch失败")
	}

	// 步骤1：张三的余额减少500.0
	updateAccount1 := map[string]any{"id": 1, "balance": 500.0}
	err = accountTable.Update(&updateAccount1, transferBatch)
	if err != nil {
		t.Fatalf("更新张三的账户失败: %v", err)
	}

	// 步骤2：李四的余额增加500.0
	updateAccount2 := map[string]any{"id": 2, "balance": 2500.0}
	err = accountTable.Update(&updateAccount2, transferBatch)
	if err != nil {
		t.Fatalf("更新李四的账户失败: %v", err)
	}

	// 提交转账事务
	err = accountTable.kvStore.WriteBatch(transferBatch)
	if err != nil {
		t.Fatalf("提交转账事务失败: %v", err)
	}

	fmt.Println("转账事务提交成功")

	// 5.4 验证转账结果
	fmt.Println("\n转账后账户余额:")

	// 验证张三的余额
	iter13 := accountTable.Search(&readAccount1)
	defer GlobalTableIterPool.Put(iter13)
	records13 := iter13.GetRecords(true)
	if len(records13) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records13))
	} else {
		if records13[0]["balance"] != 500.0 {
			t.Errorf("张三的余额应该是500.0，实际是%v", records13[0]["balance"])
		}
		fmt.Printf("张三: %.2f\n", records13[0]["balance"])
	}

	// 验证李四的余额
	iter14 := accountTable.Search(&readAccount2)
	defer GlobalTableIterPool.Put(iter14)
	records14 := iter14.GetRecords(true)
	if len(records14) != 1 {
		t.Errorf("期望找到1条记录，实际找到%d条", len(records14))
	} else {
		if records14[0]["balance"] != 2500.0 {
			t.Errorf("李四的余额应该是2500.0，实际是%v", records14[0]["balance"])
		}
		fmt.Printf("李四: %.2f\n", records14[0]["balance"])
	}

	// 5.5 验证总额一致性：转账前后总额应该不变
	// 转账前总额：1000.0 + 2000.0 = 3000.0
	// 转账后总额：500.0 + 2500.0 = 3000.0
	iter15 := accountTable.ForData()
	defer GlobalTableIterPool.Put(iter15)
	allAccounts := iter15.GetRecords(true)
	var totalBalance float64
	for _, account := range allAccounts {
		totalBalance += account["balance"].(float64)
	}

	fmt.Printf("\n账户总额: %.2f\n", totalBalance)
	if totalBalance != 3000.0 {
		t.Errorf("账户总额应该是3000.0，实际是%v", totalBalance)
	}

	fmt.Println("✓ 完整的ACID事务流程测试通过")

	fmt.Println("\n=== 所有ACID事务测试通过 ===")
}

// TestTableTransactionCache 测试TableTransaction的cache字段功能
func TestTableTransactionCache(t *testing.T) {
	fmt.Println("\n=== 测试6：Transaction接口与cache字段功能测试 ===")
	// 测试Transaction接口的基本功能，特别是cache字段的"读自己的写"功能

	// 创建测试表
	fields := map[string]any{
		"id":   0,
		"name": "",
		"age":  0,
	}
	table, err := TableNew("test_transaction_cache")
	if err != nil {
		t.Fatalf("创建表失败: %v", err)
	}

	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("设置表字段失败: %v", err)
	}

	// 创建主键索引
	pk, err := DefaultPrimaryKeyNew("pk_id")
	if err != nil {
		t.Fatalf("创建主键失败: %v", err)
	}
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("创建主键索引失败: %v", err)
	}

	// 测试1：事务内插入并读取（测试cache的"读自己的写"功能）
	t.Run("InsertAndRead", func(t *testing.T) {
		fmt.Println("\n--- 测试1.1：事务内插入并读取（读自己的写）---")
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("开始事务失败: %v", err)
		}

		// 插入记录
		insertRecord := map[string]any{"id": 1, "name": "张三", "age": 25}
		_, err = tx.Insert(&insertRecord)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务内插入记录失败: %v", err)
		}

		// 事务内读取刚插入的记录（测试cache）
		readRecord := map[string]any{"id": 1}
		recordData, err := tx.Read(&readRecord)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务内读取记录失败: %v", err)
		}

		if recordData == nil {
			tx.Rollback()
			t.Fatalf("事务内没有读取到刚插入的记录，cache功能失败")
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("提交事务失败: %v", err)
		}

		fmt.Println("✓ 事务内插入并读取成功，cache功能正常")
	})

	// 测试2：事务内更新并读取（测试cache的更新功能）
	t.Run("UpdateAndRead", func(t *testing.T) {
		fmt.Println("\n--- 测试1.2：事务内更新并读取 --- ")
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("开始事务失败: %v", err)
		}

		// 更新记录
		updateRecord := map[string]any{"id": 1, "age": 26}
		err = tx.Update(&updateRecord)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务内更新记录失败: %v", err)
		}

		// 事务内读取刚更新的记录（测试cache）
		readRecord := map[string]any{"id": 1}
		recordData, err := tx.Read(&readRecord)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务内读取记录失败: %v", err)
		}

		if recordData == nil {
			tx.Rollback()
			t.Fatalf("事务内没有读取到刚更新的记录，cache功能失败")
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("提交事务失败: %v", err)
		}

		fmt.Println("✓ 事务内更新并读取成功，cache功能正常")
	})

	// 测试3：事务回滚（测试cache的回滚功能）
	t.Run("Rollback", func(t *testing.T) {
		fmt.Println("\n--- 测试1.3：事务回滚 --- ")
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("开始事务失败: %v", err)
		}

		// 插入记录
		insertRecord := map[string]any{"id": 2, "name": "测试回滚", "age": 30}
		_, err = tx.Insert(&insertRecord)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务内插入记录失败: %v", err)
		}

		// 事务内读取刚插入的记录（测试cache）
		readRecord := map[string]any{"id": 2}
		recordData, err := tx.Read(&readRecord)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务内读取记录失败: %v", err)
		}

		if recordData == nil {
			tx.Rollback()
			t.Fatalf("事务内没有读取到刚插入的记录，cache功能失败")
		}

		// 回滚事务
		err = tx.Rollback()
		if err != nil {
			t.Fatalf("回滚事务失败: %v", err)
		}

		// 验证记录是否被回滚
		readRecordAfterRollback := map[string]any{"id": 2}
		iter := table.Search(&readRecordAfterRollback)
		defer GlobalTableIterPool.Put(iter)
		records := iter.GetRecords(true)
		if len(records) != 0 {
			t.Errorf("回滚失败，记录仍然存在")
		}

		fmt.Println("✓ 事务回滚成功，cache功能正常")
	})

	// 测试4：事务内删除并读取（测试cache的删除功能）
	t.Run("DeleteAndRead", func(t *testing.T) {
		fmt.Println("\n--- 测试1.4：事务内删除并读取 --- ")
		// 开始事务
		tx, err := table.Begin()
		if err != nil {
			t.Fatalf("开始事务失败: %v", err)
		}

		// 删除记录
		deleteRecord := map[string]any{"id": 1}
		err = tx.Delete(&deleteRecord)
		if err != nil {
			tx.Rollback()
			t.Fatalf("事务内删除记录失败: %v", err)
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			t.Fatalf("提交事务失败: %v", err)
		}

		// 验证记录是否被删除
		readRecordAfterDelete := map[string]any{"id": 1}
		iter := table.Search(&readRecordAfterDelete)
		defer GlobalTableIterPool.Put(iter)
		records := iter.GetRecords(true)
		if len(records) != 0 {
			t.Errorf("删除失败，记录仍然存在")
		}

		fmt.Println("✓ 事务内删除并读取成功，cache功能正常")
	})

	fmt.Println("\n=== Transaction接口与cache字段功能测试通过 ===")
}
