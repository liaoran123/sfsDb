package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

func main() {
	fmt.Println("sfsDb 事务冲突和死锁测试")
	fmt.Println("=========================")

	// 清理旧的数据库目录
	cleanupDatabase()

	// 测试 1: 单个表的事务操作
	fmt.Println("\n测试 1: 单个表的事务操作")
	testSingleTableTransaction()

	// 测试 2: 多个表的事务操作
	fmt.Println("\n测试 2: 多个表的事务操作")
	testMultiTableTransaction()

	// 测试 3: 并发事务操作
	fmt.Println("\n测试 3: 并发事务操作")
	testConcurrentTransactions()

	// 测试 4: 事务回滚操作
	fmt.Println("\n测试 4: 事务回滚操作")
	testTransactionRollback()

	fmt.Println("\n所有测试完成！")

}

// 清理旧的数据库目录
func cleanupDatabase() {
	// 这里可以添加清理数据库目录的代码
	// 但为了安全，我们不自动删除目录
	fmt.Println("清理旧的数据库目录...")
}

// 测试单个表的事务操作
func testSingleTableTransaction() {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_single_table_transaction")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建用户表
	userTable, err := engine.TableNew("users")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 设置字段
	userFields := map[string]any{
		"id":      0,  // 用户ID
		"name":    "", // 用户名
		"age":     0,  // 年龄
		"email":   "", // 邮箱
		"address": "", // 地址
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 插入测试数据
	users := []map[string]any{
		{"id": 1, "name": "张三", "age": 25, "email": "zhangsan@example.com", "address": "北京市"},
		{"id": 2, "name": "李四", "age": 30, "email": "lisi@example.com", "address": "上海市"},
		{"id": 3, "name": "王五", "age": 35, "email": "wangwu@example.com", "address": "广州市"},
	}

	for _, user := range users {
		currentID, err := userTable.Insert(&user)
		if err != nil {
			fmt.Printf("插入数据失败: %v\n", err)
			return
		}
		fmt.Printf("插入用户成功, ID: %d\n", currentID)
	}

	// 开始事务
	tx, err := userTable.Begin()
	if err != nil {
		fmt.Printf("开始事务失败: %v\n", err)
		return
	}

	// 更新数据
	updateData := map[string]any{
		"id":      1,                          // 用于定位记录
		"email":   "zhangsan_new@example.com", // 更新邮箱
		"address": "深圳市",                      // 更新地址
	}
	err = tx.Update(&updateData)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		tx.Rollback()
		return
	}
	fmt.Println("更新数据成功")

	// 删除数据
	deleteData := map[string]any{
		"id": 3, // 用于定位要删除的记录
	}
	err = tx.Delete(&deleteData)
	if err != nil {
		fmt.Printf("删除数据失败: %v\n", err)
		tx.Rollback()
		return
	}
	fmt.Println("删除数据成功")

	// 提交事务
	err = tx.Commit()
	if err != nil {
		fmt.Printf("提交事务失败: %v\n", err)
		return
	}
	fmt.Println("事务提交成功")

	// 验证事务结果
	iter, err := userTable.Search(&map[string]any{"id": 1})
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}
	records := iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("更新后的数据: %v\n", records[0])
	}
	engine.GlobalTableIterPool.Put(iter)

	iter, err = userTable.Search(&map[string]any{"id": 3})
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}
	records = iter.GetRecords(true)
	fmt.Printf("删除后查询结果数: %d\n", len(records))
	engine.GlobalTableIterPool.Put(iter)

}

// 测试多个表的事务操作
func testMultiTableTransaction() {
	// 初始化数据库
	db, err := storage.OpenDefaultDb("./test_multi_table_transaction")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建用户表
	userTable, err := engine.TableNew("users")
	if err != nil {
		fmt.Printf("创建用户表失败: %v\n", err)
		return
	}

	// 设置用户表字段
	userFields := map[string]any{
		"id":   0,  // 用户ID
		"name": "", // 用户名
		"age":  0,  // 年龄
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		fmt.Printf("设置用户表字段失败: %v\n", err)
		return
	}

	// 创建用户表主键索引
	userPk, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建用户表主键索引失败: %v\n", err)
		return
	}
	userPk.AddFields("id")
	err = userTable.CreateIndex(userPk)
	if err != nil {
		fmt.Printf("创建用户表索引失败: %v\n", err)
		return
	}

	// 创建订单表
	orderTable, err := engine.TableNew("orders")
	if err != nil {
		fmt.Printf("创建订单表失败: %v\n", err)
		return
	}

	// 设置订单表字段
	orderFields := map[string]any{
		"id":      0,   // 订单ID
		"user_id": 0,   // 用户ID
		"amount":  0.0, // 订单金额
		"status":  "",  // 订单状态
	}
	err = orderTable.SetFields(orderFields)
	if err != nil {
		fmt.Printf("设置订单表字段失败: %v\n", err)
		return
	}

	// 创建订单表主键索引
	orderPk, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建订单表主键索引失败: %v\n", err)
		return
	}
	orderPk.AddFields("id")
	err = orderTable.CreateIndex(orderPk)
	if err != nil {
		fmt.Printf("创建订单表索引失败: %v\n", err)
		return
	}

	// 插入测试数据
	// 插入用户
	userData := map[string]any{"id": 1, "name": "张三", "age": 25}
	_, err = userTable.Insert(&userData)
	if err != nil {
		fmt.Printf("插入用户数据失败: %v\n", err)
		return
	}
	fmt.Println("插入用户数据成功")

	// 使用事务管理器执行多表事务
	batch := db.GetBatch()
	tm := engine.NewTransactionManager(batch)

	// 添加表到事务管理器
	userTx, err := tm.AddTable(userTable)
	if err != nil {
		fmt.Printf("添加用户表到事务管理器失败: %v\n", err)
		return
	}

	orderTx, err := tm.AddTable(orderTable)
	if err != nil {
		fmt.Printf("添加订单表到事务管理器失败: %v\n", err)
		return
	}

	// 更新用户信息
	updateUserData := map[string]any{"id": 1, "age": 26}
	err = userTx.Update(&updateUserData)
	if err != nil {
		fmt.Printf("更新用户数据失败: %v\n", err)
		tm.Rollback()
		return
	}
	fmt.Println("更新用户数据成功")

	// 插入订单
	orderData := map[string]any{"id": 1, "user_id": 1, "amount": 100.0, "status": "pending"}
	_, err = orderTx.Insert(&orderData)
	if err != nil {
		fmt.Printf("插入订单数据失败: %v\n", err)
		tm.Rollback()
		return
	}
	fmt.Println("插入订单数据成功")

	// 提交事务
	err = tm.Commit()
	if err != nil {
		fmt.Printf("提交事务失败: %v\n", err)
		return
	}
	fmt.Println("多表事务提交成功")

	// 验证事务结果
	iter, err := userTable.Search(&map[string]any{"id": 1})
	if err != nil {
		fmt.Printf("搜索用户失败: %v\n", err)
		return
	}
	records := iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("更新后的用户数据: %v\n", records[0])
	}
	engine.GlobalTableIterPool.Put(iter)

	iter, err = orderTable.Search(&map[string]any{"id": 1})
	if err != nil {
		fmt.Printf("搜索订单失败: %v\n", err)
		return
	}
	records = iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("插入的订单数据: %v\n", records[0])
	}
	engine.GlobalTableIterPool.Put(iter)

}

// 测试并发事务操作
func testConcurrentTransactions() {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_concurrent_transactions")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建用户表
	userTable, err := engine.TableNew("users")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 设置字段
	userFields := map[string]any{
		"id":      0,   // 用户ID
		"name":    "",  // 用户名
		"balance": 0.0, // 余额
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 插入测试数据
	user1 := map[string]any{"id": 1, "name": "张三", "balance": 1000.0}
	user2 := map[string]any{"id": 2, "name": "李四", "balance": 1000.0}

	_, err = userTable.Insert(&user1)
	if err != nil {
		fmt.Printf("插入用户1数据失败: %v\n", err)
		return
	}

	_, err = userTable.Insert(&user2)
	if err != nil {
		fmt.Printf("插入用户2数据失败: %v\n", err)
		return
	}

	fmt.Println("插入测试数据成功")

	// 并发执行两个事务
	var wg sync.WaitGroup
	errorChan := make(chan error, 2)

	// 事务 1: 张三向李四转账 100 元
	wg.Add(1)
	go func() {
		defer wg.Done()

		// 开始事务
		tx, err := userTable.Begin()
		if err != nil {
			errorChan <- fmt.Errorf("事务1开始失败: %v", err)
			return
		}

		// 读取张三的余额
		zhangsanData := map[string]any{"id": 1}
		_, err = tx.Read(&zhangsanData)
		if err != nil {
			tx.Rollback()
			errorChan <- fmt.Errorf("事务1读取张三数据失败: %v", err)
			return
		}

		// 模拟网络延迟
		time.Sleep(100 * time.Millisecond)

		// 更新张三的余额
		updateZhangsan := map[string]any{"id": 1, "balance": 900.0}
		err = tx.Update(&updateZhangsan)
		if err != nil {
			tx.Rollback()
			errorChan <- fmt.Errorf("事务1更新张三数据失败: %v", err)
			return
		}

		// 更新李四的余额
		updateLisi := map[string]any{"id": 2, "balance": 1100.0}
		err = tx.Update(&updateLisi)
		if err != nil {
			tx.Rollback()
			errorChan <- fmt.Errorf("事务1更新李四数据失败: %v", err)
			return
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			errorChan <- fmt.Errorf("事务1提交失败: %v", err)
			return
		}

		fmt.Println("事务1执行成功: 张三向李四转账 100 元")
	}()

	// 事务 2: 李四向张三转账 50 元
	wg.Add(1)
	go func() {
		defer wg.Done()

		// 开始事务
		tx, err := userTable.Begin()
		if err != nil {
			errorChan <- fmt.Errorf("事务2开始失败: %v", err)
			return
		}

		// 读取李四的余额
		lisiData := map[string]any{"id": 2}
		_, err = tx.Read(&lisiData)
		if err != nil {
			tx.Rollback()
			errorChan <- fmt.Errorf("事务2读取李四数据失败: %v", err)
			return
		}

		// 模拟网络延迟
		time.Sleep(100 * time.Millisecond)

		// 更新李四的余额
		updateLisi := map[string]any{"id": 2, "balance": 950.0}
		err = tx.Update(&updateLisi)
		if err != nil {
			tx.Rollback()
			errorChan <- fmt.Errorf("事务2更新李四数据失败: %v", err)
			return
		}

		// 更新张三的余额
		updateZhangsan := map[string]any{"id": 1, "balance": 1050.0}
		err = tx.Update(&updateZhangsan)
		if err != nil {
			tx.Rollback()
			errorChan <- fmt.Errorf("事务2更新张三数据失败: %v", err)
			return
		}

		// 提交事务
		err = tx.Commit()
		if err != nil {
			errorChan <- fmt.Errorf("事务2提交失败: %v", err)
			return
		}

		fmt.Println("事务2执行成功: 李四向张三转账 50 元")
	}()

	// 等待所有事务完成
	wg.Wait()
	close(errorChan)

	// 检查错误
	hasError := false
	for err := range errorChan {
		fmt.Printf("并发事务错误: %v\n", err)
		hasError = true
	}

	if !hasError {
		fmt.Println("并发事务执行成功，没有死锁或冲突")
	}

	// 验证最终余额
	iter, err := userTable.Search(&map[string]any{"id": 1})
	if err != nil {
		fmt.Printf("搜索张三失败: %v\n", err)
		return
	}
	records := iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("张三最终余额: %v\n", records[0]["balance"])
	}
	engine.GlobalTableIterPool.Put(iter)

	iter, err = userTable.Search(&map[string]any{"id": 2})
	if err != nil {
		fmt.Printf("搜索李四失败: %v\n", err)
		return
	}
	records = iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("李四最终余额: %v\n", records[0]["balance"])
	}
	engine.GlobalTableIterPool.Put(iter)

}

// 测试事务回滚操作
func testTransactionRollback() {
	// 初始化数据库
	_, err := storage.OpenDefaultDb("./test_transaction_rollback")
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建用户表
	userTable, err := engine.TableNew("users")
	if err != nil {
		fmt.Printf("创建表失败: %v\n", err)
		return
	}

	// 设置字段
	userFields := map[string]any{
		"id":      0,   // 用户ID
		"name":    "",  // 用户名
		"balance": 0.0, // 余额
	}
	err = userTable.SetFields(userFields)
	if err != nil {
		fmt.Printf("设置字段失败: %v\n", err)
		return
	}

	// 创建主键索引
	primaryKey, err := engine.DefaultPrimaryKeyNew("id")
	if err != nil {
		fmt.Printf("创建主键索引失败: %v\n", err)
		return
	}
	primaryKey.AddFields("id")
	err = userTable.CreateIndex(primaryKey)
	if err != nil {
		fmt.Printf("创建索引失败: %v\n", err)
		return
	}

	// 插入测试数据
	userData := map[string]any{"id": 1, "name": "张三", "balance": 1000.0}
	_, err = userTable.Insert(&userData)
	if err != nil {
		fmt.Printf("插入用户数据失败: %v\n", err)
		return
	}

	fmt.Println("插入测试数据成功")

	// 开始事务
	tx, err := userTable.Begin()
	if err != nil {
		fmt.Printf("开始事务失败: %v\n", err)
		return
	}

	// 更新用户余额
	updateData := map[string]any{"id": 1, "balance": 2000.0}
	err = tx.Update(&updateData)
	if err != nil {
		fmt.Printf("更新数据失败: %v\n", err)
		tx.Rollback()
		return
	}

	fmt.Println("更新数据成功，准备回滚事务")

	// 回滚事务
	err = tx.Rollback()
	if err != nil {
		fmt.Printf("回滚事务失败: %v\n", err)
		return
	}

	fmt.Println("事务回滚成功")

	// 验证回滚结果
	iter, err := userTable.Search(&map[string]any{"id": 1})
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}
	records := iter.GetRecords(true)
	if len(records) > 0 {
		fmt.Printf("回滚后余额: %v\n", records[0]["balance"])
		if records[0]["balance"] == 1000.0 {
			fmt.Println("验证成功: 余额已回滚到原始值")
		} else {
			fmt.Println("验证失败: 余额没有回滚到原始值")
		}
	}
	engine.GlobalTableIterPool.Put(iter)

}
