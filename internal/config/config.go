package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config 汇总应用运行时使用的全部配置。
// 当前学习阶段只包含 TCP 服务器配置，后续模块可按需继续扩展。
type Config struct {
	Server ServerConfig `mapstructure:"server"`
}

// ServerConfig 描述 TCP 服务器的监听参数和资源限制。
type ServerConfig struct {
	IPVersion     string `mapstructure:"ipversion"`     // IPVersion 指定网络类型，例如 tcp、tcp4 或 tcp6。
	Host          string `mapstructure:"host"`          // Host 是服务器绑定的 IP 地址。
	Port          int    `mapstructure:"port"`          // Port 是服务器监听端口。
	MaxConn       int    `mapstructure:"maxconn"`       // MaxConn 是允许同时建立的最大连接数，供后续连接管理模块使用。
	MaxPacketSize uint32 `mapstructure:"maxpacketsize"` // MaxPacketSize 是单个数据包的最大字节数，供后续封包模块使用。
}

// Load 从 config/config.yaml 读取配置，并允许 APP_ 前缀的环境变量覆盖同名配置项。
// 例如 server.port 可以通过 APP_SERVER_PORT 覆盖。
func Load() (*Config, error) {
	// .env 仅作为本地开发的可选配置。
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("加载 .env 失败: %w", err)
	}

	v := viper.New()
	v.SetConfigFile("config/config.yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// server.port -> APP_SERVER_PORT
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("无效的服务端口: %d", cfg.Server.Port)
	}

	return &cfg, nil
}
