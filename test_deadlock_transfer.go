package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/storage"
)

var (
	accountTable *engine.Table
	wg           sync.WaitGroup
	errCount     int
	successCount int
	mu           sync.Mutex
)

// 初始化数据库和表
func initDB() error {
	// 清理旧的数据库目录
	_ = os.RemoveAll("./test_kvdb_deadlock")

	// 打开数据库
	_, err := storage.OpenDefaultDb("./test_kvdb_deadlock")
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 创建账户表
	table, err := engine.TableNew("accounts")
	if err != nil {
		return fmt.Errorf("创建账户表失败: %v", err)
	}

	// 设置字段
	fields := map[string]any{
		"id":      0,
		"name":    "",
		"balance": 0.0,
	}
	err = table.SetFields(fields)
	if err != nil {
		return fmt.Errorf("设置字段失败: %v", err)
	}

	accountTable = table
	return nil
}

// 初始化测试数据
func initTestData() error {
	// 插入测试账户
	accounts := []map[string]any{
		{"id": 1, "name": "张三", "balance": 1000.0},
		{"id": 2, "name": "李四", "balance": 1000.0},
		{"id": 3, "name": "王五", "balance": 1000.0},
		{"id": 4, "name": "赵六", "balance": 1000.0},
		{"id": 5, "name": "钱七", "balance": 1000.0},
	}

	for _, account := range accounts {
		_, err := accountTable.Insert(&account)
		if err != nil {
			return fmt.Errorf("插入账户数据失败: %v", err)
		}
	}

	return nil
}

// 转账函数（故意制造死锁）
func transfer(fromID, toID int, amount float64, txID int) error {
	// 故意使用不同的操作顺序，制造死锁
	if fromID < toID {
		// 顺序1: 先读转出账户，再读转入账户
		return transferWithOrder(fromID, toID, amount, txID)
	} else {
		// 顺序2: 先读转入账户，再读转出账户
		return transferWithReverseOrder(fromID, toID, amount, txID)
	}
}

// 转账函数（顺序1）
func transferWithOrder(fromID, toID int, amount float64, txID int) error {
	// 开始事务
	tx, err := accountTable.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 模拟处理时间，增加死锁概率
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	// 1. 更新转出账户（模拟减少余额）
	updateFromFields := map[string]any{
		"id":      fromID,
		"balance": 900.0, // 假设余额足够
	}
	if err := tx.Update(&updateFromFields); err != nil {
		tx.Rollback()
		return fmt.Errorf("更新转出账户失败: %v", err)
	}

	// 模拟处理时间，增加死锁概率
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	// 2. 更新转入账户（模拟增加余额）
	updateToFields := map[string]any{
		"id":      toID,
		"balance": 1100.0, // 假设余额增加
	}
	if err := tx.Update(&updateToFields); err != nil {
		tx.Rollback()
		return fmt.Errorf("更新转入账户失败: %v", err)
	}

	// 3. 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// 转账函数（顺序2，反向）
func transferWithReverseOrder(fromID, toID int, amount float64, txID int) error {
	// 开始事务
	tx, err := accountTable.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 模拟处理时间，增加死锁概率
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	// 1. 更新转入账户（注意：这里与上面顺序相反）
	updateToFields := map[string]any{
		"id":      toID,
		"balance": 1100.0, // 假设余额增加
	}
	if err := tx.Update(&updateToFields); err != nil {
		tx.Rollback()
		return fmt.Errorf("更新转入账户失败: %v", err)
	}

	// 模拟处理时间，增加死锁概率
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

	// 2. 更新转出账户
	updateFromFields := map[string]any{
		"id":      fromID,
		"balance": 900.0, // 假设余额足够
	}
	if err := tx.Update(&updateFromFields); err != nil {
		tx.Rollback()
		return fmt.Errorf("更新转出账户失败: %v", err)
	}

	// 3. 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	return nil
}

// 执行转账操作（带重试机制）
func executeTransfer(fromID, toID int, amount float64, txID int) {
	defer wg.Done()

	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := transfer(fromID, toID, amount, txID)
		if err == nil {
			// 转账成功
			mu.Lock()
			successCount++
			fmt.Printf("事务%d: 转账成功: %d -> %d, 金额: %.2f\n", txID, fromID, toID, amount)
			mu.Unlock()
			return
		}

		// 检查是否是死锁错误
		if containsDeadlockError(err.Error()) {
			fmt.Printf("事务%d: 检测到死锁，正在重试 (%d/%d)...\n", txID, attempt+1, maxRetries)
			// 指数退避
			backoffTime := time.Duration(attempt*100) * time.Millisecond
			time.Sleep(backoffTime)
			continue
		}

		// 其他错误
		mu.Lock()
		errCount++
		fmt.Printf("事务%d: 转账失败: %v\n", txID, err)
		mu.Unlock()
		return
	}

	// 达到最大重试次数
	mu.Lock()
	errCount++
	fmt.Printf("事务%d: 转账失败，已达到最大重试次数\n", txID)
	mu.Unlock()
}

// 检查错误信息是否包含死锁错误
func containsDeadlockError(errMsg string) bool {
	return len(errMsg) > 0 && (contains(errMsg, "检测到死锁") || contains(errMsg, "deadlock detected"))
}

// 字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) != -1
}

// 简单的子字符串查找
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// 打印账户余额
func printAccountBalances() {
	fmt.Println("\n=== 测试完成 ===")
	fmt.Printf("成功转账: %d 次\n", successCount)
	fmt.Printf("失败转账: %d 次\n", errCount)
}

func main20() {
	// 初始化随机数种子
	// 从 Go 1.20 起无需手动调用 rand.Seed，全局随机源已自动初始化

	// 初始化数据库
	fmt.Println("初始化数据库...")
	if err := initDB(); err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 初始化测试数据
	fmt.Println("初始化测试数据...")
	if err := initTestData(); err != nil {
		fmt.Printf("初始化测试数据失败: %v\n", err)
		return
	}

	// 打印初始余额
	printAccountBalances()

	// 配置死锁检测
	fmt.Println("\n配置死锁检测...")
	accountTable.SetDeadlockCheckInterval(100 * time.Millisecond)

	// 启动并发转账测试
	fmt.Println("\n开始并发转账测试...")
	const concurrentCount = 100 // 并发事务数

	wg.Add(concurrentCount)
	for i := 0; i < concurrentCount; i++ {
		go func(txID int) {
			// 随机选择两个不同的账户
			fromID := rand.Intn(5) + 1
			toID := rand.Intn(5) + 1
			for fromID == toID {
				toID = rand.Intn(5) + 1
			}

			// 随机转账金额
			amount := float64(rand.Intn(100)) + float64(rand.Intn(100))/100.0

			executeTransfer(fromID, toID, amount, txID)
		}(i)
	}

	// 等待所有事务完成
	wg.Wait()

	// 打印最终结果
	fmt.Printf("\n=== 测试结果 ===\n")
	fmt.Printf("成功次数: %d\n", successCount)
	fmt.Printf("失败次数: %d\n", errCount)
	fmt.Printf("总事务数: %d\n", concurrentCount)

	// 打印最终余额
	printAccountBalances()

	fmt.Println("\n测试完成!")
}
