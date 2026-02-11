package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/management"
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
		transactionStats := monitorMgr.GetTransactionStats()

		response := gin.H{
			"keyChangeStats":    keyChangeStats,
			"indexStats":        indexStats,
			"transactionStats":  transactionStats,
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
