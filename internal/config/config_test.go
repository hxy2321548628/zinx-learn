package config

import "testing"

func TestConfigValidateResourceLimits(t *testing.T) {
	t.Parallel()

	valid := Config{Server: ServerConfig{
		IPVersion:      "tcp",
		Port:           7777,
		MaxConn:        1,
		MaxPacketSize:  1,
		WorkerPoolSize: 1,
	}}

	// 每个用例只破坏一项配置，便于定位是哪条资源限制没有被校验。
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "max connections", mutate: func(cfg *Config) { cfg.Server.MaxConn = 0 }},
		{name: "max packet size", mutate: func(cfg *Config) { cfg.Server.MaxPacketSize = 0 }},
		{name: "worker pool size", mutate: func(cfg *Config) { cfg.Server.WorkerPoolSize = 0 }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := valid
			test.mutate(&cfg)
			if err := cfg.validate(); err == nil {
				t.Fatal("validate() error = nil")
			}
		})
	}
}
