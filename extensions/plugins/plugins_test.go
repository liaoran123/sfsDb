package plugins

import (
	"testing"
)

// MockPlugin 模拟插件实现
type MockPlugin struct {
	name        string
	version     string
	description string
	running     bool
}

func NewMockPlugin(name, version, description string) *MockPlugin {
	return &MockPlugin{
		name:        name,
		version:     version,
		description: description,
		running:     false,
	}
}

func (p *MockPlugin) Name() string {
	return p.name
}

func (p *MockPlugin) Version() string {
	return p.version
}

func (p *MockPlugin) Description() string {
	return p.description
}

func (p *MockPlugin) Init() error {
	return nil
}

func (p *MockPlugin) Start() error {
	p.running = true
	return nil
}

func (p *MockPlugin) Stop() error {
	p.running = false
	return nil
}

func (p *MockPlugin) IsRunning() bool {
	return p.running
}

func TestPluginManager(t *testing.T) {
	// 创建插件配置
	config := PluginConfig{
		Enabled: true,
		Dir:     "plugins",
		Plugins: []string{},
	}

	// 创建插件管理器
	pm := NewPluginManager(config)

	// 创建测试插件（用于后续扩展测试）
	_ = NewMockPlugin("plugin1", "1.0.0", "Test plugin 1")
	_ = NewMockPlugin("plugin2", "2.0.0", "Test plugin 2")

	// 注意：由于Go的插件系统需要编译为.so文件，这里我们模拟插件加载
	// 在实际测试中，需要编译插件并通过文件路径加载

	// 测试插件列表
	t.Run("ListPlugins", func(t *testing.T) {
		plugins := pm.ListPlugins()
		if len(plugins) != 0 {
			t.Errorf("Expected empty plugin list, got %d plugins", len(plugins))
		}

		// 注意：在实际实现中，这里应该测试插件的加载和管理
		// 由于插件加载需要.so文件，这里我们跳过实际加载测试
	})

	// 测试插件状态
	t.Run("PluginState", func(t *testing.T) {
		// 测试插件启动和停止
		plugin := NewMockPlugin("test", "1.0.0", "Test plugin")

		// 测试启动
		err := plugin.Start()
		if err != nil {
			t.Errorf("Expected successful start, got error: %v", err)
		}
		if !plugin.IsRunning() {
			t.Errorf("Expected plugin to be running after start")
		}

		// 测试停止
		err = plugin.Stop()
		if err != nil {
			t.Errorf("Expected successful stop, got error: %v", err)
		}
		if plugin.IsRunning() {
			t.Errorf("Expected plugin to not be running after stop")
		}
	})
}
