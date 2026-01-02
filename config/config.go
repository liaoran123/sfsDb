package config

import (
	"os"
	"path/filepath"

	"time"

	"gopkg.in/yaml.v3"
)

var Cfg *Config

func init() {
	// 加载配置文件，尝试多个可能的路径
	var err error

	// 尝试执行文件所在目录的config.yaml
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		configPath := filepath.Join(execDir, "config.yaml")
		Cfg, err = LoadConfig(configPath)
	}

	if err != nil {
		// 5. 尝试上级目录的config.yaml
		Cfg, err = LoadConfig("../config.yaml")
	}
	if err != nil {
		// 6. 尝试上上级目录的config.yaml
		Cfg, err = LoadConfig("../../config.yaml")
	}

	if err != nil {
		// 所有尝试都失败，使用默认配置
		os.Stderr.WriteString("警告：所有配置文件路径都无法加载，使用默认值\n")
		Cfg = &Config{
			Web: WebConfig{
				Port: 9981,
			},
			Db: DbConfig{
				DbPath:   "db",
				Password: "",
			},
			IterCache: IterCacheConfig{
				Timeout: time.Minute * 5,
				Max:     10000,
			},
		}
	}
}

// Config 配置结构体
type Config struct {
	Web       WebConfig       `yaml:"web"`
	Db        DbConfig        `yaml:"db"`
	IterCache IterCacheConfig `yaml:"iter_cache"`
}

// WebConfig Web配置
type WebConfig struct {
	Port int `yaml:"port"`
}

// DbConfig 数据库配置
type DbConfig struct {
	DbPath   string `yaml:"dbpath"`
	Password string `yaml:"password"`
}

// IterCacheConfig 数据迭代器缓存配置
type IterCacheConfig struct {
	Timeout time.Duration `yaml:"timeout"`
	Max     int           `yaml:"max"`
}

// LoadConfig 从yaml文件加载配置
func LoadConfig(filePath string) (*Config, error) {
	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 解析yaml
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
