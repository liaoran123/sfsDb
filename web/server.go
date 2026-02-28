package web

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/liaoran123/sfsDb/management"
)

// Server Web服务器结构
type Server struct {
	addr           string
	manager        *management.Manager
	authManager    *AuthManager
	metricsManager *MetricsManager
	tlsConfig      TLSConfig
}

// AuthManager 获取认证管理器
func (s *Server) AuthManager() *AuthManager {
	return s.authManager
}

// isPortAvailable 检查端口是否可用
func isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// findAvailablePort 查找可用的端口
func findAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort
}

// openBrowser 自动打开浏览器
func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "windows":
		err = exec.Command("cmd", "/c", "start", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		fmt.Printf("Unsupported OS: %s\n", runtime.GOOS)
		return
	}

	if err != nil {
		fmt.Printf("Error opening browser: %v\n", err)
	}
}

// NewServer 创建新的Web服务器
func NewServer(addr string, manager *management.Manager) *Server {
	return &Server{
		addr:           addr,
		manager:        manager,
		authManager:    NewAuthManager(),
		metricsManager: NewMetricsManager(),
		tlsConfig:      DefaultTLSConfig(),
	}
}

// NewServerWithTLS 创建带TLS配置的Web服务器
func NewServerWithTLS(addr string, manager *management.Manager, tlsConfig TLSConfig) *Server {
	return &Server{
		addr:           addr,
		manager:        manager,
		authManager:    NewAuthManager(),
		metricsManager: NewMetricsManager(),
		tlsConfig:      tlsConfig,
	}
}

// Start 启动Web服务器
func (s *Server) Start() error {
	// 端口自适应：检查并找到可用端口
	var port int
	if s.addr != "" {
		// 解析地址，提取端口号
		_, portStr, err := net.SplitHostPort(s.addr)
		if err != nil {
			// 如果地址格式不正确，使用默认端口8080
			port = 8080
		} else {
			port, err = strconv.Atoi(portStr)
			if err != nil {
				port = 8080
			}
		}
	} else {
		// 如果地址为空，使用默认端口8080
		port = 8080
	}

	// 查找可用端口
	availablePort := findAvailablePort(port)
	if availablePort != port {
		fmt.Printf("Port %d is in use, switching to port %d\n", port, availablePort)
		port = availablePort
		// 更新服务器地址
		s.addr = fmt.Sprintf(":%d", port)
	}

	// 创建Gin实例
	router := gin.Default()

	// 添加响应时间中间件
	router.Use(ResponseTimeMiddleware(s.metricsManager))

	// 注册认证API路由（无需认证）
	auth := router.Group("/api/auth")
	{
		auth.POST("/generate", s.handleGenerateAPIKey)
	}

	// 注册需要认证的API路由
	api := router.Group("/api")
	api.Use(AuthMiddleware(s.authManager))
	{
		// 读取权限API
		api.GET("/status", PermissionMiddleware(s.authManager, PermissionRead), s.handleStatus)
		api.GET("/system", PermissionMiddleware(s.authManager, PermissionRead), s.handleSystem)
		api.GET("/config", PermissionMiddleware(s.authManager, PermissionRead), s.handleConfig)
		api.GET("/monitor", PermissionMiddleware(s.authManager, PermissionRead), s.handleMonitor)

		// 写入权限API
		api.POST("/config", PermissionMiddleware(s.authManager, PermissionWrite), s.handleSetConfig)
		api.POST("/config/scenario", PermissionMiddleware(s.authManager, PermissionWrite), s.handleSetScenarioConfig)
		api.POST("/config/reset", PermissionMiddleware(s.authManager, PermissionWrite), s.handleResetConfig)

		// 写入权限API
		api.GET("/backup", PermissionMiddleware(s.authManager, PermissionWrite), s.handleBackup)

		// 管理权限API
		api.POST("/auth/revoke", PermissionMiddleware(s.authManager, PermissionAdmin), s.handleRevokeAPIKey)

		// 诊断API
		api.GET("/diagnostic", PermissionMiddleware(s.authManager, PermissionRead), s.handleDiagnostic)

		// 导入导出API
		api.POST("/tables/export", PermissionMiddleware(s.authManager, PermissionRead), s.handleExportTable)
		api.POST("/tables/import", PermissionMiddleware(s.authManager, PermissionWrite), s.handleImportTable)

		// 监控API
		api.GET("/metrics", PermissionMiddleware(s.authManager, PermissionRead), s.handleGetMetrics)
		api.GET("/alerts", PermissionMiddleware(s.authManager, PermissionRead), s.handleGetAlerts)
		api.POST("/thresholds", PermissionMiddleware(s.authManager, PermissionAdmin), s.handleSetThresholds)
		api.GET("/thresholds", PermissionMiddleware(s.authManager, PermissionRead), s.handleGetThresholds)
		api.POST("/metrics/collect", PermissionMiddleware(s.authManager, PermissionRead), s.handleCollectMetrics)
	}

	// 注册静态文件路由（使用嵌入的文件系统）
	router.StaticFS("/static", getStaticFS())

	// 注册首页路由（使用嵌入的文件系统）
	router.GET("/", func(c *gin.Context) {
		// 从嵌入的文件系统中读取index.html
		content, err := staticFiles.ReadFile("static/index.html")
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to read index.html"})
			return
		}
		c.Data(200, "text/html; charset=utf-8", content)
	})

	// 添加404处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Page not found"})
	})

	// 检查TLS配置
	if s.tlsConfig.Enable {
		// 加载TLS配置
		tlsConfig, err := LoadTLSConfig(s.tlsConfig)
		if err != nil {
			// TLS配置加载失败，回退到HTTP
			fmt.Printf("TLS configuration error: %v, falling back to HTTP\n", err)
			fmt.Printf("Web interface started at http://localhost%s\n", s.addr)

			// 自动打开浏览器
			go openBrowser(fmt.Sprintf("http://localhost%s", s.addr))

			return router.Run(s.addr)
		}

		// 创建HTTPS服务器
		server := &http.Server{
			Addr:      s.addr,
			Handler:   router,
			TLSConfig: tlsConfig,
		}

		fmt.Printf("Web interface started at https://localhost%s\n", s.addr)

		// 自动打开浏览器
		go openBrowser(fmt.Sprintf("https://localhost%s", s.addr))

		return server.ListenAndServeTLS(s.tlsConfig.CertFile, s.tlsConfig.KeyFile)
	} else {
		// 启动HTTP服务器
		fmt.Printf("Web interface started at http://localhost%s\n", s.addr)

		// 自动打开浏览器
		go openBrowser(fmt.Sprintf("http://localhost%s", s.addr))

		return router.Run(s.addr)
	}
}
