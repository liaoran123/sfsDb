package cmd

import (
	"fmt"
	"strconv"

	"github.com/liaoran123/sfsDb/management"
	"github.com/liaoran123/sfsDb/web"
	"github.com/spf13/cobra"
)

var (
	webPort   string
	webEnable bool
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Manage web interface",
	Long:  `Manage web interface settings and start/stop the web server.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化存储
		err := initStore()
		if err != nil {
			fmt.Printf("Failed to initialize store: %v\n", err)
			return
		}
		defer closeStore()

		// 获取或设置 web 配置
		if enable, _ := cmd.Flags().GetBool("enable"); enable {
			// 启用 web 功能
			setWebConfig(manager, true, webPort)
			fmt.Println("Web interface enabled")
		} else if disable, _ := cmd.Flags().GetBool("disable"); disable {
			// 禁用 web 功能
			setWebConfig(manager, false, "")
			fmt.Println("Web interface disabled")
		} else if start, _ := cmd.Flags().GetBool("start"); start {
			// 启动 web 服务器
			startWebServer(manager, webPort)
		} else {
			// 显示当前 web 配置
			showWebConfig(manager)
		}
	},
}

func init() {
	rootCmd.AddCommand(webCmd)

	// 添加命令行标志
	webCmd.Flags().BoolP("enable", "e", false, "Enable web interface")
	webCmd.Flags().BoolP("disable", "d", false, "Disable web interface")
	webCmd.Flags().BoolP("start", "s", false, "Start web server")
	webCmd.Flags().StringVarP(&webPort, "port", "p", ":8083", "Web server port")
}

// 设置 web 配置
func setWebConfig(manager *management.Manager, enable bool, port string) error {
	configMgr := manager.ConfigManager()

	// 设置 web 启用状态
	err := configMgr.SetConfig("web_enable", strconv.FormatBool(enable))
	if err != nil {
		return err
	}

	// 设置 web 端口
	if enable && port != "" {
		err = configMgr.SetConfig("web_port", port)
		if err != nil {
			return err
		}
	}

	return nil
}

// 显示当前 web 配置
func showWebConfig(manager *management.Manager) {
	// 检查 web 配置
	webEnable := "false"
	if value, err := manager.Store().Get([]byte("config:web_enable")); err == nil {
		webEnable = string(value)
	}

	webPort := ":8083"
	if value, err := manager.Store().Get([]byte("config:web_port")); err == nil {
		webPort = string(value)
	}

	fmt.Println("Web Interface Configuration:")
	fmt.Printf("Enabled: %s\n", webEnable)
	fmt.Printf("Port: %s\n", webPort)
	fmt.Printf("Address: http://localhost%s\n", webPort)
}

// 启动 web 服务器
func startWebServer(manager *management.Manager, port string) {
	// 检查 web 是否启用
	webEnable := false
	if value, err := manager.Store().Get([]byte("config:web_enable")); err == nil {
		webEnable, _ = strconv.ParseBool(string(value))
	}

	if !webEnable {
		fmt.Println("Web interface is disabled. Use --enable flag to enable it.")
		return
	}

	// 使用配置的端口或命令行指定的端口
	if port == ":8083" {
		if value, err := manager.Store().Get([]byte("config:web_port")); err == nil {
			port = string(value)
		}
	}

	// 创建并启动 web 服务器
	server := web.NewServer(port, manager)
	if err := server.Start(); err != nil {
		fmt.Printf("Failed to start web server: %v\n", err)
	}
}
