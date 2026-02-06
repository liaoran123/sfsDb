package main

import (
	"fmt"

	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/storage"
	"github.com/liaoran123/sfsDb/web"
)

func main() {
	fmt.Println("sfsDb Web Interface")
	fmt.Println("====================")
	fmt.Println("Starting web server...")

	// 1. 初始化数据库
	fmt.Println("1. Initializing database...")
	store, err := storage.OpenDefaultDb("./test_db")
	if err != nil {
		fmt.Printf("Error: Failed to open database: %v\n", err)
		return
	}
	defer storage.CloseDb()
	fmt.Println("✓ Database initialized successfully")

	// 2. 创建管理管理器
	fmt.Println("2. Creating management manager...")
	manager := management.NewManager(store)
	fmt.Println("✓ Management manager created successfully")

	// 3. 创建并启动Web服务器
	fmt.Println("3. Starting web server...")
	addr := ":8888"
	server := web.NewServer(addr, manager)
	if err := server.Start(); err != nil {
		fmt.Printf("Error: Failed to start web server: %v\n", err)
		return
	}

	fmt.Println("✓ Web server started successfully")
	fmt.Println("\nWeb Interface Access:")
	fmt.Printf("  URL: http://localhost%s\n", addr)
	fmt.Println("  Status: Running")
	fmt.Println("\nFeatures available:")
	fmt.Println("  • System Status - Memory and storage information")
	fmt.Println("  • System Information - Table and index details")
	fmt.Println("  • Index Management - List and analyze indexes")
	fmt.Println("  • Configuration - View database configuration")
	fmt.Println("  • Backup Management - Create and restore backups")
	fmt.Println("  • Performance Statistics - Query statistics and hotspots")
	fmt.Println("  • Memory Monitoring - Memory usage and GC statistics")
	fmt.Println("  • Key Monitoring - Key-value change statistics")
	fmt.Println("\nPress Ctrl+C to stop the server...")

	// 阻塞主协程
	select {}
}
