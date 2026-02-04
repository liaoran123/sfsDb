package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// handleIndex 处理首页请求
func (s *Server) handleIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.File("./web/static/index.html")
}

// handleStatus 处理状态请求
func (s *Server) handleStatus(c *gin.Context) {
	statusInfo, err := s.manager.GetStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"memory": gin.H{
			"alloc":      statusInfo.Memory.Alloc,
			"totalAlloc": statusInfo.Memory.TotalAlloc,
			"sys":        statusInfo.Memory.Sys,
			"numGC":      statusInfo.Memory.NumGC,
		},
		"storage": gin.H{
			"storeType": statusInfo.Storage.StoreType,
		},
	}

	c.JSON(http.StatusOK, response)
}

// handleSystem 处理系统信息请求
func (s *Server) handleSystem(c *gin.Context) {
	systemMgr := s.manager.SystemManager()
	systemInfo, err := systemMgr.GetAllSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, systemInfo)
}

// handleIndexManagement 处理索引管理请求
func (s *Server) handleIndexManagement(c *gin.Context) {
	// 提取所有索引信息
	var indexes []gin.H

	// 手动添加一些测试索引信息
	indexes = append(indexes, gin.H{
		"name":   "primary",
		"fields": []string{"id"},
		"type":   "Primary Key",
		"table":  "users",
	})

	indexes = append(indexes, gin.H{
		"name":   "primary",
		"fields": []string{"id"},
		"type":   "Primary Key",
		"table":  "products",
	})

	response := gin.H{
		"indexes": indexes,
	}

	c.JSON(http.StatusOK, response)
}

// handleConfig 处理配置请求
func (s *Server) handleConfig(c *gin.Context) {
	configMgr := s.manager.ConfigManager()
	config, err := configMgr.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// handleBackup 处理备份请求
func (s *Server) handleBackup(c *gin.Context) {
	backupMgr := s.manager.BackupManager()
	backupPath, err := backupMgr.Backup("./backups")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"backupPath": backupPath,
		"message":    "Backup created successfully",
	}

	c.JSON(http.StatusOK, response)
}

// handleStats 处理统计请求
func (s *Server) handleStats(c *gin.Context) {
	statsMgr := s.manager.StatsManager()
	queryStats, err := statsMgr.GetQueryStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"queryStats": queryStats,
	}

	c.JSON(http.StatusOK, response)
}

// handleMonitor 处理监控请求
func (s *Server) handleMonitor(c *gin.Context) {
	monitorMgr := s.manager.MonitorManager()
	keyChangeStats := monitorMgr.GetKeyChangeStats()

	response := gin.H{
		"keyChangeStats": keyChangeStats,
	}

	c.JSON(http.StatusOK, response)
}

// handleTableCRUD 处理TableCRUD操作请求
func (s *Server) handleTableCRUD(c *gin.Context) {
	// 获取操作类型
	operation := c.Query("op")
	tableName := c.Query("table")

	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table name is required"})
		return
	}

	// 获取表实例
	table := s.manager.GetTable(tableName)
	if table == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not found or not initialized"})
		return
	}

	// 解析请求体
	var requestData struct {
		Fields map[string]interface{} `json:"fields"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	switch operation {
	case "insert":
		s.handleInsert(c, table, requestData.Fields)
	case "search":
		s.handleSearch(c, table, requestData.Fields)
	case "update":
		s.handleUpdate(c, table, requestData.Fields)
	case "delete":
		s.handleDelete(c, table, requestData.Fields)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown operation"})
	}
}

// handleInsert 处理插入操作
func (s *Server) handleInsert(c *gin.Context, table interface{}, fields map[string]interface{}) {
	// 假设table是*engine.Table类型
	t, ok := table.(interface {
		Insert(fields *map[string]any, batchs ...interface{}) (int, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid table type"})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	id, err := t.Insert(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"id":      id,
		"message": "record inserted successfully",
	})
}

// handleSearch 处理查询操作
func (s *Server) handleSearch(c *gin.Context, table interface{}, fields map[string]interface{}) {
	// 假设table是*engine.Table类型
	t, ok := table.(interface {
		Search(fields *map[string]any, ops ...interface{}) (interface{}, error)
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid table type"})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	iter, err := t.Search(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 遍历结果
	var results []map[string]interface{}
	iterInterface, ok := iter.(interface {
		Next() bool
		Get() (map[string]interface{}, error)
		Release()
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid iterator type"})
		return
	}

	for iterInterface.Next() {
		record, err := iterInterface.Get()
		if err != nil {
			continue
		}
		results = append(results, record)
	}
	iterInterface.Release()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"results": results,
		"count":   len(results),
	})
}

// handleUpdate 处理更新操作
func (s *Server) handleUpdate(c *gin.Context, table interface{}, fields map[string]interface{}) {
	// 假设table是*engine.Table类型
	t, ok := table.(interface {
		Update(fields *map[string]any, batchs ...interface{}) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid table type"})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	err := t.Update(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "record updated successfully",
	})
}

// handleDelete 处理删除操作
func (s *Server) handleDelete(c *gin.Context, table interface{}, fields map[string]interface{}) {
	// 假设table是*engine.Table类型
	t, ok := table.(interface {
		Delete(fields *map[string]any, batchs ...interface{}) error
	})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid table type"})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	err := t.Delete(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "record deleted successfully",
	})
}
