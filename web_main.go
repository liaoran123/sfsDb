package main

import (
	"fmt"
	"time"

	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/web"
)

func main1111() {
	// 进行相关测试
	//main20()
	//return
	// 初始化数据库
	dbManager := storage.GetDBManager()
	// 使用时间戳创建唯一的数据库路径
	timestamp := time.Now().Format("20060102150405")
	dbPath := fmt.Sprintf("./web_test_db_%s", timestamp)
	fmt.Printf("使用数据库路径: %s\n", dbPath)
	db, err := dbManager.OpenDB(dbPath)
	if err != nil {
		fmt.Printf("打开数据库失败: %v\n", err)
		return
	}
	defer dbManager.CloseDB()

	// 初始化管理模块
	manager := management.NewManager(db)

	// 创建web服务器
	server := web.NewServer(":8080", manager)

	// 使用固定的API密钥（管理员权限）
	fixedAPIKey := "test_api_key_for_development"
	server.AuthManager().AddAPIKey(fixedAPIKey, web.RoleAdmin)
	fmt.Println("=====================================")
	fmt.Println("默认API密钥 (管理员权限):")
	fmt.Printf("%s\n", fixedAPIKey)
	fmt.Println("=====================================")
	fmt.Println("请在前端API配置中使用此密钥")
	fmt.Println("=====================================")

	// 启动web服务器
	fmt.Println("启动web服务器...")
	err = server.Start()
	if err != nil {
		fmt.Printf("启动服务器失败: %v\n", err)
		return
	}
}
