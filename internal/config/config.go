package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	IPVersion string `mapstructure:"ipversion"`
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

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

	// password 没有出现在 YAML 中，需要显式绑定以确保反序列化。
	if err := v.BindEnv("database.password"); err != nil {
		return nil, fmt.Errorf("绑定环境变量失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("无效的服务端口: %d", cfg.Server.Port)
	}

	return &cfg, nil
}
