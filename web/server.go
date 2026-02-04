package web

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/management"
)

// Server Web服务器结构
type Server struct {
	addr    string
	manager *management.Manager
}

// NewServer 创建新的Web服务器
func NewServer(addr string, manager *management.Manager) *Server {
	return &Server{
		addr:    addr,
		manager: manager,
	}
}

// Start 启动Web服务器
func (s *Server) Start() error {
	// 创建Gin实例
	router := gin.Default()

	// 注册API路由
	api := router.Group("/api")
	{
		api.GET("/status", s.handleStatus)
		api.GET("/system", s.handleSystem)
		api.GET("/index", s.handleIndexManagement)
		api.GET("/config", s.handleConfig)
		api.GET("/backup", s.handleBackup)
		api.GET("/stats", s.handleStats)
		api.GET("/monitor", s.handleMonitor)
		api.POST("/crud", s.handleTableCRUD)
	}

	// 注册静态文件路由
	router.Static("/static", "./web/static")

	// 注册首页路由
	router.GET("/", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.File("./web/static/index.html")
	})

	fmt.Printf("Web interface started at http://localhost%s\n", s.addr)
	return router.Run(s.addr)
}
