package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/engine"
	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/storage"
)

// TestCRUDOperations 测试 CRUD 操作的 API 端点
func TestCRUDOperations(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_web_crud_test")
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

	// 创建管理实例
	manager := management.NewManager(testDb)

	// 创建 web 服务器
	server := NewServer(":8080", manager)

	// 生成测试 API 密钥
	testAPIKey, err := server.authManager.GenerateAPIKey(RoleAdmin)
	if err != nil {
		t.Fatalf("Failed to generate test API key: %v", err)
	}

	// 创建 gin 引擎并注册路由
	engine := gin.Default()

	// 注册 API 路由，与 server.Start() 中的路由注册一致
	api := engine.Group("/api")
	// 添加认证中间件
	api.Use(func(c *gin.Context) {
		// 设置测试 API 密钥
		c.Request.Header.Set("X-API-Key", testAPIKey)
		// 模拟认证通过
		c.Set("role", RoleAdmin)
		c.Next()
	})
	{
		api.POST("/tables", server.handleCreateTable)
		api.GET("/tables", server.handleGetTables)
		api.GET("/tables/:name", server.handleGetTable)
		api.PUT("/tables/:name", server.handleUpdateTable)
		api.DELETE("/tables/:name", server.handleDeleteTable)
		api.POST("/tables/:name/records", server.handleInsertRecord)
		api.GET("/tables/:name/records", server.handleGetRecords)
		api.PUT("/tables/:name/records", server.handleUpdateRecord)
		api.DELETE("/tables/:name/records", server.handleDeleteRecord)
		api.POST("/tables/:name/records/batch", server.handleBatchInsertRecords)
		api.GET("/tables/:name/records/range", server.handleSearchRange)
	}

	// 测试创建表
	t.Run("CreateTable", func(t *testing.T) {
		// 创建请求体
		reqBody := TableRequest{
			Name: "test_table",
			Fields: map[string]any{
				"id":    1,
				"name":  "test",
				"age":   25,
				"active": true,
				"score":  95.5,
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// 创建请求
		req, err := http.NewRequest("POST", "/api/tables", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		t.Logf("Table created successfully")
	})

	// 测试插入记录（包含类型转换）
	t.Run("InsertRecordWithTypeConversion", func(t *testing.T) {
		// 创建请求体，使用不同类型的值来测试类型转换
		reqBody := RecordRequest{
			Fields: map[string]any{
				"id":     "1",      // 字符串转整数
				"name":   "John",
				"age":    "30",     // 字符串转整数
				"active": "true",   // 字符串转布尔值
				"score":  "98.5",   // 字符串转浮点数
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// 创建请求
		req, err := http.NewRequest("POST", "/api/tables/test_table/records", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		t.Logf("Record inserted successfully with type conversion")
	})

	// 测试查询记录
	t.Run("GetRecords", func(t *testing.T) {
		// 创建请求
		req, err := http.NewRequest("GET", "/api/tables/test_table/records", nil)
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

		t.Logf("Records retrieved successfully")
	})

	// 测试更新记录
	t.Run("UpdateRecord", func(t *testing.T) {
		// 创建请求体
		reqBody := RecordRequest{
			Fields: map[string]any{
				"id":     1,
				"name":   "John Doe",
				"age":    "35",     // 字符串转整数
				"active": "false",  // 字符串转布尔值
				"score":  "99.0",   // 字符串转浮点数
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// 创建请求
		req, err := http.NewRequest("PUT", "/api/tables/test_table/records", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		t.Logf("Record updated successfully")
	})

	// 测试批量插入记录
	t.Run("BatchInsertRecords", func(t *testing.T) {
		// 创建请求体
		reqBody := BatchRecordRequest{
			Records: []map[string]any{
				{
					"id":     "2",      // 字符串转整数
					"name":   "Jane",
					"age":    "28",     // 字符串转整数
					"active": "true",   // 字符串转布尔值
					"score":  "92.5",   // 字符串转浮点数
				},
				{
					"id":     "3",      // 字符串转整数
					"name":   "Bob",
					"age":    "32",     // 字符串转整数
					"active": "false",  // 字符串转布尔值
					"score":  "88.0",   // 字符串转浮点数
				},
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// 创建请求
		req, err := http.NewRequest("POST", "/api/tables/test_table/records/batch", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		t.Logf("Batch records inserted successfully")
	})

	// 测试删除记录
	t.Run("DeleteRecord", func(t *testing.T) {
		// 创建请求体
		reqBody := RecordRequest{
			Fields: map[string]any{
				"id": "1",  // 字符串转整数
			},
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}

		// 创建请求
		req, err := http.NewRequest("DELETE", "/api/tables/test_table/records", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		// 创建响应记录器
		w := httptest.NewRecorder()

		// 处理请求
		engine.ServeHTTP(w, req)

		// 检查响应状态码
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status code 200, got %d", w.Code)
		}

		t.Logf("Record deleted successfully")
	})

	// 测试删除表
	t.Run("DeleteTable", func(t *testing.T) {
		// 创建请求
		req, err := http.NewRequest("DELETE", "/api/tables/test_table", nil)
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

		t.Logf("Table deleted successfully")
	})
}

// TestTypeConversion 专门测试类型转换功能
func TestTypeConversion(t *testing.T) {
	// 测试目录设置
	testDir := filepath.Join(os.TempDir(), "sfsdb_web_type_test")
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

	// 创建测试表
	table, err := engine.TableNew("test_type_table")
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// 设置表字段
	fields := map[string]any{
		"id":    1,
		"name":  "test",
		"age":   25,
		"active": true,
		"score":  95.5,
	}

	if err := table.SetFields(fields); err != nil {
		t.Fatalf("Failed to set table fields: %v", err)
	}

	// 测试类型转换
	t.Run("TestConvertFieldTypes", func(t *testing.T) {
		// 测试数据，包含需要转换的类型
		testData := map[string]any{
			"id":     "10",     // 字符串转整数
			"name":   "Test",
			"age":    "30",     // 字符串转整数
			"active": "true",   // 字符串转布尔值
			"score":  "98.5",   // 字符串转浮点数
		}

		// 转换字段类型
		convertedData, err := convertFieldTypes(table, testData)
		if err != nil {
			t.Fatalf("Failed to convert field types: %v", err)
		}

		// 验证转换结果
		if id, ok := convertedData["id"].(int); !ok || id != 10 {
			t.Errorf("Expected id to be int(10), got %T(%v)", convertedData["id"], convertedData["id"])
		}

		if age, ok := convertedData["age"].(int); !ok || age != 30 {
			t.Errorf("Expected age to be int(30), got %T(%v)", convertedData["age"], convertedData["age"])
		}

		if active, ok := convertedData["active"].(bool); !ok || active != true {
			t.Errorf("Expected active to be bool(true), got %T(%v)", convertedData["active"], convertedData["active"])
		}

		if score, ok := convertedData["score"].(float64); !ok || score != 98.5 {
			t.Errorf("Expected score to be float64(98.5), got %T(%v)", convertedData["score"], convertedData["score"])
		}

		t.Logf("Type conversion test passed")
	})
}
