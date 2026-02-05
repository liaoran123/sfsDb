package web

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// 通用错误处理函数
func (s *Server) sendError(c *gin.Context, statusCode int, errorType string, message string, details ...string) {
	errorResponse := gin.H{
		"error":   errorType,
		"message": message,
	}

	if len(details) > 0 && details[0] != "" {
		errorResponse["details"] = details[0]
	}

	c.JSON(statusCode, errorResponse)
}

// 发送成功响应
func (s *Server) sendSuccess(c *gin.Context, data map[string]interface{}) {
	c.JSON(http.StatusOK, data)
}

// handleIndex 处理首页请求
func (s *Server) handleIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.File("./web/static/index.html")
}

// handleStatus 处理状态请求
func (s *Server) handleStatus(c *gin.Context) {
	statusInfo, err := s.manager.GetStatus()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "status retrieval failed", "failed to get system status", err.Error())
		return
	}

	response := gin.H{
		"Memory": gin.H{
			"Alloc":      statusInfo.Memory.Alloc,
			"TotalAlloc": statusInfo.Memory.TotalAlloc,
			"Sys":        statusInfo.Memory.Sys,
			"NumGC":      statusInfo.Memory.NumGC,
		},
		"Storage": gin.H{
			"StoreType": statusInfo.Storage.StoreType,
		},
	}

	s.sendSuccess(c, response)
}

// handleSystem 处理系统信息请求
func (s *Server) handleSystem(c *gin.Context) {
	systemMgr := s.manager.SystemManager()
	systemInfo, err := systemMgr.GetAllSystemInfo()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "system info retrieval failed", "failed to get system information", err.Error())
		return
	}

	s.sendSuccess(c, systemInfo)
}

// handleIndexManagement 处理索引管理请求
func (s *Server) handleIndexManagement(c *gin.Context) {
	// 提取所有索引信息
	var indexes []gin.H

	// 从系统信息中获取索引信息
	systemMgr := s.manager.SystemManager()
	systemInfo, err := systemMgr.GetAllSystemInfo()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "index retrieval failed", "failed to get system information for index management", err.Error())
		return
	}

	// 从系统信息中提取索引信息
	if tableDetails, ok := systemInfo["tableDetails"].(map[uint8]map[string]interface{}); ok {
		for _, details := range tableDetails {
			tableName := ""
			if name, ok := details["name"].(string); ok {
				tableName = name
			}

			// 直接获取索引信息
			if indexInfos, ok := details["indexes"]; ok {
				// 尝试使用反射处理
				if reflect.TypeOf(indexInfos).Kind() == reflect.Slice {
					sliceValue := reflect.ValueOf(indexInfos)
					for i := 0; i < sliceValue.Len(); i++ {
						item := sliceValue.Index(i)
						// 尝试获取 Name 字段
						nameField := item.FieldByName("Name")
						if nameField.IsValid() && nameField.Kind() == reflect.String {
							indexName := nameField.String()
							indexes = append(indexes, gin.H{
								"name":   indexName,
								"fields": []string{},
								"type":   "Index",
								"table":  tableName,
							})
						}
					}
				}
			}
		}
	}

	// 直接从每个表获取索引信息（备用方案）
	if len(indexes) == 0 {
		tables, err := systemMgr.GetAllTables()
		if err == nil {
			for _, table := range tables {
				indexesInfo, err := systemMgr.GetTableIndexes(table.ID)
				if err == nil {
					for _, index := range indexesInfo {
						indexes = append(indexes, gin.H{
							"name":   index.Name,
							"fields": []string{},
							"type":   "Index",
							"table":  table.Name,
						})
					}
				}
			}
		}
	}

	// 如果没有找到索引，显示提示信息
	if len(indexes) == 0 {
		s.sendSuccess(c, gin.H{
			"indexes": []interface{}{},
			"message": "No indexes found in the database",
		})
		return
	}

	s.sendSuccess(c, gin.H{"indexes": indexes})
}

// handleConfig 处理配置请求
func (s *Server) handleConfig(c *gin.Context) {
	configMgr := s.manager.ConfigManager()
	config, err := configMgr.GetConfig()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "config retrieval failed", "failed to get configuration", err.Error())
		return
	}

	// 转换为 map[string]interface{} 类型
	configData := map[string]interface{}{
		"StoreType": config.StoreType,
		"Options":   config.Options,
	}

	s.sendSuccess(c, configData)
}

// handleBackup 处理备份请求
func (s *Server) handleBackup(c *gin.Context) {
	backupMgr := s.manager.BackupManager()

	// 获取操作类型
	op := c.Query("op")

	// 获取备份路径
	path := c.Query("path")
	if path == "" {
		path = "./backups"
	}

	// 根据操作类型处理
	switch op {
	case "restore":
		// 恢复备份
		file := c.Query("file")
		if file == "" {
			s.sendError(c, http.StatusBadRequest, "missing parameter", "file parameter is required for restore operation")
			return
		}

		err := backupMgr.Restore(file)
		if err != nil {
			s.sendError(c, http.StatusInternalServerError, "restore failed", "failed to restore backup", err.Error())
			return
		}

		s.sendSuccess(c, gin.H{"message": "Backup restored successfully"})
	case "create":
		// 创建备份
		backupPath, err := backupMgr.Backup(path)
		if err != nil {
			s.sendError(c, http.StatusInternalServerError, "backup failed", "failed to create backup", err.Error())
			return
		}

		s.sendSuccess(c, gin.H{
			"backupPath": backupPath,
			"message":    "Backup created successfully",
		})
	default:
		// 默认返回备份管理页面的初始状态
		s.sendSuccess(c, gin.H{
			"message":     "Backup management ready",
			"defaultPath": path,
		})
	}
}

// handleStats 处理统计请求
func (s *Server) handleStats(c *gin.Context) {
	statsMgr := s.manager.StatsManager()
	queryStats, err := statsMgr.GetQueryStats()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "stats retrieval failed", "failed to get query statistics", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{"queryStats": queryStats})
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

	// 验证操作类型
	validOperations := map[string]bool{
		"insert": true,
		"search": true,
		"update": true,
		"delete": true,
	}

	if !validOperations[operation] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid operation",
			"message": "operation must be one of: insert, search, update, delete",
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
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
	}
}

// handleInsert 处理插入操作
func (s *Server) handleInsert(c *gin.Context, table interface{}, fields map[string]interface{}) {
	// 类型断言检查
	type Insertable interface {
		Insert(fields *map[string]any, batchs ...interface{}) (int, error)
	}

	t, ok := table.(Insertable)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "invalid table type",
			"message": "table does not implement required Insert method",
		})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	id, err := t.Insert(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "insert operation failed",
			"details": err.Error(),
		})
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
	// 类型断言检查
	type Searchable interface {
		Search(fields *map[string]any, ops ...interface{}) (interface{}, error)
	}

	t, ok := table.(Searchable)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "invalid table type",
			"message": "table does not implement required Search method",
		})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	iter, err := t.Search(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "search operation failed",
			"details": err.Error(),
		})
		return
	}

	// 遍历结果
	var results []map[string]interface{}
	type Iterator interface {
		Next() bool
		Get() (map[string]interface{}, error)
		Release()
	}

	iterInterface, ok := iter.(Iterator)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "invalid iterator type",
			"message": "search returned invalid iterator",
		})
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
	// 类型断言检查
	type Updatable interface {
		Update(fields *map[string]any, batchs ...interface{}) error
	}

	t, ok := table.(Updatable)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "invalid table type",
			"message": "table does not implement required Update method",
		})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	err := t.Update(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "update operation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "record updated successfully",
	})
}

// handleDelete 处理删除操作
func (s *Server) handleDelete(c *gin.Context, table interface{}, fields map[string]interface{}) {
	// 类型断言检查
	type Deletable interface {
		Delete(fields *map[string]any, batchs ...interface{}) error
	}

	t, ok := table.(Deletable)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "invalid table type",
			"message": "table does not implement required Delete method",
		})
		return
	}

	// 转换为*map[string]any
	fieldsMap := fields
	err := t.Delete(&fieldsMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "delete operation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "record deleted successfully",
	})
}
