package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/web"
)

func main() {
	fmt.Println("Testing sfsDb Web Interface Integration...")

	// 初始化数据库
	store, err := storage.OpenDefaultDb("./test_kvdb_perf")
	if err != nil {
		fmt.Printf("Failed to open database: %v\n", err)
		return
	}
	defer storage.CloseDb()

	// 创建管理器
	manager := management.NewManager(store)

	// 核心配置
	configMgr := manager.ConfigManager()
	configMgr.SetConfig("web_enable", "true")
	configMgr.SetConfig("web_port", ":8084")

	// 启动Web服务器
	server := web.NewServer(":8084", manager)
	fmt.Println("Starting web server...")
	fmt.Println("Web interface will be available at http://localhost:8084")
	fmt.Println("Press Ctrl+C to stop")

	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start web server: %v\n", err)
	}
}
