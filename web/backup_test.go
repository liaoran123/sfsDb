package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/storage"
)

// TestBackupEndpoint 测试备份相关的 HTTP 端点
func TestBackupEndpoint(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_web_test")
	backupDir := filepath.Join(testDir, "backups")
	testDbPath := filepath.Join(testDir, "test_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建测试数据库
	testDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDb.Close()

	// 写入测试数据
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	for k, v := range testData {
		if err := testDb.Put([]byte(k), []byte(v)); err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
	}

	// 创建管理实例
	manager := management.NewManager(testDb)

	// 创建 web 服务器
	server := NewServer(":8080", manager)

	// 创建 gin 引擎并注册路由
	engine := gin.Default()

	// 注册 API 路由，与 server.Start() 中的路由注册一致
	api := engine.Group("/api")
	{
		api.GET("/backup", server.handleBackup)
	}

	// 测试创建备份
	t.Run("CreateBackup", func(t *testing.T) {
		// 创建请求
		params := url.Values{}
		params.Add("op", "create")
		params.Add("path", backupDir)

		req, err := http.NewRequest("GET", "/api/backup?"+params.Encode(), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		// 验证备份目录是否存在
		if _, err := os.Stat(backupDir); os.IsNotExist(err) {
			t.Fatalf("Backup directory not created: %s", backupDir)
		}

		// 验证备份文件是否存在
		files, err := os.ReadDir(backupDir)
		if err != nil {
			t.Fatalf("Failed to read backup directory: %v", err)
		}

		if len(files) == 0 {
			t.Fatalf("No backup files created")
		}

		t.Logf("Backup created successfully")
	})

	// 测试获取备份状态
	t.Run("GetBackupStatus", func(t *testing.T) {
		// 创建请求
		params := url.Values{}

		req, err := http.NewRequest("GET", "/api/backup?"+params.Encode(), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		t.Logf("Backup status retrieved successfully")
	})
}

// TestBackupRestoreEndpoint 测试备份恢复的 HTTP 端点
func TestBackupRestoreEndpoint(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_web_restore_test")
	backupDir := filepath.Join(testDir, "backups")
	testDbPath := filepath.Join(testDir, "test_db")
	restoreDbPath := filepath.Join(testDir, "restore_db")

	// 清理测试目录
	defer func() {
		os.RemoveAll(testDir)
	}()

	// 创建源测试数据库
	sourceDb, err := storage.OpenDefaultDb(testDbPath)
	if err != nil {
		t.Fatalf("Failed to open source database: %v", err)
	}
	defer sourceDb.Close()

	// 写入测试数据
	testData := map[string]string{
		"restore_key1": "restore_value1",
		"restore_key2": "restore_value2",
	}
	for k, v := range testData {
		if err := sourceDb.Put([]byte(k), []byte(v)); err != nil {
			t.Fatalf("Failed to write test data: %v", err)
		}
	}

	// 创建备份
	backupManager := management.NewManager(sourceDb).BackupManager()
	backupPath, err := backupManager.Backup(backupDir)
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// 关闭源数据库
	sourceDb.Close()
	storage.KVDb = nil

	// 创建恢复目标数据库
	restoreDb, err := storage.OpenDefaultDb(restoreDbPath)
	if err != nil {
		t.Fatalf("Failed to open restore database: %v", err)
	}
	defer restoreDb.Close()

	// 创建管理实例
	manager := management.NewManager(restoreDb)

	// 创建 web 服务器
	server := NewServer(":8080", manager)

	// 创建 gin 引擎并注册路由
	engine := gin.Default()

	// 注册 API 路由，与 server.Start() 中的路由注册一致
	api := engine.Group("/api")
	{
		api.GET("/backup", server.handleBackup)
	}

	// 测试恢复备份
	t.Run("RestoreBackup", func(t *testing.T) {
		// 创建请求
		params := url.Values{}
		params.Add("op", "restore")
		params.Add("file", backupPath)

		req, err := http.NewRequest("GET", "/api/backup?"+params.Encode(), nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		// 验证恢复后的数据
		for k, expectedV := range testData {
			v, err := restoreDb.Get([]byte(k))
			if err != nil {
				t.Fatalf("Failed to get restored data for key %s: %v", k, err)
			}
			if string(v) != expectedV {
				t.Fatalf("Restored data mismatch for key %s: expected %s, got %s", k, expectedV, string(v))
			}
		}

		t.Logf("Backup restored successfully")
	})
}
