package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/management/importexport"
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

/*
// handleIndex 处理首页请求
func (s *Server) handleIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.File("./web/static/index.html")
}
*/
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
		"Scenarios": config.Scenarios,
	}

	s.sendSuccess(c, configData)
}

// handleSetConfig 处理设置配置请求
func (s *Server) handleSetConfig(c *gin.Context) {
	// 解析请求参数
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	configMgr := s.manager.ConfigManager()
	err := configMgr.SetConfig(req.Key, req.Value)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "config update failed", "Failed to update configuration", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{"message": "Configuration updated successfully"})
}

// handleSetScenarioConfig 处理设置场景配置请求
func (s *Server) handleSetScenarioConfig(c *gin.Context) {
	// 解析请求参数
	var req struct {
		Scenario string `json:"scenario"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	configMgr := s.manager.ConfigManager()
	err := configMgr.SetScenarioConfig(req.Scenario)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "scenario config update failed", "Failed to update scenario configuration", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{"message": "Scenario configuration updated successfully"})
}

// handleResetConfig 处理重置配置请求
func (s *Server) handleResetConfig(c *gin.Context) {
	configMgr := s.manager.ConfigManager()
	err := configMgr.ResetConfig()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "config reset failed", "Failed to reset configuration", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{"message": "Configuration reset successfully"})
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

// 监控器实例
var monitor *management.Monitor

// handleMonitor 处理监控请求
func (s *Server) handleMonitor(c *gin.Context) {
	// 获取操作类型
	op := c.Query("op")

	switch op {
	case "start":
		s.handleMonitorStart(c)
	case "stop":
		s.handleMonitorStop(c)
	case "status":
		s.handleMonitorStatus(c)
	case "config":
		s.handleMonitorConfig(c)
	default:
		// 获取监控统计信息
		monitorMgr := s.manager.MonitorManager()
		keyChangeStats := monitorMgr.GetKeyChangeStats()
		indexStats := monitorMgr.GetIndexStats()
		response := gin.H{
			"keyChangeStats":    keyChangeStats,
			"indexStats":        indexStats,
			"monitorRunning":    monitor != nil && s.isMonitorRunning(),
		}

		c.JSON(http.StatusOK, response)
	}
}

// handleMonitorStart 处理启动监控请求
func (s *Server) handleMonitorStart(c *gin.Context) {
	// 解析请求参数
	intervalStr := c.DefaultQuery("interval", "5s")
	memoryThresholdStr := c.DefaultQuery("memoryThreshold", "100")
	gcThresholdStr := c.DefaultQuery("gcThreshold", "100")

	// 解析监控间隔
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		interval = 5 * time.Second
	}

	// 解析内存阈值
	memoryThreshold, err := strconv.ParseFloat(memoryThresholdStr, 64)
	if err != nil {
		memoryThreshold = 100.0
	}

	// 解析GC阈值
	gcThreshold, err := strconv.Atoi(gcThresholdStr)
	if err != nil {
		gcThreshold = 100
	}

	// 创建监控阈值
	thresholds := management.Thresholds{
		MemoryUsage: memoryThreshold,
		GCCount:     gcThreshold,
	}

	// 创建并启动监控器
	monitor = management.NewMonitor(s.manager, interval, thresholds)
	err = monitor.Start()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "monitor start failed", "failed to start monitor", err.Error())
		return
	}

	s.sendSuccess(c, gin.H{
		"message":  "Monitor started successfully",
		"interval": interval.String(),
		"thresholds": gin.H{
			"memoryUsage": memoryThreshold,
			"gcCount":     gcThreshold,
		},
	})
}

// handleMonitorStop 处理停止监控请求
func (s *Server) handleMonitorStop(c *gin.Context) {
	if monitor == nil {
		s.sendError(c, http.StatusBadRequest, "monitor not running", "monitor is not started")
		return
	}

	monitor.Stop()
	monitor = nil

	s.sendSuccess(c, gin.H{"message": "Monitor stopped successfully"})
}

// handleMonitorStatus 处理获取监控状态请求
func (s *Server) handleMonitorStatus(c *gin.Context) {
	if monitor == nil {
		s.sendSuccess(c, gin.H{
			"running": false,
			"message": "Monitor is not started",
		})
		return
	}

	// 获取当前状态
	statusInfo, err := monitor.GetCurrentStatus()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "status retrieval failed", "failed to get monitor status", err.Error())
		return
	}

	// 计算内存使用
	memoryUsage := float64(statusInfo.Memory.Alloc) / 1024 / 1024

	s.sendSuccess(c, gin.H{
		"running": true,
		"currentStatus": gin.H{
			"memoryUsage": memoryUsage,
			"gcCount":     statusInfo.Memory.NumGC,
			"storeType":   statusInfo.Storage.StoreType,
		},
	})
}

// handleMonitorConfig 处理配置监控请求
func (s *Server) handleMonitorConfig(c *gin.Context) {
	// 返回默认配置
	s.sendSuccess(c, gin.H{
		"defaultInterval":        "5s",
		"defaultMemoryThreshold": 100.0,
		"defaultGCThreshold":     100,
	})
}

// isMonitorRunning 检查监控器是否运行
func (s *Server) isMonitorRunning() bool {
	// 这里简化处理，实际应该检查 monitor.running 字段
	// 但由于 running 是私有字段，这里返回 true 表示监控器存在
	return monitor != nil
}

// handleGenerateAPIKey 处理生成API密钥请求
func (s *Server) handleGenerateAPIKey(c *gin.Context) {
	// 解析请求参数
	var req struct {
		Role string `json:"role"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	// 验证角色
	var role Role
	switch req.Role {
	case string(RoleAdmin):
		role = RoleAdmin
	case string(RoleOperator):
		role = RoleOperator
	case string(RoleViewer):
		role = RoleViewer
	default:
		s.sendError(c, http.StatusBadRequest, "invalid role", "Invalid role specified")
		return
	}

	// 生成API密钥
	key, err := s.authManager.GenerateAPIKey(role)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "key generation failed", "Failed to generate API key", err.Error())
		return
	}

	// 返回API密钥
	s.sendSuccess(c, gin.H{
		"api_key": key,
		"role":    string(role),
		"message": "API key generated successfully",
	})
}

// handleRevokeAPIKey 处理撤销API密钥请求
func (s *Server) handleRevokeAPIKey(c *gin.Context) {
	// 解析请求参数
	var req struct {
		Key string `json:"key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	// 撤销API密钥
	s.authManager.RevokeAPIKey(req.Key)

	// 返回成功响应
	s.sendSuccess(c, gin.H{
		"message": "API key revoked successfully",
	})
}

// handleDiagnostic 处理诊断包导出请求
func (s *Server) handleDiagnostic(c *gin.Context) {
	// 获取系统状态
	statusInfo, err := s.manager.GetStatus()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "status retrieval failed", "failed to get system status", err.Error())
		return
	}

	// 获取系统信息
	systemMgr := s.manager.SystemManager()
	systemInfo, err := systemMgr.GetAllSystemInfo()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "system info retrieval failed", "failed to get system information", err.Error())
		return
	}

	// 获取配置信息
	configMgr := s.manager.ConfigManager()
	config, err := configMgr.GetConfig()
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "config retrieval failed", "failed to get configuration", err.Error())
		return
	}

	// 获取监控统计信息
	monitorMgr := s.manager.MonitorManager()
	keyChangeStats := monitorMgr.GetKeyChangeStats()
	indexStats := monitorMgr.GetIndexStats()
	// 构建诊断数据
	diagnosticData := gin.H{
		"timestamp": time.Now().Format(time.RFC3339),
		"status": statusInfo,
		"system": systemInfo,
		"config": map[string]interface{}{
			"StoreType": config.StoreType,
			"Options":   config.Options,
		},
		"monitor": gin.H{
			"keyChangeStats":   keyChangeStats,
			"indexStats":       indexStats,
			"monitorRunning":   monitor != nil && s.isMonitorRunning(),
		},
		"version": "1.0.0", // 诊断包版本
	}

	// 转换为JSON
	diagnosticJSON, err := json.MarshalIndent(diagnosticData, "", "  ")
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "json marshal failed", "failed to marshal diagnostic data", err.Error())
		return
	}

	// 设置响应头，使浏览器下载文件
	filename := fmt.Sprintf("diagnostic_%s.json", time.Now().Format("20060102_150405"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/json")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	c.Header("Content-Length", strconv.Itoa(len(diagnosticJSON)))

	// 发送响应
	c.Data(http.StatusOK, "application/json", diagnosticJSON)
}

// handleExportTable 处理导出表数据请求
func (s *Server) handleExportTable(c *gin.Context) {
	// 解析请求参数
	var req struct {
		TableName string `json:"tableName"`
		Format    string `json:"format"`
		Fields    []string `json:"fields"`
		WhereClause string `json:"whereClause"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	// 验证参数
	if req.TableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing parameter", "tableName parameter is required")
		return
	}

	if req.Format == "" {
		req.Format = "json" // 默认格式
	}

	// 获取导入导出管理器
	importExportMgr := s.manager.ImportExportManager()

	// 构建导出选项
	options := importexport.ExportOptions{
		Format:      req.Format,
		Fields:      req.Fields,
		WhereClause: req.WhereClause,
	}

	// 执行导出
	exportPath, err := importExportMgr.ExportTable(req.TableName, options)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "export failed", "Failed to export table data", err.Error())
		return
	}

	// 发送成功响应
	s.sendSuccess(c, gin.H{
		"message":     "Table exported successfully",
		"exportPath": exportPath,
	})
}

// handleImportTable 处理导入表数据请求
func (s *Server) handleImportTable(c *gin.Context) {
	// 解析请求参数
	var req struct {
		TableName    string `json:"tableName"`
		ImportPath   string `json:"importPath"`
		Format       string `json:"format"`
		IgnoreErrors bool   `json:"ignoreErrors"`
		Truncate     bool   `json:"truncate"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		s.sendError(c, http.StatusBadRequest, "invalid request", "Invalid request format", err.Error())
		return
	}

	// 验证参数
	if req.TableName == "" {
		s.sendError(c, http.StatusBadRequest, "missing parameter", "tableName parameter is required")
		return
	}

	if req.ImportPath == "" {
		s.sendError(c, http.StatusBadRequest, "missing parameter", "importPath parameter is required")
		return
	}

	// 获取导入导出管理器
	importExportMgr := s.manager.ImportExportManager()

	// 构建导入选项
	options := importexport.ImportOptions{
		Format:      req.Format,
		IgnoreErrors: req.IgnoreErrors,
		Truncate:     req.Truncate,
	}

	// 执行导入
	count, err := importExportMgr.ImportTable(req.TableName, req.ImportPath, options)
	if err != nil {
		s.sendError(c, http.StatusInternalServerError, "import failed", "Failed to import table data", err.Error())
		return
	}

	// 发送成功响应
	s.sendSuccess(c, gin.H{
		"message":    "Table imported successfully",
		"importedCount": count,
	})
}
